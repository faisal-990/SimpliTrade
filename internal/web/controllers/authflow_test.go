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
	resets   map[uuid.UUID]*models.PasswordReset
}

func newMemAuthRepo() *memAuthRepo {
	return &memAuthRepo{
		users:    map[uuid.UUID]*models.User{},
		accounts: map[uuid.UUID]*models.Account{},
		tokens:   map[uuid.UUID]*models.RefreshToken{},
		resets:   map[uuid.UUID]*models.PasswordReset{},
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

func (m *memAuthRepo) RevokeAllRefreshTokens(_ context.Context, userID uuid.UUID, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, t := range m.tokens {
		if t.UserID == userID && t.RevokedAt == nil {
			t.RevokedAt = &at
		}
	}
	return nil
}

func (m *memAuthRepo) CreatePasswordReset(_ context.Context, pr *models.PasswordReset) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if pr.ID == uuid.Nil {
		pr.ID = uuid.New()
	}
	cp := *pr
	m.resets[pr.ID] = &cp
	return nil
}

func (m *memAuthRepo) InvalidateUserPasswordResets(_ context.Context, userID uuid.UUID, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, pr := range m.resets {
		if pr.UserID == userID && pr.UsedAt == nil {
			pr.UsedAt = &at
		}
	}
	return nil
}

func (m *memAuthRepo) GetActivePasswordReset(_ context.Context, userID uuid.UUID) (*models.PasswordReset, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var latest *models.PasswordReset
	for _, pr := range m.resets {
		if pr.UserID == userID && pr.UsedAt == nil {
			if latest == nil || pr.CreatedAt.After(latest.CreatedAt) {
				latest = pr
			}
		}
	}
	if latest == nil {
		return nil, repository.ErrNotFound
	}
	cp := *latest
	return &cp, nil
}

func (m *memAuthRepo) IncrementPasswordResetAttempts(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if pr, ok := m.resets[id]; ok {
		pr.Attempts++
	}
	return nil
}

func (m *memAuthRepo) MarkPasswordResetUsed(_ context.Context, id uuid.UUID, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if pr, ok := m.resets[id]; ok {
		pr.UsedAt = &at
	}
	return nil
}

func (m *memAuthRepo) UpdateUserPassword(_ context.Context, userID uuid.UUID, passwordHash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if u, ok := m.users[userID]; ok {
		u.Password = passwordHash
	}
	return nil
}

// ---- test harness: build the real router with the in-memory repo ----

// memMailer captures the last reset code instead of sending email, so the reset
// flow is testable end-to-end without an email provider.
type memMailer struct {
	mu       sync.Mutex
	lastCode string
}

func (m *memMailer) SendPasswordResetCode(_ context.Context, _, code string, _ int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastCode = code
	return nil
}

func (m *memMailer) code() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastCode
}

func newTestRouter() (*gin.Engine, *memMailer) {
	gin.SetMode(gin.TestMode)
	tm := auth.NewTokenManager("test-secret", 15*time.Minute, 24*time.Hour)
	mm := &memMailer{}
	svc := service.NewAuthService(newMemAuthRepo(), tm, mm, "http://test.local")
	h := controllers.NewAuthHandler(svc)
	mw := middlewares.AuthMiddleware(tm)

	r := gin.New()
	g := r.Group("/api/auth")
	{
		g.POST("/signup", h.HandleAuthSignup)
		g.POST("/login", h.HandleAuthLogin)
		g.POST("/refresh", h.HandleAuthRefresh)
		g.POST("/logout", h.HandleAuthLogout)
		g.POST("/forgot-password", h.HandleForgotPassword)
		g.POST("/reset-password", h.HandleResetPassword)
		g.GET("/me", mw, h.HandleAuthForMe)
	}
	return r, mm
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

// ---- forgot/reset password: request link -> reset -> old pw fails, new works ----

func TestForgotResetPassword(t *testing.T) {
	r, mm := newTestRouter()
	const email = "reset@example.com"
	do(t, r, http.MethodPost, "/api/auth/signup", "",
		map[string]string{"name": "Reset User", "email": email, "password": "oldpassword1"})

	// Request a reset — always 200, link captured by the test mailer.
	if status, body := do(t, r, http.MethodPost, "/api/auth/forgot-password", "",
		map[string]string{"email": email}); status != http.StatusOK {
		t.Fatalf("forgot-password status = %d, want 200 (%v)", status, body)
	}
	code := mm.code()
	if len(code) != 6 {
		t.Fatalf("expected a 6-digit code, got %q", code)
	}

	// Unknown email still returns 200 (no account enumeration).
	if status, _ := do(t, r, http.MethodPost, "/api/auth/forgot-password", "",
		map[string]string{"email": "nobody@example.com"}); status != http.StatusOK {
		t.Errorf("forgot-password for unknown email status = %d, want 200", status)
	}

	// A wrong code is rejected.
	wrong := "000000"
	if wrong == code {
		wrong = "111111"
	}
	if status, _ := do(t, r, http.MethodPost, "/api/auth/reset-password", "",
		map[string]string{"email": email, "code": wrong, "password": "newpassword1"}); status != http.StatusBadRequest {
		t.Errorf("reset with wrong code status = %d, want 400", status)
	}

	// Correct code → 200.
	if status, body := do(t, r, http.MethodPost, "/api/auth/reset-password", "",
		map[string]string{"email": email, "code": code, "password": "newpassword1"}); status != http.StatusOK {
		t.Fatalf("reset-password status = %d, want 200 (%v)", status, body)
	}

	// Old password no longer works; new one does.
	if status, _ := do(t, r, http.MethodPost, "/api/auth/login", "",
		map[string]string{"email": email, "password": "oldpassword1"}); status != http.StatusUnauthorized {
		t.Errorf("login with old password status = %d, want 401", status)
	}
	if status, _ := do(t, r, http.MethodPost, "/api/auth/login", "",
		map[string]string{"email": email, "password": "newpassword1"}); status != http.StatusOK {
		t.Errorf("login with new password status = %d, want 200", status)
	}

	// The code is single-use — replay fails.
	if status, _ := do(t, r, http.MethodPost, "/api/auth/reset-password", "",
		map[string]string{"email": email, "code": code, "password": "anotherpass1"}); status != http.StatusBadRequest {
		t.Errorf("reused code status = %d, want 400", status)
	}
}

// ---- the full happy path: signup -> /me -> refresh -> /me -> logout ----

func TestAuthFlow_HappyPath(t *testing.T) {
	r, _ := newTestRouter()
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
	r, _ := newTestRouter()
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
	r, _ := newTestRouter()
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
