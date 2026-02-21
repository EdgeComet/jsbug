package security

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
)

// privateRanges contains all private and reserved IP ranges that should be blocked
// to prevent SSRF attacks.
var privateRanges []*net.IPNet

func init() {
	cidrs := []string{
		// IPv4
		"127.0.0.0/8",    // loopback
		"10.0.0.0/8",     // RFC 1918
		"172.16.0.0/12",  // RFC 1918
		"192.168.0.0/16", // RFC 1918
		"169.254.0.0/16", // link-local (includes AWS metadata 169.254.169.254)
		"100.64.0.0/10",  // CGNAT (RFC 6598)
		"0.0.0.0/8",      // "this" network
		"224.0.0.0/4",    // multicast

		// IPv6
		"::1/128",   // loopback
		"fe80::/10", // link-local
		"fc00::/7",  // unique local
		"ff00::/8",  // multicast
	}

	for _, cidr := range cidrs {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			panic(fmt.Sprintf("invalid CIDR in SSRF private ranges: %s", cidr))
		}
		privateRanges = append(privateRanges, ipNet)
	}
}

// blockedHostnames contains hostnames that resolve to private IPs and must be
// blocked before DNS resolution (Chrome does its own resolution).
var blockedHostnames = map[string]bool{
	"localhost": true,
}

// IsPrivateIP returns true if the given IP belongs to a private or reserved range.
func IsPrivateIP(ip net.IP) bool {
	if ip == nil {
		return false
	}

	for _, ipNet := range privateRanges {
		if ipNet.Contains(ip) {
			return true
		}
	}
	return false
}

// ValidateURL checks that a URL does not target private/internal network resources.
// It performs:
//  1. Hostname extraction (strips port, IPv6 brackets)
//  2. Blocked hostname check (localhost)
//  3. IP literal check against private ranges
//  4. DNS resolution with all resolved IPs checked against private ranges
func ValidateURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	hostname := u.Hostname() // strips port and IPv6 brackets

	if hostname == "" {
		return fmt.Errorf("URL has no hostname")
	}

	// Block known dangerous hostnames
	if blockedHostnames[strings.ToLower(hostname)] {
		return fmt.Errorf("hostname %q is not allowed", hostname)
	}

	// Check if hostname is an IP literal
	if ip := net.ParseIP(hostname); ip != nil {
		if IsPrivateIP(ip) {
			return fmt.Errorf("IP address %s is in a private/reserved range", hostname)
		}
		return nil
	}

	// Hostname is a domain name: resolve and check all IPs
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ips, err := net.DefaultResolver.LookupIPAddr(ctx, hostname)
	if err != nil {
		// DNS resolution failure is not an SSRF issue; let it fail later at fetch/render
		return nil
	}

	for _, ipAddr := range ips {
		if IsPrivateIP(ipAddr.IP) {
			return fmt.Errorf("hostname %q resolves to private/reserved IP %s", hostname, ipAddr.IP)
		}
	}

	return nil
}

// SSRFSafeDialContext returns a DialContext function that validates resolved IPs
// against private ranges before connecting. Use as http.Transport.DialContext for
// defense-in-depth against DNS rebinding in the HTTP fetcher path.
func SSRFSafeDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("invalid address %q: %w", addr, err)
	}

	// Block known dangerous hostnames
	if blockedHostnames[strings.ToLower(host)] {
		return nil, fmt.Errorf("SSRF protection: hostname %q is not allowed", host)
	}

	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("DNS resolution failed for %q: %w", host, err)
	}

	if len(ips) == 0 {
		return nil, fmt.Errorf("no IP addresses found for %q", host)
	}

	for _, ipAddr := range ips {
		if IsPrivateIP(ipAddr.IP) {
			return nil, fmt.Errorf("SSRF protection: %q resolves to private/reserved IP %s", host, ipAddr.IP)
		}
	}

	// All IPs validated as public; connect to the first one
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	return dialer.DialContext(ctx, network, net.JoinHostPort(ips[0].IP.String(), port))
}
