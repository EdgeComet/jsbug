package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"

	"github.com/user/jsbug/internal/config"
	"github.com/user/jsbug/internal/fetcher"
	"github.com/user/jsbug/internal/parser"
)

func testAPIConfig() *config.Config {
	return &config.Config{
		API: config.APIConfig{
			Enabled: true,
			Keys:    []string{"test-key-abc123"},
		},
	}
}

func newTestExtHandler() *ExtRenderHandler {
	cfg := testAPIConfig()
	logger := zap.NewNop()
	p := parser.NewParser()

	mockFetcher := &MockFetcher{
		result: &fetcher.FetchResult{
			HTML:          "<html><body><h1>Test</h1><p>Hello world</p></body></html>",
			StatusCode:    200,
			FinalURL:      "https://example.com/",
			PageSizeBytes: 100,
			FetchTime:     0.5,
		},
	}

	renderHandler := NewRenderHandler(nil, mockFetcher, p, cfg, logger, nil, nil)
	return NewExtRenderHandler(renderHandler, cfg, logger)
}

func TestExtRenderHandler_MissingAPIKey(t *testing.T) {
	handler := newTestExtHandler()

	body := `{"url":"https://example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/ext/render", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["success"] != false {
		t.Errorf("success = %v, want false", resp["success"])
	}
	errObj, ok := resp["error"].(map[string]interface{})
	if !ok {
		t.Fatal("expected error object in response")
	}
	if errObj["code"] != "API_KEY_REQUIRED" {
		t.Errorf("error.code = %v, want API_KEY_REQUIRED", errObj["code"])
	}
}

func TestExtRenderHandler_InvalidAPIKey(t *testing.T) {
	handler := newTestExtHandler()

	body := `{"url":"https://example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/ext/render", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "wrong-key")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	errObj, ok := resp["error"].(map[string]interface{})
	if !ok {
		t.Fatal("expected error object in response")
	}
	if errObj["code"] != "API_KEY_INVALID" {
		t.Errorf("error.code = %v, want API_KEY_INVALID", errObj["code"])
	}
}

func TestExtRenderHandler_GetMethodNotAllowed(t *testing.T) {
	handler := newTestExtHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/ext/render", nil)
	req.Header.Set("X-API-Key", "test-key-abc123")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	errObj, ok := resp["error"].(map[string]interface{})
	if !ok {
		t.Fatal("expected error object in response")
	}
	if errObj["code"] != "METHOD_NOT_ALLOWED" {
		t.Errorf("error.code = %v, want METHOD_NOT_ALLOWED", errObj["code"])
	}
}

func TestExtRenderHandler_MalformedJSON(t *testing.T) {
	handler := newTestExtHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/ext/render", bytes.NewBufferString("not json at all"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-key-abc123")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	errObj, ok := resp["error"].(map[string]interface{})
	if !ok {
		t.Fatal("expected error object in response")
	}
	if errObj["code"] != "INVALID_REQUEST_BODY" {
		t.Errorf("error.code = %v, want INVALID_REQUEST_BODY", errObj["code"])
	}
}

func TestExtRenderHandler_EmptyBody(t *testing.T) {
	handler := newTestExtHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/ext/render", bytes.NewBufferString(""))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-key-abc123")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	errObj, ok := resp["error"].(map[string]interface{})
	if !ok {
		t.Fatal("expected error object in response")
	}
	if errObj["code"] != "INVALID_REQUEST_BODY" {
		t.Errorf("error.code = %v, want INVALID_REQUEST_BODY", errObj["code"])
	}
}

