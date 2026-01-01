package parser

import (
	"fmt"
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

// semanticElements maps HTML5 semantic elements to their labels
var semanticElements = map[string]string{
	"nav":     "NAV",
	"header":  "HEADER",
	"main":    "MAIN CONTENT",
	"article": "ARTICLE",
	"section": "SECTION",
	"aside":   "ASIDE",
	"footer":  "FOOTER",
}

// ariaRoleMap maps ARIA roles to semantic labels
var ariaRoleMap = map[string]string{
	"navigation":    "NAV",
	"main":          "MAIN CONTENT",
	"banner":        "HEADER",
	"contentinfo":   "FOOTER",
	"complementary": "ASIDE",
	"article":       "ARTICLE",
	"region":        "SECTION",
}

// hasSemanticDescendant checks if a node has any semantic element descendants
func hasSemanticDescendant(n *html.Node) bool {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode {
			tag := strings.ToLower(c.Data)
			if _, isSemantic := semanticElements[tag]; isSemantic {
				return true
			}
			role := getAttr(c, "role")
			if _, isSemantic := ariaRoleMap[role]; isSemantic {
				return true
			}
			if hasSemanticDescendant(c) {
				return true
			}
		}
	}
	return false
}

// skipElements are elements whose content should be skipped during text extraction
var skipElements = map[string]bool{
	"svg": true,
}

