package captcha

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"
)

const turnstileVerifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

// VerifyResponse represents the response from Cloudflare's siteverify API
type VerifyResponse struct {
	Success     bool     `json:"success"`
	ErrorCodes  []string `json:"error-codes,omitempty"`
	ChallengeTS string   `json:"challenge_ts,omitempty"`
	Hostname    string   `json:"hostname,omitempty"`
}

// Verifier handles Turnstile token verification with Cloudflare
type Verifier struct {
	secretKey  string
	httpClient *http.Client
	logger     *zap.Logger
}

// NewVerifier creates a new captcha verifier
func NewVerifier(secretKey string, logger *zap.Logger) *Verifier {
	return &Verifier{
		secretKey: secretKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// Verify validates a Turnstile token with Cloudflare's API
// Returns true if the token is valid, false otherwise
// The remoteIP parameter is optional but improves security
func (v *Verifier) Verify(ctx context.Context, token string, remoteIP string) (bool, error) {
	if token == "" {
		return false, fmt.Errorf("empty token")
	}

	formData := url.Values{
		"secret":   {v.secretKey},
		"response": {token},
	}
	if remoteIP != "" {
		formData.Set("remoteip", remoteIP)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, turnstileVerifyURL,
		strings.NewReader(formData.Encode()))
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to verify token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("cloudflare API returned status %d", resp.StatusCode)
	}

	var verifyResp VerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&verifyResp); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	// Log failures only (per requirements - don't log successes)
	if !verifyResp.Success {
		v.logger.Warn("Captcha verification failed",
			zap.Strings("error_codes", verifyResp.ErrorCodes),
			zap.String("remote_ip", remoteIP),
		)
	}

	return verifyResp.Success, nil
}
