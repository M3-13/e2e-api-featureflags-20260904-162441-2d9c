package handlers

import (
	"net/http"

	"featureflag-api/internal/store"
)

func handleEvaluate(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotImplemented, "not implemented")
	}
}
