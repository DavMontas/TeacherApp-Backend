package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNotFound          = errors.New("record not found")
	QueryTimeoutDuration = time.Second * 5
)

type Storage struct {
	Users interface {
		GetAll(context.Context) ([]*User, error)
		Create(context.Context, *sql.Tx, *User) error
		CreateAndInvite(context.Context, *User, string, time.Duration) error
		GetByID(context.Context, int64) (*User, error)
		Delete(context.Context, int64) error
		Activate(context.Context, string) error
	}

	Profiles interface {
		GetByID(context.Context, int64) (*Profile, error)
		Update(ctx context.Context, profile *Profile) error
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Users:    &UsersStore{db},
		Profiles: &ProfilesStore{db},
	}
}

func withTx(db *sql.DB, ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
