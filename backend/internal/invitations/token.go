package invitations

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/jackc/pgx/v5/pgtype"
)

const (
	invitationPrefix      = "inv_"
	invitationSecretBytes = 32
	maxTokenLength        = 512
)

type ParsedToken struct {
	Selector pgtype.UUID
	Secret   []byte
}

type TokenCodec interface {
	Generate(selector pgtype.UUID) (raw string, hash []byte, err error)
	Prepare() (secret []byte, hash []byte, err error)
	Format(selector pgtype.UUID, secret []byte) (string, error)
	Parse(raw string) (ParsedToken, error)
	Verify(secret, expectedHash []byte) bool
}

type tokenCodec struct {
	key    []byte
	random io.Reader
}

func NewTokenCodec(key []byte) (TokenCodec, error) {
	return newTokenCodec(key, rand.Reader)
}

func newTokenCodec(key []byte, random io.Reader) (TokenCodec, error) {
	if len(key) < sha256.Size {
		return nil, fmt.Errorf("invitation token HMAC key must be at least %d bytes", sha256.Size)
	}
	if random == nil {
		return nil, fmt.Errorf("invitation token random source is required")
	}
	return &tokenCodec{key: append([]byte(nil), key...), random: random}, nil
}

func (c *tokenCodec) Generate(selector pgtype.UUID) (string, []byte, error) {
	secret, hash, err := c.Prepare()
	if err != nil {
		return "", nil, err
	}
	raw, err := c.Format(selector, secret)
	if err != nil {
		return "", nil, err
	}
	return raw, hash, nil
}

func (c *tokenCodec) Prepare() ([]byte, []byte, error) {
	secret := make([]byte, invitationSecretBytes)
	if _, err := io.ReadFull(c.random, secret); err != nil {
		return nil, nil, fmt.Errorf("generate invitation secret: %w", err)
	}
	return secret, c.hash(secret), nil
}

func (c *tokenCodec) Format(selector pgtype.UUID, secret []byte) (string, error) {
	if !validUUID(selector) {
		return "", fmt.Errorf("invalid invitation selector")
	}
	if len(secret) != invitationSecretBytes {
		return "", fmt.Errorf("invalid invitation secret")
	}
	raw := invitationPrefix + selector.String() + "." + base64.RawURLEncoding.EncodeToString(secret)
	return raw, nil
}

func (c *tokenCodec) Parse(raw string) (ParsedToken, error) {
	if raw == "" || len(raw) > maxTokenLength || strings.IndexFunc(raw, unicode.IsSpace) >= 0 || !strings.HasPrefix(raw, invitationPrefix) {
		return ParsedToken{}, ErrInvalidToken
	}
	payload := strings.TrimPrefix(raw, invitationPrefix)
	if strings.Count(payload, ".") != 1 {
		return ParsedToken{}, ErrInvalidToken
	}
	selectorText, encodedSecret, _ := strings.Cut(payload, ".")
	var selector pgtype.UUID
	if len(selectorText) != 36 || selector.Scan(selectorText) != nil || !validUUID(selector) || selector.String() != selectorText {
		return ParsedToken{}, ErrInvalidToken
	}
	if encodedSecret == "" || strings.Contains(encodedSecret, "=") {
		return ParsedToken{}, ErrInvalidToken
	}
	secret, err := base64.RawURLEncoding.DecodeString(encodedSecret)
	if err != nil || len(secret) != invitationSecretBytes || base64.RawURLEncoding.EncodeToString(secret) != encodedSecret {
		return ParsedToken{}, ErrInvalidToken
	}
	return ParsedToken{Selector: selector, Secret: secret}, nil
}

func (c *tokenCodec) Verify(secret, expectedHash []byte) bool {
	actual := c.hash(secret)
	var expected [sha256.Size]byte
	copy(expected[:], expectedHash)
	validLengths := subtle.ConstantTimeEq(int32(len(secret)), invitationSecretBytes) & subtle.ConstantTimeEq(int32(len(expectedHash)), sha256.Size)
	return subtle.ConstantTimeCompare(actual, expected[:])&validLengths == 1
}

func (c *tokenCodec) hash(secret []byte) []byte {
	mac := hmac.New(sha256.New, c.key)
	_, _ = mac.Write(secret)
	return mac.Sum(nil)
}
