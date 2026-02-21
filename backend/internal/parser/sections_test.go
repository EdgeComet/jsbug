package parser

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func parseHTML(t *testing.T, htmlStr string) *goquery.Document {
	t.Helper()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlStr))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}
	return doc
}

func TestExtractSections_NoHeadings(t *testing.T) {
	doc := parseHTML(t, `<html><body><p>Just some text</p></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 1 {
		t.Fatalf("Expected 1 section, got %d", len(sections))
	}

	s := sections[0]
	if s.SectionID != "s1" {
		t.Errorf("Expected section_id 's1', got %q", s.SectionID)
	}
	if s.HeadingLevel != 0 {
		t.Errorf("Expected heading_level 0, got %d", s.HeadingLevel)
	}
	if s.HeadingText != "" {
		t.Errorf("Expected empty heading_text, got %q", s.HeadingText)
	}
	if !strings.Contains(s.BodyMarkdown, "Just some text") {
		t.Errorf("Expected body_markdown to contain 'Just some text', got %q", s.BodyMarkdown)
	}
}

func TestExtractSections_SingleH1(t *testing.T) {
	doc := parseHTML(t, `<html><body><h1>Title</h1><p>Content here</p></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 1 {
		t.Fatalf("Expected 1 section, got %d", len(sections))
	}

	s := sections[0]
	if s.SectionID != "s1" {
		t.Errorf("Expected section_id 's1', got %q", s.SectionID)
	}
	if s.HeadingLevel != 1 {
		t.Errorf("Expected heading_level 1, got %d", s.HeadingLevel)
	}
	if s.HeadingText != "Title" {
		t.Errorf("Expected heading_text 'Title', got %q", s.HeadingText)
	}
	if !strings.Contains(s.BodyMarkdown, "Content here") {
		t.Errorf("Expected body_markdown to contain 'Content here', got %q", s.BodyMarkdown)
	}
}

func TestExtractSections_MultipleHeadings(t *testing.T) {
	doc := parseHTML(t, `<html><body>
		<h1>First</h1><p>Paragraph one</p>
		<h2>Second</h2><p>Paragraph two</p>
		<h3>Third</h3><p>Paragraph three</p>
	</body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 3 {
		t.Fatalf("Expected 3 sections, got %d", len(sections))
	}

	expectedLevels := []int{1, 2, 3}
	expectedIDs := []string{"s1", "s2", "s3"}
	expectedHeadings := []string{"First", "Second", "Third"}

	for i, s := range sections {
		if s.SectionID != expectedIDs[i] {
			t.Errorf("Section %d: expected section_id %q, got %q", i, expectedIDs[i], s.SectionID)
		}
		if s.HeadingLevel != expectedLevels[i] {
			t.Errorf("Section %d: expected heading_level %d, got %d", i, expectedLevels[i], s.HeadingLevel)
		}
		if s.HeadingText != expectedHeadings[i] {
			t.Errorf("Section %d: expected heading_text %q, got %q", i, expectedHeadings[i], s.HeadingText)
		}
	}
}

func TestExtractSections_IntroBeforeFirstHeading(t *testing.T) {
	doc := parseHTML(t, `<html><body><p>Intro text</p><h1>First Heading</h1><p>Body</p></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 2 {
		t.Fatalf("Expected 2 sections, got %d", len(sections))
	}

	// First section: intro with no heading
	s0 := sections[0]
	if s0.HeadingLevel != 0 {
		t.Errorf("Intro section: expected heading_level 0, got %d", s0.HeadingLevel)
	}
	if s0.HeadingText != "" {
		t.Errorf("Intro section: expected empty heading_text, got %q", s0.HeadingText)
	}
	if !strings.Contains(s0.BodyMarkdown, "Intro text") {
		t.Errorf("Intro section: expected body_markdown to contain 'Intro text', got %q", s0.BodyMarkdown)
	}

	// Second section: h1 heading
	s1 := sections[1]
	if s1.HeadingLevel != 1 {
		t.Errorf("H1 section: expected heading_level 1, got %d", s1.HeadingLevel)
	}
	if s1.HeadingText != "First Heading" {
		t.Errorf("H1 section: expected heading_text 'First Heading', got %q", s1.HeadingText)
	}
}

func TestExtractSections_HeadingWithEmptyBody(t *testing.T) {
	doc := parseHTML(t, `<html><body><h1>Heading One</h1><h2>Heading Two</h2><p>Content</p></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 2 {
		t.Fatalf("Expected 2 sections, got %d", len(sections))
	}

	// First section: heading with empty body (included because heading is non-empty)
	s0 := sections[0]
	if s0.HeadingText != "Heading One" {
		t.Errorf("First section: expected heading_text 'Heading One', got %q", s0.HeadingText)
	}
	if s0.BodyMarkdown != "" {
		t.Errorf("First section: expected empty body_markdown, got %q", s0.BodyMarkdown)
	}

	// Second section: heading with content
	s1 := sections[1]
	if s1.HeadingText != "Heading Two" {
		t.Errorf("Second section: expected heading_text 'Heading Two', got %q", s1.HeadingText)
	}
	if !strings.Contains(s1.BodyMarkdown, "Content") {
		t.Errorf("Second section: expected body_markdown to contain 'Content', got %q", s1.BodyMarkdown)
	}
}

func TestExtractSections_EmptyHeadingEmptyBody_Omitted(t *testing.T) {
	doc := parseHTML(t, `<html><body><h1></h1><h2>Real Heading</h2><p>Content</p></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 1 {
		t.Fatalf("Expected 1 section (empty heading+body omitted), got %d", len(sections))
	}

	s := sections[0]
	if s.SectionID != "s1" {
		t.Errorf("Expected section_id 's1', got %q", s.SectionID)
	}
	if s.HeadingText != "Real Heading" {
		t.Errorf("Expected heading_text 'Real Heading', got %q", s.HeadingText)
	}
}

func TestExtractSections_ScriptStyleRemoved(t *testing.T) {
	doc := parseHTML(t, `<html><body>
		<script>var x=1;</script>
		<h1>Title</h1>
		<p>Visible content</p>
		<style>.foo{color:red;}</style>
	</body></html>`)
	sections := ExtractSections(doc)

	for _, s := range sections {
		if strings.Contains(s.BodyMarkdown, "var x=1") {
			t.Errorf("Script content should not appear in body_markdown, got %q", s.BodyMarkdown)
		}
		if strings.Contains(s.BodyMarkdown, ".foo") {
			t.Errorf("Style content should not appear in body_markdown, got %q", s.BodyMarkdown)
		}
	}
}

func TestExtractSections_NavHeaderFooterRemoved(t *testing.T) {
	doc := parseHTML(t, `<html><body>
		<nav><a href="/">Home</a></nav>
		<h1>Main</h1>
		<p>Content</p>
		<footer>Copyright</footer>
	</body></html>`)
	sections := ExtractSections(doc)

	for _, s := range sections {
		if strings.Contains(s.BodyMarkdown, "Home") {
			t.Errorf("Nav content should not appear in body_markdown, got %q", s.BodyMarkdown)
		}
		if strings.Contains(s.HeadingText, "Home") {
			t.Errorf("Nav content should not appear in heading_text, got %q", s.HeadingText)
		}
		if strings.Contains(s.BodyMarkdown, "Copyright") {
			t.Errorf("Footer content should not appear in body_markdown, got %q", s.BodyMarkdown)
		}
	}

	// Verify the real content is present
	found := false
	for _, s := range sections {
		if s.HeadingText == "Main" && strings.Contains(s.BodyMarkdown, "Content") {
			found = true
		}
	}
	if !found {
		t.Errorf("Expected to find section with heading 'Main' and body containing 'Content'")
	}
}

func TestExtractSections_Lists(t *testing.T) {
	doc := parseHTML(t, `<html><body><h1>Title</h1><ul><li>Item 1</li><li>Item 2</li></ul></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 1 {
		t.Fatalf("Expected 1 section, got %d", len(sections))
	}

	s := sections[0]
	if !strings.Contains(s.BodyMarkdown, "- Item 1") {
		t.Errorf("Expected body_markdown to contain '- Item 1', got %q", s.BodyMarkdown)
	}
	if !strings.Contains(s.BodyMarkdown, "- Item 2") {
		t.Errorf("Expected body_markdown to contain '- Item 2', got %q", s.BodyMarkdown)
	}
}

func TestExtractSections_OrderedList(t *testing.T) {
	doc := parseHTML(t, `<html><body><h1>Steps</h1><ol><li>First</li><li>Second</li></ol></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 1 {
		t.Fatalf("Expected 1 section, got %d", len(sections))
	}

	s := sections[0]
	if !strings.Contains(s.BodyMarkdown, "1.") || !strings.Contains(s.BodyMarkdown, "First") {
		t.Errorf("Expected body_markdown to contain numbered item '1. First', got %q", s.BodyMarkdown)
	}
	if !strings.Contains(s.BodyMarkdown, "2.") || !strings.Contains(s.BodyMarkdown, "Second") {
		t.Errorf("Expected body_markdown to contain numbered item '2. Second', got %q", s.BodyMarkdown)
	}
}

func TestExtractSections_Blockquote(t *testing.T) {
	doc := parseHTML(t, `<html><body><h1>Quote</h1><blockquote>Famous words</blockquote></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 1 {
		t.Fatalf("Expected 1 section, got %d", len(sections))
	}

	s := sections[0]
	if !strings.Contains(s.BodyMarkdown, "> Famous words") {
		t.Errorf("Expected body_markdown to contain '> Famous words', got %q", s.BodyMarkdown)
	}
}

func TestExtractSections_InlineFormatting(t *testing.T) {
	doc := parseHTML(t, `<html><body>
		<h1>Title</h1>
		<p><strong>Bold</strong> and <em>italic</em> and <a href="https://example.com">link</a></p>
	</body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 1 {
		t.Fatalf("Expected 1 section, got %d", len(sections))
	}

	s := sections[0]
	if !strings.Contains(s.BodyMarkdown, "**Bold**") {
		t.Errorf("Expected body_markdown to contain '**Bold**', got %q", s.BodyMarkdown)
	}
	if !strings.Contains(s.BodyMarkdown, "*italic*") {
		t.Errorf("Expected body_markdown to contain '*italic*', got %q", s.BodyMarkdown)
	}
	if !strings.Contains(s.BodyMarkdown, "[link](https://example.com)") {
		t.Errorf("Expected body_markdown to contain '[link](https://example.com)', got %q", s.BodyMarkdown)
	}
}

func TestExtractSections_DeeplyNestedHeadings(t *testing.T) {
	doc := parseHTML(t, `<html><body>
		<h4>Level Four</h4><p>Content 4</p>
		<h5>Level Five</h5><p>Content 5</p>
		<h6>Level Six</h6><p>Content 6</p>
	</body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 3 {
		t.Fatalf("Expected 3 sections, got %d", len(sections))
	}

	expectedLevels := []int{4, 5, 6}
	for i, s := range sections {
		if s.HeadingLevel != expectedLevels[i] {
			t.Errorf("Section %d: expected heading_level %d, got %d", i, expectedLevels[i], s.HeadingLevel)
		}
	}
}

func TestExtractSections_SequentialIDs(t *testing.T) {
	doc := parseHTML(t, `<html><body>
		<h1>One</h1><p>P1</p>
		<h2>Two</h2><p>P2</p>
		<h3>Three</h3><p>P3</p>
		<h4>Four</h4><p>P4</p>
		<h5>Five</h5><p>P5</p>
	</body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 5 {
		t.Fatalf("Expected 5 sections, got %d", len(sections))
	}

	expectedIDs := []string{"s1", "s2", "s3", "s4", "s5"}
	for i, s := range sections {
		if s.SectionID != expectedIDs[i] {
			t.Errorf("Section %d: expected section_id %q, got %q", i, expectedIDs[i], s.SectionID)
		}
	}
}

func TestExtractSections_HeadingWithInlineElements(t *testing.T) {
	doc := parseHTML(t, `<html><body>
		<h1><a href="#">Link</a> in <strong>heading</strong></h1>
		<p>Body</p>
	</body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 1 {
		t.Fatalf("Expected 1 section, got %d", len(sections))
	}

	s := sections[0]
	// heading_text should be plain text without markdown formatting
	if strings.Contains(s.HeadingText, "**") {
		t.Errorf("heading_text should not contain markdown bold '**', got %q", s.HeadingText)
	}
	if strings.Contains(s.HeadingText, "[") || strings.Contains(s.HeadingText, "]") {
		t.Errorf("heading_text should not contain markdown link brackets, got %q", s.HeadingText)
	}
	if !strings.Contains(s.HeadingText, "Link") {
		t.Errorf("heading_text should contain 'Link', got %q", s.HeadingText)
	}
	if !strings.Contains(s.HeadingText, "heading") {
		t.Errorf("heading_text should contain 'heading', got %q", s.HeadingText)
	}
}

func TestExtractSections_EmptyDocument(t *testing.T) {
	doc := parseHTML(t, `<html><body></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 0 {
		t.Errorf("Expected 0 sections for empty document, got %d", len(sections))
	}
}
