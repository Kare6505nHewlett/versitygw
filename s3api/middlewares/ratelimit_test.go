package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDefaultRateLimitConfig(t *testing.T) {
	cfg := DefaultRateLimitConfig()
	if cfg.RequestsPerSecond != 100 {
		t.Errorf("expected RequestsPerSecond=100, got %d", cfg.RequestsPerSecond)
	}
	// BurstSize should be 2x RequestsPerSecond by default
	if cfg.BurstSize != 200 {
		t.Errorf("expected BurstSize=200, got %d", cfg.BurstSize)
	}
}

func TestRateLimitMiddleware_AllowsRequests(t *testing.T) {
	cfg := RateLimitConfig{RequestsPerSecond: 10, BurstSize: 5}
	middleware := RateLimitMiddleware(cfg)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestRateLimitMiddleware_ThrottlesExcessRequests(t *testing.T) {
	cfg := RateLimitConfig{RequestsPerSecond: 1, BurstSize: 1}
	middleware := RateLimitMiddleware(cfg)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// First request should succeed (uses burst token)
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req)
	if rr1.Code != http.StatusOK {
		t.Errorf("expected first request to succeed, got %d", rr1.Code)
	}

	// Second immediate request should be throttled
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req)
	if rr2.Code != http.StatusTooManyRequests {
		t.Errorf("expected second request to be throttled (429), got %d", rr2.Code)
	}
}

func TestRateLimitMiddleware_RetryAfterHeader(t *testing.T) {
	cfg := RateLimitConfig{RequestsPerSecond: 1, BurstSize: 1}
	middleware := RateLimitMiddleware(cfg)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req) // consume burst

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Retry-After header is required by RFC 6585 for 429 responses
	if rr.Header().Get("Retry-After") == "" {
		t.Error("expected Retry-After header on throttled response")
	}
}

func TestRateLimitMiddleware_ZeroUsesDefault(t *testing.T) {
	cfg := RateLimitConfig{RequestsPerSecond: 0}
	middleware := RateLimitMiddleware(cfg)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200 with default config, got %d", rr.Code)
	}
}

func TestRateLimitMiddleware_RefillsTokens(t *testing.T) {
	cfg := RateLimitConfig{RequestsPerSecond: 100, BurstSize: 1}
	middleware := RateLimitMiddleware(cfg)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// Consume the single burst token
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req)
	if rr1.Code != http.StatusOK {
		t.Errorf("expected first request to succeed, got %d", rr1.Code)
	}

	// Wait long enough for the limiter to refill at least one token (100 rps = 10ms per token)
	time.Sleep(20 * time.Millisecond)

	// After refill, request should succeed again
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req)
	if rr2.Code != http.StatusOK {
		t.Errorf("expected request after refill to succeed, got %d", rr2.Code)
	}
}
