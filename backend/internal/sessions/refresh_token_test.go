package sessions

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

func base64URLEncode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

func TestRefreshTokenCodec_Generate(t *testing.T) {
	hashKey := make([]byte, 32)
	_, _ = rand.Read(hashKey)
	codec := NewRefreshTokenCodec(hashKey)

	var tokenID pgtype.UUID
	_ = tokenID.Scan("550e8400-e29b-41d4-a716-446655440000")

	raw, hash, err := codec.Generate(tokenID)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !strings.HasPrefix(raw, "rt_") {
		t.Fatalf("expected prefix rt_, got %q", raw)
	}

	parts := strings.SplitN(raw[3:], ".", 2)
	if len(parts) != 2 {
		t.Fatalf("expected uuid.secret format, got %q", raw)
	}

	selectorStr, secretB64 := parts[0], parts[1]

	if selectorStr != "550e8400-e29b-41d4-a716-446655440000" {
		t.Fatalf("expected selector 550e8400-e29b-41d4-a716-446655440000, got %q", selectorStr)
	}

	if len(secretB64) != 43 { // 32 bytes in base64url without padding = 43 chars
		t.Fatalf("expected secret base64 length 43, got %d", len(secretB64))
	}

	if len(hash) != 32 {
		t.Fatalf("expected hash length 32, got %d", len(hash))
	}
}

func TestRefreshTokenCodec_Generate_DifferentToken(t *testing.T) {
	hashKey := make([]byte, 32)
	_, _ = rand.Read(hashKey)
	codec := NewRefreshTokenCodec(hashKey)

	var id1, id2 pgtype.UUID
	_ = id1.Scan("550e8400-e29b-41d4-a716-446655440000")
	_ = id2.Scan("660e8400-e29b-41d4-a716-446655440001")

	raw1, hash1, _ := codec.Generate(id1)
	raw2, hash2, _ := codec.Generate(id2)

	if raw1 == raw2 {
		t.Fatal("expected different tokens for different IDs")
	}

	if string(hash1) == string(hash2) {
		t.Fatal("expected different hashes for different tokens")
	}
}

func TestRefreshTokenCodec_Parse_Valid(t *testing.T) {
	hashKey := make([]byte, 32)
	_, _ = rand.Read(hashKey)
	codec := NewRefreshTokenCodec(hashKey)

	var tokenID pgtype.UUID
	_ = tokenID.Scan("550e8400-e29b-41d4-a716-446655440000")

	raw, _, err := codec.Generate(tokenID)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	parsed, err := codec.Parse(raw)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if parsed.Selector != tokenID {
		t.Fatalf("expected selector %v, got %v", tokenID, parsed.Selector)
	}

	if len(parsed.Secret) != 32 {
		t.Fatalf("expected secret length 32, got %d", len(parsed.Secret))
	}
}

func TestRefreshTokenCodec_HashAndVerify(t *testing.T) {
	hashKey := make([]byte, 32)
	_, _ = rand.Read(hashKey)
	codec := NewRefreshTokenCodec(hashKey)

	secret := make([]byte, 32)
	_, _ = rand.Read(secret)

	hash := codec.HashSecret(secret)

	if len(hash) != 32 {
		t.Fatalf("expected hash length 32, got %d", len(hash))
	}

	if !codec.VerifySecret(secret, hash) {
		t.Fatal("VerifySecret should return true for matching secret")
	}

	wrongSecret := make([]byte, 32)
	_, _ = rand.Read(wrongSecret)

	if codec.VerifySecret(wrongSecret, hash) {
		t.Fatal("VerifySecret should return false for wrong secret")
	}

	if codec.VerifySecret(nil, hash) {
		t.Fatal("VerifySecret should return false for nil secret")
	}
}

func TestRefreshTokenCodec_Parse_Empty(t *testing.T) {
	hashKey := make([]byte, 32)
	codec := NewRefreshTokenCodec(hashKey)

	_, err := codec.Parse("")
	if err != ErrRefreshTokenMissing {
		t.Fatalf("expected ErrRefreshTokenMissing, got %v", err)
	}
}

func TestRefreshTokenCodec_Parse_InvalidPrefix(t *testing.T) {
	hashKey := make([]byte, 32)
	codec := NewRefreshTokenCodec(hashKey)

	_, err := codec.Parse("invalid_token")
	if err == nil || !strings.Contains(err.Error(), "invalid prefix") {
		t.Fatalf("expected invalid prefix error, got %v", err)
	}
}

func TestRefreshTokenCodec_Parse_MissingSeparator(t *testing.T) {
	hashKey := make([]byte, 32)
	codec := NewRefreshTokenCodec(hashKey)

	_, err := codec.Parse("rt_uuid")
	if err == nil || !strings.Contains(err.Error(), "missing separator") {
		t.Fatalf("expected missing separator error, got %v", err)
	}
}

