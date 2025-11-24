package repository

import (
	"context"

	"github.com/faisal-990/ProjectInvestApp/backend/internal/models"
	"gorm.io/gorm"
)

type AuthRepo interface {
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) error
}

type authRepo struct {
	DB *gorm.DB
}

func NewAuthRepo(db *gorm.DB) AuthRepo {
	return &authRepo{
		DB: db,
	}
}

func (a *authRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var object models.User

	err := a.DB.WithContext(ctx).
		Where("email = ?", email).
		First(&object).Error

	if err != nil {
		return nil, err
	}

	return &object, nil
}

func (a *authRepo) CreateUser(ctx context.Context, user *models.User) error {

	return a.DB.WithContext(ctx).Create(user).Error
}

func (a *authRepo) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	return nil, nil
}
