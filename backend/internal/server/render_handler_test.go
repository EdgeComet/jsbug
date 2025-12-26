package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"

	"github.com/user/jsbug/internal/chrome"
	"github.com/user/jsbug/internal/config"
	"github.com/user/jsbug/internal/fetcher"
	"github.com/user/jsbug/internal/parser"
	"github.com/user/jsbug/internal/types"
)

// MockFetcher implements the Fetcher interface for testing
type MockFetcher struct {
	result *fetcher.FetchResult
	err    error
}

func (m *MockFetcher) Fetch(ctx context.Context, opts fetcher.FetchOptions) (*fetcher.FetchResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

func testConfig() *config.Config {
	return &config.Config{
		Chrome: config.ChromeConfig{},
	}
}

func TestNewRenderHandler(t *testing.T) {
	logger := zap.NewNop()
	cfg := testConfig()
	p := parser.NewParser()

	handler := NewRenderHandler(nil, nil, p, cfg, logger, nil, nil)

	if handler == nil {
		t.Fatal("NewRenderHandler() returned nil")
	}
	if handler.parser != p {
		t.Error("parser not set correctly")
	}
	if handler.config != cfg {
		t.Error("config not set correctly")
	}
}

func TestRenderHandler_InvalidMethod(t *testing.T) {
	logger := zap.NewNop()
	cfg := testConfig()
	p := parser.NewParser()
	handler := NewRenderHandler(nil, nil, p, cfg, logger, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/render", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestRenderHandler_InvalidJSON(t *testing.T) {
	logger := zap.NewNop()
	cfg := testConfig()
	p := parser.NewParser()
	handler := NewRenderHandler(nil, nil, p, cfg, logger, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/render", bytes.NewBufferString("invalid json"))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestRenderHandler_InvalidURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"empty URL", ""},
		{"invalid scheme", "ftp://example.com"},
		{"no host", "http://"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			cfg := testConfig()
			p := parser.NewParser()
			handler := NewRenderHandler(nil, nil, p, cfg, logger, nil, nil)

			body := map[string]interface{}{
				"url": tt.url,
			}
			jsonBody, _ := json.Marshal(body)

			req := httptest.NewRequest(http.MethodPost, "/api/render", bytes.NewBuffer(jsonBody))
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
			}

			var response types.RenderResponse
			json.NewDecoder(w.Body).Decode(&response)

			if response.Success {
				t.Error("expected success = false")
			}
			if response.Error == nil || response.Error.Code != types.ErrInvalidURL {
				t.Errorf("expected error code %s", types.ErrInvalidURL)
			}
		})
	}
}

func TestRenderHandler_InvalidTimeout(t *testing.T) {
	tests := []struct {
		name    string
		timeout int
	}{
		// Note: 0 is treated as "use default" so not tested here
		{"negative timeout", -1},
		{"timeout too high", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			cfg := testConfig()
			p := parser.NewParser()
			handler := NewRenderHandler(nil, nil, p, cfg, logger, nil, nil)

			body := map[string]interface{}{
				"url":     "https://example.com",
				"timeout": tt.timeout,
			}
			jsonBody, _ := json.Marshal(body)

			req := httptest.NewRequest(http.MethodPost, "/api/render", bytes.NewBuffer(jsonBody))
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
			}

			var response types.RenderResponse
			json.NewDecoder(w.Body).Decode(&response)

			if response.Error == nil || response.Error.Code != types.ErrInvalidTimeout {
				t.Errorf("expected error code %s", types.ErrInvalidTimeout)
			}
		})
	}
}

