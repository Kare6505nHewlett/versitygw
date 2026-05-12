package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGenerateRequestID_Length(t *testing.T) {
	id, err := GenerateRequestID()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 16 bytes encoded as hex = 32 characters
	if len(id) != 32 {
		t.Errorf("expected request ID length 32, got %d", len(id))
	}
}

func TestGenerateRequestID_Uniqueness(t *testing.T) {
	ids := make(map[string]struct{}, 100)
	for i := 0; i < 100; i++ {
		id, err := GenerateRequestID()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, exists := ids[id]; exists {
			t.Errorf("duplicate request ID generated: %s", id)
		}
		ids[id] = struct{}{}
	}
}

func TestRequestIDMiddleware_HeaderSet(t *testing.T) {
	handler := RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	reqID := rr.Header().Get("x-amz-request-id")
	if len(reqID) != 32 {
		t.Errorf("expected x-amz-request-id of length 32, got %q", reqID)
	}
}

func TestRequestIDMiddleware_ContextValue(t *testing.T) {
	var capturedID string

	handler := RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if capturedID == "" {
		t.Error("expected non-empty request ID in context")
	}
	if capturedID != rr.Header().Get("x-amz-request-id") {
		t.Errorf("context ID %q does not match header ID %q", capturedID, rr.Header().Get("x-amz-request-id"))
	}
}

func TestGetRequestID_MissingKey(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if id := GetRequestID(req.Context()); id != "" {
		t.Errorf("expected empty string for missing key, got %q", id)
	}
}
