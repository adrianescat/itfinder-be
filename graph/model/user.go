package model

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

var AnonymousUser = &User{}

type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"created_at"`
	DeletedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Lastname  string    `json:"name"`
	Email     string    `json:"email"`
	// Password  Password  `json:"-"`
	Activated bool `json:"activated"`
	Version   int  `json:"-"`
}

type Password struct {
	plaintext *string
	hash      []byte
}

type UserModel struct {
	DB *sql.DB
}

func (m UserModel) Insert(user *User) error {
	query := `
		INSERT INTO users (name, lastname, email, activated)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version
	`

	args := []any{user.Name, user.Lastname, user.Email, user.Activated}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	return nil
}
