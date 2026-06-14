package controllers_test

// End-to-end auth tests. These exercise the REAL stack — gin router, auth
// middleware, controllers, service, and token manager — via net/http/httptest
// (which runs the real router in-process). The ONLY substitution is the data
// layer: an in-memory AuthRepo so tests need no live Postgres. In production the
// same code runs against the real GORM repository.

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/auth"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/repository"
	"github.com/faisal-990/ProjectInvestApp/internal/web/controllers"
	"github.com/faisal-990/ProjectInvestApp/internal/web/middlewares"
	"github.com/faisal-990/ProjectInvestApp/internal/web/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ---- in-memory AuthRepo (the only fake; substitutes for Postgres) ----

type memAuthRepo struct {
	mu       sync.Mutex
	users    map[uuid.UUID]*models.User
	accounts map[uuid.UUID]*models.Account
	tokens   map[uuid.UUID]*models.RefreshToken
}

func newMemAuthRepo() *memAuthRepo {
	return &memAuthRepo{
		users:    map[uuid.UUID]*models.User{},
		accounts: map[uuid.UUID]*models.Account{},
		tokens:   map[uuid.UUID]*models.RefreshToken{},
	}
}

// CreateUser assigns ids and cascades nested accounts, mirroring GORM's Create.
func (m *memAuthRepo) CreateUser(_ context.Context, u *models.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	for i := range u.Accounts {
		if u.Accounts[i].ID == uuid.Nil {
			u.Accounts[i].ID = uuid.New()
		}
		u.Accounts[i].UserID = u.ID
		acct := u.Accounts[i]
		m.accounts[acct.ID] = &acct
	}
	stored := *u
	m.users[u.ID] = &stored
	return nil
}

func (m *memAuthRepo) GetUserByEmail(_ context.Context, email string) (*models.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, u := range m.users {
		if u.Email == email {
			cp := *u
			return &cp, nil
		}
	}
	return nil, repository.ErrNotFound
}

func (m *memAuthRepo) GetUserByID(_ context.Context, id uuid.UUID) (*models.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if u, ok := m.users[id]; ok {
		cp := *u
		return &cp, nil
	}
	return nil, repository.ErrNotFound
}

func (m *memAuthRepo) GetAccount(_ context.Context, userID uuid.UUID, mode models.AccountMode) (*models.Account, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, a := range m.accounts {
		if a.UserID == userID && a.Mode == mode {
			cp := *a
			return &cp, nil
		}
	}
	return nil, repository.ErrNotFound
}

func (m *memAuthRepo) UpdateLastLogin(_ context.Context, userID uuid.UUID, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if u, ok := m.users[userID]; ok {
		u.LastLoginAt = &at
	}
	return nil
}

func (m *memAuthRepo) SaveRefreshToken(_ context.Context, t *models.RefreshToken) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	stored := *t
	m.tokens[t.ID] = &stored
	return nil
}

func (m *memAuthRepo) GetRefreshTokenByHash(_ context.Context, hash string) (*models.RefreshToken, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, t := range m.tokens {
		if t.TokenHash == hash {
			cp := *t
			return &cp, nil
		}
	}
	return nil, repository.ErrNotFound
}

func (m *memAuthRepo) RevokeRefreshToken(_ context.Context, id uuid.UUID, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if t, ok := m.tokens[id]; ok {
		t.RevokedAt = &at
	}
	return nil
}

// ---- test harness: build the real router with the in-memory repo ----

func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	tm := auth.NewTokenManager("test-secret", 15*time.Minute, 24*time.Hour)
	svc := service.NewAuthService(newMemAuthRepo(), tm)
	h := controllers.NewAuthHandler(svc)
	mw := middlewares.AuthMiddleware(tm)

	r := gin.New()
	g := r.Group("/api/auth")
	{
		g.POST("/signup", h.HandleAuthSignup)
		g.POST("/login", h.HandleAuthLogin)
		g.POST("/refresh", h.HandleAuthRefresh)
		g.POST("/logout", h.HandleAuthLogout)
		g.GET("/me", mw, h.HandleAuthForMe)
	}
	return r
}

// do issues a request and returns status + decoded body envelope.
func do(t *testing.T, r *gin.Engine, method, path, bearer string, body any) (int, map[string]any) {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var decoded map[string]any
	if w.Body.Len() > 0 {
		if err := json.Unmarshal(w.Body.Bytes(), &decoded); err != nil {
			t.Fatalf("decode response (%s): %v", w.Body.String(), err)
		}
	}
	return w.Code, decoded
}

func dataOf(t *testing.T, body map[string]any) map[string]any {
	t.Helper()
	d, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatalf("response missing data envelope: %v", body)
	}
	return d
}

func errCodeOf(body map[string]any) string {
	if e, ok := body["error"].(map[string]any); ok {
		if code, ok := e["code"].(string); ok {
			return code
		}
	}
	return ""
}

// ---- the full happy path: signup -> /me -> refresh -> /me -> logout ----

