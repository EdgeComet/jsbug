package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/user/jsbug/internal/compare"
	"golang.org/x/net/html"
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

func TestParseAriaLevel(t *testing.T) {
	tests := []struct {
		name     string
		attrs    []html.Attribute
		expected int
	}{
		{"aria-level=2", []html.Attribute{{Key: "aria-level", Val: "2"}}, 2},
		{"aria-level=3", []html.Attribute{{Key: "aria-level", Val: "3"}}, 3},
		{"aria-level=1", []html.Attribute{{Key: "aria-level", Val: "1"}}, 1},
		{"aria-level=6", []html.Attribute{{Key: "aria-level", Val: "6"}}, 6},
		{"empty value", []html.Attribute{{Key: "aria-level", Val: ""}}, 2},
		{"no attribute", nil, 2},
		{"non-numeric", []html.Attribute{{Key: "aria-level", Val: "abc"}}, 2},
		{"zero clamped", []html.Attribute{{Key: "aria-level", Val: "0"}}, 1},
		{"negative clamped", []html.Attribute{{Key: "aria-level", Val: "-1"}}, 1},
		{"seven clamped", []html.Attribute{{Key: "aria-level", Val: "7"}}, 6},
		{"hundred clamped", []html.Attribute{{Key: "aria-level", Val: "100"}}, 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &html.Node{
				Type: html.ElementNode,
				Data: "div",
				Attr: tt.attrs,
			}
			got := parseAriaLevel(node)
			if got != tt.expected {
				t.Errorf("parseAriaLevel() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestIsBlockElement(t *testing.T) {
	tests := []struct {
		tag      string
		expected bool
	}{
		{"div", true},
		{"p", true},
		{"span", true},
		{"a", true},
		{"strong", true},
		{"b", true},
		{"label", true},
		{"figcaption", true},
		{"code", false},
		{"td", false},
		{"li", false},
		{"ul", false},
		{"ol", false},
		{"nav", false},
		{"h1", false},
		{"img", false},
		{"table", false},
		{"section", false},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			got := isBlockElement(tt.tag)
			if got != tt.expected {
				t.Errorf("isBlockElement(%q) = %v, want %v", tt.tag, got, tt.expected)
			}
		})
	}
}

func TestClassMatchesHeading(t *testing.T) {
	tests := []struct {
		class    string
		expected bool
	}{
		{"section-heading", true},
		{"sectionHeading", true},
		{"main_heading block", true},
		{"Section-Heading", true},
		{"HEADING", true},
		{"no-heading-here", true},
		{"section-title", false},
		{"header", false},
		{"", false},
		{"head", false},
	}

	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			got := classMatchesHeading(tt.class)
			if got != tt.expected {
				t.Errorf("classMatchesHeading(%q) = %v, want %v", tt.class, got, tt.expected)
			}
		})
	}
}

