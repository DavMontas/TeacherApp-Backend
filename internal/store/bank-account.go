package store

import (
	"context"
	"database/sql"
	"errors"
)

type BankAccount struct {
	ID                int64  `json:"id"`
	Name              string `json:"bank_name"`
	BankAccountNumber string `json:"bank_account_number"`
	UserProfileID     int64  `json:"user_profile_id"`
	CreatedAt         string `json:"created_at"`
}

type BankAccountDTO struct {
	ID                int64  `json:"id"`
	Name              string `json:"bank_name"`
	BankAccountNumber string `json:"bank_account_number"`
	CreatedAt         string `json:"created_at"`
}

type BankAccountsStore struct {
	db *sql.DB
}

func (s *BankAccountsStore) GetUserCards(ctx context.Context, userProfileID int64) ([]*BankAccount, error) {
	query := `
	SELECT id, bank_name, bank_account_number, created_at
	FROM bank_accounts
	WHERE user_profile_id = $1
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, userProfileID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var cards []*BankAccount
	for rows.Next() {
		var b BankAccount
		if err := rows.Scan(
			&b.ID,
			&b.Name,
			&b.BankAccountNumber,
			&b.CreatedAt,
		); err != nil {
			return nil, err
		}
		cards = append(cards, &b)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cards, nil
}

func (s *BankAccountsStore) Create(ctx context.Context, baDetail *BankAccount) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		query := `
		INSERT INTO bank_accounts (bank_name, bank_account_number, user_profile_id)
		VALUES ($1, $2,$3)
	`

		ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
		defer cancel()
		_, err := tx.ExecContext(
			ctx,
			query,
			baDetail.Name,
			baDetail.BankAccountNumber,
			baDetail.UserProfileID,
		)

		if err != nil {
			return err
		}

		return nil
	})
}

func (s *BankAccountsStore) GetByID(ctx context.Context, id int64) (*BankAccount, error) {
	query := `
		SELECT id, bank_name, bank_account_number, user_profile_id
		FROM bank_accounts
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var baDetail BankAccount

	err := s.db.QueryRowContext(
		ctx,
		query,
		id).Scan(
		&baDetail.ID,
		&baDetail.Name,
		&baDetail.BankAccountNumber,
		&baDetail.UserProfileID,
	)

	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			return nil, err
		default:
			return nil, err
		}
	}

	return &baDetail, nil
}

func (s *BankAccountsStore) Update(ctx context.Context, baDetail *BankAccount) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		query := `
		UPDATE bank_accounts
		SET bank_name = $1, 
		bank_account_number = $2, 
		WHERE id = $3
	`

		ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
		defer cancel()

		_, err := tx.ExecContext(
			ctx,
			query,
			baDetail.Name,
			baDetail.BankAccountNumber,
			baDetail.ID,
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

func (s *BankAccountsStore) Delete(ctx context.Context, id int64) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		query := `
		DELETE FROM bank_accounts
		WHERE id = $1
	`

		ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
		defer cancel()

		_, err := tx.ExecContext(ctx, query, id)
		if err != nil {
			return err
		}

		return nil
	})
}

func (ba BankAccount) BankAccountToDTO() BankAccountDTO {
	return BankAccountDTO{
		ID:                ba.ID,
		Name:              ba.Name,
		BankAccountNumber: ba.BankAccountNumber,
		CreatedAt:         ba.CreatedAt,
	}
}
