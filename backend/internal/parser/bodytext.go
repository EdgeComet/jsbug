package parser

import (
	"math"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	"github.com/user/jsbug/internal/types"
	"golang.org/x/net/html"
)

// whitespaceRegex matches multiple whitespace characters
var whitespaceRegex = regexp.MustCompile(`\s+`)

// extractText recursively extracts text from an HTML node tree.
// It adds a space before and after ALL elements to ensure proper word separation.
// Excess spaces are cleaned up by whitespace normalization later.
func extractText(n *html.Node, buf *strings.Builder) {
	// Add space before element
	if n.Type == html.ElementNode {
		buf.WriteString(" ")
	}
	if n.Type == html.TextNode {
		buf.WriteString(n.Data)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractText(c, buf)
	}
	// Add space after element
	if n.Type == html.ElementNode {
		buf.WriteString(" ")
	}
}

// ExtractBodyText extracts visible text content from an HTML document.
// It removes script, style, noscript, iframe, svg, and head elements,
// then extracts all remaining visible text.
// The result is normalized (whitespace collapsed) and truncated to MaxBodyTextBytes.
func ExtractBodyText(doc *goquery.Document) string {
	// Clone the document to avoid modifying the original
	clonedDoc := doc.Clone()

	// Remove non-visible elements
	clonedDoc.Find("script, style, noscript, iframe, svg, head").Remove()

	// Extract text from body (or entire document if no body)
	var buf strings.Builder
	body := clonedDoc.Find("body")
	if body.Length() > 0 {
		for _, node := range body.Nodes {
			extractText(node, &buf)
		}
	} else {
		for _, node := range clonedDoc.Nodes {
			extractText(node, &buf)
		}
	}
	text := buf.String()

	// Normalize whitespace: collapse multiple spaces/newlines to single space
	text = whitespaceRegex.ReplaceAllString(text, " ")

	// Trim leading/trailing whitespace
	text = strings.TrimSpace(text)

	// Truncate to MaxBodyTextBytes if needed
	if len(text) > types.MaxBodyTextBytes {
		text = truncateUTF8(text, types.MaxBodyTextBytes)
	}

	return text
}

// truncateUTF8 truncates a string to maxBytes while ensuring valid UTF-8.
// It won't cut in the middle of a multi-byte character.
func truncateUTF8(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}

	// Find the last valid UTF-8 boundary at or before maxBytes
	truncated := s[:maxBytes]

	// Walk backwards to find a valid UTF-8 boundary
	for len(truncated) > 0 && !utf8.ValidString(truncated) {
		truncated = truncated[:len(truncated)-1]
	}

	// Also ensure we don't cut in the middle of a rune
	// by checking if the last byte is a continuation byte
	for len(truncated) > 0 {
		lastByte := truncated[len(truncated)-1]
		// UTF-8 continuation bytes start with 10xxxxxx (0x80-0xBF)
		if lastByte >= 0x80 && lastByte <= 0xBF {
			// This is a continuation byte, need to include the start byte
			// Check if string is valid
			if utf8.ValidString(truncated) {
				break
			}
			truncated = truncated[:len(truncated)-1]
		} else {
			break
		}
	}

	return truncated
}

// CalculateTextHtmlRatio calculates the ratio of text length to HTML length.
// Returns 0.0 if html is empty (to avoid division by zero).
// The result is rounded to 4 decimal places.
func CalculateTextHtmlRatio(bodyText, html string) float64 {
	if len(html) == 0 {
		return 0.0
	}

	ratio := float64(len(bodyText)) / float64(len(html))

	// Round to 4 decimal places
	return math.Round(ratio*10000) / 10000
}
