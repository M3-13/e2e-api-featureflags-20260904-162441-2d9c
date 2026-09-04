package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"unsafe"

	"featureflag-api/internal/model"
	"featureflag-api/internal/store"
)

// seedStore builds a store pre-populated with the given flags. The store
// package does not yet expose a public way to insert flags (its Create/Get
// live in another ticket), so we write directly into its private map.
func seedStore(flags map[string]model.Flag) *store.Store {
	s := store.New()
	v := reflect.ValueOf(s).Elem()
	f := v.FieldByName("flags")
	f = reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
	f.Set(reflect.ValueOf(flags))
	return s
}

type evaluateResponse struct {
	Key     string `json:"key"`
	Enabled bool   `json:"enabled"`
	Result  bool   `json:"result"`
}

func doEvaluate(s *store.Store, target string) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	Register(mux, s)

	req := httptest.NewRequest(http.MethodGet, target, nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func TestEvaluateDeterministic(t *testing.T) {
	s := seedStore(map[string]model.Flag{
		"feature-x": {Key: "feature-x", Enabled: true, RolloutPercent: 50},
	})

	first := doEvaluate(s, "/flags/feature-x/evaluate?user=alice")
	if first.Code != http.StatusOK {
		t.Fatalf("first request: expected 200, got %d", first.Code)
	}

	var firstBody evaluateResponse
	if err := json.NewDecoder(first.Body).Decode(&firstBody); err != nil {
		t.Fatalf("decode first response: %v", err)
	}

	for i := 0; i < 20; i++ {
		rec := doEvaluate(s, "/flags/feature-x/evaluate?user=alice")
		if rec.Code != http.StatusOK {
			t.Fatalf("repeat request %d: expected 200, got %d", i, rec.Code)
		}
		var body evaluateResponse
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("decode repeat response %d: %v", i, err)
		}
		if body.Result != firstBody.Result {
			t.Fatalf("result not deterministic for same user: got %v then %v", firstBody.Result, body.Result)
		}
	}
}

func TestEvaluateDisabledReturnsFalse(t *testing.T) {
	s := seedStore(map[string]model.Flag{
		"feature-x": {Key: "feature-x", Enabled: false, RolloutPercent: 100},
	})

	rec := doEvaluate(s, "/flags/feature-x/evaluate?user=alice")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body evaluateResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Result {
		t.Fatalf("enabled=false: expected result false, got true")
	}
}

func TestEvaluateFullRolloutReturnsTrue(t *testing.T) {
	s := seedStore(map[string]model.Flag{
		"feature-x": {Key: "feature-x", Enabled: true, RolloutPercent: 100},
	})

	rec := doEvaluate(s, "/flags/feature-x/evaluate?user=alice")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body evaluateResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !body.Result {
		t.Fatalf("enabled=true, rollout_percent=100: expected result true, got false")
	}
	if body.Key != "feature-x" {
		t.Fatalf("expected key %q, got %q", "feature-x", body.Key)
	}
	if !body.Enabled {
		t.Fatalf("expected enabled true, got false")
	}
}

func TestEvaluateMissingUserReturns400(t *testing.T) {
	s := seedStore(map[string]model.Flag{
		"feature-x": {Key: "feature-x", Enabled: true, RolloutPercent: 100},
	})

	rec := doEvaluate(s, "/flags/feature-x/evaluate")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["error"] == "" {
		t.Fatalf("expected error field in body, got %v", body)
	}
}

func TestEvaluateEmptyUserReturns400(t *testing.T) {
	s := seedStore(map[string]model.Flag{
		"feature-x": {Key: "feature-x", Enabled: true, RolloutPercent: 100},
	})

	rec := doEvaluate(s, "/flags/feature-x/evaluate?user=")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["error"] == "" {
		t.Fatalf("expected error field in body, got %v", body)
	}
}

func TestEvaluateUnknownKeyReturns404(t *testing.T) {
	s := seedStore(map[string]model.Flag{
		"feature-x": {Key: "feature-x", Enabled: true, RolloutPercent: 100},
	})

	rec := doEvaluate(s, "/flags/unknown/evaluate?user=alice")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
