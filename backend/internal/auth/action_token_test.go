package auth

import (
	"bytes"
	"encoding/base64"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestActionTokenCodecGenerateAndParse(t *testing.T) {
	selector := actionTokenTestUUID(t, "550e8400-e29b-41d4-a716-446655440000")
	secret := bytes.Repeat([]byte{0xa5}, actionTokenSecretBytes)

	for _, test := range []struct {
		name    string
		purpose ActionTokenPurpose
		prefix  string
	}{
		{name: "email verification", purpose: ActionTokenPurposeEmailVerification, prefix: "evt_"},
		{name: "password reset", purpose: ActionTokenPurposePasswordReset, prefix: "prt_"},
	} {
		t.Run(test.name, func(t *testing.T) {
			codec := newActionTokenTestCodec(t, bytes.NewReader(secret))
			rawToken, secretHash, err := codec.Generate(test.purpose, selector)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			expectedToken := test.prefix + "550e8400-e29b-41d4-a716-446655440000." + base64.RawURLEncoding.EncodeToString(secret)
			if rawToken != expectedToken {
				t.Fatal("Generate() returned an unexpected token encoding")
			}
			if strings.Contains(rawToken, "=") || len(secretHash) != 32 {
				t.Fatalf("Generate() returned padding or invalid hash length %d", len(secretHash))
			}

			parsed, err := codec.Parse(rawToken, test.purpose)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if parsed.Purpose != test.purpose || parsed.Selector != selector || !bytes.Equal(parsed.Secret, secret) {
				t.Fatal("Parse() returned unexpected token components")
			}
			if !codec.VerifySecret(parsed.Secret, secretHash) {
				t.Fatal("VerifySecret() = false, want true")
			}
		})
	}
}

func TestActionTokenCodecConstructors(t *testing.T) {
	if _, err := NewActionTokenCodec(bytes.Repeat([]byte{1}, 31)); err == nil {
		t.Fatal("NewActionTokenCodec() accepted a key shorter than 32 bytes")
	}
	if _, err := NewActionTokenCodec(bytes.Repeat([]byte{1}, 32)); err != nil {
		t.Fatalf("NewActionTokenCodec() rejected a 32-byte key: %v", err)
	}
	if _, err := NewActionTokenCodec(bytes.Repeat([]byte{1}, 64)); err != nil {
		t.Fatalf("NewActionTokenCodec() rejected a longer key: %v", err)
	}
	if _, err := NewActionTokenCodecWithRand(bytes.Repeat([]byte{1}, 32), nil); err == nil {
		t.Fatal("NewActionTokenCodecWithRand() accepted a nil reader")
	}
}

func TestActionTokenCodecReadsExactly32RandomBytes(t *testing.T) {
	reader := &countingActionTokenReader{data: bytes.Repeat([]byte{7}, 64)}
	codec := newActionTokenTestCodec(t, reader)
	selector := actionTokenTestUUID(t, "550e8400-e29b-41d4-a716-446655440000")

	if _, _, err := codec.Generate(ActionTokenPurposeEmailVerification, selector); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if reader.bytesRead != actionTokenSecretBytes {
		t.Fatalf("Generate() read %d random bytes, want %d", reader.bytesRead, actionTokenSecretBytes)
	}
}

func TestActionTokenCodecGenerateRejectsInvalidInputAndReaderFailure(t *testing.T) {
	codec := newActionTokenTestCodec(t, bytes.NewReader(bytes.Repeat([]byte{1}, 32)))
	selector := actionTokenTestUUID(t, "550e8400-e29b-41d4-a716-446655440000")

	if _, _, err := codec.Generate(ActionTokenPurpose("INVITATION"), selector); err == nil {
		t.Fatal("Generate() accepted an unsupported purpose")
	}
	if _, _, err := codec.Generate(ActionTokenPurposeEmailVerification, pgtype.UUID{}); err == nil {
		t.Fatal("Generate() accepted an invalid selector")
	}

	wantErr := errors.New("random unavailable")
	failingCodec := newActionTokenTestCodec(t, actionTokenFailReader{err: wantErr})
	if _, _, err := failingCodec.Generate(ActionTokenPurposeEmailVerification, selector); !errors.Is(err, wantErr) {
		t.Fatalf("Generate() error = %v, want wrapped %v", err, wantErr)
	}

	shortCodec := newActionTokenTestCodec(t, bytes.NewReader(make([]byte, 31)))
	if _, _, err := shortCodec.Generate(ActionTokenPurposeEmailVerification, selector); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("Generate() short-reader error = %v, want %v", err, io.ErrUnexpectedEOF)
	}
}

