package model

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"itfinder.adrianescat.com/internal/validator"
	"time"
)

type Profile struct {
	ID         int64     `json:"id"`
	UserId     int64     `json:"user_id"`
	User       *User     `json:"user"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"-"`
	Title      string    `json:"title"`
	About      string    `json:"about"`
	Status     string    `json:"status"`
	Country    string    `json:"country"`
	State      string    `json:"state"`
	City       string    `json:"city"`
	PictureUrl string    `json:"picture_url"`
	WebsiteUrl string    `json:"website_url"`
	Salary     Salaries  `json:"salary"`
	Version    int       `json:"-"`
}

type ProfileModel struct {
	DB *sql.DB
}

func ValidateProfile(v *validator.Validator, profile *Profile) {
	v.Check(profile.Title != "", "title", "must be provided")
	v.Check(len(profile.Title) <= 150, "title", "must not be more than 150 bytes long")

	v.Check(profile.About != "", "about", "must be provided")
	v.Check(len(profile.About) <= 2000, "about", "must not be more than 2000 bytes long")

	v.Check(profile.Status != "", "status", "must be provided")
	v.Check(len(profile.Status) <= 20, "status", "must not be more than 20 bytes long")

	v.Check(profile.Country != "", "country", "must be provided")
	v.Check(len(profile.Country) <= 50, "country", "must not be more than 50 bytes long")

	v.Check(profile.State != "", "state", "must be provided")
	v.Check(len(profile.State) <= 50, "state", "must not be more than 50 bytes long")

	v.Check(profile.City != "", "city", "must be provided")
	v.Check(len(profile.City) <= 100, "city", "must not be more than 100 bytes long")

	validator.ValidatePictureUrl(v, profile.PictureUrl)
	validator.ValidateWebsiteUrl(v, profile.WebsiteUrl)

	v.Check(profile.Salary != nil, "salary", "must be provided")
}

func (p ProfileModel) Insert(profile *Profile) error {
	query := `
		INSERT INTO profiles (user_id, title, about, status, country, state, city, picture_url, website_url, salary)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10::jsonb)
		RETURNING id, created_at, version
	`

	salariesJSON, err := json.Marshal(profile.Salary)
	if err != nil {
		return err
	}

	args := []any{
		profile.UserId,
		profile.Title,
		profile.About,
		profile.Status,
		profile.Country,
		profile.State,
		profile.City,
		profile.PictureUrl,
		profile.WebsiteUrl,
		salariesJSON,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = p.DB.QueryRowContext(ctx, query, args...).Scan(&profile.ID, &profile.CreatedAt, &profile.Version)

	if err != nil {
		return err
	}

	return nil
}

func (p ProfileModel) GetProfileById(id int64) (*Profile, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `SELECT id, user_id, created_at, title, about, status, country, state, city, picture_url, website_url, salary, version FROM profiles WHERE id = $1`

	var profile Profile

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	var salaries []byte
	err := p.DB.QueryRowContext(ctx, query, id).Scan(
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
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	var salariesObj Salaries
	err = json.Unmarshal(salaries, &salariesObj)
	profile.Salary = salariesObj

	if err != nil {
		return nil, err
	}

	return &profile, nil
}

func (p ProfileModel) GetProfileByUserId(id int64) (*Profile, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `SELECT id, user_id, created_at, title, about, status, country, state, city, picture_url, website_url, salary, version FROM profiles WHERE user_id = $1`

	var profile Profile

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	var salaries []byte
	err := p.DB.QueryRowContext(ctx, query, id).Scan(
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
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	var salariesObj Salaries
	err = json.Unmarshal(salaries, &salariesObj)
	profile.Salary = salariesObj

	if err != nil {
		return nil, err
	}

	return &profile, nil
}