func TestAuthFlow_HappyPath(t *testing.T) {
	r := newTestRouter()
	signup := map[string]string{"name": "Ada Lovelace", "email": "ada@example.com", "password": "supersecret1"}

	// Signup → 201 with a token pair and the default sim account established.
	status, body := do(t, r, http.MethodPost, "/api/auth/signup", "", signup)
	if status != http.StatusCreated {
		t.Fatalf("signup status = %d, want 201 (%v)", status, body)
	}
	d := dataOf(t, body)
	access, _ := d["access_token"].(string)
	refresh, _ := d["refresh_token"].(string)
	if access == "" || refresh == "" {
		t.Fatalf("signup did not return tokens: %v", d)
	}

	// /me with the access token → 200 and the right user.
	status, body = do(t, r, http.MethodGet, "/api/auth/me", access, nil)
	if status != http.StatusOK {
		t.Fatalf("/me status = %d, want 200 (%v)", status, body)
	}
	if got := dataOf(t, body)["email"]; got != "ada@example.com" {
		t.Errorf("/me email = %v, want ada@example.com", got)
	}

	// Refresh → 200 with a new pair.
	status, body = do(t, r, http.MethodPost, "/api/auth/refresh", "", map[string]string{"refresh_token": refresh})
	if status != http.StatusOK {
		t.Fatalf("refresh status = %d, want 200 (%v)", status, body)
	}
	newAccess, _ := dataOf(t, body)["access_token"].(string)
	if newAccess == "" {
		t.Fatal("refresh did not return a new access token")
	}

	// The new access token still works on /me.
	if status, _ := do(t, r, http.MethodGet, "/api/auth/me", newAccess, nil); status != http.StatusOK {
		t.Errorf("/me with refreshed token status = %d, want 200", status)
	}

	// Logout (revoke the latest refresh token) is accepted.
	newRefresh, _ := dataOf(t, body)["refresh_token"].(string)
	if status, _ := do(t, r, http.MethodPost, "/api/auth/logout", "", map[string]string{"refresh_token": newRefresh}); status != http.StatusOK {
		t.Errorf("logout status = %d, want 200", status)
	}
}

func TestAuthFlow_RefreshRotationInvalidatesOldToken(t *testing.T) {
	r := newTestRouter()
	_, body := do(t, r, http.MethodPost, "/api/auth/signup", "",
		map[string]string{"name": "Grace Hopper", "email": "grace@example.com", "password": "supersecret1"})
	refresh, _ := dataOf(t, body)["refresh_token"].(string)

	// First refresh succeeds.
	if status, _ := do(t, r, http.MethodPost, "/api/auth/refresh", "", map[string]string{"refresh_token": refresh}); status != http.StatusOK {
		t.Fatalf("first refresh status = %d, want 200", status)
	}
	// Reusing the now-rotated token must fail (single-use).
	status, body := do(t, r, http.MethodPost, "/api/auth/refresh", "", map[string]string{"refresh_token": refresh})
	if status != http.StatusUnauthorized {
		t.Fatalf("reused refresh status = %d, want 401 (%v)", status, body)
	}
}

func TestAuthFlow_SecurityEdges(t *testing.T) {
	r := newTestRouter()
	creds := map[string]string{"name": "Linus", "email": "linus@example.com", "password": "supersecret1"}
	do(t, r, http.MethodPost, "/api/auth/signup", "", creds)

	t.Run("duplicate signup is 409 conflict", func(t *testing.T) {
		status, body := do(t, r, http.MethodPost, "/api/auth/signup", "", creds)
		if status != http.StatusConflict {
			t.Fatalf("status = %d, want 409", status)
		}
		if errCodeOf(body) != "conflict" {
			t.Errorf("error code = %q, want conflict", errCodeOf(body))
		}
	})

	t.Run("login wrong password is 401", func(t *testing.T) {
		status, _ := do(t, r, http.MethodPost, "/api/auth/login", "",
			map[string]string{"email": "linus@example.com", "password": "wrongpassword"})
		if status != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", status)
		}
	})

	t.Run("login unknown email is 401 (no user enumeration)", func(t *testing.T) {
		status, _ := do(t, r, http.MethodPost, "/api/auth/login", "",
			map[string]string{"email": "nobody@example.com", "password": "supersecret1"})
		if status != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", status)
		}
	})

	t.Run("login correct password is 200", func(t *testing.T) {
		status, _ := do(t, r, http.MethodPost, "/api/auth/login", "",
			map[string]string{"email": "linus@example.com", "password": "supersecret1"})
		if status != http.StatusOK {
			t.Fatalf("status = %d, want 200", status)
		}
	})

	t.Run("/me without token is 401", func(t *testing.T) {
		if status, _ := do(t, r, http.MethodGet, "/api/auth/me", "", nil); status != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", status)
		}
	})

	t.Run("/me with garbage token is 401", func(t *testing.T) {
		if status, _ := do(t, r, http.MethodGet, "/api/auth/me", "not.a.real.jwt", nil); status != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", status)
		}
	})

	t.Run("signup with short password is 400 validation", func(t *testing.T) {
		status, body := do(t, r, http.MethodPost, "/api/auth/signup", "",
			map[string]string{"name": "X", "email": "x@example.com", "password": "short"})
		if status != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", status)
		}
		if errCodeOf(body) != "validation_error" {
			t.Errorf("error code = %q, want validation_error", errCodeOf(body))
		}
	})
}
