package sessions

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
)

const (
	refreshTokenPrefix      = "rt_"
	refreshTokenSeparator   = "."
	refreshTokenSecretBytes = 32
	maxRawTokenLength       = 512
)

type ParsedRefreshToken struct {
	Selector pgtype.UUID
	Secret   []byte
}

type RefreshTokenCodec interface {
	Generate(tokenID pgtype.UUID) (rawToken string, secretHash []byte, err error)
	Parse(rawToken string) (ParsedRefreshToken, error)
	HashSecret(secret []byte) []byte
	VerifySecret(secret, expectedHash []byte) bool
}

type refreshTokenCodec struct {
	hashKey []byte
	randSrc io.Reader
}

func NewRefreshTokenCodec(hashKey []byte) RefreshTokenCodec {
	return &refreshTokenCodec{
		hashKey: hashKey,
		randSrc: rand.Reader,
	}
}

func NewRefreshTokenCodecWithRand(hashKey []byte, randSrc io.Reader) RefreshTokenCodec {
	return &refreshTokenCodec{
		hashKey: hashKey,
		randSrc: randSrc,
	}
}

func (c *refreshTokenCodec) Generate(tokenID pgtype.UUID) (string, []byte, error) {
	secret := make([]byte, refreshTokenSecretBytes)
	if _, err := io.ReadFull(c.randSrc, secret); err != nil {
		return "", nil, fmt.Errorf("failed to generate refresh token secret: %w", err)
	}

	uuid := uuidStr(tokenID)
	if uuid == "" {
		return "", nil, fmt.Errorf("invalid token ID")
	}

	secretB64 := base64.RawURLEncoding.EncodeToString(secret)

	rawToken := refreshTokenPrefix + uuid + refreshTokenSeparator + secretB64

	if len(rawToken) > maxRawTokenLength {
		return "", nil, fmt.Errorf("generated token exceeds maximum length")
	}

	secretHash := c.HashSecret(secret)

	return rawToken, secretHash, nil
}

func (c *refreshTokenCodec) Parse(rawToken string) (ParsedRefreshToken, error) {
	if rawToken == "" {
		return ParsedRefreshToken{}, ErrRefreshTokenMissing
	}

	if len(rawToken) > maxRawTokenLength {
		return ParsedRefreshToken{}, fmt.Errorf("%w: token too long", ErrRefreshTokenInvalid)
	}

	if !strings.HasPrefix(rawToken, refreshTokenPrefix) {
		return ParsedRefreshToken{}, fmt.Errorf("%w: invalid prefix", ErrRefreshTokenInvalid)
	}

	withoutPrefix := rawToken[len(refreshTokenPrefix):]

	parts := strings.SplitN(withoutPrefix, refreshTokenSeparator, 2)
	if len(parts) != 2 {
		return ParsedRefreshToken{}, fmt.Errorf("%w: missing separator", ErrRefreshTokenInvalid)
	}

	selectorStr, secretB64 := parts[0], parts[1]

	if selectorStr == "" {
		return ParsedRefreshToken{}, fmt.Errorf("%w: empty selector", ErrRefreshTokenInvalid)
	}

	var selector pgtype.UUID
	if err := selector.Scan(selectorStr); err != nil {
		return ParsedRefreshToken{}, fmt.Errorf("%w: invalid selector: %w", ErrRefreshTokenInvalid, err)
	}

	if secretB64 == "" {
		return ParsedRefreshToken{}, fmt.Errorf("%w: empty secret", ErrRefreshTokenInvalid)
	}

	if strings.ContainsAny(secretB64, "=") {
		return ParsedRefreshToken{}, fmt.Errorf("%w: secret contains padding", ErrRefreshTokenInvalid)
	}

	secret, err := base64.RawURLEncoding.DecodeString(secretB64)
	if err != nil {
		return ParsedRefreshToken{}, fmt.Errorf("%w: invalid secret encoding: %w", ErrRefreshTokenInvalid, err)
	}

	if len(secret) != refreshTokenSecretBytes {
		return ParsedRefreshToken{}, fmt.Errorf("%w: invalid secret length", ErrRefreshTokenInvalid)
	}

	return ParsedRefreshToken{
		Selector: selector,
		Secret:   secret,
	}, nil
}

func (c *refreshTokenCodec) HashSecret(secret []byte) []byte {
	mac := hmac.New(sha256.New, c.hashKey)
	mac.Write(secret)
	return mac.Sum(nil)
}

func (c *refreshTokenCodec) VerifySecret(secret, expectedHash []byte) bool {
	actualHash := c.HashSecret(secret)
	return subtle.ConstantTimeCompare(actualHash, expectedHash) == 1
}
