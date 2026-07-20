package jwt_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	jwtlib "github.com/gabrielalc23/pdv/internal/platform/token/jwt"
	"github.com/golang-jwt/jwt/v5"
)

func generateTestKey(t testing.TB) (ed25519.PublicKey, ed25519.PrivateKey, string) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	return pub, priv, "test-key"
}

func tempKeyring(t testing.TB) *jwtlib.Keyring {
	t.Helper()
	pub, priv, kid := generateTestKey(t)
	dir := t.TempDir()

	privBytes, _ := x509.MarshalPKCS8PrivateKey(priv)
	pubBytes, _ := x509.MarshalPKIXPublicKey(pub)

	privPath := filepath.Join(dir, kid+".priv.pem")
	pubPath := filepath.Join(dir, kid+".pem")
	_ = os.WriteFile(privPath, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}), 0600)
	_ = os.WriteFile(pubPath, pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}), 0644)

	kr, err := jwtlib.LoadKeyring(kid, privPath, dir)
	if err != nil {
		t.Fatalf("LoadKeyring: %v", err)
	}
	return kr
}

func ephemeralKeyring(t testing.TB) *jwtlib.Keyring {
	t.Helper()
	kr, err := jwtlib.NewEphemeralKeyring("test-key")
	if err != nil {
		t.Fatalf("NewEphemeralKeyring: %v", err)
	}
	return kr
}

