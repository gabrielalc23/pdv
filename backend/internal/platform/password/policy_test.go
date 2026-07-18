package password_test

import (
	"testing"

	"github.com/gabrielalc23/pdv/internal/platform/password"
)

func TestPolicyMinLength(t *testing.T) {
	p := password.DefaultPolicy()
	blocklist := password.NewBuiltinBlocklist()

	err := p.Validate("short", "", blocklist)
	if err != password.ErrPasswordTooShort {
		t.Fatalf("expected ErrPasswordTooShort, got %v", err)
	}
}

func TestPolicyMaxLength(t *testing.T) {
	p := password.DefaultPolicy()
	var long string
	for i := 0; i < 130; i++ {
		long += "a"
	}
	blocklist := password.NewBuiltinBlocklist()

	err := p.Validate(long, "", blocklist)
	if err != password.ErrPasswordTooLong {
		t.Fatalf("expected ErrPasswordTooLong, got %v", err)
	}
}

func TestPolicySpacesPreserved(t *testing.T) {
	p := password.DefaultPolicy()
	blocklist := password.NewBuiltinBlocklist()

	err := p.Validate("uma senha com espacos 12345", "", blocklist)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPolicyEqualsEmail(t *testing.T) {
	p := password.DefaultPolicy()
	blocklist := password.NewBuiltinBlocklist()

	err := p.Validate("user@example.com", "user@example.com", blocklist)
	if err != password.ErrPasswordEqualsEmail {
		t.Fatalf("expected ErrPasswordEqualsEmail, got %v", err)
	}
}

func TestPolicyCommonPassword(t *testing.T) {
	p := password.DefaultPolicy()
	blocklist := password.NewBuiltinBlocklist()

	err := p.Validate("123456789012345", "", blocklist)
	if err != password.ErrPasswordCommon {
		t.Fatalf("expected ErrPasswordCommon, got %v", err)
	}
}

func TestPolicyValid(t *testing.T) {
	p := password.DefaultPolicy()
	blocklist := password.NewBuiltinBlocklist()

	err := p.Validate("uma senha muito longa e segura 123", "", blocklist)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPolicyCustomBlocklist(t *testing.T) {
	p := password.DefaultPolicy()
	blocklist := password.NewBlocklist([]string{"customblockedpassword12345"})

	err := p.Validate("customblockedpassword12345", "", blocklist)
	if err != password.ErrPasswordCommon {
		t.Fatalf("expected ErrPasswordCommon, got %v", err)
	}
}

func TestPolicyBlocklistFromFile(t *testing.T) {
	_, err := password.NewBlocklistFromFile("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestPolicyBuiltinBlocklist(t *testing.T) {
	blocklist := password.NewBuiltinBlocklist()
	if blocklist == nil {
		t.Fatal("builtin blocklist should not be nil")
	}
}
