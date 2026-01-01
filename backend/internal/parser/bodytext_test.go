package parser

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/user/jsbug/internal/types"
)

func TestExtractBodyText(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		contains []string // strings that should be in the result
		excludes []string // strings that should NOT be in the result
	}{
		{
			name: "normal HTML with mixed content",
			html: `<html>
				<head><title>Test</title></head>
				<body>
					<h1>Hello World</h1>
					<p>This is a paragraph.</p>
					<div>More content here.</div>
				</body>
			</html>`,
			contains: []string{"Hello World", "This is a paragraph", "More content here"},
			excludes: []string{"<h1>", "<p>", "<div>"},
		},
		{
			name: "HTML with heavy script/style tags",
			html: `<html>
				<head>
					<style>.foo { color: red; } .bar { display: none; }</style>
					<script>var x = 1; function test() { return x; }</script>
				</head>
				<body>
					<p>Visible text</p>
					<script>console.log("hidden");</script>
					<style>.hidden { opacity: 0; }</style>
				</body>
			</html>`,
			contains: []string{"Visible text"},
			excludes: []string{"color: red", "var x", "console.log", "opacity"},
		},
		{
			name:     "empty document",
			html:     `<html><head></head><body></body></html>`,
			contains: []string{},
			excludes: []string{},
		},
		{
			name: "document with only scripts",
			html: `<html>
				<head><script>var a = 1;</script></head>
				<body>
					<script>var b = 2;</script>
					<script>var c = 3;</script>
				</body>
			</html>`,
			contains: []string{},
			excludes: []string{"var a", "var b", "var c"},
		},
		{
			name: "HTML with noscript and iframe",
			html: `<html>
				<body>
					<p>Before</p>
					<noscript>Enable JavaScript</noscript>
					<iframe src="hidden.html">Frame content</iframe>
					<p>After</p>
				</body>
			</html>`,
			contains: []string{"Before", "After"},
			excludes: []string{"Enable JavaScript", "Frame content"},
		},
		{
			name: "HTML with SVG",
			html: `<html>
				<body>
					<p>Text before</p>
					<svg><text>SVG Text</text><path d="M0 0"/></svg>
					<p>Text after</p>
				</body>
			</html>`,
			contains: []string{"Text before", "Text after"},
			excludes: []string{"SVG Text", "M0 0"},
		},
		{
			name: "whitespace normalization",
			html: `<html>
				<body>
					<p>Multiple    spaces</p>
					<p>
						Newlines
						and
						tabs	here
					</p>
				</body>
			</html>`,
			contains: []string{"Multiple spaces", "Newlines and tabs here"},
			excludes: []string{},
		},
		{
			name:     "adjacent block elements without whitespace",
			html:     `<html><body><h1>Header</h1><p>Paragraph</p><div>Div</div></body></html>`,
			contains: []string{"Header Paragraph Div"},
			excludes: []string{"HeaderParagraph", "ParagraphDiv"},
		},
		{
			name:     "list items without whitespace",
			html:     `<ul><li>One</li><li>Two</li><li>Three</li></ul>`,
			contains: []string{"One Two Three"},
			excludes: []string{"OneTwo", "TwoThree"},
		},
		{
			name:     "inline elements with punctuation",
			html:     `<p>Click <a href="#">here</a>.</p>`,
			contains: []string{"Click", "here"},
			excludes: []string{},
		},
		{
			name:     "mixed inline and block",
			html:     `<h1>Title with <strong>bold</strong></h1><p>Next paragraph</p>`,
			contains: []string{"Title with bold Next paragraph"},
			excludes: []string{"boldNext"},
		},
		{
			name:     "adjacent inline elements need spacing",
			html:     `<nav><a href="#">Home</a><a href="#">About</a><a href="#">Contact</a></nav>`,
			contains: []string{"Home About Contact"},
			excludes: []string{"HomeAbout", "AboutContact"},
		},
		{
			name:     "navigation with adjacent links",
			html:     `<a href="#">IEEE.org</a><a href="#">IEEE Xplore</a><a href="#">IEEE Standards</a>`,
			contains: []string{"IEEE.org IEEE Xplore IEEE Standards"},
			excludes: []string{"IEEE.orgIEEE", "XploreIEEE"},
		},
		{
			name:     "text after inline element",
			html:     `<a>Spectrum</a>FOR THE INSIDER`,
			contains: []string{"Spectrum FOR THE INSIDER"},
			excludes: []string{"SpectrumFOR"},
		},
		{
			name:     "inline element after text",
			html:     `Text<span>More</span>`,
			contains: []string{"Text More"},
			excludes: []string{"TextMore"},
		},
		{
			name:     "mixed text and inline elements",
			html:     `<span>IEEE Spectrum</span>FOR THE TECHNOLOGY INSIDER`,
			contains: []string{"IEEE Spectrum FOR THE TECHNOLOGY INSIDER"},
			excludes: []string{"SpectrumFOR"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			result := ExtractBodyText(doc)

			for _, want := range tt.contains {
				if !strings.Contains(result, want) {
					t.Errorf("ExtractBodyText() result should contain %q, got %q", want, result)
				}
			}

			for _, exclude := range tt.excludes {
				if strings.Contains(result, exclude) {
					t.Errorf("ExtractBodyText() result should NOT contain %q, got %q", exclude, result)
				}
			}
		})
	}
}

