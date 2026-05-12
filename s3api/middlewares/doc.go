// Package middlewares provides HTTP middleware components for the versitygw
// S3-compatible API gateway.
//
// Available middleware:
//
//   - RequestIDMiddleware: generates a unique x-amz-request-id for every
//     inbound request, stores it in the request context, and echoes it back
//     as a response header. Downstream handlers can retrieve the ID with
//     GetRequestID(ctx).
//
// Usage:
//
//	router.Use(middlewares.RequestIDMiddleware)
//
// The request ID follows the same 32-character lowercase hex format used by
// AWS S3 (16 random bytes encoded as hexadecimal).
package middlewares
