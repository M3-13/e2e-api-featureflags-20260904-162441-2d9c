package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"featureflag-api/internal/model"
	"featureflag-api/internal/store"
)

func newSeededTestMux(t *testing.T) (*store.Store, *http.ServeMux) {
	t.Helper()
	s := store.New()
	if err := s.Create(model.Flag{Key: "feature-x", Enabled: true, Description: "original", RolloutPercent: 50}); err != nil {
		t.Fatalf("seed flag: %v", err)
	}
	mux := http.NewServeMux()
	Register(mux, s)
	return s, mux
}

func doRequest(t *testing.T, mux *http.ServeMux, method, path, body, contentType string) *httptest.ResponseRecorder {
	t.Helper()
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, r)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	return rr
}

func TestUpdateFlagSuccess(t *testing.T) {
	_, mux := newSeededTestMux(t)
	rr := doRequest(t, mux, http.MethodPut, "/flags/feature-x",
		`{"enabled":false,"description":"updated","rollout_percent":75}`, "application/json")

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var got model.Flag
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Key != "feature-x" || got.Enabled != false || got.Description != "updated" || got.RolloutPercent != 75 {
		t.Fatalf("unexpected updated flag: %+v", got)
	}
}

func TestUpdateFlagPartial(t *testing.T) {
	s, mux := newSeededTestMux(t)
	rr := doRequest(t, mux, http.MethodPut, "/flags/feature-x",
		`{"description":"only-description"}`, "application/json")

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var got model.Flag
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Description != "only-description" {
		t.Fatalf("description not updated: %+v", got)
	}
	if got.Enabled != true || got.RolloutPercent != 50 {
		t.Fatalf("missing fields must not reset existing values: %+v", got)
	}
	if _, ok := s.Get("feature-x"); !ok {
		t.Fatal("flag should still exist")
	}
}

func TestUpdateFlagInvalidRollout(t *testing.T) {
	_, mux := newSeededTestMux(t)
	for _, v := range []string{"-1", "101"} {
		rr := doRequest(t, mux, http.MethodPut, "/flags/feature-x",
			`{"rollout_percent":`+v+`}`, "application/json")
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("rollout_percent=%s: expected 400, got %d", v, rr.Code)
		}
	}
}

func TestUpdateFlagUnknownKey(t *testing.T) {
	_, mux := newSeededTestMux(t)
	rr := doRequest(t, mux, http.MethodPut, "/flags/nope",
		`{"enabled":false}`, "application/json")
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestUpdateFlagTrailingGarbage(t *testing.T) {
	_, mux := newSeededTestMux(t)
	rr := doRequest(t, mux, http.MethodPut, "/flags/feature-x",
		`{"enabled":false} trailing-garbage`, "application/json")
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d (body %s)", rr.Code, rr.Body.String())
	}
	var e map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &e); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if e["error"] == "" {
		t.Fatalf("expected error object, got %v", e)
	}
}

func TestUpdateFlagTooLarge(t *testing.T) {
	_, mux := newSeededTestMux(t)
	big := strings.Repeat("a", 2<<20)
	rr := doRequest(t, mux, http.MethodPut, "/flags/feature-x",
		`{"description":"`+big+`"}`, "application/json")
	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", rr.Code)
	}
}

func TestUpdateFlagWrongContentType(t *testing.T) {
	_, mux := newSeededTestMux(t)
	rr := doRequest(t, mux, http.MethodPut, "/flags/feature-x",
		`{"enabled":false}`, "text/plain")
	if rr.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected 415, got %d", rr.Code)
	}
}

func TestUpdateFlagMissingContentType(t *testing.T) {
	_, mux := newSeededTestMux(t)
	rr := doRequest(t, mux, http.MethodPut, "/flags/feature-x",
		`{"enabled":false}`, "")
	if rr.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected 415, got %d", rr.Code)
	}
}

func TestDeleteFlagSuccess(t *testing.T) {
	s, mux := newSeededTestMux(t)
	rr := doRequest(t, mux, http.MethodDelete, "/flags/feature-x", "", "")
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
	if rr.Body.Len() != 0 {
		t.Fatalf("expected empty body, got %q", rr.Body.String())
	}
	if _, ok := s.Get("feature-x"); ok {
		t.Fatal("flag should have been deleted")
	}
}

func TestDeleteFlagSecondDelete(t *testing.T) {
	_, mux := newSeededTestMux(t)
	if rr := doRequest(t, mux, http.MethodDelete, "/flags/feature-x", "", ""); rr.Code != http.StatusNoContent {
		t.Fatalf("first delete: expected 204, got %d", rr.Code)
	}
	rr := doRequest(t, mux, http.MethodDelete, "/flags/feature-x", "", "")
	if rr.Code != http.StatusNotFound {
		t.Fatalf("second delete: expected 404, got %d", rr.Code)
	}
}

func TestDeleteFlagUnknownKey(t *testing.T) {
	_, mux := newSeededTestMux(t)
	rr := doRequest(t, mux, http.MethodDelete, "/flags/nope", "", "")
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}
