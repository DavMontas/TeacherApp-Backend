package main

import (
	"time"

	"github.com/davmontas/teacherapp/internal/db"
	"github.com/davmontas/teacherapp/internal/env"
	"github.com/davmontas/teacherapp/internal/mailer"
	"github.com/davmontas/teacherapp/internal/store"
	"go.uber.org/zap"
)

const DB_ADDR = "DB_ADDR"
const DB_MAX_OPEN_CONNS = "DB_MAX_OPEN_CONNS"
const DB_MAX_IDLE_CONNS = "DB_MAX_IDLE_CONNS"
const DB_MAX_IDLE_TIME = "DB_MAX_IDLE_TIME"

//	@title			TeacherAPP API
//	@version		1.0
//	@description	API for TeacherApp, an app for teachers
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host						petstore.swagger.io
//	@BasePath					/v1
//
//	@securityDefinitions.apiKey	ApiKeyAuth
//	@in							header
//	@name						Authorization
//	@description

func main() {

	cfg := config{
		addr:        env.GetString("ADDR", ":8000"),
		apiURL:      env.GetString("EXTERNAL_URL", "localhost:8000"),
		frontendURL: env.GetString("FRONTEND_URL", "http://localhost:4000"),
		db: dbConfig{
			addr:         env.GetString(DB_ADDR, "postgres://admin:adminpassword@localhost/teacherapp?sslmode=disable"),
			maxOpenConns: env.GetInt(DB_MAX_OPEN_CONNS, 30),
			maxIdleConns: env.GetInt(DB_MAX_IDLE_CONNS, 30),
			maxIdleTime:  env.GetString(DB_MAX_IDLE_TIME, "15m"),
		},
		env: env.GetString("ENV", "prod"),
		mail: mailConfig{
			exp:       time.Hour * 72,
			fromEmail: env.GetString("FROM_EMAIL", ""),
			configurationType: mailerConfigurationType{
				apiKey: env.GetString("MAILTRAP_API_KEY", ""),
			},
		},
	}

	// Logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	// Database
	db, err := db.New(
		cfg.db.addr,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTime,
	)

	if err != nil {
		logger.Fatal(err)
	}

	defer db.Close()
	logger.Info("database connection pool stabishled! ;) ")

	mailer, err := mailer.NewMailTrapClient(cfg.mail.configurationType.apiKey, cfg.mail.fromEmail)
	if err != nil {
		logger.Fatal(err)
	}

	store := store.NewStorage(db)

	app := &application{
		config: cfg,
		store:  store,
		logger: logger,
		mailer: mailer,
	}

	mux := app.mount()
	logger.Fatal(app.run(mux))
}
