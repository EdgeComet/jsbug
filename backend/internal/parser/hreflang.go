package parser

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/user/jsbug/internal/types"
)

// linkHeaderRegex parses Link header entries
// Format: <url>; rel="alternate"; hreflang="en"
var linkHeaderRegex = regexp.MustCompile(`<([^>]+)>(?:\s*;\s*([^,]+))?`)
var relAlternateRegex = regexp.MustCompile(`(?i)rel\s*=\s*"?alternate"?`)
var hreflangRegex = regexp.MustCompile(`(?i)hreflang\s*=\s*"?([^";,\s]+)"?`)

// ExtractHrefLangs extracts hreflang alternate links from HTML and HTTP headers.
// It combines results from both sources, deduplicating by lang+url (HTML preferred).
func ExtractHrefLangs(doc *goquery.Document, pageURL, linkHeader string) []types.HrefLang {
	// Extract from HTML first (higher priority)
	htmlHrefLangs := parseHTMLHrefLangs(doc, pageURL)

	// Extract from HTTP Link header
	headerHrefLangs := parseLinkHeaderHrefLangs(linkHeader, pageURL)

	// Deduplicate: HTML takes priority over header
	seen := make(map[string]bool)
	result := make([]types.HrefLang, 0, len(htmlHrefLangs)+len(headerHrefLangs))

	// Add HTML hreflangs first
	for _, hl := range htmlHrefLangs {
		key := strings.ToLower(hl.Lang) + "|" + hl.URL
		if !seen[key] {
			seen[key] = true
			result = append(result, hl)
		}
	}

	// Add header hreflangs (only if not duplicate)
	for _, hl := range headerHrefLangs {
		key := strings.ToLower(hl.Lang) + "|" + hl.URL
		if !seen[key] {
			seen[key] = true
			result = append(result, hl)
		}
	}

	return result
}

// parseHTMLHrefLangs extracts hreflang from HTML <link rel="alternate" hreflang="xx" href="..."> tags.
func parseHTMLHrefLangs(doc *goquery.Document, pageURL string) []types.HrefLang {
	var result []types.HrefLang

	doc.Find(`link[hreflang]`).Each(func(i int, s *goquery.Selection) {
		// Check rel="alternate" case-insensitively
		rel, _ := s.Attr("rel")
		if !strings.EqualFold(rel, "alternate") {
			return
		}

		hreflang, hasHreflang := s.Attr("hreflang")
		href, hasHref := s.Attr("href")

		if !hasHreflang || !hasHref || hreflang == "" || href == "" {
			return
		}

		// Resolve relative URL to absolute
		absoluteURL := resolveURL(href, pageURL)

		result = append(result, types.HrefLang{
			Lang:   hreflang,
			URL:    absoluteURL,
			Source: "link",
		})
	})

	return result
}

// parseLinkHeaderHrefLangs extracts hreflang from HTTP Link header.
// Format: <url>; rel="alternate"; hreflang="en", <url2>; rel="alternate"; hreflang="fr"
func parseLinkHeaderHrefLangs(linkHeader, pageURL string) []types.HrefLang {
	if linkHeader == "" {
		return nil
	}

	var result []types.HrefLang

	// Split entries properly - commas separate entries, but only outside of <>
	entries := splitLinkHeaderEntries(linkHeader)

	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		// Extract URL from <url>
		urlMatch := linkHeaderRegex.FindStringSubmatch(entry)
		if len(urlMatch) < 2 {
			continue
		}
		rawURL := urlMatch[1] // Use captured group, not full match

		// Check for rel="alternate"
		if !relAlternateRegex.MatchString(entry) {
			continue
		}

		// Extract hreflang value
		hreflangMatch := hreflangRegex.FindStringSubmatch(entry)
		if len(hreflangMatch) < 2 {
			continue
		}
		hreflang := hreflangMatch[1]

		// Resolve relative URL
		absoluteURL := resolveURL(rawURL, pageURL)

		result = append(result, types.HrefLang{
			Lang:   hreflang,
			URL:    absoluteURL,
			Source: "header",
		})
	}

	return result
}

// splitLinkHeaderEntries splits a Link header into individual entries,
// respecting that commas inside <> are part of URLs, not separators.
func splitLinkHeaderEntries(header string) []string {
	var entries []string
	var current strings.Builder
	inBrackets := false

	for _, ch := range header {
		switch ch {
		case '<':
			inBrackets = true
			current.WriteRune(ch)
		case '>':
			inBrackets = false
			current.WriteRune(ch)
		case ',':
			if inBrackets {
				current.WriteRune(ch)
			} else {
				entries = append(entries, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(ch)
		}
	}

	// Don't forget the last entry
	if current.Len() > 0 {
		entries = append(entries, current.String())
	}

	return entries
}

// resolveURL resolves a potentially relative URL against a base URL.
func resolveURL(href, baseURL string) string {
	if href == "" {
		return ""
	}

	// Already absolute
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}

	// Protocol-relative
	if strings.HasPrefix(href, "//") {
		base, err := url.Parse(baseURL)
		if err != nil {
			return href
		}
		return base.Scheme + ":" + href
	}

	// Parse base URL
	base, err := url.Parse(baseURL)
	if err != nil {
		return href
	}

	// Parse relative reference
	ref, err := url.Parse(href)
	if err != nil {
		return href
	}

	// Resolve against base
	resolved := base.ResolveReference(ref)
	return resolved.String()
}
