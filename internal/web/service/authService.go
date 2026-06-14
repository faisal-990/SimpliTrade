package service

import (
	"context"
	"errors"
	"time"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/auth"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/mailer"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/repository"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/utils"
	"github.com/faisal-990/ProjectInvestApp/internal/web/dto"
	"github.com/faisal-990/ProjectInvestApp/internal/web/httpx"
	"github.com/google/uuid"
)

const (
	// passwordResetTTL is how long a reset code stays valid.
	passwordResetTTL = 10 * time.Minute
	// maxResetAttempts caps wrong-code guesses before the code is burned.
	maxResetAttempts = 5
)

// AuthService implements signup, login, token refresh/rotation, logout, and
// identity lookup. It returns httpx.AppError values so controllers can render
// them directly; unexpected failures are wrapped as opaque 500s.
type AuthService interface {
	Signup(ctx context.Context, req dto.SignupRequest) (*dto.AuthResponse, error)
	Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error)
	Refresh(ctx context.Context, refreshToken string) (*dto.AuthResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	Me(ctx context.Context, userID string) (*dto.UserDTO, error)
	UpdateProfile(ctx context.Context, userID string, req dto.UpdateProfileRequest) (*dto.UserDTO, error)
	// ForgotPassword emails a one-time code to the address if an active account
	// exists. It never reveals whether the email is registered (no enumeration).
	ForgotPassword(ctx context.Context, email string) error
	// ResetPassword verifies the emailed code for the address and sets a new
	// password.
	ResetPassword(ctx context.Context, email, code, newPassword string) error
}

type authservice struct {
	repo       repository.AuthRepo
	tokens     *auth.TokenManager
	mailer     mailer.Mailer
	appBaseURL string
}

func NewAuthService(r repository.AuthRepo, tm *auth.TokenManager, m mailer.Mailer, appBaseURL string) AuthService {
	return &authservice{repo: r, tokens: tm, mailer: m, appBaseURL: appBaseURL}
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
			Balance:  models.StartingSimBalance,
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

func (a *authservice) ForgotPassword(ctx context.Context, email string) error {
	user, err := a.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil // never reveal whether the email is registered
		}
		return httpx.Internal("could not process request").WithCause(err)
	}
	if !user.IsActive || user.IsBot {
		return nil // bots and disabled accounts can't reset
	}

	code, err := auth.NewOTP()
	if err != nil {
		return httpx.Internal("could not create reset code").WithCause(err)
	}
	codeHash, err := auth.HashPassword(code) // bcrypt — verified by comparison, never looked up
	if err != nil {
		return httpx.Internal("could not secure reset code").WithCause(err)
	}

	// Retire any outstanding code first, so only the newest one works.
	if err := a.repo.InvalidateUserPasswordResets(ctx, user.ID, time.Now()); err != nil {
		return httpx.Internal("could not create reset code").WithCause(err)
	}
	if err := a.repo.CreatePasswordReset(ctx, &models.PasswordReset{
		UserID:    user.ID,
		CodeHash:  codeHash,
		ExpiresAt: time.Now().Add(passwordResetTTL),
	}); err != nil {
		return httpx.Internal("could not create reset code").WithCause(err)
	}

	if err := a.mailer.SendPasswordResetCode(ctx, user.Email, code, int(passwordResetTTL.Minutes())); err != nil {
		// Best-effort: a delivery failure must not leak that the email exists.
		utils.LogError("auth: send reset code", err)
	}
	return nil
}

func (a *authservice) ResetPassword(ctx context.Context, email, code, newPassword string) error {
	// Generic error for every failure mode, so the endpoint reveals nothing about
	// which emails exist or whether a code is outstanding.
	const generic = "invalid or expired code"

	user, err := a.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return httpx.BadRequest(generic)
		}
		return httpx.Internal("could not reset password").WithCause(err)
	}

	pr, err := a.repo.GetActivePasswordReset(ctx, user.ID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return httpx.BadRequest(generic)
		}
		return httpx.Internal("could not reset password").WithCause(err)
	}
	if time.Now().After(pr.ExpiresAt) || pr.Attempts >= maxResetAttempts {
		_ = a.repo.MarkPasswordResetUsed(ctx, pr.ID, time.Now()) // burn it
		return httpx.BadRequest(generic)
	}
	if !auth.VerifyPassword(pr.CodeHash, code) {
		if err := a.repo.IncrementPasswordResetAttempts(ctx, pr.ID); err != nil {
			utils.LogError("auth: bump reset attempts", err)
		}
		return httpx.BadRequest(generic)
	}

	hash, err := auth.HashPassword(newPassword)
	if err != nil {
		return httpx.Internal("could not secure password").WithCause(err)
	}
	if err := a.repo.UpdateUserPassword(ctx, user.ID, hash); err != nil {
		return httpx.Internal("could not update password").WithCause(err)
	}
	// Single-use: consume the code so it can't be replayed.
	if err := a.repo.MarkPasswordResetUsed(ctx, pr.ID, time.Now()); err != nil {
		utils.LogError("auth: mark reset used", err)
	}
	// Security: kill every existing session so a pre-reset leak can't persist.
	if err := a.repo.RevokeAllRefreshTokens(ctx, user.ID, time.Now()); err != nil {
		utils.LogError("auth: revoke sessions after reset", err)
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

func (a *authservice) UpdateProfile(ctx context.Context, userID string, req dto.UpdateProfileRequest) (*dto.UserDTO, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, httpx.Unauthorized("invalid user identity")
	}
	if err := a.repo.UpdateUserProfile(ctx, id, req.Name, req.Bio, ""); err != nil {
		return nil, httpx.Internal("could not update profile").WithCause(err)
	}
	user, err := a.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, httpx.Internal("could not load profile").WithCause(err)
	}
	d := toUserDTO(user)
	return &d, nil
}

// issueTokens mints an access token + a rotating refresh token, persists the
// refresh token's hash, and assembles the auth response.
func (a *authservice) issueTokens(ctx context.Context, user *models.User, acct *models.Account) (*dto.AuthResponse, error) {
	return issueSession(ctx, a.repo, a.tokens, user, acct)
}

// issueSession is the shared token-minting path used by password login and OAuth
// login alike: an access token plus a rotating refresh token whose hash is
// persisted. Keeping it package-level lets every auth entry point issue
// identical sessions without duplicating the logic.
func issueSession(ctx context.Context, repo repository.AuthRepo, tokens *auth.TokenManager, user *models.User, acct *models.Account) (*dto.AuthResponse, error) {
	access, expiresAt, err := tokens.GenerateAccessToken(user.ID.String(), acct.ID.String(), user.Role)
	if err != nil {
		return nil, httpx.Internal("could not issue access token").WithCause(err)
	}
	refresh, err := tokens.GenerateRefreshToken()
	if err != nil {
		return nil, httpx.Internal("could not issue refresh token").WithCause(err)
	}
	if err := repo.SaveRefreshToken(ctx, &models.RefreshToken{
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
		AvatarURL:     u.AvatarURL,
		Bio:           u.Bio,
	}
}
