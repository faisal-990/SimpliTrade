package repository

import (
	"context"
	"errors"
	"time"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrNotFound is returned by repository lookups when no row matches. Services
// switch on it (errors.Is) to distinguish "absent" from "query failed".
var ErrNotFound = errors.New("repository: record not found")

// AuthRepo is the persistence boundary for authentication and accounts.
type AuthRepo interface {
	// CreateUser inserts a user and cascades any nested Accounts in one statement.
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetAccount(ctx context.Context, userID uuid.UUID, mode models.AccountMode) (*models.Account, error)
	UpdateLastLogin(ctx context.Context, userID uuid.UUID, at time.Time) error

	SaveRefreshToken(ctx context.Context, token *models.RefreshToken) error
	GetRefreshTokenByHash(ctx context.Context, hash string) (*models.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, id uuid.UUID, at time.Time) error
}

type authRepo struct {
	DB *gorm.DB
}

func NewAuthRepo(db *gorm.DB) AuthRepo {
	return &authRepo{DB: db}
}

func (a *authRepo) CreateUser(ctx context.Context, user *models.User) error {
	return a.DB.WithContext(ctx).Create(user).Error
}

func (a *authRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := a.DB.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		utils.LogError("repo: get user by email", err)
		return nil, err
	}
	return &user, nil
}

func (a *authRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	err := a.DB.WithContext(ctx).First(&user, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		utils.LogError("repo: get user by id", err)
		return nil, err
	}
	return &user, nil
}

func (a *authRepo) GetAccount(ctx context.Context, userID uuid.UUID, mode models.AccountMode) (*models.Account, error) {
	var acct models.Account
	err := a.DB.WithContext(ctx).
		Where("user_id = ? AND mode = ?", userID, mode).
		First(&acct).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		utils.LogError("repo: get account", err)
		return nil, err
	}
	return &acct, nil
}

func (a *authRepo) UpdateLastLogin(ctx context.Context, userID uuid.UUID, at time.Time) error {
	return a.DB.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Update("last_login_at", at).Error
}

func (a *authRepo) SaveRefreshToken(ctx context.Context, token *models.RefreshToken) error {
	return a.DB.WithContext(ctx).Create(token).Error
}

func (a *authRepo) GetRefreshTokenByHash(ctx context.Context, hash string) (*models.RefreshToken, error) {
	var rt models.RefreshToken
	err := a.DB.WithContext(ctx).Where("token_hash = ?", hash).First(&rt).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		utils.LogError("repo: get refresh token", err)
		return nil, err
	}
	return &rt, nil
}

func (a *authRepo) RevokeRefreshToken(ctx context.Context, id uuid.UUID, at time.Time) error {
	return a.DB.WithContext(ctx).
		Model(&models.RefreshToken{}).
		Where("id = ?", id).
		Update("revoked_at", at).Error
}
