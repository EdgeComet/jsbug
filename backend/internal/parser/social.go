package parser

import "strings"

// socialDomains contains the base domains of major social networks.
// Used to identify social media links in web pages.
var socialDomains = map[string]bool{
	"facebook.com":  true,
	"twitter.com":   true,
	"x.com":         true,
	"linkedin.com":  true,
	"instagram.com": true,
	"youtube.com":   true,
	"tiktok.com":    true,
	"pinterest.com": true,
	"reddit.com":    true,
	"discord.com":   true,
	"snapchat.com":  true,
	"whatsapp.com":  true,
	"telegram.org":  true,
	"tumblr.com":    true,
	"threads.net":   true,
}

// IsSocialURL checks if the given URL points to a social network domain.
// It handles subdomains (e.g., m.facebook.com matches facebook.com).
// Returns false for invalid URLs (does not error).
func IsSocialURL(rawURL string) bool {
	if rawURL == "" {
		return false
	}

	// Skip non-http schemes (mailto:, javascript:, tel:, etc.)
	lowered := strings.ToLower(rawURL)
	if strings.HasPrefix(lowered, "mailto:") ||
		strings.HasPrefix(lowered, "javascript:") ||
		strings.HasPrefix(lowered, "tel:") ||
		strings.HasPrefix(lowered, "data:") {
		return false
	}

	// Extract the host from the URL
	host, err := extractHost(rawURL)
	if err != nil {
		return false
	}

	host = strings.ToLower(host)

	// Check direct match first
	if socialDomains[host] {
		return true
	}

	// Check if it's a subdomain of a social domain
	for domain := range socialDomains {
		if strings.HasSuffix(host, "."+domain) {
			return true
		}
	}

	return false
}
