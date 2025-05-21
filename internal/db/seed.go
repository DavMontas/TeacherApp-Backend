package db

import (
	"context"
	"database/sql"
	"log"
	"strconv"
	"time"

	"github.com/davmontas/teacherapp/internal/store"
	"github.com/davmontas/teacherapp/internal/store/enums"
)

var roles = []enums.Role{1, 2, 3}

func Seed(store store.Storage, db *sql.DB) {
	ctx := context.Background()

	users := generateUsers(3)
	tx, _ := db.BeginTx(ctx, nil)

	for _, user := range users {
		if err := store.Users.Create(ctx, &sql.Tx{}, user); err != nil {
			_ = tx.Rollback()
			log.Println("Error creating user: ", err)
			return
		}
	}

	tx.Commit()

	log.Println("Seeding complete")
}

func generateUsers(n int) []*store.User {
	users := make([]*store.User, n)
	for i := 0; i < n; i++ {
		users[i] = &store.User{
			Email: "user" + strconv.Itoa(i) + "@example.com",
			// Password:  "password",
			Role:      roles[i%len(roles)],
			CreatedAt: time.Now().String(),
		}
	}
	return users
}
