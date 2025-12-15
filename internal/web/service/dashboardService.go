package service

import "github.com/faisal-990/ProjectInvestApp/internal/platform/repository"

type DashboardService interface{}

type dashboardService struct {
	repo repository.DashboardRepo
}

func NewDashboardService(r repository.DashboardRepo) DashboardService {
	return &dashboardService{
		repo: r,
	}
}
