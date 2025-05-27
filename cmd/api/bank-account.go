package main

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/davmontas/teacherapp/internal/store"
)

type CreateBankAccounts struct {
	Name              string `json:"bank_name" validate:"required,max=100"`
	BankAccountNumber string `json:"bank_account_number" validate:"required,max=40"`
	UserProfileID     int64  `json:"user_profile_id" validate:"required"`
}

type UpdateBankAccounts struct {
	Name              *string `json:"bank_name" validate:"omitempty,max=100"`
	BankAccountNumber *string `json:"bank_account_number" validate:"omitempty,max=40"`
}

// GetBankAccount godoc
//
//	@Summary		Fetches a BankAccount
//	@Description	Fetches a BankAccount
//	@Tags			BankAccount
//	@Accept			json
//	@Produce		json
//	@Param			ID	path		int	true	"BankAccount ID"
//	@Success		200	{object}	store.BankAccount
//	@Failure		400	{object}	error
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//	@Security		ApiKeyAuth
//	@Router			/bank-accounts/{ID}	[get]
func (app *application) getBankAccountHandler(w http.ResponseWriter, r *http.Request) {
	entity := GetEntityFromContext[*store.BankAccount](r, bankAccountKEY)

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

// DeleteUser godoc
//
//	@Summary		Delete a  record
//	@Description	Delete a BankAccount record
//	@Tags			BankAccount
//	@Accept			json
//	@Produce		json
//	@Param			ID	path	int	true	"BankAccount ID"
//	@Success		204	"BankAccount record deleted successfully"
//	@Failure		400	{object}	error	"record not found"
//	@Security		ApiKeyAuth
//	@Router			/bank-accounts/{ID}	[delete]
func (app *application) deleteBankAccountHandler(w http.ResponseWriter, r *http.Request) {
	entity := GetEntityFromContext[*store.BankAccount](r, bankAccountKEY)

	ctx := r.Context()

	if err := app.store.BankAccount.Delete(ctx, int64(entity.ID)); err != nil {
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
//	@Summary		Update a BankAccount record
//	@Description	Update a BankAccount record
//	@Tags			BankAccount
//	@Accept			json
//	@Produce		json
//	@Param			ID	path	int	true	"BankAccount ID"
//	@Success		204	"BankAccount updated successfully"
//	@Failure		400	{object}	error	"BankAccount record not found"
//	@Security		ApiKeyAuth
//	@Router			/bank-accounts/{ID}	[patch]
func (app *application) updateBankAccountHandler(w http.ResponseWriter, r *http.Request) {
	entity := GetEntityFromContext[*store.BankAccount](r, bankAccountKEY)

	var payload UpdateBankAccounts
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if payload.Name != nil {
		entity.Name = *payload.Name
	}

	if payload.BankAccountNumber != nil {
		entity.BankAccountNumber = *payload.BankAccountNumber
	}

	if err := app.store.BankAccount.Update(r.Context(), entity); err != nil {
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

// GetUserCards godoc
//
//	@Summary		Gets all BankAccount from an user
//	@Description	Gets all BankAccount
//	@Tags			BankAccount
//	@Accept			json
//	@Produce		json
//	@Param			ID	path		int					true	"User's profile ID"
//	@Success		200	{array}		store.BankAccount	"List of user's BankAccount"
//	@Failure		400	{object}	error
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//	@Security		ApiKeyAuth
//	@Router			/bank-accounts/getUserCards/{ID} [get]
func (app *application) getUserCardsHandler(w http.ResponseWriter, r *http.Request) {
	userProfile := GetEntityFromContext[*store.UserProfile](r, userProfileKEY)
	entities, err := app.store.BankAccount.GetUserCards(r.Context(), userProfile.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerResponse(w, r, err)
		}

		return
	}

	if len(entities) == 0 {
		app.jsonResponse(w, http.StatusOK, map[string]string{
			"message": "Este usuario no posee cuentas de banco"})
		return
	}

	var response []store.BankAccountDTO
	for _, e := range entities {
		response = append(response, e.BankAccountToDTO())
	}

	app.jsonResponse(w, http.StatusOK, response)

}

// CreateBankAccount godoc
//
//	@Summary		Register an user's bankaccount
//	@Description	Register an user's bankaccount
//	@Tags			BankAccount
//	@Accept			json
//	@Produce		json
//	@Param			payload	body	CreateBankAccounts	true	"BankAccount info"
//	@Success		201
//	@Failure		400				{object}	error
//	@Failure		500				{object}	error
//	@Router			/bank-accounts	[post]
func (app *application) createBankAccountHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreateBankAccounts
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
	}

	entity := &store.BankAccount{
		Name:              payload.Name,
		BankAccountNumber: payload.BankAccountNumber,
		UserProfileID:     payload.UserProfileID,
	}

	if err := app.store.BankAccount.Create(r.Context(), entity); err != nil {
		app.internalServerResponse(w, r, err)
	}

	if err := app.jsonResponse(w, http.StatusCreated, payload); err != nil {
		app.internalServerResponse(w, r, err)
	}
}
