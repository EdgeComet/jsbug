package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/user/jsbug/internal/screenshot"
)

// ScreenshotHandler handles screenshot retrieval requests
type ScreenshotHandler struct {
	store *screenshot.ScreenshotStore
}

// NewScreenshotHandler creates a new ScreenshotHandler
func NewScreenshotHandler(store *screenshot.ScreenshotStore) *ScreenshotHandler {
	return &ScreenshotHandler{
		store: store,
	}
}

// screenshotErrorResponse represents an error response for screenshot endpoints
type screenshotErrorResponse struct {
	Error string `json:"error"`
}

// HandleGet serves a screenshot by ID
func (h *ScreenshotHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Extract ID from URL path (last segment)
	// Expected path: /api/screenshot/{id}
	path := strings.TrimPrefix(r.URL.Path, "/api/screenshot/")
	id := strings.TrimSpace(path)

	if id == "" {
		h.writeError(w, http.StatusBadRequest, "missing screenshot ID")
		return
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid screenshot ID")
		return
	}

	// Retrieve screenshot from store
	data, found := h.store.Get(id)
	if !found {
		h.writeError(w, http.StatusNotFound, "screenshot not found or expired")
		return
	}

	// Serve the PNG image
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", "inline")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// writeError writes a JSON error response
func (h *ScreenshotHandler) writeError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(screenshotErrorResponse{Error: message})
}
