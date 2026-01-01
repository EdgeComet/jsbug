package parser

import (
	"strings"
	"testing"
)

func TestNewParser(t *testing.T) {
	p := NewParser()
	if p == nil {
		t.Fatal("NewParser() returned nil")
	}
}

func TestParser_Parse_Title(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "simple title",
			html:     `<html><head><title>Hello World</title></head></html>`,
			expected: "Hello World",
		},
		{
			name:     "title with whitespace (normalized)",
			html:     `<html><head><title>  Spaced Title  </title></head></html>`,
			expected: "Spaced Title",
		},
		{
			name:     "title with newlines and multiple spaces",
			html:     "<html><head><title>Title\n\n  with   newlines</title></head></html>",
			expected: "Title with newlines",
		},
		{
			name:     "missing title",
			html:     `<html><head></head></html>`,
			expected: "",
		},
		{
			name:     "multiple titles (first wins)",
			html:     `<html><head><title>First</title><title>Second</title></head></html>`,
			expected: "First",
		},
	}

	p := NewParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := p.Parse(tt.html, "https://example.com")
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if result.Title != tt.expected {
				t.Errorf("Title = %q, want %q", result.Title, tt.expected)
			}
		})
	}
}

func TestParser_Parse_MetaDescription(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "simple description",
			html:     `<html><head><meta name="description" content="Page description"></head></html>`,
			expected: "Page description",
		},
		{
			name:     "description case insensitive",
			html:     `<html><head><meta name="Description" content="Case test"></head></html>`,
			expected: "Case test",
		},
		{
			name:     "missing description",
			html:     `<html><head></head></html>`,
			expected: "",
		},
		{
			name:     "description with newlines and multiple spaces (normalized)",
			html:     "<html><head><meta name=\"description\" content=\"Description\n  with   extra   spaces\"></head></html>",
			expected: "Description with extra spaces",
		},
	}

	p := NewParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := p.Parse(tt.html, "https://example.com")
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if result.MetaDescription != tt.expected {
				t.Errorf("MetaDescription = %q, want %q", result.MetaDescription, tt.expected)
			}
		})
	}
}

func TestParser_Parse_CanonicalURL(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "canonical link",
			html:     `<html><head><link rel="canonical" href="https://example.com/page"></head></html>`,
			expected: "https://example.com/page",
		},
		{
			name:     "canonical case insensitive",
			html:     `<html><head><link rel="Canonical" href="https://example.com/page"></head></html>`,
			expected: "https://example.com/page",
		},
		{
			name:     "missing canonical",
			html:     `<html><head></head></html>`,
			expected: "",
		},
	}

	p := NewParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := p.Parse(tt.html, "https://example.com")
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if result.CanonicalURL != tt.expected {
				t.Errorf("CanonicalURL = %q, want %q", result.CanonicalURL, tt.expected)
			}
		})
	}
}

func TestParser_Parse_MetaRobots(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "robots meta",
			html:     `<html><head><meta name="robots" content="noindex, nofollow"></head></html>`,
			expected: "noindex, nofollow",
		},
		{
			name:     "missing robots",
			html:     `<html><head></head></html>`,
			expected: "",
		},
	}

	p := NewParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := p.Parse(tt.html, "https://example.com")
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if result.MetaRobots != tt.expected {
				t.Errorf("MetaRobots = %q, want %q", result.MetaRobots, tt.expected)
			}
		})
	}
}

func TestParser_Parse_Headings(t *testing.T) {
	html := `
<html>
<body>
	<h1>First H1</h1>
	<h1>Second H1</h1>
	<h2>First H2</h2>
	<h2>Second H2</h2>
	<h2>Third H2</h2>
	<h3>First H3</h3>
</body>
</html>`

	p := NewParser()
	result, err := p.Parse(html, "https://example.com")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(result.H1) != 2 {
		t.Errorf("H1 count = %d, want 2", len(result.H1))
	}
	if len(result.H2) != 3 {
		t.Errorf("H2 count = %d, want 3", len(result.H2))
	}
	if len(result.H3) != 1 {
		t.Errorf("H3 count = %d, want 1", len(result.H3))
	}

	if result.H1[0] != "First H1" {
		t.Errorf("H1[0] = %q, want %q", result.H1[0], "First H1")
	}
}

