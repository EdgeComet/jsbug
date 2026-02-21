package server

import (
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/user/jsbug/internal/robots"
	"github.com/user/jsbug/internal/types"
)

// RobotsHandler handles robots.txt check API requests
type RobotsHandler struct {
	checker *robots.Checker
	logger  *zap.Logger
}

// NewRobotsHandler creates a new RobotsHandler
func NewRobotsHandler(checker *robots.Checker, logger *zap.Logger) *RobotsHandler {
	return &RobotsHandler{
		checker: checker,
		logger:  logger,
	}
}

// ServeHTTP handles POST /api/robots requests
func (h *RobotsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, types.ErrInvalidURL, "Method not allowed")
		return
	}

	// Parse request body
	r.Body = http.MaxBytesReader(w, r.Body, 60<<10) // 60 KB
	var req types.RobotsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, types.ErrInvalidURL, "Invalid JSON request body")
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		if validationErr, ok := err.(*types.ValidationError); ok {
			h.writeError(w, http.StatusBadRequest, validationErr.Code, validationErr.Message)
		} else {
			h.logger.Warn("Unexpected validation error", zap.Error(err))
			h.writeError(w, http.StatusBadRequest, types.ErrInvalidURL, "Invalid request")
		}
		return
	}

	// Check robots.txt
	isAllowed, err := h.checker.Check(r.Context(), req.URL)
	if err != nil {
		// This shouldn't happen as checker fails open, but log it
		h.logger.Warn("Unexpected error from robots checker",
			zap.String("url", req.URL),
			zap.Error(err),
		)
	}

	// Build response
	response := &types.RobotsResponse{
		Success: true,
		Data: &types.RobotsData{
			URL:       req.URL,
			IsAllowed: isAllowed,
		},
	}

	h.writeJSON(w, response)

	h.logger.Info("Robots check request",
		zap.String("url", req.URL),
		zap.Bool("is_allowed", isAllowed),
		zap.Float64("total_time", time.Since(startTime).Seconds()),
	)
}

// writeJSON writes a JSON response
func (h *RobotsHandler) writeJSON(w http.ResponseWriter, response *types.RobotsResponse) {
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to write response", zap.Error(err))
	}
}

// writeError writes an error response
func (h *RobotsHandler) writeError(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := &types.RobotsResponse{
		Success: false,
		Error: &types.RenderError{
			Code:    code,
			Message: message,
		},
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to write error response", zap.Error(err))
	}
}
