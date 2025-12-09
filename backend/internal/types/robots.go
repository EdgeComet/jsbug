package types

import "net/url"

// RobotsRequest represents an API request to check robots.txt
type RobotsRequest struct {
	URL string `json:"url"`
}

// Validate checks if the request has a valid URL
func (r *RobotsRequest) Validate() error {
	if r.URL == "" {
		return &ValidationError{Code: ErrInvalidURL, Message: "URL is required"}
	}

	parsed, err := url.Parse(r.URL)
	if err != nil {
		return &ValidationError{Code: ErrInvalidURL, Message: "Invalid URL format"}
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return &ValidationError{Code: ErrInvalidURL, Message: "URL must use http or https scheme"}
	}

	if parsed.Host == "" {
		return &ValidationError{Code: ErrInvalidURL, Message: "URL must have a host"}
	}

	return nil
}

// ValidationError represents a request validation error
type ValidationError struct {
	Code    string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// RobotsResponse represents the API response for robots.txt check
type RobotsResponse struct {
	Success bool         `json:"success"`
	Data    *RobotsData  `json:"data,omitempty"`
	Error   *RenderError `json:"error,omitempty"`
}

// RobotsData contains the robots.txt check result
type RobotsData struct {
	URL       string `json:"url"`
	IsAllowed bool   `json:"is_allowed"`
}
