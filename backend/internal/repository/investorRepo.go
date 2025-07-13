package repository

import "gorm.io/gorm"

type InvestorRepo interface{}

type investorRepo struct {
	DB *gorm.DB
}

func NewInvestorRepo(db *gorm.DB) InvestorRepo {
	return &investorRepo{
		DB: db,
	}
}
