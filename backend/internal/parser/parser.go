package parser

import (
	"encoding/json"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/user/jsbug/internal/types"
)

// ParseResult contains extracted content from HTML
type ParseResult struct {
	Title           string
	MetaDescription string
	CanonicalURL    string
	MetaRobots      string
	H1              []string
	H2              []string
	H3              []string
	WordCount       int
	InternalLinks   int
	ExternalLinks   int
	OpenGraph       map[string]string
	StructuredData  []json.RawMessage
	// New fields for extended extraction
	BodyText      string
	TextHtmlRatio float64
	HrefLangs     []types.HrefLang
	Links         []types.Link
	Images        []types.Image
	MetaIndexable bool
	MetaFollow    bool
}

// ParseOptions contains options for parsing HTML
type ParseOptions struct {
	PageURL    string
	XRobotsTag string
	LinkHeader string
}

// Parser extracts SEO-relevant content from HTML
type Parser struct{}

// NewParser creates a new Parser
func NewParser() *Parser {
	return &Parser{}
}

// Parse extracts content from HTML (backward compatible wrapper)
func (p *Parser) Parse(htmlContent string, pageURL string) (*ParseResult, error) {
	return p.ParseWithOptions(htmlContent, ParseOptions{PageURL: pageURL})
}

// ParseWithOptions extracts content from HTML with additional options
func (p *Parser) ParseWithOptions(htmlContent string, opts ParseOptions) (*ParseResult, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	pageHost := ""
	if u, err := url.Parse(opts.PageURL); err == nil {
		pageHost = u.Host
	}

	result := &ParseResult{
		H1:             make([]string, 0),
		H2:             make([]string, 0),
		H3:             make([]string, 0),
		OpenGraph:      make(map[string]string),
		StructuredData: make([]json.RawMessage, 0),
		HrefLangs:      make([]types.HrefLang, 0),
		Links:          make([]types.Link, 0),
		Images:         make([]types.Image, 0),
		MetaIndexable:  true, // default to true
		MetaFollow:     true, // default to true
	}

	// Extract title (first one wins, normalized)
	result.Title = normalizeWhitespace(doc.Find("title").First().Text())

	// Extract meta tags (description, robots, googlebot)
	var googlebotContent string
	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		name, _ := s.Attr("name")
		content, _ := s.Attr("content")

		switch strings.ToLower(name) {
		case "description":
			if result.MetaDescription == "" {
				result.MetaDescription = normalizeWhitespace(content)
			}
		case "robots":
			if result.MetaRobots == "" {
				result.MetaRobots = content
			}
		case "googlebot":
			if googlebotContent == "" {
				googlebotContent = content
			}
		}
	})

	// Extract Open Graph tags
	doc.Find("meta[property^='og:']").Each(func(i int, s *goquery.Selection) {
		property, _ := s.Attr("property")
		content, _ := s.Attr("content")
		if property != "" {
			ogKey := strings.TrimPrefix(strings.ToLower(property), "og:")
			result.OpenGraph[ogKey] = content
		}
	})

	// Extract canonical URL (case insensitive)
	doc.Find("link").Each(func(i int, s *goquery.Selection) {
		rel, _ := s.Attr("rel")
		href, _ := s.Attr("href")
		if strings.ToLower(rel) == "canonical" && href != "" {
			result.CanonicalURL = href
		}
	})

	// Extract headings (normalized and unique)
	seenH1 := make(map[string]bool)
	seenH2 := make(map[string]bool)
	seenH3 := make(map[string]bool)
	doc.Find("h1").Each(func(i int, s *goquery.Selection) {
		result.H1 = addUniqueHeader(result.H1, s.Text(), seenH1)
	})
	doc.Find("h2").Each(func(i int, s *goquery.Selection) {
		result.H2 = addUniqueHeader(result.H2, s.Text(), seenH2)
	})
	doc.Find("h3").Each(func(i int, s *goquery.Selection) {
		result.H3 = addUniqueHeader(result.H3, s.Text(), seenH3)
	})

	// Extract links (count internal/external for backward compatibility)
	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists || href == "" {
			return
		}

		// Skip javascript:, mailto:, tel:, #anchors
		if strings.HasPrefix(href, "javascript:") ||
			strings.HasPrefix(href, "mailto:") ||
			strings.HasPrefix(href, "tel:") ||
			strings.HasPrefix(href, "#") {
			return
		}

		// Parse the URL to determine if internal or external
		u, err := url.Parse(href)
		if err != nil {
			return
		}

		// Relative URLs are internal
		if u.Host == "" {
			result.InternalLinks++
			return
		}

		// Compare hosts
		if strings.EqualFold(u.Host, pageHost) {
			result.InternalLinks++
		} else {
			result.ExternalLinks++
		}
	})

	// Extract structured data (JSON-LD)
	doc.Find("script[type='application/ld+json']").Each(func(i int, s *goquery.Selection) {
		content := strings.TrimSpace(s.Text())
		if content == "" {
			return
		}

		// Validate JSON
		var js json.RawMessage
		if err := json.Unmarshal([]byte(content), &js); err == nil {
			result.StructuredData = append(result.StructuredData, js)
		}
	})

	// Count words in body (excluding script, style, noscript)
	body := doc.Find("body")
	if body.Length() > 0 {
		// Clone body and remove unwanted elements
		bodyClone := body.Clone()
		bodyClone.Find("script, style, noscript, head").Remove()
		text := bodyClone.Text()
		result.WordCount = countWords(text)
	}

	// === NEW EXTRACTION FUNCTIONS ===

	// Extract robots directives (MetaIndexable, MetaFollow)
	result.MetaIndexable, result.MetaFollow = GetRobotsFromMeta(googlebotContent, result.MetaRobots, opts.XRobotsTag)

	// Extract body text and calculate ratio
	result.BodyText = ExtractBodyText(doc)
	result.TextHtmlRatio = CalculateTextHtmlRatio(result.BodyText, htmlContent)

	// Extract hreflang tags
	result.HrefLangs = ExtractHrefLangs(doc, opts.PageURL, opts.LinkHeader)

	// Extract links with full metadata
	result.Links = ExtractLinks(doc, opts.PageURL)

	// Extract images with full metadata
	result.Images = ExtractImages(doc, opts.PageURL)

	return result, nil
}

// countWords counts words in text
func countWords(text string) int {
	// Normalize whitespace
	text = normalizeWhitespace(text)
	if text == "" {
		return 0
	}
	return len(strings.Fields(text))
}

// normalizeWhitespace replaces multiple whitespace with single space
func normalizeWhitespace(s string) string {
	re := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(re.ReplaceAllString(s, " "))
}

// addUniqueHeader adds a normalized header to slice if not already present
func addUniqueHeader(slice []string, text string, seen map[string]bool) []string {
	normalized := normalizeWhitespace(text)
	if normalized == "" || seen[normalized] {
		return slice
	}
	seen[normalized] = true
	return append(slice, normalized)
}
