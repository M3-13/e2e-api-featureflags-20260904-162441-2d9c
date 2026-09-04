package handlers

import (
	"net/http"

	"featureflag-api/internal/rollout"
	"featureflag-api/internal/store"
)

func handleEvaluate(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.URL.Query().Get("user")
		if user == "" {
			writeError(w, http.StatusBadRequest, "user parameter is required")
			return
		}

		key := r.PathValue("key")
		flag, ok := s.Get(key)
		if !ok {
			writeError(w, http.StatusNotFound, "flag not found")
			return
		}

		result := rollout.Evaluate(key, user, flag.RolloutPercent, flag.Enabled)

		writeJSON(w, http.StatusOK, map[string]any{
			"key":     key,
			"enabled": flag.Enabled,
			"result":  result,
		})
	}
}
