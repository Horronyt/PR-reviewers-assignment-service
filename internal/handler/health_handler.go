package handler

import (
	"encoding/json"
	"net/http"
)

// HealthHandler для health check
type HealthHandler struct{}

// NewHealthHandler создает handler для health check
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health обработчик GET /health
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "PR Reviewer Assignment Service",
	})
}

// Ready обработчик GET /ready для readiness probe
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ready": true,
	})
}