func TestParser_Parse_Headings_Deduplication(t *testing.T) {
	html := `
<html>
<body>
	<h1>Same H1</h1>
	<h1>Same H1</h1>
	<h1>Different H1</h1>
	<h2>Repeated H2</h2>
	<h2>Repeated H2</h2>
	<h2>Repeated H2</h2>
</body>
</html>`

	p := NewParser()
	result, err := p.Parse(html, "https://example.com")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Duplicates should be removed
	if len(result.H1) != 2 {
		t.Errorf("H1 count = %d, want 2 (deduplicated)", len(result.H1))
	}
	if len(result.H2) != 1 {
		t.Errorf("H2 count = %d, want 1 (deduplicated)", len(result.H2))
	}
}

func TestParser_Parse_Headings_Normalization(t *testing.T) {
	html := `
<html>
<body>
	<h1>Heading  with   multiple    spaces</h1>
	<h1>Heading
with
newlines</h1>
</body>
</html>`

	p := NewParser()
	result, err := p.Parse(html, "https://example.com")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(result.H1) != 2 {
		t.Errorf("H1 count = %d, want 2", len(result.H1))
	}
	if result.H1[0] != "Heading with multiple spaces" {
		t.Errorf("H1[0] = %q, want %q", result.H1[0], "Heading with multiple spaces")
	}
	if result.H1[1] != "Heading with newlines" {
		t.Errorf("H1[1] = %q, want %q", result.H1[1], "Heading with newlines")
	}
}

func TestParser_Parse_Links(t *testing.T) {
	html := `
<html>
<body>
	<a href="/page1">Internal relative</a>
	<a href="https://example.com/page2">Internal absolute</a>
	<a href="https://other.com/page">External</a>
	<a href="https://EXAMPLE.COM/page3">Internal case insensitive</a>
	<a href="javascript:void(0)">JavaScript</a>
	<a href="mailto:test@example.com">Email</a>
	<a href="tel:+1234567890">Phone</a>
	<a href="#section">Anchor</a>
	<a href="">Empty</a>
</body>
</html>`

	p := NewParser()
	result, err := p.Parse(html, "https://example.com")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if result.InternalLinks != 3 {
		t.Errorf("InternalLinks = %d, want 3", result.InternalLinks)
	}
	if result.ExternalLinks != 1 {
		t.Errorf("ExternalLinks = %d, want 1", result.ExternalLinks)
	}
}

func TestParser_Parse_WordCount(t *testing.T) {
	html := `
<html>
<head>
	<title>Title</title>
	<script>var x = "ignored script";</script>
	<style>body { color: black; }</style>
</head>
<body>
	<h1>Hello World</h1>
	<p>This is a paragraph with some words.</p>
	<script>console.log("also ignored");</script>
	<noscript>Noscript ignored</noscript>
	<div>More   text   with   spaces.</div>
</body>
</html>`

	p := NewParser()
	result, err := p.Parse(html, "https://example.com")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Expected: "Hello World" (2) + "This is a paragraph with some words." (7) + "More text with spaces." (4) = 13
	if result.WordCount < 10 || result.WordCount > 20 {
		t.Errorf("WordCount = %d, expected around 13", result.WordCount)
	}
}

func TestParser_Parse_OpenGraph(t *testing.T) {
	html := `
<html>
<head>
	<meta property="og:title" content="OG Title">
	<meta property="og:description" content="OG Description">
	<meta property="og:image" content="https://example.com/image.jpg">
	<meta property="og:url" content="https://example.com/page">
	<meta property="og:type" content="article">
</head>
</html>`

	p := NewParser()
	result, err := p.Parse(html, "https://example.com")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(result.OpenGraph) != 5 {
		t.Errorf("OpenGraph count = %d, want 5", len(result.OpenGraph))
	}

	if result.OpenGraph["title"] != "OG Title" {
		t.Errorf("og:title = %q, want %q", result.OpenGraph["title"], "OG Title")
	}
	if result.OpenGraph["image"] != "https://example.com/image.jpg" {
		t.Errorf("og:image = %q", result.OpenGraph["image"])
	}
}

