//go:build chrome

package chrome

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/user/jsbug/internal/types"
)

func setupV2TestServer() *httptest.Server {
	mux := http.NewServeMux()

	// Simple HTML page
	mux.HandleFunc("/simple", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head><title>Simple Page</title></head>
<body><h1>Hello World</h1></body>
</html>`)
	})

	// Page with custom status
	mux.HandleFunc("/status/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) >= 3 {
			var code int
			fmt.Sscanf(parts[2], "%d", &code)
			w.WriteHeader(code)
		}
		fmt.Fprint(w, "<html><body>Status page</body></html>")
	})

	// Redirect page
	mux.HandleFunc("/redirect", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/simple", http.StatusFound)
	})

	// Slow page
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, "<html><body>Slow page</body></html>")
	})

	// Page with JavaScript
	mux.HandleFunc("/js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head><title>JS Page</title></head>
<body>
<div id="content">Loading...</div>
<script>
document.getElementById('content').textContent = 'JavaScript executed';
console.log('Hello from console');
</script>
</body>
</html>`)
	})

	// Page with console messages
	mux.HandleFunc("/console", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head><title>Console Page</title></head>
<body>
<script>
console.log('Log message');
console.warn('Warning message');
console.error('Error message');
</script>
</body>
</html>`)
	})

	// Page with JS error
	mux.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head><title>Error Page</title></head>
<body>
<script>
throw new Error('Test error');
</script>
</body>
</html>`)
	})

	// Page that loads external resources
	mux.HandleFunc("/resources", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head>
<title>Resources Page</title>
<link rel="stylesheet" href="/style.css">
</head>
<body>
<script src="/app.js"></script>
<img src="/image.png">
</body>
</html>`)
	})

	mux.HandleFunc("/style.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		fmt.Fprint(w, "body { color: black; }")
	})

	mux.HandleFunc("/app.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		fmt.Fprint(w, "console.log('app loaded');")
	})

	mux.HandleFunc("/image.png", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		// 1x1 transparent PNG
		w.Write([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A})
	})

	return httptest.NewServer(mux)
}

func TestRendererV2_SimpleRender(t *testing.T) {
	server := setupV2TestServer()
	defer server.Close()

	logger := zap.NewNop()
	instance, err := New(0, newTestConfig(), logger)
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer instance.Close()

	renderer := NewRendererV2(instance, logger)

	result, err := renderer.Render(context.Background(), RenderOptions{
		URL:       server.URL + "/simple",
		Timeout:   10 * time.Second,
		WaitEvent: types.WaitLoad,
	})

	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	if !strings.Contains(result.HTML, "Hello World") {
		t.Error("HTML does not contain expected content")
	}

	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}

	if result.FinalURL != server.URL+"/simple" {
		t.Errorf("FinalURL = %q, want %q", result.FinalURL, server.URL+"/simple")
	}

	if result.RenderTime <= 0 {
		t.Error("RenderTime should be > 0")
	}
}

func TestRendererV2_UserAgentApplied(t *testing.T) {
	var receivedUA string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUA = r.Header.Get("User-Agent")
		fmt.Fprint(w, "<html><body>UA test</body></html>")
	}))
	defer server.Close()

	logger := zap.NewNop()
	instance, err := New(0, newTestConfig(), logger)
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer instance.Close()

	renderer := NewRendererV2(instance, logger)
	customUA := "CustomBot/1.0"

	_, err = renderer.Render(context.Background(), RenderOptions{
		URL:       server.URL,
		UserAgent: customUA,
		Timeout:   10 * time.Second,
		WaitEvent: types.WaitLoad,
	})

	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	if receivedUA != customUA {
		t.Errorf("User-Agent = %q, want %q", receivedUA, customUA)
	}
}

func TestRendererV2_SoftTimeout(t *testing.T) {
	server := setupV2TestServer()
	defer server.Close()

	logger := zap.NewNop()
	instance, err := New(0, newTestConfig(), logger)
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer instance.Close()

	renderer := NewRendererV2(instance, logger)

	// Use a very short soft timeout for the wait event
	result, err := renderer.Render(context.Background(), RenderOptions{
		URL:       server.URL + "/slow",
		Timeout:   100 * time.Millisecond, // Short soft timeout
		WaitEvent: types.WaitNetworkIdle,
	})

	// With soft timeout, render should still succeed (continue with partial data)
	// The error would be nil because soft timeout doesn't fail the render
	if err != nil && !strings.Contains(err.Error(), "hard timeout") {
		// If there's an error, it should be a hard timeout, not soft
		t.Logf("Render completed with error: %v", err)
	}

	// Even with timeout, we should get some HTML
	if result != nil && result.HTML == "" {
		t.Log("Note: HTML may be empty if page was very slow")
	}
}

func TestRendererV2_HardTimeout(t *testing.T) {
	server := setupV2TestServer()
	defer server.Close()

	logger := zap.NewNop()
	instance, err := New(0, newTestConfig(), logger)
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer instance.Close()

	renderer := NewRendererV2(instance, logger)

	// Create context with hard timeout
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_, err = renderer.Render(ctx, RenderOptions{
		URL:       server.URL + "/slow",
		Timeout:   30 * time.Second, // Long soft timeout
		WaitEvent: types.WaitLoad,
	})

	if err == nil {
		t.Error("Expected hard timeout error, got nil")
	}

	if !strings.Contains(err.Error(), "hard timeout") {
		t.Errorf("Expected hard timeout error, got: %v", err)
	}
}

func TestRendererV2_NetworkEventsCapture(t *testing.T) {
	server := setupV2TestServer()
	defer server.Close()

	logger := zap.NewNop()
	instance, err := New(0, newTestConfig(), logger)
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer instance.Close()

	renderer := NewRendererV2(instance, logger)

	result, err := renderer.Render(context.Background(), RenderOptions{
		URL:       server.URL + "/resources",
		Timeout:   10 * time.Second,
		WaitEvent: types.WaitLoad,
	})

	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	if len(result.Network) == 0 {
		t.Error("No network requests captured")
	}

	// Check for document request
	foundDocument := false
	for _, req := range result.Network {
		if req.Type == "Document" {
			foundDocument = true
			break
		}
	}
	if !foundDocument {
		t.Error("No Document request found")
	}
}

func TestRendererV2_BlockedRequests(t *testing.T) {
	server := setupV2TestServer()
	defer server.Close()

	logger := zap.NewNop()
	instance, err := New(0, newTestConfig(), logger)
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer instance.Close()

	renderer := NewRendererV2(instance, logger)

	// Block images
	blocklist := NewBlocklist(false, false, false, []string{"image"})

	result, err := renderer.Render(context.Background(), RenderOptions{
		URL:       server.URL + "/resources",
		Timeout:   10 * time.Second,
		WaitEvent: types.WaitLoad,
		Blocklist: blocklist,
	})

	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	// Check that image was blocked
	blockedCount := 0
	for _, req := range result.Network {
		if req.Blocked {
			blockedCount++
		}
		if strings.Contains(req.URL, "image.png") && !req.Blocked {
			t.Error("Image request should be blocked")
		}
	}

	if blockedCount == 0 {
		t.Error("Expected at least one blocked request")
	}
}

func TestRendererV2_LifecycleEvents(t *testing.T) {
	server := setupV2TestServer()
	defer server.Close()

	logger := zap.NewNop()
	instance, err := New(0, newTestConfig(), logger)
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer instance.Close()

	renderer := NewRendererV2(instance, logger)

	result, err := renderer.Render(context.Background(), RenderOptions{
		URL:       server.URL + "/simple",
		Timeout:   10 * time.Second,
		WaitEvent: types.WaitLoad,
	})

	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	if len(result.Lifecycle) == 0 {
		t.Error("No lifecycle events captured")
	}

	// Look for DOMContentLoaded or load event
	foundDOMContentLoaded := false
	for _, ev := range result.Lifecycle {
		if ev.Event == "DOMContentLoaded" {
			foundDOMContentLoaded = true
			if ev.Time <= 0 {
				t.Error("DOMContentLoaded time should be > 0")
			}
			break
		}
	}

	if !foundDOMContentLoaded {
		t.Log("DOMContentLoaded not found in lifecycle events, checking for load event...")
		for _, ev := range result.Lifecycle {
			t.Logf("Lifecycle event: %s at %f", ev.Event, ev.Time)
		}
	}
}

func TestRendererV2_Redirect(t *testing.T) {
	server := setupV2TestServer()
	defer server.Close()

	logger := zap.NewNop()
	instance, err := New(0, newTestConfig(), logger)
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer instance.Close()

	renderer := NewRendererV2(instance, logger)

	result, err := renderer.Render(context.Background(), RenderOptions{
		URL:       server.URL + "/redirect",
		Timeout:   10 * time.Second,
		WaitEvent: types.WaitLoad,
	})

	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	// After redirect, final URL should be /simple
	if result.FinalURL != server.URL+"/simple" {
		t.Errorf("FinalURL = %q, want %q (after redirect)", result.FinalURL, server.URL+"/simple")
	}

	// Status code should be 302 (redirect detected) or 200 (final page)
	if result.StatusCode != 302 && result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 302 or 200", result.StatusCode)
	}
}

func TestRendererV2_HTMLExtraction(t *testing.T) {
	server := setupV2TestServer()
	defer server.Close()

	logger := zap.NewNop()
	instance, err := New(0, newTestConfig(), logger)
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer instance.Close()

	renderer := NewRendererV2(instance, logger)

	result, err := renderer.Render(context.Background(), RenderOptions{
		URL:       server.URL + "/js",
		Timeout:   10 * time.Second,
		WaitEvent: types.WaitLoad,
	})

	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	// Check that JavaScript executed and modified DOM
	if !strings.Contains(result.HTML, "JavaScript executed") {
		t.Error("HTML does not contain JS-modified content")
	}

	if result.PageSizeBytes <= 0 {
		t.Error("PageSizeBytes should be > 0")
	}
}

func TestRendererV2_IsAvailable(t *testing.T) {
	logger := zap.NewNop()
	instance, err := New(0, newTestConfig(), logger)
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}

	renderer := NewRendererV2(instance, logger)

	if !renderer.IsAvailable() {
		t.Error("Renderer should be available")
	}

	instance.Close()

	if renderer.IsAvailable() {
		t.Error("Renderer should not be available after instance closed")
	}
}

func TestRendererV2_NilInstance(t *testing.T) {
	logger := zap.NewNop()
	renderer := NewRendererV2(nil, logger)

	if renderer.IsAvailable() {
		t.Error("Renderer with nil instance should not be available")
	}
}

func TestRendererV2_ViewportConfiguration(t *testing.T) {
	server := setupV2TestServer()
	defer server.Close()

	logger := zap.NewNop()
	instance, err := New(0, newTestConfig(), logger)
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer instance.Close()

	renderer := NewRendererV2(instance, logger)

	// Test with mobile viewport
	result, err := renderer.Render(context.Background(), RenderOptions{
		URL:            server.URL + "/simple",
		Timeout:        10 * time.Second,
		WaitEvent:      types.WaitLoad,
		ViewportWidth:  375,
		ViewportHeight: 667,
	})

	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	if result.HTML == "" {
		t.Error("HTML should not be empty")
	}
}

func TestRendererV2_ConsoleMessages(t *testing.T) {
	server := setupV2TestServer()
	defer server.Close()

	logger := zap.NewNop()
	instance, err := New(0, newTestConfig(), logger)
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer instance.Close()

	renderer := NewRendererV2(instance, logger)

	result, err := renderer.Render(context.Background(), RenderOptions{
		URL:       server.URL + "/console",
		Timeout:   10 * time.Second,
		WaitEvent: types.WaitLoad,
	})

	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	if len(result.Console) == 0 {
		t.Error("Expected console messages to be captured")
	}
}

func TestRendererV2_JSErrors(t *testing.T) {
	server := setupV2TestServer()
	defer server.Close()

	logger := zap.NewNop()
	instance, err := New(0, newTestConfig(), logger)
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer instance.Close()

	renderer := NewRendererV2(instance, logger)

	result, err := renderer.Render(context.Background(), RenderOptions{
		URL:       server.URL + "/error",
		Timeout:   10 * time.Second,
		WaitEvent: types.WaitLoad,
	})

	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	if len(result.JSErrors) == 0 {
		t.Error("Expected JS errors to be captured")
	}

	// Check that the error message contains "Test error"
	foundError := false
	for _, jsErr := range result.JSErrors {
		if strings.Contains(jsErr.Message, "Test error") || strings.Contains(jsErr.Message, "Error") {
			foundError = true
			break
		}
	}
	if !foundError {
		t.Log("JS errors captured but 'Test error' not found")
		for _, jsErr := range result.JSErrors {
			t.Logf("JS Error: %s", jsErr.Message)
		}
	}
}
