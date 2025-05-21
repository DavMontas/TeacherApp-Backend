package store

import (
	"context"
	"database/sql"
	"errors"
)

type Profile struct {
	ID             int    `json:"id"`
	Identification string `json:"identification"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
	UserId         int64  `json:"user_id"`
}

type ProfilesStore struct {
	db *sql.DB
}

func (s *ProfilesStore) GetByID(ctx context.Context, id int64) (*Profile, error) {
	query := `
		SELECT id, identification, first_name, last_name, created_at
		FROM profiles
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var profile Profile

	err := s.db.QueryRowContext(
		ctx,
		query,
		id).Scan(
		&profile.ID,
		&profile.Identification,
		&profile.FirstName,
		&profile.LastName,
		&profile.CreatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			return nil, err
		default:
			return nil, err
		}
	}

	return &profile, nil
}

func (s *ProfilesStore) Update(ctx context.Context, profile *Profile) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		query := `
		UPDATE profiles
		SET first_name = $1, 
		last_name = $2, 
		identification = $3, 
		WHERE id = $4
	`

		ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
		defer cancel()

		_, err := tx.ExecContext(
			ctx,
			query,
			profile.FirstName,
			profile.LastName,
			profile.Identification,
			profile.ID,
		)

		if err != nil {
			switch err {
			case sql.ErrNoRows:
				return ErrNotFound
			default:
				return err
			}
		}

		return nil
	})
}
