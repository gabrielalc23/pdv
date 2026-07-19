package auth

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeSmallJSONStrictness(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name    string
		body    string
		wantErr bool
	}{
		{name: "valid", body: `{"email":"owner@example.com"}`},
		{name: "valid with trailing whitespace", body: "{\"email\":\"owner@example.com\"}\n\t "},
		{name: "unknown field", body: `{"email":"owner@example.com","unexpected":true}`, wantErr: true},
		{name: "second document", body: `{"email":"owner@example.com"}{"email":"other@example.com"}`, wantErr: true},
		{name: "wrong top-level type", body: `[{"email":"owner@example.com"}]`, wantErr: true},
		{name: "malformed", body: `{"email":`, wantErr: true},
		{name: "empty", body: ``, wantErr: true},
		{name: "over size limit", body: strings.Repeat(" ", (8<<10)+1), wantErr: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest("POST", "/auth/password/forgot", strings.NewReader(test.body))
			var input EmailActionRequest
			err := decodeSmallJSON(recorder, request, &input)
			if (err != nil) != test.wantErr {
				t.Fatalf("decodeSmallJSON() error = %v, wantErr %v", err, test.wantErr)
			}
			if !test.wantErr && input.Email != "owner@example.com" {
				t.Fatalf("decodeSmallJSON() input = %+v", input)
			}
		})
	}
}

func TestDecodeSmallJSONAppliesStrictnessToActionDTOs(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name string
		body string
		dst  func() any
	}{
		{name: "password reset", body: `{"token":"token","newPassword":"secret","extra":1}`, dst: func() any { return &PasswordResetRequest{} }},
		{name: "email verification", body: `{"token":"token","extra":1}`, dst: func() any { return &EmailVerifyRequest{} }},
		{name: "password change", body: `{"currentPassword":"old","newPassword":"new","extra":1}`, dst: func() any { return &ChangePasswordRequest{} }},
		{name: "profile update", body: `{"displayName":"Name","email":"other@example.com"}`, dst: func() any { return &UpdateMeRequest{} }},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest("POST", "/", strings.NewReader(test.body))
			if err := decodeSmallJSON(recorder, request, test.dst()); err == nil {
				t.Fatal("decodeSmallJSON() accepted an unknown DTO field")
			}
		})
	}
}
