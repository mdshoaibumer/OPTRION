package server

import (
	"net/http"
	"testing"
)

func TestClientIP_XForwardedFor(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	r.RemoteAddr = "10.0.0.1:12345"
	r.Header.Set("X-Forwarded-For", "203.0.113.50, 70.41.3.18, 150.172.238.178")

	ip := ClientIP(r)
	if ip != "203.0.113.50" {
		t.Fatalf("expected 203.0.113.50, got %s", ip)
	}
}

func TestClientIP_XForwardedForSingle(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	r.RemoteAddr = "10.0.0.1:12345"
	r.Header.Set("X-Forwarded-For", "198.51.100.1")

	ip := ClientIP(r)
	if ip != "198.51.100.1" {
		t.Fatalf("expected 198.51.100.1, got %s", ip)
	}
}

func TestClientIP_XRealIP(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	r.RemoteAddr = "10.0.0.1:12345"
	r.Header.Set("X-Real-IP", "198.51.100.2")

	ip := ClientIP(r)
	if ip != "198.51.100.2" {
		t.Fatalf("expected 198.51.100.2, got %s", ip)
	}
}

func TestClientIP_FallbackToRemoteAddr(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	r.RemoteAddr = "192.0.2.1:8080"

	ip := ClientIP(r)
	if ip != "192.0.2.1" {
		t.Fatalf("expected 192.0.2.1, got %s", ip)
	}
}

func TestClientIP_XForwardedForPriority(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	r.RemoteAddr = "10.0.0.1:12345"
	r.Header.Set("X-Forwarded-For", "203.0.113.50")
	r.Header.Set("X-Real-IP", "198.51.100.2")

	// X-Forwarded-For takes priority over X-Real-IP
	ip := ClientIP(r)
	if ip != "203.0.113.50" {
		t.Fatalf("expected X-Forwarded-For to take priority, got %s", ip)
	}
}

func TestClientIP_EmptyHeaders(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	r.RemoteAddr = "127.0.0.1:9999"

	ip := ClientIP(r)
	if ip != "127.0.0.1" {
		t.Fatalf("expected 127.0.0.1, got %s", ip)
	}
}

func TestClientIP_WhitespaceInXForwardedFor(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	r.RemoteAddr = "10.0.0.1:12345"
	r.Header.Set("X-Forwarded-For", "  203.0.113.50 , 70.41.3.18")

	ip := ClientIP(r)
	if ip != "203.0.113.50" {
		t.Fatalf("expected trimmed IP 203.0.113.50, got %s", ip)
	}
}

func TestClientIP_InvalidIPInXForwardedFor(t *testing.T) {
	// Reset trusted proxies (allow all headers in test mode)
	SetTrustedProxies(nil)

	r, _ := http.NewRequest("GET", "/", nil)
	r.RemoteAddr = "10.0.0.1:12345"
	r.Header.Set("X-Forwarded-For", "not-an-ip-address")

	ip := ClientIP(r)
	// Invalid IP should be rejected, fall back to RemoteAddr
	if ip != "10.0.0.1" {
		t.Fatalf("expected fallback to RemoteAddr 10.0.0.1, got %s", ip)
	}
}

func TestClientIP_InvalidIPInXRealIP(t *testing.T) {
	SetTrustedProxies(nil)

	r, _ := http.NewRequest("GET", "/", nil)
	r.RemoteAddr = "10.0.0.1:12345"
	r.Header.Set("X-Real-IP", "evil-string")

	ip := ClientIP(r)
	if ip != "10.0.0.1" {
		t.Fatalf("expected fallback to RemoteAddr 10.0.0.1, got %s", ip)
	}
}

func TestClientIP_TrustedProxy_AllowsHeaders(t *testing.T) {
	SetTrustedProxies([]string{"10.0.0.1"})
	defer SetTrustedProxies(nil)

	r, _ := http.NewRequest("GET", "/", nil)
	r.RemoteAddr = "10.0.0.1:12345" // From trusted proxy
	r.Header.Set("X-Forwarded-For", "203.0.113.50")

	ip := ClientIP(r)
	if ip != "203.0.113.50" {
		t.Fatalf("expected 203.0.113.50 from trusted proxy, got %s", ip)
	}
}

func TestClientIP_UntrustedProxy_IgnoresHeaders(t *testing.T) {
	SetTrustedProxies([]string{"172.16.0.1"}) // Only trust 172.16.0.1
	defer SetTrustedProxies(nil)

	r, _ := http.NewRequest("GET", "/", nil)
	r.RemoteAddr = "192.168.1.100:12345"       // NOT a trusted proxy
	r.Header.Set("X-Forwarded-For", "1.2.3.4") // Spoofed header

	ip := ClientIP(r)
	// Should ignore X-Forwarded-For because RemoteAddr is not trusted
	if ip != "192.168.1.100" {
		t.Fatalf("expected RemoteAddr 192.168.1.100 (untrusted proxy), got %s", ip)
	}
}

func TestClientIP_EmptyTrustedProxies_TrustsAll(t *testing.T) {
	SetTrustedProxies(nil) // Empty = trust all (dev mode)

	r, _ := http.NewRequest("GET", "/", nil)
	r.RemoteAddr = "10.0.0.1:12345"
	r.Header.Set("X-Forwarded-For", "8.8.8.8")

	ip := ClientIP(r)
	if ip != "8.8.8.8" {
		t.Fatalf("expected 8.8.8.8 (all trusted in dev mode), got %s", ip)
	}
}
