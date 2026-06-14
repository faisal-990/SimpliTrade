package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() { gin.SetMode(gin.TestMode) }

func TestRequestID_SetOnResponse(t *testing.T) {
	r := gin.New()
	r.Use(RequestID())
	r.GET("/x", func(c *gin.Context) {
		if RequestIDOf(c) == "" {
			t.Error("request id should be set in context")
		}
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", nil))
	if w.Header().Get("X-Request-ID") == "" {
		t.Error("X-Request-ID response header should be set")
	}
}

func TestRequestID_HonorsInbound(t *testing.T) {
	r := gin.New()
	r.Use(RequestID())
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("X-Request-ID", "abc-123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if got := w.Header().Get("X-Request-ID"); got != "abc-123" {
		t.Errorf("X-Request-ID = %q, want abc-123 (inbound honored)", got)
	}
}

func TestRecovery_PanicBecomes500Envelope(t *testing.T) {
	r := gin.New()
	r.Use(Recovery())
	r.GET("/boom", func(_ *gin.Context) { panic("kaboom") })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/boom", nil))

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", w.Code)
	}
	if body := w.Body.String(); !contains(body, "internal_error") {
		t.Errorf("expected error envelope, got %s", body)
	}
	if contains(w.Body.String(), "kaboom") {
		t.Error("panic detail must not leak to the client")
	}
}

func TestSecurityHeaders(t *testing.T) {
	r := gin.New()
	r.Use(SecurityHeaders())
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", nil))
	if w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("X-Content-Type-Options should be nosniff")
	}
	if w.Header().Get("X-Frame-Options") != "DENY" {
		t.Error("X-Frame-Options should be DENY")
	}
}

func TestRateLimit_BlocksBurst(t *testing.T) {
	r := gin.New()
	r.Use(RateLimit(1, 1)) // 1 rps, burst 1
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	do := func() int {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		req.RemoteAddr = "203.0.113.7:5555" // same client IP for both
		r.ServeHTTP(w, req)
		return w.Code
	}

	if code := do(); code != http.StatusOK {
		t.Fatalf("first request = %d, want 200", code)
	}
	if code := do(); code != http.StatusTooManyRequests {
		t.Fatalf("second (burst) request = %d, want 429", code)
	}
}

func TestRateLimit_DisabledWhenZero(t *testing.T) {
	r := gin.New()
	r.Use(RateLimit(0, 0)) // disabled
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	for range 5 {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (rate limiting disabled)", w.Code)
		}
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
