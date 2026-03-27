package middleware

import (
	"encoding/json"
	"net/http"
)

// Small shared helpers so middleware can return responses without depending on handlers.

func respondError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"data": nil, "error": msg})
}
