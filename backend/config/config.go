package config

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	// App
	AppEnv  string
	Address string

	// Database
	DatabaseURL string

	// Valkey
	ValkeyAddr     string
	ValkeyPassword string
	ValkeyDB       int

	// JWT
	JWTIssuer         string
	JWTAudience       string
	JWTActiveKeyID    string
	JWTPrivateKeyPath string
	JWTPublicKeysDir  string
	AccessTokenTTL    time.Duration
	JWTClockSkew      time.Duration

	// Session / Refresh
	RefreshIdleTTL     time.Duration
	SessionAbsoluteTTL time.Duration

	// Secrets (decoded binary)
	AuthTokenHashKey   []byte
	AuthCSRFSecret     []byte
	RateLimitKeySecret []byte

	// Auth feature flags
	AuthRegistrationEnabled   bool
	AuthRequireVerifiedEmail  bool
	AuthAllowEphemeralDevKey  bool
	AuthTenantCreationEnabled bool

	// Cookies
	CookieSecure      bool
	CookieSameSite    string
	CookieRefreshName string
	CookieCSRFName    string

	// Network
	CORSAllowedOrigins []string
	TrustedProxyCIDRs  []string
	AppPublicURL       string

	// Password (Argon2 parameters)
	PasswordArgon2MemoryKiB   uint32
	PasswordArgon2Iterations  uint32
	PasswordArgon2Parallelism uint32
	PasswordBlocklistPath     string

	// Mail
	MailDriver   string
	MailFrom     string
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPStartTLS bool
}

