package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"log"
	"time"

	"github.com/davmontas/teacherapp/internal/store/enums"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail    = errors.New(("a user with that email already exists"))
	ErrDuplicateUsername = errors.New(("a user with that username already exists"))
)

type User struct {
	ID        int64      `json:"id" `
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	Password  password   `json:"-"`
	Role      enums.Role `json:"role"`
	IsActive  bool       `json:"is_active"`
	CreatedAt string     `json:"created_at"`
	UpdatedAt string     `json:"updated_at"`
	Version   int        `json:"version"`
}

type password struct {
	text *string
	hash []byte
}

type UsersStore struct {
	db *sql.DB
}

func (p *password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)

	if err != nil {
		return err
	}

	p.text = &text
	var test = &p.text
	log.Println(test, p.text)
	p.hash = hash

	return nil
}

func (s *UsersStore) GetAll(ctx context.Context) ([]*User, error) {
	query := `
	SELECT id, username, email, role, created_at, version
	FROM users
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var users []*User
	err := s.db.QueryRowContext(
		ctx,
		query,
	).Scan(
		pq.Array(users),
	)

	if err != nil {
		return nil, err
	}

	return users, nil
}

func (s *UsersStore) Create(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `
		INSERT INTO users (username, email, password, role)
		VALUES ($1, $2, $3, $4) 
		RETURNING id, created_at
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := tx.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.Email,
		user.Password.hash,
		user.Role,
	).Scan(
		&user.ID,
		&user.CreatedAt,
	)

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case err.Error() == `pq: duplicate key value violates unique constraint "users_username_key"`:
			return ErrDuplicateUsername
		default:
			return err
		}
	}

	// User profile creation
	if err := s.createUserProfile(ctx, tx, user.ID); err != nil {
		return err
	}

	return nil
}

func (s *UsersStore) CreateAndInvite(ctx context.Context, user *User, token string, invitationExp time.Duration) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		// create the user
		if err := s.Create(ctx, tx, user); err != nil {
			return err
		}

		// create user invitation
		if err := s.createUserInvitation(ctx, tx, token, invitationExp, user.ID); err != nil {
			return err
		}

		return nil
	})
}

func (s *UsersStore) GetByID(ctx context.Context, id int64) (*User, error) {
	query := `
		SELECT id, username, email, role, created_at, version, is_active
		FROM users
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var user User

	err := s.db.QueryRowContext(
		ctx,
		query,
		id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Role,
		&user.CreatedAt,
		&user.Version,
		&user.IsActive,
	)

	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			return nil, err
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (s *UsersStore) Delete(ctx context.Context, id int64) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		if err := s.delete(ctx, tx, id); err != nil {
			return err
		}

		if err := s.deleteUserProfile(ctx, tx, id); err != nil {
			return err
		}

		if err := s.deleteUserInvitation(ctx, tx, id); err != nil {
			return err
		}

		return nil
	})
}

func (s *UsersStore) delete(ctx context.Context, tx *sql.Tx, id int64) error {
	query := `
		DELETE FROM users
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *UsersStore) Update(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `
		UPDATE users
		SET username = $1, 
		email = $2, 
		password = $3, 
		role = $4, 
		is_active = $5
		WHERE id = $6
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(
		ctx,
		query,
		user.Username,
		user.Email,
		user.Password.hash,
		user.Role,
		user.IsActive,
		user.ID,
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
}

func (s *UsersStore) Activate(ctx context.Context, token string) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		user, err := s.getUserFromInvitation(ctx, tx, token)
		if err != nil {
			return err
		}

		user.IsActive = true
		if err := s.Update(ctx, tx, user); err != nil {
			return err
		}

		if err := s.deleteUserInvitation(ctx, tx, user.ID); err != nil {
			return err
		}

		return nil
	})
}

func (*UsersStore) getUserFromInvitation(ctx context.Context, tx *sql.Tx, token string) (*User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.created_at, u.is_active
		FROM users u
		JOIN user_invitations ui ON u.id = ui.user_id
		WHERE ui.token = $1 AND ui.expiration > $2
	`

	hash := sha256.Sum256([]byte(token))
	hashToken := hex.EncodeToString(hash[:])

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	entity := &User{}
	err := tx.QueryRowContext(
		ctx,
		query,
		hashToken,
		time.Now(),
	).Scan(
		&entity.ID,
		&entity.Username,
		&entity.Email,
		&entity.CreatedAt,
		&entity.IsActive,
	)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return entity, nil
}

func (s *UsersStore) createUserInvitation(ctx context.Context, tx *sql.Tx, token string, exp time.Duration, id int64) error {
	query := `
		INSERT INTO user_invitations(token, user_id, expiration)
		VALUES($1, $2, $3)
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, token, id, time.Now().Add(exp))
	if err != nil {
		return err
	}

	return nil
}

func (*UsersStore) deleteUserInvitation(ctx context.Context, tx *sql.Tx, userID int64) error {
	query := `
		DELETE from user_invitations
		WHERE user_id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}

	return nil
}

// # User Profile #

func (s *UsersStore) createUserProfile(ctx context.Context, tx *sql.Tx, userID int64) error {
	query := `
	INSERT INTO profiles (user_id)
	VALUES ($1)
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(
		ctx,
		query,
		userID,
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *UsersStore) deleteUserProfile(ctx context.Context, tx *sql.Tx, id int64) error {
	query := `
		DELETE FROM profiles
		WHERE user_id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}
