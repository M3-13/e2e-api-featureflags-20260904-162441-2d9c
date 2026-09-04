package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"featureflag-api/internal/model"
	"featureflag-api/internal/store"
)

func newTestMux() (*http.ServeMux, *store.Store) {
	s := store.New()
	mux := http.NewServeMux()
	Register(mux, s)
	return mux, s
}

func TestListFlagsEmpty(t *testing.T) {
	mux, _ := newTestMux()

	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", ct)
	}

	var got []model.Flag
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if got == nil {
		t.Fatalf("expected empty JSON array, got null")
	}
	if len(got) != 0 {
		t.Fatalf("expected 0 flags, got %d", len(got))
	}
}

func TestListFlagsAfterCreate(t *testing.T) {
	mux, s := newTestMux()

	f := model.Flag{Key: "feature-x", Enabled: true, RolloutPercent: 50}
	if err := s.Create(f); err != nil {
		t.Fatalf("failed to create flag: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var got []model.Flag
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 flag, got %d", len(got))
	}
	if got[0].Key != "feature-x" {
		t.Fatalf("expected key %q, got %q", "feature-x", got[0].Key)
	}
}

func TestGetFlag(t *testing.T) {
	mux, s := newTestMux()

	f := model.Flag{Key: "feature-x", Enabled: true, Description: "cool feature", RolloutPercent: 100}
	if err := s.Create(f); err != nil {
		t.Fatalf("failed to create flag: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/flags/feature-x", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var got model.Flag
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if got.Key != "feature-x" {
		t.Fatalf("expected key %q, got %q", "feature-x", got.Key)
	}
	if !got.Enabled {
		t.Fatalf("expected enabled true, got false")
	}
	if got.RolloutPercent != 100 {
		t.Fatalf("expected rollout_percent 100, got %d", got.RolloutPercent)
	}
}

func TestGetFlagNotFound(t *testing.T) {
	mux, _ := newTestMux()

	req := httptest.NewRequest(http.MethodGet, "/flags/unknown", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", ct)
	}

	var got map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if _, ok := got["error"]; !ok {
		t.Fatalf("expected error field in response, got %v", got)
	}
}
