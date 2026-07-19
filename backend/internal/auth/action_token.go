package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/platform/database"
)

const (
	actionTokenSecretBytes = 32
	maxActionTokenLength   = 512
)

type ActionTokenPurpose = database.AuthActionTokenPurpose

const (
	ActionTokenPurposeEmailVerification = database.AuthActionTokenPurposeEMAILVERIFICATION
	ActionTokenPurposePasswordReset     = database.AuthActionTokenPurposePASSWORDRESET
)

var (
	ErrActionTokenMissing = errors.New("action token is missing")
	ErrActionTokenInvalid = errors.New("action token is invalid")
)

type ParsedActionToken struct {
	Purpose  ActionTokenPurpose
	Selector pgtype.UUID
	Secret   []byte
}

type ActionTokenCodec interface {
	Generate(purpose ActionTokenPurpose, selector pgtype.UUID) (rawToken string, secretHash []byte, err error)
	Parse(rawToken string, expectedPurpose ActionTokenPurpose) (ParsedActionToken, error)
	HashSecret(secret []byte) []byte
	VerifySecret(secret, expectedHash []byte) bool
}

type actionTokenCodec struct {
	hashKey []byte
	randSrc io.Reader
}

func NewActionTokenCodec(hashKey []byte) (ActionTokenCodec, error) {
	return NewActionTokenCodecWithRand(hashKey, rand.Reader)
}

func NewActionTokenCodecWithRand(hashKey []byte, randSrc io.Reader) (ActionTokenCodec, error) {
	if len(hashKey) < sha256.Size {
		return nil, fmt.Errorf("action token HMAC key must be at least %d bytes", sha256.Size)
	}
	if randSrc == nil {
		return nil, fmt.Errorf("action token random source is required")
	}

	return &actionTokenCodec{
		hashKey: append([]byte(nil), hashKey...),
		randSrc: randSrc,
	}, nil
}

func (c *actionTokenCodec) Generate(purpose ActionTokenPurpose, selector pgtype.UUID) (string, []byte, error) {
	prefix, err := actionTokenPrefix(purpose)
	if err != nil {
		return "", nil, err
	}

	selectorString := uuidString(selector)
	if selectorString == "" {
		return "", nil, fmt.Errorf("invalid action token selector")
	}

	secret := make([]byte, actionTokenSecretBytes)
	if _, err := io.ReadFull(c.randSrc, secret); err != nil {
		return "", nil, fmt.Errorf("generate action token secret: %w", err)
	}

	rawToken := prefix + selectorString + "." + base64.RawURLEncoding.EncodeToString(secret)
	if len(rawToken) > maxActionTokenLength {
		return "", nil, fmt.Errorf("generated action token exceeds maximum length")
	}

	return rawToken, c.HashSecret(secret), nil
}

func (c *actionTokenCodec) Parse(rawToken string, expectedPurpose ActionTokenPurpose) (ParsedActionToken, error) {
	if rawToken == "" {
		return ParsedActionToken{}, ErrActionTokenMissing
	}
	if len(rawToken) > maxActionTokenLength {
		return ParsedActionToken{}, invalidActionToken("token too long", nil)
	}
	if strings.IndexFunc(rawToken, unicode.IsSpace) >= 0 {
		return ParsedActionToken{}, invalidActionToken("whitespace is not allowed", nil)
	}

	prefix, err := actionTokenPrefix(expectedPurpose)
	if err != nil {
		return ParsedActionToken{}, invalidActionToken("invalid purpose", err)
	}
	if !strings.HasPrefix(rawToken, prefix) {
		return ParsedActionToken{}, invalidActionToken("wrong purpose or prefix", nil)
	}

	payload := rawToken[len(prefix):]
	if strings.Count(payload, ".") != 1 {
		return ParsedActionToken{}, invalidActionToken("token must contain exactly one separator", nil)
	}
	selectorString, encodedSecret, _ := strings.Cut(payload, ".")

	var selector pgtype.UUID
	if len(selectorString) != 36 || selector.Scan(selectorString) != nil || uuidString(selector) != selectorString {
		return ParsedActionToken{}, invalidActionToken("invalid or noncanonical selector", nil)
	}
	if encodedSecret == "" || strings.Contains(encodedSecret, "=") {
		return ParsedActionToken{}, invalidActionToken("invalid secret encoding", nil)
	}

	secret, err := base64.RawURLEncoding.DecodeString(encodedSecret)
	if err != nil || base64.RawURLEncoding.EncodeToString(secret) != encodedSecret {
		return ParsedActionToken{}, invalidActionToken("invalid or noncanonical secret encoding", err)
	}
	if len(secret) != actionTokenSecretBytes {
		return ParsedActionToken{}, invalidActionToken("invalid secret length", nil)
	}

	return ParsedActionToken{Purpose: expectedPurpose, Selector: selector, Secret: secret}, nil
}

func (c *actionTokenCodec) HashSecret(secret []byte) []byte {
	mac := hmac.New(sha256.New, c.hashKey)
	_, _ = mac.Write(secret)
	return mac.Sum(nil)
}

func (c *actionTokenCodec) VerifySecret(secret, expectedHash []byte) bool {
	actualHash := c.HashSecret(secret)
	var fixedExpected [sha256.Size]byte
	copy(fixedExpected[:], expectedHash)

	validLengths := subtle.ConstantTimeEq(int32(len(secret)), actionTokenSecretBytes) &
		subtle.ConstantTimeEq(int32(len(expectedHash)), sha256.Size)
	return subtle.ConstantTimeCompare(actualHash, fixedExpected[:])&validLengths == 1
}

func actionTokenPrefix(purpose ActionTokenPurpose) (string, error) {
	switch purpose {
	case ActionTokenPurposeEmailVerification:
		return "evt_", nil
	case ActionTokenPurposePasswordReset:
		return "prt_", nil
	default:
		return "", fmt.Errorf("unsupported action token purpose %q", purpose)
	}
}

func invalidActionToken(reason string, cause error) error {
	if cause != nil {
		return fmt.Errorf("%w: %s: %w", ErrActionTokenInvalid, reason, cause)
	}
	return fmt.Errorf("%w: %s", ErrActionTokenInvalid, reason)
}
