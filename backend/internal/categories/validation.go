package categories

import (
	"strings"
	"unicode"
)

func normalizeName(value string) (string, error) {
	name := strings.TrimSpace(value)
	if name == "" {
		return "", newValidationError("name", "is required")
	}
	return name, nil
}

func slugify(value string) string {
	var builder strings.Builder
	previousDash := false

	for _, char := range strings.ToLower(value) {
		switch {
		case unicode.IsLetter(char) || unicode.IsDigit(char):
			builder.WriteRune(char)
			previousDash = false
		case !previousDash:
			builder.WriteRune('-')
			previousDash = true
		}
	}

	return strings.Trim(builder.String(), "-")
}
