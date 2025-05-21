package main

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/davmontas/teacherapp/internal/store"
)

type UpdateUserProfilePayload struct {
	Identification *string `json:"identification" validate:"omitempty, max=11"`
	FirstName      *string `json:"first_name" validate:"omitempty, max=50"`
	LastName       *string `json:"last_name" validate:"omitempty, max=50"`
}

// GetProfile godoc
//
//	@Summary		Fetches an user profile
//	@Description	Fetches an user profile by ID
//	@Tags			Profiles
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Profile ID"
//	@Success		200	{object}	store.Profile
//	@Failure		400	{object}	error
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//	@Security		ApiKeyAuth
//	@Router			/profiles/{ID}	[get]
func (app *application) getProfilesHandler(w http.ResponseWriter, r *http.Request) {
	entity := GetEntityFromContext[*store.Profile](r, profileKey)

	if err := app.jsonResponse(w, http.StatusOK, entity); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerResponse(w, r, err)
		}

		return
	}

}

// UpdateUser godoc
//
//	@Summary		Update an user profile
//	@Description	Update an user profile
//	@Tags			Profiles
//	@Accept			json
//	@Produce		json
//	@Param			ID	path	int	true	"User ID"
//	@Success		204	"User's Profile updated successfully"
//	@Failure		400	{object}	error	"user's profile not found"
//	@Security		ApiKeyAuth
//	@Router			/profiles/{ID}	[patch]
func (app *application) updateProfilesHandler(w http.ResponseWriter, r *http.Request) {
	entity := GetEntityFromContext[*store.Profile](r, profileKey)

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
		entity.Identification = *payload.Identification
	}

	if payload.FirstName != nil {
		entity.FirstName = *payload.FirstName
	}

	if payload.LastName != nil {
		entity.LastName = *payload.LastName
	}

	if err := app.store.Profiles.Update(r.Context(), entity); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerResponse(w, r, err)
		}

		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, ""); err != nil {
		app.internalServerResponse(w, r, err)
		return
	}

}
