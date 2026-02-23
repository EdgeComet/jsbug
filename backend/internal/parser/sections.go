package parser

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	"github.com/user/jsbug/internal/types"
	"golang.org/x/net/html"
)

// ExtractSections splits an HTML document into heading-delimited sections.
// Each section has a sequential ID, heading level, heading text, and markdown body.
func ExtractSections(doc *goquery.Document) []types.Section {
	clone := doc.Clone()
	clone.Find("script, style, noscript, iframe, head, template").Remove()
	clone.Find("nav, header, footer").Remove()

	body := clone.Find("body")
	if body.Length() == 0 {
		return nil
	}

	deepest := maxHeadingLevel(body.Get(0))
	fallbackLevel := 2
	if deepest > 0 {
		fallbackLevel = deepest + 1
		if fallbackLevel > 6 {
			fallbackLevel = 6
		}
	}

	var sections []types.Section
	var currentHeading *html.Node
	var currentBody strings.Builder
	headingLevel := 0
	detectionMethod := ""

	finalize := func() {
		headingText := ""
		if currentHeading != nil {
			headingText = strings.TrimSpace(getTextContent(currentHeading))
		}
		bodyMd := strings.TrimSpace(normalizeMarkdown(currentBody.String()))

		if headingText == "" && bodyMd == "" {
			return
		}

		sections = append(sections, types.Section{
			SectionID:       fmt.Sprintf("s%d", len(sections)+1),
			HeadingLevel:    headingLevel,
			HeadingText:     headingText,
			BodyMarkdown:    bodyMd,
			DetectionMethod: detectionMethod,
		})
	}

	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode {
				tag := strings.ToLower(c.Data)
				if isSectionHeading(tag) {
					finalize()
					currentHeading = c
					headingLevel = int(tag[1] - '0')
					detectionMethod = "semantic"
					currentBody.Reset()
				} else if ok, ariaLevel := isARIAHeading(c); ok {
					finalize()
					currentHeading = c
					headingLevel = ariaLevel
					detectionMethod = "aria"
					currentBody.Reset()
				} else if isSectionContainer(tag) {
					if ok, vLevel, vMethod := isVisualHeading(c, fallbackLevel); ok {
						finalize()
						currentHeading = c
						headingLevel = vLevel
						detectionMethod = vMethod
						currentBody.Reset()
					} else {
						walk(c)
					}
				} else {
					if ok, vLevel, vMethod := isVisualHeading(c, fallbackLevel); ok {
						finalize()
						currentHeading = c
						headingLevel = vLevel
						detectionMethod = vMethod
						currentBody.Reset()
					} else {
						currentBody.WriteString(sectionNodeToMarkdown(c))
					}
				}
			} else if c.Type == html.TextNode {
				text := strings.TrimSpace(c.Data)
				if text != "" {
					currentBody.WriteString(text)
					currentBody.WriteString("\n")
				}
			}
		}
	}

	walk(body.Get(0))
	finalize()

	return sections
}

func isSectionHeading(tag string) bool {
	return len(tag) == 2 && tag[0] == 'h' && tag[1] >= '1' && tag[1] <= '6'
}

func isSectionContainer(tag string) bool {
	switch tag {
	case "div", "section", "article", "main", "aside", "figure", "details", "summary":
		return true
	}
	return false
}

// parseAriaLevel reads the aria-level attribute from a node and returns it as
// an integer clamped to the 1-6 range. If the attribute is missing or
// non-numeric the WAI-ARIA default of 2 is returned.
func parseAriaLevel(n *html.Node) int {
	raw := getAttr(n, "aria-level")
	if raw == "" {
		return 2
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 2
	}
	if v <= 0 {
		return 1
	}
	if v >= 7 {
		return 6
	}
	return v
}

// isBlockElement returns true when tag is an inline-or-block element that can
// serve as a visual heading container.
func isBlockElement(tag string) bool {
	switch tag {
	case "div", "p", "span", "a", "strong", "b", "label", "figcaption":
		return true
	}
	return false
}

// classMatchesHeading returns true if the class string contains the
// substring "heading" (case-insensitive).
func classMatchesHeading(class string) bool {
	return strings.Contains(strings.ToLower(class), "heading")
}

// dataAttrMatchesHeading returns true when the node carries a data-heading
// attribute with a truthy value or a data-role="heading" attribute.
func dataAttrMatchesHeading(n *html.Node) bool {
	dh := getAttr(n, "data-heading")
	if dh != "" {
		v := strings.ToLower(dh)
		if v != "false" && v != "0" && v != "no" {
			return true
		}
	}
	if strings.EqualFold(getAttr(n, "data-role"), "heading") {
		return true
	}
	return false
}

