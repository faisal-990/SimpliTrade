package repository

import "gorm.io/gorm"

type PortfolioRepo interface{}

type portfolioRepo struct {
	DB *gorm.DB
}

func NewPortfolioRepo(db *gorm.DB) PortfolioRepo {
	return &portfolioRepo{
		DB: db,
	}
}