func TestRefreshTokenCodec_Parse_EmptySelector(t *testing.T) {
	hashKey := make([]byte, 32)
	codec := NewRefreshTokenCodec(hashKey)

	_, err := codec.Parse("rt_.abcdef")
	if err == nil || !strings.Contains(err.Error(), "empty selector") {
		t.Fatalf("expected empty selector error, got %v", err)
	}
}

func TestRefreshTokenCodec_Parse_InvalidUUID(t *testing.T) {
	hashKey := make([]byte, 32)
	codec := NewRefreshTokenCodec(hashKey)

	_, err := codec.Parse("rt_not-a-uuid.secret")
	if err == nil || !strings.Contains(err.Error(), "invalid selector") {
		t.Fatalf("expected invalid selector error, got %v", err)
	}
}

func TestRefreshTokenCodec_Parse_EmptySecret(t *testing.T) {
	hashKey := make([]byte, 32)
	codec := NewRefreshTokenCodec(hashKey)

	_, err := codec.Parse("rt_550e8400-e29b-41d4-a716-446655440000.")
	if err == nil || !strings.Contains(err.Error(), "empty secret") {
		t.Fatalf("expected empty secret error, got %v", err)
	}
}

func TestRefreshTokenCodec_Parse_InvalidBase64(t *testing.T) {
	hashKey := make([]byte, 32)
	codec := NewRefreshTokenCodec(hashKey)

	_, err := codec.Parse("rt_550e8400-e29b-41d4-a716-446655440000.!!!invalid!!!")
	if err == nil || !strings.Contains(err.Error(), "invalid secret encoding") {
		t.Fatalf("expected invalid secret encoding error, got %v", err)
	}
}

func TestRefreshTokenCodec_Parse_Padding(t *testing.T) {
	hashKey := make([]byte, 32)
	codec := NewRefreshTokenCodec(hashKey)

	_, err := codec.Parse("rt_550e8400-e29b-41d4-a716-446655440000.dGhpcyBpcyBhIHRlc3Q=")
	if err == nil || !strings.Contains(err.Error(), "padding") {
		t.Fatalf("expected padding error, got %v", err)
	}
}

func TestRefreshTokenCodec_Parse_ShortSecret(t *testing.T) {
	hashKey := make([]byte, 32)
	codec := NewRefreshTokenCodec(hashKey)

	shortB64 := "dGhpcyBpcyBhIHRlc3Q" // 16 bytes, not 32
	_, err := codec.Parse("rt_550e8400-e29b-41d4-a716-446655440000." + shortB64)
	if err == nil || !strings.Contains(err.Error(), "invalid secret length") {
		t.Fatalf("expected invalid secret length error, got %v", err)
	}
}

func TestRefreshTokenCodec_Parse_LongSecret(t *testing.T) {
	hashKey := make([]byte, 32)
	codec := NewRefreshTokenCodec(hashKey)

	longBytes := make([]byte, 64)
	_, _ = rand.Read(longBytes)
	longB64 := strings.TrimRight(base64URLEncode(longBytes), "=")
	_, err := codec.Parse("rt_550e8400-e29b-41d4-a716-446655440000." + longB64)
	if err == nil || !strings.Contains(err.Error(), "invalid secret length") {
		t.Fatalf("expected invalid secret length error, got %v", err)
	}
}

func TestRefreshTokenCodec_TokenTooLong(t *testing.T) {
	hashKey := make([]byte, 32)
	codec := NewRefreshTokenCodec(hashKey)

	longToken := "rt_" + strings.Repeat("a", 600) + "." + strings.Repeat("b", 100)
	_, err := codec.Parse(longToken)
	if err == nil || !strings.Contains(err.Error(), "too long") {
		t.Fatalf("expected too long error, got %v", err)
	}
}

func TestRefreshTokenCodec_ConstantTimeCompare(t *testing.T) {
	hashKey := make([]byte, 32)
	_, _ = rand.Read(hashKey)
	codec := NewRefreshTokenCodec(hashKey)

	s1 := []byte("same_secret_32_bytes_long_for_test!")
	s2 := []byte("same_secret_32_bytes_long_for_test!")
	s3 := []byte("different_secret_32_bytes_long_for")

	h := codec.HashSecret(s1)

	if !codec.VerifySecret(s1, h) {
		t.Fatal("same hash should match")
	}
	if !codec.VerifySecret(s2, h) {
		t.Fatal("identical secret should match")
	}
	if codec.VerifySecret(s3, h) {
		t.Fatal("different secret should not match")
	}
}

func TestRefreshTokenCodec_RandFailure(t *testing.T) {
	hashKey := make([]byte, 32)
	failingReader := &failReader{}
	codec := NewRefreshTokenCodecWithRand(hashKey, failingReader)

	var id pgtype.UUID
	_ = id.Scan("550e8400-e29b-41d4-a716-446655440000")

	_, _, err := codec.Generate(id)
	if err == nil {
		t.Fatal("expected error from failing rand reader")
	}
}

type failReader struct{}

func (r *failReader) Read(p []byte) (int, error) {
	return 0, failReaderErr
}

var failReaderErr = &failReaderError{}

type failReaderError struct{}

func (e *failReaderError) Error() string { return "test read error" }
