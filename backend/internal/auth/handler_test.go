package auth

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gabrielalc23/pdv/internal/platform/cookie"
	"github.com/gabrielalc23/pdv/internal/platform/csrf"
)

func TestCSRFEndpoint(t *testing.T) {
	cookies, err := cookie.NewManager(cookie.Config{Env: "test", SameSite: "Lax", RefreshName: "pdv_refresh", CSRFName: "pdv_csrf"})
	if err != nil {
		t.Fatal(err)
	}
	manager, err := csrf.NewManager([]byte("0123456789abcdef0123456789abcdef"), []string{"http://example.com"})
	if err != nil {
		t.Fatal(err)
	}
	handler := NewHandler(nil, cookies, manager, nil, nil, nil, nil)
	var previous string
	for range 2 {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest("GET", "/auth/csrf", nil)
		handler.CSRF(recorder, request)
		if recorder.Code != 200 || recorder.Header().Get("Cache-Control") != "no-store" || recorder.Header().Get("Pragma") != "no-cache" {
			t.Fatalf("unexpected response: %d %v", recorder.Code, recorder.Header())
		}
		var body CSRFResponse
		if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		responseCookies := recorder.Result().Cookies()
		if len(responseCookies) != 1 {
			t.Fatalf("cookies = %d", len(responseCookies))
		}
		csrfCookie := responseCookies[0]
		if csrfCookie.Name != "pdv_csrf" || csrfCookie.Value != body.CSRFToken || csrfCookie.HttpOnly || csrfCookie.Path != "/" {
			t.Fatalf("invalid csrf cookie: %+v", csrfCookie)
		}
		if body.CSRFToken == previous {
			t.Fatal("consecutive csrf tokens must differ")
		}
		previous = body.CSRFToken
	}
}
