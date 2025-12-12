package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/user/jsbug/internal/session"
)

// CaptchaVerifier interface for captcha token verification
type CaptchaVerifier interface {
	Verify(ctx context.Context, token string, remoteIP string) (bool, error)
}

// AuthRequest represents the request body for POST /api/auth/captcha
type AuthRequest struct {
	CaptchaToken string `json:"captcha_token"`
}

// AuthResponse represents the response for POST /api/auth/captcha
type AuthResponse struct {
	SessionToken string `json:"session_token"`
	ExpiresAt    string `json:"expires_at"`
}

// AuthErrorResponse represents an error response
type AuthErrorResponse struct {
	Error AuthError `json:"error"`
}

// AuthError represents an error detail
type AuthError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// AuthHandler handles POST /api/auth/captcha requests
type AuthHandler struct {
	captchaVerifier CaptchaVerifier
	tokenManager    *session.TokenManager
	logger          *zap.Logger
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(captchaVerifier CaptchaVerifier, tokenManager *session.TokenManager, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		captchaVerifier: captchaVerifier,
		tokenManager:    tokenManager,
		logger:          logger,
	}
}

// ServeHTTP handles POST /api/auth/captcha requests
func (h *AuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	// Parse request body
	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid JSON request body")
		return
	}

	// Validate captcha token is present
	if req.CaptchaToken == "" {
		h.writeError(w, http.StatusBadRequest, "CAPTCHA_TOKEN_REQUIRED", "Captcha token is required")
		return
	}

	// Extract fingerprint from request headers
	fingerprint := session.HashFingerprint(
		r.Header.Get("User-Agent"),
		r.Header.Get("Accept-Language"),
		r.Header.Get("Accept-Encoding"),
	)

	// Get client IP for Turnstile verification
	clientIP := getClientIP(r)

	// Verify the Turnstile captcha token with Cloudflare
	valid, err := h.captchaVerifier.Verify(r.Context(), req.CaptchaToken, clientIP)
	if err != nil {
		h.logger.Warn("Captcha verification error", zap.Error(err))
		h.writeError(w, http.StatusServiceUnavailable, "CAPTCHA_SERVICE_UNAVAILABLE", "Captcha verification service temporarily unavailable")
		return
	}
	if !valid {
		h.writeError(w, http.StatusForbidden, "CAPTCHA_INVALID", "Captcha verification failed")
		return
	}

	// Generate JWT session token
	sessionToken, expiresAt, err := h.tokenManager.GenerateToken(fingerprint)
	if err != nil {
		h.logger.Error("Failed to generate session token", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "TOKEN_GENERATION_FAILED", "Failed to generate session token")
		return
	}

	// Return success response
	response := AuthResponse{
		SessionToken: sessionToken,
		ExpiresAt:    expiresAt.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to write response", zap.Error(err))
	}

	h.logger.Info("Session token issued",
		zap.String("client_ip", clientIP),
		zap.Time("expires_at", expiresAt),
	)
}

// writeError writes an error response
func (h *AuthHandler) writeError(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := AuthErrorResponse{
		Error: AuthError{
			Code:    code,
			Message: message,
		},
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to write error response", zap.Error(err))
	}
}
