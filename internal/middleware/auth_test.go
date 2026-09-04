package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestAuthHealthzWithoutKey(t *testing.T) {
	h := Auth(okHandler(), "secret")
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for /healthz without key, got %d", rec.Code)
	}
}

func TestAuthProtectedWithoutKey(t *testing.T) {
	h := Auth(okHandler(), "secret")
	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without key, got %d", rec.Code)
	}
}

func TestAuthProtectedWithCorrectKey(t *testing.T) {
	h := Auth(okHandler(), "secret")
	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	req.Header.Set("Authorization", "Bearer secret")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 with correct key, got %d", rec.Code)
	}
}

func TestAuthProtectedWithWrongKey(t *testing.T) {
	h := Auth(okHandler(), "secret")
	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	req.Header.Set("Authorization", "Bearer wrong")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 with wrong key, got %d", rec.Code)
	}
}
