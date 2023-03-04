package model

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Users    UserModel
	Offers   OfferModel
	Profiles ProfileModel
	Tokens   TokenModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Users:    UserModel{DB: db},
		Offers:   OfferModel{DB: db},
		Profiles: ProfileModel{DB: db},
		Tokens:   TokenModel{DB: db},
	}
}
