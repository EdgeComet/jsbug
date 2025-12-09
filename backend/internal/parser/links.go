package parser

import (
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/user/jsbug/internal/types"
)

// ExtractLinks extracts all links from an HTML document with metadata.
// It skips javascript:, mailto:, tel:, and empty hrefs.
func ExtractLinks(doc *goquery.Document, pageURL string) []types.Link {
	var links []types.Link

	// Extract page host for internal/external detection
	pageHost, _ := extractHost(pageURL)

	// Parse base URL for resolving relative paths
	baseURL, _ := url.Parse(pageURL)

	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists || href == "" || strings.HasPrefix(href, "#") {
			return
		}

		// Skip non-http schemes
		hrefLower := strings.ToLower(strings.TrimSpace(href))
		if strings.HasPrefix(hrefLower, "javascript:") ||
			strings.HasPrefix(hrefLower, "mailto:") ||
			strings.HasPrefix(hrefLower, "tel:") ||
			strings.HasPrefix(hrefLower, "data:") {
			return
		}

		// Extract link text
		text := extractLinkText(s)

		// Parse rel attribute for flags
		rel, _ := s.Attr("rel")
		nofollow, ugc, sponsored := parseRelAttribute(rel)

		// Check if link contains an image
		isImageLink := s.Find("img").Length() > 0

		// Determine if URL is absolute
		isAbsolute := strings.HasPrefix(hrefLower, "http://") ||
			strings.HasPrefix(hrefLower, "https://")

		// Resolve to absolute URL
		resolvedHref := href
		if baseURL != nil && !isAbsolute {
			if resolved, err := baseURL.Parse(href); err == nil {
				resolvedHref = resolved.String()
			}
		}

		// Determine if external
		isExternal := determineExternal(href, pageHost)

		// Determine if social
		isSocial := determineSocial(href, isAbsolute, pageURL)

		links = append(links, types.Link{
			Href:        resolvedHref,
			Text:        text,
			IsExternal:  isExternal,
			IsDofollow:  !nofollow,
			IsImageLink: isImageLink,
			IsAbsolute:  isAbsolute,
			IsSocial:    isSocial,
			IsUgc:       ugc,
			IsSponsored: sponsored,
		})
	})

	return links
}

// extractLinkText extracts and normalizes the text content of a link.
func extractLinkText(s *goquery.Selection) string {
	text := s.Text()

	// Collapse whitespace
	text = whitespaceRegex.ReplaceAllString(text, " ")

	// Trim leading/trailing whitespace
	text = strings.TrimSpace(text)

	return text
}

// parseRelAttribute parses the rel attribute and returns flags for nofollow, ugc, and sponsored.
func parseRelAttribute(rel string) (nofollow, ugc, sponsored bool) {
	if rel == "" {
		return false, false, false
	}

	// Split by whitespace and check each token
	relLower := strings.ToLower(rel)
	tokens := strings.Fields(relLower)

	for _, token := range tokens {
		switch token {
		case "nofollow":
			nofollow = true
		case "ugc":
			ugc = true
		case "sponsored":
			sponsored = true
		}
	}

	return
}

// determineExternal checks if a URL is external to the page's domain.
func determineExternal(href, pageHost string) bool {
	if pageHost == "" {
		return false
	}

	hrefLower := strings.ToLower(strings.TrimSpace(href))

	// Relative URLs are internal
	if !strings.HasPrefix(hrefLower, "http://") &&
		!strings.HasPrefix(hrefLower, "https://") &&
		!strings.HasPrefix(hrefLower, "//") {
		return false
	}

	// Extract host from href
	linkHost, err := extractHost(href)
	if err != nil {
		return false
	}

	// Compare hosts (including subdomain check)
	// Link is internal if it's the same domain or a subdomain
	pageHostLower := strings.ToLower(pageHost)
	linkHostLower := strings.ToLower(linkHost)

	// Exact match
	if linkHostLower == pageHostLower {
		return false
	}

	// Check if link is subdomain of page
	if strings.HasSuffix(linkHostLower, "."+pageHostLower) {
		return false
	}

	// Check if page is subdomain of link (both internal)
	if strings.HasSuffix(pageHostLower, "."+linkHostLower) {
		return false
	}

	// Check base domains
	linkBaseDomain, _ := ExtractBaseDomain(href)
	pageBaseDomain, _ := ExtractBaseDomain("https://" + pageHost)

	if linkBaseDomain != "" && pageBaseDomain != "" && linkBaseDomain == pageBaseDomain {
		return false
	}

	return true
}

// determineSocial checks if a URL points to a social network.
func determineSocial(href string, isAbsolute bool, pageURL string) bool {
	// For relative URLs, they can't be social (they're on the same domain)
	if !isAbsolute && !strings.HasPrefix(href, "//") {
		return false
	}

	// For protocol-relative URLs, prepend https:
	checkURL := href
	if strings.HasPrefix(href, "//") {
		checkURL = "https:" + href
	}

	return IsSocialURL(checkURL)
}
