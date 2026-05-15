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
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDefaultTimeoutConfig(t *testing.T) {
	cfg := DefaultTimeoutConfig()
	// Default timeout is 30s; adjust this if the default changes upstream.
	if cfg.Timeout != 30*time.Second {
		t.Errorf("expected 30s default timeout, got %v", cfg.Timeout)
	}
}

func TestTimeoutMiddleware_CompletesInTime(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := TimeoutMiddleware(TimeoutConfig{Timeout: 5 * time.Second})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	mw(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestTimeoutMiddleware_Exceeds(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sleep longer than the configured timeout to trigger a 504.
		time.Sleep(300 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	mw := TimeoutMiddleware(TimeoutConfig{Timeout: 50 * time.Millisecond})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	mw(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusGatewayTimeout {
		t.Errorf("expected 504, got %d", rec.Code)
	}
}

func TestTimeoutMiddleware_ZeroUsesDefault(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	// Zero timeout should fall back to 30s default — handler completes fine.
	mw := TimeoutMiddleware(TimeoutConfig{Timeout: 0})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	mw(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestTimeoutMiddleware_ContextCancelled(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
			// context was cancelled as expected
		case <-time.After(500 * time.Millisecond):
			t.Error("expected context to be cancelled before handler timeout")
		}
	})

	mw := TimeoutMiddleware(TimeoutConfig{Timeout: 50 * time.Millisecond})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	mw(handler).ServeHTTP(rec, req)
}
