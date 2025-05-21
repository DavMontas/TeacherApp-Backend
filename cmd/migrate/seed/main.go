package main

import (
	"log"

	"github.com/davmontas/teacherapp/internal/db"
	"github.com/davmontas/teacherapp/internal/env"
	"github.com/davmontas/teacherapp/internal/store"
)

func main() {
	addr := env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/teacherapp?sslmode=disable")
	conn, err := db.New(addr, 3, 3, "15m")
	if err != nil {
		log.Fatal("There was an error with the db: ", err)
	}

	defer conn.Close()

	store := store.NewStorage(conn)

	db.Seed(store, conn)
}