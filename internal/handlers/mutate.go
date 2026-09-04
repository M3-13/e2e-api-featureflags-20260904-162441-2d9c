package handlers

import (
	"encoding/json"
	"errors"
	"mime"
	"net/http"

	"featureflag-api/internal/store"
)

const maxBodyBytes = 1 << 20

func handleUpdate(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || mediaType != "application/json" {
			writeError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
			return
		}

		key := r.PathValue("key")

		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
		var body struct {
			Enabled        *bool   `json:"enabled"`
			Description    *string `json:"description"`
			RolloutPercent *int    `json:"rollout_percent"`
		}
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&body); err != nil {
			var mbe *http.MaxBytesError
			if errors.As(err, &mbe) {
				writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
				return
			}
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if dec.More() {
			writeError(w, http.StatusBadRequest, "invalid JSON: trailing data")
			return
		}

		if body.RolloutPercent != nil && (*body.RolloutPercent < 0 || *body.RolloutPercent > 100) {
			writeError(w, http.StatusBadRequest, "rollout_percent must be between 0 and 100")
			return
		}

		existing, ok := s.Get(key)
		if !ok {
			writeError(w, http.StatusNotFound, "flag not found")
			return
		}

		if body.Enabled != nil {
			existing.Enabled = *body.Enabled
		}
		if body.Description != nil {
			existing.Description = *body.Description
		}
		if body.RolloutPercent != nil {
			existing.RolloutPercent = *body.RolloutPercent
		}

		if !s.Update(key, existing) {
			writeError(w, http.StatusNotFound, "flag not found")
			return
		}

		writeJSON(w, http.StatusOK, existing)
	}
}

func handleDelete(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")
		if !s.Delete(key) {
			writeError(w, http.StatusNotFound, "flag not found")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
