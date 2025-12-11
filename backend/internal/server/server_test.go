package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"

	"github.com/user/jsbug/internal/config"
)

func newTestServer(corsOrigins []string) *Server {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host:        "127.0.0.1",
			Port:        8080,
			Timeout:     30,
			CORSOrigins: corsOrigins,
		},
	}
	logger := zap.NewNop()
	return New(cfg, logger)
}

func TestHealthHandler_ReturnsOK(t *testing.T) {
	srv := newTestServer([]string{"*"})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	srv.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("health handler returned status %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestHealthHandler_ReturnsCorrectJSON(t *testing.T) {
	srv := newTestServer([]string{"*"})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	srv.mux.ServeHTTP(rec, req)

	var response HealthResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Status != "healthy" {
		t.Errorf("status = %q, want %q", response.Status, "healthy")
	}
	if response.Chrome != chromeStatusUnknown {
		t.Errorf("chrome = %q, want %q", response.Chrome, chromeStatusUnknown)
	}
	if response.UptimeSeconds < 0 {
		t.Errorf("uptime_seconds = %d, want >= 0", response.UptimeSeconds)
	}
}

func TestHealthHandler_ContentType(t *testing.T) {
	srv := newTestServer([]string{"*"})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	srv.mux.ServeHTTP(rec, req)

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Content-Type = %q, want %q", contentType, "application/json")
	}
}

func TestHealthHandler_MethodNotAllowed(t *testing.T) {
	srv := newTestServer([]string{"*"})

	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rec := httptest.NewRecorder()

	srv.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("health handler with POST returned status %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestCORS_AllowedOrigin(t *testing.T) {
	srv := newTestServer([]string{"http://localhost:3000"})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()

	srv.corsMiddleware(srv.mux).ServeHTTP(rec, req)

	allowOrigin := rec.Header().Get("Access-Control-Allow-Origin")
	if allowOrigin != "http://localhost:3000" {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", allowOrigin, "http://localhost:3000")
	}

	allowMethods := rec.Header().Get("Access-Control-Allow-Methods")
	if allowMethods != "GET, POST, OPTIONS" {
		t.Errorf("Access-Control-Allow-Methods = %q, want %q", allowMethods, "GET, POST, OPTIONS")
	}

	allowHeaders := rec.Header().Get("Access-Control-Allow-Headers")
	if allowHeaders != "Content-Type, X-Request-ID" {
		t.Errorf("Access-Control-Allow-Headers = %q, want %q", allowHeaders, "Content-Type, X-Request-ID")
	}
}

func TestCORS_WildcardOrigin(t *testing.T) {
	srv := newTestServer([]string{"*"})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("Origin", "http://any-origin.com")
	rec := httptest.NewRecorder()

	srv.corsMiddleware(srv.mux).ServeHTTP(rec, req)

	allowOrigin := rec.Header().Get("Access-Control-Allow-Origin")
	if allowOrigin != "http://any-origin.com" {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", allowOrigin, "http://any-origin.com")
	}
}

func TestCORS_DisallowedOrigin(t *testing.T) {
	srv := newTestServer([]string{"http://allowed.com"})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("Origin", "http://disallowed.com")
	rec := httptest.NewRecorder()

	srv.corsMiddleware(srv.mux).ServeHTTP(rec, req)

	allowOrigin := rec.Header().Get("Access-Control-Allow-Origin")
	if allowOrigin != "" {
		t.Errorf("Access-Control-Allow-Origin = %q, want empty for disallowed origin", allowOrigin)
	}
}

func TestCORS_NoOriginHeader(t *testing.T) {
	srv := newTestServer([]string{"*"})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	srv.corsMiddleware(srv.mux).ServeHTTP(rec, req)

	allowOrigin := rec.Header().Get("Access-Control-Allow-Origin")
	if allowOrigin != "" {
		t.Errorf("Access-Control-Allow-Origin = %q, want empty when no Origin header", allowOrigin)
	}
}

func TestCORS_PreflightRequest(t *testing.T) {
	srv := newTestServer([]string{"http://localhost:3000"})

	req := httptest.NewRequest(http.MethodOptions, "/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()

	srv.corsMiddleware(srv.mux).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("OPTIONS request returned status %d, want %d", rec.Code, http.StatusNoContent)
	}

	allowOrigin := rec.Header().Get("Access-Control-Allow-Origin")
	if allowOrigin != "http://localhost:3000" {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", allowOrigin, "http://localhost:3000")
	}
}

func TestCORS_CaseInsensitiveOriginMatch(t *testing.T) {
	srv := newTestServer([]string{"http://LOCALHOST:3000"})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()

	srv.corsMiddleware(srv.mux).ServeHTTP(rec, req)

	allowOrigin := rec.Header().Get("Access-Control-Allow-Origin")
	if allowOrigin != "http://localhost:3000" {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q (case-insensitive match)", allowOrigin, "http://localhost:3000")
	}
}

func TestServer_Uptime(t *testing.T) {
	srv := newTestServer([]string{"*"})

	uptime := srv.Uptime()
	if uptime < 0 {
		t.Errorf("Uptime() = %d, want >= 0", uptime)
	}
}