// extractText recursively extracts text from an HTML node tree.
// It adds a space before and after ALL elements to ensure proper word separation.
// Excess spaces are cleaned up by whitespace normalization later.
func extractText(n *html.Node, buf *strings.Builder) {
	// Add space before element
	if n.Type == html.ElementNode {
		buf.WriteString(" ")
		// Skip content of certain elements (like svg) but preserve word boundary
		if skipElements[strings.ToLower(n.Data)] {
			return
		}
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

	// Remove non-visible elements (svg is handled during text extraction to preserve word boundaries)
	clonedDoc.Find("script, style, noscript, iframe, head").Remove()

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

// ExtractBodyMarkdown extracts structured markdown from HTML document
func ExtractBodyMarkdown(doc *goquery.Document) string {
	if doc == nil {
		return ""
	}

	clonedDoc := doc.Clone()
	// Remove non-visible elements (svg is handled during text extraction to preserve word boundaries)
	clonedDoc.Find("script, style, noscript, iframe, head").Remove()

	var buf strings.Builder

	body := clonedDoc.Find("body")
	if body.Length() == 0 {
		return ""
	}

	// Track semantic section counts
	sectionCounts := make(map[string]int)

	extractMarkdownWithSections(body.Get(0), &buf, sectionCounts)

	result := normalizeMarkdown(buf.String())

	// Truncate to MaxBodyMarkdownBytes at block boundary
	return truncateMarkdownAtBlock(result, types.MaxBodyMarkdownBytes)
}

// extractMarkdown is a convenience wrapper for extractMarkdownWithSections
func extractMarkdown(n *html.Node, buf *strings.Builder) {
	extractMarkdownWithSections(n, buf, nil)
}

// extractMarkdownWithSections recursively processes DOM nodes and writes markdown
func extractMarkdownWithSections(n *html.Node, buf *strings.Builder, sectionCounts map[string]int) {
	if n == nil {
		return
	}

	if n.Type == html.TextNode {
		text := strings.TrimSpace(n.Data)
		if text != "" {
			buf.WriteString(text)
		}
		return
	}

	if n.Type != html.ElementNode {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractMarkdownWithSections(c, buf, sectionCounts)
		}
		return
	}

	tag := strings.ToLower(n.Data)

	// Skip elements that should not be included in markdown output
	if skipElements[tag] {
		return
	}

	// Check for semantic element or ARIA role
	label, isSemantic := semanticElements[tag]
	if !isSemantic {
		role := getAttr(n, "role")
		label, isSemantic = ariaRoleMap[role]
	}

	if isSemantic && sectionCounts != nil {
		// Check if section has content
		content := getTextContent(n)
		if strings.TrimSpace(content) == "" {
			return // Skip empty sections
		}

		// Only output label for innermost semantic elements (no semantic descendants)
		if hasSemanticDescendant(n) {
			// Has nested semantic elements - just process children without label
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				extractMarkdownWithSections(c, buf, sectionCounts)
			}
			return
		}

		// Update count and build label
		sectionCounts[label]++
		count := sectionCounts[label]

		fullLabel := label
		// Check if we'll have multiple of this type (look ahead not possible, so always number if count > 1)
		if count > 1 {
			fullLabel = fmt.Sprintf("%s %d", label, count)
		}

		// Check for aria-label
		ariaLabel := getAttr(n, "aria-label")
		if ariaLabel != "" {
			fullLabel = fmt.Sprintf("%s: %s", label, ariaLabel)
		}

		buf.WriteString("\n\n---\n")
		buf.WriteString("[")
		buf.WriteString(fullLabel)
		buf.WriteString("]\n")

		// Process children
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractMarkdownWithSections(c, buf, sectionCounts)
		}

		buf.WriteString("\n")
		return
	}

	switch tag {
	case "h1", "h2", "h3", "h4", "h5", "h6":
		level := int(tag[1] - '0')
		prefix := strings.Repeat("#", level)
		text := getFormattedTextContent(n)
		if text != "" {
			buf.WriteString("\n\n")
			buf.WriteString(prefix)
			buf.WriteString(" ")
			buf.WriteString(text)
			buf.WriteString("\n\n")
		}
	case "p":
		text := getFormattedTextContent(n)
		if text != "" {
			buf.WriteString("\n\n")
			buf.WriteString(text)
			buf.WriteString("\n\n")
		}
	case "ul":
		buf.WriteString("\n\n")
		extractList(n, buf, false, nil)
		buf.WriteString("\n")
	case "ol":
		buf.WriteString("\n\n")
		extractList(n, buf, true, nil)
		buf.WriteString("\n")
	case "li":
		// Orphan li without parent list - treat as bullet
		text := getTextContent(n)
		if text != "" {
			buf.WriteString("\n- ")
			buf.WriteString(text)
		}
	case "dl":
		buf.WriteString("\n\n")
		extractDefinitionList(n, buf)
		buf.WriteString("\n")
	case "blockquote":
		text := getFormattedTextContent(n)
		if text != "" {
			buf.WriteString("\n\n")
			// Prefix each line with >
			lines := strings.Split(text, "\n")
			for i, line := range lines {
				buf.WriteString("> ")
				buf.WriteString(strings.TrimSpace(line))
				if i < len(lines)-1 {
					buf.WriteString("\n")
				}
			}
			buf.WriteString("\n\n")
		}
	case "table":
		buf.WriteString("\n\n")
		extractTable(n, buf)
		buf.WriteString("\n")
	default:
		// Only add spacing if element has text content (for word separation)
		hasContent := strings.TrimSpace(getTextContent(n)) != ""
		if hasContent {
			buf.WriteString(" ")
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractMarkdownWithSections(c, buf, sectionCounts)
		}
		if hasContent {
			buf.WriteString(" ")
		}
	}
}

// getTextContent extracts all text content from an element
func getTextContent(n *html.Node) string {
	var buf strings.Builder
	extractPlainText(n, &buf)
	return strings.TrimSpace(buf.String())
}

