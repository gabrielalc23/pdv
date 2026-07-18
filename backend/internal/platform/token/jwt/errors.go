package jwt

import "errors"

var (
	ErrKeyNotFound       = errors.New("jwt key not found")
	ErrKeyDuplicate      = errors.New("jwt key id already exists")
	ErrKeyInvalid        = errors.New("jwt key is invalid")
	ErrPrivateKeyExposed = errors.New("attempted to expose private key material")
	ErrTokenInvalid      = errors.New("token is invalid")
	ErrTokenExpired      = errors.New("token has expired")
	ErrTokenSignature    = errors.New("token has invalid signature")
	ErrTokenAlgorithm    = errors.New("token uses unsupported algorithm")
	ErrTokenTyp          = errors.New("token has invalid typ header")
	ErrTokenKID          = errors.New("token has invalid or missing kid")
	ErrTokenIssuer       = errors.New("token has invalid issuer")
	ErrTokenAudience     = errors.New("token has invalid audience")
	ErrTokenSize         = errors.New("token exceeds maximum size")
	ErrClaimsInvalid     = errors.New("token claims are invalid")
	ErrClaimsIncoherent  = errors.New("token claims are incoherent with context")
)
