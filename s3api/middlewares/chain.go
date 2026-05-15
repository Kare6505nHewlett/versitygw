package middlewares

import "net/http"

// Chain applies a sequence of middleware functions to an http.Handler,
// executing them in the order they are provided (outermost first).
func Chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// DefaultMiddlewares returns the standard set of middlewares used by versitygw,
// applied in the recommended order: recovery → request ID → logging → CORS.
func DefaultMiddlewares() []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{
		RecoveryMiddleware,
		RequestIDMiddleware,
		LoggingMiddleware,
		CORSMiddleware(DefaultCORSConfig()),
	}
}

// DefaultMiddlewaresWithRateLimit returns the standard middleware stack with
// an additional rate limiting layer inserted before CORS handling.
// Note: if cfg is invalid, falls back to DefaultRateLimitConfig. Rate limiting
// is placed after logging so that rejected requests are still logged.
func DefaultMiddlewaresWithRateLimit(cfg RateLimitConfig) []func(http.Handler) http.Handler {
	if cfg.RequestsPerSecond <= 0 {
		cfg = DefaultRateLimitConfig()
	}
	return []func(http.Handler) http.Handler{
		RecoveryMiddleware,
		RequestIDMiddleware,
		LoggingMiddleware,
		RateLimitMiddleware(cfg),
		CORSMiddleware(DefaultCORSConfig()),
	}
}
