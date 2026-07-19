package auth

import (
	"errors"
	"strings"
	"testing"

	"github.com/gabrielalc23/pdv/internal/platform/password"
	"github.com/gabrielalc23/pdv/internal/platform/requestmeta"
)

func TestValidateNewPasswordMapsPolicyErrors(t *testing.T) {
	t.Parallel()

	service := &Service{policy: password.DefaultPolicy(), blocklist: password.NewBuiltinBlocklist()}
	for _, test := range []struct {
		name            string
		value           string
		normalizedEmail string
		wantErr         error
	}{
		{name: "empty", value: "", wantErr: ErrWeakPassword},
		{name: "too short", value: "short", wantErr: ErrWeakPassword},
		{name: "too long", value: strings.Repeat("a", 129), wantErr: ErrWeakPassword},
		{name: "equals email", value: "owner@example.com", normalizedEmail: "owner@example.com", wantErr: ErrWeakPassword},
		{name: "common", value: "123456789012345", wantErr: ErrCommonPassword},
		{name: "valid", value: "a sufficiently long password 42"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := service.validateNewPassword(test.value, test.normalizedEmail)
			if test.wantErr == nil {
				if err != nil {
					t.Fatalf("validateNewPassword() error = %v", err)
				}
				return
			}
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("validateNewPassword() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestRegisterMapsPasswordPolicyErrorsBeforePersistence(t *testing.T) {
	t.Parallel()

	service := &Service{
		policy:    password.DefaultPolicy(),
		blocklist: password.NewBuiltinBlocklist(),
		cfg:       Config{RegistrationEnabled: true},
	}
	for _, test := range []struct {
		name    string
		value   string
		wantErr error
	}{
		{name: "weak", value: "short", wantErr: ErrWeakPassword},
		{name: "common", value: "123456789012345", wantErr: ErrCommonPassword},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			input := validRegisterRequest()
			input.Password = test.value
			_, err := service.Register(t.Context(), input, requestmeta.RequestMetadata{})
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("Register() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}