func TestParser_Parse_StructuredData(t *testing.T) {
	html := `
<html>
<head>
	<script type="application/ld+json">
	{
		"@context": "https://schema.org",
		"@type": "Article",
		"headline": "Test Article"
	}
	</script>
	<script type="application/ld+json">
	{
		"@context": "https://schema.org",
		"@type": "Organization",
		"name": "Test Org"
	}
	</script>
	<script type="text/javascript">
		// Regular script, should be ignored
		var x = 1;
	</script>
</head>
</html>`

	p := NewParser()
	result, err := p.Parse(html, "https://example.com")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(result.StructuredData) != 2 {
		t.Errorf("StructuredData count = %d, want 2", len(result.StructuredData))
	}
}

func TestParser_Parse_StructuredData_InvalidJSON(t *testing.T) {
	html := `
<html>
<head>
	<script type="application/ld+json">
	{invalid json}
	</script>
	<script type="application/ld+json">
	{"valid": "json"}
	</script>
</head>
</html>`

	p := NewParser()
	result, err := p.Parse(html, "https://example.com")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Only valid JSON should be included
	if len(result.StructuredData) != 1 {
		t.Errorf("StructuredData count = %d, want 1 (only valid JSON)", len(result.StructuredData))
	}
}

func TestParser_Parse_MalformedHTML(t *testing.T) {
	// Meta before unclosed title - parser can extract meta
	html := `
<html>
<head>
	<meta name="description" content="Test">
	<title>Unclosed title
<body>
	<h1>Unclosed heading
	<p>Unclosed paragraph
</html>`

	p := NewParser()
	result, err := p.Parse(html, "https://example.com")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Parser should still extract what it can
	if result.MetaDescription != "Test" {
		t.Errorf("MetaDescription = %q, want %q", result.MetaDescription, "Test")
	}
}

func TestParser_Parse_EmptyHTML(t *testing.T) {
	p := NewParser()
	result, err := p.Parse("", "https://example.com")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if result.Title != "" {
		t.Errorf("Title should be empty")
	}
	if result.WordCount != 0 {
		t.Errorf("WordCount = %d, want 0", result.WordCount)
	}
}

func TestParser_Parse_CompleteExample(t *testing.T) {
	html := `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<title>Complete Test Page</title>
	<meta name="description" content="A complete test page for parsing">
	<meta name="robots" content="index, follow">
	<link rel="canonical" href="https://example.com/complete">
	<meta property="og:title" content="OG Complete">
	<meta property="og:description" content="OG Description">
	<script type="application/ld+json">
	{"@type": "WebPage", "name": "Test"}
	</script>
</head>
<body>
	<h1>Main Heading</h1>
	<p>This is the main content of the page with several words.</p>
	<h2>Section One</h2>
	<p>More content here.</p>
	<a href="/internal">Internal Link</a>
	<a href="https://external.com">External Link</a>
	<h2>Section Two</h2>
	<h3>Subsection</h3>
</body>
</html>`

	p := NewParser()
	result, err := p.Parse(html, "https://example.com")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if result.Title != "Complete Test Page" {
		t.Errorf("Title = %q", result.Title)
	}
	if result.MetaDescription != "A complete test page for parsing" {
		t.Errorf("MetaDescription = %q", result.MetaDescription)
	}
	if result.MetaRobots != "index, follow" {
		t.Errorf("MetaRobots = %q", result.MetaRobots)
	}
	if result.CanonicalURL != "https://example.com/complete" {
		t.Errorf("CanonicalURL = %q", result.CanonicalURL)
	}
	if len(result.H1) != 1 || result.H1[0] != "Main Heading" {
		t.Errorf("H1 = %v", result.H1)
	}
	if len(result.H2) != 2 {
		t.Errorf("H2 count = %d, want 2", len(result.H2))
	}
	if len(result.H3) != 1 {
		t.Errorf("H3 count = %d, want 1", len(result.H3))
	}
	if result.InternalLinks != 1 {
		t.Errorf("InternalLinks = %d, want 1", result.InternalLinks)
	}
	if result.ExternalLinks != 1 {
		t.Errorf("ExternalLinks = %d, want 1", result.ExternalLinks)
	}
	if len(result.OpenGraph) != 2 {
		t.Errorf("OpenGraph count = %d, want 2", len(result.OpenGraph))
	}
	if len(result.StructuredData) != 1 {
		t.Errorf("StructuredData count = %d, want 1", len(result.StructuredData))
	}
	if result.WordCount < 10 {
		t.Errorf("WordCount = %d, expected >= 10", result.WordCount)
	}
}

