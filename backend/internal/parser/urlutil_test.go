package parser

import (
	"testing"
)

func TestExtractBaseDomain(t *testing.T) {
	tests := []struct {
		name    string
		rawURL  string
		want    string
		wantErr bool
	}{
		// Standard URLs with various schemes
		{
			name:   "https URL",
			rawURL: "https://www.example.com/path/to/page",
			want:   "example.com",
		},
		{
			name:   "http URL",
			rawURL: "http://example.com/path",
			want:   "example.com",
		},
		{
			name:   "URL without scheme",
			rawURL: "www.example.com/path",
			want:   "example.com",
		},
		{
			name:   "protocol-relative URL",
			rawURL: "//cdn.example.com/assets",
			want:   "example.com",
		},

		// URLs with ports
		{
			name:   "URL with port",
			rawURL: "https://example.com:8080/path",
			want:   "example.com",
		},
		{
			name:   "URL with port no scheme",
			rawURL: "example.com:3000",
			want:   "example.com",
		},

		// URLs with subdomains
		{
			name:   "www subdomain",
			rawURL: "https://www.example.com",
			want:   "example.com",
		},
		{
			name:   "cdn subdomain",
			rawURL: "https://cdn.example.com",
			want:   "example.com",
		},
		{
			name:   "api subdomain",
			rawURL: "https://api.v2.example.com/endpoint",
			want:   "example.com",
		},
		{
			name:   "deep subdomain",
			rawURL: "https://a.b.c.example.com",
			want:   "example.com",
		},

		// Multi-part TLDs
		{
			name:   "co.uk TLD",
			rawURL: "https://www.example.co.uk/page",
			want:   "example.co.uk",
		},
		{
			name:   "com.au TLD",
			rawURL: "https://shop.example.com.au",
			want:   "example.com.au",
		},
		{
			name:   "co.jp TLD",
			rawURL: "https://example.co.jp",
			want:   "example.co.jp",
		},

		// IP addresses
		{
			name:   "IPv4 address",
			rawURL: "http://192.168.1.1/admin",
			want:   "192.168.1.1",
		},
		{
			name:   "IPv4 with port",
			rawURL: "http://10.0.0.1:8080",
			want:   "10.0.0.1",
		},
		{
			name:   "localhost IP",
			rawURL: "http://127.0.0.1:3000/api",
			want:   "127.0.0.1",
		},

		// Case sensitivity
		{
			name:   "uppercase domain",
			rawURL: "https://WWW.EXAMPLE.COM/Path",
			want:   "example.com",
		},
		{
			name:   "mixed case",
			rawURL: "https://Cdn.ExAmPlE.cOm",
			want:   "example.com",
		},

		// Edge cases - errors
		{
			name:    "empty string",
			rawURL:  "",
			wantErr: true,
		},
		{
			name:   "just domain",
			rawURL: "example.com",
			want:   "example.com",
		},
		{
			name:   "single part domain",
			rawURL: "localhost",
			want:   "localhost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractBaseDomain(tt.rawURL)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ExtractBaseDomain(%q) expected error, got nil", tt.rawURL)
				}
				return
			}
			if err != nil {
				t.Errorf("ExtractBaseDomain(%q) unexpected error: %v", tt.rawURL, err)
				return
			}
			if got != tt.want {
				t.Errorf("ExtractBaseDomain(%q) = %q, want %q", tt.rawURL, got, tt.want)
			}
		})
	}
}

func TestIsSameDomain(t *testing.T) {
	tests := []struct {
		name string
		urlA string
		urlB string
		want bool
	}{
		{
			name: "same domain different paths",
			urlA: "https://example.com/page1",
			urlB: "https://example.com/page2",
			want: true,
		},
		{
			name: "same domain with www",
			urlA: "https://www.example.com",
			urlB: "https://example.com",
			want: true,
		},
		{
			name: "different subdomains same domain",
			urlA: "https://cdn.example.com",
			urlB: "https://api.example.com",
			want: true,
		},
		{
			name: "different domains",
			urlA: "https://example.com",
			urlB: "https://other.com",
			want: false,
		},
		{
			name: "different TLDs",
			urlA: "https://example.com",
			urlB: "https://example.org",
			want: false,
		},
		{
			name: "case insensitive",
			urlA: "https://EXAMPLE.COM",
			urlB: "https://example.com",
			want: true,
		},
		{
			name: "with and without scheme",
			urlA: "https://example.com",
			urlB: "example.com",
			want: true,
		},
		{
			name: "empty URL A",
			urlA: "",
			urlB: "https://example.com",
			want: false,
		},
		{
			name: "empty URL B",
			urlA: "https://example.com",
			urlB: "",
			want: false,
		},
		{
			name: "both empty",
			urlA: "",
			urlB: "",
			want: false,
		},
		{
			name: "multi-part TLD same domain",
			urlA: "https://www.example.co.uk",
			urlB: "https://shop.example.co.uk",
			want: true,
		},
		{
			name: "IP addresses same",
			urlA: "http://192.168.1.1/path",
			urlB: "http://192.168.1.1:8080/other",
			want: true,
		},
		{
			name: "IP addresses different",
			urlA: "http://192.168.1.1",
			urlB: "http://192.168.1.2",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsSameDomain(tt.urlA, tt.urlB)
			if got != tt.want {
				t.Errorf("IsSameDomain(%q, %q) = %v, want %v", tt.urlA, tt.urlB, got, tt.want)
			}
		})
	}
}