func TestExtractBodyText_Truncation(t *testing.T) {
	// Create HTML with body text exceeding 3MB
	largeText := strings.Repeat("A", types.MaxBodyTextBytes+1000)
	html := "<html><body><p>" + largeText + "</p></body></html>"

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	result := ExtractBodyText(doc)

	if len(result) > types.MaxBodyTextBytes {
		t.Errorf("ExtractBodyText() result length = %d, want <= %d", len(result), types.MaxBodyTextBytes)
	}
}

func TestTruncateUTF8(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxBytes int
		wantLen  int // expected length (may be less than maxBytes due to UTF-8 boundaries)
	}{
		{
			name:     "ASCII under limit",
			input:    "Hello World",
			maxBytes: 100,
			wantLen:  11,
		},
		{
			name:     "ASCII at limit",
			input:    "Hello",
			maxBytes: 5,
			wantLen:  5,
		},
		{
			name:     "ASCII over limit",
			input:    "Hello World",
			maxBytes: 5,
			wantLen:  5,
		},
		{
			name:     "UTF-8 emoji under limit",
			input:    "Hello 😀 World",
			maxBytes: 100,
			wantLen:  16, // "Hello " (6) + emoji (4) + " World" (6) = 16
		},
		{
			name:     "UTF-8 emoji cut at boundary",
			input:    "Hi 😀",
			maxBytes: 5, // "Hi " = 3, emoji = 4 bytes, so should only get "Hi "
			wantLen:  3,
		},
		{
			name:     "CJK characters",
			input:    "Hello 世界",
			maxBytes: 10, // "Hello " = 6, each CJK char = 3 bytes
			wantLen:  9,  // "Hello " (6) + "世" (3) = 9
		},
		{
			name:     "empty string",
			input:    "",
			maxBytes: 10,
			wantLen:  0,
		},
		{
			name:     "zero maxBytes",
			input:    "Hello",
			maxBytes: 0,
			wantLen:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateUTF8(tt.input, tt.maxBytes)

			if len(result) != tt.wantLen {
				t.Errorf("truncateUTF8(%q, %d) length = %d, want %d", tt.input, tt.maxBytes, len(result), tt.wantLen)
			}

			// Verify result is valid UTF-8
			if !isValidUTF8(result) {
				t.Errorf("truncateUTF8(%q, %d) produced invalid UTF-8: %q", tt.input, tt.maxBytes, result)
			}
		})
	}
}

func TestTruncateUTF8_PreservesValidity(t *testing.T) {
	// Test with various multi-byte sequences
	inputs := []string{
		"Hello 🌍🌎🌏 World",      // 4-byte emojis
		"日本語テスト",               // 3-byte CJK
		"Ελληνικά",             // 2-byte Greek
		"Mix: café 日本 🎉",       // mixed
		"👨‍👩‍👧‍👦 family emoji", // complex emoji with ZWJ
	}

	for _, input := range inputs {
		for maxBytes := 1; maxBytes <= len(input)+5; maxBytes++ {
			result := truncateUTF8(input, maxBytes)
			if !isValidUTF8(result) {
				t.Errorf("truncateUTF8(%q, %d) produced invalid UTF-8", input, maxBytes)
			}
			if len(result) > maxBytes {
				t.Errorf("truncateUTF8(%q, %d) result too long: %d", input, maxBytes, len(result))
			}
		}
	}
}

func isValidUTF8(s string) bool {
	for i := 0; i < len(s); {
		r, size := decodeRune(s[i:])
		if r == 0xFFFD && size == 1 {
			return false
		}
		i += size
	}
	return true
}

