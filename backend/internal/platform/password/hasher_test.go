package password_test

import (
	"strings"
	"testing"

	"github.com/gabrielalc23/pdv/internal/platform/password"
)

func TestHashAndVerify(t *testing.T) {
	h, err := password.NewHasher(password.DefaultParams())
	if err != nil {
		t.Fatalf("NewHasher: %v", err)
	}

	pwd := "uma senha bem longa e segura 12345"
	encoded, err := h.Hash(pwd)
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}

	if !strings.HasPrefix(encoded, "$argon2id$") {
		t.Fatalf("expected PHC prefix, got %q", encoded)
	}

	match, needsRehash, err := h.Verify(pwd, encoded)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if !match {
		t.Fatal("expected match")
	}
	if needsRehash {
		t.Fatal("unexpected needsRehash")
	}
}

func TestTwoHashesDiffer(t *testing.T) {
	h, err := password.NewHasher(password.DefaultParams())
	if err != nil {
		t.Fatalf("NewHasher: %v", err)
	}

	pwd := "minha senha super segura 123"
	h1, _ := h.Hash(pwd)
	h2, _ := h.Hash(pwd)

	if h1 == h2 {
		t.Fatal("expected different hashes due to random salt")
	}
}

func TestWrongPassword(t *testing.T) {
	h, err := password.NewHasher(password.DefaultParams())
	if err != nil {
		t.Fatalf("NewHasher: %v", err)
	}

	encoded, _ := h.Hash("senha correta longo 12345")
	match, _, err := h.Verify("senha errada longo 12345!!", encoded)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if match {
		t.Fatal("expected no match")
	}
}

func TestPHCMalformed(t *testing.T) {
	h, err := password.NewHasher(password.DefaultParams())
	if err != nil {
		t.Fatalf("NewHasher: %v", err)
	}

	tests := []struct {
		name string
		hash string
	}{
		{"empty", ""},
		{"no_prefix", "argon2id$v=19$m=8,t=1,p=1$a$b"},
		{"too_few_segments", "$argon2id$v=19$m=8,t=1,p=1"},
		{"too_many_segments", "$argon2id$v=19$m=8,t=1,p=1$a$b$c$d$e"},
		{"bad_algo", "$bcrypt$v=19$m=8,t=1,p=1$a$b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := h.Verify("password", tt.hash)
			if err == nil {
				t.Fatal("expected error for malformed PHC")
			}
		})
	}
}

func TestInvalidAlgorithm(t *testing.T) {
	h, err := password.NewHasher(password.DefaultParams())
	if err != nil {
		t.Fatalf("NewHasher: %v", err)
	}

	_, _, err = h.Verify("pwd", "$argon2i$v=19$m=65536,t=3,p=1$c2FsdHNhbHRzYWx0c2FsdA$cGFzc3dvcmRoYXNo")
	if err == nil || !strings.Contains(err.Error(), "algorithm") {
		t.Fatalf("expected algorithm error, got %v", err)
	}
}

func TestInvalidVersion(t *testing.T) {
	h, err := password.NewHasher(password.DefaultParams())
	if err != nil {
		t.Fatalf("NewHasher: %v", err)
	}

	_, _, err = h.Verify("pwd", "$argon2id$v=18$m=65536,t=3,p=1$c2FsdHNhbHRzYWx0c2FsdA$cGFzc3dvcmRoYXNo")
	if err == nil || !strings.Contains(err.Error(), "version") {
		t.Fatalf("expected version error, got %v", err)
	}
}

func TestExcessiveParams(t *testing.T) {
	h, err := password.NewHasher(password.DefaultParams())
	if err != nil {
		t.Fatalf("NewHasher: %v", err)
	}

	_, _, err = h.Verify("pwd", "$argon2id$v=19$m=9999999999999,t=999,p=999$c2FsdA$cGFzcw")
	if err == nil {
		t.Fatal("expected error for excessive params")
	}
}

func TestNeedsRehash(t *testing.T) {
	weakParams := password.Params{
		MemoryKiB:   8,
		Iterations:  1,
		Parallelism: 1,
		SaltLength:  16,
		KeyLength:   16,
	}
	weakHasher, err := password.NewHasher(weakParams)
	if err != nil {
		t.Fatalf("NewHasher weak: %v", err)
	}

	encoded, _ := weakHasher.Hash("minha senha longa 12345")

	strongHasher, err := password.NewHasher(password.DefaultParams())
	if err != nil {
		t.Fatalf("NewHasher strong: %v", err)
	}

	match, needsRehash, err := strongHasher.Verify("minha senha longa 12345", encoded)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if !match {
		t.Fatal("expected match")
	}
	if !needsRehash {
		t.Fatal("expected needsRehash when upgrading params")
	}
}

func TestDummyHash(t *testing.T) {
	if !password.IsDummyHash(password.DummyHashValue) {
		t.Fatal("expected DummyHashValue to be recognized as dummy")
	}

	if password.IsDummyHash("$argon2id$v=19$m=8,t=1,p=1$a$b") {
		t.Fatal("expected non-dummy hash to be rejected")
	}
}

func TestParametersValidation(t *testing.T) {
	tests := []struct {
		name    string
		params  password.Params
		wantErr bool
	}{
		{"default", password.DefaultParams(), false},
		{"zero_memory", password.Params{MemoryKiB: 0, Iterations: 1, Parallelism: 1, SaltLength: 16, KeyLength: 16}, true},
		{"high_memory", password.Params{MemoryKiB: 1 << 25, Iterations: 1, Parallelism: 1, SaltLength: 16, KeyLength: 16}, true},
		{"zero_iters", password.Params{MemoryKiB: 8, Iterations: 0, Parallelism: 1, SaltLength: 16, KeyLength: 16}, true},
		{"high_parallelism", password.Params{MemoryKiB: 8, Iterations: 1, Parallelism: 0, SaltLength: 16, KeyLength: 16}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := password.NewHasher(tt.params)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewHasher error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestUnicodePassword(t *testing.T) {
	h, err := password.NewHasher(password.DefaultParams())
	if err != nil {
		t.Fatalf("NewHasher: %v", err)
	}

	pwd := "こんにちは世界１２３４５６７８９０"
	encoded, err := h.Hash(pwd)
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}

	match, _, err := h.Verify(pwd, encoded)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if !match {
		t.Fatal("expected Unicode password to match")
	}
}

func TestEmptyHash(t *testing.T) {
	h, err := password.NewHasher(password.DefaultParams())
	if err != nil {
		t.Fatalf("NewHasher: %v", err)
	}

	_, _, err = h.Verify("pwd", "$argon2id$v=19$m=8,t=1,p=1$c2FsdA$")
	if err == nil {
		t.Fatal("expected error for empty hash")
	}
}