func Load() (Config, error) {
	appEnv := getEnv("APP_ENV", "development")

	accessTokenTTL := mustParseDuration(getEnv("ACCESS_TOKEN_TTL", "5m"))

	refreshIdleTTL := mustParseDuration(getEnv("REFRESH_IDLE_TTL", "720h"))

	sessionAbsoluteTTL := mustParseDuration(getEnv("SESSION_ABSOLUTE_TTL", "2160h"))

	jwtClockSkew := mustParseDuration(getEnv("JWT_CLOCK_SKEW", "30s"))

	cookieSecure := getEnv("COOKIE_SECURE", "")
	if cookieSecure == "" {
		cookieSecure = strconv.FormatBool(appEnv == "production")
	}

	corsRaw := getEnv("CORS_ALLOWED_ORIGINS", "")
	var corsOrigins []string
	if corsRaw != "" {
		corsOrigins = strings.Split(corsRaw, ",")
	}

	cidrsRaw := getEnv("TRUSTED_PROXY_CIDRS", "")
	var cidrs []string
	if cidrsRaw != "" {
		cidrs = strings.Split(cidrsRaw, ",")
	}

	var authTokenHashKey []byte
	if raw := os.Getenv("AUTH_TOKEN_HASH_KEY"); raw != "" {
		authTokenHashKey = decodeSecret(raw)
	}

	var authCSRFSecret []byte
	if raw := os.Getenv("AUTH_CSRF_SECRET"); raw != "" {
		authCSRFSecret = decodeSecret(raw)
	}

	var rateLimitKeySecret []byte
	if raw := os.Getenv("RATE_LIMIT_KEY_SECRET"); raw != "" {
		rateLimitKeySecret = decodeSecret(raw)
	}

	valkeyDB, _ := strconv.Atoi(getEnv("VALKEY_DB", "0"))

	argon2Memory, _ := strconv.ParseUint(getEnv("PASSWORD_ARGON2_MEMORY_KIB", "65536"), 10, 32)
	argon2Iterations, _ := strconv.ParseUint(getEnv("PASSWORD_ARGON2_ITERATIONS", "3"), 10, 32)
	argon2Parallelism, _ := strconv.ParseUint(getEnv("PASSWORD_ARGON2_PARALLELISM", "1"), 10, 32)

	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))

	cfg := Config{
		AppEnv:  appEnv,
		Address: getEnv("HTTP_ADDRESS", ":8080"),

		DatabaseURL: os.Getenv("DATABASE_URL"),

		ValkeyAddr:     getEnv("VALKEY_ADDR", "localhost:6379"),
		ValkeyPassword: os.Getenv("VALKEY_PASSWORD"),
		ValkeyDB:       valkeyDB,

		JWTIssuer:         getEnv("JWT_ISSUER", "pdv-auth"),
		JWTAudience:       getEnv("JWT_AUDIENCE", "pdv-api"),
		JWTActiveKeyID:    os.Getenv("JWT_ACTIVE_KEY_ID"),
		JWTPrivateKeyPath: os.Getenv("JWT_PRIVATE_KEY_PATH"),
		JWTPublicKeysDir:  os.Getenv("JWT_PUBLIC_KEYS_DIR"),
		AccessTokenTTL:    accessTokenTTL,
		JWTClockSkew:      jwtClockSkew,

		RefreshIdleTTL:     refreshIdleTTL,
		SessionAbsoluteTTL: sessionAbsoluteTTL,

		AuthTokenHashKey:   authTokenHashKey,
		AuthCSRFSecret:     authCSRFSecret,
		RateLimitKeySecret: rateLimitKeySecret,

		AuthRegistrationEnabled:   getEnvBool("AUTH_REGISTRATION_ENABLED", "true"),
		AuthRequireVerifiedEmail:  getEnvBool("AUTH_REQUIRE_VERIFIED_EMAIL", "false"),
		AuthAllowEphemeralDevKey:  getEnvBool("AUTH_ALLOW_EPHEMERAL_DEV_KEY", "false"),
		AuthTenantCreationEnabled: getEnvBool("AUTH_TENANT_CREATION_ENABLED", "true"),

		CookieSecure:      getEnvBool("COOKIE_SECURE", cookieSecure),
		CookieSameSite:    getEnv("COOKIE_SAME_SITE", "Lax"),
		CookieRefreshName: getEnv("COOKIE_REFRESH_NAME", defaultCookieRefreshName(appEnv)),
		CookieCSRFName:    getEnv("COOKIE_CSRF_NAME", defaultCookieCSRFName(appEnv)),

		CORSAllowedOrigins: corsOrigins,
		TrustedProxyCIDRs:  cidrs,
		AppPublicURL:       os.Getenv("APP_PUBLIC_URL"),

		PasswordArgon2MemoryKiB:   uint32(argon2Memory),
		PasswordArgon2Iterations:  uint32(argon2Iterations),
		PasswordArgon2Parallelism: uint32(argon2Parallelism),
		PasswordBlocklistPath:     os.Getenv("PASSWORD_BLOCKLIST_PATH"),

		MailDriver:   getEnv("MAIL_DRIVER", "log"),
		MailFrom:     getEnv("MAIL_FROM", "noreply@pdv.local"),
		SMTPHost:     os.Getenv("SMTP_HOST"),
		SMTPPort:     smtpPort,
		SMTPUsername: os.Getenv("SMTP_USERNAME"),
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
		SMTPStartTLS: getEnvBool("SMTP_STARTTLS", "true"),
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	switch c.AppEnv {
	case "development", "test", "production":
	default:
		return fmt.Errorf("APP_ENV must be one of: development, test, production (got %q)", c.AppEnv)
	}

	if c.AccessTokenTTL < time.Minute || c.AccessTokenTTL > 15*time.Minute {
		return fmt.Errorf("ACCESS_TOKEN_TTL must be between 1m and 15m (got %s)", c.AccessTokenTTL)
	}

	if c.RefreshIdleTTL <= 0 {
		return fmt.Errorf("REFRESH_IDLE_TTL must be positive")
	}

	if c.SessionAbsoluteTTL <= 0 {
		return fmt.Errorf("SESSION_ABSOLUTE_TTL must be positive")
	}

	if c.RefreshIdleTTL > c.SessionAbsoluteTTL {
		return fmt.Errorf("REFRESH_IDLE_TTL (%s) must not exceed SESSION_ABSOLUTE_TTL (%s)", c.RefreshIdleTTL, c.SessionAbsoluteTTL)
	}

	if len(c.AuthTokenHashKey) > 0 && len(c.AuthTokenHashKey) < 32 {
		return fmt.Errorf("AUTH_TOKEN_HASH_KEY must decode to at least 32 bytes (got %d)", len(c.AuthTokenHashKey))
	}

	if len(c.AuthCSRFSecret) > 0 && len(c.AuthCSRFSecret) < 32 {
		return fmt.Errorf("AUTH_CSRF_SECRET must decode to at least 32 bytes (got %d)", len(c.AuthCSRFSecret))
	}

	if len(c.RateLimitKeySecret) > 0 && len(c.RateLimitKeySecret) < 32 {
		return fmt.Errorf("RATE_LIMIT_KEY_SECRET must decode to at least 32 bytes (got %d)", len(c.RateLimitKeySecret))
	}

	if c.AppEnv == "production" {
		if !c.CookieSecure {
			return fmt.Errorf("COOKIE_SECURE must be true in production")
		}

		if c.JWTPrivateKeyPath == "" && !c.AuthAllowEphemeralDevKey {
			return fmt.Errorf("JWT_PRIVATE_KEY_PATH is required in production")
		}

		if slices.Contains(c.CORSAllowedOrigins, "*") {
			return fmt.Errorf("CORS_ALLOWED_ORIGINS must not contain wildcard in production")
		}
	}

	if c.AppEnv != "production" {
		if c.JWTPrivateKeyPath == "" && !c.AuthAllowEphemeralDevKey {
			return fmt.Errorf("JWT_PRIVATE_KEY_PATH or AUTH_ALLOW_EPHEMERAL_DEV_KEY=true is required")
		}
	}

	if c.AuthAllowEphemeralDevKey && c.AppEnv == "production" && c.JWTPrivateKeyPath != "" {
		return fmt.Errorf("ephemeral dev key must not be enabled with a persisted key in production")
	}

	if c.JWTActiveKeyID != "" && c.JWTPrivateKeyPath == "" && !c.AuthAllowEphemeralDevKey {
		return fmt.Errorf("JWT_ACTIVE_KEY_ID requires a private key (JWT_PRIVATE_KEY_PATH or ephemeral)")
	}

	if c.JWTClockSkew <= 0 || c.JWTClockSkew > 30*time.Second {
		return fmt.Errorf("JWT_CLOCK_SKEW must be positive and at most 30s")
	}

	if c.PasswordArgon2MemoryKiB < 8 || c.PasswordArgon2MemoryKiB > 1<<24 {
		return fmt.Errorf("PASSWORD_ARGON2_MEMORY_KIB must be between 8 and 16777216")
	}
	if c.PasswordArgon2Iterations < 1 || c.PasswordArgon2Iterations > 100 {
		return fmt.Errorf("PASSWORD_ARGON2_ITERATIONS must be between 1 and 100")
	}
	if c.PasswordArgon2Parallelism < 1 || c.PasswordArgon2Parallelism > 256 {
		return fmt.Errorf("PASSWORD_ARGON2_PARALLELISM must be between 1 and 256")
	}

	for _, cidr := range c.TrustedProxyCIDRs {
		if _, _, err := net.ParseCIDR(strings.TrimSpace(cidr)); err != nil {
			return fmt.Errorf("TRUSTED_PROXY_CIDRS contains invalid CIDR %q: %w", cidr, err)
		}
	}

	switch strings.ToLower(c.CookieSameSite) {
	case "lax", "strict", "none":
	default:
		return fmt.Errorf("COOKIE_SAME_SITE must be one of: Lax, Strict, None (got %q)", c.CookieSameSite)
	}

	if c.AppPublicURL != "" {
		if _, err := url.ParseRequestURI(c.AppPublicURL); err != nil {
			return fmt.Errorf("APP_PUBLIC_URL must be a valid absolute URL: %w", err)
		}
	}

	return nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvBool(key, fallback string) bool {
	raw := os.Getenv(key)
	if raw == "" {
		raw = fallback
	}
	switch strings.ToLower(raw) {
	case "true", "1", "yes":
		return true
	default:
		return false
	}
}

func mustParseDuration(raw string) time.Duration {
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0
	}
	return d
}

func decodeSecret(raw string) []byte {
	data, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		data, err = base64.URLEncoding.DecodeString(raw)
	}
	if err != nil {
		return []byte(raw)
	}
	return data
}

func defaultCookieRefreshName(appEnv string) string {
	if appEnv == "production" {
		return "__Host-pdv_refresh"
	}
	return "pdv_refresh"
}

func defaultCookieCSRFName(appEnv string) string {
	if appEnv == "production" {
		return "__Host-pdv_csrf"
	}
	return "pdv_csrf"
}
