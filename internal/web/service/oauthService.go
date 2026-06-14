package service

import (
	"context"
	"errors"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/auth"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/oauth"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/repository"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/utils"
	"github.com/faisal-990/ProjectInvestApp/internal/web/dto"
	"github.com/faisal-990/ProjectInvestApp/internal/web/httpx"
)

// OAuthService logs users in through an external identity provider (Google, etc.)
// and issues the same session a password login would. It is provider-agnostic:
// adding Robinhood is a new oauth.Provider in the registry, no changes here.
type OAuthService interface {
	// AuthCodeURL returns the provider's consent-screen URL to redirect the user to.
	AuthCodeURL(provider, state, redirectURI string) (string, error)
	// Complete exchanges the returned code, finds-or-creates-and-links the user,
	// and issues a session.
	Complete(ctx context.Context, provider, code, redirectURI string) (*dto.AuthResponse, error)
}

type oauthService struct {
	repo      repository.AuthRepo
	tokens    *auth.TokenManager
	providers oauth.Registry
}

func NewOAuthService(repo repository.AuthRepo, tm *auth.TokenManager, providers oauth.Registry) OAuthService {
	return &oauthService{repo: repo, tokens: tm, providers: providers}
}

func (s *oauthService) AuthCodeURL(provider, state, redirectURI string) (string, error) {
	p, ok := s.providers.Get(provider)
	if !ok {
		return "", httpx.NotFound("unknown or disabled login provider")
	}
	return p.AuthCodeURL(state, redirectURI), nil
}

func (s *oauthService) Complete(ctx context.Context, provider, code, redirectURI string) (*dto.AuthResponse, error) {
	p, ok := s.providers.Get(provider)
	if !ok {
		return nil, httpx.NotFound("unknown or disabled login provider")
	}
	profile, err := p.Exchange(ctx, code, redirectURI)
	if err != nil {
		return nil, httpx.BadRequest("could not complete sign-in with " + provider).WithCause(err)
	}

	user, err := s.resolveUser(ctx, provider, profile)
	if err != nil {
		return nil, err
	}

	acct, err := s.repo.GetAccount(ctx, user.ID, models.ModeSim)
	if err != nil {
		return nil, httpx.Internal("could not load account").WithCause(err)
	}
	return issueSession(ctx, s.repo, s.tokens, user, acct)
}

// resolveUser maps an external profile to a local user in priority order:
//  1. an existing link for (provider, providerUserID) → that user;
//  2. an existing user with the same email → link the provider to it;
//  3. otherwise → create a new user (verified, random password) + sim account.
func (s *oauthService) resolveUser(ctx context.Context, provider string, profile oauth.Profile) (*models.User, error) {
	user, err := s.repo.GetUserByOAuth(ctx, provider, profile.ProviderUserID)
	switch {
	case err == nil:
		return user, nil
	case !errors.Is(err, repository.ErrNotFound):
		return nil, httpx.Internal("could not look up identity").WithCause(err)
	}

	// No link yet — match by email, else create.
	existing, err := s.repo.GetUserByEmail(ctx, profile.Email)
	switch {
	case err == nil:
		if !existing.IsActive {
			return nil, httpx.Forbidden("this account is disabled")
		}
		if linkErr := s.repo.LinkOAuthAccount(ctx, &models.OAuthAccount{
			UserID: existing.ID, Provider: provider, ProviderUserID: profile.ProviderUserID,
		}); linkErr != nil {
			return nil, httpx.Internal("could not link identity").WithCause(linkErr)
		}
		// Backfill an avatar if the local account doesn't have one.
		if existing.AvatarURL == "" && profile.AvatarURL != "" {
			if uErr := s.repo.UpdateUserProfile(ctx, existing.ID, existing.Name, existing.Bio, profile.AvatarURL); uErr != nil {
				utils.LogError("oauth: backfill avatar", uErr)
			}
		}
		return existing, nil
	case !errors.Is(err, repository.ErrNotFound):
		return nil, httpx.Internal("could not look up account").WithCause(err)
	}

	// Brand-new user. OAuth users never type a password, so set an unguessable
	// random one (they can use the provider or the reset flow).
	randomPw, _, err := auth.NewOpaqueToken()
	if err != nil {
		return nil, httpx.Internal("could not create account").WithCause(err)
	}
	hash, err := auth.HashPassword(randomPw)
	if err != nil {
		return nil, httpx.Internal("could not secure account").WithCause(err)
	}
	newUser := &models.User{
		Name:          profile.Name,
		Email:         profile.Email,
		Password:      hash,
		Role:          "user",
		IsActive:      true,
		EmailVerified: true, // the provider already verified the address
		AvatarURL:     profile.AvatarURL,
		Accounts: []models.Account{{
			Mode: models.ModeSim, Currency: "USD", Balance: models.StartingSimBalance, IsActive: true,
		}},
	}
	if err := s.repo.CreateUser(ctx, newUser); err != nil {
		return nil, httpx.Internal("could not create account").WithCause(err)
	}
	if err := s.repo.LinkOAuthAccount(ctx, &models.OAuthAccount{
		UserID: newUser.ID, Provider: provider, ProviderUserID: profile.ProviderUserID,
	}); err != nil {
		return nil, httpx.Internal("could not link identity").WithCause(err)
	}
	return newUser, nil
}
