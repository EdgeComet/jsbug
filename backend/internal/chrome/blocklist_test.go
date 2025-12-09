package chrome

import "testing"

func TestNewBlocklist(t *testing.T) {
	t.Run("empty blocklist", func(t *testing.T) {
		bl := NewBlocklist(false, false, false, nil)
		if !bl.IsEmpty() {
			t.Error("expected empty blocklist")
		}
	})

	t.Run("analytics enabled", func(t *testing.T) {
		bl := NewBlocklist(true, false, false, nil)
		if bl.IsEmpty() {
			t.Error("expected non-empty blocklist with analytics")
		}
		if len(bl.patterns) != len(analyticsPatterns) {
			t.Errorf("patterns count = %d, want %d", len(bl.patterns), len(analyticsPatterns))
		}
	})

	t.Run("all categories enabled", func(t *testing.T) {
		bl := NewBlocklist(true, true, true, nil)
		expectedCount := len(analyticsPatterns) + len(adsPatterns) + len(socialPatterns)
		if len(bl.patterns) != expectedCount {
			t.Errorf("patterns count = %d, want %d", len(bl.patterns), expectedCount)
		}
	})

	t.Run("with blocked types", func(t *testing.T) {
		bl := NewBlocklist(false, false, false, []string{"image", "font"})
		if len(bl.blockedTypes) != 2 {
			t.Errorf("blocked types count = %d, want 2", len(bl.blockedTypes))
		}
	})
}

func TestBlocklist_ShouldBlock_Analytics(t *testing.T) {
	bl := NewBlocklist(true, false, false, nil)

	tests := []struct {
		url      string
		expected bool
	}{
		{"https://www.google-analytics.com/analytics.js", true},
		{"https://www.googletagmanager.com/gtm.js?id=GTM-XXX", true},
		{"https://www.googletagmanager.com/gtag/js?id=G-XXX", true},
		{"https://static.hotjar.com/c/hotjar-123.js", true},
		{"https://cdn.segment.com/analytics.js/v1/xxx/analytics.min.js", true},
		{"https://api.mixpanel.com/track", true},
		{"https://api.amplitude.com/2/httpapi", true},
		{"https://cdn.heapanalytics.com/js/heap-123.js", true},
		{"https://plausible.io/api/event", true},
		{"https://example.com/matomo.js", true},
		{"https://clarity.ms/tag/xxx", true},
		{"https://cdn.mouseflow.com/projects/xxx.js", true},
		{"https://edge.fullstory.com/s/fs.js", true},
		{"https://cdn.logrocket.com/LogRocket.min.js", true},
		{"https://example.com/page.html", false},
		{"https://cdn.example.com/app.js", false},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			if got := bl.ShouldBlock(tt.url, ""); got != tt.expected {
				t.Errorf("ShouldBlock(%q) = %v, want %v", tt.url, got, tt.expected)
			}
		})
	}
}

func TestBlocklist_ShouldBlock_Ads(t *testing.T) {
	bl := NewBlocklist(false, true, false, nil)

	tests := []struct {
		url      string
		expected bool
	}{
		{"https://ad.doubleclick.net/ddm/trackclk/xxx", true},
		{"https://pagead2.googlesyndication.com/pagead/js/adsbygoogle.js", true},
		{"https://www.googleadservices.com/pagead/conversion/xxx", true},
		{"https://ib.adnxs.com/px?id=xxx", true},
		{"https://static.criteo.net/js/ld/publishertag.js", true},
		{"https://aax.amazon-adsystem.com/aax2/apstag.js", true},
		{"https://z.moatads.com/xxx/moatad.js", true},
		{"https://example.com/page.html", false},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			if got := bl.ShouldBlock(tt.url, ""); got != tt.expected {
				t.Errorf("ShouldBlock(%q) = %v, want %v", tt.url, got, tt.expected)
			}
		})
	}
}

func TestBlocklist_ShouldBlock_Social(t *testing.T) {
	bl := NewBlocklist(false, false, true, nil)

	tests := []struct {
		url      string
		expected bool
	}{
		{"https://www.facebook.com/tr?id=xxx", true},
		{"https://connect.facebook.net/en_US/fbevents.js", true},
		{"https://platform.twitter.com/widgets.js", true},
		{"https://analytics.twitter.com/i/adsct", true},
		{"https://snap.licdn.com/li.lms-analytics/insight.min.js", true},
		{"https://analytics.tiktok.com/i18n/pixel/events.js", true},
		{"https://ct.pinterest.com/v3/?event=init", true},
		{"https://example.com/page.html", false},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			if got := bl.ShouldBlock(tt.url, ""); got != tt.expected {
				t.Errorf("ShouldBlock(%q) = %v, want %v", tt.url, got, tt.expected)
			}
		})
	}
}

