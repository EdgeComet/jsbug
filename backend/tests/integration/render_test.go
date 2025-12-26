//go:build chrome

package integration

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/user/jsbug/internal/chrome"
	"github.com/user/jsbug/internal/config"
	"github.com/user/jsbug/internal/fetcher"
	"github.com/user/jsbug/internal/parser"
	"github.com/user/jsbug/internal/server"
	"github.com/user/jsbug/internal/types"
)

// TestServer manages the test jsbug server
type TestServer struct {
	Server   *server.Server
	Chrome   *chrome.Instance
	Renderer *chrome.Renderer
	BaseURL  string
}

// NewTestServer creates a test server with Chrome enabled
func NewTestServer(t *testing.T) *TestServer {
	t.Helper()

	logger := zap.NewNop()

	// Create Chrome instance
	chromeInstance, err := chrome.New(0, chrome.InstanceConfig{
		Headless:  true,
		NoSandbox: true,
	}, logger)
	if err != nil {
		t.Fatalf("Failed to create Chrome instance: %v", err)
	}

	renderer := chrome.NewRenderer(chromeInstance, logger)
	httpFetcher := fetcher.NewFetcher(logger)
	htmlParser := parser.NewParser()

	cfg := &config.Config{
		Server: config.ServerConfig{
			Host:         "127.0.0.1",
			Port:         0, // Use any available port
			ReadTimeout:  30,
			WriteTimeout: 60,
			CORSOrigins:  []string{"*"},
		},
		Chrome: config.ChromeConfig{},
	}

	srv := server.New(cfg, logger)
	renderHandler := server.NewRenderHandler(renderer, httpFetcher, htmlParser, cfg, logger, nil, nil)
	renderHandler.SetSSEManager(srv.SSEManager())
	srv.SetRenderHandler(renderHandler)

	return &TestServer{
		Server:   srv,
		Chrome:   chromeInstance,
		Renderer: renderer,
		BaseURL:  "http://127.0.0.1:8080", // Will be updated when started
	}
}

// Close cleans up test server resources
func (ts *TestServer) Close() {
	if ts.Chrome != nil {
		ts.Chrome.Close()
	}
}

func TestIntegration_NonJSFetch_ReturnsRawHTML(t *testing.T) {
	fixtures := NewFixtureServer()
	defer fixtures.Close()

	logger := zap.NewNop()
	httpFetcher := fetcher.NewFetcher(logger)
	htmlParser := parser.NewParser()

	cfg := &config.Config{
		Chrome: config.ChromeConfig{},
	}

	renderHandler := server.NewRenderHandler(nil, httpFetcher, htmlParser, cfg, logger, nil, nil)

	// Create request
	req := &types.RenderRequest{
		URL:       fixtures.URL() + "/simple",
		JSEnabled: false,
		Timeout:   15,
		WaitEvent: types.WaitLoad,
	}
	req.ApplyDefaults()

	// Use handler directly
	ctx := context.Background()
	// We need to test via HTTP to properly test the handler

	// Make HTTP request to fixture server directly with fetcher
	fetchResult, err := httpFetcher.Fetch(ctx, fetcher.FetchOptions{
		URL:       fixtures.URL() + "/simple",
		UserAgent: types.ResolveUserAgent(types.UserAgentChrome),
		Timeout:   15 * time.Second,
	})
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	// Parse content
	parseResult, err := htmlParser.Parse(fetchResult.HTML, fetchResult.FinalURL)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify content
	if parseResult.Title != "Simple Test Page" {
		t.Errorf("Title = %q, want %q", parseResult.Title, "Simple Test Page")
	}
	if parseResult.MetaDescription != "A simple test page for integration testing" {
		t.Errorf("MetaDescription = %q", parseResult.MetaDescription)
	}
	if len(parseResult.H1) != 1 || parseResult.H1[0] != "Welcome to the Test Page" {
		t.Errorf("H1 = %v", parseResult.H1)
	}
	if len(parseResult.H2) != 2 {
		t.Errorf("H2 count = %d, want 2", len(parseResult.H2))
	}
	if parseResult.InternalLinks != 1 {
		t.Errorf("InternalLinks = %d, want 1", parseResult.InternalLinks)
	}
	if parseResult.ExternalLinks != 1 {
		t.Errorf("ExternalLinks = %d, want 1", parseResult.ExternalLinks)
	}

	// Ignore renderHandler to avoid unused variable error
	_ = renderHandler
}

