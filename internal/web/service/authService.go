package service

import (
	"context"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/repository"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/utils"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	RegisterUser(ctx context.Context, user *models.User) error
	AuthenticateUser(ctx context.Context, email, password string) (*models.User, error)
	RequestResetPassword(ctx context.Context, email string) error
}

type authservice struct {
	repo repository.AuthRepo
}

func NewAuthService(r repository.AuthRepo) AuthService {
	return &authservice{
		repo: r,
	}
}

func (a *authservice) RegisterUser(ctx context.Context, user *models.User) error {
	// Extract plain password
	plainPassword := user.Password

	// Hash it
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		utils.LogError("failed to hash password", err)
		return err
	}

	// Replace plain with hashed
	user.Password = string(hashedPassword)

	// Store user
	if err := a.repo.AddUser(ctx, user); err != nil {
		utils.LogError("failed to register user", err)
		return err
	}

	return nil
}

func (a *authservice) AuthenticateUser(ctx context.Context, email, password string) (*models.User, error) {
	// check if user exist , and if the password is correct
	return nil, nil
}

func (a *authservice) RequestResetPassword(ctx context.Context, email string) error {
	return nil
}
