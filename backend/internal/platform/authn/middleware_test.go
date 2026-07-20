package authn

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/gabrielalc23/pdv/internal/platform/clock"
	jwtlib "github.com/gabrielalc23/pdv/internal/platform/token/jwt"
)

func ephemeralKeyring(t testing.TB) *jwtlib.Keyring {
	t.Helper()
	kr, err := jwtlib.NewEphemeralKeyring("test-key")
	if err != nil {
		t.Fatalf("NewEphemeralKeyring: %v", err)
	}
	return kr
}

func testMiddleware(t testing.TB) *Middleware {
	t.Helper()
	kr := ephemeralKeyring(t)
	validator := jwtlib.NewValidator(kr, "pdv-auth", "pdv-api", 30*time.Second)
	return NewMiddleware(
		validator,
		NewPersistenceStore(nil),
		NewSessionCache(nil, 60*time.Second),
		NewTouchThrottle(nil, 30*time.Second),
		clock.RealClock{},
	)
}

func okHandler(t testing.TB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}
}

func assertHeader(t testing.TB, rec *httptest.ResponseRecorder, key, expected string) {
	t.Helper()
	got := rec.Header().Get(key)
	if got != expected {
		t.Fatalf("expected header %s: %q, got %q", key, expected, got)
	}
}

