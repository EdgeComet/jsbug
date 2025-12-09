//go:build chrome

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/user/jsbug/internal/types"
)

// Test fixtures - HTML pages for various test scenarios

const SimpleHTML = `<!DOCTYPE html>
<html>
<head>
	<title>Simple Test Page</title>
	<meta name="description" content="A simple test page for integration testing">
	<meta name="robots" content="index, follow">
	<link rel="canonical" href="http://localhost/simple">
</head>
<body>
	<h1>Welcome to the Test Page</h1>
	<h2>Section One</h2>
	<p>This is a paragraph with some content.</p>
	<h2>Section Two</h2>
	<p>More content here.</p>
	<a href="/internal">Internal Link</a>
	<a href="https://external.com">External Link</a>
</body>
</html>`

const SPAHTML = `<!DOCTYPE html>
<html>
<head>
	<title>SPA Test Page</title>
</head>
<body>
	<div id="app">Loading...</div>
	<script>
		setTimeout(function() {
			document.getElementById('app').innerHTML = '<h1>Rendered by JavaScript</h1><p>Dynamic content loaded</p>';
			document.title = 'SPA - Loaded';
		}, 100);
	</script>
</body>
</html>`

const AnalyticsHTML = `<!DOCTYPE html>
<html>
<head>
	<title>Analytics Test</title>
	<script async src="https://www.google-analytics.com/analytics.js"></script>
	<script async src="https://www.googletagmanager.com/gtag/js"></script>
</head>
<body>
	<h1>Page with Analytics</h1>
</body>
</html>`

const StructuredDataHTML = `<!DOCTYPE html>
<html>
<head>
	<title>Structured Data Test</title>
	<meta property="og:title" content="OG Title">
	<meta property="og:description" content="OG Description">
	<meta property="og:image" content="https://example.com/image.jpg">
	<script type="application/ld+json">
	{
		"@context": "https://schema.org",
		"@type": "Article",
		"headline": "Test Article",
		"author": "Test Author"
	}
	</script>
</head>
<body>
	<h1>Structured Data Page</h1>
</body>
</html>`

const SlowHTML = `<!DOCTYPE html>
<html>
<head>
	<title>Slow Page</title>
</head>
<body>
	<h1>This page is slow</h1>
</body>
</html>`

// FixtureServer creates a test HTTP server with fixture pages
type FixtureServer struct {
	Server *httptest.Server
}

// NewFixtureServer creates a new fixture server
func NewFixtureServer() *FixtureServer {
	mux := http.NewServeMux()

	mux.HandleFunc("/simple", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, SimpleHTML)
	})

	mux.HandleFunc("/spa", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, SPAHTML)
	})

	mux.HandleFunc("/analytics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, AnalyticsHTML)
	})

	mux.HandleFunc("/structured", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, StructuredDataHTML)
	})

	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, SlowHTML)
	})

	mux.HandleFunc("/redirect", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/simple", http.StatusFound)
	})

	mux.HandleFunc("/404", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `<html><head><title>Not Found</title></head><body><h1>404 Not Found</h1></body></html>`)
	})

	return &FixtureServer{
		Server: httptest.NewServer(mux),
	}
}

// Close closes the fixture server
func (f *FixtureServer) Close() {
	f.Server.Close()
}

// URL returns the base URL of the fixture server
func (f *FixtureServer) URL() string {
	return f.Server.URL
}

// RenderClient helps make render requests
type RenderClient struct {
	BaseURL string
	Client  *http.Client
}

// NewRenderClient creates a new render client
func NewRenderClient(baseURL string) *RenderClient {
	return &RenderClient{
		BaseURL: baseURL,
		Client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Render makes a render request
func (c *RenderClient) Render(req *types.RenderRequest) (*types.RenderResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Post(c.BaseURL+"/api/render", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result types.RenderResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}