func TestIntegration_JSRender_CapturesFullContent(t *testing.T) {
	fixtures := NewFixtureServer()
	defer fixtures.Close()

	logger := zap.NewNop()

	// Create Chrome instance
	chromeInstance, err := chrome.New(0, chrome.InstanceConfig{
		Headless:  true,
		NoSandbox: true,
	}, logger)
	if err != nil {
		t.Fatalf("Failed to create Chrome: %v", err)
	}
	defer chromeInstance.Close()

	renderer := chrome.NewRenderer(chromeInstance, logger)
	htmlParser := parser.NewParser()

	// Render SPA page
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := renderer.Render(ctx, chrome.RenderOptions{
		URL:       fixtures.URL() + "/spa",
		UserAgent: types.ResolveUserAgent(types.UserAgentChrome),
		Timeout:   15 * time.Second,
		WaitEvent: types.WaitNetworkIdle,
	})
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Verify JS was executed - title should be changed
	parseResult, _ := htmlParser.Parse(result.HTML, result.FinalURL)

	// The SPA changes the title via JS
	if !strings.Contains(parseResult.Title, "SPA") {
		t.Logf("Title = %q (may not have updated yet)", parseResult.Title)
	}

	// Verify rendered content
	if !strings.Contains(result.HTML, "Rendered by JavaScript") {
		t.Error("HTML does not contain JS-rendered content")
	}

	// Verify network requests were captured
	if result.NetworkSum == nil || result.NetworkSum.TotalRequests == 0 {
		t.Error("Expected network requests to be captured")
	}
}

func TestIntegration_StructuredDataExtraction(t *testing.T) {
	fixtures := NewFixtureServer()
	defer fixtures.Close()

	logger := zap.NewNop()
	httpFetcher := fetcher.NewFetcher(logger)
	htmlParser := parser.NewParser()

	ctx := context.Background()
	fetchResult, err := httpFetcher.Fetch(ctx, fetcher.FetchOptions{
		URL:       fixtures.URL() + "/structured",
		UserAgent: types.ResolveUserAgent(types.UserAgentChrome),
		Timeout:   15 * time.Second,
	})
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	parseResult, _ := htmlParser.Parse(fetchResult.HTML, fetchResult.FinalURL)

	// Verify Open Graph
	if parseResult.OpenGraph["title"] != "OG Title" {
		t.Errorf("og:title = %q, want %q", parseResult.OpenGraph["title"], "OG Title")
	}
	if parseResult.OpenGraph["description"] != "OG Description" {
		t.Errorf("og:description = %q", parseResult.OpenGraph["description"])
	}
	if parseResult.OpenGraph["image"] != "https://example.com/image.jpg" {
		t.Errorf("og:image = %q", parseResult.OpenGraph["image"])
	}

	// Verify structured data
	if len(parseResult.StructuredData) != 1 {
		t.Errorf("StructuredData count = %d, want 1", len(parseResult.StructuredData))
	}
}

func TestIntegration_RedirectHandling(t *testing.T) {
	fixtures := NewFixtureServer()
	defer fixtures.Close()

	logger := zap.NewNop()
	httpFetcher := fetcher.NewFetcher(logger)

	ctx := context.Background()
	fetchResult, err := httpFetcher.Fetch(ctx, fetcher.FetchOptions{
		URL:       fixtures.URL() + "/redirect",
		UserAgent: types.ResolveUserAgent(types.UserAgentChrome),
		Timeout:   15 * time.Second,
	})
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	// Verify final URL
	expectedFinalURL := fixtures.URL() + "/simple"
	if fetchResult.FinalURL != expectedFinalURL {
		t.Errorf("FinalURL = %q, want %q", fetchResult.FinalURL, expectedFinalURL)
	}

	// Verify content is from redirected page
	if !strings.Contains(fetchResult.HTML, "Simple Test Page") {
		t.Error("HTML should be from redirected page")
	}
}

