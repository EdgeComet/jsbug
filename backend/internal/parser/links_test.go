package parser

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestExtractLinks_Basic(t *testing.T) {
	html := `<html>
		<body>
			<a href="https://example.com/page1">Link 1</a>
			<a href="/page2">Link 2</a>
			<a href="page3">Link 3</a>
		</body>
	</html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	links := ExtractLinks(doc, "https://example.com")

	if len(links) != 3 {
		t.Fatalf("Expected 3 links, got %d", len(links))
	}

	// Check first link (absolute)
	if links[0].Href != "https://example.com/page1" {
		t.Errorf("Link[0].Href = %q, want %q", links[0].Href, "https://example.com/page1")
	}
	if links[0].Text != "Link 1" {
		t.Errorf("Link[0].Text = %q, want %q", links[0].Text, "Link 1")
	}
	if !links[0].IsAbsolute {
		t.Error("Link[0] should be absolute")
	}

	// Check second link (relative with /) - should be resolved to absolute
	if links[1].Href != "https://example.com/page2" {
		t.Errorf("Link[1].Href = %q, want %q", links[1].Href, "https://example.com/page2")
	}
	if links[1].IsAbsolute {
		t.Error("Link[1] should not be absolute")
	}
}

func TestExtractLinks_SkippedSchemes(t *testing.T) {
	html := `<html>
		<body>
			<a href="javascript:void(0)">JS Link</a>
			<a href="mailto:test@example.com">Email</a>
			<a href="tel:+1234567890">Phone</a>
			<a href="data:text/html,test">Data</a>
			<a href="https://example.com/valid">Valid</a>
			<a href="">Empty</a>
		</body>
	</html>`

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	links := ExtractLinks(doc, "https://example.com")

	// Only the valid HTTPS link should be extracted
	if len(links) != 1 {
		t.Fatalf("Expected 1 link (skipping js/mailto/tel/data/empty), got %d", len(links))
	}

	if links[0].Href != "https://example.com/valid" {
		t.Errorf("Expected valid link, got %q", links[0].Href)
	}
}

func TestExtractLinks_TextExtraction(t *testing.T) {
	html := `<html>
		<body>
			<a href="/a">  Multiple   Spaces  </a>
			<a href="/b">
				Newlines
				And
				Tabs
			</a>
			<a href="/c"><span>Nested <strong>Text</strong></span></a>
			<a href="/d"></a>
		</body>
	</html>`

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	links := ExtractLinks(doc, "https://example.com")

	if len(links) != 4 {
		t.Fatalf("Expected 4 links, got %d", len(links))
	}

	// Multiple spaces collapsed
	if links[0].Text != "Multiple Spaces" {
		t.Errorf("Link[0].Text = %q, want %q", links[0].Text, "Multiple Spaces")
	}

	// Newlines/tabs collapsed
	if links[1].Text != "Newlines And Tabs" {
		t.Errorf("Link[1].Text = %q, want %q", links[1].Text, "Newlines And Tabs")
	}

	// Nested elements
	if links[2].Text != "Nested Text" {
		t.Errorf("Link[2].Text = %q, want %q", links[2].Text, "Nested Text")
	}

	// Empty text
	if links[3].Text != "" {
		t.Errorf("Link[3].Text = %q, want empty", links[3].Text)
	}
}

func TestExtractLinks_RelAttributes(t *testing.T) {
	tests := []struct {
		name          string
		rel           string
		wantDofollow  bool
		wantUgc       bool
		wantSponsored bool
	}{
		{"no rel", "", true, false, false},
		{"nofollow", "nofollow", false, false, false},
		{"ugc", "ugc", true, true, false},
		{"sponsored", "sponsored", true, false, true},
		{"nofollow ugc", "nofollow ugc", false, true, false},
		{"nofollow sponsored", "nofollow sponsored", false, false, true},
		{"all three", "nofollow ugc sponsored", false, true, true},
		{"noopener noreferrer", "noopener noreferrer", true, false, false},
		{"uppercase NOFOLLOW", "NOFOLLOW", false, false, false},
		{"mixed case NoFollow", "NoFollow", false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			relAttr := ""
			if tt.rel != "" {
				relAttr = ` rel="` + tt.rel + `"`
			}
			html := `<html><body><a href="/link"` + relAttr + `>Link</a></body></html>`

			doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
			links := ExtractLinks(doc, "https://example.com")

			if len(links) != 1 {
				t.Fatalf("Expected 1 link, got %d", len(links))
			}

			if links[0].IsDofollow != tt.wantDofollow {
				t.Errorf("IsDofollow = %v, want %v", links[0].IsDofollow, tt.wantDofollow)
			}
			if links[0].IsUgc != tt.wantUgc {
				t.Errorf("IsUgc = %v, want %v", links[0].IsUgc, tt.wantUgc)
			}
			if links[0].IsSponsored != tt.wantSponsored {
				t.Errorf("IsSponsored = %v, want %v", links[0].IsSponsored, tt.wantSponsored)
			}
		})
	}
}

func TestExtractLinks_ImageLinks(t *testing.T) {
	html := `<html>
		<body>
			<a href="/text-link">Text Link</a>
			<a href="/image-link"><img src="/img.jpg" alt="Image"></a>
			<a href="/mixed-link"><span>Text</span><img src="/img.jpg"></a>
			<a href="/nested-image"><div><span><img src="/deep.jpg"></span></div></a>
		</body>
	</html>`

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	links := ExtractLinks(doc, "https://example.com")

	if len(links) != 4 {
		t.Fatalf("Expected 4 links, got %d", len(links))
	}

	// Text only - not image link
	if links[0].IsImageLink {
		t.Error("Link[0] should not be an image link")
	}

	// Direct image - is image link
	if !links[1].IsImageLink {
		t.Error("Link[1] should be an image link")
	}

	// Mixed text and image - is image link
	if !links[2].IsImageLink {
		t.Error("Link[2] should be an image link")
	}

	// Deeply nested image - is image link
	if !links[3].IsImageLink {
		t.Error("Link[3] should be an image link")
	}
}

func TestExtractLinks_ExternalDetection(t *testing.T) {
	tests := []struct {
		name         string
		href         string
		pageURL      string
		wantExternal bool
	}{
		{"same domain", "https://example.com/page", "https://example.com", false},
		{"subdomain is internal", "https://cdn.example.com/file", "https://example.com", false},
		{"www subdomain internal", "https://www.example.com/page", "https://example.com", false},
		{"different domain external", "https://other.com/page", "https://example.com", true},
		{"relative URL internal", "/page", "https://example.com", false},
		{"relative no slash internal", "page", "https://example.com", false},
		{"protocol-relative same domain", "//example.com/page", "https://example.com", false},
		{"protocol-relative external", "//other.com/page", "https://example.com", true},
		{"social external", "https://facebook.com/page", "https://example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html := `<html><body><a href="` + tt.href + `">Link</a></body></html>`

			doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
			links := ExtractLinks(doc, tt.pageURL)

			if len(links) != 1 {
				t.Fatalf("Expected 1 link, got %d", len(links))
			}

			if links[0].IsExternal != tt.wantExternal {
				t.Errorf("IsExternal = %v, want %v", links[0].IsExternal, tt.wantExternal)
			}
		})
	}
}

func TestExtractLinks_SocialDetection(t *testing.T) {
	tests := []struct {
		name       string
		href       string
		wantSocial bool
	}{
		{"facebook", "https://facebook.com/page", true},
		{"twitter", "https://twitter.com/user", true},
		{"x.com", "https://x.com/user", true},
		{"linkedin", "https://linkedin.com/in/user", true},
		{"instagram", "https://instagram.com/user", true},
		{"youtube", "https://youtube.com/watch", true},
		{"mobile facebook", "https://m.facebook.com/page", true},
		{"www twitter", "https://www.twitter.com/user", true},
		{"non-social external", "https://other.com/page", false},
		{"relative URL", "/page", false},
		{"internal link", "https://example.com/page", false},
		{"protocol-relative social", "//facebook.com/page", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html := `<html><body><a href="` + tt.href + `">Link</a></body></html>`

			doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
			links := ExtractLinks(doc, "https://example.com")

			if len(links) != 1 {
				t.Fatalf("Expected 1 link, got %d", len(links))
			}

			if links[0].IsSocial != tt.wantSocial {
				t.Errorf("IsSocial = %v, want %v", links[0].IsSocial, tt.wantSocial)
			}
		})
	}
}

func TestExtractLinks_SocialAndExternal(t *testing.T) {
	// Social links should be both external AND social
	html := `<html><body><a href="https://facebook.com/page">Facebook</a></body></html>`

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	links := ExtractLinks(doc, "https://example.com")

	if len(links) != 1 {
		t.Fatalf("Expected 1 link, got %d", len(links))
	}

	if !links[0].IsExternal {
		t.Error("Social link should be external")
	}
	if !links[0].IsSocial {
		t.Error("Social link should be marked as social")
	}
}

func TestExtractLinks_MixedPage(t *testing.T) {
	html := `<html>
		<body>
			<a href="https://example.com/internal">Internal</a>
			<a href="https://cdn.example.com/asset">CDN (Internal)</a>
			<a href="https://other.com/page">External</a>
			<a href="https://facebook.com/page" rel="nofollow">Facebook</a>
			<a href="/relative">Relative</a>
			<a href="https://twitter.com/user" rel="ugc sponsored">Twitter UGC</a>
			<a href="/image"><img src="/img.jpg"></a>
		</body>
	</html>`

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	links := ExtractLinks(doc, "https://example.com")

	if len(links) != 7 {
		t.Fatalf("Expected 7 links, got %d", len(links))
	}

	// Internal link
	if links[0].IsExternal || links[0].IsSocial {
		t.Error("First link should be internal and not social")
	}

	// CDN subdomain (internal)
	if links[1].IsExternal {
		t.Error("CDN link should be internal")
	}

	// External non-social
	if !links[2].IsExternal || links[2].IsSocial {
		t.Error("Third link should be external but not social")
	}

	// Facebook with nofollow
	if !links[3].IsExternal || !links[3].IsSocial || links[3].IsDofollow {
		t.Error("Facebook link should be external, social, and nofollow")
	}

	// Relative (internal)
	if links[4].IsExternal || links[4].IsAbsolute {
		t.Error("Relative link should be internal and not absolute")
	}

	// Twitter with UGC and sponsored
	if !links[5].IsSocial || !links[5].IsUgc || !links[5].IsSponsored {
		t.Error("Twitter link should be social, UGC, and sponsored")
	}

	// Image link
	if !links[6].IsImageLink {
		t.Error("Image link should be marked as image")
	}
}

func TestParseRelAttribute(t *testing.T) {
	tests := []struct {
		rel       string
		nofollow  bool
		ugc       bool
		sponsored bool
	}{
		{"", false, false, false},
		{"nofollow", true, false, false},
		{"ugc", false, true, false},
		{"sponsored", false, false, true},
		{"nofollow ugc", true, true, false},
		{"nofollow sponsored", true, false, true},
		{"ugc sponsored", false, true, true},
		{"nofollow ugc sponsored", true, true, true},
		{"noopener noreferrer", false, false, false},
		{"NOFOLLOW", true, false, false},
		{"UGC SPONSORED", false, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.rel, func(t *testing.T) {
			nofollow, ugc, sponsored := parseRelAttribute(tt.rel)
			if nofollow != tt.nofollow {
				t.Errorf("nofollow = %v, want %v", nofollow, tt.nofollow)
			}
			if ugc != tt.ugc {
				t.Errorf("ugc = %v, want %v", ugc, tt.ugc)
			}
			if sponsored != tt.sponsored {
				t.Errorf("sponsored = %v, want %v", sponsored, tt.sponsored)
			}
		})
	}
}

func TestExtractLinks_EmptyDocument(t *testing.T) {
	html := `<html><body></body></html>`

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	links := ExtractLinks(doc, "https://example.com")

	if len(links) != 0 {
		t.Errorf("Expected 0 links for empty document, got %d", len(links))
	}
}

func TestExtractLinks_FragmentOnly(t *testing.T) {
	html := `<html><body><a href="#section">Jump to section</a><a href="#">Top</a></body></html>`

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	links := ExtractLinks(doc, "https://example.com/page")

	// Fragment-only links (#section, #) should be skipped
	if len(links) != 0 {
		t.Errorf("Expected 0 links (fragment-only should be skipped), got %d", len(links))
	}
}
