package password

import (
	"errors"
	"strings"
	"unicode/utf8"
)

var (
	ErrPasswordTooShort    = errors.New("password must be at least 15 characters")
	ErrPasswordTooLong     = errors.New("password must be at most 128 characters")
	ErrPasswordCommon      = errors.New("password is too common")
	ErrPasswordEqualsEmail = errors.New("password must not be the same as email")
)

type Policy struct {
	MinLength int
	MaxLength int
}

func DefaultPolicy() Policy {
	return Policy{
		MinLength: 15,
		MaxLength: 128,
	}
}

func (p Policy) Validate(password, normalizedEmail string, blocklist Blocklist) error {
	if utf8.RuneCountInString(password) < p.MinLength {
		return ErrPasswordTooShort
	}
	if utf8.RuneCountInString(password) > p.MaxLength {
		return ErrPasswordTooLong
	}
	lowerPwd := strings.ToLower(strings.TrimSpace(password))
	lowerEmail := strings.ToLower(strings.TrimSpace(normalizedEmail))
	if lowerPwd == lowerEmail && lowerEmail != "" {
		return ErrPasswordEqualsEmail
	}
	if blocklist != nil && blocklist.Contains(password) {
		return ErrPasswordCommon
	}
	return nil
}
