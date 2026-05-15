package middlewares

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRecoveryMiddleware_NoPanic(t *testing.T) {
	handler := RecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestRecoveryMiddleware_WithPanic(t *testing.T) {
	handler := RecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something went wrong")
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "InternalError") {
		t.Errorf("expected InternalError in response body, got: %s", body)
	}
}

func TestRecoveryMiddleware_WithRequestID(t *testing.T) {
	const testRequestID = "test-req-123"

	handler := RecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("panic with request id")
	}))

	req := httptest.NewRequest(http.MethodPut, "/bucket/key", nil)
	ctx := context.WithValue(req.Context(), requestIDKey, testRequestID)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rr.Code)
	}

	// Also verify the response body contains InternalError when a request ID is present
	body := rr.Body.String()
	if !strings.Contains(body, "InternalError") {
		t.Errorf("expected InternalError in response body when request ID is set, got: %s", body)
	}
}

func TestRecoveryMiddleware_ContentType(t *testing.T) {
	handler := RecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(42)
	}))

	req := httptest.NewRequest(http.MethodDelete, "/bucket", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	ct := rr.Header().Get("Content-Type")
	if ct != "application/xml" {
		t.Errorf("expected Content-Type application/xml, got %s", ct)
	}
}
