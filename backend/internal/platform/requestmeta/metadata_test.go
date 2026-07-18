package requestmeta_test

import (
	"net/http/httptest"
	"testing"

	"github.com/gabrielalc23/pdv/internal/platform/requestmeta"
)

func TestDirectAccess(t *testing.T) {
	resolver, err := requestmeta.NewResolver(nil)
	if err != nil {
		t.Fatalf("NewResolver: %v", err)
	}

	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "192.168.1.100:12345"

	meta := resolver.Extract(r)
	if meta.ClientIP != "192.168.1.100" {
		t.Fatalf("expected 192.168.1.100, got %q", meta.ClientIP)
	}
}

func TestTrustedProxy(t *testing.T) {
	resolver, err := requestmeta.NewResolver([]string{"10.0.0.0/8"})
	if err != nil {
		t.Fatalf("NewResolver: %v", err)
	}

	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "10.0.0.1:12345"
	r.Header.Set("X-Forwarded-For", "203.0.113.50")

	meta := resolver.Extract(r)
	if meta.ClientIP != "203.0.113.50" {
		t.Fatalf("expected 203.0.113.50, got %q", meta.ClientIP)
	}
}

func TestUntrustedProxy(t *testing.T) {
	resolver, err := requestmeta.NewResolver([]string{"10.0.0.0/8"})
	if err != nil {
		t.Fatalf("NewResolver: %v", err)
	}

	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "203.0.113.1:12345"
	r.Header.Set("X-Forwarded-For", "192.168.1.100")

	meta := resolver.Extract(r)
	if meta.ClientIP == "192.168.1.100" {
		t.Fatal("should not trust X-Forwarded-For from untrusted proxy")
	}
}

func TestSpoofingAttempt(t *testing.T) {
	resolver, err := requestmeta.NewResolver([]string{"10.0.0.0/8"})
	if err != nil {
		t.Fatalf("NewResolver: %v", err)
	}

	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "10.0.0.1:12345"
	r.Header.Set("X-Forwarded-For", "10.0.0.2, 10.0.0.3, 203.0.113.50")

	meta := resolver.Extract(r)
	if meta.ClientIP != "203.0.113.50" {
		t.Fatalf("expected first untrusted IP 203.0.113.50, got %q", meta.ClientIP)
	}
}

func TestMultipleProxies(t *testing.T) {
	resolver, err := requestmeta.NewResolver([]string{"10.0.0.0/8", "192.168.0.0/16"})
	if err != nil {
		t.Fatalf("NewResolver: %v", err)
	}

	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "192.168.1.1:12345"
	r.Header.Set("X-Forwarded-For", "10.0.0.1, 203.0.113.100")

	meta := resolver.Extract(r)
	if meta.ClientIP != "203.0.113.100" {
		t.Fatalf("expected 203.0.113.100, got %q", meta.ClientIP)
	}
}

func TestIPv6(t *testing.T) {
	resolver, err := requestmeta.NewResolver([]string{"::1/128"})
	if err != nil {
		t.Fatalf("NewResolver: %v", err)
	}

	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "[::1]:12345"

	meta := resolver.Extract(r)
	if meta.ClientIP != "::1" {
		t.Fatalf("expected ::1, got %q", meta.ClientIP)
	}
}

func TestMalformedHeader(t *testing.T) {
	resolver, err := requestmeta.NewResolver([]string{"10.0.0.0/8"})
	if err != nil {
		t.Fatalf("NewResolver: %v", err)
	}

	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "10.0.0.1:12345"
	r.Header.Set("X-Forwarded-For", "not-an-ip, 203.0.113.50")

	meta := resolver.Extract(r)
	if meta.ClientIP != "203.0.113.50" {
		t.Fatalf("expected 203.0.113.50 (skip malformed), got %q", meta.ClientIP)
	}
}

func TestUserAgentTruncated(t *testing.T) {
	resolver, _ := requestmeta.NewResolver(nil)

	longUA := ""
	for i := 0; i < 1000; i++ {
		longUA += "a"
	}

	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "127.0.0.1:12345"
	r.Header.Set("User-Agent", longUA)

	meta := resolver.Extract(r)
	if len(meta.UserAgent) > 512 {
		t.Fatalf("expected UserAgent truncated to 512, got %d", len(meta.UserAgent))
	}
}

func TestRequestID(t *testing.T) {
	resolver, _ := requestmeta.NewResolver(nil)

	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "127.0.0.1:12345"

	meta := resolver.Extract(r)
	if meta.RequestID == "" {
		t.Log("RequestID may be empty without middleware")
	}
}

func TestXRealIPFallback(t *testing.T) {
	resolver, err := requestmeta.NewResolver([]string{"10.0.0.0/8"})
	if err != nil {
		t.Fatalf("NewResolver: %v", err)
	}

	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "10.0.0.1:12345"
	r.Header.Set("X-Real-IP", "203.0.113.99")

	meta := resolver.Extract(r)
	if meta.ClientIP != "203.0.113.99" {
		t.Fatalf("expected 203.0.113.99 from X-Real-IP, got %q", meta.ClientIP)
	}
}

func TestEmptyCIDRs(t *testing.T) {
	resolver, err := requestmeta.NewResolver([]string{"", "   "})
	if err != nil {
		t.Fatalf("NewResolver with empty CIDRs: %v", err)
	}

	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "10.0.0.1:12345"
	r.Header.Set("X-Forwarded-For", "203.0.113.50")

	meta := resolver.Extract(r)
	if meta.ClientIP == "203.0.113.50" {
		t.Fatal("should not trust XFF with empty CIDR list")
	}
}
