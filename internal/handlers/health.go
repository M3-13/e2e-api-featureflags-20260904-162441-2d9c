package handlers

import "net/http"

var ServiceVersion = "unknown"

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "version": ServiceVersion})
}