func TestRenderHandler_InvalidWaitEvent(t *testing.T) {
	logger := zap.NewNop()
	cfg := testConfig()
	p := parser.NewParser()
	handler := NewRenderHandler(nil, nil, p, cfg, logger, nil, nil)

	body := map[string]interface{}{
		"url":        "https://example.com",
		"wait_event": "invalid_event",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/render", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var response types.RenderResponse
	json.NewDecoder(w.Body).Decode(&response)

	if response.Error == nil || response.Error.Code != types.ErrInvalidWaitEvent {
		t.Errorf("expected error code %s", types.ErrInvalidWaitEvent)
	}
}

func TestRenderHandler_JSRender_PoolUnavailable(t *testing.T) {
	logger := zap.NewNop()
	cfg := testConfig()
	p := parser.NewParser()

	// Handler with nil pool
	handler := NewRenderHandler(nil, nil, p, cfg, logger, nil, nil)

	body := map[string]interface{}{
		"url":        "https://example.com",
		"js_enabled": true,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/render", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}

	var response types.RenderResponse
	json.NewDecoder(w.Body).Decode(&response)

	if response.Error == nil || response.Error.Code != types.ErrChromeUnavailable {
		t.Errorf("expected error code %s, got %v", types.ErrChromeUnavailable, response.Error)
	}
}

func TestRenderHandler_JSRender_PoolExhausted(t *testing.T) {
	logger := zap.NewNop()
	cfg := testConfig()
	p := parser.NewParser()

	// Create a mock pool that returns ErrNoInstanceAvailable
	pool := createMockExhaustedPool(logger)

	handler := NewRenderHandler(pool, nil, p, cfg, logger, nil, nil)

	body := map[string]interface{}{
		"url":        "https://example.com",
		"js_enabled": true,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/render", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}

	var response types.RenderResponse
	json.NewDecoder(w.Body).Decode(&response)

	if response.Error == nil || response.Error.Code != types.ErrPoolExhausted {
		t.Errorf("expected error code %s, got %v", types.ErrPoolExhausted, response.Error)
	}
}

func TestRenderHandler_JSRender_PoolShuttingDown(t *testing.T) {
	logger := zap.NewNop()
	cfg := testConfig()
	p := parser.NewParser()

	// Create a mock pool that returns ErrPoolShuttingDown
	pool := createMockShuttingDownPool(logger)

	handler := NewRenderHandler(pool, nil, p, cfg, logger, nil, nil)

	body := map[string]interface{}{
		"url":        "https://example.com",
		"js_enabled": true,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/render", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}

	var response types.RenderResponse
	json.NewDecoder(w.Body).Decode(&response)

	if response.Error == nil || response.Error.Code != types.ErrPoolShuttingDown {
		t.Errorf("expected error code %s, got %v", types.ErrPoolShuttingDown, response.Error)
	}
}

// createMockExhaustedPool creates a pool with no available instances
func createMockExhaustedPool(logger *zap.Logger) *chrome.ChromePool {
	return chrome.NewMockPool(logger, chrome.ErrNoInstanceAvailable)
}

// createMockShuttingDownPool creates a pool that is shutting down
func createMockShuttingDownPool(logger *zap.Logger) *chrome.ChromePool {
	return chrome.NewMockPool(logger, chrome.ErrPoolShuttingDown)
}

func TestRenderHandler_Fetch_Success(t *testing.T) {
	logger := zap.NewNop()
	cfg := testConfig()
	p := parser.NewParser()

	mockFetcher := &MockFetcher{
		result: &fetcher.FetchResult{
			HTML:          "<html><head><title>Fetch Page</title></head><body><h1>Content</h1></body></html>",
			FinalURL:      "https://example.com/",
			StatusCode:    200,
			PageSizeBytes: 80,
			FetchTime:     0.1,
			Headers:       http.Header{},
		},
	}

	handler := NewRenderHandler(nil, mockFetcher, p, cfg, logger, nil, nil)

	body := map[string]interface{}{
		"url":        "https://example.com",
		"js_enabled": false,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/render", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var response types.RenderResponse
	json.NewDecoder(w.Body).Decode(&response)

	if !response.Success {
		t.Errorf("expected success = true, got error: %v", response.Error)
	}
	if response.Data == nil {
		t.Fatal("expected data to be set")
	}
	if response.Data.Title != "Fetch Page" {
		t.Errorf("Title = %q, want %q", response.Data.Title, "Fetch Page")
	}
	if len(response.Data.H1) != 1 || response.Data.H1[0] != "Content" {
		t.Errorf("H1 = %v, want [Content]", response.Data.H1)
	}
}

func TestRenderHandler_Fetch_WithHeaders(t *testing.T) {
	logger := zap.NewNop()
	cfg := testConfig()
	p := parser.NewParser()

	headers := http.Header{}
	headers.Set("X-Robots-Tag", "noindex")
	headers.Set("Link", `<https://example.com/canonical>; rel="canonical"`)

	mockFetcher := &MockFetcher{
		result: &fetcher.FetchResult{
			HTML:          "<html><head><title>Test</title></head></html>",
			FinalURL:      "https://example.com/",
			StatusCode:    200,
			PageSizeBytes: 50,
			FetchTime:     0.05,
			Headers:       headers,
		},
	}

	handler := NewRenderHandler(nil, mockFetcher, p, cfg, logger, nil, nil)

	body := map[string]interface{}{
		"url":        "https://example.com",
		"js_enabled": false,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/render", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	var response types.RenderResponse
	json.NewDecoder(w.Body).Decode(&response)

	if response.Data.XRobotsTag != "noindex" {
		t.Errorf("XRobotsTag = %q, want %q", response.Data.XRobotsTag, "noindex")
	}
	if response.Data.CanonicalURL != "https://example.com/canonical" {
		t.Errorf("CanonicalURL = %q, want %q", response.Data.CanonicalURL, "https://example.com/canonical")
	}
}

func TestRenderHandler_Fetch_Timeout(t *testing.T) {
	logger := zap.NewNop()
	cfg := testConfig()
	p := parser.NewParser()

	mockFetcher := &MockFetcher{
		err: errors.New("Client.Timeout exceeded"),
	}

	handler := NewRenderHandler(nil, mockFetcher, p, cfg, logger, nil, nil)

	body := map[string]interface{}{
		"url":        "https://example.com",
		"js_enabled": false,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/render", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusRequestTimeout {
		t.Errorf("status = %d, want %d", w.Code, http.StatusRequestTimeout)
	}

	var response types.RenderResponse
	json.NewDecoder(w.Body).Decode(&response)

	if response.Error == nil || response.Error.Code != types.ErrRenderTimeout {
		t.Errorf("expected error code %s", types.ErrRenderTimeout)
	}
}

func TestRenderHandler_Fetch_Error(t *testing.T) {
	logger := zap.NewNop()
	cfg := testConfig()
	p := parser.NewParser()

	mockFetcher := &MockFetcher{
		err: errors.New("connection refused"),
	}

	handler := NewRenderHandler(nil, mockFetcher, p, cfg, logger, nil, nil)

	body := map[string]interface{}{
		"url":        "https://example.com",
		"js_enabled": false,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/render", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}

	var response types.RenderResponse
	json.NewDecoder(w.Body).Decode(&response)

	if response.Error == nil || response.Error.Code != types.ErrFetchFailed {
		t.Errorf("expected error code %s", types.ErrFetchFailed)
	}
}

func TestRenderHandler_ValidateRequest(t *testing.T) {
	logger := zap.NewNop()
	cfg := testConfig()
	p := parser.NewParser()
	handler := NewRenderHandler(nil, nil, p, cfg, logger, nil, nil)

	tests := []struct {
		name        string
		req         *types.RenderRequest
		expectError bool
		errorCode   string
	}{
		{
			name: "valid request",
			req: &types.RenderRequest{
				URL:       "https://example.com",
				Timeout:   15,
				WaitEvent: "load",
			},
			expectError: false,
		},
		{
			name: "valid request with defaults",
			req: &types.RenderRequest{
				URL:     "http://example.com/page",
				Timeout: 30,
			},
			expectError: false,
		},
		{
			name: "empty URL",
			req: &types.RenderRequest{
				URL:     "",
				Timeout: 15,
			},
			expectError: true,
			errorCode:   types.ErrInvalidURL,
		},
		{
			name: "invalid scheme",
			req: &types.RenderRequest{
				URL:     "file:///etc/passwd",
				Timeout: 15,
			},
			expectError: true,
			errorCode:   types.ErrInvalidURL,
		},
		{
			name: "timeout too low",
			req: &types.RenderRequest{
				URL:     "https://example.com",
				Timeout: 0,
			},
			expectError: true,
			errorCode:   types.ErrInvalidTimeout,
		},
		{
			name: "timeout too high",
			req: &types.RenderRequest{
				URL:     "https://example.com",
				Timeout: 61,
			},
			expectError: true,
			errorCode:   types.ErrInvalidTimeout,
		},
		{
			name: "invalid wait event",
			req: &types.RenderRequest{
				URL:       "https://example.com",
				Timeout:   15,
				WaitEvent: "invalid",
			},
			expectError: true,
			errorCode:   types.ErrInvalidWaitEvent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.validateRequest(tt.req)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				} else if err.Code != tt.errorCode {
					t.Errorf("error code = %s, want %s", err.Code, tt.errorCode)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestRenderHandler_UserAgentResolution(t *testing.T) {
	// Test that user agent presets are correctly resolved
	tests := []struct {
		preset   string
		expected string
	}{
		{"mobile", types.UserAgentPresets[types.UserAgentMobile]},
		{"chrome", types.UserAgentPresets[types.UserAgentChrome]},
		{"bot", types.UserAgentPresets[types.UserAgentBot]},
		{"custom-agent", "custom-agent"},
	}

	for _, tt := range tests {
		t.Run(tt.preset, func(t *testing.T) {
			resolved := types.ResolveUserAgent(tt.preset)
			if resolved != tt.expected {
				t.Errorf("ResolveUserAgent(%q) = %q, want %q", tt.preset, resolved, tt.expected)
			}
		})
	}
}

// Note: TestRenderHandler_BlockingOptions was removed as it requires real Chrome instances.
// Blocking functionality is tested via integration tests.
