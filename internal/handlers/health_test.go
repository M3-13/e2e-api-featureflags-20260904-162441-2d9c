package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"featureflag-api/internal/store"
)

func TestHealth(t *testing.T) {
	s := store.New()
	mux := http.NewServeMux()
	Register(mux, s)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", ct)
	}

	var got map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if got["status"] != "ok" {
		t.Fatalf("expected status %q, got %q", "ok", got["status"])
	}
	if got["version"] == "" {
		t.Fatalf("expected a version field, got empty")
	}
}
