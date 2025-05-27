package main

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/davmontas/teacherapp/internal/store"
)

type UpdateUserProfilePayload struct {
	Identification *string `json:"identification" validate:"omitempty,max=11"`
	FirstName      *string `json:"first_name" validate:"omitempty,max=50"`
	LastName       *string `json:"last_name" validate:"omitempty,max=50"`
}

// GetProfile godoc
//
//	@Summary		Fetches an user profile
//	@Description	Fetches an user profile by ID
//	@Tags			Profiles
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Profile ID"
//	@Success		200	{object}	store.UserProfileDTO
//	@Failure		400	{object}	error
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//	@Security		ApiKeyAuth
//	@Router			/user-profiles/{id}	[get]
func (app *application) getProfileHandler(w http.ResponseWriter, r *http.Request) {
	entity := GetEntityFromContext[*store.UserProfile](r, userProfileKEY)

	if err := app.jsonResponse(w, http.StatusOK, entity.UserProfileToDTO()); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerResponse(w, r, err)
		}

		return
	}

}

// GetAllProfile godoc
//
//	@Summary		List profiles
//	@Description	Get profiles
//	@Tags			Profiles
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}		store.UserProfileDTO
//	@Failure		400	{object}	error
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//	@Security		ApiKeyAuth
//	@Router			/user-profiles/get-all	[get]
func (app *application) getAllProfileHandler(w http.ResponseWriter, r *http.Request) {
	entities, err := app.store.UserProfiles.GetAll(r.Context())
	if err != nil {
		app.internalServerResponse(w, r, err)
		return
	}

	var response []store.UserProfileDTO
	for _, p := range entities {
		response = append(response, p.UserProfileToDTO())
	}
	app.jsonResponse(w, http.StatusOK, response)
}

// UpdateUser godoc
//
//	@Summary		Update an user profile
//	@Description	Update an user profile
//	@Tags			Profiles
//	@Accept			json
//	@Produce		json
//	@Param			id		path	int							true	"User ID"
//	@Param			payload	body	UpdateUserProfilePayload	true	"User Credentials"
//	@Success		204		"User's Profile updated successfully"
//	@Failure		400		{object}	error	"user's profile not found"
//	@Security		ApiKeyAuth
//	@Router			/user-profiles/{id}	[patch]
func (app *application) updateProfileHandler(w http.ResponseWriter, r *http.Request) {
	entity := GetEntityFromContext[*store.UserProfile](r, userProfileKEY)

	var payload UpdateUserProfilePayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if payload.Identification != nil {
		entity.Identification = sql.NullString{String: *payload.Identification, Valid: true}
	}

	if payload.FirstName != nil {
		entity.FirstName = sql.NullString{String: *payload.FirstName, Valid: true}
	}

	if payload.LastName != nil {
		entity.LastName = sql.NullString{String: *payload.LastName, Valid: true}
	}

	if err := app.store.UserProfiles.Update(r.Context(), entity); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerResponse(w, r, err)
		}

		return
	}

	if err := app.jsonResponse(w, http.StatusNoContent, ""); err != nil {
		app.internalServerResponse(w, r, err)
		return
	}
}
