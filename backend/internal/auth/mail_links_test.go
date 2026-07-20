package auth

import (
	"net/url"
	"strings"
	"testing"
)

func TestMailLinkBuilderBuildsPurposeRoutes(t *testing.T) {
	builder, err := NewMailLinkBuilder("https://app.example.com/tenant/base/")
	if err != nil {
		t.Fatalf("NewMailLinkBuilder() error = %v", err)
	}

	for _, test := range []struct {
		name string
		got  string
		path string
	}{
		{name: "email verification", got: builder.BuildEmailVerification("evt_selector.secret"), path: "/tenant/base/verify-email"},
		{name: "password reset", got: builder.BuildPasswordReset("prt_selector.secret"), path: "/tenant/base/reset-password"},
	} {
		t.Run(test.name, func(t *testing.T) {
			parsed, err := url.Parse(test.got)
			if err != nil {
				t.Fatalf("parse generated link: %v", err)
			}
			if parsed.Scheme != "https" || parsed.Host != "app.example.com" || parsed.Path != test.path {
				t.Fatalf("generated URL = %q", test.got)
			}
			if parsed.RawQuery != "" {
				t.Fatalf("token leaked into query: %q", parsed.RawQuery)
			}
			fragment, err := url.ParseQuery(parsed.Fragment)
			if err != nil {
				t.Fatalf("parse fragment: %v", err)
			}
			if fragment.Get("token") == "" {
				t.Fatalf("fragment has no token: %q", parsed.Fragment)
			}
			if strings.Contains(parsed.Path, fragment.Get("token")) {
				t.Fatalf("token leaked into path: %q", parsed.Path)
			}
		})
	}
}

func TestMailLinkBuilderEncodesTokenOnlyInFragment(t *testing.T) {
	builder, err := NewMailLinkBuilder("http://localhost:5173/app?old=secret#old-fragment")
	if err != nil {
		t.Fatalf("NewMailLinkBuilder() error = %v", err)
	}
	token := "evt_value&admin=true /?#%"
	link := builder.BuildEmailVerification(token)

	parsed, err := url.Parse(link)
	if err != nil {
		t.Fatalf("parse generated link: %v", err)
	}
	if parsed.Path != "/app/verify-email" || parsed.RawQuery != "" {
		t.Fatalf("generated URL path/query = %q / %q", parsed.Path, parsed.RawQuery)
	}
	if strings.Contains(parsed.Path, token) || strings.Contains(parsed.RawQuery, token) {
		t.Fatal("token leaked outside fragment")
	}
	fragment, err := url.ParseQuery(parsed.Fragment)
	if err != nil {
		t.Fatalf("parse fragment: %v", err)
	}
	if got := fragment.Get("token"); got != token {
		t.Fatalf("fragment token = %q, want %q", got, token)
	}
	if strings.Contains(parsed.Fragment, "&admin=true") {
		t.Fatalf("fragment token was not URL-encoded: %q", parsed.Fragment)
	}
}

func TestMailLinkBuilderPreservesEscapedBasePath(t *testing.T) {
	for _, test := range []struct {
		publicURL string
		wantPath  string
	}{
		{publicURL: "https://app.example.com/companies/acme%2Fbr/", wantPath: "/companies/acme%2Fbr/reset-password"},
		{publicURL: "https://app.example.com/companies/acme%2F", wantPath: "/companies/acme%2F/reset-password"},
	} {
		builder, err := NewMailLinkBuilder(test.publicURL)
		if err != nil {
			t.Fatalf("NewMailLinkBuilder() error = %v", err)
		}

		parsed, err := url.Parse(builder.BuildPasswordReset("prt_token"))
		if err != nil {
			t.Fatalf("parse generated link: %v", err)
		}
		if got := parsed.EscapedPath(); got != test.wantPath {
			t.Fatalf("escaped path = %q, want %q", got, test.wantPath)
		}
	}
}

func TestNewMailLinkBuilderRejectsInvalidPublicURLs(t *testing.T) {
	invalid := []string{
		"",
		" app.example.com ",
		"app.example.com",
		"/relative/path",
		"//app.example.com/path",
		"ftp://app.example.com/path",
		"https:///missing-host",
		"https://user:password@app.example.com",
		"https://",
		"https://app.example.com/%zz",
	}

	for _, publicURL := range invalid {
		t.Run(publicURL, func(t *testing.T) {
			if _, err := NewMailLinkBuilder(publicURL); err == nil {
				t.Fatalf("NewMailLinkBuilder(%q) succeeded", publicURL)
			}
		})
	}
}

func TestMailLinkBuilderAcceptsHTTPAndHTTPS(t *testing.T) {
	for _, publicURL := range []string{"http://localhost:3000", "https://app.example.com"} {
		if _, err := NewMailLinkBuilder(publicURL); err != nil {
			t.Fatalf("NewMailLinkBuilder(%q) error = %v", publicURL, err)
		}
	}
}