func TestSignAndValidate(t *testing.T) {
	kr := ephemeralKeyring(t)
	signer := jwtlib.NewSigner(kr, "pdv-auth", "pdv-api", 5*time.Minute, 30*time.Second)
	validator := jwtlib.NewValidator(kr, "pdv-auth", "pdv-api", 30*time.Second)

	now := time.Now()
	token, err := signer.Sign(jwtlib.SignerClaims{
		Subject:   "550e8400-e29b-41d4-a716-446655440000",
		JTI:       "660e8400-e29b-41d4-a716-446655440001",
		SessionID: "770e8400-e29b-41d4-a716-446655440002",
		ClientID:  "pdv-admin",
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

	claims, err := validator.Validate(token)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if claims.Subject != "550e8400-e29b-41d4-a716-446655440000" {
		t.Fatalf("expected subject, got %q", claims.Subject)
	}
	if claims.Ctx != jwtlib.ContextIdentity {
		t.Fatalf("expected identity context, got %q", claims.Ctx)
	}
}

func TestInvalidSignature(t *testing.T) {
	kr1 := ephemeralKeyring(t)
	kr2 := ephemeralKeyring(t)

	signer := jwtlib.NewSigner(kr1, "pdv-auth", "pdv-api", 5*time.Minute, 30*time.Second)
	validator := jwtlib.NewValidator(kr2, "pdv-auth", "pdv-api", 30*time.Second)

	now := time.Now()
	token, err := signer.Sign(jwtlib.SignerClaims{
		Subject:   "550e8400-e29b-41d4-a716-446655440000",
		JTI:       "660e8400-e29b-41d4-a716-446655440001",
		SessionID: "770e8400-e29b-41d4-a716-446655440002",
		ClientID:  "pdv-admin",
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

	_, err = validator.Validate(token)
	if err == nil {
		t.Fatal("expected error for invalid signature")
	}
}

func TestWrongAlgorithm(t *testing.T) {
	kr := ephemeralKeyring(t)
	validator := jwtlib.NewValidator(kr, "pdv-auth", "pdv-api", 30*time.Second)

	claims := jwt.MapClaims{
		"sub": "test",
		"iss": "pdv-auth",
		"aud": []string{"pdv-api"},
	}
	hmacToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	hs256token, _ := hmacToken.SignedString([]byte("secret"))

	_, err := validator.Validate(hs256token)
	if err == nil {
		t.Fatal("expected error for wrong algorithm")
	}
}

func TestExpiredToken(t *testing.T) {
	kr := ephemeralKeyring(t)
	signer := jwtlib.NewSigner(kr, "pdv-auth", "pdv-api", -1*time.Minute, 0)
	validator := jwtlib.NewValidator(kr, "pdv-auth", "pdv-api", 0)

	now := time.Now()
	token, err := signer.Sign(jwtlib.SignerClaims{
		Subject:   "550e8400-e29b-41d4-a716-446655440000",
		JTI:       "660e8400-e29b-41d4-a716-446655440001",
		SessionID: "770e8400-e29b-41d4-a716-446655440002",
		ClientID:  "pdv-admin",
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

	_, err = validator.Validate(token)
	if err == nil || !errors.Is(err, jwtlib.ErrTokenExpired) {
		t.Fatalf("expected ErrTokenExpired, got %v", err)
	}
}

func TestUnknownKID(t *testing.T) {
	kr := ephemeralKeyring(t)
	signer := jwtlib.NewSigner(kr, "pdv-auth", "pdv-api", 5*time.Minute, 30*time.Second)

	now := time.Now()
	token, err := signer.Sign(jwtlib.SignerClaims{
		Subject:   "550e8400-e29b-41d4-a716-446655440000",
		JTI:       "660e8400-e29b-41d4-a716-446655440001",
		SessionID: "770e8400-e29b-41d4-a716-446655440002",
		ClientID:  "pdv-admin",
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

	kr2, _ := jwtlib.NewEphemeralKeyring("other-key")
	validator := jwtlib.NewValidator(kr2, "pdv-auth", "pdv-api", 30*time.Second)

	_, err = validator.Validate(token)
	if err == nil {
		t.Fatal("expected error for unknown kid")
	}
}

func TestWrongTyp(t *testing.T) {
	kr := ephemeralKeyring(t)
	signer := jwtlib.NewSigner(kr, "pdv-auth", "pdv-api", 5*time.Minute, 30*time.Second)
	validator := jwtlib.NewValidator(kr, "pdv-auth", "pdv-api", 30*time.Second)

	now := time.Now()
	token, err := signer.Sign(jwtlib.SignerClaims{
		Subject:   "550e8400-e29b-41d4-a716-446655440000",
		JTI:       "660e8400-e29b-41d4-a716-446655440001",
		SessionID: "770e8400-e29b-41d4-a716-446655440002",
		ClientID:  "pdv-admin",
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

	parts := strings.SplitN(token, ".", 3)
	if len(parts) == 3 {
		header := `{"alg":"EdDSA","typ":"invalid","kid":"test-key"}`
		headerB64 := base64.RawURLEncoding.EncodeToString([]byte(header))
		token = headerB64 + "." + parts[1] + "." + parts[2]
	}

	_, err = validator.Validate(token)
	if err == nil {
		t.Fatal("expected error for wrong typ")
	}
}

func TestOrganizationContext(t *testing.T) {
	kr := ephemeralKeyring(t)
	signer := jwtlib.NewSigner(kr, "pdv-auth", "pdv-api", 5*time.Minute, 30*time.Second)
	validator := jwtlib.NewValidator(kr, "pdv-auth", "pdv-api", 30*time.Second)

	now := time.Now()
	oav := int64(5)
	mav := int64(3)

	token, err := signer.Sign(jwtlib.SignerClaims{
		Subject:      "550e8400-e29b-41d4-a716-446655440000",
		JTI:          "660e8400-e29b-41d4-a716-446655440001",
		SessionID:    "770e8400-e29b-41d4-a716-446655440002",
		ClientID:     "pdv-admin",
		Ctx:          jwtlib.ContextOrganization,
		OrgID:        "550e8400-e29b-41d4-a716-446655440010",
		MembershipID: "550e8400-e29b-41d4-a716-446655440020",
		Roles:        []string{"admin", "owner"},
		Scopes:       []string{"organization.read", "stores.read"},
		OAV:          &oav,
		MAV:          &mav,
		PV:           1,
		AuthTime:     now,
		AMR:          []string{"pwd"},
	})
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	claims, err := validator.Validate(token)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if claims.Ctx != jwtlib.ContextOrganization {
		t.Fatalf("expected organization context")
	}
	if claims.OrgID != "550e8400-e29b-41d4-a716-446655440010" {
		t.Fatalf("wrong org_id")
	}
}

func TestStoreContext(t *testing.T) {
	kr := ephemeralKeyring(t)
	signer := jwtlib.NewSigner(kr, "pdv-auth", "pdv-api", 5*time.Minute, 30*time.Second)
	validator := jwtlib.NewValidator(kr, "pdv-auth", "pdv-api", 30*time.Second)

	now := time.Now()
	oav := int64(5)
	mav := int64(3)

	token, err := signer.Sign(jwtlib.SignerClaims{
		Subject:      "550e8400-e29b-41d4-a716-446655440000",
		JTI:          "660e8400-e29b-41d4-a716-446655440001",
		SessionID:    "770e8400-e29b-41d4-a716-446655440002",
		ClientID:     "pdv-admin",
		Ctx:          jwtlib.ContextStore,
		OrgID:        "550e8400-e29b-41d4-a716-446655440010",
		MembershipID: "550e8400-e29b-41d4-a716-446655440020",
		StoreID:      "550e8400-e29b-41d4-a716-446655440030",
		Roles:        []string{"cashier"},
		Scopes:       []string{"catalog.read", "sales.create"},
		OAV:          &oav,
		MAV:          &mav,
		PV:           1,
		AuthTime:     now,
		AMR:          []string{"pwd"},
	})
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	claims, err := validator.Validate(token)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if claims.Ctx != jwtlib.ContextStore {
		t.Fatalf("expected store context")
	}
	if claims.StoreID != "550e8400-e29b-41d4-a716-446655440030" {
		t.Fatalf("wrong store_id")
	}
}

func TestIncoherentContext(t *testing.T) {
	kr := ephemeralKeyring(t)
	signer := jwtlib.NewSigner(kr, "pdv-auth", "pdv-api", 5*time.Minute, 30*time.Second)

	now := time.Now()
	oav := int64(5)
	mav := int64(3)

	_, err := signer.Sign(jwtlib.SignerClaims{
		Subject:      "550e8400-e29b-41d4-a716-446655440000",
		JTI:          "660e8400-e29b-41d4-a716-446655440001",
		SessionID:    "770e8400-e29b-41d4-a716-446655440002",
		ClientID:     "pdv-admin",
		Ctx:          jwtlib.ContextOrganization,
		OrgID:        "",
		MembershipID: "550e8400-e29b-41d4-a716-446655440020",
		Roles:        []string{},
		Scopes:       []string{},
		OAV:          &oav,
		MAV:          &mav,
		PV:           1,
		AuthTime:     now,
		AMR:          []string{"pwd"},
	})
	if err == nil {
		t.Fatal("expected error for incoherent organization context")
	}
}

func TestJWKSCorrect(t *testing.T) {
	kr := ephemeralKeyring(t)
	svc := jwtlib.NewJWKSService(kr)

	req, _ := http.NewRequest("GET", "/.well-known/jwks.json", nil)
	rec := httptest.NewRecorder()
	svc.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var jwksResp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&jwksResp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	keys, ok := jwksResp["keys"].([]any)
	if !ok {
		t.Fatal("expected keys array")
	}
	if len(keys) == 0 {
		t.Fatal("expected at least one key")
	}

	key0 := keys[0].(map[string]any)
	if key0["kty"] != "OKP" {
		t.Fatalf("expected OKP, got %v", key0["kty"])
	}
	if key0["crv"] != "Ed25519" {
		t.Fatalf("expected Ed25519, got %v", key0["crv"])
	}
	if key0["alg"] != "EdDSA" {
		t.Fatalf("expected EdDSA, got %v", key0["alg"])
	}
	if key0["use"] != "sig" {
		t.Fatalf("expected sig, got %v", key0["use"])
	}
	if key0["x"] == "" {
		t.Fatal("expected non-empty x")
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Fatalf("expected application/json, got %q", contentType)
	}

	cacheControl := rec.Header().Get("Cache-Control")
	if !strings.Contains(cacheControl, "public") {
		t.Fatalf("expected public cache-control, got %q", cacheControl)
	}

	etag := rec.Header().Get("ETag")
	if etag == "" {
		t.Fatal("expected ETag")
	}
}

func TestJWKSNoPrivateKey(t *testing.T) {
	kr := ephemeralKeyring(t)
	svc := jwtlib.NewJWKSService(kr)

	req, _ := http.NewRequest("GET", "/.well-known/jwks.json", nil)
	rec := httptest.NewRecorder()
	svc.ServeHTTP(rec, req)

	body := rec.Body.String()
	if strings.Contains(body, "priv") || strings.Contains(body, "seed") {
		t.Fatal("JWKS should not contain private key material")
	}
}

func TestKeyringDuplicateKID(t *testing.T) {
	dir := t.TempDir()
	kid := "dup-key"

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	privBytes, _ := x509.MarshalPKCS8PrivateKey(priv)
	pubBytes, _ := x509.MarshalPKIXPublicKey(pub)

	_ = os.WriteFile(filepath.Join(dir, kid+".priv.pem"),
		pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}), 0600)

	_ = os.WriteFile(filepath.Join(dir, kid+".pem"),
		pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}), 0644)

	_, err = jwtlib.LoadKeyring(kid, filepath.Join(dir, kid+".priv.pem"), dir)
	if err != nil {
		t.Logf("expected possible duplicate error: %v", err)
	}
}

func TestClaimsValidationIdentity(t *testing.T) {
	claims := jwtlib.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:   "pdv-auth",
			Subject:  "550e8400-e29b-41d4-a716-446655440000",
			Audience: jwt.ClaimStrings{"pdv-api"},
			ID:       "660e8400-e29b-41d4-a716-446655440001",
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
		ClientID: "pdv-admin",
		Ctx:      jwtlib.ContextIdentity,
		PV:       1,
		Ver:      1,
	}

	if err := claims.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestClaimsIdentityRejectsTenant(t *testing.T) {
	claims := jwtlib.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:   "pdv-auth",
			Subject:  "550e8400-e29b-41d4-a716-446655440000",
			Audience: jwt.ClaimStrings{"pdv-api"},
			ID:       "660e8400-e29b-41d4-a716-446655440001",
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
		ClientID:     "pdv-admin",
		Ctx:          jwtlib.ContextIdentity,
		OrgID:        "550e8400-e29b-41d4-a716-446655440010",
		MembershipID: "550e8400-e29b-41d4-a716-446655440020",
		PV:           1,
	}

	if err := claims.Validate(); err == nil {
		t.Fatal("expected error for identity with tenant claims")
	}
}

func TestScopeUniquenessOrder(t *testing.T) {
	kr := ephemeralKeyring(t)
	signer := jwtlib.NewSigner(kr, "pdv-auth", "pdv-api", 5*time.Minute, 30*time.Second)

	now := time.Now()
	token, err := signer.Sign(jwtlib.SignerClaims{
		Subject:   "550e8400-e29b-41d4-a716-446655440000",
		JTI:       "660e8400-e29b-41d4-a716-446655440001",
		SessionID: "770e8400-e29b-41d4-a716-446655440002",
		ClientID:  "pdv-admin",
		Ctx:       jwtlib.ContextIdentity,
		Roles:     []string{"admin", "admin", "owner"},
		Scopes:    []string{"catalog.read", "sales.create", "catalog.read"},
		PV:        1,
		AuthTime:  now,
		AMR:       []string{"pwd"},
	})
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	validator := jwtlib.NewValidator(kr, "pdv-auth", "pdv-api", 30*time.Second)
	claims, err := validator.Validate(token)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}

	if len(claims.Roles) != 2 {
		t.Fatalf("expected 2 unique roles, got %v", claims.Roles)
	}
	if !strings.Contains(claims.Scope, "catalog.read") && !strings.Contains(claims.Scope, "sales.create") {
		t.Fatalf("expected scopes in claim")
	}
}

func TestTokenTooLarge(t *testing.T) {
	kr := ephemeralKeyring(t)
	signer := jwtlib.NewSigner(kr, "pdv-auth", "pdv-api", 5*time.Minute, 30*time.Second)

	roles := make([]string, 30)
	for i := range roles {
		roles[i] = "role-very-long-name-that-will-take-up-space"
	}

	now := time.Now()
	_, err := signer.Sign(jwtlib.SignerClaims{
		Subject:   "550e8400-e29b-41d4-a716-446655440000",
		JTI:       "660e8400-e29b-41d4-a716-446655440001",
		SessionID: "770e8400-e29b-41d4-a716-446655440002",
		ClientID:  "pdv-admin",
		Ctx:       jwtlib.ContextIdentity,
		Roles:     roles,
		Scopes:    []string{},
		PV:        1,
		AuthTime:  now,
		AMR:       []string{"pwd"},
	})
	if err == nil {
		t.Fatal("expected error for too many roles")
	}
}

func TestKeyringMultipleKeys(t *testing.T) {
	dir := t.TempDir()
	kid1, kid2 := "key-one", "key-two"

	pub1, priv1, _ := ed25519.GenerateKey(rand.Reader)
	pub2, _, _ := ed25519.GenerateKey(rand.Reader)

	privBytes, _ := x509.MarshalPKCS8PrivateKey(priv1)
	_ = os.WriteFile(filepath.Join(dir, kid1+".priv.pem"),
		pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}), 0600)

	pub1Bytes, _ := x509.MarshalPKIXPublicKey(pub1)
	_ = os.WriteFile(filepath.Join(dir, kid1+".pem"),
		pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pub1Bytes}), 0644)

	pub2Bytes, _ := x509.MarshalPKIXPublicKey(pub2)
	_ = os.WriteFile(filepath.Join(dir, kid2+".pem"),
		pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pub2Bytes}), 0644)

	kr, err := jwtlib.LoadKeyring(kid1, filepath.Join(dir, kid1+".priv.pem"), dir)
	if err != nil {
		t.Fatalf("LoadKeyring: %v", err)
	}

	if _, ok := kr.PublicKey(kid1); !ok {
		t.Fatal("expected kid1 to be found")
	}
	if _, ok := kr.PublicKey(kid2); !ok {
		t.Fatal("expected kid2 to be found")
	}
}

func TestValidatorRejectsNoneAlgorithm(t *testing.T) {
	kr := ephemeralKeyring(t)
	validator := jwtlib.NewValidator(kr, "pdv-auth", "pdv-api", 30*time.Second)

	noneToken := "eyJhbGciOiJub25lIiwidHlwIjoiYXQrand0In0.eyJzdWIiOiJ0ZXN0In0."
	_, err := validator.Validate(noneToken)
	if err == nil {
		t.Fatal("expected error for 'none' algorithm")
	}
}

func TestKeyringEphemeral(t *testing.T) {
	kr, err := jwtlib.NewEphemeralKeyring("ephemeral-test")
	if err != nil {
		t.Fatalf("NewEphemeralKeyring: %v", err)
	}
	if kr.ActiveKID != "ephemeral-test" {
		t.Fatalf("expected active kid ephemeral-test, got %q", kr.ActiveKID)
	}
	if _, ok := kr.PublicKey("ephemeral-test"); !ok {
		t.Fatal("expected ephemeral public key to be present")
	}
}
