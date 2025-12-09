package robots

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/temoto/robotstxt"
	"go.uber.org/zap"
)

const (
	// fetchTimeout is the timeout for fetching robots.txt
	fetchTimeout = 5 * time.Second

	// userAgent is the User-Agent used when fetching robots.txt
	userAgent = "jsbug-robots/1.0"

	// botName is the bot name to check rules against
	botName = "Googlebot"
)

// Checker performs robots.txt checks
type Checker struct {
	logger *zap.Logger
}

// NewChecker creates a new Checker
func NewChecker(logger *zap.Logger) *Checker {
	return &Checker{
		logger: logger,
	}
}

// Check determines if the given URL is allowed by robots.txt for Googlebot.
// Returns true (allowed) on any error (fail-open behavior).
func (c *Checker) Check(ctx context.Context, targetURL string) (bool, error) {
	// Parse the target URL
	parsed, err := url.Parse(targetURL)
	if err != nil {
		return true, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Construct robots.txt URL
	robotsURL := fmt.Sprintf("%s://%s/robots.txt", parsed.Scheme, parsed.Host)

	c.logger.Debug("Fetching robots.txt",
		zap.String("robots_url", robotsURL),
		zap.String("target_url", targetURL),
	)

	// Fetch and parse robots.txt using FromResponse
	// This handles status codes automatically (404/5xx = allow all)
	robots, err := c.fetchAndParseRobotsTxt(ctx, robotsURL)
	if err != nil {
		c.logger.Debug("Failed to fetch/parse robots.txt, allowing access",
			zap.String("robots_url", robotsURL),
			zap.Error(err),
		)
		return true, nil // Fail open
	}

	// Check if the URL is allowed for Googlebot
	group := robots.FindGroup(botName)
	allowed := group.Test(parsed.Path)

	c.logger.Debug("Robots.txt check completed",
		zap.String("target_url", targetURL),
		zap.String("path", parsed.Path),
		zap.Bool("is_allowed", allowed),
	)

	return allowed, nil
}

// fetchAndParseRobotsTxt fetches and parses robots.txt using FromResponse
func (c *Checker) fetchAndParseRobotsTxt(ctx context.Context, robotsURL string) (*robotstxt.RobotsData, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, fetchTimeout)
	defer cancel()

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, robotsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)

	// Execute request
	client := &http.Client{
		Timeout: fetchTimeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// FromResponse handles status codes per robots.txt spec:
	// - 2xx: parse the body
	// - 4xx: allow all (no restrictions)
	// - 5xx: allow all (temporary failure)
	robots, err := robotstxt.FromResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse robots.txt: %w", err)
	}

	return robots, nil
}
