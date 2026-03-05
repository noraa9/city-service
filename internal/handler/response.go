package handler

import (
	"encoding/json"
	"net/http"
)

// We keep response helpers in handler package so all HTTP endpoints return the same shape:
// - Success: { "data": <any>, "error": null }
// - Error:   { "data": null, "error": "human readable message" }

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"data": data, "error": nil})
}

func respondError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"data": nil, "error": msg})
}

