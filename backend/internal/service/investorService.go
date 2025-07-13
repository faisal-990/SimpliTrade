package service

import "github.com/faisal-990/ProjectInvestApp/backend/internal/repository"

type InvestorService interface{}

type investorService struct {
	repo repository.InvestorRepo
}

func NewInvestorService(r repository.InvestorRepo) InvestorService {
	return &investorService{
		repo: r,
	}
}
