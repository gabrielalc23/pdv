package invitations

import (
	"bytes"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestTokenCodecRoundTripStoresOnlyHMAC(t *testing.T) {
	codec, err := newTokenCodec(bytes.Repeat([]byte{0x42}, 32), bytes.NewReader(bytes.Repeat([]byte{0x17}, 32)))
	if err != nil {
		t.Fatal(err)
	}
	selector := testUUID(t, "018f2f9a-8d4b-7f35-8b31-84b75f216456")
	raw, hash, err := codec.Generate(selector)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(hash, []byte(raw)) || len(hash) != 32 {
		t.Fatalf("persisted value must be a 32-byte HMAC, got %d bytes", len(hash))
	}
	parsed, err := codec.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Selector != selector || !codec.Verify(parsed.Secret, hash) {
		t.Fatal("generated invitation did not verify")
	}
	if codec.Verify(bytes.Repeat([]byte{0x18}, 32), hash) {
		t.Fatal("different secret verified")
	}
}

func TestTokenCodecRotationInvalidatesOldSecret(t *testing.T) {
	random := append(bytes.Repeat([]byte{0x11}, 32), bytes.Repeat([]byte{0x22}, 32)...)
	codec, err := newTokenCodec(bytes.Repeat([]byte{0x42}, 32), bytes.NewReader(random))
	if err != nil {
		t.Fatal(err)
	}
	selector := testUUID(t, "018f2f9a-8d4b-7f35-8b31-84b75f216456")
	oldRaw, _, err := codec.Generate(selector)
	if err != nil {
		t.Fatal(err)
	}
	newRaw, newHash, err := codec.Generate(selector)
	if err != nil {
		t.Fatal(err)
	}
	oldParsed, _ := codec.Parse(oldRaw)
	newParsed, _ := codec.Parse(newRaw)
	if codec.Verify(oldParsed.Secret, newHash) {
		t.Fatal("rotated hash accepted the old secret")
	}
	if !codec.Verify(newParsed.Secret, newHash) {
		t.Fatal("rotated hash rejected the new secret")
	}
}

func TestTokenCodecRejectsNonCanonicalTokens(t *testing.T) {
	codec, err := newTokenCodec(bytes.Repeat([]byte{0x42}, 32), bytes.NewReader(bytes.Repeat([]byte{0x17}, 32)))
	if err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{"", "inv_bad", "INV_018f2f9a-8d4b-7f35-8b31-84b75f216456.secret", "inv_018f2f9a-8d4b-7f35-8b31-84b75f216456.AA==", " inv_018f2f9a-8d4b-7f35-8b31-84b75f216456.AA"} {
		if _, err := codec.Parse(raw); !errors.Is(err, ErrInvalidToken) {
			t.Fatalf("Parse(%q) error = %v", raw, err)
		}
	}
}

func testUUID(t *testing.T, value string) pgtype.UUID {
	t.Helper()
	var id pgtype.UUID
	if err := id.Scan(value); err != nil {
		t.Fatal(err)
	}
	return id
}
