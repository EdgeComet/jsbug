package fetcher

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/user/jsbug/internal/security"
)

// Default timeouts
const (
	defaultDialTimeout    = 10 * time.Second
	defaultRequestTimeout = 30 * time.Second
)

// FetchOptions contains options for fetching a URL
type FetchOptions struct {
	URL             string
	UserAgent       string
	Timeout         time.Duration
	FollowRedirects bool // default should be true
}

// FetchResult contains the results of fetching a URL
type FetchResult struct {
	HTML          string
	FinalURL      string
	RedirectURL   string // Set when redirect detected but not followed
	StatusCode    int
	PageSizeBytes int
	FetchTime     float64 // seconds
	Headers       http.Header
}

// Fetcher performs HTTP requests for non-JS rendering
type Fetcher struct {
	client    *http.Client
	transport *http.Transport
	logger    *zap.Logger
}

// NewFetcher creates a new Fetcher with SSRF-safe transport
func NewFetcher(logger *zap.Logger) *Fetcher {
	transport := &http.Transport{
		DialContext: security.SSRFSafeDialContext,
	}

	client := &http.Client{
		Timeout:   defaultRequestTimeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	return &Fetcher{
		client:    client,
		transport: transport,
		logger:    logger,
	}
}

// NewUnsafeFetcher creates a Fetcher without SSRF protection. Use only in tests.
func NewUnsafeFetcher(logger *zap.Logger) *Fetcher {
	transport := &http.Transport{}

	client := &http.Client{
		Timeout:   defaultRequestTimeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	return &Fetcher{
		client:    client,
		transport: transport,
		logger:    logger,
	}
}

// ErrRedirectStopped is returned when redirect is detected but not followed
var ErrRedirectStopped = fmt.Errorf("redirect stopped")

// Fetch retrieves a URL using HTTP GET
func (f *Fetcher) Fetch(ctx context.Context, opts FetchOptions) (*FetchResult, error) {
	startTime := time.Now()

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, opts.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set User-Agent
	if opts.UserAgent != "" {
		req.Header.Set("User-Agent", opts.UserAgent)
	}

	// Set Accept header for HTML
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	// Create client with appropriate redirect handling
	var redirectURL string
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = defaultRequestTimeout
	}

	var checkRedirect func(req *http.Request, via []*http.Request) error
	if opts.FollowRedirects {
		checkRedirect = func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		}
	} else {
		checkRedirect = func(req *http.Request, via []*http.Request) error {
			redirectURL = req.URL.String()
			return http.ErrUseLastResponse
		}
	}

	client := &http.Client{
		Timeout:       timeout,
		Transport:     f.transport,
		CheckRedirect: checkRedirect,
	}

	f.logger.Debug("Fetching URL",
		zap.String("url", opts.URL),
		zap.String("user_agent", opts.UserAgent),
		zap.Bool("follow_redirects", opts.FollowRedirects),
	)

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	fetchTime := time.Since(startTime).Seconds()

	result := &FetchResult{
		HTML:          string(body),
		FinalURL:      resp.Request.URL.String(),
		RedirectURL:   redirectURL,
		StatusCode:    resp.StatusCode,
		PageSizeBytes: len(body),
		FetchTime:     fetchTime,
		Headers:       resp.Header,
	}

	f.logger.Debug("Fetch completed",
		zap.String("url", opts.URL),
		zap.String("final_url", result.FinalURL),
		zap.String("redirect_url", result.RedirectURL),
		zap.Int("status_code", result.StatusCode),
		zap.Int("size_bytes", result.PageSizeBytes),
		zap.Float64("fetch_time", fetchTime),
	)

	return result, nil
}

// GetContentType extracts Content-Type from headers
func (r *FetchResult) GetContentType() string {
	return r.Headers.Get("Content-Type")
}

// GetXRobotsTag extracts X-Robots-Tag from headers
func (r *FetchResult) GetXRobotsTag() string {
	return r.Headers.Get("X-Robots-Tag")
}

// GetLinkHeader returns the Link header value
func (r *FetchResult) GetLinkHeader() string {
	return r.Headers.Get("Link")
}

// GetCanonicalFromHeader extracts canonical URL from Link header
func (r *FetchResult) GetCanonicalFromHeader() string {
	link := r.Headers.Get("Link")
	if link == "" {
		return ""
	}

	// Parse Link header for rel="canonical"
	// Format: <URL>; rel="canonical"
	// This is a simplified parser
	for _, part := range splitLinkHeader(link) {
		if containsRel(part, "canonical") {
			return extractURL(part)
		}
	}

	return ""
}

// splitLinkHeader splits a Link header by commas (outside angle brackets)
func splitLinkHeader(header string) []string {
	var parts []string
	var current string
	inBrackets := false

	for _, c := range header {
		switch c {
		case '<':
			inBrackets = true
		case '>':
			inBrackets = false
		}

		if c == ',' && !inBrackets {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(c)
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}

// containsRel checks if a link part contains rel="canonical"
func containsRel(part, rel string) bool {
	// Look for rel="canonical" or rel=canonical
	target1 := fmt.Sprintf(`rel="%s"`, rel)
	target2 := fmt.Sprintf(`rel=%s`, rel)

	return strings.Contains(part, target1) || strings.Contains(part, target2)
}

// extractURL extracts URL from <URL>
func extractURL(part string) string {
	start := -1
	end := -1

	for i, c := range part {
		if c == '<' {
			start = i + 1
		} else if c == '>' && start >= 0 {
			end = i
			break
		}
	}

	if start >= 0 && end > start {
		return part[start:end]
	}

	return ""
}
