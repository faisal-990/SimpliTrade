package service

import "github.com/faisal-990/ProjectInvestApp/backend/internal/repository"

type PortfolioService interface{}

type portfolioService struct {
	repo repository.PortfolioRepo
}

func NewPortfolioService(r repository.PortfolioRepo) PortfolioService {
	return &portfolioService{
		repo: r,
	}
}
