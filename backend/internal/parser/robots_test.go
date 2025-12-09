package parser

import "testing"

func TestParseRobotsDirectives(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		wantIndexable bool
		wantFollow    bool
	}{
		// Single directives
		{
			name:          "noindex only",
			content:       "noindex",
			wantIndexable: false,
			wantFollow:    true,
		},
		{
			name:          "nofollow only",
			content:       "nofollow",
			wantIndexable: true,
			wantFollow:    false,
		},
		{
			name:          "none directive",
			content:       "none",
			wantIndexable: false,
			wantFollow:    false,
		},
		{
			name:          "all directive",
			content:       "all",
			wantIndexable: true,
			wantFollow:    true,
		},
		{
			name:          "index directive",
			content:       "index",
			wantIndexable: true,
			wantFollow:    true,
		},
		{
			name:          "follow directive",
			content:       "follow",
			wantIndexable: true,
			wantFollow:    true,
		},

		// Combined directives
		{
			name:          "noindex nofollow combined",
			content:       "noindex, nofollow",
			wantIndexable: false,
			wantFollow:    false,
		},
		{
			name:          "index nofollow combined",
			content:       "index, nofollow",
			wantIndexable: true,
			wantFollow:    false,
		},
		{
			name:          "noindex follow combined",
			content:       "noindex, follow",
			wantIndexable: false,
			wantFollow:    true,
		},

		// Case insensitivity
		{
			name:          "uppercase NOINDEX",
			content:       "NOINDEX",
			wantIndexable: false,
			wantFollow:    true,
		},
		{
			name:          "mixed case NoFollow",
			content:       "NoFollow",
			wantIndexable: true,
			wantFollow:    false,
		},
		{
			name:          "uppercase NONE",
			content:       "NONE",
			wantIndexable: false,
			wantFollow:    false,
		},

		// Whitespace handling
		{
			name:          "extra whitespace",
			content:       "  noindex  ,  nofollow  ",
			wantIndexable: false,
			wantFollow:    false,
		},
		{
			name:          "no spaces after comma",
			content:       "noindex,nofollow",
			wantIndexable: false,
			wantFollow:    false,
		},

		// Empty/missing content (defaults)
		{
			name:          "empty string",
			content:       "",
			wantIndexable: true,
			wantFollow:    true,
		},
		{
			name:          "whitespace only",
			content:       "   ",
			wantIndexable: true,
			wantFollow:    true,
		},

		// Unknown directives (should be ignored)
		{
			name:          "noarchive ignored",
			content:       "noarchive",
			wantIndexable: true,
			wantFollow:    true,
		},
		{
			name:          "nosnippet ignored",
			content:       "nosnippet",
			wantIndexable: true,
			wantFollow:    true,
		},
		{
			name:          "mixed known and unknown",
			content:       "noindex, noarchive, nosnippet",
			wantIndexable: false,
			wantFollow:    true,
		},
		{
			name:          "max-snippet ignored",
			content:       "max-snippet:50, nofollow",
			wantIndexable: true,
			wantFollow:    false,
		},

		// Order matters - later overrides earlier
		{
			name:          "noindex then index",
			content:       "noindex, index",
			wantIndexable: true,
			wantFollow:    true,
		},
		{
			name:          "index then noindex",
			content:       "index, noindex",
			wantIndexable: false,
			wantFollow:    true,
		},
		{
			name:          "none then all",
			content:       "none, all",
			wantIndexable: true,
			wantFollow:    true,
		},
		{
			name:          "all then none",
			content:       "all, none",
			wantIndexable: false,
			wantFollow:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIndexable, gotFollow := ParseRobotsDirectives(tt.content)
			if gotIndexable != tt.wantIndexable {
				t.Errorf("ParseRobotsDirectives(%q) indexable = %v, want %v", tt.content, gotIndexable, tt.wantIndexable)
			}
			if gotFollow != tt.wantFollow {
				t.Errorf("ParseRobotsDirectives(%q) follow = %v, want %v", tt.content, gotFollow, tt.wantFollow)
			}
		})
	}
}

func TestGetRobotsFromMeta(t *testing.T) {
	tests := []struct {
		name          string
		googlebot     string
		robots        string
		xRobotsHeader string
		wantIndexable bool
		wantFollow    bool
	}{
		// Single source tests
		{
			name:          "only googlebot noindex",
			googlebot:     "noindex",
			robots:        "",
			xRobotsHeader: "",
			wantIndexable: false,
			wantFollow:    true,
		},
		{
			name:          "only robots nofollow",
			googlebot:     "",
			robots:        "nofollow",
			xRobotsHeader: "",
			wantIndexable: true,
			wantFollow:    false,
		},
		{
			name:          "only X-Robots-Tag none",
			googlebot:     "",
			robots:        "",
			xRobotsHeader: "none",
			wantIndexable: false,
			wantFollow:    false,
		},

		// All empty (defaults)
		{
			name:          "all empty",
			googlebot:     "",
			robots:        "",
			xRobotsHeader: "",
			wantIndexable: true,
			wantFollow:    true,
		},

		// Priority/conflict resolution - more restrictive wins
		{
			name:          "googlebot index vs xRobotsHeader noindex",
			googlebot:     "index",
			robots:        "",
			xRobotsHeader: "noindex",
			wantIndexable: false, // more restrictive wins
			wantFollow:    true,
		},
		{
			name:          "robots follow vs googlebot nofollow",
			googlebot:     "nofollow",
			robots:        "follow",
			xRobotsHeader: "",
			wantIndexable: true,
			wantFollow:    false, // more restrictive wins
		},
		{
			name:          "all sources conflicting",
			googlebot:     "index, follow",
			robots:        "noindex",
			xRobotsHeader: "nofollow",
			wantIndexable: false, // robots says noindex
			wantFollow:    false, // header says nofollow
		},

		// Combined restrictive settings
		{
			name:          "header none overrides all",
			googlebot:     "all",
			robots:        "all",
			xRobotsHeader: "none",
			wantIndexable: false,
			wantFollow:    false,
		},
		{
			name:          "multiple sources all permissive",
			googlebot:     "index, follow",
			robots:        "all",
			xRobotsHeader: "index",
			wantIndexable: true,
			wantFollow:    true,
		},

		// Real-world examples
		{
			name:          "typical noindex page",
			googlebot:     "",
			robots:        "noindex, follow",
			xRobotsHeader: "",
			wantIndexable: false,
			wantFollow:    true,
		},
		{
			name:          "googlebot specific noindex",
			googlebot:     "noindex",
			robots:        "index, follow",
			xRobotsHeader: "",
			wantIndexable: false, // googlebot is more restrictive
			wantFollow:    true,
		},
		{
			name:          "header overrides permissive meta",
			googlebot:     "",
			robots:        "index, follow",
			xRobotsHeader: "noindex, nofollow",
			wantIndexable: false,
			wantFollow:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIndexable, gotFollow := GetRobotsFromMeta(tt.googlebot, tt.robots, tt.xRobotsHeader)
			if gotIndexable != tt.wantIndexable {
				t.Errorf("GetRobotsFromMeta(%q, %q, %q) indexable = %v, want %v",
					tt.googlebot, tt.robots, tt.xRobotsHeader, gotIndexable, tt.wantIndexable)
			}
			if gotFollow != tt.wantFollow {
				t.Errorf("GetRobotsFromMeta(%q, %q, %q) follow = %v, want %v",
					tt.googlebot, tt.robots, tt.xRobotsHeader, gotFollow, tt.wantFollow)
			}
		})
	}
}
