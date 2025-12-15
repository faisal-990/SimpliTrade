package repository

import "gorm.io/gorm"

type DashboardRepo interface{}

type dashboardRepo struct {
	DB *gorm.DB
}

func NewDashboardRepo(db *gorm.DB) DashboardRepo {
	return &dashboardRepo{
		DB: db,
	}
}
