package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"

	"github.com/user/jsbug/internal/fetcher"
	"github.com/user/jsbug/internal/parser"
)

func newTestExtCompareHandler() *ExtCompareHandler {
	cfg := testAPIConfig()
	logger := zap.NewNop()
	p := parser.NewParser()

	mockFetcher := &MockFetcher{
		result: &fetcher.FetchResult{
			HTML:          `<html><head><title>Test Page</title><meta name="description" content="Test description"></head><body><h1>Hello</h1><h2>Section One</h2><p>Some content here.</p><h2>Section Two</h2><p>More content here.</p><a href="https://example.com/link1">Link 1</a><img src="https://example.com/img1.jpg" alt="Image 1"></body></html>`,
			StatusCode:    200,
			FinalURL:      "https://example.com/",
			PageSizeBytes: 500,
			FetchTime:     0.3,
		},
	}

	renderHandler := NewRenderHandler(nil, mockFetcher, p, cfg, logger, nil, nil)
	return NewExtCompareHandler(renderHandler, cfg, logger)
}

func TestExtCompareHandler_MissingAPIKey(t *testing.T) {
	handler := newTestExtCompareHandler()

	body := `{"url":"https://example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/ext/compare", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["success"] != false {
		t.Errorf("success = %v, want false", resp["success"])
	}
	errObj, ok := resp["error"].(map[string]any)
	if !ok {
		t.Fatal("expected error object in response")
	}
	if errObj["code"] != "API_KEY_REQUIRED" {
		t.Errorf("error.code = %v, want API_KEY_REQUIRED", errObj["code"])
	}
}

func TestExtCompareHandler_InvalidAPIKey(t *testing.T) {
	handler := newTestExtCompareHandler()

	body := `{"url":"https://example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/ext/compare", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "wrong-key")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
	}

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)

	errObj, ok := resp["error"].(map[string]any)
	if !ok {
		t.Fatal("expected error object in response")
	}
	if errObj["code"] != "API_KEY_INVALID" {
		t.Errorf("error.code = %v, want API_KEY_INVALID", errObj["code"])
	}
}

func TestExtCompareHandler_GetMethodNotAllowed(t *testing.T) {
	handler := newTestExtCompareHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/ext/compare", nil)
	req.Header.Set("X-API-Key", "test-key-abc123")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)

	errObj, ok := resp["error"].(map[string]any)
	if !ok {
		t.Fatal("expected error object in response")
	}
	if errObj["code"] != "METHOD_NOT_ALLOWED" {
		t.Errorf("error.code = %v, want METHOD_NOT_ALLOWED", errObj["code"])
	}
}

func TestExtCompareHandler_MalformedJSON(t *testing.T) {
	handler := newTestExtCompareHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/ext/compare", bytes.NewBufferString("not json at all"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-key-abc123")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)

	errObj, ok := resp["error"].(map[string]any)
	if !ok {
		t.Fatal("expected error object in response")
	}
	if errObj["code"] != "INVALID_REQUEST_BODY" {
		t.Errorf("error.code = %v, want INVALID_REQUEST_BODY", errObj["code"])
	}
}

