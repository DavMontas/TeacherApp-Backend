package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/davmontas/teacherapp/internal/mailer"
	"github.com/davmontas/teacherapp/internal/store"
	"github.com/davmontas/teacherapp/internal/store/enums"
	"github.com/google/uuid"
)

type RegisterUserPayload struct {
	Username string     `json:"username" validate:"required,max=100"`
	Email    string     `json:"email" validate:"required,email,max=255"`
	Password string     `json:"password" validate:"required,min=3,max=72"`
	Role     enums.Role `json:"role" validate:"required"`
}

type UserWithToken struct {
	*store.User
	Token string `json:"token"`
}

// RegisterUser godoc
//
//	@Summary		Register an user
//	@Description	Register an user
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			payload						body		RegisterUserPayload	true	"User Credentials"
//	@Success		201							{object}	UserWithToken		"User registered"
//	@Failure		400							{object}	error
//	@Failure		500							{object}	error
//	@Router			/authentication/register	[post]
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload RegisterUserPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &store.User{
		Role:     payload.Role,
		Username: payload.Username,
		Email:    payload.Email,
	}

	// hash user's password
	if err := user.Password.Set(payload.Password); err != nil {
		app.internalServerResponse(w, r, err)
		return
	}

	ctx := r.Context()

	plainToken := uuid.New().String()

	// hash the token for storage but keep the plain token for email
	hash := sha256.Sum256([]byte(plainToken))
	hashToken := hex.EncodeToString(hash[:])

	// store the user
	err := app.store.Users.CreateAndInvite(ctx, user, hashToken, app.config.mail.exp)
	if err != nil {
		switch err {
		case store.ErrDuplicateEmail:
			app.badRequestResponse(w, r, err)
		case store.ErrDuplicateUsername:
			app.badRequestResponse(w, r, err)
		default:
			app.internalServerResponse(w, r, err)
		}
		return
	}

	userWithToken := UserWithToken{
		User:  user,
		Token: plainToken,
	}

	isProdEnv := app.config.env == "prod"
	activationURL := fmt.Sprintf("%s/confirm/%s", app.config.frontendURL, plainToken)
	vars := struct {
		Username      string
		ActivationURL string
	}{
		Username:      user.Username,
		ActivationURL: activationURL,
	}
	// send mail
	_, err = app.mailer.Send(mailer.UserWelcomeTemplate, user.Username, user.Email, vars, !isProdEnv)
	if err != nil {
		app.logger.Errorw("error sending the welcome email", err)

		//rollback user creation if email fails (saga pattern)
		if err := app.store.Users.Delete(ctx, user.ID); err != nil {
			app.logger.Errorw("error deleting the user after the welcome email failed", err)
			app.internalServerResponse(w, r, err)
			return
		}

		app.internalServerResponse(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, userWithToken); err != nil {
		app.internalServerResponse(w, r, err)
		return
	}

}
