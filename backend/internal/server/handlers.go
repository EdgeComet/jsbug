package server

import (
	"encoding/json"
	"net/http"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status        string `json:"status"`
	Chrome        string `json:"chrome"`
	UptimeSeconds int64  `json:"uptime_seconds"`
}

// Chrome status values
const (
	chromeStatusUnknown     = "unknown"
	chromeStatusAvailable   = "available"
	chromeStatusUnavailable = "unavailable"
)

// healthHandler returns the server health status
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := HealthResponse{
		Status:        "healthy",
		Chrome:        chromeStatusUnknown,
		UptimeSeconds: s.Uptime(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
