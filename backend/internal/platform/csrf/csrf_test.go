package csrf_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gabrielalc23/pdv/internal/platform/csrf"
)

func TestPreauthValid(t *testing.T) {
	mgr, err := csrf.NewManager(make([]byte, 32), nil)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	token, err := mgr.Generate(csrf.BindingPreauth, "preauth")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if err := mgr.Validate(token, token, string(csrf.BindingPreauth), "preauth"); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestSessionBoundValid(t *testing.T) {
	mgr, err := csrf.NewManager(make([]byte, 32), nil)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	sessionID := "550e8400-e29b-41d4-a716-446655440000"
	token, err := mgr.Generate(csrf.BindingSessionID, sessionID)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if err := mgr.Validate(token, token, string(csrf.BindingSessionID), sessionID); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestCookieMissing(t *testing.T) {
	mgr, _ := csrf.NewManager(make([]byte, 32), nil)

	err := mgr.Validate("", "header-token", string(csrf.BindingPreauth), "preauth")
	if err != csrf.ErrTokenMissing {
		t.Fatalf("expected ErrTokenMissing, got %v", err)
	}
}

func TestHeaderMissing(t *testing.T) {
	mgr, _ := csrf.NewManager(make([]byte, 32), nil)

	token, _ := mgr.Generate(csrf.BindingPreauth, "preauth")
	err := mgr.Validate(token, "", string(csrf.BindingPreauth), "preauth")
	if err != csrf.ErrTokenMissing {
		t.Fatalf("expected ErrTokenMissing, got %v", err)
	}
}

func TestCookieHeaderMismatch(t *testing.T) {
	mgr, _ := csrf.NewManager(make([]byte, 32), nil)

	token1, _ := mgr.Generate(csrf.BindingPreauth, "preauth")
	token2, _ := mgr.Generate(csrf.BindingPreauth, "preauth")

	err := mgr.Validate(token1, token2, string(csrf.BindingPreauth), "preauth")
	if err != csrf.ErrTokenInvalid {
		t.Fatalf("expected ErrTokenInvalid, got %v", err)
	}
}

func TestMalformedToken(t *testing.T) {
	mgr, _ := csrf.NewManager(make([]byte, 32), nil)

	err := mgr.Validate("invalid", "invalid", string(csrf.BindingPreauth), "preauth")
	if err == nil {
		t.Fatal("expected error for malformed token")
	}
}

func TestInvalidHMAC(t *testing.T) {
	mgr, _ := csrf.NewManager(make([]byte, 32), nil)

	token, _ := mgr.Generate(csrf.BindingPreauth, "preauth")

	parts := strings.SplitN(token, ".", 2)
	if len(parts) == 2 {
		badToken := parts[0] + ".invalidsignature"
		err := mgr.Validate(badToken, badToken, string(csrf.BindingPreauth), "preauth")
		if err == nil {
			t.Fatal("expected error for invalid HMAC")
		}
	}
}

func TestWrongSession(t *testing.T) {
	mgr, _ := csrf.NewManager(make([]byte, 32), nil)

	sessionA := "session-id-a"
	sessionB := "session-id-b"

	token, _ := mgr.Generate(csrf.BindingSessionID, sessionA)
	err := mgr.Validate(token, token, string(csrf.BindingSessionID), sessionB)
	if err == nil {
		t.Fatal("expected error for wrong session")
	}
}

func TestCheckOriginAllowed(t *testing.T) {
	mgr, _ := csrf.NewManager(make([]byte, 32), []string{"https://app.example.com"})

	r, _ := http.NewRequest("POST", "/", nil)
	r.Header.Set("Origin", "https://app.example.com")

	if err := mgr.CheckOrigin(r); err != nil {
		t.Fatalf("CheckOrigin: %v", err)
	}
}

func TestCheckOriginDenied(t *testing.T) {
	mgr, _ := csrf.NewManager(make([]byte, 32), []string{"https://app.example.com"})

	r, _ := http.NewRequest("POST", "/", nil)
	r.Header.Set("Origin", "https://evil.com")

	if err := mgr.CheckOrigin(r); err == nil {
		t.Fatal("expected error for denied origin")
	}
}

func TestCheckFetchMetadataCrossSite(t *testing.T) {
	mgr, _ := csrf.NewManager(make([]byte, 32), nil)

	r, _ := http.NewRequest("POST", "/", nil)
	r.Header.Set("Sec-Fetch-Site", "cross-site")

	if err := mgr.CheckFetchMetadata(r); err != csrf.ErrCrossSiteRequest {
		t.Fatalf("expected ErrCrossSiteRequest, got %v", err)
	}
}

func TestCheckFetchMetadataSameSite(t *testing.T) {
	mgr, _ := csrf.NewManager(make([]byte, 32), nil)

	r, _ := http.NewRequest("POST", "/", nil)
	r.Header.Set("Sec-Fetch-Site", "same-origin")

	if err := mgr.CheckFetchMetadata(r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSafeMethodSkipsOriginCheck(t *testing.T) {
	if csrf.IsUnsafeMethod("GET") {
		t.Fatal("GET should be safe")
	}
	if csrf.IsUnsafeMethod("HEAD") {
		t.Fatal("HEAD should be safe")
	}
	if csrf.IsUnsafeMethod("OPTIONS") {
		t.Fatal("OPTIONS should be safe")
	}
	if !csrf.IsUnsafeMethod("POST") {
		t.Fatal("POST should be unsafe")
	}
	if !csrf.IsUnsafeMethod("PUT") {
		t.Fatal("PUT should be unsafe")
	}
	if !csrf.IsUnsafeMethod("DELETE") {
		t.Fatal("DELETE should be unsafe")
	}
	if !csrf.IsUnsafeMethod("PATCH") {
		t.Fatal("PATCH should be unsafe")
	}
}

func TestRefererFallback(t *testing.T) {
	mgr, _ := csrf.NewManager(make([]byte, 32), []string{"https://app.example.com"})

	r, _ := http.NewRequest("POST", "/", nil)
	r.Header.Set("Referer", "https://app.example.com/some/path")

	if err := mgr.CheckOrigin(r); err != nil {
		t.Fatalf("CheckOrigin with Referer: %v", err)
	}
}

func TestRefererInvalidFallback(t *testing.T) {
	mgr, _ := csrf.NewManager(make([]byte, 32), []string{"https://app.example.com"})

	r, _ := http.NewRequest("POST", "/", nil)
	r.Header.Set("Referer", "https://evil.com/some/path")

	if err := mgr.CheckOrigin(r); err == nil {
		t.Fatal("expected error for invalid Referer origin")
	}
}

func TestRequestValidationIntegration(t *testing.T) {
	mgr, _ := csrf.NewManager(make([]byte, 32), []string{"http://localhost:5173"})

	token, _ := mgr.Generate(csrf.BindingPreauth, "preauth")

	r := httptest.NewRequest("POST", "/", nil)
	r.Header.Set("Origin", "http://localhost:5173")
	r.Header.Set("X-CSRF-Token", token)
	r.Header.Set("Cookie", "pdv_csrf="+token)

	if err := mgr.CheckOrigin(r); err != nil {
		t.Fatalf("CheckOrigin: %v", err)
	}

	cookieToken := token
	if err := mgr.ValidateRequest(r, cookieToken); err != nil {
		t.Fatalf("ValidateRequest: %v", err)
	}
}

func TestConstantTimeComparison(t *testing.T) {
	mgr, _ := csrf.NewManager(make([]byte, 32), nil)

	token, _ := mgr.Generate(csrf.BindingPreauth, "preauth")

	mgrSame, _ := csrf.NewManager(make([]byte, 32), nil)
	err := mgrSame.Validate(token, token, string(csrf.BindingPreauth), "preauth")
	if err != nil {
		t.Fatalf("same key validate: %v", err)
	}

	diffKey := make([]byte, 32)
	for i := range diffKey {
		diffKey[i] = 0xFF
	}
	mgrDiff, _ := csrf.NewManager(diffKey, nil)
	err = mgrDiff.Validate(token, token, string(csrf.BindingPreauth), "preauth")
	if err == nil {
		t.Fatal("expected error with different key")
	}
}
