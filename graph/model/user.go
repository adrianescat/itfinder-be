package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"itfinder.adrianescat.com/internal/validator"
	"strings"
	"time"
)

var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

var AnonymousUser = &User{}

type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"-"`
	Name      string    `json:"name"`
	Lastname  string    `json:"lastname"`
	Email     string    `json:"email"`
	Password  Password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

type Password struct {
	Plaintext *string
	Hash      []byte
}

type UserModel struct {
	DB *sql.DB
}

func (p *Password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)

	if err != nil {
		return err
	}

	p.Plaintext = &plaintextPassword
	p.Hash = hash

	return nil
}

// Matches method checks whether the provided plaintext password matches the
// hashed password stored in the struct, returning true if it matches and false
// otherwise.
func (p *Password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.Hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 500, "name", "must not be more than 500 bytes long")

	v.Check(user.Lastname != "", "lastname", "must be provided")
	v.Check(len(user.Lastname) <= 500, "lastname", "must not be more than 500 bytes long")

	// Call the standalone ValidateEmail() helper.
	ValidateEmail(v, user.Email)

	// If the plaintext password is not nil, call the standalone
	// ValidatePasswordPlaintext() helper.
	if user.Password.Plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.Plaintext)
	}

	// If the password hash is ever nil, this will be due to a logic error in our
	// codebase (probably because we forgot to set a password for the user). It's a
	// useful sanity check to include here, but it's not a problem with the data
	// provided by the client. So rather than adding an error to the validation map we
	// raise a panic instead.
	if user.Password.Hash == nil {
		panic("missing password hash for user")
	}
}

func (m UserModel) Insert(user *User) error {
	query := `
		INSERT INTO users (name, lastname, email, password_hash, activated)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, version
	`

	args := []any{user.Name, user.Lastname, user.Email, user.Password.Hash, user.Activated}
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

func (m UserModel) GetAll() ([]*User, error) {
	query := `
		SELECT id, created_at, updated_at, name, lastname, email, activated, version
		FROM users`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var args []any

	rows, err := m.DB.QueryContext(ctx, query, args...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []*User

	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.ID,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.Name,
			&user.Lastname,
			&user.Email,
			&user.Activated,
			&user.Version,
		)

		if err != nil {
			return nil, err // Update this to return an empty Metadata struct.
		}

		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, err // Update this to return an empty Metadata struct.
	}

	return users, nil
}

func (m UserModel) GetUsersByIds(ids []string) ([]*User, error) {
	idString := strings.Join(ids, ",")

	query := fmt.Sprintf("SELECT id, created_at, updated_at, name, lastname, email, activated, version FROM users WHERE id IN (%s)", idString)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []*User

	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.ID,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.Name,
			&user.Lastname,
			&user.Email,
			&user.Activated,
			&user.Version,
		)

		if err != nil {
			return nil, err // Update this to return an empty Metadata struct.
		}

		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, err // Update this to return an empty Metadata struct.
	}

	return users, nil
}

func (m UserModel) GetById(id int64) (*User, error) {
	query := `
		SELECT id, created_at, name, lastname, email, activated, version
		FROM users
		WHERE id = $1
	`
	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Lastname,
		&user.Email,
		&user.Activated,
		&user.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}
