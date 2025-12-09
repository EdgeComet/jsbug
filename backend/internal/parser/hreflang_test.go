package parser

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/user/jsbug/internal/types"
)

func TestExtractHrefLangs_HTML(t *testing.T) {
	tests := []struct {
		name    string
		html    string
		pageURL string
		want    []types.HrefLang
	}{
		{
			name: "multiple hreflang links",
			html: `<html>
				<head>
					<link rel="alternate" hreflang="en" href="https://example.com/en">
					<link rel="alternate" hreflang="fr" href="https://example.com/fr">
					<link rel="alternate" hreflang="de" href="https://example.com/de">
				</head>
				<body></body>
			</html>`,
			pageURL: "https://example.com",
			want: []types.HrefLang{
				{Lang: "en", URL: "https://example.com/en", Source: "link"},
				{Lang: "fr", URL: "https://example.com/fr", Source: "link"},
				{Lang: "de", URL: "https://example.com/de", Source: "link"},
			},
		},
		{
			name: "with x-default",
			html: `<html>
				<head>
					<link rel="alternate" hreflang="en" href="https://example.com/en">
					<link rel="alternate" hreflang="x-default" href="https://example.com/">
				</head>
				<body></body>
			</html>`,
			pageURL: "https://example.com",
			want: []types.HrefLang{
				{Lang: "en", URL: "https://example.com/en", Source: "link"},
				{Lang: "x-default", URL: "https://example.com/", Source: "link"},
			},
		},
		{
			name: "relative URLs resolved",
			html: `<html>
				<head>
					<link rel="alternate" hreflang="en" href="/en/page">
					<link rel="alternate" hreflang="fr" href="../fr/page">
				</head>
				<body></body>
			</html>`,
			pageURL: "https://example.com/section/page",
			want: []types.HrefLang{
				{Lang: "en", URL: "https://example.com/en/page", Source: "link"},
				{Lang: "fr", URL: "https://example.com/fr/page", Source: "link"},
			},
		},
		{
			name: "no hreflang links",
			html: `<html>
				<head>
					<link rel="stylesheet" href="/style.css">
					<link rel="canonical" href="https://example.com/page">
				</head>
				<body></body>
			</html>`,
			pageURL: "https://example.com",
			want:    []types.HrefLang{},
		},
		{
			name: "missing href ignored",
			html: `<html>
				<head>
					<link rel="alternate" hreflang="en">
					<link rel="alternate" hreflang="fr" href="https://example.com/fr">
				</head>
				<body></body>
			</html>`,
			pageURL: "https://example.com",
			want: []types.HrefLang{
				{Lang: "fr", URL: "https://example.com/fr", Source: "link"},
			},
		},
		{
			name: "empty hreflang ignored",
			html: `<html>
				<head>
					<link rel="alternate" hreflang="" href="https://example.com/empty">
					<link rel="alternate" hreflang="en" href="https://example.com/en">
				</head>
				<body></body>
			</html>`,
			pageURL: "https://example.com",
			want: []types.HrefLang{
				{Lang: "en", URL: "https://example.com/en", Source: "link"},
			},
		},
		{
			name: "rel attribute case-insensitive",
			html: `<html>
				<head>
					<link rel="Alternate" hreflang="en" href="https://example.com/en">
					<link rel="ALTERNATE" hreflang="fr" href="https://example.com/fr">
					<link rel="AlTeRnAtE" hreflang="de" href="https://example.com/de">
				</head>
				<body></body>
			</html>`,
			pageURL: "https://example.com",
			want: []types.HrefLang{
				{Lang: "en", URL: "https://example.com/en", Source: "link"},
				{Lang: "fr", URL: "https://example.com/fr", Source: "link"},
				{Lang: "de", URL: "https://example.com/de", Source: "link"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			got := ExtractHrefLangs(doc, tt.pageURL, "")

			if len(got) != len(tt.want) {
				t.Fatalf("ExtractHrefLangs() returned %d items, want %d", len(got), len(tt.want))
			}

			for i, want := range tt.want {
				if got[i].Lang != want.Lang {
					t.Errorf("HrefLang[%d].Lang = %q, want %q", i, got[i].Lang, want.Lang)
				}
				if got[i].URL != want.URL {
					t.Errorf("HrefLang[%d].URL = %q, want %q", i, got[i].URL, want.URL)
				}
				if got[i].Source != want.Source {
					t.Errorf("HrefLang[%d].Source = %q, want %q", i, got[i].Source, want.Source)
				}
			}
		})
	}
}

func TestExtractHrefLangs_LinkHeader(t *testing.T) {
	tests := []struct {
		name       string
		linkHeader string
		pageURL    string
		want       []types.HrefLang
	}{
		{
			name:       "single hreflang in header",
			linkHeader: `<https://example.com/en>; rel="alternate"; hreflang="en"`,
			pageURL:    "https://example.com",
			want: []types.HrefLang{
				{Lang: "en", URL: "https://example.com/en", Source: "header"},
			},
		},
		{
			name:       "multiple hreflangs in header",
			linkHeader: `<https://example.com/en>; rel="alternate"; hreflang="en", <https://example.com/fr>; rel="alternate"; hreflang="fr"`,
			pageURL:    "https://example.com",
			want: []types.HrefLang{
				{Lang: "en", URL: "https://example.com/en", Source: "header"},
				{Lang: "fr", URL: "https://example.com/fr", Source: "header"},
			},
		},
		{
			name:       "header without quotes",
			linkHeader: `<https://example.com/en>; rel=alternate; hreflang=en`,
			pageURL:    "https://example.com",
			want: []types.HrefLang{
				{Lang: "en", URL: "https://example.com/en", Source: "header"},
			},
		},
		{
			name:       "non-alternate links ignored",
			linkHeader: `<https://example.com/style.css>; rel="stylesheet", <https://example.com/en>; rel="alternate"; hreflang="en"`,
			pageURL:    "https://example.com",
			want: []types.HrefLang{
				{Lang: "en", URL: "https://example.com/en", Source: "header"},
			},
		},
		{
			name:       "empty header",
			linkHeader: "",
			pageURL:    "https://example.com",
			want:       nil,
		},
		{
			name:       "malformed entry skipped",
			linkHeader: `malformed, <https://example.com/en>; rel="alternate"; hreflang="en"`,
			pageURL:    "https://example.com",
			want: []types.HrefLang{
				{Lang: "en", URL: "https://example.com/en", Source: "header"},
			},
		},
		{
			name:       "URL with comma in query param",
			linkHeader: `<https://example.com/page?a=1,2>; rel="alternate"; hreflang="en"`,
			pageURL:    "https://example.com",
			want: []types.HrefLang{
				{Lang: "en", URL: "https://example.com/page?a=1,2", Source: "header"},
			},
		},
		{
			name:       "duplicate URLs with different hreflangs",
			linkHeader: `<https://example.com/page>; rel="alternate"; hreflang="en", <https://example.com/page>; rel="alternate"; hreflang="en-US"`,
			pageURL:    "https://example.com",
			want: []types.HrefLang{
				{Lang: "en", URL: "https://example.com/page", Source: "header"},
				{Lang: "en-US", URL: "https://example.com/page", Source: "header"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, _ := goquery.NewDocumentFromReader(strings.NewReader("<html></html>"))
			got := ExtractHrefLangs(doc, tt.pageURL, tt.linkHeader)

			if len(got) != len(tt.want) {
				t.Fatalf("ExtractHrefLangs() returned %d items, want %d", len(got), len(tt.want))
			}

			for i, want := range tt.want {
				if got[i].Lang != want.Lang {
					t.Errorf("HrefLang[%d].Lang = %q, want %q", i, got[i].Lang, want.Lang)
				}
				if got[i].URL != want.URL {
					t.Errorf("HrefLang[%d].URL = %q, want %q", i, got[i].URL, want.URL)
				}
				if got[i].Source != want.Source {
					t.Errorf("HrefLang[%d].Source = %q, want %q", i, got[i].Source, want.Source)
				}
			}
		})
	}
}

func TestExtractHrefLangs_Deduplication(t *testing.T) {
	// Same lang+url from both HTML and header - HTML should win
	html := `<html>
		<head>
			<link rel="alternate" hreflang="en" href="https://example.com/en">
		</head>
		<body></body>
	</html>`
	linkHeader := `<https://example.com/en>; rel="alternate"; hreflang="en"`
	pageURL := "https://example.com"

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	got := ExtractHrefLangs(doc, pageURL, linkHeader)

	// Should only have one entry (deduplicated)
	if len(got) != 1 {
		t.Fatalf("Expected 1 deduplicated entry, got %d", len(got))
	}

	// HTML source should win
	if got[0].Source != "link" {
		t.Errorf("Expected Source='link' (HTML priority), got %q", got[0].Source)
	}
}

func TestExtractHrefLangs_MixedSources(t *testing.T) {
	// Different langs from HTML and header
	html := `<html>
		<head>
			<link rel="alternate" hreflang="en" href="https://example.com/en">
			<link rel="alternate" hreflang="fr" href="https://example.com/fr">
		</head>
		<body></body>
	</html>`
	linkHeader := `<https://example.com/de>; rel="alternate"; hreflang="de", <https://example.com/es>; rel="alternate"; hreflang="es"`
	pageURL := "https://example.com"

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	got := ExtractHrefLangs(doc, pageURL, linkHeader)

	// Should have 4 entries
	if len(got) != 4 {
		t.Fatalf("Expected 4 entries, got %d", len(got))
	}

	// First two should be from HTML (link), last two from header
	for i := 0; i < 2; i++ {
		if got[i].Source != "link" {
			t.Errorf("Entry %d should be from 'link', got %q", i, got[i].Source)
		}
	}
	for i := 2; i < 4; i++ {
		if got[i].Source != "header" {
			t.Errorf("Entry %d should be from 'header', got %q", i, got[i].Source)
		}
	}
}

func TestResolveURL(t *testing.T) {
	tests := []struct {
		name    string
		href    string
		baseURL string
		want    string
	}{
		{
			name:    "already absolute",
			href:    "https://other.com/page",
			baseURL: "https://example.com/",
			want:    "https://other.com/page",
		},
		{
			name:    "relative path",
			href:    "/en/page",
			baseURL: "https://example.com/section/",
			want:    "https://example.com/en/page",
		},
		{
			name:    "relative with ../",
			href:    "../fr/page",
			baseURL: "https://example.com/en/section/",
			want:    "https://example.com/en/fr/page",
		},
		{
			name:    "protocol-relative",
			href:    "//cdn.example.com/file.js",
			baseURL: "https://example.com/page",
			want:    "https://cdn.example.com/file.js",
		},
		{
			name:    "empty href",
			href:    "",
			baseURL: "https://example.com/",
			want:    "",
		},
		{
			name:    "fragment only",
			href:    "#section",
			baseURL: "https://example.com/page",
			want:    "https://example.com/page#section",
		},
		{
			name:    "query string",
			href:    "?lang=en",
			baseURL: "https://example.com/page",
			want:    "https://example.com/page?lang=en",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveURL(tt.href, tt.baseURL)
			if got != tt.want {
				t.Errorf("resolveURL(%q, %q) = %q, want %q", tt.href, tt.baseURL, got, tt.want)
			}
		})
	}
}

func TestExtractHrefLangs_EmptyInputs(t *testing.T) {
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader("<html></html>"))

	// Empty HTML and empty header
	got := ExtractHrefLangs(doc, "https://example.com", "")

	if len(got) != 0 {
		t.Errorf("Expected empty result for empty inputs, got %d items", len(got))
	}
}
