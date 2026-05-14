package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthMiddleware_NoHeader(t *testing.T) {
	cfg := DefaultAuthConfig()
	mw := AuthMiddleware(cfg)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/bucket/key", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestAuthMiddleware_ValidHeader(t *testing.T) {
	cfg := DefaultAuthConfig()
	mw := AuthMiddleware(cfg)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/bucket/key", nil)
	req.Header.Set("Authorization", "AWS4-HMAC-SHA256 Credential=test")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestAuthMiddleware_SkipPath(t *testing.T) {
	cfg := DefaultAuthConfig()
	cfg.SkipPaths = []string{"/health"}
	mw := AuthMiddleware(cfg)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	// No Authorization header — should still pass due to skip.
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for skipped path, got %d", rec.Code)
	}
}

// TestAuthMiddleware_SkipPath_Nested verifies that skip path matching is
// exact and does not accidentally skip sub-paths like /healthz or /health/check.
func TestAuthMiddleware_SkipPath_Nested(t *testing.T) {
	cfg := DefaultAuthConfig()
	cfg.SkipPaths = []string{"/health"}
	mw := AuthMiddleware(cfg)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// /healthz should NOT be skipped — it is a different path.
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for /healthz (not in skip list), got %d", rec.Code)
	}
}

func TestAuthMiddleware_CustomValidator(t *testing.T) {
	cfg := AuthConfig{
		SkipPaths: []string{},
		Validator: func(authHeader string) bool {
			return authHeader == "Bearer secret-token"
		},
	}
	mw := AuthMiddleware(cfg)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Wrong token.
	req := httptest.NewRequest(http.MethodGet, "/bucket", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for wrong token, got %d", rec.Code)
	}

	// Correct token.
	req2 := httptest.NewRequest(http.MethodGet, "/bucket", nil)
	req2.Header.Set("Authorization", "Bearer secret-token")
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Errorf("expected 200 for correct token, got %d", rec2.Code)
	}
}