func TestParseWithOptions_BodyMarkdown(t *testing.T) {
	html := `<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body>
    <h1>Welcome</h1>
    <p>This is a <strong>test</strong> page.</p>
    <nav>Home | About</nav>
</body>
</html>`

	parser := NewParser()
	result, err := parser.ParseWithOptions(html, ParseOptions{
		PageURL: "https://example.com",
	})
	if err != nil {
		t.Fatalf("ParseWithOptions() error = %v", err)
	}

	// Verify body_markdown contains expected structure
	if !strings.Contains(result.BodyMarkdown, "# Welcome") {
		t.Error("BodyMarkdown should contain h1 as markdown heading")
	}
	if !strings.Contains(result.BodyMarkdown, "**test**") {
		t.Error("BodyMarkdown should contain bold formatting")
	}
	if !strings.Contains(result.BodyMarkdown, "[NAV]") {
		t.Error("BodyMarkdown should contain semantic section label")
	}
}

func TestFullMarkdownExtraction(t *testing.T) {
	html := `<!DOCTYPE html>
<html>
<body>
    <header>
        <nav aria-label="Main">
            <a href="https://example.com">Home</a>
        </nav>
    </header>
    <main>
        <article>
            <h1>Welcome</h1>
            <p>This is <strong>important</strong> content.</p>
            <ul>
                <li>First item</li>
                <li>Second item</li>
            </ul>
            <table>
                <tr><td>A</td><td>B</td></tr>
            </table>
        </article>
    </main>
    <footer>Copyright 2024</footer>
</body>
</html>`

	parser := NewParser()
	result, err := parser.ParseWithOptions(html, ParseOptions{
		PageURL: "https://example.com",
	})
	if err != nil {
		t.Fatalf("ParseWithOptions() error = %v", err)
	}

	// Verify structure - per spec, only innermost semantic labels appear
	// header contains nav, so only [NAV] appears (not [HEADER])
	// main contains article, so only [ARTICLE] appears (not [MAIN CONTENT])
	expected := []string{
		"[NAV: Main]",
		"Home", // Link text inside nav section (rendered as plain text)
		"[ARTICLE]",
		"# Welcome",
		"**important**",
		"- First item",
		"- Second item",
		"- A | B",
		"[FOOTER]",
		"Copyright 2024",
	}

	for _, exp := range expected {
		if !strings.Contains(result.BodyMarkdown, exp) {
			t.Errorf("BodyMarkdown missing expected content: %s\nGot:\n%s", exp, result.BodyMarkdown)
		}
	}

	// Verify outer semantic labels do NOT appear (innermost only)
	shouldNotContain := []string{"[HEADER]", "[MAIN CONTENT]"}
	for _, s := range shouldNotContain {
		if strings.Contains(result.BodyMarkdown, s) {
			t.Errorf("BodyMarkdown should NOT contain outer semantic %s (innermost only)\nGot:\n%s", s, result.BodyMarkdown)
		}
	}
}

