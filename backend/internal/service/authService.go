package service

import (
	"context"
	"strings"

	"github.com/faisal-990/ProjectInvestApp/backend/internal/dto"
	"github.com/faisal-990/ProjectInvestApp/backend/internal/models"
	"github.com/faisal-990/ProjectInvestApp/backend/internal/repository"
	"github.com/faisal-990/ProjectInvestApp/backend/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	RegisterUser(ctx context.Context, user *models.User) error
	AuthenticateUser(ctx context.Context, user *dto.Login) (*models.User, error)
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
	return nil
}

func (a *authservice) AuthenticateUser(ctx context.Context, input *dto.Login) (*models.User, error) {
	// check if user exist , and if the password is correct
	email := input.Email
	password := input.Password
	//basic checks on email and password
	if email == "" {
		return nil, utils.ErrInvalidEmail
	}
	if len(password) < 8 {
		return nil, utils.ErrInvalidPassword
	}
	email = strings.TrimSpace(strings.ToLower(email))

	//get the user from the databse

	user, err := a.repo.GetUserByEmail(ctx, email)

	//Check if user exist or not
	if err != nil {
		//TODO: define more custom error types for various different scenerios while db querying
		//for the moment the error is begin sent as is
		return nil, err
	}
	//check if password matches
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, utils.Wrap(utils.ErrWrongPassword, err)
	}

	//User exist and the password matches as well
	return user, nil
}

func (a *authservice) RequestResetPassword(ctx context.Context, email string) error {
	return nil
}
