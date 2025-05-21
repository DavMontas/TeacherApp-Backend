package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/davmontas/teacherapp/internal/store"
	"github.com/davmontas/teacherapp/internal/store/enums"
	"github.com/go-chi/chi/v5"
)

type CreateUserPayload struct {
	// FirstName string     `json:"first_name"`
	// LastName  string     `json:"last_name"`
	Username string     `json:"username" validate:"required,100"`
	Email    string     `json:"email" validate:"required,max=255" `
	Password string     `json:"password"`
	Role     enums.Role `json:"role"`
}

type UpdateUserPayload struct {
	Username *string     `json:"username" validate:"omitempty,max=100"`
	Email    *string     `json:"email" validate:"omitempty,max=255"`
	Password *string     `json:"password" validate:"omitempty"`
	Role     *enums.Role `json:"role" validate:"omitempty"`
}

// GetUser godoc
//
//	@Summary		Fetches an user
//	@Description	Fetches an user by ID
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	store.User
//	@Failure		400	{object}	error
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/{userID}	[get]
func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	user := GetEntityFromContext[*store.User](r, userKey)

	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerResponse(w, r, err)
		}

		return
	}
}

// GetAll User godoc
//
//	@Summary		Gets all users
//	@Description	Gets all users
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}		store.User	"List of users"
//	@Failure		400	{object}	error
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/getAll [get]
func (app *application) getAllUserHandler(w http.ResponseWriter, r *http.Request) {
	users, err := app.store.Users.GetAll(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			app.notFoundResponse(w, r, err)
		case len(users) == 0:
			log.Printf("Eta vaina ta' vacia")
		default:
			app.internalServerResponse(w, r, err)
		}

		return
	}

	app.jsonResponse(w, http.StatusOK, users)
}

// DeleteUser godoc
//
//	@Summary		Delete an user profile
//	@Description	Delete an user profile
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			userID	path	int	true	"User ID"
//	@Success		204		"User deleted successfully"
//	@Failure		400		{object}	error	"user not found"
//	@Security		ApiKeyAuth
//	@Router			/users/{userID}	[delete]
func (app *application) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	user := GetEntityFromContext[*store.User](r, userKey)

	ctx := r.Context()

	if err := app.store.Users.Delete(ctx, int64(user.ID)); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerResponse(w, r, err)
		}

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateUser godoc
//
//	@Summary		Update an user
//	@Description	Update an user
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			userID	path	int	true	"User ID"
//	@Success		204		"User updated successfully"
//	@Failure		400		{object}	error	"user not found"
//	@Security		ApiKeyAuth
//	@Router			/users/{userID}	[patch]
func (app *application) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	user := GetEntityFromContext[*store.User](r, userKey)

	var payload UpdateUserPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if payload.Email != nil {
		user.Email = *payload.Email
	}

	if payload.Password != nil {
		user.Password.Set(*payload.Password)
	}

	if payload.Role != nil {
		user.Role = *payload.Role
	}

	// TODO: Evaluar la necesidad de este metodo
	// if err := app.store.Users.Update(r.Context(), user); err != nil {
	// 	switch {
	// 	case errors.Is(err, store.ErrNotFound):
	// 		app.notFoundResponse(w, r, err)
	// 	default:
	// 		app.internalServerResponse(w, r, err)
	// 	}
	// 	return
	// }

	if err := app.jsonResponse(w, http.StatusCreated, user); err != nil {
		app.internalServerResponse(w, r, err)
		return
	}
}

// ActivateUser godoc
//
//	@Summary		Activates an user
//	@Description	Activates an user by an invitation token
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			token	path	string	true	"Invitation token"
//	@Success		204		"User activated"
//	@Failure		404		{object}	error	"user not found"
//	@Failure		500		{object}	error	"internal error"
//	@Security		ApiKeyAuth
//	@Router			/users/activate/{token}	[put]
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	err := app.store.Users.Activate(r.Context(), token)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerResponse(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusNoContent, "user activated"); err != nil {
		app.internalServerResponse(w, r, err)
	}

}