func TestValidateBearer(t *testing.T) {
	tests := []struct {
		name    string
		header  string
		wantErr bool
		errCode string
	}{
		{"missing header", "", true, "ACCESS_TOKEN_MISSING"},
		{"no bearer prefix", "Basic dG9rZW4=", true, "ACCESS_TOKEN_INVALID"},
		{"bearer with empty token", "Bearer ", true, "ACCESS_TOKEN_INVALID"},
		{"bearer lowercase", "bearer token123", false, ""},
		{"valid bearer", "Bearer eyJhbGciOiJFZERTQSJ9.eyJzdWIiOiIxMjMifQ.test", false, ""},
		{"multiple parts", "Bearer token extra", true, "ACCESS_TOKEN_INVALID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validateBearer(tt.header)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				ae := mapErr(err)
				if ae.Code != tt.errCode {
					t.Fatalf("expected code %s, got %s", tt.errCode, ae.Code)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateBearerLargeToken(t *testing.T) {
	token := strings.Repeat("a", 20*1024)
	_, err := validateBearer("Bearer " + token)
	if err == nil {
		t.Fatal("expected error for large token")
	}
}

func TestMiddlewareNoAuthHeader(t *testing.T) {
	mw := testMiddleware(t)
	handler := mw.RequireAccessToken(okHandler(t))

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 401 {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
	assertHeader(t, rec, "WWW-Authenticate", "Bearer")
	assertHeader(t, rec, "Cache-Control", "no-store")
	assertHeader(t, rec, "Pragma", "no-cache")
}

func TestMiddlewareInvalidBearer(t *testing.T) {
	mw := testMiddleware(t)
	handler := mw.RequireAccessToken(okHandler(t))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Basic dG9rZW4=")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 401 {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestMiddlewareInvalidJWT(t *testing.T) {
	mw := testMiddleware(t)
	handler := mw.RequireAccessToken(okHandler(t))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 401 {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestMiddlewareExpiredJWT(t *testing.T) {
	kr := ephemeralKeyring(t)
	signer := jwtlib.NewSigner(kr, "pdv-auth", "pdv-api", -1*time.Minute, 0)
	validator := jwtlib.NewValidator(kr, "pdv-auth", "pdv-api", 0)

	now := time.Now()
	tokenStr, err := signer.Sign(jwtlib.SignerClaims{
		Subject:   "550e8400-e29b-41d4-a716-446655440000",
		JTI:       "660e8400-e29b-41d4-a716-446655440001",
		SessionID: "770e8400-e29b-41d4-a716-446655440002",
		ClientID:  "test-client",
		Ctx:       jwtlib.ContextIdentity,
		Roles:     []string{},
		Scopes:    []string{},
		PV:        1,
		AuthTime:  now,
		AMR:       []string{"pwd"},
	})
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	mw := NewMiddleware(validator,
		NewPersistenceStore(nil),
		NewSessionCache(nil, 60*time.Second),
		NewTouchThrottle(nil, 30*time.Second),
		clock.RealClock{},
	)
	handler := mw.RequireAccessToken(okHandler(t))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 401 {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestMiddlewareUnknownKID(t *testing.T) {
	kr1 := ephemeralKeyring(t)
	kr2 := ephemeralKeyring(t)

	signer := jwtlib.NewSigner(kr1, "pdv-auth", "pdv-api", 5*time.Minute, 30*time.Second)
	validator := jwtlib.NewValidator(kr2, "pdv-auth", "pdv-api", 30*time.Second)

	now := time.Now()
	tokenStr, err := signer.Sign(jwtlib.SignerClaims{
		Subject:   "550e8400-e29b-41d4-a716-446655440000",
		JTI:       "660e8400-e29b-41d4-a716-446655440001",
		SessionID: "770e8400-e29b-41d4-a716-446655440002",
		ClientID:  "test-client",
		Ctx:       jwtlib.ContextIdentity,
		Roles:     []string{},
		Scopes:    []string{},
		PV:        1,
		AuthTime:  now,
		AMR:       []string{"pwd"},
	})
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	mw := NewMiddleware(validator,
		NewPersistenceStore(nil),
		NewSessionCache(nil, 60*time.Second),
		NewTouchThrottle(nil, 30*time.Second),
		clock.RealClock{},
	)
	handler := mw.RequireAccessToken(okHandler(t))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 401 {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestBuildPrincipalIdentity(t *testing.T) {
	claims := testClaims(t, jwtlib.ContextIdentity)
	state := sessionState{}

	p := buildPrincipal(claims, state)

	if p.ContextKind != "identity" {
		t.Fatalf("expected identity context, got %s", p.ContextKind)
	}
	if !p.IsIdentity() {
		t.Fatal("expected IsIdentity")
	}
}

func TestBuildPrincipalOrganization(t *testing.T) {
	claims := testClaims(t, jwtlib.ContextOrganization)
	claims.OrgID = "880e8400-e29b-41d4-a716-446655440010"
	claims.MembershipID = "990e8400-e29b-41d4-a716-446655440020"
	oav := int64(1)
	mav := int64(1)
	claims.OAV = &oav
	claims.MAV = &mav

	state := sessionState{}
	p := buildPrincipal(claims, state)

	if p.ContextKind != "organization" {
		t.Fatalf("expected organization context, got %s", p.ContextKind)
	}
	if !p.IsOrganization() {
		t.Fatal("expected IsOrganization")
	}
	if !p.HasOrganizationScope() {
		t.Fatal("expected organization scope")
	}
}

func TestBuildPrincipalStore(t *testing.T) {
	claims := testClaims(t, jwtlib.ContextStore)
	claims.OrgID = "880e8400-e29b-41d4-a716-446655440010"
	claims.MembershipID = "990e8400-e29b-41d4-a716-446655440020"
	claims.StoreID = "aa0e8400-e29b-41d4-a716-446655440030"
	oav := int64(1)
	mav := int64(1)
	claims.OAV = &oav
	claims.MAV = &mav

	state := sessionState{}
	p := buildPrincipal(claims, state)

	if p.ContextKind != "store" {
		t.Fatalf("expected store context, got %s", p.ContextKind)
	}
	if !p.IsStore() {
		t.Fatal("expected IsStore")
	}
	if !p.HasStoreScope() {
		t.Fatal("expected store scope")
	}
}

func testClaims(t testing.TB, ctxKind jwtlib.ContextKind) *jwtlib.Claims {
	t.Helper()
	return &jwtlib.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "550e8400-e29b-41d4-a716-446655440000",
			ID:        "660e8400-e29b-41d4-a716-446655440001",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
			Issuer:    "pdv-auth",
			Audience:  jwt.ClaimStrings{"pdv-api"},
		},
		ClientID: "test-client",
		Ctx:      ctxKind,
		SID:      "770e8400-e29b-41d4-a716-446655440002",
		Roles:    []string{"admin"},
		Scope:    "read write",
		PV:       1,
		AuthTime: time.Now().Unix(),
		Ver:      1,
	}
}

func TestJWTErrMapping(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code string
	}{
		{"expired", jwtlib.ErrTokenExpired, "ACCESS_TOKEN_EXPIRED"},
		{"invalid signature", jwtlib.ErrTokenSignature, "ACCESS_TOKEN_INVALID"},
		{"invalid algorithm", jwtlib.ErrTokenAlgorithm, "ACCESS_TOKEN_INVALID"},
		{"invalid claims", jwtlib.ErrClaimsInvalid, "ACCESS_TOKEN_INVALID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapped := jwtErrToAuthnErr(tt.err)
			ae := mapErr(mapped)
			if ae.Code != tt.code {
				t.Fatalf("expected code %s, got %s", tt.code, ae.Code)
			}
		})
	}
}
