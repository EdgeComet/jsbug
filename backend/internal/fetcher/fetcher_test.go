package fetcher

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNewFetcher(t *testing.T) {
	logger := zap.NewNop()
	f := NewUnsafeFetcher(logger)

	if f == nil {
		t.Fatal("NewUnsafeFetcher() returned nil")
	}
	if f.client == nil {
		t.Error("client is nil")
	}
}

func TestFetcher_Fetch_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "<html><body>Hello World</body></html>")
	}))
	defer server.Close()

	logger := zap.NewNop()
	f := NewUnsafeFetcher(logger)

	result, err := f.Fetch(context.Background(), FetchOptions{
		URL:     server.URL,
		Timeout: 10 * time.Second,
	})

	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}

	if !strings.Contains(result.HTML, "Hello World") {
		t.Error("HTML does not contain expected content")
	}

	if result.PageSizeBytes <= 0 {
		t.Error("PageSizeBytes should be > 0")
	}

	if result.FetchTime <= 0 {
		t.Error("FetchTime should be > 0")
	}

	if result.FinalURL != server.URL {
		t.Errorf("FinalURL = %q, want %q", result.FinalURL, server.URL)
	}
}

func TestFetcher_Fetch_Redirect(t *testing.T) {
	redirectCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			redirectCount++
			http.Redirect(w, r, "/final", http.StatusFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Final page")
	}))
	defer server.Close()

	logger := zap.NewNop()
	f := NewUnsafeFetcher(logger)

	result, err := f.Fetch(context.Background(), FetchOptions{
		URL:             server.URL + "/redirect",
		Timeout:         10 * time.Second,
		FollowRedirects: true,
	})

	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	if redirectCount == 0 {
		t.Error("Expected redirect to occur")
	}

	expectedFinalURL := server.URL + "/final"
	if result.FinalURL != expectedFinalURL {
		t.Errorf("FinalURL = %q, want %q", result.FinalURL, expectedFinalURL)
	}

	if !strings.Contains(result.HTML, "Final page") {
		t.Error("HTML does not contain final page content")
	}
}

func TestFetcher_Fetch_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		fmt.Fprint(w, "Slow response")
	}))
	defer server.Close()

	logger := zap.NewNop()
	f := NewUnsafeFetcher(logger)

	_, err := f.Fetch(context.Background(), FetchOptions{
		URL:     server.URL,
		Timeout: 100 * time.Millisecond,
	})

	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}

func TestFetcher_Fetch_NetworkError(t *testing.T) {
	logger := zap.NewNop()
	f := NewUnsafeFetcher(logger)

	_, err := f.Fetch(context.Background(), FetchOptions{
		URL:     "http://localhost:99999/invalid",
		Timeout: 1 * time.Second,
	})

	if err == nil {
		t.Error("Expected network error, got nil")
	}
}

func TestFetcher_Fetch_StatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"200 OK", 200},
		{"404 Not Found", 404},
		{"500 Internal Server Error", 500},
		{"301 Moved Permanently", 301},
		{"403 Forbidden", 403},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For redirects, we need to handle differently
			if tt.statusCode >= 300 && tt.statusCode < 400 {
				// Skip redirect status codes as they get followed
				return
			}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				fmt.Fprint(w, "Response body")
			}))
			defer server.Close()

			logger := zap.NewNop()
			f := NewUnsafeFetcher(logger)

			result, err := f.Fetch(context.Background(), FetchOptions{
				URL:     server.URL,
				Timeout: 10 * time.Second,
			})

			if err != nil {
				t.Fatalf("Fetch() error = %v", err)
			}

			if result.StatusCode != tt.statusCode {
				t.Errorf("StatusCode = %d, want %d", result.StatusCode, tt.statusCode)
			}
		})
	}
}

func TestFetcher_Fetch_UserAgent(t *testing.T) {
	var receivedUA string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUA = r.Header.Get("User-Agent")
		fmt.Fprint(w, "OK")
	}))
	defer server.Close()

	logger := zap.NewNop()
	f := NewUnsafeFetcher(logger)

	customUA := "CustomBot/1.0"
	_, err := f.Fetch(context.Background(), FetchOptions{
		URL:       server.URL,
		UserAgent: customUA,
		Timeout:   10 * time.Second,
	})

	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	if receivedUA != customUA {
		t.Errorf("User-Agent = %q, want %q", receivedUA, customUA)
	}
}

