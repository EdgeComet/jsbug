package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"

	"github.com/user/jsbug/internal/session"
)

// mockCaptchaVerifier is a mock implementation of captcha verification
type mockCaptchaVerifier struct {
	shouldSucceed bool
	shouldError   bool
}

func (m *mockCaptchaVerifier) Verify(ctx context.Context, token string, remoteIP string) (bool, error) {
	if m.shouldError {
		return false, context.DeadlineExceeded
	}
	return m.shouldSucceed, nil
}

func TestAuthHandler_ServeHTTP_Success(t *testing.T) {
	logger := zap.NewNop()
	tokenManager, err := session.NewTokenManager("test-secret-key-32-bytes-long!!!", logger)
	if err != nil {
		t.Fatalf("NewTokenManager failed: %v", err)
	}

	// Create a mock captcha verifier that always succeeds
	mockVerifier := &mockCaptchaVerifierWrapper{
		verifyFunc: func(ctx context.Context, token, remoteIP string) (bool, error) {
			return true, nil
		},
	}

	handler := &AuthHandler{
		captchaVerifier: mockVerifier,
		tokenManager:    tokenManager,
		logger:          logger,
	}

	// Create request
	body := AuthRequest{CaptchaToken: "valid-turnstile-token"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/captcha", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Accept-Language", "en-US")
	req.Header.Set("Accept-Encoding", "gzip")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp AuthResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.SessionToken == "" {
		t.Error("SessionToken should not be empty")
	}

	if resp.ExpiresAt == "" {
		t.Error("ExpiresAt should not be empty")
	}
}

func TestAuthHandler_ServeHTTP_MethodNotAllowed(t *testing.T) {
	logger := zap.NewNop()
	tokenManager, err := session.NewTokenManager("test-secret-key-32-bytes-long!!!", logger)
	if err != nil {
		t.Fatalf("NewTokenManager failed: %v", err)
	}

	handler := &AuthHandler{
		tokenManager: tokenManager,
		logger:       logger,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/auth/captcha", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestAuthHandler_ServeHTTP_MissingCaptchaToken(t *testing.T) {
	logger := zap.NewNop()
	tokenManager, err := session.NewTokenManager("test-secret-key-32-bytes-long!!!", logger)
	if err != nil {
		t.Fatalf("NewTokenManager failed: %v", err)
	}

	handler := &AuthHandler{
		tokenManager: tokenManager,
		logger:       logger,
	}

	body := AuthRequest{CaptchaToken: ""}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/captcha", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var resp AuthErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Error.Code != "CAPTCHA_TOKEN_REQUIRED" {
		t.Errorf("Expected error code CAPTCHA_TOKEN_REQUIRED, got %s", resp.Error.Code)
	}
}

func TestAuthHandler_ServeHTTP_InvalidCaptcha(t *testing.T) {
	logger := zap.NewNop()
	tokenManager, err := session.NewTokenManager("test-secret-key-32-bytes-long!!!", logger)
	if err != nil {
		t.Fatalf("NewTokenManager failed: %v", err)
	}

	// Create a mock captcha verifier that always fails
	mockVerifier := &mockCaptchaVerifierWrapper{
		verifyFunc: func(ctx context.Context, token, remoteIP string) (bool, error) {
			return false, nil
		},
	}

	handler := &AuthHandler{
		captchaVerifier: mockVerifier,
		tokenManager:    tokenManager,
		logger:          logger,
	}

	body := AuthRequest{CaptchaToken: "invalid-token"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/captcha", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
	}

	var resp AuthErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Error.Code != "CAPTCHA_INVALID" {
		t.Errorf("Expected error code CAPTCHA_INVALID, got %s", resp.Error.Code)
	}
}

func TestAuthHandler_ServeHTTP_CaptchaServiceError(t *testing.T) {
	logger := zap.NewNop()
	tokenManager, err := session.NewTokenManager("test-secret-key-32-bytes-long!!!", logger)
	if err != nil {
		t.Fatalf("NewTokenManager failed: %v", err)
	}

	// Create a mock captcha verifier that returns an error
	mockVerifier := &mockCaptchaVerifierWrapper{
		verifyFunc: func(ctx context.Context, token, remoteIP string) (bool, error) {
			return false, context.DeadlineExceeded
		},
	}

	handler := &AuthHandler{
		captchaVerifier: mockVerifier,
		tokenManager:    tokenManager,
		logger:          logger,
	}

	body := AuthRequest{CaptchaToken: "some-token"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/captcha", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
	}

	var resp AuthErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Error.Code != "CAPTCHA_SERVICE_UNAVAILABLE" {
		t.Errorf("Expected error code CAPTCHA_SERVICE_UNAVAILABLE, got %s", resp.Error.Code)
	}
}

func TestAuthHandler_ServeHTTP_InvalidJSON(t *testing.T) {
	logger := zap.NewNop()
	tokenManager, err := session.NewTokenManager("test-secret-key-32-bytes-long!!!", logger)
	if err != nil {
		t.Fatalf("NewTokenManager failed: %v", err)
	}

	handler := &AuthHandler{
		tokenManager: tokenManager,
		logger:       logger,
	}

	req := httptest.NewRequest(http.MethodPost, "/api/auth/captcha", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// mockCaptchaVerifierWrapper wraps a function to implement the captcha verifier interface
type mockCaptchaVerifierWrapper struct {
	verifyFunc func(ctx context.Context, token, remoteIP string) (bool, error)
}

func (m *mockCaptchaVerifierWrapper) Verify(ctx context.Context, token string, remoteIP string) (bool, error) {
	return m.verifyFunc(ctx, token, remoteIP)
}
