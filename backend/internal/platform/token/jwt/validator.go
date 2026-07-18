package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Validator struct {
	keyring   *Keyring
	issuer    string
	audience  string
	clockSkew time.Duration
}

func NewValidator(keyring *Keyring, issuer, audience string, clockSkew time.Duration) *Validator {
	return &Validator{
		keyring:   keyring,
		issuer:    issuer,
		audience:  audience,
		clockSkew: clockSkew,
	}
}

func (v *Validator) Validate(tokenStr string) (*Claims, error) {
	if len(tokenStr) > MaxTokenSize {
		return nil, fmt.Errorf("%w: token exceeds %d bytes", ErrTokenSize, MaxTokenSize)
	}

	claims := &Claims{}

	keyFunc := func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != SigningAlgorithm {
			return nil, fmt.Errorf("%w: expected %s, got %s", ErrTokenAlgorithm, SigningAlgorithm, t.Method.Alg())
		}

		kid, ok := t.Header["kid"].(string)
		if !ok || kid == "" {
			return nil, fmt.Errorf("%w: missing kid", ErrTokenKID)
		}

		pubKey, exists := v.keyring.PublicKey(kid)
		if !exists {
			return nil, fmt.Errorf("%w: unknown kid %q", ErrTokenKID, kid)
		}

		return pubKey, nil
	}

	parsed, err := jwt.ParseWithClaims(tokenStr, claims, keyFunc,
		jwt.WithValidMethods([]string{SigningAlgorithm}),
		jwt.WithIssuer(v.issuer),
		jwt.WithAudience(v.audience),
		jwt.WithLeeway(v.clockSkew),
		jwt.WithIssuedAt(),
	)

	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, ErrTokenExpired
		case errors.Is(err, jwt.ErrTokenSignatureInvalid):
			return nil, ErrTokenSignature
		case errors.Is(err, jwt.ErrTokenInvalidIssuer):
			return nil, ErrTokenIssuer
		case errors.Is(err, jwt.ErrTokenInvalidAudience):
			return nil, ErrTokenAudience
		default:
			return nil, fmt.Errorf("%w: %w", ErrTokenInvalid, err)
		}
	}

	var ok bool
	claims, ok = parsed.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("%w: failed to extract claims", ErrTokenInvalid)
	}

	typ, ok := parsed.Header["typ"].(string)
	if !ok || typ != TokenType {
		return nil, fmt.Errorf("%w: expected %s, got %q", ErrTokenTyp, TokenType, typ)
	}

	if err := claims.Validate(); err != nil {
		return nil, err
	}

	return claims, nil
}
