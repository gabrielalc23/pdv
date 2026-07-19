package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	platformhttp "github.com/gabrielalc23/pdv/internal/platform/http"
)

func TestMapHTTPErrorActionTokensAndPasswordPolicy(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name string
		err  error
		want httpError
	}{
		{name: "missing action token", err: ErrActionTokenMissing, want: httpError{status: http.StatusBadRequest, code: "INVALID_REQUEST", message: "Requisição inválida."}},
		{name: "wrapped malformed action token", err: fmt.Errorf("parse token: %w", ErrActionTokenInvalid), want: httpError{status: http.StatusBadRequest, code: "INVALID_REQUEST", message: "Requisição inválida."}},
		{name: "wrapped expired action token", err: fmt.Errorf("consume token: %w", ErrActionTokenExpired), want: httpError{status: http.StatusGone, code: "ACTION_TOKEN_EXPIRED", message: "Token de ação expirado."}},
		{name: "weak password", err: ErrWeakPassword, want: httpError{status: http.StatusUnprocessableEntity, code: "WEAK_PASSWORD", message: "A senha não atende aos requisitos de segurança.", field: "password"}},
		{name: "common password", err: ErrCommonPassword, want: httpError{status: http.StatusUnprocessableEntity, code: "COMMON_PASSWORD", message: "A senha informada é muito comum.", field: "password"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := mapHTTPError(test.err); got != test.want {
				t.Fatalf("mapHTTPError() = %+v, want %+v", got, test.want)
			}
		})
	}
}

func TestWriteErrorSetsSensitiveHeadersAndBearerChallenge(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name          string
		err           error
		bearer        bool
		wantStatus    int
		wantCode      string
		wantField     string
		wantChallenge bool
	}{
		{name: "bearer unauthorized", err: ErrInvalidCredentials, bearer: true, wantStatus: http.StatusUnauthorized, wantCode: "INVALID_CREDENTIALS", wantChallenge: true},
		{name: "non-bearer unauthorized", err: ErrInvalidCredentials, wantStatus: http.StatusUnauthorized, wantCode: "INVALID_CREDENTIALS"},
		{name: "bearer validation failure", err: ErrWeakPassword, bearer: true, wantStatus: http.StatusUnprocessableEntity, wantCode: "WEAK_PASSWORD", wantField: "password"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			recorder := httptest.NewRecorder()
			writeError(recorder, test.err, test.bearer)

			if recorder.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, test.wantStatus)
			}
			if recorder.Header().Get("Cache-Control") != "no-store" || recorder.Header().Get("Pragma") != "no-cache" {
				t.Fatalf("sensitive headers = %v", recorder.Header())
			}
			if got := recorder.Header().Get("WWW-Authenticate"); (got == "Bearer") != test.wantChallenge {
				t.Fatalf("WWW-Authenticate = %q, want challenge %v", got, test.wantChallenge)
			}

			var body platformhttp.ErrorResponse
			if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if body.Error.Code != test.wantCode || body.Error.Field != test.wantField {
				t.Fatalf("error response = %+v", body.Error)
			}
		})
	}
}

func TestSensitiveHeadersMiddleware(t *testing.T) {
	t.Parallel()

	handler := SensitiveHeaders(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Test", "reached")
		w.WriteHeader(http.StatusAccepted)
	}))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/me", nil))

	if recorder.Code != http.StatusAccepted || recorder.Header().Get("X-Test") != "reached" {
		t.Fatalf("middleware did not delegate: status=%d headers=%v", recorder.Code, recorder.Header())
	}
	if recorder.Header().Get("Cache-Control") != "no-store" || recorder.Header().Get("Pragma") != "no-cache" {
		t.Fatalf("sensitive headers = %v", recorder.Header())
	}
}