func decodeRune(s string) (rune, int) {
	if len(s) == 0 {
		return 0, 0
	}
	// Simple UTF-8 decode
	b := s[0]
	if b < 0x80 {
		return rune(b), 1
	}
	if b < 0xC0 {
		return 0xFFFD, 1 // invalid start byte
	}
	if b < 0xE0 {
		if len(s) < 2 {
			return 0xFFFD, 1
		}
		return rune(b&0x1F)<<6 | rune(s[1]&0x3F), 2
	}
	if b < 0xF0 {
		if len(s) < 3 {
			return 0xFFFD, 1
		}
		return rune(b&0x0F)<<12 | rune(s[1]&0x3F)<<6 | rune(s[2]&0x3F), 3
	}
	if len(s) < 4 {
		return 0xFFFD, 1
	}
	return rune(b&0x07)<<18 | rune(s[1]&0x3F)<<12 | rune(s[2]&0x3F)<<6 | rune(s[3]&0x3F), 4
}

func TestCalculateTextHtmlRatio(t *testing.T) {
	tests := []struct {
		name     string
		bodyText string
		html     string
		want     float64
	}{
		{
			name:     "normal ratio",
			bodyText: "Hello World",        // 11 bytes
			html:     "<p>Hello World</p>", // 18 bytes
			want:     0.6111,               // 11/18 = 0.6111...
		},
		{
			name:     "empty html",
			bodyText: "Hello",
			html:     "",
			want:     0.0,
		},
		{
			name:     "empty text",
			bodyText: "",
			html:     "<p></p>",
			want:     0.0,
		},
		{
			name:     "both empty",
			bodyText: "",
			html:     "",
			want:     0.0,
		},
		{
			name:     "equal lengths",
			bodyText: "12345",
			html:     "12345",
			want:     1.0,
		},
		{
			name:     "text longer than html (edge case)",
			bodyText: "Hello World Extended",
			html:     "<p>Hi</p>",
			want:     2.2222, // 20/9
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateTextHtmlRatio(tt.bodyText, tt.html)
			// Allow small floating point difference
			if got < tt.want-0.0001 || got > tt.want+0.0001 {
				t.Errorf("CalculateTextHtmlRatio(%q, %q) = %v, want %v", tt.bodyText, tt.html, got, tt.want)
			}
		})
	}
}

func TestCalculateTextHtmlRatio_Precision(t *testing.T) {
	// Test that ratio is rounded to 4 decimal places
	bodyText := "A"
	html := "ABC"

	got := CalculateTextHtmlRatio(bodyText, html)
	expected := 0.3333 // 1/3 rounded to 4 decimals

	if got != expected {
		t.Errorf("CalculateTextHtmlRatio() = %v, want %v (4 decimal precision)", got, expected)
	}
}

