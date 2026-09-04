package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLoggingSingleEntryPerRequest(t *testing.T) {
	var buf bytes.Buffer
	prev := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(prev)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	handler := Logging(next)

	req := httptest.NewRequest(http.MethodGet, "/flags/foo/evaluate?user=alice", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	got := buf.String()
	lines := strings.Split(strings.TrimSpace(got), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected exactly one log entry, got %d: %q", len(lines), got)
	}
}

func TestLoggingCorrectFields(t *testing.T) {
	var buf bytes.Buffer
	prev := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(prev)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})
	handler := Logging(next)

	req := httptest.NewRequest(http.MethodPost, "/flags", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	line := strings.TrimSpace(buf.String())
	if !strings.Contains(line, "POST") {
		t.Fatalf("log entry missing method: %q", line)
	}
	if !strings.Contains(line, "/flags") {
		t.Fatalf("log entry missing path: %q", line)
	}
	if !strings.Contains(line, "201") {
		t.Fatalf("log entry missing status code: %q", line)
	}
}

func TestLoggingNoQueryStringOrUser(t *testing.T) {
	var buf bytes.Buffer
	prev := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(prev)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := Logging(next)

	req := httptest.NewRequest(http.MethodGet, "/flags/secret/evaluate?user=hunter2", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	line := strings.TrimSpace(buf.String())
	if strings.Contains(line, "hunter2") {
		t.Fatalf("log entry leaks user value: %q", line)
	}
	if strings.Contains(line, "?") {
		t.Fatalf("log entry contains query string: %q", line)
	}
	if !strings.Contains(line, "/flags/secret/evaluate") {
		t.Fatalf("log entry missing path: %q", line)
	}
}

func TestLoggingStripsControlCharactersFromPath(t *testing.T) {
	var buf bytes.Buffer
	prev := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(prev)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := Logging(next)

	req := httptest.NewRequest(http.MethodGet, "/flags%0Afoo", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	got := buf.String()
	lines := strings.Split(strings.TrimSpace(got), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected exactly one log entry, got %d: %q", len(lines), got)
	}

	line := lines[0]
	if strings.Contains(line, "\n") || strings.Contains(line, "\r") {
		t.Fatalf("log entry contains newline/control character: %q", line)
	}
	for _, r := range req.URL.Path {
		if r < 0x20 || r == 0x7f {
			if strings.ContainsRune(line, r) {
				t.Fatalf("log entry still contains control rune %q: %q", r, line)
			}
		}
	}
}
