package service

import (
	"context"
	"errors"
	"time"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/auth"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/repository"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/utils"
	"github.com/faisal-990/ProjectInvestApp/internal/web/dto"
	"github.com/faisal-990/ProjectInvestApp/internal/web/httpx"
	"github.com/google/uuid"
)

// startingSimBalance is the virtual cash every new user receives in their
// simulated account.
const startingSimBalance = 100000

// AuthService implements signup, login, token refresh/rotation, logout, and
// identity lookup. It returns httpx.AppError values so controllers can render
// them directly; unexpected failures are wrapped as opaque 500s.
type AuthService interface {
	Signup(ctx context.Context, req dto.SignupRequest) (*dto.AuthResponse, error)
	Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error)
	Refresh(ctx context.Context, refreshToken string) (*dto.AuthResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	Me(ctx context.Context, userID string) (*dto.UserDTO, error)
}

type authservice struct {
	repo   repository.AuthRepo
	tokens *auth.TokenManager
}

func NewAuthService(r repository.AuthRepo, tm *auth.TokenManager) AuthService {
	return &authservice{repo: r, tokens: tm}
}

func (a *authservice) Signup(ctx context.Context, req dto.SignupRequest) (*dto.AuthResponse, error) {
	// Reject duplicates up front for a clear 409 (a unique-index race still
	// protects us at the DB level).
	switch _, err := a.repo.GetUserByEmail(ctx, req.Email); {
	case err == nil:
		return nil, httpx.Conflict("an account with this email already exists")
	case !errors.Is(err, repository.ErrNotFound):
		return nil, httpx.Internal("could not verify email availability").WithCause(err)
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, httpx.Internal("could not secure password").WithCause(err)
	}

	// Create the user together with their default simulated account in one
	// cascading insert.
	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hash,
		Role:     "user",
		IsActive: true,
		Accounts: []models.Account{{
			Mode:     models.ModeSim,
			Currency: "USD",
			Balance:  startingSimBalance,
			IsActive: true,
		}},
	}
	if err := a.repo.CreateUser(ctx, user); err != nil {
		return nil, httpx.Internal("could not create account").WithCause(err)
	}

	return a.issueTokens(ctx, user, &user.Accounts[0])
}

func (a *authservice) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := a.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			// Same message as a bad password — never reveal which emails exist.
			return nil, httpx.Unauthorized("invalid email or password")
		}
		return nil, httpx.Internal("could not complete login").WithCause(err)
	}
	if !user.IsActive {
		return nil, httpx.Forbidden("this account is disabled")
	}
	if !auth.VerifyPassword(user.Password, req.Password) {
		return nil, httpx.Unauthorized("invalid email or password")
	}

	acct, err := a.repo.GetAccount(ctx, user.ID, models.ModeSim)
	if err != nil {
		return nil, httpx.Internal("could not load account").WithCause(err)
	}

	// Best-effort: a failed timestamp update must not block login.
	if err := a.repo.UpdateLastLogin(ctx, user.ID, time.Now()); err != nil {
		utils.LogError("auth: update last login", err)
	}

	return a.issueTokens(ctx, user, acct)
}

func (a *authservice) Refresh(ctx context.Context, refreshToken string) (*dto.AuthResponse, error) {
	rt, err := a.repo.GetRefreshTokenByHash(ctx, auth.HashToken(refreshToken))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, httpx.Unauthorized("invalid refresh token")
		}
		return nil, httpx.Internal("could not refresh session").WithCause(err)
	}
	if rt.RevokedAt != nil || time.Now().After(rt.ExpiresAt) {
		return nil, httpx.Unauthorized("refresh token expired or revoked")
	}

	user, err := a.repo.GetUserByID(ctx, rt.UserID)
	if err != nil {
		return nil, httpx.Internal("could not load user").WithCause(err)
	}
	acct, err := a.repo.GetAccount(ctx, user.ID, models.ModeSim)
	if err != nil {
		return nil, httpx.Internal("could not load account").WithCause(err)
	}

	// Rotation: revoke the presented token before issuing the next one, so a
	// stolen token is single-use.
	if err := a.repo.RevokeRefreshToken(ctx, rt.ID, time.Now()); err != nil {
		return nil, httpx.Internal("could not rotate refresh token").WithCause(err)
	}

	return a.issueTokens(ctx, user, acct)
}

func (a *authservice) Logout(ctx context.Context, refreshToken string) error {
	rt, err := a.repo.GetRefreshTokenByHash(ctx, auth.HashToken(refreshToken))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil // already gone — logout is idempotent
		}
		return httpx.Internal("could not log out").WithCause(err)
	}
	if rt.RevokedAt != nil {
		return nil
	}
	if err := a.repo.RevokeRefreshToken(ctx, rt.ID, time.Now()); err != nil {
		return httpx.Internal("could not log out").WithCause(err)
	}
	return nil
}

func (a *authservice) Me(ctx context.Context, userID string) (*dto.UserDTO, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, httpx.Unauthorized("invalid user identity")
	}
	user, err := a.repo.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, httpx.NotFound("user not found")
		}
		return nil, httpx.Internal("could not load user").WithCause(err)
	}
	d := toUserDTO(user)
	return &d, nil
}

// issueTokens mints an access token + a rotating refresh token, persists the
// refresh token's hash, and assembles the auth response.
func (a *authservice) issueTokens(ctx context.Context, user *models.User, acct *models.Account) (*dto.AuthResponse, error) {
	access, expiresAt, err := a.tokens.GenerateAccessToken(user.ID.String(), acct.ID.String(), user.Role)
	if err != nil {
		return nil, httpx.Internal("could not issue access token").WithCause(err)
	}
	refresh, err := a.tokens.GenerateRefreshToken()
	if err != nil {
		return nil, httpx.Internal("could not issue refresh token").WithCause(err)
	}
	if err := a.repo.SaveRefreshToken(ctx, &models.RefreshToken{
		UserID:    user.ID,
		TokenHash: refresh.Hash,
		ExpiresAt: refresh.ExpiresAt,
	}); err != nil {
		return nil, httpx.Internal("could not persist session").WithCause(err)
	}

	return &dto.AuthResponse{
		AccessToken:  access,
		RefreshToken: refresh.Plaintext,
		ExpiresAt:    expiresAt.Unix(),
		User:         toUserDTO(user),
	}, nil
}

func toUserDTO(u *models.User) dto.UserDTO {
	return dto.UserDTO{
		ID:            u.ID.String(),
		Name:          u.Name,
		Email:         u.Email,
		Role:          u.Role,
		EmailVerified: u.EmailVerified,
	}
}
