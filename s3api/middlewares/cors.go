package middlewares

import (
	"net/http"
	"strings"
)

// CORSConfig holds configuration for CORS middleware.
type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
	MaxAge         string
}

// DefaultCORSConfig returns a CORSConfig with sensible defaults for S3-compatible APIs.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "PUT", "POST", "DELETE", "HEAD", "OPTIONS"},
		AllowedHeaders: []string{"Authorization", "Content-Type", "Content-MD5", "x-amz-date", "x-amz-content-sha256", "x-amz-security-token"},
		MaxAge:         "86400",
	}
}

// CORSMiddleware adds CORS headers to responses based on the provided config.
func CORSMiddleware(cfg CORSConfig) func(http.Handler) http.Handler {
	allowedOrigins := make(map[string]struct{}, len(cfg.AllowedOrigins))
	for _, o := range cfg.AllowedOrigins {
		allowedOrigins[o] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}

			_, allowAll := allowedOrigins["*"]
			_, allowOrigin := allowedOrigins[origin]

			if allowAll {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if allowOrigin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Add("Vary", "Origin")
			} else {
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("Access-Control-Allow-Methods", strings.Join(cfg.AllowedMethods, ", "))
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(cfg.AllowedHeaders, ", "))
			w.Header().Set("Access-Control-Max-Age", cfg.MaxAge)

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
