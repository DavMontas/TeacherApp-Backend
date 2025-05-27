package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/davmontas/teacherapp/internal/store"
	"github.com/go-chi/chi/v5"
)

const (
	userKEY        string = "user"
	userProfileKEY string = "profile"
	bankAccountKEY string = "bank_account"
)

func EntityContextMiddleware[T any](
	app *application,
	key any,
	param string,
	getByID func(ctx context.Context, id int64) (T, error),
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			idParam := chi.URLParam(r, param)
			id, err := strconv.ParseInt(idParam, 10, 64)
			if err != nil {
				app.internalServerResponse(w, r, err)
				return
			}

			entity, err := getByID(r.Context(), id)
			if err != nil {
				switch {
				case errors.Is(err, store.ErrNotFound):
					app.notFoundResponse(w, r, err)
				default:
					app.internalServerResponse(w, r, err)
				}
				return
			}

			ctx := context.WithValue(r.Context(), key, entity)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetEntityFromContext[T any](r *http.Request, key string) T {
	value := r.Context().Value(key)

	if typed, ok := value.(T); ok {
		return typed
	}

	var zero T
	return zero
}
