package handlers

import (
	"encoding/json"
	"net/http"

	"featureflag-api/internal/store"
)

func Register(mux *http.ServeMux, s *store.Store) {
	mux.HandleFunc("POST /flags", handleCreate(s))
	mux.HandleFunc("GET /flags", handleList(s))
	mux.HandleFunc("GET /flags/{key}", handleGet(s))
	mux.HandleFunc("PUT /flags/{key}", handleUpdate(s))
	mux.HandleFunc("DELETE /flags/{key}", handleDelete(s))
	mux.HandleFunc("GET /flags/{key}/evaluate", handleEvaluate(s))
	mux.HandleFunc("GET /healthz", handleHealth)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
