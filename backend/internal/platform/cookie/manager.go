package cookie

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

const hostPrefix = "__Host-"

type Config struct {
	Secure      bool
	SameSite    string
	RefreshName string
	CSRFName    string
	Env         string
}

func (c Config) Validate() error {
	if c.Env == "production" {
		if !c.Secure {
			return fmt.Errorf("cookie secure must be true in production")
		}
		if !strings.HasPrefix(c.RefreshName, hostPrefix) {
			return fmt.Errorf("refresh cookie name must use __Host- prefix in production")
		}
		if !strings.HasPrefix(c.CSRFName, hostPrefix) {
			return fmt.Errorf("csrf cookie name must use __Host- prefix in production")
		}
	}

	if strings.HasPrefix(c.RefreshName, hostPrefix) || strings.HasPrefix(c.CSRFName, hostPrefix) {
		if !c.Secure {
			return fmt.Errorf("__Host- prefix requires Secure flag")
		}
	}

	switch strings.ToLower(c.SameSite) {
	case "lax", "strict", "none":
	default:
		return fmt.Errorf("invalid SameSite value: %q", c.SameSite)
	}

	if c.SameSiteMode() == http.SameSiteNoneMode && !c.Secure {
		return fmt.Errorf("SameSite=None requires Secure flag")
	}

	return nil
}

func (c Config) SameSiteMode() http.SameSite {
	switch strings.ToLower(c.SameSite) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}

type Manager struct {
	config Config
}

func NewManager(cfg Config) (*Manager, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &Manager{config: cfg}, nil
}

func (m *Manager) SetRefreshCookie(w http.ResponseWriter, value string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     m.config.RefreshName,
		Value:    value,
		Path:     "/",
		Domain:   "",
		Expires:  expiresAt,
		MaxAge:   int(time.Until(expiresAt).Seconds()),
		Secure:   m.config.Secure,
		HttpOnly: true,
		SameSite: m.config.SameSiteMode(),
	})
}

func (m *Manager) SetCSRFCookie(w http.ResponseWriter, value string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     m.config.CSRFName,
		Value:    value,
		Path:     "/",
		Domain:   "",
		Expires:  expiresAt,
		MaxAge:   int(time.Until(expiresAt).Seconds()),
		Secure:   m.config.Secure,
		HttpOnly: false,
		SameSite: m.config.SameSiteMode(),
	})
}

func (m *Manager) ClearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     m.config.RefreshName,
		Value:    "",
		Path:     "/",
		Domain:   "",
		MaxAge:   -1,
		Secure:   m.config.Secure,
		HttpOnly: true,
		SameSite: m.config.SameSiteMode(),
	})
}

func (m *Manager) ClearCSRFCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     m.config.CSRFName,
		Value:    "",
		Path:     "/",
		Domain:   "",
		MaxAge:   -1,
		Secure:   m.config.Secure,
		HttpOnly: false,
		SameSite: m.config.SameSiteMode(),
	})
}

func (m *Manager) ClearAuthCookies(w http.ResponseWriter) {
	m.ClearRefreshCookie(w)
	m.ClearCSRFCookie(w)
}

func (m *Manager) RefreshCookieName() string {
	return m.config.RefreshName
}

func (m *Manager) CSRFCookieName() string {
	return m.config.CSRFName
}

func (m *Manager) IsSecure() bool {
	return m.config.Secure
}
