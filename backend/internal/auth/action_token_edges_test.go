package auth

import (
	"bytes"
	"strings"
	"testing"
)

func TestActionTokenCodecAcceptsRawURLAlphabet(t *testing.T) {
	t.Parallel()

	secret := bytes.Repeat([]byte{0xfb, 0xff, 0xff}, 11)[:actionTokenSecretBytes]
	codec := newActionTokenTestCodec(t, bytes.NewReader(secret))
	selector := actionTokenTestUUID(t, "550e8400-e29b-41d4-a716-446655440000")
	rawToken, _, err := codec.Generate(ActionTokenPurposePasswordReset, selector)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	encodedSecret := strings.Split(rawToken, ".")[1]
	if !strings.Contains(encodedSecret, "-") || !strings.Contains(encodedSecret, "_") {
		t.Fatal("test secret did not exercise both URL-safe characters")
	}
	parsed, err := codec.Parse(rawToken, ActionTokenPurposePasswordReset)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if !bytes.Equal(parsed.Secret, secret) {
		t.Fatal("Parse() returned an unexpected secret")
	}
}

func TestActionTokenCodecCopiesHashKey(t *testing.T) {
	t.Parallel()

	key := bytes.Repeat([]byte{0x41}, 32)
	codec, err := NewActionTokenCodecWithRand(key, bytes.NewReader(make([]byte, 32)))
	if err != nil {
		t.Fatalf("NewActionTokenCodecWithRand() error = %v", err)
	}
	secret := bytes.Repeat([]byte{0x24}, actionTokenSecretBytes)
	want := codec.HashSecret(secret)
	for i := range key {
		key[i] ^= 0xff
	}
	if got := codec.HashSecret(secret); !bytes.Equal(got, want) {
		t.Fatalf("HashSecret() changed after caller mutated constructor key: got %x, want %x", got, want)
	}
}

func TestActionTokenCodecVerifySecretRejectsNilAndOversizedInputs(t *testing.T) {
	t.Parallel()

	codec := newActionTokenTestCodec(t, bytes.NewReader(make([]byte, 32)))
	secret := bytes.Repeat([]byte{3}, actionTokenSecretBytes)
	hash := codec.HashSecret(secret)
	for _, test := range []struct {
		name   string
		secret []byte
		hash   []byte
	}{
		{name: "nil secret", hash: hash},
		{name: "nil hash", secret: secret},
		{name: "oversized secret", secret: append(append([]byte(nil), secret...), 0), hash: hash},
		{name: "oversized hash", secret: secret, hash: append(append([]byte(nil), hash...), 0)},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if codec.VerifySecret(test.secret, test.hash) {
				t.Fatal("VerifySecret() accepted invalid input lengths")
			}
		})
	}
}
