package store

import (
	"context"
	"database/sql"
	"errors"
)

// swagger:model UserProfile
type UserProfile struct {
	ID             int64          `json:"id"`
	Identification sql.NullString `json:"identification"`
	FirstName      sql.NullString `json:"first_name"`
	LastName       sql.NullString `json:"last_name"`
	CreatedAt      string         `json:"created_at"`
	UpdatedAt      string         `json:"updated_at"`
	UserId         int64          `json:"user_id"`
}

// swagger:model UserProfileDTO
type UserProfileDTO struct {
	ID             int64  `json:"id"`
	Identification string `json:"identification"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	CreatedAt      string `json:"created_at"`
	UserId         int64  `json:"user_id"`
}

type UserProfilesStore struct {
	db *sql.DB
}

func (s *UserProfilesStore) GetByID(ctx context.Context, id int64) (*UserProfile, error) {
	query := `
		SELECT id, identification, first_name, last_name, user_id, created_at
		FROM user_profiles
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var profile UserProfile
	err := s.db.QueryRowContext(
		ctx,
		query,
		id).Scan(
		&profile.ID,
		&profile.Identification,
		&profile.FirstName,
		&profile.LastName,
		&profile.UserId,
		&profile.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, err
		}
		return nil, err
	}

	return &profile, nil
}

func (s *UserProfilesStore) Update(ctx context.Context, profile *UserProfile) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		query := `
		UPDATE user_profiles
		SET first_name = $1, 
		last_name = $2, 
		identification = $3
		WHERE id = $4
	`

		ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
		defer cancel()

		res, err := tx.ExecContext(
			ctx,
			query,
			profile.FirstName,
			profile.LastName,
			profile.Identification,
			profile.ID,
		)

		if err != nil {
			return err
		}

		if rowsAffected, err := res.RowsAffected(); err != nil {
			return err
		} else if rowsAffected == 0 {
			return ErrNotFound
		}

		return nil
	})
}

func (s *UserProfilesStore) GetAll(ctx context.Context) ([]*UserProfile, error) {
	query := `
	SELECT id, identification, first_name, last_name, user_id, created_at
	FROM user_profiles
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var entities []*UserProfile
	for rows.Next() {
		var e UserProfile
		if err := rows.Scan(
			&e.ID,
			&e.Identification,
			&e.FirstName,
			&e.LastName,
			&e.UserId,
			&e.CreatedAt,
		); err != nil {
			return nil, err
		}
		entities = append(entities, &e)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return entities, nil
}

func (p UserProfile) UserProfileToDTO() UserProfileDTO {
	response := UserProfileDTO{
		ID:        p.ID,
		CreatedAt: p.CreatedAt,
		UserId:    p.UserId,
	}

	if p.Identification.Valid {
		response.Identification = p.Identification.String
	}

	if p.FirstName.Valid {
		response.FirstName = p.FirstName.String
	}

	if p.LastName.Valid {
		response.LastName = p.LastName.String
	}

	return response
}
