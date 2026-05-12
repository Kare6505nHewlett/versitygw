package middlewares

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

type contextKey string

const RequestIDKey contextKey = "requestID"

// GenerateRequestID creates a random 16-byte hex-encoded request ID.
func GenerateRequestID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// RequestIDMiddleware injects a unique request ID into each request context
// and sets the x-amz-request-id response header.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID, err := GenerateRequestID()
		if err != nil {
			// Fall back to a static value if random generation fails.
			reqID = "00000000000000000000000000000000"
		}

		ctx := context.WithValue(r.Context(), RequestIDKey, reqID)
		w.Header().Set("x-amz-request-id", reqID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID retrieves the request ID from the context.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}
