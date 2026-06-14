package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Google OAuth2 / OpenID Connect endpoints.
const (
	googleAuthEndpoint  = "https://accounts.google.com/o/oauth2/v2/auth"
	googleTokenEndpoint = "https://oauth2.googleapis.com/token"
	googleUserInfo      = "https://openidconnect.googleapis.com/v1/userinfo"
)

// GoogleProvider implements the authorization-code flow against Google using the
// standard library only (no SDK): build consent URL -> exchange code for an
// access token -> read the userinfo endpoint.
type GoogleProvider struct {
	clientID     string
	clientSecret string
	http         *http.Client
}

func NewGoogleProvider(clientID, clientSecret string) *GoogleProvider {
	return &GoogleProvider{
		clientID:     clientID,
		clientSecret: clientSecret,
		http:         &http.Client{Timeout: 10 * time.Second},
	}
}

func (g *GoogleProvider) Name() string { return "google" }

func (g *GoogleProvider) AuthCodeURL(state, redirectURI string) string {
	q := url.Values{
		"client_id":     {g.clientID},
		"redirect_uri":  {redirectURI},
		"response_type": {"code"},
		"scope":         {"openid email profile"},
		"state":         {state},
		"access_type":   {"online"},
		"prompt":        {"select_account"},
	}
	return googleAuthEndpoint + "?" + q.Encode()
}

func (g *GoogleProvider) Exchange(ctx context.Context, code, redirectURI string) (Profile, error) {
	// 1. Exchange the authorization code for an access token.
	form := url.Values{
		"code":          {code},
		"client_id":     {g.clientID},
		"client_secret": {g.clientSecret},
		"redirect_uri":  {redirectURI},
		"grant_type":    {"authorization_code"},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, googleTokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return Profile{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := g.http.Do(req)
	if err != nil {
		return Profile{}, fmt.Errorf("google: token exchange: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Profile{}, fmt.Errorf("google: token exchange status %d", resp.StatusCode)
	}
	var tok struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return Profile{}, fmt.Errorf("google: decode token: %w", err)
	}
	if tok.AccessToken == "" {
		return Profile{}, fmt.Errorf("google: empty access token")
	}

	// 2. Fetch the user's profile.
	ureq, err := http.NewRequestWithContext(ctx, http.MethodGet, googleUserInfo, nil)
	if err != nil {
		return Profile{}, err
	}
	ureq.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	uresp, err := g.http.Do(ureq)
	if err != nil {
		return Profile{}, fmt.Errorf("google: userinfo: %w", err)
	}
	defer uresp.Body.Close()
	if uresp.StatusCode != http.StatusOK {
		return Profile{}, fmt.Errorf("google: userinfo status %d", uresp.StatusCode)
	}
	var info struct {
		Sub     string `json:"sub"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.NewDecoder(uresp.Body).Decode(&info); err != nil {
		return Profile{}, fmt.Errorf("google: decode userinfo: %w", err)
	}
	if info.Sub == "" || info.Email == "" {
		return Profile{}, fmt.Errorf("google: incomplete profile")
	}
	name := info.Name
	if name == "" {
		name = strings.Split(info.Email, "@")[0]
	}
	return Profile{
		ProviderUserID: info.Sub,
		Email:          info.Email,
		Name:           name,
		AvatarURL:      info.Picture,
	}, nil
}
