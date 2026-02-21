package security

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsPrivateIP(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		private bool
	}{
		// Loopback
		{"loopback 127.0.0.1", "127.0.0.1", true},
		{"loopback 127.255.255.255", "127.255.255.255", true},
		{"loopback IPv6", "::1", true},

		// RFC 1918
		{"rfc1918 10.0.0.1", "10.0.0.1", true},
		{"rfc1918 10.255.255.255", "10.255.255.255", true},
		{"rfc1918 172.16.0.1", "172.16.0.1", true},
		{"rfc1918 172.31.255.255", "172.31.255.255", true},
		{"rfc1918 192.168.0.1", "192.168.0.1", true},
		{"rfc1918 192.168.255.255", "192.168.255.255", true},

		// Link-local (includes AWS metadata)
		{"link-local 169.254.0.1", "169.254.0.1", true},
		{"link-local 169.254.169.254", "169.254.169.254", true},
		{"link-local IPv6 fe80::1", "fe80::1", true},

		// CGNAT (RFC 6598)
		{"cgnat 100.64.0.1", "100.64.0.1", true},
		{"cgnat 100.127.255.255", "100.127.255.255", true},

		// "This" network
		{"this-network 0.0.0.0", "0.0.0.0", true},
		{"this-network 0.255.255.255", "0.255.255.255", true},

		// Multicast
		{"multicast 224.0.0.1", "224.0.0.1", true},
		{"multicast 239.255.255.255", "239.255.255.255", true},
		{"multicast IPv6 ff02::1", "ff02::1", true},

		// IPv6 unique local
		{"unique-local fd00::1", "fd00::1", true},
		{"unique-local fc00::1", "fc00::1", true},

		// Public IPs (should NOT be private)
		{"public 8.8.8.8", "8.8.8.8", false},
		{"public 1.1.1.1", "1.1.1.1", false},
		{"public 93.184.216.34", "93.184.216.34", false},
		{"public 172.32.0.1", "172.32.0.1", false},   // just outside 172.16.0.0/12
		{"public 100.128.0.1", "100.128.0.1", false}, // just outside 100.64.0.0/10
		{"public 11.0.0.1", "11.0.0.1", false},       // just outside 10.0.0.0/8
		{"public IPv6 2001:db8::1", "2001:db8::1", false},
		{"public IPv6 2607:f8b0:4004:800::200e", "2607:f8b0:4004:800::200e", false},

		// Nil
		{"nil IP", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ip net.IP
			if tt.ip != "" {
				ip = net.ParseIP(tt.ip)
				require.NotNil(t, ip, "failed to parse test IP: %s", tt.ip)
			}
			assert.Equal(t, tt.private, IsPrivateIP(ip))
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantError bool
		errMsg    string
	}{
		// Private IP literals should be blocked
		{"blocks loopback", "http://127.0.0.1/", true, "private/reserved"},
		{"blocks loopback with port", "http://127.0.0.1:8080/path", true, "private/reserved"},
		{"blocks rfc1918 10.x", "http://10.0.0.1/", true, "private/reserved"},
		{"blocks rfc1918 172.16.x", "http://172.16.0.1/", true, "private/reserved"},
		{"blocks rfc1918 192.168.x", "http://192.168.1.1/", true, "private/reserved"},
		{"blocks link-local", "http://169.254.169.254/latest/meta-data/", true, "private/reserved"},
		{"blocks cgnat", "http://100.64.0.1/", true, "private/reserved"},
		{"blocks zero", "http://0.0.0.0/", true, "private/reserved"},
		{"blocks IPv6 loopback", "http://[::1]/", true, "private/reserved"},
		{"blocks IPv6 link-local", "http://[fe80::1]/", true, "private/reserved"},
		{"blocks IPv6 unique-local", "http://[fd00::1]/", true, "private/reserved"},

		// Blocked hostnames
		{"blocks localhost", "http://localhost/", true, "not allowed"},
		{"blocks localhost with port", "http://localhost:3000/", true, "not allowed"},
		{"blocks LOCALHOST", "http://LOCALHOST/", true, "not allowed"},

		// Public IPs should pass
		{"allows public IP", "http://8.8.8.8/", false, ""},
		{"allows public IP with port", "http://93.184.216.34:443/", false, ""},
		{"allows public IPv6", "http://[2607:f8b0:4004:800::200e]/", false, ""},

		// Edge cases for range boundaries
		{"allows 172.32.0.1", "http://172.32.0.1/", false, ""},   // just outside /12
		{"allows 100.128.0.1", "http://100.128.0.1/", false, ""}, // just outside /10
		{"allows 11.0.0.1", "http://11.0.0.1/", false, ""},       // just outside 10.0.0.0/8

		// Public domains (DNS resolution tested only if available)
		{"allows public domain", "https://example.com/", false, ""},

		// No hostname
		{"rejects empty hostname", "http:///path", true, "no hostname"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.url)
			if tt.wantError {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
