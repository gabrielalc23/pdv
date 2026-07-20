package users

import (
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"
)

var ErrInvalidDisplayName = errors.New("invalid display name")

func NormalizeDisplayName(value string) (string, error) {
	value = strings.TrimSpace(value)
	if utf8.RuneCountInString(value) < 1 || utf8.RuneCountInString(value) > 150 {
		return "", ErrInvalidDisplayName
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return "", ErrInvalidDisplayName
		}
	}
	return value, nil
}