func TestBlocklist_ShouldBlock_ResourceTypes(t *testing.T) {
	bl := NewBlocklist(false, false, false, []string{"image", "font", "media"})

	tests := []struct {
		resourceType string
		expected     bool
	}{
		{"image", true},
		{"Image", true},
		{"IMAGE", true},
		{"font", true},
		{"media", true},
		{"script", false},
		{"stylesheet", false},
		{"document", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.resourceType, func(t *testing.T) {
			if got := bl.ShouldBlock("https://example.com/file", tt.resourceType); got != tt.expected {
				t.Errorf("ShouldBlock(type=%q) = %v, want %v", tt.resourceType, got, tt.expected)
			}
		})
	}
}

func TestBlocklist_ShouldBlock_Combined(t *testing.T) {
	bl := NewBlocklist(true, true, true, []string{"image"})

	tests := []struct {
		name         string
		url          string
		resourceType string
		expected     bool
	}{
		{"analytics URL", "https://google-analytics.com/collect", "", true},
		{"ads URL", "https://doubleclick.net/ad", "", true},
		{"social URL", "https://facebook.com/tr", "", true},
		{"image resource", "https://example.com/photo.jpg", "image", true},
		{"allowed URL and type", "https://example.com/app.js", "script", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := bl.ShouldBlock(tt.url, tt.resourceType); got != tt.expected {
				t.Errorf("ShouldBlock(%q, %q) = %v, want %v", tt.url, tt.resourceType, got, tt.expected)
			}
		})
	}
}

func TestBlocklist_ShouldBlock_EmptyBlocklist(t *testing.T) {
	bl := NewBlocklist(false, false, false, nil)

	if bl.ShouldBlock("https://google-analytics.com/collect", "") {
		t.Error("empty blocklist should not block anything")
	}
	if bl.ShouldBlock("https://example.com/image.jpg", "image") {
		t.Error("empty blocklist should not block any resource type")
	}
}

func TestBlocklist_ShouldBlock_NilBlocklist(t *testing.T) {
	var bl *Blocklist

	if bl.ShouldBlock("https://google-analytics.com/collect", "") {
		t.Error("nil blocklist should not block anything")
	}
}

func TestBlocklist_IsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		bl       *Blocklist
		expected bool
	}{
		{"nil blocklist", nil, true},
		{"empty blocklist", NewBlocklist(false, false, false, nil), true},
		{"with analytics", NewBlocklist(true, false, false, nil), false},
		{"with blocked types only", NewBlocklist(false, false, false, []string{"image"}), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.bl.IsEmpty(); got != tt.expected {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestWildcardMatch(t *testing.T) {
	tests := []struct {
		pattern  string
		text     string
		expected bool
	}{
		// Basic cases
		{"*", "anything", true},
		{"*", "", true},
		{"", "", true},
		{"", "text", false},

		// Prefix matching
		{"hello*", "hello world", true},
		{"hello*", "hello", true},
		{"hello*", "hi", false},

		// Suffix matching
		{"*world", "hello world", true},
		{"*world", "world", true},
		{"*world", "worldx", false},

		// Contains matching
		{"*test*", "this is a test case", true},
		{"*test*", "testing", true},
		{"*test*", "no match", false},

		// Multiple wildcards
		{"*google*analytics*", "https://www.google-analytics.com/collect", true},
		{"*a*b*c*", "aXXbYYc", true},
		{"*a*b*c*", "cba", false},

		// Case insensitivity (pattern is lowercased)
		{"*GOOGLE*", "https://google.com", true},
		{"*Google*", "https://GOOGLE.com", false}, // text is not lowercased in wildcardMatch

		// Exact match (no wildcards)
		{"exact", "exact", true},
		{"exact", "not exact", false},

		// Edge cases
		{"**", "anything", true},
		{"a*b*c", "abc", true},
		{"a*b*c", "aXXXbYYYc", true},
		{"a*b*c", "abcd", false},
	}

	for _, tt := range tests {
		name := tt.pattern + " matches " + tt.text
		t.Run(name, func(t *testing.T) {
			if got := wildcardMatch(tt.pattern, tt.text); got != tt.expected {
				t.Errorf("wildcardMatch(%q, %q) = %v, want %v", tt.pattern, tt.text, got, tt.expected)
			}
		})
	}
}

func TestBlocklist_CaseInsensitive(t *testing.T) {
	bl := NewBlocklist(true, false, false, nil)

	tests := []struct {
		url      string
		expected bool
	}{
		{"https://GOOGLE-ANALYTICS.COM/collect", true},
		{"https://Google-Analytics.com/collect", true},
		{"https://google-analytics.com/collect", true},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			if got := bl.ShouldBlock(tt.url, ""); got != tt.expected {
				t.Errorf("ShouldBlock(%q) = %v, want %v (case insensitive)", tt.url, got, tt.expected)
			}
		})
	}
}
