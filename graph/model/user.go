package model

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
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
	Roles     []string  `json:"roles"`
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

func ValidateRoles(v *validator.Validator, roles []string) {
	permittedRoles := []string{"recruiter", "candidate"}
	for _, role := range roles {
		v.Check(role != "", "role", "must be provided")
		v.Check(validator.PermittedValue(role, permittedRoles...), "role", fmt.Sprintf("%s is not a permitted role", role))
	}
	v.Check(validator.Unique(roles), "roles", "roles must be unique")
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

	ValidateRoles(v, user.Roles)
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

	role := user.Roles[0]
	query = `SELECT id from roles WHERE code = $1`
	ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var roleId int64

	err = m.DB.QueryRowContext(ctx, query, role).Scan(&roleId)
	if err != nil {
		return err
	}

	query = `
		INSERT INTO users_roles (user_id, role_id)
		VALUES ($1, $2)
		RETURNING user_id
	`

	args = []any{user.ID, roleId}

	ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var result int64
	err = m.DB.QueryRowContext(ctx, query, args...).Scan(&result)
	if err != nil {
		return err
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

func (m UserModel) GetRolesByUserId(id int64) ([]string, error) {
	query := `
		SELECT code FROM roles
		INNER JOIN users_roles ur on roles.id = ur.role_id
		WHERE ur.user_id = $1;
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, id)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var roles []string

	for rows.Next() {
		var role string
		err := rows.Scan(
			&role,
		)

		if err != nil {
			return nil, err
		}

		roles = append(roles, role)
	}

	if err = rows.Err(); err != nil {
		return nil, err // Update this to return an empty Metadata struct.
	}

	return roles, nil
}

func (m UserModel) GetForToken(tokenScope, tokenPlaintext string) (*User, error) {
	// Calculate the SHA-256 hash of the plaintext token provided by the client.
	// Remember that this returns a byte *array* with length 32, not a slice.
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	// Set up the SQL query.
	query := `
		SELECT id, created_at, name, lastname,  email, activated, version
		FROM users
		INNER JOIN tokens
		ON id = tokens.user_id
		WHERE tokens.hash = $1
		AND tokens.scope = $2
		AND tokens.expiry > $3
	`

	// Create a slice containing the query arguments. Notice how we use the [:] operator
	// to get a slice containing the token hash, rather than passing in the array (which
	// is not supported by the pq driver), and that we pass the current time as the
	// value to check against the token expiry.
	args := []any{tokenHash[:], tokenScope, time.Now()}

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Execute the query, scanning the return values into a User struct. If no matching
	// record is found we return an ErrRecordNotFound error.
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
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

	// Return the matching user.
	return &user, nil
}

func (m UserModel) GetByEmail(email string) (*User, error) {
	query := `
		SELECT id, created_at, updated_at, name, lastname, email, activated, password_hash, version
		FROM users
		WHERE email = $1
	`
	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Name,
		&user.Lastname,
		&user.Email,
		&user.Activated,
		&user.Password.Hash,
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

func (m UserModel) GetAllBookmarksByUserId(id int64) ([]*Profile, error) {
	query := `
		SELECT p.id, p.user_id, p.created_at, p.title, p.about, p.status, p.country, p.state, p.city, p.picture_url, p.website_url, p.salary, p.version
		FROM profiles p
		INNER JOIN profile_bookmarks pb on p.id = pb.profile_id
		LEFT JOIN users u on u.id = pb.user_id
		WHERE u.id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, id)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var profiles []*Profile

	for rows.Next() {
		var profile Profile
		var salaries []byte
		err := rows.Scan(
			&profile.ID,
			&profile.UserId,
			&profile.CreatedAt,
			&profile.Title,
			&profile.About,
			&profile.Status,
			&profile.Country,
			&profile.State,
			&profile.City,
			&profile.PictureUrl,
			&profile.WebsiteUrl,
			&salaries,
			&profile.Version,
		)

		if err != nil {
			return nil, err
		}

		var salariesObj Salaries
		err = json.Unmarshal(salaries, &salariesObj)
		profile.Salary = salariesObj

		if err != nil {
			return nil, err
		}

		profiles = append(profiles, &profile)
	}

	if err = rows.Err(); err != nil {
		return nil, err // Update this to return an empty Metadata struct.
	}

	return profiles, nil
}

func (m *UserModel) CreateProfileBookmark(userId int64, profileId int64) error {
	query := `
		INSERT INTO profile_bookmarks (user_id, profile_id)
		VALUES ($1, $2)
		RETURNING user_id, profile_id 
	`

	args := []any{userId, profileId}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (m *UserModel) DeleteProfileBookmark(userId int64, profileId int64) error {
	query := `
		DELETE FROM profile_bookmarks
		WHERE user_id = $1 AND profile_id = $2
	`

	args := []any{userId, profileId}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}
