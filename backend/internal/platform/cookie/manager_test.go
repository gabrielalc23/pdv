package cookie_test

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gabrielalc23/pdv/internal/platform/cookie"
)

func TestProductionConfig(t *testing.T) {
	cfg := cookie.Config{
		Secure:      true,
		SameSite:    "Lax",
		RefreshName: "__Host-pdv_refresh",
		CSRFName:    "__Host-pdv_csrf",
		Env:         "production",
	}

	mgr, err := cookie.NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	if mgr == nil {
		t.Fatal("expected non-nil manager")
	}
}

func TestDevelopmentConfig(t *testing.T) {
	cfg := cookie.Config{
		Secure:      false,
		SameSite:    "Lax",
		RefreshName: "pdv_refresh",
		CSRFName:    "pdv_csrf",
		Env:         "development",
	}

	mgr, err := cookie.NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	if mgr == nil {
		t.Fatal("expected non-nil manager")
	}
}

func TestProductionRejectsInsecure(t *testing.T) {
	cfg := cookie.Config{
		Secure:      false,
		SameSite:    "Lax",
		RefreshName: "__Host-pdv_refresh",
		CSRFName:    "__Host-pdv_csrf",
		Env:         "production",
	}

	_, err := cookie.NewManager(cfg)
	if err == nil {
		t.Fatal("expected error for insecure production config")
	}
}

func TestProductionRequiresHostPrefix(t *testing.T) {
	cfg := cookie.Config{
		Secure:      true,
		SameSite:    "Lax",
		RefreshName: "pdv_refresh",
		CSRFName:    "__Host-pdv_csrf",
		Env:         "production",
	}

	_, err := cookie.NewManager(cfg)
	if err == nil {
		t.Fatal("expected error for missing __Host- prefix in production")
	}
}

func TestSetRefreshCookie(t *testing.T) {
	mgr, _ := cookie.NewManager(cookie.Config{
		Secure:      true,
		SameSite:    "Lax",
		RefreshName: "__Host-pdv_refresh",
		CSRFName:    "__Host-pdv_csrf",
		Env:         "production",
	})

	w := httptest.NewRecorder()
	mgr.SetRefreshCookie(w, "test-value", time.Now().Add(time.Hour))

	cookies := w.Result().Cookies()
	var found bool
	for _, c := range cookies {
		if c.Name == "__Host-pdv_refresh" {
			found = true
			if !c.HttpOnly {
				t.Fatal("refresh cookie should be HttpOnly")
			}
			if !c.Secure {
				t.Fatal("refresh cookie should be Secure")
			}
			if c.Domain != "" {
				t.Fatal("refresh cookie should not have Domain")
			}
			if c.Path != "/" {
				t.Fatalf("expected path /, got %q", c.Path)
			}
		}
	}
	if !found {
		t.Fatal("refresh cookie not found")
	}
}

func TestSetCSRFCookie(t *testing.T) {
	mgr, _ := cookie.NewManager(cookie.Config{
		Secure:      true,
		SameSite:    "Lax",
		RefreshName: "__Host-pdv_refresh",
		CSRFName:    "__Host-pdv_csrf",
		Env:         "production",
	})

	w := httptest.NewRecorder()
	mgr.SetCSRFCookie(w, "test-csrf", time.Now().Add(time.Hour))

	cookies := w.Result().Cookies()
	var found bool
	for _, c := range cookies {
		if c.Name == "__Host-pdv_csrf" {
			found = true
			if c.HttpOnly {
				t.Fatal("CSRF cookie should not be HttpOnly")
			}
			if !c.Secure {
				t.Fatal("CSRF cookie should be Secure")
			}
			if c.Domain != "" {
				t.Fatal("CSRF cookie should not have Domain")
			}
		}
	}
	if !found {
		t.Fatal("CSRF cookie not found")
	}
}

func TestClearAuthCookies(t *testing.T) {
	mgr, _ := cookie.NewManager(cookie.Config{
		Secure:      true,
		SameSite:    "Lax",
		RefreshName: "__Host-pdv_refresh",
		CSRFName:    "__Host-pdv_csrf",
		Env:         "production",
	})

	w := httptest.NewRecorder()
	mgr.ClearAuthCookies(w)

	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected cookies to be cleared")
	}
	for _, c := range cookies {
		if c.MaxAge != -1 {
			t.Fatalf("expected maxAge -1 for removal, got %d for %s", c.MaxAge, c.Name)
		}
	}
}

func TestInvalidSameSite(t *testing.T) {
	cfg := cookie.Config{
		Secure:      true,
		SameSite:    "Invalid",
		RefreshName: "__Host-pdv_refresh",
		CSRFName:    "__Host-pdv_csrf",
		Env:         "production",
	}

	_, err := cookie.NewManager(cfg)
	if err == nil {
		t.Fatal("expected error for invalid SameSite")
	}
}
