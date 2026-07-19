package config_test

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/gabrielalc23/pdv/config"
)

func setProductionSecrets(t *testing.T) {
	t.Helper()
	secret := base64.StdEncoding.EncodeToString(make([]byte, 32))
	t.Setenv("AUTH_TOKEN_HASH_KEY", secret)
	t.Setenv("AUTH_CSRF_SECRET", secret)
	t.Setenv("RATE_LIMIT_KEY_SECRET", secret)
}

func setProductionMailer(t *testing.T) {
	t.Helper()
	t.Setenv("MAIL_DRIVER", "smtp")
	t.Setenv("MAIL_FROM", "noreply@example.com")
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_PORT", "587")
	t.Setenv("SMTP_USERNAME", "")
	t.Setenv("SMTP_PASSWORD", "")
	t.Setenv("SMTP_STARTTLS", "true")
	t.Setenv("APP_PUBLIC_URL", "https://app.example.com")
}

func TestLoadDefaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/pdv?sslmode=disable")
	t.Setenv("APP_ENV", "development")
	t.Setenv("HTTP_ADDRESS", "")
	t.Setenv("AUTH_ALLOW_EPHEMERAL_DEV_KEY", "true")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.Address != ":8080" {
		t.Fatalf("expected default address :8080, got %q", cfg.Address)
	}
	if cfg.AppEnv != "development" {
		t.Fatalf("expected APP_ENV development, got %q", cfg.AppEnv)
	}
	if cfg.DatabaseURL == "" {
		t.Fatal("expected database url to be set")
	}
	if cfg.ValkeyAddr != "localhost:6379" {
		t.Fatalf("expected VALKEY_ADDR localhost:6379, got %q", cfg.ValkeyAddr)
	}
	if cfg.ValkeyDB != 0 {
		t.Fatalf("expected VALKEY_DB 0, got %d", cfg.ValkeyDB)
	}
	if cfg.JWTIssuer != "pdv-auth" {
		t.Fatalf("expected JWT_ISSUER pdv-auth, got %q", cfg.JWTIssuer)
	}
	if cfg.JWTAudience != "pdv-api" {
		t.Fatalf("expected JWT_AUDIENCE pdv-api, got %q", cfg.JWTAudience)
	}
	if cfg.AccessTokenTTL != 5*time.Minute {
		t.Fatalf("expected ACCESS_TOKEN_TTL 5m, got %s", cfg.AccessTokenTTL)
	}
	if cfg.RefreshIdleTTL != 720*time.Hour {
		t.Fatalf("expected REFRESH_IDLE_TTL 720h, got %s", cfg.RefreshIdleTTL)
	}
	if cfg.SessionAbsoluteTTL != 2160*time.Hour {
		t.Fatalf("expected SESSION_ABSOLUTE_TTL 2160h, got %s", cfg.SessionAbsoluteTTL)
	}
	if cfg.JWTClockSkew != 30*time.Second {
		t.Fatalf("expected JWT_CLOCK_SKEW 30s, got %s", cfg.JWTClockSkew)
	}
	if !cfg.AuthRegistrationEnabled {
		t.Fatal("expected AUTH_REGISTRATION_ENABLED true by default")
	}
	if cfg.AuthRequireVerifiedEmail {
		t.Fatal("expected AUTH_REQUIRE_VERIFIED_EMAIL false by default")
	}
	if !cfg.AuthAllowEphemeralDevKey {
		t.Fatal("expected AUTH_ALLOW_EPHEMERAL_DEV_KEY true by default in development")
	}
	if !cfg.AuthTenantCreationEnabled {
		t.Fatal("expected AUTH_TENANT_CREATION_ENABLED true by default")
	}
	if cfg.CookieSameSite != "Lax" {
		t.Fatalf("expected COOKIE_SAME_SITE Lax, got %q", cfg.CookieSameSite)
	}
	if cfg.CookieRefreshName != "pdv_refresh" {
		t.Fatalf("expected dev cookie refresh name pdv_refresh, got %q", cfg.CookieRefreshName)
	}
	if cfg.CookieCSRFName != "pdv_csrf" {
		t.Fatalf("expected dev cookie CSRF name pdv_csrf, got %q", cfg.CookieCSRFName)
	}
	if cfg.PasswordArgon2MemoryKiB != 65536 {
		t.Fatalf("expected PASSWORD_ARGON2_MEMORY_KIB 65536, got %d", cfg.PasswordArgon2MemoryKiB)
	}
	if cfg.PasswordArgon2Iterations != 3 {
		t.Fatalf("expected PASSWORD_ARGON2_ITERATIONS 3, got %d", cfg.PasswordArgon2Iterations)
	}
	if cfg.PasswordArgon2Parallelism != 1 {
		t.Fatalf("expected PASSWORD_ARGON2_PARALLELISM 1, got %d", cfg.PasswordArgon2Parallelism)
	}
	if cfg.MailDriver != "log" {
		t.Fatalf("expected MAIL_DRIVER log, got %q", cfg.MailDriver)
	}
	if cfg.MailFrom != "noreply@pdv.local" {
		t.Fatalf("expected MAIL_FROM noreply@pdv.local, got %q", cfg.MailFrom)
	}
}

func TestLoadRequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	if _, err := config.Load(); err == nil {
		t.Fatal("expected error when DATABASE_URL is missing")
	}
}

func TestLoadValidAppEnv(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/pdv?sslmode=disable")
	t.Setenv("AUTH_ALLOW_EPHEMERAL_DEV_KEY", "true")
	setProductionSecrets(t)
	for _, env := range []string{"development", "test", "production"} {
		t.Setenv("APP_ENV", env)
		if env == "production" {
			setProductionMailer(t)
		}
		if _, err := config.Load(); err != nil {
			t.Fatalf("expected valid APP_ENV %q: %v", env, err)
		}
	}
}

func TestLoadInvalidAppEnv(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/pdv?sslmode=disable")
	t.Setenv("APP_ENV", "staging")
	if _, err := config.Load(); err == nil {
		t.Fatal("expected error for invalid APP_ENV")
	}
}

func TestLoadAccessTokenTTLValidation(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/pdv?sslmode=disable")
	t.Setenv("AUTH_ALLOW_EPHEMERAL_DEV_KEY", "true")

	t.Run("zero", func(t *testing.T) {
		t.Setenv("ACCESS_TOKEN_TTL", "0s")
		if _, err := config.Load(); err == nil {
			t.Fatal("expected error for zero TTL")
		}
	})

	t.Run("too_long", func(t *testing.T) {
		t.Setenv("ACCESS_TOKEN_TTL", "30m")
		if _, err := config.Load(); err == nil {
			t.Fatal("expected error for TTL > 15m")
		}
	})

	t.Run("valid", func(t *testing.T) {
		t.Setenv("ACCESS_TOKEN_TTL", "10m")
		if _, err := config.Load(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestLoadTTLCoherence(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/pdv?sslmode=disable")
	t.Setenv("REFRESH_IDLE_TTL", "100h")
	t.Setenv("SESSION_ABSOLUTE_TTL", "50h")

	if _, err := config.Load(); err == nil {
		t.Fatal("expected error when REFRESH_IDLE_TTL > SESSION_ABSOLUTE_TTL")
	}
}

func TestLoadProductionCookieSecure(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/pdv?sslmode=disable")
	t.Setenv("APP_ENV", "production")
	t.Setenv("COOKIE_SECURE", "false")

	if _, err := config.Load(); err == nil {
		t.Fatal("expected error when COOKIE_SECURE is false in production")
	}
}

func TestLoadProductionCookieNames(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/pdv?sslmode=disable")
	t.Setenv("APP_ENV", "production")
	t.Setenv("AUTH_ALLOW_EPHEMERAL_DEV_KEY", "true")
	setProductionSecrets(t)
	setProductionMailer(t)
	t.Setenv("COOKIE_REFRESH_NAME", "")
	t.Setenv("COOKIE_CSRF_NAME", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.CookieRefreshName != "__Host-pdv_refresh" {
		t.Fatalf("expected __Host-pdv_refresh in production, got %q", cfg.CookieRefreshName)
	}
	if cfg.CookieCSRFName != "__Host-pdv_csrf" {
		t.Fatalf("expected __Host-pdv_csrf in production, got %q", cfg.CookieCSRFName)
	}
}

func TestLoadProductionRequiresJWTKey(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/pdv?sslmode=disable")
	t.Setenv("APP_ENV", "production")
	t.Setenv("JWT_PRIVATE_KEY_PATH", "")
	t.Setenv("AUTH_ALLOW_EPHEMERAL_DEV_KEY", "false")

	if _, err := config.Load(); err == nil {
		t.Fatal("expected error when JWT_PRIVATE_KEY_PATH is empty in production")
	}
}

func TestLoadProductionAllowsEphemeralKey(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/pdv?sslmode=disable")
	t.Setenv("APP_ENV", "production")
	t.Setenv("JWT_PRIVATE_KEY_PATH", "")
	t.Setenv("AUTH_ALLOW_EPHEMERAL_DEV_KEY", "true")
	setProductionSecrets(t)
	setProductionMailer(t)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected ephemeral dev key to be allowed: %v", err)
	}
	if !cfg.AuthAllowEphemeralDevKey {
		t.Fatal("expected AUTH_ALLOW_EPHEMERAL_DEV_KEY to be true")
	}
}

func TestLoadProductionMailConfiguration(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/pdv?sslmode=disable")
	t.Setenv("APP_ENV", "production")
	t.Setenv("AUTH_ALLOW_EPHEMERAL_DEV_KEY", "true")
	setProductionSecrets(t)
	setProductionMailer(t)

	t.Run("log_driver_rejected", func(t *testing.T) {
		t.Setenv("MAIL_DRIVER", "log")
		if _, err := config.Load(); err == nil {
			t.Fatal("expected MAIL_DRIVER=log to be rejected in production")
		}
	})

	t.Run("smtp_valid", func(t *testing.T) {
		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("expected valid production smtp configuration: %v", err)
		}
		if cfg.MailDriver != "smtp" {
			t.Fatalf("expected MAIL_DRIVER smtp, got %q", cfg.MailDriver)
		}
		if cfg.SMTPHost != "smtp.example.com" || cfg.SMTPPort != 587 {
			t.Fatalf("expected smtp.example.com:587, got %s:%d", cfg.SMTPHost, cfg.SMTPPort)
		}
		if !cfg.SMTPStartTLS {
			t.Fatal("expected SMTP_STARTTLS true")
		}
	})

	t.Run("starttls_required", func(t *testing.T) {
		t.Setenv("SMTP_STARTTLS", "false")
		if _, err := config.Load(); err == nil {
			t.Fatal("expected production SMTP without STARTTLS to be rejected")
		}
	})

	t.Run("https_public_url_required", func(t *testing.T) {
		t.Setenv("APP_PUBLIC_URL", "http://app.example.com")
		if _, err := config.Load(); err == nil {
			t.Fatal("expected production HTTP APP_PUBLIC_URL to be rejected")
		}
	})
}

func TestLoadSMTPValidation(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/pdv?sslmode=disable")
	t.Setenv("APP_ENV", "development")
	t.Setenv("AUTH_ALLOW_EPHEMERAL_DEV_KEY", "true")
	setProductionMailer(t)

	tests := []struct {
		name  string
		key   string
		value string
	}{
		{name: "missing_host", key: "SMTP_HOST", value: ""},
		{name: "zero_port", key: "SMTP_PORT", value: "0"},
		{name: "port_above_maximum", key: "SMTP_PORT", value: "65536"},
		{name: "invalid_mail_from", key: "MAIL_FROM", value: "noreply@"},
		{name: "password_without_username", key: "SMTP_PASSWORD", value: "secret"},
		{name: "invalid_starttls_text", key: "SMTP_STARTTLS", value: "sometimes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(tt.key, tt.value)
			if _, err := config.Load(); err == nil {
				t.Fatalf("expected error for %s=%q", tt.key, tt.value)
			}
		})
	}
}

func TestLoadProductionWildcardCORS(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/pdv?sslmode=disable")
	t.Setenv("APP_ENV", "production")
	t.Setenv("CORS_ALLOWED_ORIGINS", "*")

	if _, err := config.Load(); err == nil {
		t.Fatal("expected error when CORS_ALLOWED_ORIGINS contains wildcard in production")
	}
}

func TestLoadInvalidCIDRs(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/pdv?sslmode=disable")
	t.Setenv("TRUSTED_PROXY_CIDRS", "not-a-cidr,10.0.0.0/8")

	if _, err := config.Load(); err == nil {
		t.Fatal("expected error for invalid CIDR")
	}
}

func TestLoadValidCIDRs(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/pdv?sslmode=disable")
	t.Setenv("AUTH_ALLOW_EPHEMERAL_DEV_KEY", "true")
	t.Setenv("TRUSTED_PROXY_CIDRS", "10.0.0.0/8,192.168.0.0/16")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.TrustedProxyCIDRs) != 2 {
		t.Fatalf("expected 2 CIDRs, got %d", len(cfg.TrustedProxyCIDRs))
	}
	if cfg.TrustedProxyCIDRs[0] != "10.0.0.0/8" {
		t.Fatalf("expected 10.0.0.0/8, got %q", cfg.TrustedProxyCIDRs[0])
	}
}

func TestLoadInvalidSameSite(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/pdv?sslmode=disable")
	t.Setenv("COOKIE_SAME_SITE", "Invalid")

	if _, err := config.Load(); err == nil {
		t.Fatal("expected error for invalid SameSite")
	}
}

func TestLoadValidSameSite(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/pdv?sslmode=disable")
	t.Setenv("AUTH_ALLOW_EPHEMERAL_DEV_KEY", "true")

	for _, val := range []string{"Lax", "Strict", "None"} {
		t.Setenv("COOKIE_SAME_SITE", val)
		if _, err := config.Load(); err != nil {
			t.Fatalf("expected valid SameSite %q: %v", val, err)
		}
	}
}

func TestLoadAppPublicURLValidation(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/pdv?sslmode=disable")
	t.Setenv("AUTH_ALLOW_EPHEMERAL_DEV_KEY", "true")

	for _, tt := range []struct {
		name  string
		value string
	}{
		{name: "relative", value: "/pdv"},
		{name: "unsupported_scheme", value: "ftp://app.example.com"},
		{name: "userinfo", value: "https://user:pass@app.example.com"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("APP_PUBLIC_URL", tt.value)
			if _, err := config.Load(); err == nil {
				t.Fatalf("expected error for APP_PUBLIC_URL=%q", tt.value)
			}
		})
	}

	t.Run("absolute_base_path", func(t *testing.T) {
		t.Setenv("APP_PUBLIC_URL", "https://app.example.com/pdv")
		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.AppPublicURL != "https://app.example.com/pdv" {
			t.Fatalf("expected https://app.example.com/pdv, got %q", cfg.AppPublicURL)
		}
	})
}

