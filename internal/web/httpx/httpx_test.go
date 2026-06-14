package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() { gin.SetMode(gin.TestMode) }

// testCtx returns a gin context bound to a recorder for asserting responses.
func testCtx() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

func TestOK_Envelope(t *testing.T) {
	c, w := testCtx()
	OK(c, map[string]string{"hello": "world"})

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var body struct {
		Data map[string]string `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body.Data["hello"] != "world" {
		t.Errorf("data.hello = %q, want world", body.Data["hello"])
	}
}

func TestFail_AppErrorMapsStatusAndCode(t *testing.T) {
	tests := []struct {
		name       string
		err        *AppError
		wantStatus int
		wantCode   Code
	}{
		{"bad request", BadRequest("nope"), http.StatusBadRequest, CodeBadRequest},
		{"unauthorized", Unauthorized("no"), http.StatusUnauthorized, CodeUnauthorized},
		{"not found", NotFound("gone"), http.StatusNotFound, CodeNotFound},
		{"conflict", Conflict("dup"), http.StatusConflict, CodeConflict},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, w := testCtx()
			Fail(c, tc.err)

			if w.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tc.wantStatus)
			}
			var body errorEnvelope
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if body.Error.Code != tc.wantCode {
				t.Errorf("error.code = %q, want %q", body.Error.Code, tc.wantCode)
			}
		})
	}
}

func TestFail_UnknownErrorIsOpaque500(t *testing.T) {
	c, w := testCtx()
	Fail(c, errors.New("raw db driver detail that must not leak"))

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", w.Code)
	}
	if got := w.Body.String(); contains(got, "db driver") {
		t.Errorf("internal detail leaked to client: %s", got)
	}
	var body errorEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body.Error.Code != CodeInternal {
		t.Errorf("error.code = %q, want %q", body.Error.Code, CodeInternal)
	}
}

func TestAppError_WrapsCauseButHidesFromClient(t *testing.T) {
	cause := errors.New("sql: no rows")
	appErr := NotFound("user not found").WithCause(cause)

	// errors.Is sees through to the cause server-side.
	if !errors.Is(appErr, cause) {
		t.Error("errors.Is should find the wrapped cause")
	}
	// errors.As recovers the AppError from a wrapped chain.
	wrapped := fmt.Errorf("service layer: %w", appErr)
	var got *AppError
	if !errors.As(wrapped, &got) {
		t.Fatal("errors.As should recover AppError from wrapped chain")
	}
	if got.Code != CodeNotFound {
		t.Errorf("recovered code = %q, want %q", got.Code, CodeNotFound)
	}

	// But the client envelope never includes the cause string.
	c, w := testCtx()
	Fail(c, wrapped)
	if contains(w.Body.String(), "sql: no rows") {
		t.Errorf("wrapped cause leaked to client: %s", w.Body.String())
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
