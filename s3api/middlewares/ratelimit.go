package middlewares

import (
	"net/http"
	"sync"
	"time"
)

// RateLimitConfig holds configuration for the rate limiting middleware.
type RateLimitConfig struct {
	// RequestsPerSecond is the maximum number of requests allowed per second.
	RequestsPerSecond int
	// BurstSize is the maximum number of requests allowed in a burst.
	BurstSize int
}

// DefaultRateLimitConfig returns a RateLimitConfig with sensible defaults.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		RequestsPerSecond: 100,
		BurstSize:         200,
	}
}

type tokenBucket struct {
	tokens    float64
	maxTokens float64
	rate      float64
	lastCheck time.Time
	mu        sync.Mutex
}

func newTokenBucket(rate float64, burst float64) *tokenBucket {
	return &tokenBucket{
		tokens:    burst,
		maxTokens: burst,
		rate:      rate,
		lastCheck: time.Now(),
	}
}

func (tb *tokenBucket) allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastCheck).Seconds()
	tb.lastCheck = now

	tb.tokens += elapsed * tb.rate
	if tb.tokens > tb.maxTokens {
		tb.tokens = tb.maxTokens
	}

	if tb.tokens < 1 {
		return false
	}
	tb.tokens--
	return true
}

// RateLimitMiddleware returns an HTTP middleware that limits request rates
// using a token bucket algorithm. Requests exceeding the limit receive a
// 429 Too Many Requests response.
func RateLimitMiddleware(cfg RateLimitConfig) func(http.Handler) http.Handler {
	if cfg.RequestsPerSecond <= 0 {
		cfg = DefaultRateLimitConfig()
	}
	bucket := newTokenBucket(float64(cfg.RequestsPerSecond), float64(cfg.BurstSize))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !bucket.allow() {
				w.Header().Set("Content-Type", "application/xml")
				w.Header().Set("Retry-After", "1")
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><Error><Code>SlowDown</Code><Message>Please reduce your request rate.</Message></Error>`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