func TestExtRenderHandler_UnknownFields(t *testing.T) {
	handler := newTestExtHandler()

	body := `{"url":"https://example.com","inclue_html":true}`
	req := httptest.NewRequest(http.MethodPost, "/api/ext/render", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-key-abc123")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	errObj, ok := resp["error"].(map[string]interface{})
	if !ok {
		t.Fatal("expected error object in response")
	}
	if errObj["code"] != "INVALID_REQUEST_BODY" {
		t.Errorf("error.code = %v, want INVALID_REQUEST_BODY", errObj["code"])
	}
}

func TestExtRenderHandler_MissingURL(t *testing.T) {
	handler := newTestExtHandler()

	body := `{"js_enabled":false}`
	req := httptest.NewRequest(http.MethodPost, "/api/ext/render", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-key-abc123")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	errObj, ok := resp["error"].(map[string]interface{})
	if !ok {
		t.Fatal("expected error object in response")
	}
	if errObj["code"] != "INVALID_URL" {
		t.Errorf("error.code = %v, want INVALID_URL", errObj["code"])
	}
}

func TestExtRenderHandler_InvalidTimeout(t *testing.T) {
	handler := newTestExtHandler()

	body := `{"url":"https://example.com","timeout":100}`
	req := httptest.NewRequest(http.MethodPost, "/api/ext/render", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-key-abc123")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	errObj, ok := resp["error"].(map[string]interface{})
	if !ok {
		t.Fatal("expected error object in response")
	}
	if errObj["code"] != "INVALID_TIMEOUT" {
		t.Errorf("error.code = %v, want INVALID_TIMEOUT", errObj["code"])
	}
}

func TestExtRenderHandler_InvalidWaitEvent(t *testing.T) {
	handler := newTestExtHandler()

	body := `{"url":"https://example.com","wait_event":"invalid"}`
	req := httptest.NewRequest(http.MethodPost, "/api/ext/render", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-key-abc123")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	errObj, ok := resp["error"].(map[string]interface{})
	if !ok {
		t.Fatal("expected error object in response")
	}
	if errObj["code"] != "INVALID_WAIT_EVENT" {
		t.Errorf("error.code = %v, want INVALID_WAIT_EVENT", errObj["code"])
	}
}

func TestExtRenderHandler_MetadataOnlyResponse(t *testing.T) {
	handler := newTestExtHandler()

	body := `{"url":"https://example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/ext/render", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-key-abc123")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["success"] != true {
		t.Errorf("success = %v, want true", resp["success"])
	}

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected data object in response")
	}

	// Check required metadata fields are present
	if _, ok := data["status_code"]; !ok {
		t.Error("expected status_code in data")
	}
	if _, ok := data["final_url"]; !ok {
		t.Error("expected final_url in data")
	}
	if _, ok := data["title"]; !ok {
		t.Error("expected title in data")
	}

	// Check opt-in content fields are NOT present
	if _, ok := data["html"]; ok {
		t.Error("html should not be present when include_html is not set")
	}
	if _, ok := data["body_text"]; ok {
		t.Error("body_text should not be present when include_text is not set")
	}
	if _, ok := data["body_markdown"]; ok {
		t.Error("body_markdown should not be present when include_markdown is not set")
	}
	if _, ok := data["sections"]; ok {
		t.Error("sections should not be present when include_sections is not set")
	}
	if _, ok := data["links"]; ok {
		t.Error("links should not be present when include_links is not set")
	}
}

func TestExtRenderHandler_WithIncludeText(t *testing.T) {
	handler := newTestExtHandler()

	body := `{"url":"https://example.com","include_text":true}`
	req := httptest.NewRequest(http.MethodPost, "/api/ext/render", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-key-abc123")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected data object in response")
	}

	if _, ok := data["body_text"]; !ok {
		t.Error("expected body_text in data when include_text is true")
	}
	if _, ok := data["body_text_tokens_count"]; !ok {
		t.Error("expected body_text_tokens_count in data when include_text is true")
	}

	// Check that other opt-in fields are not present
	if _, ok := data["html"]; ok {
		t.Error("html should not be present when include_html is not set")
	}
	if _, ok := data["body_markdown"]; ok {
		t.Error("body_markdown should not be present when include_markdown is not set")
	}
}

func TestExtRenderHandler_WithIncludeHTML(t *testing.T) {
	handler := newTestExtHandler()

	body := `{"url":"https://example.com","include_html":true}`
	req := httptest.NewRequest(http.MethodPost, "/api/ext/render", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-key-abc123")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected data object in response")
	}

	htmlVal, ok := data["html"]
	if !ok {
		t.Fatal("expected html in data when include_html is true")
	}
	htmlStr, ok := htmlVal.(string)
	if !ok || htmlStr == "" {
		t.Error("expected html to be a non-empty string")
	}
}

func TestExtRenderHandler_WithIncludeSections(t *testing.T) {
	handler := newTestExtHandler()

	body := `{"url":"https://example.com","include_sections":true}`
	req := httptest.NewRequest(http.MethodPost, "/api/ext/render", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-key-abc123")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected data object in response")
	}

	sectionsVal, ok := data["sections"]
	if !ok {
		t.Fatal("expected sections in data when include_sections is true")
	}
	sections, ok := sectionsVal.([]interface{})
	if !ok {
		t.Fatal("expected sections to be an array")
	}
	if len(sections) < 1 {
		t.Error("expected at least 1 section from mock HTML with h1")
	}
}

func TestExtRenderHandler_ValidAPIKey(t *testing.T) {
	handler := newTestExtHandler()

	body := `{"url":"https://example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/ext/render", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-key-abc123")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["success"] != true {
		t.Errorf("success = %v, want true", resp["success"])
	}
}
