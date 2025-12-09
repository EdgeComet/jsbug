package robots

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNewChecker(t *testing.T) {
	logger := zap.NewNop()
	c := NewChecker(logger)

	if c == nil {
		t.Fatal("NewChecker() returned nil")
	}
	if c.logger == nil {
		t.Error("logger is nil")
	}
}

func TestChecker_Check_Allowed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `
User-agent: *
Allow: /

User-agent: Googlebot
Allow: /public/
Disallow: /private/
`)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := zap.NewNop()
	c := NewChecker(logger)

	// Test allowed URL
	allowed, err := c.Check(context.Background(), server.URL+"/public/page")
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !allowed {
		t.Error("Check() returned false, expected true for allowed URL")
	}
}

func TestChecker_Check_Disallowed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `
User-agent: Googlebot
Disallow: /private/
`)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := zap.NewNop()
	c := NewChecker(logger)

	// Test disallowed URL
	allowed, err := c.Check(context.Background(), server.URL+"/private/secret")
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if allowed {
		t.Error("Check() returned true, expected false for disallowed URL")
	}
}

func TestChecker_Check_MissingRobotsTxt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "Not Found")
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := zap.NewNop()
	c := NewChecker(logger)

	// Missing robots.txt should return allowed (fail open)
	allowed, err := c.Check(context.Background(), server.URL+"/any/page")
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !allowed {
		t.Error("Check() returned false, expected true when robots.txt is missing")
	}
}

func TestChecker_Check_MalformedRobotsTxt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			w.WriteHeader(http.StatusOK)
			// Return some garbage that can't be parsed properly
			// Note: robotstxt library is quite lenient, so we return empty/invalid content
			fmt.Fprint(w, "")
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := zap.NewNop()
	c := NewChecker(logger)

	// Empty/malformed robots.txt should return allowed
	allowed, err := c.Check(context.Background(), server.URL+"/any/page")
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !allowed {
		t.Error("Check() returned false, expected true for empty robots.txt")
	}
}

func TestChecker_Check_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "Internal Server Error")
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := zap.NewNop()
	c := NewChecker(logger)

	// Server error should return allowed (fail open)
	allowed, err := c.Check(context.Background(), server.URL+"/any/page")
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !allowed {
		t.Error("Check() returned false, expected true on server error")
	}
}

func TestChecker_Check_NetworkError(t *testing.T) {
	logger := zap.NewNop()
	c := NewChecker(logger)

	// Invalid URL should fail but return allowed (fail open)
	allowed, err := c.Check(context.Background(), "http://localhost:99999/page")
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !allowed {
		t.Error("Check() returned false, expected true on network error")
	}
}

func TestChecker_Check_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			// Sleep longer than the checker's timeout
			time.Sleep(10 * time.Second)
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "User-agent: *\nDisallow: /")
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := zap.NewNop()
	c := NewChecker(logger)

	// Context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Timeout should return allowed (fail open)
	allowed, err := c.Check(ctx, server.URL+"/any/page")
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !allowed {
		t.Error("Check() returned false, expected true on timeout")
	}
}

func TestChecker_Check_UserAgent(t *testing.T) {
	var receivedUA string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			receivedUA = r.Header.Get("User-Agent")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "User-agent: *\nAllow: /")
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := zap.NewNop()
	c := NewChecker(logger)

	_, err := c.Check(context.Background(), server.URL+"/page")
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	expectedUA := "jsbug-robots/1.0"
	if receivedUA != expectedUA {
		t.Errorf("User-Agent = %q, want %q", receivedUA, expectedUA)
	}
}

func TestChecker_Check_GooglebotSpecificRules(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `
User-agent: *
Allow: /

User-agent: Googlebot
Disallow: /googlebot-only/

User-agent: Bingbot
Disallow: /bingbot-only/
`)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := zap.NewNop()
	c := NewChecker(logger)

	// Googlebot-specific disallow should be respected
	allowed, err := c.Check(context.Background(), server.URL+"/googlebot-only/page")
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if allowed {
		t.Error("Check() returned true, expected false for Googlebot-specific disallow")
	}

	// Bingbot-specific disallow should NOT affect Googlebot
	allowed, err = c.Check(context.Background(), server.URL+"/bingbot-only/page")
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !allowed {
		t.Error("Check() returned false, expected true (Bingbot rule should not affect Googlebot)")
	}
}

func TestChecker_Check_DisallowAll(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `
User-agent: *
Disallow: /
`)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := zap.NewNop()
	c := NewChecker(logger)

	// Disallow all should block everything
	allowed, err := c.Check(context.Background(), server.URL+"/any/page")
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if allowed {
		t.Error("Check() returned true, expected false for Disallow: /")
	}
}

func TestChecker_Check_AllowAll(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `
User-agent: *
Allow: /
`)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := zap.NewNop()
	c := NewChecker(logger)

	// Allow all should permit everything
	allowed, err := c.Check(context.Background(), server.URL+"/any/page")
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !allowed {
		t.Error("Check() returned false, expected true for Allow: /")
	}
}
