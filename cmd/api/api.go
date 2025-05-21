package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/davmontas/teacherapp/cmd/api/configurations"
	mailer "github.com/davmontas/teacherapp/internal/mailer"
	"github.com/davmontas/teacherapp/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2" // http-swagger middleware
	"go.uber.org/zap"
)

type application struct {
	config config
	store  store.Storage
	logger *zap.SugaredLogger
	mailer mailer.Client
}

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

type mailConfig struct {
	exp               time.Duration
	configurationType mailerConfigurationType
	fromEmail         string
}

type mailerConfigurationType struct {
	apiKey string
}

type config struct {
	addr        string
	db          dbConfig
	env         string
	apiURL      string
	mail        mailConfig
	frontendURL string
}

const version = "0.0.1"

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", app.healthCheckHandler)

		docsURL := fmt.Sprintf("%s/swagger/doc.json", app.config.addr)
		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(docsURL)))

		r.Route("/users", func(r chi.Router) {
			r.Get("/getAll", app.getAllUserHandler)
			r.Put("/activate/{token}", app.activateUserHandler)

			r.Route("/{userID}", func(r chi.Router) {
				r.Use(EntityContextMiddleware(
					app,
					userKey,
					"userID",
					app.store.Users.GetByID,
				))

				r.Get("/", app.getUserHandler)

				r.Delete("/", app.deleteUserHandler)
				r.Patch("/", app.updateUserHandler)
			})
		})

		// Public routes
		r.Route("/authentication", func(r chi.Router) {
			r.Post("/register", app.registerUserHandler)
		})

		r.Route("/user-profiles", func(r chi.Router) {

			r.Route("/{ID}", func(r chi.Router) {
				r.Use(EntityContextMiddleware(
					app,
					userKey,
					"ID",
					app.store.Profiles.GetByID,
				))

				r.Get("/", app.getProfilesHandler)
				r.Patch("/", app.updateProfilesHandler)
			})
		})
	})

	return r
}

func (app *application) run(mux http.Handler) error {

	// Swag Docs Config
	configurations.SwaggerInfo(version, app.config.apiURL)

	srv := http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	app.logger.Infow("App running", "Address", srv.Addr)

	return srv.ListenAndServe()
}