func TestLoadSecretDecoding(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/pdv?sslmode=disable")
	t.Setenv("AUTH_ALLOW_EPHEMERAL_DEV_KEY", "true")

	t.Run("base64_32_bytes", func(t *testing.T) {
		key := "QUJDREVGR0hJSktMTU5PUFFSU1RVVldYWVoxMjM0NTY="
		t.Setenv("AUTH_TOKEN_HASH_KEY", key)
		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(cfg.AuthTokenHashKey) != 32 {
			t.Fatalf("expected 32 bytes, got %d", len(cfg.AuthTokenHashKey))
		}
	})

	t.Run("too_short", func(t *testing.T) {
		key := "aGVsbG8="
		t.Setenv("AUTH_CSRF_SECRET", key)
		if _, err := config.Load(); err == nil {
			t.Fatal("expected error for short secret")
		}
	})

	t.Run("empty_allowed", func(t *testing.T) {
		t.Setenv("AUTH_TOKEN_HASH_KEY", "")
		t.Setenv("AUTH_CSRF_SECRET", "")
		t.Setenv("RATE_LIMIT_KEY_SECRET", "")
		_, err := config.Load()
		if err != nil {
			t.Fatalf("expected empty secrets to be allowed: %v", err)
		}
	})
}

func TestLoadCORSOrigins(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/pdv?sslmode=disable")
	t.Setenv("AUTH_ALLOW_EPHEMERAL_DEV_KEY", "true")
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:4173")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.CORSAllowedOrigins) != 2 {
		t.Fatalf("expected 2 origins, got %d", len(cfg.CORSAllowedOrigins))
	}
}

func TestLoadCustomCookieNames(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/pdv?sslmode=disable")
	t.Setenv("AUTH_ALLOW_EPHEMERAL_DEV_KEY", "true")
	t.Setenv("COOKIE_REFRESH_NAME", "my_refresh")
	t.Setenv("COOKIE_CSRF_NAME", "my_csrf")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.CookieRefreshName != "my_refresh" {
		t.Fatalf("expected my_refresh, got %q", cfg.CookieRefreshName)
	}
	if cfg.CookieCSRFName != "my_csrf" {
		t.Fatalf("expected my_csrf, got %q", cfg.CookieCSRFName)
	}
}
