package handlers

import (
	"net/http"

	"featureflag-api/internal/model"
	"featureflag-api/internal/store"
)

func handleList(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flags := s.List()
		if flags == nil {
			flags = []model.Flag{}
		}
		writeJSON(w, http.StatusOK, flags)
	}
}

func handleGet(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")
		f, ok := s.Get(key)
		if !ok {
			writeError(w, http.StatusNotFound, "flag not found")
			return
		}
		writeJSON(w, http.StatusOK, f)
	}
}