func TestMarkdownExtractionWithComparison(t *testing.T) {
	// Simulate JS-rendered page with additional dynamic content
	jsRenderedHTML := `<!DOCTYPE html>
<html>
<body>
    <main>
        <h1>Product Page</h1>
        <p>Base content visible to all.</p>
        <div class="js-loaded">
            <h2>Reviews</h2>
            <p>Customer reviews loaded by JavaScript.</p>
            <ul>
                <li>Great product! - User1</li>
                <li>Highly recommended - User2</li>
            </ul>
        </div>
    </main>
</body>
</html>`

	// Simulate non-JS page without dynamic content
	nonJSHTML := `<!DOCTYPE html>
<html>
<body>
    <main>
        <h1>Product Page</h1>
        <p>Base content visible to all.</p>
    </main>
</body>
</html>`

	parser := NewParser()

	jsResult, err := parser.ParseWithOptions(jsRenderedHTML, ParseOptions{
		PageURL: "https://example.com/product",
	})
	if err != nil {
		t.Fatalf("ParseWithOptions() error for JS HTML = %v", err)
	}

	nonJSResult, err := parser.ParseWithOptions(nonJSHTML, ParseOptions{
		PageURL: "https://example.com/product",
	})
	if err != nil {
		t.Fatalf("ParseWithOptions() error for non-JS HTML = %v", err)
	}

	// Both should have the base content
	baseContent := []string{
		"[MAIN CONTENT]",
		"# Product Page",
		"Base content visible to all",
	}

	for _, exp := range baseContent {
		if !strings.Contains(jsResult.BodyMarkdown, exp) {
			t.Errorf("JS BodyMarkdown missing base content: %s", exp)
		}
		if !strings.Contains(nonJSResult.BodyMarkdown, exp) {
			t.Errorf("Non-JS BodyMarkdown missing base content: %s", exp)
		}
	}

	// Only JS version should have the dynamic content
	jsOnlyContent := []string{
		"## Reviews",
		"Customer reviews loaded by JavaScript",
		"- Great product! - User1",
		"- Highly recommended - User2",
	}

	for _, exp := range jsOnlyContent {
		if !strings.Contains(jsResult.BodyMarkdown, exp) {
			t.Errorf("JS BodyMarkdown missing dynamic content: %s", exp)
		}
		if strings.Contains(nonJSResult.BodyMarkdown, exp) {
			t.Errorf("Non-JS BodyMarkdown should NOT contain dynamic content: %s", exp)
		}
	}

	// JS version should have more content
	if len(jsResult.BodyMarkdown) <= len(nonJSResult.BodyMarkdown) {
		t.Errorf("JS BodyMarkdown should be longer than non-JS BodyMarkdown. JS: %d, Non-JS: %d",
			len(jsResult.BodyMarkdown), len(nonJSResult.BodyMarkdown))
	}
}

func TestMarkdownBlockquoteAndDefinitionList(t *testing.T) {
	html := `<!DOCTYPE html>
<html>
<body>
    <blockquote>This is a famous quote that should be formatted properly.</blockquote>
    <dl>
        <dt>HTML</dt>
        <dd>HyperText Markup Language</dd>
        <dt>CSS</dt>
        <dd>Cascading Style Sheets</dd>
    </dl>
</body>
</html>`

	parser := NewParser()
	result, err := parser.ParseWithOptions(html, ParseOptions{
		PageURL: "https://example.com",
	})
	if err != nil {
		t.Fatalf("ParseWithOptions() error = %v", err)
	}

	// Check blockquote formatting
	if !strings.Contains(result.BodyMarkdown, "> This is a famous quote") {
		t.Error("BodyMarkdown should contain blockquote with > prefix")
	}

	// Check definition list formatting
	if !strings.Contains(result.BodyMarkdown, "**HTML**") {
		t.Error("BodyMarkdown should contain bold term for definition list")
	}
	if !strings.Contains(result.BodyMarkdown, "HyperText Markup Language") {
		t.Error("BodyMarkdown should contain definition")
	}
	if !strings.Contains(result.BodyMarkdown, "**CSS**") {
		t.Error("BodyMarkdown should contain bold CSS term")
	}
}
