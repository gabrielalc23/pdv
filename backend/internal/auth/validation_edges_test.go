package auth

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateActionEmail(t *testing.T) {
	t.Parallel()

	t.Run("normalizes case and surrounding whitespace", func(t *testing.T) {
		t.Parallel()
		got, err := validateActionEmail("  Owner@Example.COM  ")
		if err != nil {
			t.Fatalf("validateActionEmail() error = %v", err)
		}
		if got != "owner@example.com" {
			t.Fatalf("validateActionEmail() = %q, want %q", got, "owner@example.com")
		}
	})

	for _, test := range []struct {
		name  string
		value string
	}{
		{name: "empty", value: " \t\n "},
		{name: "missing at sign", value: "owner.example.com"},
		{name: "display address", value: "Owner <owner@example.com>"},
		{name: "too long", value: strings.Repeat("a", 310) + "@example.com"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := validateActionEmail(test.value)
			if got != "" {
				t.Fatalf("validateActionEmail() = %q, want empty result", got)
			}
			var validation *ValidationError
			if !errors.As(err, &validation) || validation.Field != "email" {
				t.Fatalf("validateActionEmail() error = %#v, want email ValidationError", err)
			}
		})
	}
}

func TestValidateRegisterDisplayNameRuneBoundaries(t *testing.T) {
	t.Parallel()

	input := validRegisterRequest()
	want := strings.Repeat("界", 150)
	input.DisplayName = "  " + want + "  "
	if err := validateRegister(&input); err != nil {
		t.Fatalf("validateRegister() rejected 150-rune display name: %v", err)
	}
	if input.DisplayName != want {
		t.Fatalf("validateRegister() display name = %q, want trimmed value", input.DisplayName)
	}

	input = validRegisterRequest()
	input.DisplayName = strings.Repeat("界", 151)
	err := validateRegister(&input)
	var validation *ValidationError
	if !errors.As(err, &validation) || validation.Field != "displayName" {
		t.Fatalf("validateRegister() error = %#v, want displayName ValidationError", err)
	}
}
