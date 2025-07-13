package service

import "github.com/faisal-990/ProjectInvestApp/backend/internal/repository"

type AuthService interface{}

type authservice struct {
	repo repository.AuthRepo
}

func NewAuthService(r repository.AuthRepo) AuthService {
	return &authservice{
		repo: r,
	}
}
