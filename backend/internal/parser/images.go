package parser

import (
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/user/jsbug/internal/types"
)

// ExtractImages extracts all images from an HTML document with metadata.
// It skips empty src and data: URIs. Src is always resolved to absolute URL.
func ExtractImages(doc *goquery.Document, pageURL string) []types.Image {
	var images []types.Image

	// Extract page host for internal/external detection
	pageHost, _ := extractHost(pageURL)

	// Parse base URL for resolving relative paths
	baseURL, _ := url.Parse(pageURL)

	doc.Find("img[src]").Each(func(i int, s *goquery.Selection) {
		src, exists := s.Attr("src")
		if !exists || src == "" {
			return
		}

		// Skip data: URIs
		srcLower := strings.ToLower(strings.TrimSpace(src))
		if strings.HasPrefix(srcLower, "data:") {
			return
		}

		alt, _ := s.Attr("alt")

		// Determine if original URL was absolute (before resolving)
		isAbsolute := strings.HasPrefix(srcLower, "http://") ||
			strings.HasPrefix(srcLower, "https://")

		// Resolve to absolute URL
		resolvedSrc := src
		if baseURL != nil && !isAbsolute {
			if resolved, err := baseURL.Parse(src); err == nil {
				resolvedSrc = resolved.String()
			}
		}

		// Determine if external (use resolved URL for accurate detection)
		isExternal := determineExternal(resolvedSrc, pageHost)

		// Check if inside <a> tag
		isInLink := false
		linkHref := ""
		parentLink := s.ParentsFiltered("a").First()
		if parentLink.Length() > 0 {
			isInLink = true
			linkHref, _ = parentLink.Attr("href")
		}

		images = append(images, types.Image{
			Src:        resolvedSrc,
			Alt:        alt,
			IsExternal: isExternal,
			IsAbsolute: isAbsolute,
			IsInLink:   isInLink,
			LinkHref:   linkHref,
		})
	})

	return images
}
