package collector

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// ValidateTargetURL checks that a target URL is safe to fetch (no SSRF).
// It rejects URLs pointing to private/internal networks, localhost, link-local, etc.
func ValidateTargetURL(targetURL string) error {
	parsed, err := url.Parse(targetURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Only allow HTTP/HTTPS
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("unsupported scheme %q, only http and https are allowed", parsed.Scheme)
	}

	// Extract hostname (without port)
	host := parsed.Hostname()
	if host == "" {
		return fmt.Errorf("empty hostname")
	}

	// Reject localhost variants
	lowerHost := strings.ToLower(host)
	if lowerHost == "localhost" || lowerHost == "127.0.0.1" || lowerHost == "::1" || lowerHost == "0.0.0.0" {
		return fmt.Errorf("localhost targets are not allowed")
	}

	// Resolve and check the IP
	ips, err := net.LookupIP(host)
	if err != nil {
		// DNS resolution failure — fail closed to prevent SSRF bypass
		return fmt.Errorf("DNS resolution failed for %q: %w (request blocked for safety)", host, err)
	}

	if len(ips) == 0 {
		return fmt.Errorf("DNS returned no addresses for %q", host)
	}

	for _, ip := range ips {
		if isPrivateIP(ip) {
			return fmt.Errorf("target resolves to private IP %s, which is not allowed", ip.String())
		}
	}

	return nil
}

// isPrivateIP checks if an IP address is in a private/reserved range.
func isPrivateIP(ip net.IP) bool {
	// Check standard private ranges
	privateRanges := []struct {
		network *net.IPNet
	}{
		{mustParseCIDR("10.0.0.0/8")},
		{mustParseCIDR("172.16.0.0/12")},
		{mustParseCIDR("192.168.0.0/16")},
		{mustParseCIDR("127.0.0.0/8")},     // Loopback
		{mustParseCIDR("169.254.0.0/16")},  // Link-local
		{mustParseCIDR("::1/128")},         // IPv6 loopback
		{mustParseCIDR("fc00::/7")},        // IPv6 unique local
		{mustParseCIDR("fe80::/10")},       // IPv6 link-local
		{mustParseCIDR("100.64.0.0/10")},   // Shared address space (CGNAT)
		{mustParseCIDR("0.0.0.0/8")},       // "This" network
		{mustParseCIDR("192.0.0.0/24")},    // IANA special purpose
		{mustParseCIDR("192.0.2.0/24")},    // TEST-NET-1
		{mustParseCIDR("198.51.100.0/24")}, // TEST-NET-2
		{mustParseCIDR("203.0.113.0/24")},  // TEST-NET-3
		{mustParseCIDR("224.0.0.0/4")},     // Multicast
		{mustParseCIDR("240.0.0.0/4")},     // Reserved
	}

	for _, r := range privateRanges {
		if r.network.Contains(ip) {
			return true
		}
	}

	return false
}

func mustParseCIDR(s string) *net.IPNet {
	_, network, err := net.ParseCIDR(s)
	if err != nil {
		panic(fmt.Sprintf("invalid CIDR: %s", s))
	}
	return network
}
