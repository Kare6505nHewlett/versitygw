// Copyright 2025 Versity Software
// This file is licensed under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance
// with the License. You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
// either express or implied. See the License for the specific
// language governing permissions and limitations under the License.

package middlewares

import (
	"context"
	"net/http"
	"time"
)

// TimeoutConfig holds the configuration for the timeout middleware.
type TimeoutConfig struct {
	// Timeout is the maximum duration for a request to complete.
	// Defaults to 30 seconds if zero.
	Timeout time.Duration
}

// DefaultTimeoutConfig returns a TimeoutConfig with sensible defaults.
func DefaultTimeoutConfig() TimeoutConfig {
	return TimeoutConfig{
		Timeout: 30 * time.Second,
	}
}

// TimeoutMiddleware cancels the request context after the configured duration,
// allowing handlers to respect context cancellation and abort long-running work.
func TimeoutMiddleware(cfg TimeoutConfig) func(http.Handler) http.Handler {
	if cfg.Timeout <= 0 {
		cfg.Timeout = DefaultTimeoutConfig().Timeout
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), cfg.Timeout)
			defer cancel()

			r = r.WithContext(ctx)
			done := make(chan struct{})

			go func() {
				defer close(done)
				next.ServeHTTP(w, r)
			}()

			select {
			case <-done:
				// handler completed in time
			case <-ctx.Done():
				http.Error(w, "request timeout", http.StatusGatewayTimeout)
			}
		})
	}
}
