package store

import (
	"errors"
	"sync"
	"testing"

	"featureflag-api/internal/model"
)

func TestCreateAndGet(t *testing.T) {
	s := New()
	f := model.Flag{Key: "feature-x", Enabled: true, Description: "desc", RolloutPercent: 50}

	if err := s.Create(f); err != nil {
		t.Fatalf("Create returned unexpected error: %v", err)
	}

	got, ok := s.Get("feature-x")
	if !ok {
		t.Fatal("Get returned ok=false for an existing key")
	}
	if got != f {
		t.Fatalf("Get returned %+v, want %+v", got, f)
	}
}

func TestCreateDuplicateReturnsErrKeyExists(t *testing.T) {
	s := New()
	f := model.Flag{Key: "feature-x", Enabled: true}

	if err := s.Create(f); err != nil {
		t.Fatalf("first Create returned unexpected error: %v", err)
	}
	if err := s.Create(f); !errors.Is(err, ErrKeyExists) {
		t.Fatalf("second Create returned %v, want ErrKeyExists", err)
	}
}

func TestGetUnknownKey(t *testing.T) {
	s := New()

	got, ok := s.Get("missing")
	if ok {
		t.Fatal("Get returned ok=true for an unknown key")
	}
	if got != (model.Flag{}) {
		t.Fatalf("Get returned non-zero flag %+v for an unknown key", got)
	}
}

func TestList(t *testing.T) {
	s := New()

	if got := s.List(); len(got) != 0 {
		t.Fatalf("List on empty store returned %d flags, want 0", len(got))
	}

	for _, f := range []model.Flag{
		{Key: "a", Enabled: true},
		{Key: "b", Enabled: false},
		{Key: "c", Enabled: true},
	} {
		if err := s.Create(f); err != nil {
			t.Fatalf("Create(%q) returned unexpected error: %v", f.Key, err)
		}
	}

	got := s.List()
	if len(got) != 3 {
		t.Fatalf("List returned %d flags, want 3", len(got))
	}

	seen := make(map[string]bool, len(got))
	for _, f := range got {
		seen[f.Key] = true
	}
	for _, key := range []string{"a", "b", "c"} {
		if !seen[key] {
			t.Fatalf("List did not include flag %q", key)
		}
	}
}

func TestUpdate(t *testing.T) {
	s := New()
	if err := s.Create(model.Flag{Key: "feature-x", Enabled: true, Description: "old", RolloutPercent: 10}); err != nil {
		t.Fatalf("Create returned unexpected error: %v", err)
	}

	updated := model.Flag{Enabled: false, Description: "new", RolloutPercent: 90}
	if !s.Update("feature-x", updated) {
		t.Fatal("Update returned false for an existing key")
	}

	got, ok := s.Get("feature-x")
	if !ok {
		t.Fatal("Get returned ok=false after Update")
	}
	if got.Enabled != false || got.Description != "new" || got.RolloutPercent != 90 {
		t.Fatalf("Update did not replace fields: %+v", got)
	}
	if got.Key != "feature-x" {
		t.Fatalf("Update changed the key to %q, want it preserved", got.Key)
	}
}

func TestUpdateUnknownKey(t *testing.T) {
	s := New()
	if s.Update("missing", model.Flag{Enabled: true}) {
		t.Fatal("Update returned true for an unknown key")
	}
}

func TestDelete(t *testing.T) {
	s := New()
	if err := s.Create(model.Flag{Key: "feature-x", Enabled: true}); err != nil {
		t.Fatalf("Create returned unexpected error: %v", err)
	}

	if !s.Delete("feature-x") {
		t.Fatal("Delete returned false for an existing key")
	}
	if _, ok := s.Get("feature-x"); ok {
		t.Fatal("Get returned ok=true after Delete")
	}
}

func TestDeleteUnknownKey(t *testing.T) {
	s := New()
	if s.Delete("missing") {
		t.Fatal("Delete returned true for an unknown key")
	}
}

func TestConcurrentAccess(t *testing.T) {
	s := New()

	const writers = 8
	const ops = 200

	var wg sync.WaitGroup
	wg.Add(writers)
	for w := 0; w < writers; w++ {
		go func(w int) {
			defer wg.Done()
			for i := 0; i < ops; i++ {
				key := "flag"
				f := model.Flag{Key: key, Enabled: i%2 == 0, RolloutPercent: i % 101}
				if err := s.Create(f); err != nil && !errors.Is(err, ErrKeyExists) {
					t.Errorf("Create returned unexpected error: %v", err)
					return
				}
				_ = s.Update(key, model.Flag{Enabled: true, RolloutPercent: 50})
				_, _ = s.Get(key)
				_ = s.List()
				_ = s.Delete(key)
			}
		}(w)
	}
	wg.Wait()
}
