package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"featureflag-api/internal/model"
	"featureflag-api/internal/store"
)

func postFlags(t *testing.T, body string, contentType string) *httptest.ResponseRecorder {
	t.Helper()
	s := store.New()
	mux := http.NewServeMux()
	Register(mux, s)

	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(body))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func decodeFlag(t *testing.T, rec *httptest.ResponseRecorder) model.Flag {
	t.Helper()
	var flag model.Flag
	if err := json.NewDecoder(rec.Body).Decode(&flag); err != nil {
		t.Fatalf("failed to decode flag response: %v", err)
	}
	return flag
}

func decodeErr(t *testing.T, rec *httptest.ResponseRecorder) map[string]string {
	t.Helper()
	var e map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&e); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	return e
}

func TestCreateFlag(t *testing.T) {
	rec := postFlags(t, `{"key":"feature-x","enabled":true,"description":"hello","rollout_percent":50}`, "application/json")

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d (body %s)", rec.Code, rec.Body.String())
	}

	flag := decodeFlag(t, rec)
	if flag.Key != "feature-x" {
		t.Fatalf("expected key %q, got %q", "feature-x", flag.Key)
	}
	if !flag.Enabled {
		t.Fatalf("expected enabled true, got false")
	}
	if flag.Description != "hello" {
		t.Fatalf("expected description %q, got %q", "hello", flag.Description)
	}
	if flag.RolloutPercent != 50 {
		t.Fatalf("expected rollout_percent 50, got %d", flag.RolloutPercent)
	}
}

func TestCreateFlagDuplicate(t *testing.T) {
	s := store.New()
	mux := http.NewServeMux()
	Register(mux, s)

	mk := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(`{"key":"dup","enabled":true}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		return rec
	}

	first := mk()
	if first.Code != http.StatusCreated {
		t.Fatalf("expected first create 201, got %d (body %s)", first.Code, first.Body.String())
	}

	second := mk()
	if second.Code != http.StatusConflict {
		t.Fatalf("expected duplicate 409, got %d (body %s)", second.Code, second.Body.String())
	}
	if e := decodeErr(t, second); e["error"] == "" {
		t.Fatalf("expected error object, got %v", e)
	}
}

func TestCreateFlagEmptyKey(t *testing.T) {
	rec := postFlags(t, `{"key":"","enabled":true}`, "application/json")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateFlagMissingEnabled(t *testing.T) {
	rec := postFlags(t, `{"key":"feature-x"}`, "application/json")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateFlagRolloutOutOfRange(t *testing.T) {
	for _, v := range []int{-1, 101} {
		b, _ := json.Marshal(map[string]any{"key": "feature-x", "enabled": true, "rollout_percent": v})
		rec := postFlags(t, string(b), "application/json")
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("rollout_percent %d: expected 400, got %d", v, rec.Code)
		}
	}
}

func TestCreateFlagBodyTooLarge(t *testing.T) {
	big := `{"key":"` + strings.Repeat("a", maxCreateBodyBytes) + `","enabled":true}`
	rec := postFlags(t, big, "application/json")
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", rec.Code)
	}
	if e := decodeErr(t, rec); e["error"] == "" {
		t.Fatalf("expected error object, got %v", e)
	}
}

func TestCreateFlagWrongContentType(t *testing.T) {
	rec := postFlags(t, `{"key":"feature-x","enabled":true}`, "text/plain")
	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected 415, got %d", rec.Code)
	}
	if e := decodeErr(t, rec); e["error"] == "" {
		t.Fatalf("expected error object, got %v", e)
	}
}

func TestCreateFlagMissingContentType(t *testing.T) {
	rec := postFlags(t, `{"key":"feature-x","enabled":true}`, "")
	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected 415, got %d", rec.Code)
	}
}

func TestCreateFlagCharsetContentType(t *testing.T) {
	rec := postFlags(t, `{"key":"feature-x","enabled":true}`, "application/json; charset=utf-8")

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d (body %s)", rec.Code, rec.Body.String())
	}

	flag := decodeFlag(t, rec)
	if flag.Key != "feature-x" {
		t.Fatalf("expected key %q, got %q", "feature-x", flag.Key)
	}
}

func TestCreateFlagTooManyFlags(t *testing.T) {
	s := store.New()
	mux := http.NewServeMux()
	Register(mux, s)

	mk := func(key string) *httptest.ResponseRecorder {
		body := `{"key":"` + key + `","enabled":true}`
		req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		return rec
	}

	for i := 0; i < 1000; i++ {
		rec := mk(fmt.Sprintf("flag-%d", i))
		if rec.Code != http.StatusCreated {
			t.Fatalf("flag #%d: expected 201, got %d (body %s)", i, rec.Code, rec.Body.String())
		}
	}

	overflow := mk("overflow")
	if overflow.Code != http.StatusInsufficientStorage {
		t.Fatalf("expected 507, got %d (body %s)", overflow.Code, overflow.Body.String())
	}
	if e := decodeErr(t, overflow); e["error"] == "" {
		t.Fatalf("expected error object, got %v", e)
	}
}
