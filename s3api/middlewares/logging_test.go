package middlewares

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestLoggingMiddleware_StatusCode(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	handler := LoggingMiddleware(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}
}

func TestLoggingMiddleware_DefaultStatus(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	handler := LoggingMiddleware(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// no explicit WriteHeader call
		_, _ = w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected default status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestLoggingMiddleware_WithRequestID(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Chain RequestIDMiddleware before LoggingMiddleware so request ID is in context.
	handler := RequestIDMiddleware(LoggingMiddleware(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := GetRequestID(r.Context())
		if id == "" {
			t.Error("expected non-empty request ID in handler")
		}
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest(http.MethodPut, "/bucket/key", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestResponseWriter_WriteHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := newResponseWriter(rec)

	rw.WriteHeader(http.StatusNotFound)

	if rw.statusCode != http.StatusNotFound {
		t.Errorf("expected captured status %d, got %d", http.StatusNotFound, rw.statusCode)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected underlying recorder status %d, got %d", http.StatusNotFound, rec.Code)
	}
}