func TestActionTokenCodecParseRejectsInvalidTokens(t *testing.T) {
	codec := newActionTokenTestCodec(t, bytes.NewReader(make([]byte, 32)))
	secret := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{2}, 32))
	valid := "evt_550e8400-e29b-41d4-a716-446655440000." + secret

	tests := []struct {
		name    string
		token   string
		purpose ActionTokenPurpose
		missing bool
	}{
		{name: "missing", token: "", purpose: ActionTokenPurposeEmailVerification, missing: true},
		{name: "too long", token: valid + strings.Repeat("a", maxActionTokenLength), purpose: ActionTokenPurposeEmailVerification},
		{name: "leading whitespace", token: " " + valid, purpose: ActionTokenPurposeEmailVerification},
		{name: "embedded unicode whitespace", token: strings.Replace(valid, ".", ".\u00a0", 1), purpose: ActionTokenPurposeEmailVerification},
		{name: "wrong purpose", token: valid, purpose: ActionTokenPurposePasswordReset},
		{name: "wrong prefix", token: "bad_" + valid[4:], purpose: ActionTokenPurposeEmailVerification},
		{name: "unsupported expected purpose", token: valid, purpose: ActionTokenPurpose("INVITATION")},
		{name: "missing separator", token: strings.Replace(valid, ".", "", 1), purpose: ActionTokenPurposeEmailVerification},
		{name: "extra separator", token: valid + ".extra", purpose: ActionTokenPurposeEmailVerification},
		{name: "invalid UUID", token: "evt_not-a-uuid." + secret, purpose: ActionTokenPurposeEmailVerification},
		{name: "uppercase UUID", token: "evt_550E8400-E29B-41D4-A716-446655440000." + secret, purpose: ActionTokenPurposeEmailVerification},
		{name: "compact UUID", token: "evt_550e8400e29b41d4a716446655440000." + secret, purpose: ActionTokenPurposeEmailVerification},
		{name: "empty secret", token: "evt_550e8400-e29b-41d4-a716-446655440000.", purpose: ActionTokenPurposeEmailVerification},
		{name: "padding", token: valid + "=", purpose: ActionTokenPurposeEmailVerification},
		{name: "standard base64 alphabet", token: "evt_550e8400-e29b-41d4-a716-446655440000." + strings.Repeat("/", 43), purpose: ActionTokenPurposeEmailVerification},
		{name: "malformed base64", token: "evt_550e8400-e29b-41d4-a716-446655440000.!not-base64!", purpose: ActionTokenPurposeEmailVerification},
		{name: "noncanonical base64", token: "evt_550e8400-e29b-41d4-a716-446655440000." + secret[:42] + "x", purpose: ActionTokenPurposeEmailVerification},
		{name: "short secret", token: "evt_550e8400-e29b-41d4-a716-446655440000." + base64.RawURLEncoding.EncodeToString(make([]byte, 31)), purpose: ActionTokenPurposeEmailVerification},
		{name: "long secret", token: "evt_550e8400-e29b-41d4-a716-446655440000." + base64.RawURLEncoding.EncodeToString(make([]byte, 33)), purpose: ActionTokenPurposeEmailVerification},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := codec.Parse(test.token, test.purpose)
			if test.missing {
				if !errors.Is(err, ErrActionTokenMissing) {
					t.Fatalf("Parse() error = %v, want ErrActionTokenMissing", err)
				}
				return
			}
			if !errors.Is(err, ErrActionTokenInvalid) {
				t.Fatalf("Parse() error = %v, want ErrActionTokenInvalid", err)
			}
		})
	}
}

func TestActionTokenCodecVerifySecret(t *testing.T) {
	codec := newActionTokenTestCodec(t, bytes.NewReader(make([]byte, 32)))
	secret := bytes.Repeat([]byte{3}, 32)
	hash := codec.HashSecret(secret)

	if !codec.VerifySecret(secret, hash) {
		t.Fatal("VerifySecret() rejected matching secret and hash")
	}
	wrongSecret := append([]byte(nil), secret...)
	wrongSecret[31] ^= 1
	if codec.VerifySecret(wrongSecret, hash) {
		t.Fatal("VerifySecret() accepted a different secret")
	}
	if codec.VerifySecret(secret[:31], hash) {
		t.Fatal("VerifySecret() accepted a short secret")
	}
	if codec.VerifySecret(secret, hash[:31]) {
		t.Fatal("VerifySecret() accepted a short hash")
	}
}

func newActionTokenTestCodec(t *testing.T, reader io.Reader) ActionTokenCodec {
	t.Helper()
	codec, err := NewActionTokenCodecWithRand(bytes.Repeat([]byte{0x5c}, 32), reader)
	if err != nil {
		t.Fatalf("NewActionTokenCodecWithRand() error = %v", err)
	}
	return codec
}

func actionTokenTestUUID(t *testing.T, value string) pgtype.UUID {
	t.Helper()
	var id pgtype.UUID
	if err := id.Scan(value); err != nil {
		t.Fatalf("scan UUID: %v", err)
	}
	return id
}

type countingActionTokenReader struct {
	data      []byte
	bytesRead int
}

func (r *countingActionTokenReader) Read(p []byte) (int, error) {
	n := copy(p, r.data[r.bytesRead:])
	r.bytesRead += n
	return n, nil
}

type actionTokenFailReader struct {
	err error
}

func (r actionTokenFailReader) Read([]byte) (int, error) {
	return 0, r.err
}
