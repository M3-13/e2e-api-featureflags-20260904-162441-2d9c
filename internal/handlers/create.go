package handlers

import (
	"encoding/json"
	"errors"
	"mime"
	"net/http"

	"featureflag-api/internal/model"
	"featureflag-api/internal/store"
)

const maxCreateBodyBytes = 1 << 20

type createFlagRequest struct {
	Key            *string `json:"key"`
	Enabled        *bool   `json:"enabled"`
	Description    *string `json:"description"`
	RolloutPercent *int    `json:"rollout_percent"`
}

func handleCreate(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || mediaType != "application/json" {
			writeError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxCreateBodyBytes)

		var req createFlagRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			var maxBytesErr *http.MaxBytesError
			if errors.As(err, &maxBytesErr) {
				writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
				return
			}
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		if req.Key == nil || *req.Key == "" {
			writeError(w, http.StatusBadRequest, "key is required")
			return
		}
		if req.Enabled == nil {
			writeError(w, http.StatusBadRequest, "enabled is required")
			return
		}
		if req.RolloutPercent != nil && (*req.RolloutPercent < 0 || *req.RolloutPercent > 100) {
			writeError(w, http.StatusBadRequest, "rollout_percent must be between 0 and 100")
			return
		}

		flag := model.Flag{
			Key:     *req.Key,
			Enabled: *req.Enabled,
		}
		if req.Description != nil {
			flag.Description = *req.Description
		}
		if req.RolloutPercent != nil {
			flag.RolloutPercent = *req.RolloutPercent
		}

		if err := s.Create(flag); err != nil {
			if errors.Is(err, store.ErrKeyExists) {
				writeError(w, http.StatusConflict, "flag already exists")
				return
			}
			if errors.Is(err, store.ErrTooManyFlags) {
				writeError(w, http.StatusInsufficientStorage, "too many flags")
				return
			}
			writeError(w, http.StatusInternalServerError, "failed to create flag")
			return
		}

		writeJSON(w, http.StatusCreated, flag)
	}
}
