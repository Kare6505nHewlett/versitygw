package middlewares

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/rs/zerolog/log"
)

// RecoveryMiddleware recovers from panics, logs the stack trace, and returns
// a 500 Internal Server Error response to the client.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				stack := debug.Stack()
				requestID := GetRequestID(r.Context())

				log.Error().
					Str("request_id", requestID).
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Str("panic", fmt.Sprintf("%v", rec)).
					Str("stack", string(stack)).
					Msg("recovered from panic")

				w.Header().Set("Content-Type", "application/xml")
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><Error><Code>InternalError</Code><Message>We encountered an internal error. Please try again.</Message></Error>`))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