// hasChildHeadingOrARIA recursively checks all descendants of n (not n itself)
// and returns true if any descendant is an ElementNode that is either an h1-h6
// tag or has role="heading".
func hasChildHeadingOrARIA(n *html.Node) bool {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode {
			tag := strings.ToLower(c.Data)
			if isSectionHeading(tag) {
				return true
			}
			if strings.EqualFold(getAttr(c, "role"), "heading") {
				return true
			}
		}
		if hasChildHeadingOrARIA(c) {
			return true
		}
	}
	return false
}

// maxHeadingLevel recursively walks the entire subtree under root and returns
// the highest (deepest/largest number) heading level found. It considers both
// h1-h6 tags and elements with role="heading" (using parseAriaLevel for the
// level). Returns 0 if no headings are found.
func maxHeadingLevel(root *html.Node) int {
	max := 0
	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode {
				tag := strings.ToLower(c.Data)
				if isSectionHeading(tag) {
					level := int(tag[1] - '0')
					if level > max {
						max = level
					}
				} else if strings.EqualFold(getAttr(c, "role"), "heading") {
					level := parseAriaLevel(c)
					if level > max {
						max = level
					}
				}
			}
			walk(c)
		}
	}
	walk(root)
	return max
}

// isARIAHeading returns true (and the resolved heading level) when the node
// has role="heading" and is not a native h1-h6 element. It applies container,
// text-length, and child-heading filters to avoid false positives.
func isARIAHeading(n *html.Node) (bool, int) {
	if n.Type != html.ElementNode {
		return false, 0
	}
	tag := strings.ToLower(n.Data)
	if isSectionHeading(tag) {
		return false, 0
	}
	if !strings.EqualFold(getAttr(n, "role"), "heading") {
		return false, 0
	}
	// Container filter: if the tag is a section container, reject nodes with
	// more than one direct child element (likely a wrapper, not a heading).
	if isSectionContainer(tag) {
		count := 0
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode {
				count++
			}
		}
		if count > 1 {
			return false, 0
		}
	}
	// Text filter: reject empty or overly long text.
	text := getTextContent(n)
	if text == "" || utf8.RuneCountInString(text) > 200 {
		return false, 0
	}
	// Child heading filter.
	if hasChildHeadingOrARIA(n) {
		return false, 0
	}
	return true, parseAriaLevel(n)
}

// isVisualHeading returns true (along with the heading level and detection
// method) when a block-level element looks like a heading based on its CSS
// class or data attributes, but is not a native h1-h6 element.
func isVisualHeading(n *html.Node, fallbackLevel int) (bool, int, string) {
	if n.Type != html.ElementNode {
		return false, 0, ""
	}
	tag := strings.ToLower(n.Data)
	if isSectionHeading(tag) {
		return false, 0, ""
	}
	if !isBlockElement(tag) {
		return false, 0, ""
	}
	// Check signals.
	classMatch := classMatchesHeading(getAttr(n, "class"))
	dataMatch := dataAttrMatchesHeading(n)
	if !classMatch && !dataMatch {
		return false, 0, ""
	}
	// Text filter.
	text := getTextContent(n)
	if text == "" || utf8.RuneCountInString(text) > 200 {
		return false, 0, ""
	}
	// Child heading filter.
	if hasChildHeadingOrARIA(n) {
		return false, 0, ""
	}
	method := "data_attr"
	if classMatch {
		method = "css_class"
	}
	return true, fallbackLevel, method
}

func sectionNodeToMarkdown(n *html.Node) string {
	if n == nil {
		return ""
	}
	tag := strings.ToLower(n.Data)
	var buf strings.Builder

	switch tag {
	case "p":
		text := getFormattedTextContent(n)
		if text != "" {
			buf.WriteString(text)
			buf.WriteString("\n\n")
		}
	case "ul":
		extractList(n, &buf, false, nil)
		buf.WriteString("\n")
	case "ol":
		extractList(n, &buf, true, nil)
		buf.WriteString("\n")
	case "blockquote":
		text := getFormattedTextContent(n)
		if text != "" {
			lines := strings.Split(text, "\n")
			for _, line := range lines {
				buf.WriteString("> ")
				buf.WriteString(strings.TrimSpace(line))
				buf.WriteString("\n")
			}
			buf.WriteString("\n")
		}
	case "table":
		extractTable(n, &buf)
		buf.WriteString("\n")
	case "pre":
		buf.WriteString("```\n")
		buf.WriteString(getTextContent(n))
		buf.WriteString("\n```\n\n")
	default:
		text := getFormattedTextContent(n)
		if text != "" {
			buf.WriteString(text)
			buf.WriteString("\n")
		}
	}
	return buf.String()
}
