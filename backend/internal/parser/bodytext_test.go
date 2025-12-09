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
