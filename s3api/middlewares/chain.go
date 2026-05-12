package middlewares

import "net/http"

// Middleware defines a standard middleware type that wraps an http.Handler.
type Middleware func(http.Handler) http.Handler

// Chain applies a list of middlewares to a handler in order, so that the
// first middleware in the list is the outermost (executed first).
//
// Example:
//
//	handler := Chain(myHandler, RecoveryMiddleware, RequestIDMiddleware, LoggingMiddleware)
func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	// Apply in reverse so the first middleware wraps the outermost layer.
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// DefaultMiddlewares returns the standard middleware stack used by versitygw:
// recovery → request ID → logging.
func DefaultMiddlewares() []Middleware {
	return []Middleware{
		RecoveryMiddleware,
		RequestIDMiddleware,
		LoggingMiddleware,
	}
}
