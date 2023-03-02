package model

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"itfinder.adrianescat.com/internal/validator"
	"time"
)

type Salaries []*SalaryByRole

type Offer struct {
	ID          int64     `json:"id"`
	UserId      int64     `json:"user_id"`
	User        *User     `json:"user"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"-"`
	Title       string    `json:"title"`
	PictureUrl  string    `json:"picture_url"`
	Description string    `json:"description"`
	Salary      Salaries  `json:"salary"`
	Active      bool      `json:"-"`
	Version     int       `json:"-"`
}

type OfferModel struct {
	DB *sql.DB
}

func (s Salaries) MarshalGQL(w io.Writer) {
	err := json.NewEncoder(w).Encode(s)
	if err != nil {
		panic(err)
	}
}

func (s *Salaries) UnmarshalGQL(v interface{}) error {
	m, ok := v.([]*SalaryByRole)

	if !ok {
		return fmt.Errorf("%T is not a map", v)
	}

	*s = m
	return nil
}

func ValidateOffer(v *validator.Validator, offer *Offer) {
	v.Check(offer.Title != "", "title", "must be provided")
	v.Check(len(offer.Title) <= 150, "title", "must not be more than 100 bytes long")

	validator.ValidatePictureUrl(v, offer.PictureUrl)

	v.Check(offer.Description != "", "description", "must be provided")
	v.Check(offer.Salary != nil, "salary", "must be provided")
}

func (m OfferModel) Insert(offer *Offer) error {
	query := `
		INSERT INTO offers (user_id, title, picture_url, description, salary)
		VALUES ($1, $2, $3, $4, $5::jsonb)
		RETURNING id, created_at, version
	`

	salariesJSON, err := json.Marshal(offer.Salary)
	if err != nil {
		return err
	}

	args := []any{offer.UserId, offer.Title, offer.PictureUrl, offer.Description, salariesJSON}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	err = m.DB.QueryRowContext(ctx, query, args...).Scan(&offer.ID, &offer.CreatedAt, &offer.Version)

	if err != nil {
		return err
	}

	return nil
}

func (m OfferModel) GetAll() ([]*Offer, error) {
	query := `SELECT id, created_at, title, description, salary, picture_url, user_id, active FROM offers`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var args []any

	rows, err := m.DB.QueryContext(ctx, query, args...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var offers []*Offer

	for rows.Next() {
		var offer Offer
		var salaries []byte
		err := rows.Scan(
			&offer.ID,
			&offer.CreatedAt,
			&offer.Title,
			&offer.Description,
			&salaries,
			&offer.PictureUrl,
			&offer.UserId,
			&offer.Active,
		)

		if err != nil {
			return nil, err // Update this to return an empty Metadata struct.
		}

		// Decode JSON-encoded byte slice into Salaries object
		var salariesObj Salaries
		err = json.Unmarshal(salaries, &salariesObj)

		offer.Salary = salariesObj

		offers = append(offers, &offer)
	}

	if err = rows.Err(); err != nil {
		return nil, err // Update this to return an empty Metadata struct.
	}

	return offers, nil
}
