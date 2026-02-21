package parser

import (
	"fmt"
	"strings"

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

	var sections []types.Section
	var currentHeading *html.Node
	var currentBody strings.Builder
	headingLevel := 0

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
			SectionID:    fmt.Sprintf("s%d", len(sections)+1),
			HeadingLevel: headingLevel,
			HeadingText:  headingText,
			BodyMarkdown: bodyMd,
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
					currentBody.Reset()
				} else if isSectionContainer(tag) {
					walk(c)
				} else {
					currentBody.WriteString(sectionNodeToMarkdown(c))
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