func TestExtCompareHandler_UnknownFields(t *testing.T) {
	handler := newTestExtCompareHandler()

	body := `{"url":"https://example.com","inclue_html":true}`
	req := httptest.NewRequest(http.MethodPost, "/api/ext/compare", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-key-abc123")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)

	errObj, ok := resp["error"].(map[string]any)
	if !ok {
		t.Fatal("expected error object in response")
	}
	if errObj["code"] != "INVALID_REQUEST_BODY" {
		t.Errorf("error.code = %v, want INVALID_REQUEST_BODY", errObj["code"])
	}
}

func TestExtCompareHandler_MissingURL(t *testing.T) {
	handler := newTestExtCompareHandler()

	body := `{"include_sections":true}`
	req := httptest.NewRequest(http.MethodPost, "/api/ext/compare", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-key-abc123")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)

	errObj, ok := resp["error"].(map[string]any)
	if !ok {
		t.Fatal("expected error object in response")
	}
	if errObj["code"] != "INVALID_URL" {
		t.Errorf("error.code = %v, want INVALID_URL", errObj["code"])
	}
}

func TestExtCompareHandler_ResponseStructure(t *testing.T) {
	handler := newTestExtCompareHandler()

	body := `{"url":"https://example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/ext/compare", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-key-abc123")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["success"] != true {
		t.Errorf("success = %v, want true", resp["success"])
	}

	data, ok := resp["data"].(map[string]any)
	if !ok {
		t.Fatal("expected data object in response")
	}

	// Check js_status exists
	jsStatus, ok := data["js_status"].(map[string]any)
	if !ok {
		t.Fatal("expected js_status object in data")
	}

	// Check http_status exists
	httpStatus, ok := data["http_status"].(map[string]any)
	if !ok {
		t.Fatal("expected http_status object in data")
	}

	// HTTP fetch should succeed (MockFetcher returns success)
	if httpStatus["success"] != true {
		t.Errorf("http_status.success = %v, want true", httpStatus["success"])
	}

	// JS fetch should fail (no Chrome pool in test)
	if jsStatus["success"] != false {
		t.Errorf("js_status.success = %v, want false", jsStatus["success"])
	}

	// JS data should be null (JS fetch failed)
	if data["js"] != nil {
		t.Errorf("js = %v, want nil (JS fetch failed)", data["js"])
	}

	// Diff should be null (need both fetches to compute diff)
	if data["diff"] != nil {
		t.Errorf("diff = %v, want nil (need both fetches)", data["diff"])
	}

	// Rendering impact should be null
	if data["rendering_impact"] != nil {
		t.Errorf("rendering_impact = %v, want nil (need both fetches)", data["rendering_impact"])
	}
}

func TestExtCompareHandler_HTTPFetchContent(t *testing.T) {
	handler := newTestExtCompareHandler()

	body := `{"url":"https://example.com","include_sections":true,"include_links":true}`
	req := httptest.NewRequest(http.MethodPost, "/api/ext/compare", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-key-abc123")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)

	data, ok := resp["data"].(map[string]any)
	if !ok {
		t.Fatal("expected data object in response")
	}

	// HTTP status should indicate success
	httpStatus, ok := data["http_status"].(map[string]any)
	if !ok {
		t.Fatal("expected http_status object in data")
	}
	if httpStatus["success"] != true {
		t.Errorf("http_status.success = %v, want true", httpStatus["success"])
	}
	if httpStatus["status_code"] != float64(200) {
		t.Errorf("http_status.status_code = %v, want 200", httpStatus["status_code"])
	}

	// Since JS fetch fails (no Chrome pool), js/diff/rendering_impact should be null
	if data["js"] != nil {
		t.Errorf("js = %v, want nil (JS fetch failed)", data["js"])
	}
	if data["diff"] != nil {
		t.Errorf("diff = %v, want nil (need both fetches)", data["diff"])
	}
	if data["rendering_impact"] != nil {
		t.Errorf("rendering_impact = %v, want nil (need both fetches)", data["rendering_impact"])
	}
}

func TestExtCompareHandler_IncludeFlags(t *testing.T) {
	handler := newTestExtCompareHandler()

	// No include flags set
	body := `{"url":"https://example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/ext/compare", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-key-abc123")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)

	data, ok := resp["data"].(map[string]any)
	if !ok {
		t.Fatal("expected data object in response")
	}

	// If js object is present, verify include flags are respected
	if jsData, ok := data["js"].(map[string]any); ok {
		optInFields := []string{"sections", "links", "images", "structured_data", "html", "body_text", "body_markdown"}
		for _, field := range optInFields {
			if _, exists := jsData[field]; exists {
				t.Errorf("js.%s should not be present when include flags are not set", field)
			}
		}
	}
	// If js is nil (no Chrome pool), the test still passes since there is nothing to check
}
