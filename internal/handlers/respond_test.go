package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSONSetsCacheControlNoStore(t *testing.T) {
	rec := httptest.NewRecorder()

	writeJSON(rec, http.StatusOK, map[string]string{"status": "ok"})

	if cc := rec.Header().Get("Cache-Control"); cc != "no-store" {
		t.Fatalf("expected Cache-Control no-store, got %q", cc)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func TestWriteErrorSetsCacheControlNoStore(t *testing.T) {
	rec := httptest.NewRecorder()

	writeError(rec, http.StatusNotFound, "not found")

	if cc := rec.Header().Get("Cache-Control"); cc != "no-store" {
		t.Fatalf("expected Cache-Control no-store, got %q", cc)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
}
