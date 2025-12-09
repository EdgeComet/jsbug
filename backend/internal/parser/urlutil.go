package parser

import (
	"errors"
	"net/url"
	"strings"
)

var (
	// ErrInvalidURL is returned when a URL cannot be parsed
	ErrInvalidURL = errors.New("invalid URL")
	// ErrEmptyURL is returned when an empty URL is provided
	ErrEmptyURL = errors.New("empty URL")
)

// ExtractBaseDomain extracts the base domain from a URL.
// It handles URLs with or without scheme, ports, and subdomains.
// Examples:
//   - "https://www.example.com/path" -> "example.com"
//   - "cdn.example.com:8080" -> "example.com"
//   - "192.168.1.1" -> "192.168.1.1" (IP addresses returned as-is)
func ExtractBaseDomain(rawURL string) (string, error) {
	if rawURL == "" {
		return "", ErrEmptyURL
	}

	// Add scheme if missing to help url.Parse
	normalized := rawURL
	if !strings.Contains(rawURL, "://") {
		// Check for protocol-relative URLs
		if strings.HasPrefix(rawURL, "//") {
			normalized = "https:" + rawURL
		} else {
			normalized = "https://" + rawURL
		}
	}

	parsed, err := url.Parse(normalized)
	if err != nil {
		return "", ErrInvalidURL
	}

	host := parsed.Hostname() // Strips port
	if host == "" {
		return "", ErrInvalidURL
	}

	host = strings.ToLower(host)

	// If it's an IP address, return as-is
	if isIPAddress(host) {
		return host, nil
	}

	// Extract base domain (e.g., "example.com" from "www.example.com")
	return getBaseDomain(host), nil
}

// IsSameDomain returns true if both URLs have the same base domain.
// Subdomains are treated as the same domain (cdn.example.com == www.example.com).
// Comparison is case-insensitive.
func IsSameDomain(urlA, urlB string) bool {
	domainA, errA := ExtractBaseDomain(urlA)
	domainB, errB := ExtractBaseDomain(urlB)

	if errA != nil || errB != nil {
		return false
	}

	return domainA == domainB
}

// IsSubdomainOf checks if the URL's domain is a subdomain of (or equal to) parentDomain.
// Examples:
//   - IsSubdomainOf("https://cdn.example.com", "example.com") -> true
//   - IsSubdomainOf("https://example.com", "example.com") -> true
//   - IsSubdomainOf("https://other.com", "example.com") -> false
func IsSubdomainOf(childURL, parentDomain string) bool {
	if childURL == "" || parentDomain == "" {
		return false
	}

	// Extract the host from childURL
	childDomain, err := extractHost(childURL)
	if err != nil {
		return false
	}

	// Normalize both domains
	childDomain = strings.ToLower(childDomain)
	parentDomain = strings.ToLower(strings.TrimSpace(parentDomain))

	// Remove any scheme/path from parentDomain if accidentally included
	if strings.Contains(parentDomain, "://") {
		if pd, err := extractHost(parentDomain); err == nil {
			parentDomain = pd
		}
	}

	// Exact match
	if childDomain == parentDomain {
		return true
	}

	// Check if childDomain ends with ".parentDomain"
	return strings.HasSuffix(childDomain, "."+parentDomain)
}

// extractHost extracts just the hostname (lowercase, no port) from a URL.
func extractHost(rawURL string) (string, error) {
	if rawURL == "" {
		return "", ErrEmptyURL
	}

	// Add scheme if missing
	normalized := rawURL
	if !strings.Contains(rawURL, "://") {
		if strings.HasPrefix(rawURL, "//") {
			normalized = "https:" + rawURL
		} else {
			normalized = "https://" + rawURL
		}
	}

	parsed, err := url.Parse(normalized)
	if err != nil {
		return "", ErrInvalidURL
	}

	host := parsed.Hostname()
	if host == "" {
		return "", ErrInvalidURL
	}

	return strings.ToLower(host), nil
}

// getBaseDomain extracts the base domain from a hostname.
// For "www.example.com" it returns "example.com".
// For "sub.sub.example.co.uk" it returns "example.co.uk".
func getBaseDomain(host string) string {
	// Handle known multi-part TLDs
	multiPartTLDs := map[string]bool{
		"co.uk": true, "org.uk": true, "ac.uk": true, "gov.uk": true,
		"co.jp": true, "co.nz": true, "co.za": true,
		"com.au": true, "net.au": true, "org.au": true,
		"com.br": true, "org.br": true,
	}

	parts := strings.Split(host, ".")
	n := len(parts)

	if n <= 2 {
		return host
	}

	// Check for multi-part TLD
	if n >= 3 {
		possibleTLD := parts[n-2] + "." + parts[n-1]
		if multiPartTLDs[possibleTLD] {
			// Return domain + multi-part TLD (e.g., "example.co.uk")
			if n >= 4 {
				return parts[n-3] + "." + possibleTLD
			}
			return host
		}
	}

	// Standard TLD: return last two parts
	return parts[n-2] + "." + parts[n-1]
}

// isIPAddress checks if the host is an IPv4 or IPv6 address.
func isIPAddress(host string) bool {
	// IPv6 addresses contain colons
	if strings.Contains(host, ":") {
		return true
	}

	// Check for IPv4: all parts are numbers
	parts := strings.Split(host, ".")
	if len(parts) != 4 {
		return false
	}

	for _, part := range parts {
		if part == "" {
			return false
		}
		for _, c := range part {
			if c < '0' || c > '9' {
				return false
			}
		}
	}
	return true
}
