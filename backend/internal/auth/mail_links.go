package auth

import (
	"fmt"
	"net/url"
	"strings"
)

const (
	emailVerificationPath = "/verify-email"
	passwordResetPath     = "/reset-password"
)

type MailLinkBuilder struct {
	baseURL *url.URL
}

func NewMailLinkBuilder(appPublicURL string) (*MailLinkBuilder, error) {
	if appPublicURL == "" || strings.TrimSpace(appPublicURL) != appPublicURL {
		return nil, fmt.Errorf("APP_PUBLIC_URL must be a non-empty absolute HTTP(S) URL")
	}

	baseURL, err := url.Parse(appPublicURL)
	if err != nil {
		return nil, fmt.Errorf("parse APP_PUBLIC_URL: %w", err)
	}
	if (baseURL.Scheme != "http" && baseURL.Scheme != "https") || baseURL.Host == "" || baseURL.Hostname() == "" || baseURL.Opaque != "" {
		return nil, fmt.Errorf("APP_PUBLIC_URL must be an absolute HTTP(S) URL")
	}
	if baseURL.User != nil {
		return nil, fmt.Errorf("APP_PUBLIC_URL must not contain user information")
	}

	baseURL.RawQuery = ""
	baseURL.ForceQuery = false
	baseURL.Fragment = ""
	baseURL.RawFragment = ""

	return &MailLinkBuilder{baseURL: baseURL}, nil
}

func (b *MailLinkBuilder) BuildEmailVerification(token string) string {
	return b.build(emailVerificationPath, token)
}

func (b *MailLinkBuilder) BuildPasswordReset(token string) string {
	return b.build(passwordResetPath, token)
}

func (b *MailLinkBuilder) build(route, token string) string {
	link := *b.baseURL
	escapedBasePath := link.EscapedPath()
	if link.RawPath != "" {
		link.RawPath = strings.TrimRight(escapedBasePath, "/") + route
		link.Path, _ = url.PathUnescape(link.RawPath)
	} else {
		link.Path = strings.TrimRight(link.Path, "/") + route
	}
	link.RawQuery = ""
	link.ForceQuery = false
	link.Fragment = url.Values{"token": []string{token}}.Encode()
	link.RawFragment = ""
	return link.String()
}