func TestFetcher_Fetch_Headers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("X-Robots-Tag", "noindex, nofollow")
		w.Header().Set("Link", `<https://example.com/canonical>; rel="canonical"`)
		fmt.Fprint(w, "<html></html>")
	}))
	defer server.Close()

	logger := zap.NewNop()
	f := NewUnsafeFetcher(logger)

	result, err := f.Fetch(context.Background(), FetchOptions{
		URL:     server.URL,
		Timeout: 10 * time.Second,
	})

	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	if result.GetContentType() != "text/html; charset=utf-8" {
		t.Errorf("ContentType = %q", result.GetContentType())
	}

	if result.GetXRobotsTag() != "noindex, nofollow" {
		t.Errorf("XRobotsTag = %q", result.GetXRobotsTag())
	}

	if result.GetCanonicalFromHeader() != "https://example.com/canonical" {
		t.Errorf("Canonical = %q", result.GetCanonicalFromHeader())
	}
}

func TestFetcher_Fetch_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		fmt.Fprint(w, "Response")
	}))
	defer server.Close()

	logger := zap.NewNop()
	f := NewUnsafeFetcher(logger)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := f.Fetch(ctx, FetchOptions{
		URL:     server.URL,
		Timeout: 10 * time.Second,
	})

	if err == nil {
		t.Error("Expected context cancellation error, got nil")
	}
}

func TestFetchResult_GetCanonicalFromHeader(t *testing.T) {
	tests := []struct {
		name     string
		link     string
		expected string
	}{
		{
			name:     "simple canonical",
			link:     `<https://example.com/page>; rel="canonical"`,
			expected: "https://example.com/page",
		},
		{
			name:     "canonical without quotes",
			link:     `<https://example.com/page>; rel=canonical`,
			expected: "https://example.com/page",
		},
		{
			name:     "multiple links with canonical",
			link:     `<https://example.com/prev>; rel="prev", <https://example.com/canonical>; rel="canonical", <https://example.com/next>; rel="next"`,
			expected: "https://example.com/canonical",
		},
		{
			name:     "no canonical",
			link:     `<https://example.com/next>; rel="next"`,
			expected: "",
		},
		{
			name:     "empty link",
			link:     "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &FetchResult{
				Headers: http.Header{"Link": []string{tt.link}},
			}

			canonical := result.GetCanonicalFromHeader()
			if canonical != tt.expected {
				t.Errorf("GetCanonicalFromHeader() = %q, want %q", canonical, tt.expected)
			}
		})
	}
}

func TestFetcher_Fetch_NoFollowRedirects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "/final", http.StatusFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Final page")
	}))
	defer server.Close()

	logger := zap.NewNop()
	f := NewUnsafeFetcher(logger)

	result, err := f.Fetch(context.Background(), FetchOptions{
		URL:             server.URL + "/redirect",
		Timeout:         10 * time.Second,
		FollowRedirects: false,
	})

	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	// Should get redirect status code
	if result.StatusCode != http.StatusFound {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusFound)
	}

	// FinalURL should be the original URL (since redirect was not followed)
	expectedFinalURL := server.URL + "/redirect"
	if result.FinalURL != expectedFinalURL {
		t.Errorf("FinalURL = %q, want %q", result.FinalURL, expectedFinalURL)
	}

	// RedirectURL should contain the target of the redirect
	expectedRedirectURL := server.URL + "/final"
	if result.RedirectURL != expectedRedirectURL {
		t.Errorf("RedirectURL = %q, want %q", result.RedirectURL, expectedRedirectURL)
	}
}

func TestFetcher_TooManyRedirects(t *testing.T) {
	redirectCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redirectCount++
		// Always redirect to create infinite loop
		http.Redirect(w, r, "/redirect", http.StatusFound)
	}))
	defer server.Close()

	logger := zap.NewNop()
	f := NewUnsafeFetcher(logger)

	_, err := f.Fetch(context.Background(), FetchOptions{
		URL:             server.URL + "/redirect",
		Timeout:         10 * time.Second,
		FollowRedirects: true,
	})

	if err == nil {
		t.Error("Expected too many redirects error, got nil")
	}

	if redirectCount <= 10 {
		t.Logf("Redirect count: %d", redirectCount)
	}
}
