package parser

import "testing"

func TestIsSocialURL(t *testing.T) {
	tests := []struct {
		name   string
		rawURL string
		want   bool
	}{
		// Direct matches
		{
			name:   "facebook direct",
			rawURL: "https://facebook.com/page",
			want:   true,
		},
		{
			name:   "twitter direct",
			rawURL: "https://twitter.com/user",
			want:   true,
		},
		{
			name:   "x.com direct",
			rawURL: "https://x.com/user",
			want:   true,
		},
		{
			name:   "linkedin direct",
			rawURL: "https://linkedin.com/in/user",
			want:   true,
		},
		{
			name:   "instagram direct",
			rawURL: "https://instagram.com/user",
			want:   true,
		},
		{
			name:   "youtube direct",
			rawURL: "https://youtube.com/watch?v=123",
			want:   true,
		},
		{
			name:   "tiktok direct",
			rawURL: "https://tiktok.com/@user",
			want:   true,
		},
		{
			name:   "pinterest direct",
			rawURL: "https://pinterest.com/pin/123",
			want:   true,
		},
		{
			name:   "reddit direct",
			rawURL: "https://reddit.com/r/subreddit",
			want:   true,
		},
		{
			name:   "discord direct",
			rawURL: "https://discord.com/invite/abc",
			want:   true,
		},
		{
			name:   "snapchat direct",
			rawURL: "https://snapchat.com/add/user",
			want:   true,
		},
		{
			name:   "whatsapp direct",
			rawURL: "https://whatsapp.com/channel/123",
			want:   true,
		},
		{
			name:   "telegram direct",
			rawURL: "https://telegram.org/channel",
			want:   true,
		},
		{
			name:   "tumblr direct",
			rawURL: "https://tumblr.com/blog",
			want:   true,
		},
		{
			name:   "threads direct",
			rawURL: "https://threads.net/@user",
			want:   true,
		},

		// Subdomain matches
		{
			name:   "mobile facebook",
			rawURL: "https://m.facebook.com/page",
			want:   true,
		},
		{
			name:   "mobile twitter",
			rawURL: "https://m.twitter.com/user",
			want:   true,
		},
		{
			name:   "www facebook",
			rawURL: "https://www.facebook.com/page",
			want:   true,
		},
		{
			name:   "www linkedin",
			rawURL: "https://www.linkedin.com/in/user",
			want:   true,
		},
		{
			name:   "youtube with subdomain",
			rawURL: "https://www.youtube.com/watch?v=123",
			want:   true,
		},
		{
			name:   "deep subdomain instagram",
			rawURL: "https://api.graph.instagram.com/endpoint",
			want:   true,
		},
		{
			name:   "t.co (twitter short URL)",
			rawURL: "https://t.co/abc123",
			want:   false, // t.co is a different domain, not twitter.com
		},

		// Non-social URLs
		{
			name:   "google",
			rawURL: "https://www.google.com/search",
			want:   false,
		},
		{
			name:   "amazon",
			rawURL: "https://amazon.com/product",
			want:   false,
		},
		{
			name:   "github",
			rawURL: "https://github.com/user/repo",
			want:   false,
		},
		{
			name:   "example.com",
			rawURL: "https://example.com/page",
			want:   false,
		},
		{
			name:   "random blog",
			rawURL: "https://myblog.wordpress.com",
			want:   false,
		},

		// URLs without scheme
		{
			name:   "facebook without scheme",
			rawURL: "facebook.com/page",
			want:   true,
		},
		{
			name:   "twitter without scheme",
			rawURL: "m.twitter.com/user",
			want:   true,
		},

		// Protocol-relative URLs
		{
			name:   "protocol-relative facebook",
			rawURL: "//www.facebook.com/page",
			want:   true,
		},

		// Edge cases - invalid URLs (should return false, not panic)
		{
			name:   "empty string",
			rawURL: "",
			want:   false,
		},
		{
			name:   "relative URL",
			rawURL: "/page/about",
			want:   false,
		},
		{
			name:   "fragment only",
			rawURL: "#section",
			want:   false,
		},
		{
			name:   "javascript URL",
			rawURL: "javascript:void(0)",
			want:   false,
		},
		{
			name:   "mailto URL",
			rawURL: "mailto:user@facebook.com",
			want:   false, // mailto is not a social URL even with social domain in email
		},

		// Case insensitivity
		{
			name:   "uppercase FACEBOOK",
			rawURL: "https://FACEBOOK.COM/page",
			want:   true,
		},
		{
			name:   "mixed case Twitter",
			rawURL: "https://Twitter.Com/user",
			want:   true,
		},

		// Similar but not social domains
		{
			name:   "facebookmail.com (not facebook)",
			rawURL: "https://facebookmail.com",
			want:   false,
		},
		{
			name:   "nottwitter.com",
			rawURL: "https://nottwitter.com",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsSocialURL(tt.rawURL)
			if got != tt.want {
				t.Errorf("IsSocialURL(%q) = %v, want %v", tt.rawURL, got, tt.want)
			}
		})
	}
}

func TestSocialDomainsCompleteness(t *testing.T) {
	// Verify all expected social domains are in the map
	expectedDomains := []string{
		"facebook.com",
		"twitter.com",
		"x.com",
		"linkedin.com",
		"instagram.com",
		"youtube.com",
		"tiktok.com",
		"pinterest.com",
		"reddit.com",
		"discord.com",
		"snapchat.com",
		"whatsapp.com",
		"telegram.org",
		"tumblr.com",
		"threads.net",
	}

	for _, domain := range expectedDomains {
		if !socialDomains[domain] {
			t.Errorf("Expected domain %q to be in socialDomains map", domain)
		}
	}

	// Verify count matches
	if len(socialDomains) != len(expectedDomains) {
		t.Errorf("socialDomains has %d entries, expected %d", len(socialDomains), len(expectedDomains))
	}
}