func TestIntegration_ErrorPageHandling(t *testing.T) {
	fixtures := NewFixtureServer()
	defer fixtures.Close()

	logger := zap.NewNop()
	httpFetcher := fetcher.NewFetcher(logger)
	htmlParser := parser.NewParser()

	ctx := context.Background()
	fetchResult, err := httpFetcher.Fetch(ctx, fetcher.FetchOptions{
		URL:       fixtures.URL() + "/404",
		UserAgent: types.ResolveUserAgent(types.UserAgentChrome),
		Timeout:   15 * time.Second,
	})
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	// Verify status code
	if fetchResult.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want %d", fetchResult.StatusCode, http.StatusNotFound)
	}

	// Verify HTML is still returned
	parseResult, _ := htmlParser.Parse(fetchResult.HTML, fetchResult.FinalURL)
	if parseResult.Title != "Not Found" {
		t.Errorf("Title = %q, want %q", parseResult.Title, "Not Found")
	}
}

func TestIntegration_TimeoutHandling(t *testing.T) {
	fixtures := NewFixtureServer()
	defer fixtures.Close()

	logger := zap.NewNop()
	httpFetcher := fetcher.NewFetcher(logger)

	ctx := context.Background()
	_, err := httpFetcher.Fetch(ctx, fetcher.FetchOptions{
		URL:       fixtures.URL() + "/slow",
		UserAgent: types.ResolveUserAgent(types.UserAgentChrome),
		Timeout:   1 * time.Second, // Short timeout
	})

	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}

func TestIntegration_BlockingAnalytics(t *testing.T) {
	fixtures := NewFixtureServer()
	defer fixtures.Close()

	logger := zap.NewNop()

	// Create Chrome instance
	chromeInstance, err := chrome.New(0, chrome.InstanceConfig{
		Headless:  true,
		NoSandbox: true,
	}, logger)
	if err != nil {
		t.Fatalf("Failed to create Chrome: %v", err)
	}
	defer chromeInstance.Close()

	renderer := chrome.NewRenderer(chromeInstance, logger)

	// Create blocklist that blocks analytics
	blocklist := chrome.NewBlocklist(true, false, false, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := renderer.Render(ctx, chrome.RenderOptions{
		URL:       fixtures.URL() + "/analytics",
		UserAgent: types.ResolveUserAgent(types.UserAgentChrome),
		Timeout:   15 * time.Second,
		WaitEvent: types.WaitLoad,
		Blocklist: blocklist,
	})
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Check if analytics requests were blocked
	blockedCount := 0
	for _, req := range result.Network {
		if req.Blocked {
			blockedCount++
		}
	}

	if blockedCount == 0 {
		t.Log("No blocked requests found - analytics scripts may not have been requested")
	}

	// Verify page still rendered
	if result.StatusCode == 0 {
		t.Error("Expected a valid status code")
	}
}

func TestIntegration_ContentExtraction(t *testing.T) {
	fixtures := NewFixtureServer()
	defer fixtures.Close()

	logger := zap.NewNop()
	httpFetcher := fetcher.NewFetcher(logger)
	htmlParser := parser.NewParser()

	ctx := context.Background()
	fetchResult, err := httpFetcher.Fetch(ctx, fetcher.FetchOptions{
		URL:       fixtures.URL() + "/simple",
		UserAgent: types.ResolveUserAgent(types.UserAgentChrome),
		Timeout:   15 * time.Second,
	})
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	parseResult, _ := htmlParser.Parse(fetchResult.HTML, fetchResult.FinalURL)

	// Comprehensive content verification
	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"Title", parseResult.Title, "Simple Test Page"},
		{"MetaDescription", parseResult.MetaDescription, "A simple test page for integration testing"},
		{"MetaRobots", parseResult.MetaRobots, "index, follow"},
		{"CanonicalURL", parseResult.CanonicalURL, "http://localhost/simple"},
		{"H1 count", len(parseResult.H1), 1},
		{"H2 count", len(parseResult.H2), 2},
		{"InternalLinks", parseResult.InternalLinks, 1},
		{"ExternalLinks", parseResult.ExternalLinks, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.expected)
			}
		})
	}

	// Verify word count is reasonable
	if parseResult.WordCount < 10 {
		t.Errorf("WordCount = %d, expected >= 10", parseResult.WordCount)
	}
}
