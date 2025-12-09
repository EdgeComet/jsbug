package parser

import "strings"

// ParseRobotsDirectives parses a robots directive content string and returns
// indexable and follow status.
//
// Supported directives (case-insensitive):
//   - "noindex" → indexable=false
//   - "nofollow" → follow=false
//   - "none" → both false
//   - "all" → both true (explicit default)
//   - "index" → indexable=true (explicit)
//   - "follow" → follow=true (explicit)
//
// Unknown directives (e.g., "noarchive", "nosnippet") are ignored.
// Default values are indexable=true, follow=true.
func ParseRobotsDirectives(content string) (indexable bool, follow bool) {
	// Defaults
	indexable = true
	follow = true

	if content == "" {
		return
	}

	// Split by comma and process each directive
	directives := strings.Split(content, ",")
	for _, d := range directives {
		directive := strings.ToLower(strings.TrimSpace(d))

		switch directive {
		case "noindex":
			indexable = false
		case "nofollow":
			follow = false
		case "none":
			indexable = false
			follow = false
		case "all":
			indexable = true
			follow = true
		case "index":
			indexable = true
		case "follow":
			follow = true
			// Ignore unknown directives like "noarchive", "nosnippet", "max-snippet", etc.
		}
	}

	return
}

// GetRobotsFromMeta determines the final indexable and follow status from multiple sources.
//
// Priority order (most specific wins):
//  1. googlebot meta tag (most specific for Google)
//  2. robots meta tag (general)
//  3. X-Robots-Tag header (HTTP header)
//
// When multiple sources exist, more restrictive values win.
// For example: googlebot="index", xRobotsHeader="noindex" → indexable=false
func GetRobotsFromMeta(googlebotContent, robotsContent, xRobotsHeader string) (indexable bool, follow bool) {
	// Start with defaults
	indexable = true
	follow = true

	// Parse all sources
	headerIndexable, headerFollow := ParseRobotsDirectives(xRobotsHeader)
	robotsIndexable, robotsFollow := ParseRobotsDirectives(robotsContent)
	googlebotIndexable, googlebotFollow := ParseRobotsDirectives(googlebotContent)

	// Apply more restrictive wins logic
	// If ANY source says noindex, result is noindex
	// If ANY source says nofollow, result is nofollow

	// Check X-Robots-Tag header
	if xRobotsHeader != "" {
		if !headerIndexable {
			indexable = false
		}
		if !headerFollow {
			follow = false
		}
	}

	// Check robots meta (overrides header if more restrictive)
	if robotsContent != "" {
		if !robotsIndexable {
			indexable = false
		}
		if !robotsFollow {
			follow = false
		}
	}

	// Check googlebot meta (overrides others if more restrictive)
	if googlebotContent != "" {
		if !googlebotIndexable {
			indexable = false
		}
		if !googlebotFollow {
			follow = false
		}
	}

	return
}