// extractPlainText recursively extracts plain text from nodes
// Adds spaces around elements to ensure proper word separation
func extractPlainText(n *html.Node, buf *strings.Builder) {
	if n.Type == html.TextNode {
		buf.WriteString(n.Data)
		return
	}
	if n.Type == html.ElementNode {
		buf.WriteString(" ")
		// Skip content of certain elements (like svg) but preserve word boundary
		if skipElements[strings.ToLower(n.Data)] {
			return
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractPlainText(c, buf)
	}
	if n.Type == html.ElementNode {
		buf.WriteString(" ")
	}
}

// formattingState tracks current inline formatting context
type formattingState struct {
	bold   bool
	italic bool
}

// getFormattedTextContent extracts text with inline formatting
func getFormattedTextContent(n *html.Node) string {
	var buf strings.Builder
	extractInlineContent(n, &buf, formattingState{})
	return strings.TrimSpace(buf.String())
}

// extractInlineContent processes nodes with inline formatting awareness
func extractInlineContent(n *html.Node, buf *strings.Builder, state formattingState) {
	if n == nil {
		return
	}

	if n.Type == html.TextNode {
		text := n.Data
		// Normalize internal whitespace but preserve leading/trailing spaces
		// if original had them (to maintain word separation)
		hasLeadingSpace := len(text) > 0 && (text[0] == ' ' || text[0] == '\t' || text[0] == '\n')
		hasTrailingSpace := len(text) > 0 && (text[len(text)-1] == ' ' || text[len(text)-1] == '\t' || text[len(text)-1] == '\n')

		normalized := strings.Join(strings.Fields(text), " ")
		if normalized != "" {
			if hasLeadingSpace {
				buf.WriteString(" ")
			}
			buf.WriteString(normalized)
			if hasTrailingSpace {
				buf.WriteString(" ")
			}
		} else if hasLeadingSpace || hasTrailingSpace {
			// Whitespace-only node - preserve one space
			buf.WriteString(" ")
		}
		return
	}

	if n.Type != html.ElementNode {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractInlineContent(c, buf, state)
		}
		return
	}

	tag := strings.ToLower(n.Data)

	// Skip elements like svg, outputting space to preserve word boundary
	if skipElements[tag] {
		buf.WriteString(" ")
		return
	}

	switch tag {
	case "strong", "b":
		if !state.bold {
			buf.WriteString("**")
			newState := state
			newState.bold = true
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				extractInlineContent(c, buf, newState)
			}
			buf.WriteString("**")
		} else {
			// Already bold, don't double
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				extractInlineContent(c, buf, state)
			}
		}
	case "em", "i":
		if !state.italic {
			buf.WriteString("*")
			newState := state
			newState.italic = true
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				extractInlineContent(c, buf, newState)
			}
			buf.WriteString("*")
		} else {
			// Already italic, don't double
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				extractInlineContent(c, buf, state)
			}
		}
	case "a":
		href := getAttr(n, "href")
		// Extract text content from children, but process inline formatting
		var textBuf strings.Builder
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractInlineContent(c, &textBuf, state)
		}
		text := strings.TrimSpace(textBuf.String())

		if text == "" {
			return
		}

		// Only create link for http/https URLs
		if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
			// Escape parentheses in URL (both must be escaped for valid markdown)
			escapedHref := strings.ReplaceAll(href, "(", "%28")
			escapedHref = strings.ReplaceAll(escapedHref, ")", "%29")
			// Escape brackets in text
			escapedText := strings.ReplaceAll(text, "[", "\\[")
			escapedText = strings.ReplaceAll(escapedText, "]", "\\]")
			buf.WriteString(fmt.Sprintf("[%s](%s)", escapedText, escapedHref))
		} else {
			// Non-http links: just output text
			buf.WriteString(text)
		}
	default:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractInlineContent(c, buf, state)
		}
	}
}

// normalizeMarkdown cleans up extra whitespace in markdown
func normalizeMarkdown(md string) string {
	// Replace 3+ newlines with 2
	re := regexp.MustCompile(`\n{3,}`)
	md = re.ReplaceAllString(md, "\n\n")
	// Collapse multiple spaces (not newlines) to single space
	spaceRe := regexp.MustCompile(` {2,}`)
	md = spaceRe.ReplaceAllString(md, " ")
	return strings.TrimSpace(md)
}

// truncateMarkdownAtBlock truncates markdown at a block boundary (double newline)
// to avoid cutting content mid-paragraph. Appends truncation indicator when truncated.
func truncateMarkdownAtBlock(markdown string, maxBytes int) string {
	if len(markdown) <= maxBytes {
		return markdown
	}

	truncationIndicator := "\n\n[Content truncated...]"
	targetLen := maxBytes - len(truncationIndicator)

	if targetLen <= 0 {
		return truncationIndicator
	}

	// Find last double newline before target length
	candidate := markdown[:targetLen]
	lastBlock := strings.LastIndex(candidate, "\n\n")

	if lastBlock > 0 {
		return markdown[:lastBlock] + truncationIndicator
	}

	// No block boundary found, use UTF-8 safe truncation
	return truncateUTF8(markdown, targetLen) + truncationIndicator
}

