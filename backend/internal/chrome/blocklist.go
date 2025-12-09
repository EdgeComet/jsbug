package chrome

import (
	"strings"
)

// Predefined blocking patterns
var (
	analyticsPatterns = []string{
		"*google-analytics.com*",
		"*googletagmanager.com*",
		"*gtm.js*",
		"*gtag/js*",
		"*hotjar.com*",
		"*segment.com*",
		"*segment.io*",
		"*mixpanel.com*",
		"*amplitude.com*",
		"*heap.io*",
		"*heapanalytics.com*",
		"*plausible.io*",
		"*matomo.*",
		"*piwik.*",
		"*clarity.ms*",
		"*mouseflow.com*",
		"*fullstory.com*",
		"*logrocket.com*",
	}

	adsPatterns = []string{
		"*doubleclick.net*",
		"*googlesyndication.com*",
		"*googleadservices.com*",
		"*adnxs.com*",
		"*criteo.com*",
		"*criteo.net*",
		"*amazon-adsystem.com*",
		"*moatads.com*",
		"*adsrvr.org*",
		"*adroll.com*",
		"*outbrain.com*",
		"*taboola.com*",
	}

	socialPatterns = []string{
		"*facebook.com/tr*",
		"*fbevents.js*",
		"*connect.facebook.net*",
		"*platform.twitter.com*",
		"*ads-twitter.com*",
		"*analytics.twitter.com*",
		"*linkedin.com/px*",
		"*snap.licdn.com*",
		"*tiktok.com/i18n/pixel*",
		"*analytics.tiktok.com*",
		"*ct.pinterest.com*",
		"*pinimg.com/ct*",
	}
)

// Resource types that can be blocked
const (
	ResourceTypeImage      = "image"
	ResourceTypeFont       = "font"
	ResourceTypeMedia      = "media"
	ResourceTypeStylesheet = "stylesheet"
	ResourceTypeScript     = "script"
)

// Blocklist handles URL and resource type blocking
type Blocklist struct {
	patterns     []string
	blockedTypes map[string]bool
}

// NewBlocklist creates a new Blocklist with the specified blocking options
func NewBlocklist(blockAnalytics, blockAds, blockSocial bool, blockedTypes []string) *Blocklist {
	var patterns []string

	if blockAnalytics {
		patterns = append(patterns, analyticsPatterns...)
	}
	if blockAds {
		patterns = append(patterns, adsPatterns...)
	}
	if blockSocial {
		patterns = append(patterns, socialPatterns...)
	}

	blockedTypesMap := make(map[string]bool)
	for _, t := range blockedTypes {
		blockedTypesMap[strings.ToLower(t)] = true
	}

	return &Blocklist{
		patterns:     patterns,
		blockedTypes: blockedTypesMap,
	}
}

// ShouldBlock checks if a URL or resource type should be blocked
func (b *Blocklist) ShouldBlock(url string, resourceType string) bool {
	if b == nil {
		return false
	}

	// Check resource type first
	if b.blockedTypes[strings.ToLower(resourceType)] {
		return true
	}

	// Check URL patterns
	urlLower := strings.ToLower(url)
	for _, pattern := range b.patterns {
		if wildcardMatch(pattern, urlLower) {
			return true
		}
	}

	return false
}

// IsEmpty returns true if no blocking is configured
func (b *Blocklist) IsEmpty() bool {
	if b == nil {
		return true
	}
	return len(b.patterns) == 0 && len(b.blockedTypes) == 0
}

// wildcardMatch performs case-insensitive wildcard matching
// * matches any sequence of characters
func wildcardMatch(pattern, text string) bool {
	pattern = strings.ToLower(pattern)

	// Handle simple cases
	if pattern == "*" {
		return true
	}
	if pattern == "" {
		return text == ""
	}

	// Split pattern by wildcards
	parts := strings.Split(pattern, "*")

	// If pattern doesn't start with *, text must start with first part
	if !strings.HasPrefix(pattern, "*") {
		if !strings.HasPrefix(text, parts[0]) {
			return false
		}
	}

	// If pattern doesn't end with *, text must end with last part
	if !strings.HasSuffix(pattern, "*") {
		if !strings.HasSuffix(text, parts[len(parts)-1]) {
			return false
		}
	}

	// Check all non-empty parts appear in order
	pos := 0
	for _, part := range parts {
		if part == "" {
			continue
		}
		idx := strings.Index(text[pos:], part)
		if idx < 0 {
			return false
		}
		pos += idx + len(part)
	}

	return true
}