func TestIsSubdomainOf(t *testing.T) {
	tests := []struct {
		name         string
		childURL     string
		parentDomain string
		want         bool
	}{
		// Basic cases
		{
			name:         "exact match",
			childURL:     "https://example.com",
			parentDomain: "example.com",
			want:         true,
		},
		{
			name:         "www subdomain",
			childURL:     "https://www.example.com",
			parentDomain: "example.com",
			want:         true,
		},
		{
			name:         "cdn subdomain",
			childURL:     "https://cdn.example.com/assets/file.js",
			parentDomain: "example.com",
			want:         true,
		},
		{
			name:         "api subdomain",
			childURL:     "https://api.example.com",
			parentDomain: "example.com",
			want:         true,
		},
		{
			name:         "deep subdomain",
			childURL:     "https://a.b.c.example.com",
			parentDomain: "example.com",
			want:         true,
		},

		// Non-matches
		{
			name:         "different domain",
			childURL:     "https://other.com",
			parentDomain: "example.com",
			want:         false,
		},
		{
			name:         "similar but different domain",
			childURL:     "https://notexample.com",
			parentDomain: "example.com",
			want:         false,
		},
		{
			name:         "domain contains parent as substring",
			childURL:     "https://myexample.com",
			parentDomain: "example.com",
			want:         false,
		},

		// Case insensitivity
		{
			name:         "case insensitive child",
			childURL:     "https://CDN.EXAMPLE.COM",
			parentDomain: "example.com",
			want:         true,
		},
		{
			name:         "case insensitive parent",
			childURL:     "https://cdn.example.com",
			parentDomain: "EXAMPLE.COM",
			want:         true,
		},

		// Edge cases
		{
			name:         "empty child URL",
			childURL:     "",
			parentDomain: "example.com",
			want:         false,
		},
		{
			name:         "empty parent domain",
			childURL:     "https://example.com",
			parentDomain: "",
			want:         false,
		},
		{
			name:         "parent domain with scheme (normalized)",
			childURL:     "https://cdn.example.com",
			parentDomain: "https://example.com",
			want:         true,
		},
		{
			name:         "parent domain with whitespace",
			childURL:     "https://cdn.example.com",
			parentDomain: "  example.com  ",
			want:         true,
		},

		// Protocol-relative URLs
		{
			name:         "protocol-relative subdomain",
			childURL:     "//cdn.example.com/file.js",
			parentDomain: "example.com",
			want:         true,
		},

		// URLs without scheme
		{
			name:         "child without scheme",
			childURL:     "cdn.example.com/path",
			parentDomain: "example.com",
			want:         true,
		},

		// Multi-part TLDs
		{
			name:         "co.uk subdomain",
			childURL:     "https://www.example.co.uk",
			parentDomain: "example.co.uk",
			want:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsSubdomainOf(tt.childURL, tt.parentDomain)
			if got != tt.want {
				t.Errorf("IsSubdomainOf(%q, %q) = %v, want %v", tt.childURL, tt.parentDomain, got, tt.want)
			}
		})
	}
}

func TestIsIPAddress(t *testing.T) {
	tests := []struct {
		host string
		want bool
	}{
		{"192.168.1.1", true},
		{"10.0.0.1", true},
		{"127.0.0.1", true},
		{"255.255.255.255", true},
		{"0.0.0.0", true},
		{"example.com", false},
		{"www.example.com", false},
		{"192.168.1", false},     // incomplete
		{"192.168.1.1.1", false}, // too many parts
		{"192.168.1.a", false},   // non-numeric
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			got := isIPAddress(tt.host)
			if got != tt.want {
				t.Errorf("isIPAddress(%q) = %v, want %v", tt.host, got, tt.want)
			}
		})
	}
}

func TestGetBaseDomain(t *testing.T) {
	tests := []struct {
		host string
		want string
	}{
		{"example.com", "example.com"},
		{"www.example.com", "example.com"},
		{"cdn.example.com", "example.com"},
		{"a.b.c.example.com", "example.com"},
		{"example.co.uk", "example.co.uk"},
		{"www.example.co.uk", "example.co.uk"},
		{"shop.example.com.au", "example.com.au"},
		{"localhost", "localhost"},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			got := getBaseDomain(tt.host)
			if got != tt.want {
				t.Errorf("getBaseDomain(%q) = %q, want %q", tt.host, got, tt.want)
			}
		})
	}
}