func TestHasChildHeadingOrARIA(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected bool
	}{
		{
			"contains h2 child",
			`<html><body><div id="target"><h2>Child</h2><p>text</p></div></body></html>`,
			true,
		},
		{
			"contains ARIA heading child",
			`<html><body><div id="target"><div role="heading">Child</div></div></body></html>`,
			true,
		},
		{
			"deeply nested h3",
			`<html><body><div id="target"><div><div><h3>Deep</h3></div></div></div></body></html>`,
			true,
		},
		{
			"no headings",
			`<html><body><div id="target"><p>No headings</p><span>text</span></div></body></html>`,
			false,
		},
		{
			"self is heading but no child headings",
			`<html><body><h2 id="target">Self</h2></body></html>`,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := parseHTML(t, tt.html)
			sel := doc.Find("#target")
			if sel.Length() == 0 {
				t.Fatal("target node not found")
			}
			node := sel.Get(0)
			got := hasChildHeadingOrARIA(node)
			if got != tt.expected {
				t.Errorf("hasChildHeadingOrARIA() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMaxHeadingLevel(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected int
	}{
		{
			"only h1",
			`<html><body><h1>Title</h1></body></html>`,
			1,
		},
		{
			"h1 and h3",
			`<html><body><h1>Title</h1><h3>Sub</h3></body></html>`,
			3,
		},
		{
			"h1 through h6",
			`<html><body><h1>A</h1><h2>B</h2><h3>C</h3><h4>D</h4><h5>E</h5><h6>F</h6></body></html>`,
			6,
		},
		{
			"ARIA heading level 3",
			`<html><body><div role="heading" aria-level="3">ARIA</div></body></html>`,
			3,
		},
		{
			"h1 and ARIA level 4",
			`<html><body><h1>Title</h1><div role="heading" aria-level="4">ARIA</div></body></html>`,
			4,
		},
		{
			"no headings",
			`<html><body><p>text</p></body></html>`,
			0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := parseHTML(t, tt.html)
			sel := doc.Find("body")
			if sel.Length() == 0 {
				t.Fatal("body node not found")
			}
			bodyNode := sel.Get(0)
			got := maxHeadingLevel(bodyNode)
			if got != tt.expected {
				t.Errorf("maxHeadingLevel() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestDataAttrMatchesHeading(t *testing.T) {
	tests := []struct {
		name     string
		attrs    []html.Attribute
		expected bool
	}{
		{"data-heading=true", []html.Attribute{{Key: "data-heading", Val: "true"}}, true},
		{"data-heading=yes", []html.Attribute{{Key: "data-heading", Val: "yes"}}, true},
		{"data-heading=1", []html.Attribute{{Key: "data-heading", Val: "1"}}, true},
		{"data-heading=anything", []html.Attribute{{Key: "data-heading", Val: "anything"}}, true},
		{"data-role=heading", []html.Attribute{{Key: "data-role", Val: "heading"}}, true},
		{"data-role=Heading", []html.Attribute{{Key: "data-role", Val: "Heading"}}, true},
		{"data-heading=empty", []html.Attribute{{Key: "data-heading", Val: ""}}, false},
		{"data-heading=false", []html.Attribute{{Key: "data-heading", Val: "false"}}, false},
		{"data-heading=0", []html.Attribute{{Key: "data-heading", Val: "0"}}, false},
		{"data-heading=no", []html.Attribute{{Key: "data-heading", Val: "no"}}, false},
		{"data-heading=FALSE", []html.Attribute{{Key: "data-heading", Val: "FALSE"}}, false},
		{"data-heading=No", []html.Attribute{{Key: "data-heading", Val: "No"}}, false},
		{"data-role=button", []html.Attribute{{Key: "data-role", Val: "button"}}, false},
		{"no data attributes", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &html.Node{
				Type: html.ElementNode,
				Data: "div",
				Attr: tt.attrs,
			}
			got := dataAttrMatchesHeading(node)
			if got != tt.expected {
				t.Errorf("dataAttrMatchesHeading() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsARIAHeading(t *testing.T) {
	longText := strings.Repeat("x", 250)

	tests := []struct {
		name          string
		html          string
		selector      string
		expectedMatch bool
		expectedLevel int
	}{
		{
			"div role=heading aria-level=2",
			`<html><body><div role="heading" aria-level="2">Title</div></body></html>`,
			"div[role]",
			true, 2,
		},
		{
			"span role=heading aria-level=3",
			`<html><body><span role="heading" aria-level="3">Title</span></body></html>`,
			"span[role]",
			true, 3,
		},
		{
			"default level when aria-level missing",
			`<html><body><div role="heading">Title</div></body></html>`,
			"div[role]",
			true, 2,
		},
		{
			"h2 already handled",
			`<html><body><h2 role="heading">Title</h2></body></html>`,
			"h2",
			false, 0,
		},
		{
			"wrong role",
			`<html><body><div role="button">Title</div></body></html>`,
			"div[role]",
			false, 0,
		},
		{
			"text too long",
			`<html><body><div role="heading" aria-level="2">` + longText + `</div></body></html>`,
			"div[role]",
			false, 0,
		},
		{
			"empty text",
			`<html><body><div role="heading" aria-level="2"></div></body></html>`,
			"div[role]",
			false, 0,
		},
		{
			"container with multiple children",
			`<html><body><section role="heading"><div>A</div><div>B</div></section></body></html>`,
			"section[role]",
			false, 0,
		},
		{
			"container with single child OK",
			`<html><body><div role="heading"><span>Single child</span></div></body></html>`,
			"div[role]",
			true, 2,
		},
		{
			"contains child heading",
			`<html><body><div role="heading"><h2>Nested</h2></div></body></html>`,
			"div[role]",
			false, 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := parseHTML(t, tt.html)
			sel := doc.Find(tt.selector)
			if sel.Length() == 0 {
				t.Fatal("target node not found")
			}
			node := sel.Get(0)
			gotMatch, gotLevel := isARIAHeading(node)
			if gotMatch != tt.expectedMatch {
				t.Errorf("isARIAHeading() match = %v, want %v", gotMatch, tt.expectedMatch)
			}
			if gotLevel != tt.expectedLevel {
				t.Errorf("isARIAHeading() level = %d, want %d", gotLevel, tt.expectedLevel)
			}
		})
	}
}

func TestIsVisualHeading(t *testing.T) {
	longText := strings.Repeat("x", 250)

	tests := []struct {
		name           string
		html           string
		selector       string
		fallbackLevel  int
		expectedMatch  bool
		expectedLevel  int
		expectedMethod string
	}{
		{
			"class signal",
			`<html><body><div class="section-heading">Title</div></body></html>`,
			"div.section-heading",
			2,
			true, 2, "css_class",
		},
		{
			"data-heading signal",
			`<html><body><div data-heading="true">Title</div></body></html>`,
			"div[data-heading]",
			3,
			true, 3, "data_attr",
		},
		{
			"class wins over data-attr",
			`<html><body><div class="section-heading" data-heading="true">Title</div></body></html>`,
			"div.section-heading",
			2,
			true, 2, "css_class",
		},
		{
			"h2 already handled",
			`<html><body><h2 class="section-heading">Title</h2></body></html>`,
			"h2",
			2,
			false, 0, "",
		},
		{
			"not eligible tag",
			`<html><body><code class="section-heading">Title</code></body></html>`,
			"code",
			2,
			false, 0, "",
		},
		{
			"text too long",
			`<html><body><div class="section-heading">` + longText + `</div></body></html>`,
			"div.section-heading",
			2,
			false, 0, "",
		},
		{
			"empty text",
			`<html><body><div class="section-heading"></div></body></html>`,
			"div.section-heading",
			2,
			false, 0, "",
		},
		{
			"contains child heading",
			`<html><body><div class="heading-group"><h2>Title</h2></div></body></html>`,
			"div.heading-group",
			2,
			false, 0, "",
		},
		{
			"no heading signal",
			`<html><body><div class="other-class">Title</div></body></html>`,
			"div.other-class",
			2,
			false, 0, "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := parseHTML(t, tt.html)
			sel := doc.Find(tt.selector)
			if sel.Length() == 0 {
				t.Fatal("target node not found")
			}
			node := sel.Get(0)
			gotMatch, gotLevel, gotMethod := isVisualHeading(node, tt.fallbackLevel)
			if gotMatch != tt.expectedMatch {
				t.Errorf("isVisualHeading() match = %v, want %v", gotMatch, tt.expectedMatch)
			}
			if gotLevel != tt.expectedLevel {
				t.Errorf("isVisualHeading() level = %d, want %d", gotLevel, tt.expectedLevel)
			}
			if gotMethod != tt.expectedMethod {
				t.Errorf("isVisualHeading() method = %q, want %q", gotMethod, tt.expectedMethod)
			}
		})
	}
}

func TestExtractSections_ClassHeading(t *testing.T) {
	doc := parseHTML(t, `<html><body><h1>Title</h1><div class="section-heading">Features</div><p>Content</p></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 2 {
		t.Fatalf("Expected 2 sections, got %d", len(sections))
	}

	s1 := sections[0]
	if s1.HeadingLevel != 1 {
		t.Errorf("s1: expected heading_level 1, got %d", s1.HeadingLevel)
	}
	if s1.HeadingText != "Title" {
		t.Errorf("s1: expected heading_text 'Title', got %q", s1.HeadingText)
	}

	s2 := sections[1]
	if s2.HeadingText != "Features" {
		t.Errorf("s2: expected heading_text 'Features', got %q", s2.HeadingText)
	}
	if s2.HeadingLevel != 2 {
		t.Errorf("s2: expected heading_level 2 (fallback: deepest=1, so 1+1=2), got %d", s2.HeadingLevel)
	}
	if s2.DetectionMethod != "css_class" {
		t.Errorf("s2: expected detection_method 'css_class', got %q", s2.DetectionMethod)
	}
	if !strings.Contains(s2.BodyMarkdown, "Content") {
		t.Errorf("s2: expected body_markdown to contain 'Content', got %q", s2.BodyMarkdown)
	}
}

func TestExtractSections_ClassCaseInsensitive(t *testing.T) {
	doc := parseHTML(t, `<html><body><h1>Title</h1><div class="Section-Heading">Features</div><p>Content</p></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 2 {
		t.Fatalf("Expected 2 sections, got %d", len(sections))
	}

	s2 := sections[1]
	if s2.HeadingText != "Features" {
		t.Errorf("s2: expected heading_text 'Features', got %q", s2.HeadingText)
	}
	if s2.DetectionMethod != "css_class" {
		t.Errorf("s2: expected detection_method 'css_class' (case-insensitive match), got %q", s2.DetectionMethod)
	}
}

func TestExtractSections_ClassMultipleClasses(t *testing.T) {
	doc := parseHTML(t, `<html><body><h1>Title</h1><div class="section-heading other-class">Features</div><p>Content</p></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 2 {
		t.Fatalf("Expected 2 sections, got %d", len(sections))
	}

	s2 := sections[1]
	if s2.HeadingText != "Features" {
		t.Errorf("s2: expected heading_text 'Features', got %q", s2.HeadingText)
	}
	if s2.DetectionMethod != "css_class" {
		t.Errorf("s2: expected detection_method 'css_class', got %q", s2.DetectionMethod)
	}
}

func TestExtractSections_ClassSpanHeading(t *testing.T) {
	doc := parseHTML(t, `<html><body><h1>Title</h1><span class="section-heading">Features</span><p>Content</p></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 2 {
		t.Fatalf("Expected 2 sections, got %d", len(sections))
	}

	s2 := sections[1]
	if s2.HeadingText != "Features" {
		t.Errorf("s2: expected heading_text 'Features', got %q", s2.HeadingText)
	}
	if s2.DetectionMethod != "css_class" {
		t.Errorf("s2: expected detection_method 'css_class' (span is eligible), got %q", s2.DetectionMethod)
	}
}

func TestExtractSections_ClassParagraphHeading(t *testing.T) {
	doc := parseHTML(t, `<html><body><h1>Title</h1><p class="page-heading">Features</p><p>Content</p></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 2 {
		t.Fatalf("Expected 2 sections, got %d", len(sections))
	}

	s2 := sections[1]
	if s2.HeadingText != "Features" {
		t.Errorf("s2: expected heading_text 'Features', got %q", s2.HeadingText)
	}
	if s2.DetectionMethod != "css_class" {
		t.Errorf("s2: expected detection_method 'css_class' (p is eligible), got %q", s2.DetectionMethod)
	}
}

func TestExtractSections_ClassAnchorHeading(t *testing.T) {
	doc := parseHTML(t, `<html><body><h1>Title</h1><a class="section-heading" href="#features">Features</a><p>Content</p></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 2 {
		t.Fatalf("Expected 2 sections, got %d", len(sections))
	}

	s2 := sections[1]
	if s2.HeadingText != "Features" {
		t.Errorf("s2: expected heading_text 'Features', got %q", s2.HeadingText)
	}
	if s2.DetectionMethod != "css_class" {
		t.Errorf("s2: expected detection_method 'css_class' (a is eligible), got %q", s2.DetectionMethod)
	}
}

func TestExtractSections_ClassStrongHeading(t *testing.T) {
	doc := parseHTML(t, `<html><body><h1>Title</h1><strong class="section-heading">Features</strong><p>Content</p></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 2 {
		t.Fatalf("Expected 2 sections, got %d", len(sections))
	}

	s2 := sections[1]
	if s2.HeadingText != "Features" {
		t.Errorf("s2: expected heading_text 'Features', got %q", s2.HeadingText)
	}
	if s2.DetectionMethod != "css_class" {
		t.Errorf("s2: expected detection_method 'css_class', got %q", s2.DetectionMethod)
	}
}

func TestExtractSections_ClassBoldHeading(t *testing.T) {
	doc := parseHTML(t, `<html><body><h1>Title</h1><b class="section-heading">Features</b><p>Content</p></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 2 {
		t.Fatalf("Expected 2 sections, got %d", len(sections))
	}

	s2 := sections[1]
	if s2.HeadingText != "Features" {
		t.Errorf("s2: expected heading_text 'Features', got %q", s2.HeadingText)
	}
	if s2.DetectionMethod != "css_class" {
		t.Errorf("s2: expected detection_method 'css_class', got %q", s2.DetectionMethod)
	}
}

func TestExtractSections_ARIAHeading(t *testing.T) {
	doc := parseHTML(t, `<html><body><h1>Title</h1><div role="heading" aria-level="2">Features</div><p>Content</p></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 2 {
		t.Fatalf("Expected 2 sections, got %d", len(sections))
	}

	s1 := sections[0]
	if s1.HeadingLevel != 1 {
		t.Errorf("s1: expected heading_level 1, got %d", s1.HeadingLevel)
	}
	if s1.HeadingText != "Title" {
		t.Errorf("s1: expected heading_text 'Title', got %q", s1.HeadingText)
	}
	if s1.DetectionMethod != "semantic" {
		t.Errorf("s1: expected detection_method 'semantic', got %q", s1.DetectionMethod)
	}

	s2 := sections[1]
	if s2.HeadingLevel != 2 {
		t.Errorf("s2: expected heading_level 2, got %d", s2.HeadingLevel)
	}
	if s2.HeadingText != "Features" {
		t.Errorf("s2: expected heading_text 'Features', got %q", s2.HeadingText)
	}
	if s2.DetectionMethod != "aria" {
		t.Errorf("s2: expected detection_method 'aria', got %q", s2.DetectionMethod)
	}
	if !strings.Contains(s2.BodyMarkdown, "Content") {
		t.Errorf("s2: expected body_markdown to contain 'Content', got %q", s2.BodyMarkdown)
	}
}

func TestExtractSections_ARIAHeadingDefaultLevel(t *testing.T) {
	doc := parseHTML(t, `<html><body><h1>Title</h1><div role="heading">Features</div><p>Content</p></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 2 {
		t.Fatalf("Expected 2 sections, got %d", len(sections))
	}

	s2 := sections[1]
	if s2.HeadingLevel != 2 {
		t.Errorf("s2: expected heading_level 2 (default when aria-level missing), got %d", s2.HeadingLevel)
	}
	if s2.DetectionMethod != "aria" {
		t.Errorf("s2: expected detection_method 'aria', got %q", s2.DetectionMethod)
	}
}

func TestExtractSections_ARIAHeadingAnyTag(t *testing.T) {
	doc := parseHTML(t, `<html><body><h1>Title</h1><span role="heading" aria-level="3">Features</span><p>Content</p></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 2 {
		t.Fatalf("Expected 2 sections, got %d", len(sections))
	}

	s2 := sections[1]
	if s2.HeadingLevel != 3 {
		t.Errorf("s2: expected heading_level 3, got %d", s2.HeadingLevel)
	}
	if s2.DetectionMethod != "aria" {
		t.Errorf("s2: expected detection_method 'aria', got %q", s2.DetectionMethod)
	}
}

func TestExtractSections_ARIAHeadingPriority(t *testing.T) {
	doc := parseHTML(t, `<html><body><h1>Title</h1><div role="heading" aria-level="4" class="section-heading">Features</div><p>Content</p></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 2 {
		t.Fatalf("Expected 2 sections, got %d", len(sections))
	}

	s2 := sections[1]
	if s2.DetectionMethod != "aria" {
		t.Errorf("s2: expected detection_method 'aria' (ARIA takes priority over CSS class), got %q", s2.DetectionMethod)
	}
	if s2.HeadingLevel != 4 {
		t.Errorf("s2: expected heading_level 4 (ARIA level used, NOT fallbackLevel), got %d", s2.HeadingLevel)
	}
}

func TestExtractSections_ARIAHeadingLongText_Rejected(t *testing.T) {
	longText := strings.Repeat("a", 300)
	doc := parseHTML(t, `<html><body><h1>Title</h1><div role="heading" aria-level="2">`+longText+`</div><p>Content</p></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 1 {
		t.Fatalf("Expected 1 section (ARIA heading rejected for text > 200 runes), got %d", len(sections))
	}

	s := sections[0]
	if s.HeadingText != "Title" {
		t.Errorf("Expected heading_text 'Title', got %q", s.HeadingText)
	}
	if !strings.Contains(s.BodyMarkdown, longText) {
		t.Errorf("Expected body_markdown to contain the 300-char text, got %q", s.BodyMarkdown)
	}
	if !strings.Contains(s.BodyMarkdown, "Content") {
		t.Errorf("Expected body_markdown to contain 'Content', got %q", s.BodyMarkdown)
	}
}

func TestExtractSections_ARIAHeadingInvalidLevel(t *testing.T) {
	tests := []struct {
		name          string
		ariaLevel     string
		expectedLevel int
	}{
		{"zero", "0", 1},
		{"seven", "7", 6},
		{"negative", "-1", 1},
		{"non_numeric", "abc", 2},
		{"empty", "", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			htmlStr := `<html><body><h1>Title</h1><div role="heading" aria-level="` + tt.ariaLevel + `">Features</div><p>Content</p></body></html>`
			doc := parseHTML(t, htmlStr)
			sections := ExtractSections(doc)

			if len(sections) != 2 {
				t.Fatalf("Expected 2 sections, got %d", len(sections))
			}

			s2 := sections[1]
			if s2.HeadingLevel != tt.expectedLevel {
				t.Errorf("Expected heading_level %d for aria-level=%q, got %d", tt.expectedLevel, tt.ariaLevel, s2.HeadingLevel)
			}
			if s2.DetectionMethod != "aria" {
				t.Errorf("Expected detection_method 'aria', got %q", s2.DetectionMethod)
			}
		})
	}
}

func TestExtractSections_ARIAHeadingInNavRemoved(t *testing.T) {
	doc := parseHTML(t, `<html><body><nav><div role="heading" aria-level="2">Navigation</div></nav><h1>Title</h1><p>Content</p></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 1 {
		t.Fatalf("Expected 1 section (ARIA heading inside nav removed during cleanup), got %d", len(sections))
	}

	s := sections[0]
	if s.HeadingText != "Title" {
		t.Errorf("Expected heading_text 'Title', got %q", s.HeadingText)
	}
	if !strings.Contains(s.BodyMarkdown, "Content") {
		t.Errorf("Expected body_markdown to contain 'Content', got %q", s.BodyMarkdown)
	}
}

// --- Data attribute integration tests ---

func TestExtractSections_DataHeadingAttr(t *testing.T) {
	doc := parseHTML(t, `<html><body><h1>Title</h1><div data-heading="true">Features</div><p>Content</p></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 2 {
		t.Fatalf("Expected 2 sections, got %d", len(sections))
	}

	s2 := sections[1]
	if s2.HeadingText != "Features" {
		t.Errorf("s2: expected heading_text 'Features', got %q", s2.HeadingText)
	}
	if s2.DetectionMethod != "data_attr" {
		t.Errorf("s2: expected detection_method 'data_attr', got %q", s2.DetectionMethod)
	}
}

func TestExtractSections_DataRoleHeading(t *testing.T) {
	doc := parseHTML(t, `<html><body><h1>Title</h1><div data-role="heading">Features</div><p>Content</p></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 2 {
		t.Fatalf("Expected 2 sections, got %d", len(sections))
	}

	s2 := sections[1]
	if s2.HeadingText != "Features" {
		t.Errorf("s2: expected heading_text 'Features', got %q", s2.HeadingText)
	}
	if s2.DetectionMethod != "data_attr" {
		t.Errorf("s2: expected detection_method 'data_attr', got %q", s2.DetectionMethod)
	}
}

func TestExtractSections_DataRoleCaseInsensitive(t *testing.T) {
	doc := parseHTML(t, `<html><body><h1>Title</h1><div data-role="Heading">Features</div><p>Content</p></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 2 {
		t.Fatalf("Expected 2 sections, got %d", len(sections))
	}

	s2 := sections[1]
	if s2.HeadingText != "Features" {
		t.Errorf("s2: expected heading_text 'Features', got %q", s2.HeadingText)
	}
	if s2.DetectionMethod != "data_attr" {
		t.Errorf("s2: expected detection_method 'data_attr', got %q", s2.DetectionMethod)
	}
}

func TestExtractSections_DataHeadingEmpty_Rejected(t *testing.T) {
	doc := parseHTML(t, `<html><body><h1>Title</h1><div data-heading="">Features</div><p>Content</p></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 1 {
		t.Fatalf("Expected 1 section (empty data-heading is not truthy), got %d", len(sections))
	}

	s := sections[0]
	if s.HeadingText != "Title" {
		t.Errorf("Expected heading_text 'Title', got %q", s.HeadingText)
	}
}

func TestExtractSections_DataHeadingFalse_Rejected(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"false", "false"},
		{"zero", "0"},
		{"no", "no"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			htmlStr := `<html><body><h1>Title</h1><div data-heading="` + tt.value + `">Features</div><p>Content</p></body></html>`
			doc := parseHTML(t, htmlStr)
			sections := ExtractSections(doc)

			if len(sections) != 1 {
				t.Fatalf("Expected 1 section (data-heading=%q should not be truthy), got %d", tt.value, len(sections))
			}

			s := sections[0]
			if s.HeadingText != "Title" {
				t.Errorf("Expected heading_text 'Title', got %q", s.HeadingText)
			}
		})
	}
}

// --- False-positive rejection integration tests ---

func TestExtractSections_ClassHeadingWrapper_Rejected(t *testing.T) {
	doc := parseHTML(t, `<html><body><div class="heading-group"><h2>Real Heading</h2><p>subtitle</p></div><p>Body</p></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 1 {
		t.Fatalf("Expected 1 section (wrapper div rejected, h2 creates section), got %d", len(sections))
	}

	s := sections[0]
	if s.HeadingText != "Real Heading" {
		t.Errorf("Expected heading_text 'Real Heading', got %q", s.HeadingText)
	}
	if s.DetectionMethod != "semantic" {
		t.Errorf("Expected detection_method 'semantic', got %q", s.DetectionMethod)
	}
}

func TestExtractSections_ClassHeadingLongText_Rejected(t *testing.T) {
	longText := strings.Repeat("x", 300)
	doc := parseHTML(t, `<html><body><h1>Title</h1><div class="section-heading">`+longText+`</div><p>Content</p></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 1 {
		t.Fatalf("Expected 1 section (long-text div rejected), got %d", len(sections))
	}

	s := sections[0]
	if s.HeadingText != "Title" {
		t.Errorf("Expected heading_text 'Title', got %q", s.HeadingText)
	}
}

func TestExtractSections_ClassHeadingEmptyText_Rejected(t *testing.T) {
	doc := parseHTML(t, `<html><body><h1>Title</h1><div class="heading"></div><p>Content</p></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 1 {
		t.Fatalf("Expected 1 section (empty-text div rejected), got %d", len(sections))
	}

	s := sections[0]
	if s.HeadingText != "Title" {
		t.Errorf("Expected heading_text 'Title', got %q", s.HeadingText)
	}
}

func TestExtractSections_ClassHeadingNonEligibleTag_Rejected(t *testing.T) {
	doc := parseHTML(t, `<html><body><h1>Title</h1><div><code class="section-heading">Code</code></div><p>Content</p></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 1 {
		t.Fatalf("Expected 1 section (code tag not eligible for visual heading), got %d", len(sections))
	}

	s := sections[0]
	if s.HeadingText != "Title" {
		t.Errorf("Expected heading_text 'Title', got %q", s.HeadingText)
	}
}

func TestExtractSections_ClassHeadingNestedContainer(t *testing.T) {
	doc := parseHTML(t, `<html><body><h1>Title</h1><article><div class="section-heading">Sub</div><p>Text</p></article></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 2 {
		t.Fatalf("Expected 2 sections (visual heading detected inside nested container), got %d", len(sections))
	}

	s2 := sections[1]
	if s2.HeadingText != "Sub" {
		t.Errorf("s2: expected heading_text 'Sub', got %q", s2.HeadingText)
	}
	if s2.DetectionMethod != "css_class" {
		t.Errorf("s2: expected detection_method 'css_class', got %q", s2.DetectionMethod)
	}
}

func TestExtractSections_ClassHeadingDeeplyNested(t *testing.T) {
	doc := parseHTML(t, `<html><body><h1>Title</h1><div><div><div><div class="section-heading">Deep</div></div></div></div><p>Content</p></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 2 {
		t.Fatalf("Expected 2 sections (visual heading detected through multiple nested containers), got %d", len(sections))
	}

	s2 := sections[1]
	if s2.HeadingText != "Deep" {
		t.Errorf("s2: expected heading_text 'Deep', got %q", s2.HeadingText)
	}
	if s2.DetectionMethod != "css_class" {
		t.Errorf("s2: expected detection_method 'css_class', got %q", s2.DetectionMethod)
	}
}

// --- Fallback level inference tests ---

func TestExtractSections_FallbackLevel_OnlyH1(t *testing.T) {
	doc := parseHTML(t, `<html><body><h1>Title</h1><div class="section-heading">Sub</div><p>Content</p></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 2 {
		t.Fatalf("Expected 2 sections, got %d", len(sections))
	}

	s1 := sections[0]
	if s1.HeadingLevel != 1 {
		t.Errorf("s1: expected heading_level 1, got %d", s1.HeadingLevel)
	}
	if s1.HeadingText != "Title" {
		t.Errorf("s1: expected heading_text 'Title', got %q", s1.HeadingText)
	}

	s2 := sections[1]
	if s2.HeadingLevel != 2 {
		t.Errorf("s2: expected heading_level 2 (deepest=1, fallback=2), got %d", s2.HeadingLevel)
	}
	if s2.HeadingText != "Sub" {
		t.Errorf("s2: expected heading_text 'Sub', got %q", s2.HeadingText)
	}
}

func TestExtractSections_FallbackLevel_H1H2(t *testing.T) {
	doc := parseHTML(t, `<html><body><h1>Title</h1><h2>Sub</h2><p>P1</p><div class="section-heading">Visual</div><p>P2</p></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 3 {
		t.Fatalf("Expected 3 sections, got %d", len(sections))
	}

	s1 := sections[0]
	if s1.HeadingLevel != 1 {
		t.Errorf("s1: expected heading_level 1, got %d", s1.HeadingLevel)
	}

	s2 := sections[1]
	if s2.HeadingLevel != 2 {
		t.Errorf("s2: expected heading_level 2 (h2), got %d", s2.HeadingLevel)
	}

	s3 := sections[2]
	if s3.HeadingLevel != 3 {
		t.Errorf("s3: expected heading_level 3 (deepest=2, fallback=3), got %d", s3.HeadingLevel)
	}
	if s3.DetectionMethod != "css_class" {
		t.Errorf("s3: expected detection_method 'css_class', got %q", s3.DetectionMethod)
	}
}

func TestExtractSections_FallbackLevel_NoRealHeadings(t *testing.T) {
	doc := parseHTML(t, `<html><body><div class="section-heading">A</div><p>P1</p><div class="section-heading">B</div><p>P2</p></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 2 {
		t.Fatalf("Expected 2 sections, got %d", len(sections))
	}

	s1 := sections[0]
	if s1.HeadingLevel != 2 {
		t.Errorf("s1: expected heading_level 2 (deepest=0, fallback=2), got %d", s1.HeadingLevel)
	}
	if s1.HeadingText != "A" {
		t.Errorf("s1: expected heading_text 'A', got %q", s1.HeadingText)
	}

	s2 := sections[1]
	if s2.HeadingLevel != 2 {
		t.Errorf("s2: expected heading_level 2 (deepest=0, fallback=2), got %d", s2.HeadingLevel)
	}
	if s2.HeadingText != "B" {
		t.Errorf("s2: expected heading_text 'B', got %q", s2.HeadingText)
	}
}

func TestExtractSections_FallbackLevel_CappedAt6(t *testing.T) {
	doc := parseHTML(t, `<html><body><h1>H1</h1><h2>H2</h2><h3>H3</h3><h4>H4</h4><h5>H5</h5><h6>H6</h6><div class="section-heading">Visual</div><p>Content</p></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 7 {
		t.Fatalf("Expected 7 sections, got %d", len(sections))
	}

	expectedLevels := []int{1, 2, 3, 4, 5, 6, 6}
	for i, s := range sections {
		if s.HeadingLevel != expectedLevels[i] {
			t.Errorf("Section %d: expected heading_level %d, got %d", i, expectedLevels[i], s.HeadingLevel)
		}
	}

	// The visual heading (last section) must be capped at 6
	s7 := sections[6]
	if s7.HeadingText != "Visual" {
		t.Errorf("s7: expected heading_text 'Visual', got %q", s7.HeadingText)
	}
	if s7.HeadingLevel != 6 {
		t.Errorf("s7: expected heading_level 6 (deepest=6, fallback=min(7,6)=6), got %d", s7.HeadingLevel)
	}
}

// --- Detection method tracking tests ---

func TestExtractSections_DetectionMethod_Semantic(t *testing.T) {
	doc := parseHTML(t, `<html><body><h1>Title</h1><h2>Sub</h2><p>Content</p></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 2 {
		t.Fatalf("Expected 2 sections, got %d", len(sections))
	}

	s1 := sections[0]
	if s1.DetectionMethod != "semantic" {
		t.Errorf("s1: expected detection_method 'semantic', got %q", s1.DetectionMethod)
	}

	s2 := sections[1]
	if s2.DetectionMethod != "semantic" {
		t.Errorf("s2: expected detection_method 'semantic', got %q", s2.DetectionMethod)
	}
}

func TestExtractSections_DetectionMethod_Intro(t *testing.T) {
	doc := parseHTML(t, `<html><body><p>Intro</p><h1>Title</h1><p>Content</p></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 2 {
		t.Fatalf("Expected 2 sections, got %d", len(sections))
	}

	s1 := sections[0]
	if s1.DetectionMethod != "" {
		t.Errorf("s1 (intro): expected detection_method '' (empty), got %q", s1.DetectionMethod)
	}

	s2 := sections[1]
	if s2.DetectionMethod != "semantic" {
		t.Errorf("s2: expected detection_method 'semantic', got %q", s2.DetectionMethod)
	}
}

func TestExtractSections_DetectionMethod_Mixed(t *testing.T) {
	doc := parseHTML(t, `<html><body><h1>Title</h1><div role="heading" aria-level="2">ARIA</div><p>P1</p><div class="section-heading">CSS</div><p>P2</p><div data-heading="true">Data</div><p>P3</p></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 4 {
		t.Fatalf("Expected 4 sections, got %d", len(sections))
	}

	expectedMethods := []string{"semantic", "aria", "css_class", "data_attr"}
	expectedTexts := []string{"Title", "ARIA", "CSS", "Data"}

	for i, s := range sections {
		if s.DetectionMethod != expectedMethods[i] {
			t.Errorf("Section %d (%q): expected detection_method %q, got %q", i, expectedTexts[i], expectedMethods[i], s.DetectionMethod)
		}
		if s.HeadingText != expectedTexts[i] {
			t.Errorf("Section %d: expected heading_text %q, got %q", i, expectedTexts[i], s.HeadingText)
		}
	}
}

// --- Integration tests: multiple features interacting together ---

func TestExtractSections_MixedRealAndVisual(t *testing.T) {
	doc := parseHTML(t, `<html><body><h1>Title</h1><p>Intro</p><h2>Real Sub</h2><p>P1</p><div class="section-heading">Visual Sub</div><p>P2</p><h2>Another Real</h2><p>P3</p></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 4 {
		t.Fatalf("Expected 4 sections, got %d", len(sections))
	}

	// s1: level=1, text="Title", method="semantic"
	s1 := sections[0]
	if s1.SectionID != "s1" {
		t.Errorf("s1: expected section_id 's1', got %q", s1.SectionID)
	}
	if s1.HeadingLevel != 1 {
		t.Errorf("s1: expected heading_level 1, got %d", s1.HeadingLevel)
	}
	if s1.HeadingText != "Title" {
		t.Errorf("s1: expected heading_text 'Title', got %q", s1.HeadingText)
	}
	if s1.DetectionMethod != "semantic" {
		t.Errorf("s1: expected detection_method 'semantic', got %q", s1.DetectionMethod)
	}

	// s2: level=2, text="Real Sub", method="semantic"
	s2 := sections[1]
	if s2.SectionID != "s2" {
		t.Errorf("s2: expected section_id 's2', got %q", s2.SectionID)
	}
	if s2.HeadingLevel != 2 {
		t.Errorf("s2: expected heading_level 2, got %d", s2.HeadingLevel)
	}
	if s2.HeadingText != "Real Sub" {
		t.Errorf("s2: expected heading_text 'Real Sub', got %q", s2.HeadingText)
	}
	if s2.DetectionMethod != "semantic" {
		t.Errorf("s2: expected detection_method 'semantic', got %q", s2.DetectionMethod)
	}

	// s3: level=3, text="Visual Sub", method="css_class" (fallback: deepest=2, so 3)
	s3 := sections[2]
	if s3.SectionID != "s3" {
		t.Errorf("s3: expected section_id 's3', got %q", s3.SectionID)
	}
	if s3.HeadingLevel != 3 {
		t.Errorf("s3: expected heading_level 3 (deepest=2, fallback=3), got %d", s3.HeadingLevel)
	}
	if s3.HeadingText != "Visual Sub" {
		t.Errorf("s3: expected heading_text 'Visual Sub', got %q", s3.HeadingText)
	}
	if s3.DetectionMethod != "css_class" {
		t.Errorf("s3: expected detection_method 'css_class', got %q", s3.DetectionMethod)
	}

	// s4: level=2, text="Another Real", method="semantic"
	s4 := sections[3]
	if s4.SectionID != "s4" {
		t.Errorf("s4: expected section_id 's4', got %q", s4.SectionID)
	}
	if s4.HeadingLevel != 2 {
		t.Errorf("s4: expected heading_level 2, got %d", s4.HeadingLevel)
	}
	if s4.HeadingText != "Another Real" {
		t.Errorf("s4: expected heading_text 'Another Real', got %q", s4.HeadingText)
	}
	if s4.DetectionMethod != "semantic" {
		t.Errorf("s4: expected detection_method 'semantic', got %q", s4.DetectionMethod)
	}
}

func TestExtractSections_JetOctopusLike(t *testing.T) {
	doc := parseHTML(t, `<html><body>
<h1>Main Title</h1>
<p>Intro paragraph</p>
<div class="section-heading">Solutions</div>
<ul><li>Item 1</li><li>Item 2</li></ul>
<div class="section-heading">Unique Features</div>
<p>Feature text here</p>
<div class="section-heading">Values</div>
<p>Value text here</p>
<div class="section-heading">Articles</div>
<ul><li>Article 1</li><li>Article 2</li></ul>
</body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 5 {
		t.Fatalf("Expected 5 sections (1 h1 + 4 visual), got %d", len(sections))
	}

	// s1: "Main Title", body contains "Intro paragraph"
	s1 := sections[0]
	if s1.HeadingText != "Main Title" {
		t.Errorf("s1: expected heading_text 'Main Title', got %q", s1.HeadingText)
	}
	if !strings.Contains(s1.BodyMarkdown, "Intro paragraph") {
		t.Errorf("s1: expected body_markdown to contain 'Intro paragraph', got %q", s1.BodyMarkdown)
	}

	// s2: "Solutions", body contains "Item 1"
	s2 := sections[1]
	if s2.HeadingText != "Solutions" {
		t.Errorf("s2: expected heading_text 'Solutions', got %q", s2.HeadingText)
	}
	if s2.HeadingLevel != 2 {
		t.Errorf("s2: expected heading_level 2 (deepest=1, fallback=2), got %d", s2.HeadingLevel)
	}
	if s2.DetectionMethod != "css_class" {
		t.Errorf("s2: expected detection_method 'css_class', got %q", s2.DetectionMethod)
	}
	if !strings.Contains(s2.BodyMarkdown, "Item 1") {
		t.Errorf("s2: expected body_markdown to contain 'Item 1', got %q", s2.BodyMarkdown)
	}

	// s3: "Unique Features", body contains "Feature text"
	s3 := sections[2]
	if s3.HeadingText != "Unique Features" {
		t.Errorf("s3: expected heading_text 'Unique Features', got %q", s3.HeadingText)
	}
	if s3.DetectionMethod != "css_class" {
		t.Errorf("s3: expected detection_method 'css_class', got %q", s3.DetectionMethod)
	}
	if !strings.Contains(s3.BodyMarkdown, "Feature text") {
		t.Errorf("s3: expected body_markdown to contain 'Feature text', got %q", s3.BodyMarkdown)
	}

	// s4: "Values", body contains "Value text"
	s4 := sections[3]
	if s4.HeadingText != "Values" {
		t.Errorf("s4: expected heading_text 'Values', got %q", s4.HeadingText)
	}
	if s4.DetectionMethod != "css_class" {
		t.Errorf("s4: expected detection_method 'css_class', got %q", s4.DetectionMethod)
	}
	if !strings.Contains(s4.BodyMarkdown, "Value text") {
		t.Errorf("s4: expected body_markdown to contain 'Value text', got %q", s4.BodyMarkdown)
	}

	// s5: "Articles", body contains "Article 1"
	s5 := sections[4]
	if s5.HeadingText != "Articles" {
		t.Errorf("s5: expected heading_text 'Articles', got %q", s5.HeadingText)
	}
	if s5.DetectionMethod != "css_class" {
		t.Errorf("s5: expected detection_method 'css_class', got %q", s5.DetectionMethod)
	}
	if !strings.Contains(s5.BodyMarkdown, "Article 1") {
		t.Errorf("s5: expected body_markdown to contain 'Article 1', got %q", s5.BodyMarkdown)
	}
}

func TestExtractSections_VisualHeadingInlineFormatting(t *testing.T) {
	doc := parseHTML(t, `<html><body><h1>Title</h1><div class="section-heading"><strong>Bold</strong> and <em>italic</em> heading</div><p>Content</p></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 2 {
		t.Fatalf("Expected 2 sections, got %d", len(sections))
	}

	s2 := sections[1]
	// getTextContent adds spaces around inline elements, so adjacent inline
	// elements produce double spaces in the plain-text extraction.
	if s2.HeadingText != "Bold  and  italic  heading" {
		t.Errorf("s2: expected heading_text 'Bold  and  italic  heading' (plain text), got %q", s2.HeadingText)
	}
	if s2.DetectionMethod != "css_class" {
		t.Errorf("s2: expected detection_method 'css_class', got %q", s2.DetectionMethod)
	}
	// Ensure no markdown formatting markers in heading text
	if strings.Contains(s2.HeadingText, "**") {
		t.Errorf("s2: heading_text should not contain markdown bold markers, got %q", s2.HeadingText)
	}
	if strings.Contains(s2.HeadingText, "[") || strings.Contains(s2.HeadingText, "]") {
		t.Errorf("s2: heading_text should not contain markdown link brackets, got %q", s2.HeadingText)
	}
	// Verify it contains the key words as plain text
	if !strings.Contains(s2.HeadingText, "Bold") {
		t.Errorf("s2: heading_text should contain 'Bold', got %q", s2.HeadingText)
	}
	if !strings.Contains(s2.HeadingText, "italic") {
		t.Errorf("s2: heading_text should contain 'italic', got %q", s2.HeadingText)
	}
	if !strings.Contains(s2.HeadingText, "heading") {
		t.Errorf("s2: heading_text should contain 'heading', got %q", s2.HeadingText)
	}
}

func TestExtractSections_FigcaptionHeading(t *testing.T) {
	doc := parseHTML(t, `<html><body><h1>Title</h1><figure><figcaption class="section-heading">Caption Heading</figcaption><p>Figure content</p></figure></body></html>`)
	sections := ExtractSections(doc)

	if len(sections) != 2 {
		t.Fatalf("Expected 2 sections, got %d", len(sections))
	}

	s2 := sections[1]
	if s2.HeadingText != "Caption Heading" {
		t.Errorf("s2: expected heading_text 'Caption Heading', got %q", s2.HeadingText)
	}
	if s2.DetectionMethod != "css_class" {
		t.Errorf("s2: expected detection_method 'css_class', got %q", s2.DetectionMethod)
	}
}

// --- Fixture-based integration tests ---

func loadFixture(t *testing.T, name string) *goquery.Document {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("Failed to read fixture: %v", err)
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(data)))
	if err != nil {
		t.Fatalf("Failed to parse fixture: %v", err)
	}
	return doc
}

func TestExtractSections_Fixture_CSSClassLanding(t *testing.T) {
	doc := loadFixture(t, "css_class_landing.html")
	sections := ExtractSections(doc)

	if len(sections) != 5 {
		for i, s := range sections {
			t.Logf("  s%d: level=%d method=%q text=%q body=%q", i+1, s.HeadingLevel, s.DetectionMethod, s.HeadingText, s.BodyMarkdown)
		}
		t.Fatalf("Expected 5 sections, got %d", len(sections))
	}

	// s1: level=1, text="JavaScript Rendering for SEO", method="semantic"
	s1 := sections[0]
	if s1.HeadingLevel != 1 {
		t.Errorf("s1: expected heading_level 1, got %d", s1.HeadingLevel)
	}
	if s1.HeadingText != "JavaScript Rendering for SEO" {
		t.Errorf("s1: expected heading_text 'JavaScript Rendering for SEO', got %q", s1.HeadingText)
	}
	if s1.DetectionMethod != "semantic" {
		t.Errorf("s1: expected detection_method 'semantic', got %q", s1.DetectionMethod)
	}

	// s2: level=2, text="Solutions", method="css_class"
	s2 := sections[1]
	if s2.HeadingLevel != 2 {
		t.Errorf("s2: expected heading_level 2, got %d", s2.HeadingLevel)
	}
	if s2.HeadingText != "Solutions" {
		t.Errorf("s2: expected heading_text 'Solutions', got %q", s2.HeadingText)
	}
	if s2.DetectionMethod != "css_class" {
		t.Errorf("s2: expected detection_method 'css_class', got %q", s2.DetectionMethod)
	}
	if !strings.Contains(s2.BodyMarkdown, "rendering solutions") {
		t.Errorf("s2: expected body to contain 'rendering solutions', got %q", s2.BodyMarkdown)
	}

	// s3: level=2, text="Unique Features", method="css_class"
	s3 := sections[2]
	if s3.HeadingLevel != 2 {
		t.Errorf("s3: expected heading_level 2, got %d", s3.HeadingLevel)
	}
	if s3.HeadingText != "Unique Features" {
		t.Errorf("s3: expected heading_text 'Unique Features', got %q", s3.HeadingText)
	}
	if s3.DetectionMethod != "css_class" {
		t.Errorf("s3: expected detection_method 'css_class', got %q", s3.DetectionMethod)
	}

	// s4: level=2, text="Pricing", method="css_class"
	s4 := sections[3]
	if s4.HeadingLevel != 2 {
		t.Errorf("s4: expected heading_level 2, got %d", s4.HeadingLevel)
	}
	if s4.HeadingText != "Pricing" {
		t.Errorf("s4: expected heading_text 'Pricing', got %q", s4.HeadingText)
	}
	if s4.DetectionMethod != "css_class" {
		t.Errorf("s4: expected detection_method 'css_class', got %q", s4.DetectionMethod)
	}

	// s5: level=2, text="FAQ", method="css_class"
	s5 := sections[4]
	if s5.HeadingLevel != 2 {
		t.Errorf("s5: expected heading_level 2, got %d", s5.HeadingLevel)
	}
	if s5.HeadingText != "FAQ" {
		t.Errorf("s5: expected heading_text 'FAQ', got %q", s5.HeadingText)
	}
	if s5.DetectionMethod != "css_class" {
		t.Errorf("s5: expected detection_method 'css_class', got %q", s5.DetectionMethod)
	}
	if !strings.Contains(s5.BodyMarkdown, "Frequently asked") {
		t.Errorf("s5: expected body to contain 'Frequently asked', got %q", s5.BodyMarkdown)
	}
}

func TestExtractSections_Fixture_ARIAWebapp(t *testing.T) {
	doc := loadFixture(t, "aria_webapp.html")
	sections := ExtractSections(doc)

	if len(sections) != 4 {
		for i, s := range sections {
			t.Logf("  s%d: level=%d method=%q text=%q", i+1, s.HeadingLevel, s.DetectionMethod, s.HeadingText)
		}
		t.Fatalf("Expected 4 sections, got %d", len(sections))
	}

	expectedTexts := []string{"Dashboard Overview", "Recent Activity", "Statistics", "Settings"}
	expectedLevels := []int{2, 3, 3, 2}

	for i, s := range sections {
		if s.HeadingText != expectedTexts[i] {
			t.Errorf("s%d: expected heading_text %q, got %q", i+1, expectedTexts[i], s.HeadingText)
		}
		if s.HeadingLevel != expectedLevels[i] {
			t.Errorf("s%d: expected heading_level %d, got %d", i+1, expectedLevels[i], s.HeadingLevel)
		}
		if s.DetectionMethod != "aria" {
			t.Errorf("s%d: expected detection_method 'aria', got %q", i+1, s.DetectionMethod)
		}
	}
}

func TestExtractSections_Fixture_MixedDetection(t *testing.T) {
	doc := loadFixture(t, "mixed_detection.html")
	sections := ExtractSections(doc)

	if len(sections) != 5 {
		for i, s := range sections {
			t.Logf("  s%d: level=%d method=%q text=%q body=%q", i+1, s.HeadingLevel, s.DetectionMethod, s.HeadingText, s.BodyMarkdown)
		}
		t.Fatalf("Expected 5 sections, got %d", len(sections))
	}

	// s1: method="semantic", text="Page Title"
	s1 := sections[0]
	if s1.DetectionMethod != "semantic" {
		t.Errorf("s1: expected detection_method 'semantic', got %q", s1.DetectionMethod)
	}
	if s1.HeadingText != "Page Title" {
		t.Errorf("s1: expected heading_text 'Page Title', got %q", s1.HeadingText)
	}

	// s2: method="aria", text="ARIA Section"
	s2 := sections[1]
	if s2.DetectionMethod != "aria" {
		t.Errorf("s2: expected detection_method 'aria', got %q", s2.DetectionMethod)
	}
	if s2.HeadingText != "ARIA Section" {
		t.Errorf("s2: expected heading_text 'ARIA Section', got %q", s2.HeadingText)
	}

	// s3: method="aria", text="ARIA Priority Test" (ARIA wins over CSS class)
	s3 := sections[2]
	if s3.DetectionMethod != "aria" {
		t.Errorf("s3: expected detection_method 'aria' (ARIA wins over CSS class), got %q", s3.DetectionMethod)
	}
	if s3.HeadingText != "ARIA Priority Test" {
		t.Errorf("s3: expected heading_text 'ARIA Priority Test', got %q", s3.HeadingText)
	}

	// s4: method="css_class", text="CSS Section"
	s4 := sections[3]
	if s4.DetectionMethod != "css_class" {
		t.Errorf("s4: expected detection_method 'css_class', got %q", s4.DetectionMethod)
	}
	if s4.HeadingText != "CSS Section" {
		t.Errorf("s4: expected heading_text 'CSS Section', got %q", s4.HeadingText)
	}

	// s5: method="data_attr", text="Data Section"
	s5 := sections[4]
	if s5.DetectionMethod != "data_attr" {
		t.Errorf("s5: expected detection_method 'data_attr', got %q", s5.DetectionMethod)
	}
	if s5.HeadingText != "Data Section" {
		t.Errorf("s5: expected heading_text 'Data Section', got %q", s5.HeadingText)
	}

	// The "Not a Heading" div (data-heading="false") should appear in s5's body text
	if !strings.Contains(s5.BodyMarkdown, "Not a Heading") {
		t.Errorf("s5: expected body to contain 'Not a Heading' (data-heading=false rejected), got %q", s5.BodyMarkdown)
	}
}

func TestExtractSections_Fixture_DeepNesting(t *testing.T) {
	doc := loadFixture(t, "deep_nesting.html")
	sections := ExtractSections(doc)

	if len(sections) != 4 {
		for i, s := range sections {
			t.Logf("  s%d: level=%d method=%q text=%q", i+1, s.HeadingLevel, s.DetectionMethod, s.HeadingText)
		}
		t.Fatalf("Expected 4 sections, got %d", len(sections))
	}

	// s1: level=1, method="semantic", text="Page Title"
	s1 := sections[0]
	if s1.HeadingLevel != 1 {
		t.Errorf("s1: expected heading_level 1, got %d", s1.HeadingLevel)
	}
	if s1.DetectionMethod != "semantic" {
		t.Errorf("s1: expected detection_method 'semantic', got %q", s1.DetectionMethod)
	}
	if s1.HeadingText != "Page Title" {
		t.Errorf("s1: expected heading_text 'Page Title', got %q", s1.HeadingText)
	}

	// s2-s4: level=2, method="css_class"
	expectedTexts := []string{"Wrapped Heading", "Deep in Article", "Span Deep"}
	for i := 1; i < 4; i++ {
		s := sections[i]
		if s.HeadingLevel != 2 {
			t.Errorf("s%d: expected heading_level 2, got %d", i+1, s.HeadingLevel)
		}
		if s.DetectionMethod != "css_class" {
			t.Errorf("s%d: expected detection_method 'css_class', got %q", i+1, s.DetectionMethod)
		}
		if s.HeadingText != expectedTexts[i-1] {
			t.Errorf("s%d: expected heading_text %q, got %q", i+1, expectedTexts[i-1], s.HeadingText)
		}
	}
}

func TestExtractSections_Fixture_FalsePositiveStress(t *testing.T) {
	doc := loadFixture(t, "false_positive_stress.html")
	sections := ExtractSections(doc)

	if len(sections) != 4 {
		for i, s := range sections {
			t.Logf("  s%d: level=%d method=%q text=%q body=%q", i+1, s.HeadingLevel, s.DetectionMethod, s.HeadingText, s.BodyMarkdown)
		}
		t.Fatalf("Expected 4 sections, got %d", len(sections))
	}

	// s1: level=1, text="Real Title", method="semantic"
	s1 := sections[0]
	if s1.HeadingLevel != 1 {
		t.Errorf("s1: expected heading_level 1, got %d", s1.HeadingLevel)
	}
	if s1.HeadingText != "Real Title" {
		t.Errorf("s1: expected heading_text 'Real Title', got %q", s1.HeadingText)
	}
	if s1.DetectionMethod != "semantic" {
		t.Errorf("s1: expected detection_method 'semantic', got %q", s1.DetectionMethod)
	}

	// s2: level=2, text="Grouped Heading", method="semantic" (the h2 inside heading-group)
	s2 := sections[1]
	if s2.HeadingLevel != 2 {
		t.Errorf("s2: expected heading_level 2, got %d", s2.HeadingLevel)
	}
	if s2.HeadingText != "Grouped Heading" {
		t.Errorf("s2: expected heading_text 'Grouped Heading', got %q", s2.HeadingText)
	}
	if s2.DetectionMethod != "semantic" {
		t.Errorf("s2: expected detection_method 'semantic', got %q", s2.DetectionMethod)
	}

	// s3: text="Legit Visual", method="css_class", level=3 (deepest=2, fallback=3)
	s3 := sections[2]
	if s3.HeadingLevel != 3 {
		t.Errorf("s3: expected heading_level 3 (deepest=2, fallback=3), got %d", s3.HeadingLevel)
	}
	if s3.HeadingText != "Legit Visual" {
		t.Errorf("s3: expected heading_text 'Legit Visual', got %q", s3.HeadingText)
	}
	if s3.DetectionMethod != "css_class" {
		t.Errorf("s3: expected detection_method 'css_class', got %q", s3.DetectionMethod)
	}

	// s4: level=2, text="Another Real", method="semantic"
	s4 := sections[3]
	if s4.HeadingLevel != 2 {
		t.Errorf("s4: expected heading_level 2, got %d", s4.HeadingLevel)
	}
	if s4.HeadingText != "Another Real" {
		t.Errorf("s4: expected heading_text 'Another Real', got %q", s4.HeadingText)
	}
	if s4.DetectionMethod != "semantic" {
		t.Errorf("s4: expected detection_method 'semantic', got %q", s4.DetectionMethod)
	}
}

func TestExtractSections_Fixture_NoSemanticHeadings(t *testing.T) {
	doc := loadFixture(t, "no_semantic_headings.html")
	sections := ExtractSections(doc)

	if len(sections) != 5 {
		for i, s := range sections {
			t.Logf("  s%d: level=%d method=%q text=%q body=%q", i+1, s.HeadingLevel, s.DetectionMethod, s.HeadingText, s.BodyMarkdown)
		}
		t.Fatalf("Expected 5 sections, got %d", len(sections))
	}

	// s1: intro, level=0, method="", body contains "Intro paragraph"
	s1 := sections[0]
	if s1.HeadingLevel != 0 {
		t.Errorf("s1: expected heading_level 0, got %d", s1.HeadingLevel)
	}
	if s1.DetectionMethod != "" {
		t.Errorf("s1: expected detection_method '' (empty), got %q", s1.DetectionMethod)
	}
	if !strings.Contains(s1.BodyMarkdown, "Intro paragraph") {
		t.Errorf("s1: expected body to contain 'Intro paragraph', got %q", s1.BodyMarkdown)
	}

	// s2: level=2, text="First Section", method="css_class"
	s2 := sections[1]
	if s2.HeadingLevel != 2 {
		t.Errorf("s2: expected heading_level 2, got %d", s2.HeadingLevel)
	}
	if s2.HeadingText != "First Section" {
		t.Errorf("s2: expected heading_text 'First Section', got %q", s2.HeadingText)
	}
	if s2.DetectionMethod != "css_class" {
		t.Errorf("s2: expected detection_method 'css_class', got %q", s2.DetectionMethod)
	}

	// s3: level=2, text="Second Section", method="css_class"
	s3 := sections[2]
	if s3.HeadingLevel != 2 {
		t.Errorf("s3: expected heading_level 2, got %d", s3.HeadingLevel)
	}
	if s3.HeadingText != "Second Section" {
		t.Errorf("s3: expected heading_text 'Second Section', got %q", s3.HeadingText)
	}
	if s3.DetectionMethod != "css_class" {
		t.Errorf("s3: expected detection_method 'css_class', got %q", s3.DetectionMethod)
	}

	// s4: level=2, text="Third Section", method="data_attr"
	s4 := sections[3]
	if s4.HeadingLevel != 2 {
		t.Errorf("s4: expected heading_level 2, got %d", s4.HeadingLevel)
	}
	if s4.HeadingText != "Third Section" {
		t.Errorf("s4: expected heading_text 'Third Section', got %q", s4.HeadingText)
	}
	if s4.DetectionMethod != "data_attr" {
		t.Errorf("s4: expected detection_method 'data_attr', got %q", s4.DetectionMethod)
	}

	// s5: level=2, text="Fourth Section", method="css_class"
	s5 := sections[4]
	if s5.HeadingLevel != 2 {
		t.Errorf("s5: expected heading_level 2, got %d", s5.HeadingLevel)
	}
	if s5.HeadingText != "Fourth Section" {
		t.Errorf("s5: expected heading_text 'Fourth Section', got %q", s5.HeadingText)
	}
	if s5.DetectionMethod != "css_class" {
		t.Errorf("s5: expected detection_method 'css_class', got %q", s5.DetectionMethod)
	}
}

func TestExtractSections_Fixture_HeadingsInRemoved(t *testing.T) {
	doc := loadFixture(t, "headings_in_removed.html")
	sections := ExtractSections(doc)

	if len(sections) != 2 {
		for i, s := range sections {
			t.Logf("  s%d: level=%d method=%q text=%q", i+1, s.HeadingLevel, s.DetectionMethod, s.HeadingText)
		}
		t.Fatalf("Expected 2 sections (headings in header/nav/footer stripped), got %d", len(sections))
	}

	// s1: level=1, text="Page Title", method="semantic"
	s1 := sections[0]
	if s1.HeadingLevel != 1 {
		t.Errorf("s1: expected heading_level 1, got %d", s1.HeadingLevel)
	}
	if s1.HeadingText != "Page Title" {
		t.Errorf("s1: expected heading_text 'Page Title', got %q", s1.HeadingText)
	}
	if s1.DetectionMethod != "semantic" {
		t.Errorf("s1: expected detection_method 'semantic', got %q", s1.DetectionMethod)
	}

	// s2: text="Real Visual", method="css_class"
	s2 := sections[1]
	if s2.HeadingText != "Real Visual" {
		t.Errorf("s2: expected heading_text 'Real Visual', got %q", s2.HeadingText)
	}
	if s2.DetectionMethod != "css_class" {
		t.Errorf("s2: expected detection_method 'css_class', got %q", s2.DetectionMethod)
	}

	// Verify none of the stripped headings appear
	for _, s := range sections {
		if strings.Contains(s.HeadingText, "Site Title") {
			t.Errorf("'Site Title' from header should be stripped, found in heading_text: %q", s.HeadingText)
		}
		if strings.Contains(s.HeadingText, "Nav Section") {
			t.Errorf("'Nav Section' from nav should be stripped, found in heading_text: %q", s.HeadingText)
		}
		if strings.Contains(s.HeadingText, "Footer Heading") {
			t.Errorf("'Footer Heading' from footer should be stripped, found in heading_text: %q", s.HeadingText)
		}
	}
}

func TestExtractSections_Fixture_ComparePair(t *testing.T) {
	// --- Load and parse both fixtures ---
	httpDoc := loadFixture(t, "compare_pair_http.html")
	httpSections := ExtractSections(httpDoc)

	jsDoc := loadFixture(t, "compare_pair_js.html")
	jsSections := ExtractSections(jsDoc)

	// --- Assert HTTP version: 4 sections ---
	if len(httpSections) != 4 {
		for i, s := range httpSections {
			t.Logf("  http s%d: level=%d method=%q text=%q body=%q", i+1, s.HeadingLevel, s.DetectionMethod, s.HeadingText, s.BodyMarkdown)
		}
		t.Fatalf("HTTP: expected 4 sections, got %d", len(httpSections))
	}

	// h1 Product Page: level=1, semantic
	httpS1 := httpSections[0]
	if httpS1.HeadingLevel != 1 {
		t.Errorf("HTTP s1: expected heading_level 1, got %d", httpS1.HeadingLevel)
	}
	if httpS1.HeadingText != "Product Page" {
		t.Errorf("HTTP s1: expected heading_text 'Product Page', got %q", httpS1.HeadingText)
	}
	if httpS1.DetectionMethod != "semantic" {
		t.Errorf("HTTP s1: expected detection_method 'semantic', got %q", httpS1.DetectionMethod)
	}

	// 3 CSS class headings: level=2, css_class
	httpExpectedTexts := []string{"Features", "Pricing", "FAQ"}
	for i := 1; i <= 3; i++ {
		s := httpSections[i]
		if s.HeadingLevel != 2 {
			t.Errorf("HTTP s%d: expected heading_level 2, got %d", i+1, s.HeadingLevel)
		}
		if s.HeadingText != httpExpectedTexts[i-1] {
			t.Errorf("HTTP s%d: expected heading_text %q, got %q", i+1, httpExpectedTexts[i-1], s.HeadingText)
		}
		if s.DetectionMethod != "css_class" {
			t.Errorf("HTTP s%d: expected detection_method 'css_class', got %q", i+1, s.DetectionMethod)
		}
	}

	// --- Assert JS version: 4 sections ---
	if len(jsSections) != 4 {
		for i, s := range jsSections {
			t.Logf("  js s%d: level=%d method=%q text=%q body=%q", i+1, s.HeadingLevel, s.DetectionMethod, s.HeadingText, s.BodyMarkdown)
		}
		t.Fatalf("JS: expected 4 sections, got %d", len(jsSections))
	}

	// h1 Product Page: level=1, semantic
	jsS1 := jsSections[0]
	if jsS1.HeadingLevel != 1 {
		t.Errorf("JS s1: expected heading_level 1, got %d", jsS1.HeadingLevel)
	}
	if jsS1.HeadingText != "Product Page" {
		t.Errorf("JS s1: expected heading_text 'Product Page', got %q", jsS1.HeadingText)
	}
	if jsS1.DetectionMethod != "semantic" {
		t.Errorf("JS s1: expected detection_method 'semantic', got %q", jsS1.DetectionMethod)
	}

	// 3 h2 headings: level=2, semantic
	jsExpectedTexts := []string{"Features", "Pricing", "FAQ"}
	for i := 1; i <= 3; i++ {
		s := jsSections[i]
		if s.HeadingLevel != 2 {
			t.Errorf("JS s%d: expected heading_level 2, got %d", i+1, s.HeadingLevel)
		}
		if s.HeadingText != jsExpectedTexts[i-1] {
			t.Errorf("JS s%d: expected heading_text %q, got %q", i+1, jsExpectedTexts[i-1], s.HeadingText)
		}
		if s.DetectionMethod != "semantic" {
			t.Errorf("JS s%d: expected detection_method 'semantic', got %q", i+1, s.DetectionMethod)
		}
	}

	// --- Diff sections: JS vs HTTP ---
	diffs := compare.DiffSections(jsSections, httpSections)

	if len(diffs) != 4 {
		for i, d := range diffs {
			t.Logf("  diff %d: heading=%q level=%d status=%q", i, d.HeadingText, d.HeadingLevel, d.Status)
		}
		t.Fatalf("Expected 4 section diffs, got %d", len(diffs))
	}

	// Product Page (h1): matched by exact key h1:product page. Same body -> unchanged.
	d0 := diffs[0]
	if d0.HeadingText != "Product Page" {
		t.Errorf("diff[0]: expected heading_text 'Product Page', got %q", d0.HeadingText)
	}
	if d0.HeadingLevel != 1 {
		t.Errorf("diff[0]: expected heading_level 1, got %d", d0.HeadingLevel)
	}
	if d0.Status != "unchanged" {
		t.Errorf("diff[0]: expected status 'unchanged', got %q", d0.Status)
	}

	// Features: exact match on h2:features. Different body -> changed.
	d1 := diffs[1]
	if d1.HeadingText != "Features" {
		t.Errorf("diff[1]: expected heading_text 'Features', got %q", d1.HeadingText)
	}
	if d1.HeadingLevel != 2 {
		t.Errorf("diff[1]: expected heading_level 2, got %d", d1.HeadingLevel)
	}
	if d1.Status != "changed" {
		t.Errorf("diff[1]: expected status 'changed', got %q", d1.Status)
	}

	// Pricing: exact match on h2:pricing. Different body -> changed.
	d2 := diffs[2]
	if d2.HeadingText != "Pricing" {
		t.Errorf("diff[2]: expected heading_text 'Pricing', got %q", d2.HeadingText)
	}
	if d2.HeadingLevel != 2 {
		t.Errorf("diff[2]: expected heading_level 2, got %d", d2.HeadingLevel)
	}
	if d2.Status != "changed" {
		t.Errorf("diff[2]: expected status 'changed', got %q", d2.Status)
	}

	// FAQ: exact match on h2:faq. Different body -> changed.
	d3 := diffs[3]
	if d3.HeadingText != "FAQ" {
		t.Errorf("diff[3]: expected heading_text 'FAQ', got %q", d3.HeadingText)
	}
	if d3.HeadingLevel != 2 {
		t.Errorf("diff[3]: expected heading_level 2, got %d", d3.HeadingLevel)
	}
	if d3.Status != "changed" {
		t.Errorf("diff[3]: expected status 'changed', got %q", d3.Status)
	}

	// No sections should be added_by_js or removed_by_js.
	for i, d := range diffs {
		if d.Status == "added_by_js" {
			t.Errorf("diff[%d]: unexpected status 'added_by_js' for heading %q", i, d.HeadingText)
		}
		if d.Status == "removed_by_js" {
			t.Errorf("diff[%d]: unexpected status 'removed_by_js' for heading %q", i, d.HeadingText)
		}
	}
}
