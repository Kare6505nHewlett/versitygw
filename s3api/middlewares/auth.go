package middlewares

import (
	"net/http"
	"strings"
)

// AuthConfig holds configuration for the authentication middleware.
type AuthConfig struct {
	// SkipPaths is a list of URL paths that bypass authentication.
	SkipPaths []string
	// Validator is a function that validates the Authorization header value.
	// It returns true if the request is authenticated.
	Validator func(authHeader string) bool
}

// DefaultAuthConfig returns an AuthConfig with sensible defaults.
// By default, no paths are skipped and all requests must have an
// Authorization header that is non-empty.
func DefaultAuthConfig() AuthConfig {
	return AuthConfig{
		SkipPaths: []string{},
		Validator: func(authHeader string) bool {
			return strings.TrimSpace(authHeader) != ""
		},
	}
}

// AuthMiddleware returns an HTTP middleware that validates the Authorization
// header of incoming requests using the provided AuthConfig.
// Requests to paths listed in cfg.SkipPaths bypass validation.
// Unauthenticated requests receive a 401 Unauthorized response.
func AuthMiddleware(cfg AuthConfig) func(http.Handler) http.Handler {
	skipSet := make(map[string]struct{}, len(cfg.SkipPaths))
	for _, p := range cfg.SkipPaths {
		skipSet[p] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, skip := skipSet[r.URL.Path]; skip {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if cfg.Validator == nil || !cfg.Validator(authHeader) {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