// extractList processes list elements recursively, tracking ordered vs unordered
func extractList(n *html.Node, buf *strings.Builder, ordered bool, counter *int) {
	if counter == nil {
		c := 0
		counter = &c
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type != html.ElementNode {
			continue
		}

		tag := strings.ToLower(c.Data)

		switch tag {
		case "li":
			*counter++
			// Get direct text content (not from nested lists)
			text := getDirectTextContent(c)
			if text != "" {
				if ordered {
					buf.WriteString(fmt.Sprintf("%d. %s\n", *counter, text))
				} else {
					buf.WriteString(fmt.Sprintf("- %s\n", text))
				}
			}
			// Process nested lists
			for gc := c.FirstChild; gc != nil; gc = gc.NextSibling {
				if gc.Type == html.ElementNode {
					gcTag := strings.ToLower(gc.Data)
					if gcTag == "ul" {
						nestedCounter := 0
						extractList(gc, buf, false, &nestedCounter)
					} else if gcTag == "ol" {
						nestedCounter := 0
						extractList(gc, buf, true, &nestedCounter)
					}
				}
			}
		case "ul":
			nestedCounter := 0
			extractList(c, buf, false, &nestedCounter)
		case "ol":
			nestedCounter := 0
			extractList(c, buf, true, &nestedCounter)
		}
	}
}

// getDirectTextContent gets text content excluding nested lists
func getDirectTextContent(n *html.Node) string {
	var buf strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			buf.WriteString(c.Data)
		} else if c.Type == html.ElementNode {
			tag := strings.ToLower(c.Data)
			// Skip nested lists
			if tag != "ul" && tag != "ol" {
				extractPlainText(c, &buf)
			}
		}
	}
	return strings.TrimSpace(buf.String())
}

// getAttr gets an attribute value from an element node
func getAttr(n *html.Node, name string) string {
	for _, attr := range n.Attr {
		if strings.ToLower(attr.Key) == name {
			return attr.Val
		}
	}
	return ""
}

// extractDefinitionList processes <dl> elements
func extractDefinitionList(n *html.Node, buf *strings.Builder) {
	var currentTerm string
	var definitions []string

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type != html.ElementNode {
			continue
		}

		tag := strings.ToLower(c.Data)

		switch tag {
		case "dt":
			// If we have a pending term with definitions, write it
			if currentTerm != "" && len(definitions) > 0 {
				buf.WriteString(fmt.Sprintf("- **%s**: %s\n", currentTerm, strings.Join(definitions, "; ")))
			}
			currentTerm = getTextContent(c)
			definitions = nil
		case "dd":
			text := getTextContent(c)
			if text != "" {
				definitions = append(definitions, text)
			}
		}
	}

	// Write final term if exists
	if currentTerm != "" && len(definitions) > 0 {
		buf.WriteString(fmt.Sprintf("- **%s**: %s\n", currentTerm, strings.Join(definitions, "; ")))
	}
}

// extractTable converts table to pipe-separated list format
func extractTable(n *html.Node, buf *strings.Builder) {
	// Find all rows (tr elements)
	rows := findElements(n, "tr")

	for _, row := range rows {
		cells := findElements(row, "th", "td")
		var cellTexts []string

		for _, cell := range cells {
			text := getFormattedTextContent(cell)
			if text != "" {
				cellTexts = append(cellTexts, text)
			}
		}

		if len(cellTexts) > 0 {
			buf.WriteString("- ")
			buf.WriteString(strings.Join(cellTexts, " | "))
			buf.WriteString("\n")
		}
	}
}

// findElements finds all descendant elements with given tag names
func findElements(n *html.Node, tags ...string) []*html.Node {
	var result []*html.Node
	tagSet := make(map[string]bool)
	for _, t := range tags {
		tagSet[t] = true
	}

	var find func(*html.Node)
	find = func(node *html.Node) {
		if node.Type == html.ElementNode && tagSet[strings.ToLower(node.Data)] {
			result = append(result, node)
			return // Don't recurse into matched elements
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			find(c)
		}
	}
	find(n)
	return result
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
