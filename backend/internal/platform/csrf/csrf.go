package csrf

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
)

var (
	ErrTokenMissing     = errors.New("csrf token is missing")
	ErrTokenMalformed   = errors.New("csrf token is malformed")
	ErrTokenInvalid     = errors.New("csrf token is invalid")
	ErrTokenBinding     = errors.New("csrf token binding mismatch")
	ErrHMACInvalid      = errors.New("csrf token hmac is invalid")
	ErrOriginDenied     = errors.New("origin not allowed")
	ErrCrossSiteRequest = errors.New("cross-site request rejected")
)

const (
	nonceSize     = 32
	headerCSRF    = "X-CSRF-Token"
	minSecretSize = 32
)

type Binding string

const (
	BindingPreauth   Binding = "preauth"
	BindingSessionID Binding = "session-id"
)

type Manager struct {
	secret         []byte
	allowedOrigins []string
}

func NewManager(secret []byte, allowedOrigins []string) (*Manager, error) {
	if len(secret) < minSecretSize {
		return nil, fmt.Errorf("csrf secret must be at least %d bytes", minSecretSize)
	}
	return &Manager{secret: secret, allowedOrigins: allowedOrigins}, nil
}

func (m *Manager) Generate(binding Binding, bindingValue string) (string, error) {
	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	nonceB64 := base64.RawURLEncoding.EncodeToString(nonce)

	mac := hmac.New(sha256.New, m.secret)
	mac.Write([]byte(binding))
	mac.Write([]byte("|"))
	mac.Write([]byte(bindingValue))
	mac.Write([]byte("|"))
	mac.Write(nonce)
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return nonceB64 + "." + sig, nil
}

func (m *Manager) Validate(cookieToken, headerToken, binding string, bindingValue string) error {
	if cookieToken == "" {
		return ErrTokenMissing
	}
	if headerToken == "" {
		return ErrTokenMissing
	}
	if !hmac.Equal([]byte(cookieToken), []byte(headerToken)) {
		return ErrTokenInvalid
	}

	parts := strings.SplitN(cookieToken, ".", 2)
	if len(parts) != 2 {
		return fmt.Errorf("%w: expected nonce.sig format", ErrTokenMalformed)
	}
	nonceB64, sigB64 := parts[0], parts[1]

	nonce, err := base64.RawURLEncoding.DecodeString(nonceB64)
	if err != nil || len(nonce) != nonceSize {
		return fmt.Errorf("%w: invalid nonce encoding", ErrTokenMalformed)
	}

	sig, err := base64.RawURLEncoding.DecodeString(sigB64)
	if err != nil || len(sig) == 0 {
		return fmt.Errorf("%w: invalid signature encoding", ErrTokenMalformed)
	}

	mac := hmac.New(sha256.New, m.secret)
	mac.Write([]byte(binding))
	mac.Write([]byte("|"))
	mac.Write([]byte(bindingValue))
	mac.Write([]byte("|"))
	mac.Write(nonce)
	expected := mac.Sum(nil)

	if !hmac.Equal(sig, expected) {
		return ErrHMACInvalid
	}

	if len(bindingValue) > 0 && binding == "session-id" {
		sigMac := hmac.New(sha256.New, m.secret)
		sigMac.Write([]byte(binding))
		sigMac.Write([]byte("|"))
		sigMac.Write([]byte(bindingValue))
		sigMac.Write([]byte("|"))
		sigMac.Write(nonce)
		expectedSig := sigMac.Sum(nil)

		if !hmac.Equal(sig, expectedSig) {
			return ErrTokenBinding
		}
	}

	return nil
}

func (m *Manager) ValidateRequest(r *http.Request, cookieValue string) error {
	headerToken := r.Header.Get(headerCSRF)
	return m.Validate(cookieValue, headerToken, string(BindingPreauth), "preauth")
}

func (m *Manager) ValidateRequestWithSession(r *http.Request, cookieValue, sessionID string) error {
	headerToken := r.Header.Get(headerCSRF)
	return m.Validate(cookieValue, headerToken, string(BindingSessionID), sessionID)
}

func (m *Manager) CheckOrigin(r *http.Request) error {
	origin := r.Header.Get("Origin")

	if origin == "" {
		referer := r.Header.Get("Referer")
		if referer == "" {
			return fmt.Errorf("%w: origin or referer is required", ErrOriginDenied)
		}
		origin = extractOriginFromReferer(referer)
	}

	if origin == "" {
		return fmt.Errorf("%w: invalid referer", ErrOriginDenied)
	}

	if slices.Contains(m.allowedOrigins, origin) {
		return nil
	}

	return fmt.Errorf("%w: %q", ErrOriginDenied, origin)
}

func (m *Manager) CheckFetchMetadata(r *http.Request) error {
	site := r.Header.Get("Sec-Fetch-Site")
	if site == "cross-site" {
		return ErrCrossSiteRequest
	}
	return nil
}

func IsUnsafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return false
	}
	return true
}

func extractOriginFromReferer(referer string) string {
	ref := strings.TrimSuffix(referer, "/")
	if idx := strings.Index(ref, "://"); idx >= 0 {
		rest := ref[idx+3:]
		if slashIdx := strings.Index(rest, "/"); slashIdx >= 0 {
			return ref[:idx+3+slashIdx]
		}
		return ref
	}
	return ""
}

func NonceSize() int {
	return nonceSize
}

func MinSecretSize() int {
	return minSecretSize
}

func HeaderName() string {
	return headerCSRF
}