func TestExtractBodyMarkdown_SemanticSections(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "nav section",
			html:     "<html><body><nav>Home | About | Contact</nav></body></html>",
			expected: "---\n[NAV]\nHome | About | Contact",
		},
		{
			name:     "multiple nav sections numbered",
			html:     "<html><body><nav>Main</nav><nav>Footer</nav></body></html>",
			expected: "---\n[NAV]\nMain\n\n---\n[NAV 2]\nFooter",
		},
		{
			name:     "nav with aria-label",
			html:     `<html><body><nav aria-label="Main Menu">Links</nav></body></html>`,
			expected: "---\n[NAV: Main Menu]\nLinks",
		},
		{
			name:     "main content section",
			html:     "<html><body><main><h1>Title</h1><p>Content</p></main></body></html>",
			expected: "---\n[MAIN CONTENT]\n\n# Title\n\nContent",
		},
		{
			name:     "ARIA role navigation",
			html:     `<html><body><div role="navigation">Nav content</div></body></html>`,
			expected: "---\n[NAV]\nNav content",
		},
		{
			name:     "empty section skipped",
			html:     "<html><body><nav></nav><p>Content</p></body></html>",
			expected: "Content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("failed to parse HTML: %v", err)
			}

			result := ExtractBodyMarkdown(doc)
			if result != tt.expected {
				t.Errorf("ExtractBodyMarkdown() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExtractBodyMarkdown_Tables(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name: "simple table",
			html: `<html><body><table>
				<tr><th>Name</th><th>Price</th></tr>
				<tr><td>Apple</td><td>$1</td></tr>
			</table></body></html>`,
			expected: "- Name | Price\n- Apple | $1",
		},
		{
			name: "table with empty cells skipped",
			html: `<html><body><table>
				<tr><td>A</td><td></td><td>B</td></tr>
			</table></body></html>`,
			expected: "- A | B",
		},
		{
			name: "table with formatting",
			html: `<html><body><table>
				<tr><td><strong>Bold</strong></td><td>Normal</td></tr>
			</table></body></html>`,
			expected: "- **Bold** | Normal",
		},
		{
			name: "table after paragraph",
			html: `<html><body><p>Data:</p><table>
				<tr><td>X</td><td>Y</td></tr>
			</table></body></html>`,
			expected: "Data:\n\n- X | Y",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("failed to parse HTML: %v", err)
			}

			result := ExtractBodyMarkdown(doc)
			if result != tt.expected {
				t.Errorf("ExtractBodyMarkdown() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExtractBodyMarkdown_Blockquotes(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "simple blockquote",
			html:     "<html><body><blockquote>Quoted text</blockquote></body></html>",
			expected: "> Quoted text",
		},
		{
			name:     "blockquote with paragraph",
			html:     "<html><body><p>Before</p><blockquote>Quote</blockquote><p>After</p></body></html>",
			expected: "Before\n\n> Quote\n\nAfter",
		},
		{
			name:     "blockquote with formatting",
			html:     "<html><body><blockquote>This is <strong>important</strong></blockquote></body></html>",
			expected: "> This is **important**",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("failed to parse HTML: %v", err)
			}

			result := ExtractBodyMarkdown(doc)
			if result != tt.expected {
				t.Errorf("ExtractBodyMarkdown() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExtractBodyMarkdown_Links(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "http link",
			html:     `<html><body><p>Visit <a href="http://example.com">Example</a></p></body></html>`,
			expected: "Visit [Example](http://example.com)",
		},
		{
			name:     "https link",
			html:     `<html><body><p>Visit <a href="https://example.com">Example</a></p></body></html>`,
			expected: "Visit [Example](https://example.com)",
		},
		{
			name:     "javascript link becomes plain text",
			html:     `<html><body><p>Click <a href="javascript:void(0)">here</a></p></body></html>`,
			expected: "Click here",
		},
		{
			name:     "mailto link becomes plain text",
			html:     `<html><body><p>Email <a href="mailto:test@example.com">us</a></p></body></html>`,
			expected: "Email us",
		},
		{
			name:     "empty href becomes plain text",
			html:     `<html><body><p>Click <a href="">here</a></p></body></html>`,
			expected: "Click here",
		},
		{
			name:     "no href becomes plain text",
			html:     `<html><body><p>Click <a>here</a></p></body></html>`,
			expected: "Click here",
		},
		{
			name:     "URL with parentheses escaped",
			html:     `<html><body><p><a href="https://example.com/page(1)">Link</a></p></body></html>`,
			expected: "[Link](https://example.com/page%281%29)",
		},
		{
			name:     "text with brackets escaped",
			html:     `<html><body><p><a href="https://example.com">[Click]</a></p></body></html>`,
			expected: "[\\[Click\\]](https://example.com)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("failed to parse HTML: %v", err)
			}

			result := ExtractBodyMarkdown(doc)
			if result != tt.expected {
				t.Errorf("ExtractBodyMarkdown() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExtractBodyMarkdown_InlineFormatting(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "bold text",
			html:     "<html><body><p>This is <strong>bold</strong> text</p></body></html>",
			expected: "This is **bold** text",
		},
		{
			name:     "italic text",
			html:     "<html><body><p>This is <em>italic</em> text</p></body></html>",
			expected: "This is *italic* text",
		},
		{
			name:     "bold with b tag",
			html:     "<html><body><p>This is <b>bold</b> text</p></body></html>",
			expected: "This is **bold** text",
		},
		{
			name:     "italic with i tag",
			html:     "<html><body><p>This is <i>italic</i> text</p></body></html>",
			expected: "This is *italic* text",
		},
		{
			name:     "nested bold and italic",
			html:     "<html><body><p>This is <strong><em>bold italic</em></strong> text</p></body></html>",
			expected: "This is ***bold italic*** text",
		},
		{
			name:     "redundant bold collapsed",
			html:     "<html><body><p><b><strong>text</strong></b></p></body></html>",
			expected: "**text**",
		},
		{
			name:     "mixed content",
			html:     "<html><body><p>Normal <strong>bold</strong> and <em>italic</em> mix</p></body></html>",
			expected: "Normal **bold** and *italic* mix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("failed to parse HTML: %v", err)
			}

			result := ExtractBodyMarkdown(doc)
			if result != tt.expected {
				t.Errorf("ExtractBodyMarkdown() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExtractBodyMarkdown_DefinitionLists(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name: "simple definition list",
			html: `<html><body><dl>
				<dt>HTML</dt>
				<dd>Markup language</dd>
			</dl></body></html>`,
			expected: "- **HTML**: Markup language",
		},
		{
			name: "multiple definitions",
			html: `<html><body><dl>
				<dt>Term</dt>
				<dd>Definition 1</dd>
				<dd>Definition 2</dd>
			</dl></body></html>`,
			expected: "- **Term**: Definition 1; Definition 2",
		},
		{
			name: "multiple terms",
			html: `<html><body><dl>
				<dt>First</dt>
				<dd>First def</dd>
				<dt>Second</dt>
				<dd>Second def</dd>
			</dl></body></html>`,
			expected: "- **First**: First def\n- **Second**: Second def",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("failed to parse HTML: %v", err)
			}

			result := ExtractBodyMarkdown(doc)
			if result != tt.expected {
				t.Errorf("ExtractBodyMarkdown() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExtractBodyMarkdown_Lists(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "unordered list",
			html:     "<html><body><ul><li>One</li><li>Two</li><li>Three</li></ul></body></html>",
			expected: "- One\n- Two\n- Three",
		},
		{
			name:     "ordered list",
			html:     "<html><body><ol><li>First</li><li>Second</li><li>Third</li></ol></body></html>",
			expected: "1. First\n2. Second\n3. Third",
		},
		{
			name:     "nested list flattened",
			html:     "<html><body><ul><li>Parent<ol><li>Child 1</li><li>Child 2</li></ol></li></ul></body></html>",
			expected: "- Parent\n1. Child 1\n2. Child 2",
		},
		{
			name:     "orphan li treated as bullet",
			html:     "<html><body><li>Orphan item</li></body></html>",
			expected: "- Orphan item",
		},
		{
			name:     "list after paragraph",
			html:     "<html><body><p>Intro text</p><ul><li>Item</li></ul></body></html>",
			expected: "Intro text\n\n- Item",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("failed to parse HTML: %v", err)
			}

			result := ExtractBodyMarkdown(doc)
			if result != tt.expected {
				t.Errorf("ExtractBodyMarkdown() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExtractBodyMarkdown_Paragraphs(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "single paragraph",
			html:     "<html><body><p>Hello world</p></body></html>",
			expected: "Hello world",
		},
		{
			name:     "multiple paragraphs",
			html:     "<html><body><p>First</p><p>Second</p></body></html>",
			expected: "First\n\nSecond",
		},
		{
			name:     "heading and paragraph",
			html:     "<html><body><h1>Title</h1><p>Content here.</p></body></html>",
			expected: "# Title\n\nContent here.",
		},
		{
			name:     "empty paragraph skipped",
			html:     "<html><body><p></p><p>Valid</p></body></html>",
			expected: "Valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("failed to parse HTML: %v", err)
			}

			result := ExtractBodyMarkdown(doc)
			if result != tt.expected {
				t.Errorf("ExtractBodyMarkdown() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExtractBodyMarkdown_Headings(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "single h1",
			html:     "<html><body><h1>Title</h1></body></html>",
			expected: "# Title",
		},
		{
			name:     "multiple heading levels",
			html:     "<html><body><h1>Title</h1><h2>Subtitle</h2><h3>Section</h3></body></html>",
			expected: "# Title\n\n## Subtitle\n\n### Section",
		},
		{
			name:     "all heading levels",
			html:     "<html><body><h1>H1</h1><h2>H2</h2><h3>H3</h3><h4>H4</h4><h5>H5</h5><h6>H6</h6></body></html>",
			expected: "# H1\n\n## H2\n\n### H3\n\n#### H4\n\n##### H5\n\n###### H6",
		},
		{
			name:     "empty heading skipped",
			html:     "<html><body><h1></h1><h2>Valid</h2></body></html>",
			expected: "## Valid",
		},
		{
			name:     "whitespace-only heading skipped",
			html:     "<html><body><h1>   </h1><h2>Valid</h2></body></html>",
			expected: "## Valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("failed to parse HTML: %v", err)
			}

			result := ExtractBodyMarkdown(doc)
			if result != tt.expected {
				t.Errorf("ExtractBodyMarkdown() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestTruncateMarkdownAtBlock(t *testing.T) {
	// truncationIndicator is "\n\n[Content truncated...]" = 24 bytes
	tests := []struct {
		name     string
		markdown string
		maxBytes int
		expected string
	}{
		{
			name:     "no truncation needed",
			markdown: "Short text", // 10 bytes
			maxBytes: 100,
			expected: "Short text",
		},
		{
			name:     "truncate at block boundary",
			markdown: "First paragraph\n\nSecond paragraph\n\nThird paragraph", // 51 bytes
			maxBytes: 45, // 51 > 45, targetLen = 45 - 25 = 20, find \n\n at position 15
			expected: "First paragraph\n\n[Content truncated...]",
		},
		{
			name:     "truncate without block boundary",
			markdown: "Single very long paragraph without any block breaks here", // 56 bytes, no \n\n
			maxBytes: 50, // 56 > 50, targetLen = 50 - 25 = 25, UTF-8 truncate
			expected: "Single very long paragraph\n\n[Content truncated...]",
		},
		{
			name:     "maxBytes smaller than indicator",
			markdown: "This text is longer than the max bytes value", // 44 bytes
			maxBytes: 10, // 44 > 10, targetLen = 10 - 25 = -15, returns just indicator
			expected: "\n\n[Content truncated...]",
		},
		{
			name:     "exact size no truncation",
			markdown: "Exact", // 5 bytes
			maxBytes: 5,
			expected: "Exact",
		},
		{
			name:     "truncate at second block boundary",
			markdown: "Para one\n\nPara two\n\nPara three long text here extra", // 51 bytes
			maxBytes: 50, // 51 > 50, targetLen = 50 - 25 = 25, find last \n\n at position 18
			expected: "Para one\n\nPara two\n\n[Content truncated...]",
		},
		{
			name:     "just over limit truncates with no block boundary",
			markdown: "Short text here more text", // 25 bytes
			maxBytes: 28,                           // 25 <= 28, no truncation needed
			expected: "Short text here more text",
		},
		{
			name:     "truncate long text without block boundary",
			markdown: "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", // 36 bytes, no \n\n
			maxBytes: 35,                                     // 36 > 35, targetLen = 35 - 25 = 10
			expected: "ABCDEFGHIJK\n\n[Content truncated...]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateMarkdownAtBlock(tt.markdown, tt.maxBytes)
			if result != tt.expected {
				t.Errorf("truncateMarkdownAtBlock() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExtractBodyMarkdown_Truncation(t *testing.T) {
	// Create HTML with body markdown exceeding MaxBodyMarkdownBytes
	// We need many paragraphs to generate a large markdown output
	var htmlBuilder strings.Builder
	htmlBuilder.WriteString("<html><body>")

	// Each paragraph generates markdown like "This is paragraph N.\n\n"
	// We need enough paragraphs to exceed MaxBodyMarkdownBytes
	paragraphCount := types.MaxBodyMarkdownBytes/25 + 100
	for i := 0; i < paragraphCount; i++ {
		htmlBuilder.WriteString("<p>This is paragraph content.</p>")
	}
	htmlBuilder.WriteString("</body></html>")

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlBuilder.String()))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	result := ExtractBodyMarkdown(doc)

	if len(result) > types.MaxBodyMarkdownBytes {
		t.Errorf("ExtractBodyMarkdown() result length = %d, want <= %d", len(result), types.MaxBodyMarkdownBytes)
	}

	// Verify it ends with truncation indicator
	if !strings.HasSuffix(result, "[Content truncated...]") {
		t.Errorf("ExtractBodyMarkdown() should end with truncation indicator, got %q", result[len(result)-50:])
	}
}

func TestExtractBodyMarkdown_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
		isNil    bool // use nil doc instead of parsing html
	}{
		{
			name:     "nil document",
			html:     "",
			expected: "",
			isNil:    true,
		},
		{
			name:     "no body element",
			html:     "<html><head></head></html>",
			expected: "",
		},
		{
			name:     "script and style removed",
			html:     "<html><body><script>alert('x')</script><p>Text</p><style>.x{}</style></body></html>",
			expected: "Text",
		},
		{
			name:     "deeply nested divs",
			html:     "<html><body><div><div><div><p>Deep</p></div></div></div></body></html>",
			expected: "Deep",
		},
		{
			name:     "whitespace normalization",
			html:     "<html><body><p>  Multiple   spaces   </p></body></html>",
			expected: "Multiple spaces",
		},
		{
			name:     "pre element content",
			html:     "<html><body><pre>code here</pre></body></html>",
			expected: "code here",
		},
		{
			name:     "code element content",
			html:     "<html><body><code>inline code</code></body></html>",
			expected: "inline code",
		},
		{
			name:     "noscript removed",
			html:     "<html><body><p>Before</p><noscript>Enable JS</noscript><p>After</p></body></html>",
			expected: "Before\n\nAfter",
		},
		{
			name:     "iframe removed",
			html:     "<html><body><p>Content</p><iframe src=\"x\">Frame</iframe></body></html>",
			expected: "Content",
		},
		{
			name:     "svg removed",
			html:     "<html><body><p>Text</p><svg><text>SVG</text></svg></body></html>",
			expected: "Text",
		},
		{
			name:     "empty body",
			html:     "<html><body></body></html>",
			expected: "",
		},
		{
			name:     "whitespace only body",
			html:     "<html><body>   \n\t   </body></html>",
			expected: "",
		},
		{
			name:     "comment nodes ignored",
			html:     "<html><body><!-- comment --><p>Text</p></body></html>",
			expected: "Text",
		},
		{
			name:     "br tags concatenate text",
			html:     "<html><body><p>Line1<br/>Line2</p></body></html>",
			expected: "Line1Line2",
		},
		{
			name:     "span without formatting",
			html:     "<html><body><p>Text <span>in span</span> here</p></body></html>",
			expected: "Text in span here",
		},
		{
			name:     "mixed case tags",
			html:     "<HTML><BODY><P>Text</P></BODY></HTML>",
			expected: "Text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var doc *goquery.Document
			var err error

			if tt.isNil {
				doc = nil
			} else {
				doc, err = goquery.NewDocumentFromReader(strings.NewReader(tt.html))
				if err != nil {
					t.Fatalf("failed to parse HTML: %v", err)
				}
			}

			result := ExtractBodyMarkdown(doc)
			if result != tt.expected {
				t.Errorf("ExtractBodyMarkdown() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExtractBodyMarkdown_ComplexDocument(t *testing.T) {
	html := `<html><body>
		<header><h1>Site Title</h1></header>
		<nav><a href="https://example.com">Home</a></nav>
		<main>
			<article>
				<h2>Article Title</h2>
				<p>First <strong>paragraph</strong>.</p>
				<ul>
					<li>Item one</li>
					<li>Item two</li>
				</ul>
			</article>
		</main>
		<footer>Copyright</footer>
	</body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("failed to parse HTML: %v", err)
	}

	result := ExtractBodyMarkdown(doc)

	// Check for key elements - per spec, only innermost semantic labels appear
	// main contains article, so only [ARTICLE] should be labeled (not [MAIN CONTENT])
	shouldContain := []struct {
		desc    string
		content string
	}{
		{"header section", "[HEADER]"},
		{"h1 heading", "# Site Title"},
		{"nav section", "[NAV]"},
		{"nav link text", "Home"},
		{"article section", "[ARTICLE]"},
		{"h2 heading", "## Article Title"},
		{"bold text", "**paragraph**"},
		{"list item 1", "- Item one"},
		{"list item 2", "- Item two"},
		{"footer section", "[FOOTER]"},
		{"footer content", "Copyright"},
	}

	for _, check := range shouldContain {
		if !strings.Contains(result, check.content) {
			t.Errorf("Expected %s to contain %q, got:\n%s", check.desc, check.content, result)
		}
	}

	// Per spec: main contains article, so [MAIN CONTENT] should NOT appear
	if strings.Contains(result, "[MAIN CONTENT]") {
		t.Errorf("Result should NOT contain [MAIN CONTENT] (outer semantic, article is innermost), got:\n%s", result)
	}
}

func TestExtractBodyMarkdown_MalformedHTML(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		contains []string // content that should be present
	}{
		{
			name:     "unclosed tags",
			html:     "<html><body><p>Paragraph<div>Div content</body></html>",
			contains: []string{"Paragraph", "Div content"},
		},
		{
			name:     "nested incorrectly",
			html:     "<html><body><p><strong>Bold<em>Italic</strong>Mixed</em></p></body></html>",
			contains: []string{"Bold", "Italic", "Mixed"},
		},
		{
			name:     "missing closing html/body",
			html:     "<html><body><p>Content",
			contains: []string{"Content"},
		},
		{
			name:     "extra closing tags",
			html:     "<html><body><p>Text</p></p></div></body></html>",
			contains: []string{"Text"},
		},
		{
			name:     "attribute without quotes",
			html:     `<html><body><a href=https://example.com>Link</a></body></html>`,
			contains: []string{"Link"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("failed to parse HTML: %v", err)
			}

			result := ExtractBodyMarkdown(doc)

			for _, want := range tt.contains {
				if !strings.Contains(result, want) {
					t.Errorf("ExtractBodyMarkdown() should contain %q, got %q", want, result)
				}
			}
		})
	}
}

func TestExtractBodyMarkdown_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "HTML entities",
			html:     "<html><body><p>&amp; &lt; &gt; &quot;</p></body></html>",
			expected: `& < > "`,
		},
		{
			name:     "numeric entities",
			html:     "<html><body><p>&#169; &#8364;</p></body></html>",
			expected: "\u00A9 \u20AC", // copyright and euro symbols
		},
		{
			name:     "unicode text",
			html:     "<html><body><p>Hello \u4e16\u754c</p></body></html>",
			expected: "Hello \u4e16\u754c", // Hello World in Chinese
		},
		{
			name:     "emoji",
			html:     "<html><body><p>Hello \U0001F600</p></body></html>",
			expected: "Hello \U0001F600", // grinning face emoji
		},
		{
			name:     "newlines in paragraph",
			html:     "<html><body><p>Line1\nLine2\nLine3</p></body></html>",
			expected: "Line1 Line2 Line3",
		},
		{
			name:     "tabs in text",
			html:     "<html><body><p>Col1\tCol2\tCol3</p></body></html>",
			expected: "Col1 Col2 Col3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("failed to parse HTML: %v", err)
			}

			result := ExtractBodyMarkdown(doc)
			if result != tt.expected {
				t.Errorf("ExtractBodyMarkdown() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExtractBodyMarkdown_NestedSemanticElements(t *testing.T) {
	html := `<html><body>
		<main>
			<article>
				<header><h2>Article Header</h2></header>
				<section>
					<h3>Section Title</h3>
					<p>Section content.</p>
				</section>
				<aside>Related links</aside>
				<footer>Article footer</footer>
			</article>
		</main>
	</body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("failed to parse HTML: %v", err)
	}

	result := ExtractBodyMarkdown(doc)

	// Per spec: nested semantic elements should use innermost label only
	// MAIN and ARTICLE have semantic descendants, so should NOT be labeled
	shouldNotContain := []string{
		"[MAIN CONTENT]",
		"[ARTICLE]",
	}

	for _, check := range shouldNotContain {
		if strings.Contains(result, check) {
			t.Errorf("Result should NOT contain outer semantic label %q (innermost only), got:\n%s", check, result)
		}
	}

	// Only innermost semantic elements should be labeled
	shouldContain := []string{
		"[HEADER]",
		"[SECTION]",
		"[ASIDE]",
		"[FOOTER]",
		"## Article Header",
		"### Section Title",
		"Section content",
		"Related links",
		"Article footer",
	}

	for _, check := range shouldContain {
		if !strings.Contains(result, check) {
			t.Errorf("Expected result to contain %q, got:\n%s", check, result)
		}
	}
}

func TestExtractBodyMarkdown_NestedSemanticElements_SpecExample(t *testing.T) {
	// Exact example from markdown.md spec (lines 172-181)
	html := `<html><body><article><section><aside>Sidebar</aside></section></article></body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("failed to parse HTML: %v", err)
	}

	result := ExtractBodyMarkdown(doc)

	// Should contain only the innermost label
	if !strings.Contains(result, "[ASIDE]") {
		t.Errorf("Expected result to contain [ASIDE], got:\n%s", result)
	}
	if !strings.Contains(result, "Sidebar") {
		t.Errorf("Expected result to contain 'Sidebar', got:\n%s", result)
	}

	// Should NOT contain outer semantic labels
	if strings.Contains(result, "[ARTICLE]") {
		t.Errorf("Result should NOT contain [ARTICLE] (outer element), got:\n%s", result)
	}
	if strings.Contains(result, "[SECTION]") {
		t.Errorf("Result should NOT contain [SECTION] (outer element), got:\n%s", result)
	}
}

func TestExtractBodyMarkdown_MultipleSemanticElementsNumbered(t *testing.T) {
	html := `<html><body>
		<nav>First nav</nav>
		<nav>Second nav</nav>
		<nav>Third nav</nav>
		<section>First section</section>
		<section>Second section</section>
	</body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("failed to parse HTML: %v", err)
	}

	result := ExtractBodyMarkdown(doc)

	// First occurrences should not be numbered
	if !strings.Contains(result, "[NAV]") {
		t.Errorf("Expected first NAV without number, got:\n%s", result)
	}

	// Second and third should be numbered
	if !strings.Contains(result, "[NAV 2]") {
		t.Errorf("Expected second NAV with number 2, got:\n%s", result)
	}
	if !strings.Contains(result, "[NAV 3]") {
		t.Errorf("Expected third NAV with number 3, got:\n%s", result)
	}

	// Same for sections
	if !strings.Contains(result, "[SECTION]") {
		t.Errorf("Expected first SECTION without number, got:\n%s", result)
	}
	if !strings.Contains(result, "[SECTION 2]") {
		t.Errorf("Expected second SECTION with number 2, got:\n%s", result)
	}
}
