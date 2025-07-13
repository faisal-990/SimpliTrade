package repository

import "gorm.io/gorm"

type AuthRepo interface{}

type authRepo struct {
	DB *gorm.DB
}

func NewAuthRepo(db *gorm.DB) AuthRepo {
	return &authRepo{
		DB: db,
	}
}
