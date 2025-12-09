package parser

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestExtractImages_Basic(t *testing.T) {
	html := `<html>
		<body>
			<img src="https://example.com/image1.jpg" alt="Image 1">
			<img src="/image2.png" alt="Image 2">
			<img src="image3.gif">
		</body>
	</html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	images := ExtractImages(doc, "https://example.com")

	if len(images) != 3 {
		t.Fatalf("Expected 3 images, got %d", len(images))
	}

	// Check first image (absolute)
	if images[0].Src != "https://example.com/image1.jpg" {
		t.Errorf("Image[0].Src = %q, want %q", images[0].Src, "https://example.com/image1.jpg")
	}
	if images[0].Alt != "Image 1" {
		t.Errorf("Image[0].Alt = %q, want %q", images[0].Alt, "Image 1")
	}
	if !images[0].IsAbsolute {
		t.Error("Image[0] should be absolute")
	}

	// Check second image (relative with / - now resolved to absolute)
	if images[1].Src != "https://example.com/image2.png" {
		t.Errorf("Image[1].Src = %q, want %q", images[1].Src, "https://example.com/image2.png")
	}
	if images[1].IsAbsolute {
		t.Error("Image[1].IsAbsolute should be false (original was relative)")
	}

	// Check third image (relative - now resolved to absolute, no alt)
	if images[2].Src != "https://example.com/image3.gif" {
		t.Errorf("Image[2].Src = %q, want %q", images[2].Src, "https://example.com/image3.gif")
	}
	if images[2].Alt != "" {
		t.Errorf("Image[2].Alt = %q, want empty", images[2].Alt)
	}
}

func TestExtractImages_SkipDataURIs(t *testing.T) {
	html := `<html>
		<body>
			<img src="data:image/png;base64,iVBORw0KGgo=" alt="Data URI">
			<img src="DATA:image/gif;base64,R0lGODlh" alt="Uppercase Data">
			<img src="https://example.com/valid.jpg" alt="Valid">
		</body>
	</html>`

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	images := ExtractImages(doc, "https://example.com")

	// Only the valid image should be extracted
	if len(images) != 1 {
		t.Fatalf("Expected 1 image (skipping data URIs), got %d", len(images))
	}

	if images[0].Src != "https://example.com/valid.jpg" {
		t.Errorf("Expected valid image, got %q", images[0].Src)
	}
}

func TestExtractImages_SkipEmpty(t *testing.T) {
	html := `<html>
		<body>
			<img src="" alt="Empty src">
			<img alt="No src attribute">
			<img src="https://example.com/valid.jpg" alt="Valid">
		</body>
	</html>`

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	images := ExtractImages(doc, "https://example.com")

	// Only the valid image should be extracted
	if len(images) != 1 {
		t.Fatalf("Expected 1 image (skipping empty/missing src), got %d", len(images))
	}

	if images[0].Src != "https://example.com/valid.jpg" {
		t.Errorf("Expected valid image, got %q", images[0].Src)
	}
}

func TestExtractImages_ExternalDetection(t *testing.T) {
	tests := []struct {
		name         string
		src          string
		pageURL      string
		wantExternal bool
	}{
		{"same domain", "https://example.com/image.jpg", "https://example.com", false},
		{"subdomain is internal", "https://cdn.example.com/image.jpg", "https://example.com", false},
		{"www subdomain internal", "https://www.example.com/image.jpg", "https://example.com", false},
		{"different domain external", "https://other.com/image.jpg", "https://example.com", true},
		{"relative URL internal", "/image.jpg", "https://example.com", false},
		{"relative no slash internal", "image.jpg", "https://example.com", false},
		{"protocol-relative same domain", "//example.com/image.jpg", "https://example.com", false},
		{"protocol-relative external", "//other.com/image.jpg", "https://example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html := `<html><body><img src="` + tt.src + `" alt="Test"></body></html>`

			doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
			images := ExtractImages(doc, tt.pageURL)

			if len(images) != 1 {
				t.Fatalf("Expected 1 image, got %d", len(images))
			}

			if images[0].IsExternal != tt.wantExternal {
				t.Errorf("IsExternal = %v, want %v", images[0].IsExternal, tt.wantExternal)
			}
		})
	}
}

func TestExtractImages_AbsoluteDetection(t *testing.T) {
	tests := []struct {
		name         string
		src          string
		wantAbsolute bool
	}{
		{"http absolute", "http://example.com/image.jpg", true},
		{"https absolute", "https://example.com/image.jpg", true},
		{"protocol-relative", "//example.com/image.jpg", false},
		{"root-relative", "/images/photo.jpg", false},
		{"relative", "images/photo.jpg", false},
		{"relative with dot", "./images/photo.jpg", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html := `<html><body><img src="` + tt.src + `" alt="Test"></body></html>`

			doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
			images := ExtractImages(doc, "https://example.com")

			if len(images) != 1 {
				t.Fatalf("Expected 1 image, got %d", len(images))
			}

			if images[0].IsAbsolute != tt.wantAbsolute {
				t.Errorf("IsAbsolute = %v, want %v", images[0].IsAbsolute, tt.wantAbsolute)
			}
		})
	}
}

func TestExtractImages_InLink(t *testing.T) {
	html := `<html>
		<body>
			<img src="/standalone.jpg" alt="Standalone">
			<a href="/page1"><img src="/linked1.jpg" alt="Linked 1"></a>
			<a href="https://example.com/page2"><img src="/linked2.jpg" alt="Linked 2"></a>
			<div><a href="/page3"><span><img src="/deeply-nested.jpg" alt="Deep"></span></a></div>
		</body>
	</html>`

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	images := ExtractImages(doc, "https://example.com")

	if len(images) != 4 {
		t.Fatalf("Expected 4 images, got %d", len(images))
	}

	// Standalone image - not in link (src resolved)
	if images[0].IsInLink {
		t.Error("Image[0] should not be in link")
	}
	if images[0].Src != "https://example.com/standalone.jpg" {
		t.Errorf("Image[0].Src = %q, want %q", images[0].Src, "https://example.com/standalone.jpg")
	}
	if images[0].LinkHref != "" {
		t.Errorf("Image[0].LinkHref = %q, want empty", images[0].LinkHref)
	}

	// Direct child of anchor (src resolved)
	if !images[1].IsInLink {
		t.Error("Image[1] should be in link")
	}
	if images[1].Src != "https://example.com/linked1.jpg" {
		t.Errorf("Image[1].Src = %q, want %q", images[1].Src, "https://example.com/linked1.jpg")
	}
	if images[1].LinkHref != "/page1" {
		t.Errorf("Image[1].LinkHref = %q, want %q", images[1].LinkHref, "/page1")
	}

	// Direct child of anchor (absolute href, src resolved)
	if !images[2].IsInLink {
		t.Error("Image[2] should be in link")
	}
	if images[2].Src != "https://example.com/linked2.jpg" {
		t.Errorf("Image[2].Src = %q, want %q", images[2].Src, "https://example.com/linked2.jpg")
	}
	if images[2].LinkHref != "https://example.com/page2" {
		t.Errorf("Image[2].LinkHref = %q, want %q", images[2].LinkHref, "https://example.com/page2")
	}

	// Deeply nested in anchor (src resolved)
	if !images[3].IsInLink {
		t.Error("Image[3] should be in link")
	}
	if images[3].Src != "https://example.com/deeply-nested.jpg" {
		t.Errorf("Image[3].Src = %q, want %q", images[3].Src, "https://example.com/deeply-nested.jpg")
	}
	if images[3].LinkHref != "/page3" {
		t.Errorf("Image[3].LinkHref = %q, want %q", images[3].LinkHref, "/page3")
	}
}

func TestExtractImages_EmptyDocument(t *testing.T) {
	html := `<html><body></body></html>`

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	images := ExtractImages(doc, "https://example.com")

	if len(images) != 0 {
		t.Errorf("Expected 0 images for empty document, got %d", len(images))
	}
}

func TestExtractImages_MixedPage(t *testing.T) {
	html := `<html>
		<body>
			<img src="https://example.com/internal.jpg" alt="Internal">
			<img src="https://cdn.example.com/asset.jpg" alt="CDN (Internal)">
			<img src="https://other.com/external.jpg" alt="External">
			<a href="/gallery"><img src="/gallery-thumb.jpg" alt="Thumbnail"></a>
			<img src="data:image/png;base64,abc" alt="Should be skipped">
			<img src="" alt="Should be skipped too">
		</body>
	</html>`

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	images := ExtractImages(doc, "https://example.com")

	if len(images) != 4 {
		t.Fatalf("Expected 4 images (skipping data URI and empty), got %d", len(images))
	}

	// Internal image
	if images[0].IsExternal {
		t.Error("First image should be internal")
	}
	if !images[0].IsAbsolute {
		t.Error("First image should be absolute")
	}

	// CDN subdomain (internal)
	if images[1].IsExternal {
		t.Error("CDN image should be internal")
	}

	// External image
	if !images[2].IsExternal {
		t.Error("Third image should be external")
	}

	// Linked image (src resolved to absolute)
	if !images[3].IsInLink {
		t.Error("Gallery image should be in link")
	}
	if images[3].Src != "https://example.com/gallery-thumb.jpg" {
		t.Errorf("Gallery image Src = %q, want %q", images[3].Src, "https://example.com/gallery-thumb.jpg")
	}
	if images[3].LinkHref != "/gallery" {
		t.Errorf("Gallery image LinkHref = %q, want %q", images[3].LinkHref, "/gallery")
	}
}

func TestExtractImages_URLResolution(t *testing.T) {
	tests := []struct {
		name         string
		src          string
		pageURL      string
		wantSrc      string
		wantAbsolute bool
	}{
		{
			name:         "absolute URL unchanged",
			src:          "https://example.com/image.jpg",
			pageURL:      "https://example.com/page",
			wantSrc:      "https://example.com/image.jpg",
			wantAbsolute: true,
		},
		{
			name:         "root-relative resolved",
			src:          "/images/photo.jpg",
			pageURL:      "https://example.com/page/subpage",
			wantSrc:      "https://example.com/images/photo.jpg",
			wantAbsolute: false,
		},
		{
			name:         "relative resolved",
			src:          "photo.jpg",
			pageURL:      "https://example.com/page/",
			wantSrc:      "https://example.com/page/photo.jpg",
			wantAbsolute: false,
		},
		{
			name:         "relative with dot-dot resolved",
			src:          "../images/photo.jpg",
			pageURL:      "https://example.com/page/subpage/",
			wantSrc:      "https://example.com/page/images/photo.jpg",
			wantAbsolute: false,
		},
		{
			name:         "protocol-relative resolved",
			src:          "//cdn.example.com/image.jpg",
			pageURL:      "https://example.com/page",
			wantSrc:      "https://cdn.example.com/image.jpg",
			wantAbsolute: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html := `<html><body><img src="` + tt.src + `" alt="Test"></body></html>`

			doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
			images := ExtractImages(doc, tt.pageURL)

			if len(images) != 1 {
				t.Fatalf("Expected 1 image, got %d", len(images))
			}

			if images[0].Src != tt.wantSrc {
				t.Errorf("Src = %q, want %q", images[0].Src, tt.wantSrc)
			}
			if images[0].IsAbsolute != tt.wantAbsolute {
				t.Errorf("IsAbsolute = %v, want %v", images[0].IsAbsolute, tt.wantAbsolute)
			}
		})
	}
}
