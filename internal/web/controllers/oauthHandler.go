package controllers

import (
	"net/http"
	"net/url"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/auth"
	"github.com/faisal-990/ProjectInvestApp/internal/web/service"
	"github.com/gin-gonic/gin"
)

// oauthStateCookie holds the CSRF state between the login redirect and the
// callback (double-submit: the value in the cookie must match the one the
// provider echoes back).
const oauthStateCookie = "oauth_state"

type OAuthHandler struct {
	service    service.OAuthService
	apiBaseURL string // public base of THIS server, for the provider redirect_uri
	appBaseURL string // public base of the frontend, for the final redirect
}

func NewOAuthHandler(s service.OAuthService, apiBaseURL, appBaseURL string) *OAuthHandler {
	return &OAuthHandler{service: s, apiBaseURL: apiBaseURL, appBaseURL: appBaseURL}
}

func (h *OAuthHandler) redirectURI(provider string) string {
	return h.apiBaseURL + "/api/auth/oauth/" + provider + "/callback"
}

// HandleLogin starts the OAuth flow: set a CSRF state cookie and redirect to the
// provider's consent screen. GET /auth/oauth/:provider/login
func (h *OAuthHandler) HandleLogin(c *gin.Context) {
	provider := c.Param("provider")
	state, _, err := auth.NewOpaqueToken()
	if err != nil {
		c.Redirect(http.StatusFound, h.appBaseURL+"/login?error=oauth")
		return
	}
	authURL, err := h.service.AuthCodeURL(provider, state, h.redirectURI(provider))
	if err != nil {
		c.Redirect(http.StatusFound, h.appBaseURL+"/login?error=oauth_unavailable")
		return
	}
	// 10-minute, httpOnly, lax cookie scoped to the callback path's parent.
	c.SetCookie(oauthStateCookie, state, 600, "/api/auth/oauth", "", false, true)
	c.Redirect(http.StatusFound, authURL)
}

// HandleCallback completes the flow: verify state, exchange the code, and redirect
// to the frontend with the refresh token in the URL fragment (kept out of logs
// and the query string). GET /auth/oauth/:provider/callback
func (h *OAuthHandler) HandleCallback(c *gin.Context) {
	provider := c.Param("provider")

	if errParam := c.Query("error"); errParam != "" {
		// User denied consent or the provider errored.
		c.Redirect(http.StatusFound, h.appBaseURL+"/login?error=oauth_denied")
		return
	}

	state := c.Query("state")
	cookieState, err := c.Cookie(oauthStateCookie)
	c.SetCookie(oauthStateCookie, "", -1, "/api/auth/oauth", "", false, true) // clear it
	if err != nil || state == "" || state != cookieState {
		c.Redirect(http.StatusFound, h.appBaseURL+"/login?error=oauth_state")
		return
	}

	code := c.Query("code")
	if code == "" {
		c.Redirect(http.StatusFound, h.appBaseURL+"/login?error=oauth")
		return
	}

	resp, err := h.service.Complete(c.Request.Context(), provider, code, h.redirectURI(provider))
	if err != nil {
		c.Redirect(http.StatusFound, h.appBaseURL+"/login?error=oauth")
		return
	}

	// Hand the refresh token to the SPA via the fragment; it exchanges it for an
	// access token + profile through the normal refresh flow.
	target := h.appBaseURL + "/oauth/callback#refresh_token=" + url.QueryEscape(resp.RefreshToken)
	c.Redirect(http.StatusFound, target)
}
