package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gabrielalc23/pdv/internal/app"
	"github.com/gabrielalc23/pdv/internal/platform/mailer"
)

type authResponse struct {
	AccessToken string `json:"accessToken"`
	TokenType   string `json:"tokenType"`
	ExpiresIn   int64  `json:"expiresIn"`
	User        struct {
		ID, Email, DisplayName string
		EmailVerified          bool `json:"emailVerified"`
	} `json:"user"`
	Session struct{ ID, ClientID, CreatedAt, IdleExpiresAt, AbsoluteExpiresAt string } `json:"session"`
	Context struct {
		Kind         string                           `json:"kind"`
		MembershipID *string                          `json:"membershipId"`
		Roles        []string                         `json:"roles"`
		Scopes       []string                         `json:"scopes"`
		Organization *struct{ ID, Name, Slug string } `json:"organization"`
		Store        *struct{ ID, Code, Name string } `json:"store"`
	} `json:"context"`
}

func TestAuthOwnerHTTPFlow(t *testing.T) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Jar: jar}
	csrfToken := getCSRFToken(t, client)
	register := map[string]any{
		"email": "owner.task11b@example.com", "password": "uma senha longa segura 2026", "displayName": "Owner E2E",
		"organization": map[string]any{"name": "Empresa Auth E2E", "slug": "empresa-auth-e2e", "timezone": "America/Sao_Paulo", "locale": "pt-BR", "currency": "BRL"},
		"store":        map[string]any{"code": "MATRIZ", "name": "Matriz", "timezone": "America/Sao_Paulo"},
		"clientId":     "pdv-admin", "deviceName": "E2E Browser",
	}

	missing := authRequest(t, client, http.MethodPost, "/auth/register", register, "", "")
	if missing.StatusCode != http.StatusForbidden {
		t.Fatalf("register without csrf = %d", missing.StatusCode)
	}
	if code := responseErrorCode(t, missing); code != "CSRF_TOKEN_MISSING" {
		t.Fatalf("missing csrf code = %q", code)
	}

	registered := authRequest(t, client, http.MethodPost, "/auth/register", register, csrfToken, "")
	if registered.StatusCode != http.StatusCreated {
		t.Fatalf("register = %d: %s", registered.StatusCode, readBody(registered))
	}
	rawRegistered := readBodyBytes(t, registered)
	if bytes.Contains(rawRegistered, []byte("refreshToken")) || bytes.Contains(rawRegistered, []byte("passwordHash")) {
		t.Fatal("sensitive value exposed in register response")
	}
	var auth authResponse
	if err := json.Unmarshal(rawRegistered, &auth); err != nil {
		t.Fatal(err)
	}
	if auth.AccessToken == "" || auth.TokenType != "Bearer" || auth.ExpiresIn != 300 {
		t.Fatalf("invalid auth response: %+v", auth)
	}
	if auth.Context.Kind != "store" || auth.Context.Store == nil || auth.Context.Organization == nil {
		t.Fatalf("unexpected initial context: %+v", auth.Context)
	}
	if !contains(auth.Context.Roles, "owner") || !contains(auth.Context.Scopes, "organization.owners.manage") {
		t.Fatalf("owner authorization missing: %+v", auth.Context)
	}
	assertAuthCookies(t, jar)
	assertBootstrapRows(t, auth.User.ID, auth.Context.Organization.ID, auth.Context.Store.ID)

	csrfToken = cookieValue(t, jar, "pdv_csrf")
	refreshedResp := authRequest(t, client, http.MethodPost, "/auth/refresh", nil, csrfToken, "")
	if refreshedResp.StatusCode != http.StatusOK {
		t.Fatalf("refresh = %d: %s", refreshedResp.StatusCode, readBody(refreshedResp))
	}
	var refreshed authResponse
	decodeResponse(t, refreshedResp, &refreshed)
	if refreshed.AccessToken == auth.AccessToken || refreshed.Session.ID != auth.Session.ID {
		t.Fatal("refresh did not rotate access token while preserving session")
	}

	me := authRequest(t, client, http.MethodGet, "/me", nil, "", refreshed.AccessToken)
	if me.StatusCode != http.StatusOK {
		t.Fatalf("me = %d: %s", me.StatusCode, readBody(me))
	}
	meBody := readBodyBytes(t, me)
	if bytes.Contains(meBody, []byte("accessToken")) || bytes.Contains(meBody, []byte("refresh")) {
		t.Fatal("me exposed token material")
	}

	storeA2, organizationB, storeB1 := prepareAdditionalContexts(t, auth.User.ID, auth.Context.Organization.ID)
	storeA2Resp := authRequest(t, client, http.MethodPost, "/auth/context", map[string]any{"organizationId": auth.Context.Organization.ID, "storeId": storeA2}, "", refreshed.AccessToken)
	if storeA2Resp.StatusCode != http.StatusOK {
		t.Fatalf("store A2 context = %d: %s", storeA2Resp.StatusCode, readBody(storeA2Resp))
	}
	var storeA2Auth authResponse
	decodeResponse(t, storeA2Resp, &storeA2Auth)
	if storeA2Auth.Context.Store == nil || storeA2Auth.Context.Store.ID != storeA2 || !contains(storeA2Auth.Context.Roles, "owner") {
		t.Fatalf("invalid store A2 context: %+v", storeA2Auth.Context)
	}
	stale := authRequest(t, client, http.MethodGet, "/me", nil, "", refreshed.AccessToken)
	if stale.StatusCode != http.StatusUnauthorized {
		t.Fatalf("stale token = %d", stale.StatusCode)
	}
	if code := responseErrorCode(t, stale); code != "AUTH_CONTEXT_STALE" {
		t.Fatalf("stale token code = %q", code)
	}

	organizationBResp := authRequest(t, client, http.MethodPost, "/auth/context", map[string]any{"organizationId": organizationB, "storeId": nil}, "", storeA2Auth.AccessToken)
	if organizationBResp.StatusCode != http.StatusOK {
		t.Fatalf("organization B context = %d: %s", organizationBResp.StatusCode, readBody(organizationBResp))
	}
	var organizationBAuth authResponse
	decodeResponse(t, organizationBResp, &organizationBAuth)
	if organizationBAuth.Context.Kind != "organization" || organizationBAuth.Context.Organization == nil || organizationBAuth.Context.Organization.ID != organizationB {
		t.Fatalf("invalid organization B context: %+v", organizationBAuth.Context)
	}

	storeBResp := authRequest(t, client, http.MethodPost, "/auth/context", map[string]any{"organizationId": organizationB, "storeId": storeB1}, "", organizationBAuth.AccessToken)
	if storeBResp.StatusCode != http.StatusOK {
		t.Fatalf("store B1 context = %d: %s", storeBResp.StatusCode, readBody(storeBResp))
	}
	var storeBAuth authResponse
	decodeResponse(t, storeBResp, &storeBAuth)
	if storeBAuth.Context.Store == nil || storeBAuth.Context.Store.ID != storeB1 || !contains(storeBAuth.Context.Roles, "cashier") || contains(storeBAuth.Context.Roles, "owner") {
		t.Fatalf("invalid store B1 authorization: %+v", storeBAuth.Context)
	}

	unauthorized := authRequest(t, client, http.MethodPost, "/auth/context", map[string]any{"organizationId": organizationB, "storeId": storeA2}, "", storeBAuth.AccessToken)
	if unauthorized.StatusCode != http.StatusNotFound {
		t.Fatalf("cross-organization store context = %d: %s", unauthorized.StatusCode, readBody(unauthorized))
	}
	unauthorized.Body.Close()

	identityResp := authRequest(t, client, http.MethodPost, "/auth/context", map[string]any{"organizationId": nil, "storeId": nil}, "", storeBAuth.AccessToken)
	if identityResp.StatusCode != http.StatusOK {
		t.Fatalf("context = %d: %s", identityResp.StatusCode, readBody(identityResp))
	}
	var identity authResponse
	decodeResponse(t, identityResp, &identity)
	if identity.Context.Kind != "identity" || identity.Context.Organization != nil || len(identity.Context.Roles) != 0 {
		t.Fatalf("invalid identity context: %+v", identity.Context)
	}

	sessionsResp := authRequest(t, client, http.MethodGet, "/me/sessions", nil, "", identity.AccessToken)
	if sessionsResp.StatusCode != http.StatusOK {
		t.Fatalf("sessions = %d: %s", sessionsResp.StatusCode, readBody(sessionsResp))
	}
	var sessionsBody struct {
		Data []struct {
			ID        string `json:"id"`
			IsCurrent bool   `json:"isCurrent"`
		} `json:"data"`
	}
	decodeResponse(t, sessionsResp, &sessionsBody)
	if len(sessionsBody.Data) == 0 || !sessionsBody.Data[0].IsCurrent {
		t.Fatalf("current session missing: %+v", sessionsBody.Data)
	}

	logout := authRequest(t, client, http.MethodPost, "/auth/logout", nil, "", identity.AccessToken)
	if logout.StatusCode != http.StatusNoContent {
		t.Fatalf("logout = %d: %s", logout.StatusCode, readBody(logout))
	}
	afterLogout := authRequest(t, client, http.MethodGet, "/me", nil, "", identity.AccessToken)
	if afterLogout.StatusCode != http.StatusUnauthorized {
		t.Fatalf("token after logout = %d", afterLogout.StatusCode)
	}
	if code := responseErrorCode(t, afterLogout); code != "SESSION_REVOKED" {
		t.Fatalf("after logout code = %q", code)
	}

	preauth := getCSRFToken(t, client)
	invalidUnknown := authRequest(t, client, http.MethodPost, "/auth/login", map[string]any{"email": "unknown.task11b@example.com", "password": "senha errada mas comprida", "clientId": "pdv-admin"}, preauth, "")
	if invalidUnknown.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unknown login = %d", invalidUnknown.StatusCode)
	}
	var unknownError errorBody
	decodeResponse(t, invalidUnknown, &unknownError)
	invalidPassword := authRequest(t, client, http.MethodPost, "/auth/login", map[string]any{"email": "owner.task11b@example.com", "password": "senha totalmente incorreta", "clientId": "pdv-admin"}, preauth, "")
	if invalidPassword.StatusCode != http.StatusUnauthorized {
		t.Fatalf("wrong password = %d", invalidPassword.StatusCode)
	}
	var passwordError errorBody
	decodeResponse(t, invalidPassword, &passwordError)
	if unknownError.Error.Code != "INVALID_CREDENTIALS" || unknownError.Error.Code != passwordError.Error.Code || unknownError.Error.Message != passwordError.Error.Message {
		t.Fatalf("credential responses differ: %+v %+v", unknownError, passwordError)
	}

	loginPayload := map[string]any{"email": "owner.task11b@example.com", "password": "uma senha longa segura 2026", "clientId": "pdv-admin", "deviceName": "Login One"}
	loginOneResp := authRequest(t, client, http.MethodPost, "/auth/login", loginPayload, preauth, "")
	if loginOneResp.StatusCode != http.StatusOK {
		t.Fatalf("login one = %d: %s", loginOneResp.StatusCode, readBody(loginOneResp))
	}
	var loginOne authResponse
	decodeResponse(t, loginOneResp, &loginOne)
	preauth = getCSRFToken(t, client)
	loginPayload["deviceName"] = "Login Two"
	loginTwoResp := authRequest(t, client, http.MethodPost, "/auth/login", loginPayload, preauth, "")
	if loginTwoResp.StatusCode != http.StatusOK {
		t.Fatalf("login two = %d: %s", loginTwoResp.StatusCode, readBody(loginTwoResp))
	}
	var loginTwo authResponse
	decodeResponse(t, loginTwoResp, &loginTwo)

	revokeOther := authRequest(t, client, http.MethodDelete, "/me/sessions/"+loginOne.Session.ID, nil, "", loginTwo.AccessToken)
	if revokeOther.StatusCode != http.StatusNoContent {
		t.Fatalf("revoke other = %d: %s", revokeOther.StatusCode, readBody(revokeOther))
	}
	currentStillActive := authRequest(t, client, http.MethodGet, "/me", nil, "", loginTwo.AccessToken)
	if currentStillActive.StatusCode != http.StatusOK {
		t.Fatalf("current session after revoke = %d", currentStillActive.StatusCode)
	}
	currentStillActive.Body.Close()
	otherRevoked := authRequest(t, client, http.MethodGet, "/me", nil, "", loginOne.AccessToken)
	if otherRevoked.StatusCode != http.StatusUnauthorized {
		t.Fatalf("other token after revoke = %d", otherRevoked.StatusCode)
	}
	otherRevoked.Body.Close()

	logoutAll := authRequest(t, client, http.MethodPost, "/auth/logout-all", nil, "", loginTwo.AccessToken)
	if logoutAll.StatusCode != http.StatusNoContent {
		t.Fatalf("logout all = %d: %s", logoutAll.StatusCode, readBody(logoutAll))
	}
	afterLogoutAll := authRequest(t, client, http.MethodGet, "/me", nil, "", loginTwo.AccessToken)
	if afterLogoutAll.StatusCode != http.StatusUnauthorized {
		t.Fatalf("token after logout all = %d", afterLogoutAll.StatusCode)
	}
	afterLogoutAll.Body.Close()

	preauth = getCSRFToken(t, client)
	reuseLoginResp := authRequest(t, client, http.MethodPost, "/auth/login", loginPayload, preauth, "")
	if reuseLoginResp.StatusCode != http.StatusOK {
		t.Fatalf("reuse login = %d: %s", reuseLoginResp.StatusCode, readBody(reuseLoginResp))
	}
	var reuseLogin authResponse
	decodeResponse(t, reuseLoginResp, &reuseLogin)
	oldRefresh := cookieValue(t, client.Jar, "pdv_refresh")
	sessionCSRF := cookieValue(t, client.Jar, "pdv_csrf")
	validRotation := authRequest(t, client, http.MethodPost, "/auth/refresh", nil, sessionCSRF, "")
	if validRotation.StatusCode != http.StatusOK {
		t.Fatalf("reuse setup refresh = %d: %s", validRotation.StatusCode, readBody(validRotation))
	}
	var rotated authResponse
	decodeResponse(t, validRotation, &rotated)
	currentCSRF := cookieValue(t, client.Jar, "pdv_csrf")
	reuseRequest, err := http.NewRequest(http.MethodPost, baseURL+"/auth/refresh", nil)
	if err != nil {
		t.Fatal(err)
	}
	reuseRequest.Header.Set("Origin", baseURL)
	reuseRequest.Header.Set("Sec-Fetch-Site", "same-origin")
	reuseRequest.Header.Set("X-CSRF-Token", currentCSRF)
	reuseRequest.Header.Set("Cookie", "pdv_refresh="+oldRefresh+"; pdv_csrf="+currentCSRF)
	reused, err := (&http.Client{}).Do(reuseRequest)
	if err != nil {
		t.Fatal(err)
	}
	if reused.StatusCode != http.StatusUnauthorized {
		t.Fatalf("reuse = %d: %s", reused.StatusCode, readBody(reused))
	}
	if code := responseErrorCode(t, reused); code != "REFRESH_TOKEN_REUSED" {
		t.Fatalf("reuse code = %q", code)
	}
	afterReuse := authRequest(t, client, http.MethodGet, "/me", nil, "", rotated.AccessToken)
	if afterReuse.StatusCode != http.StatusUnauthorized {
		t.Fatalf("token after reuse = %d", afterReuse.StatusCode)
	}
	afterReuse.Body.Close()

	preauth = getCSRFToken(t, client)
	rateLimited := authRequest(t, client, http.MethodPost, "/auth/login", loginPayload, preauth, "")
	if rateLimited.StatusCode != http.StatusTooManyRequests || rateLimited.Header.Get("Retry-After") == "" {
		t.Fatalf("login rate limit = %d retry=%q", rateLimited.StatusCode, rateLimited.Header.Get("Retry-After"))
	}
	rateLimited.Body.Close()
}

type captureMailer struct {
	mu                 sync.Mutex
	link               string
	committed          bool
	verificationLinks  []string
	passwordResetLinks []string
	invitationLinks    []string
}

func (m *captureMailer) SendEmailVerification(ctx context.Context, to, _ string, link string) error {
	exists, err := capturedActionTokenCommitted(ctx, to, link, "EMAIL_VERIFICATION")
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.link, m.committed = link, exists
	m.verificationLinks = append(m.verificationLinks, link)
	return nil
}

func (m *captureMailer) SendPasswordReset(ctx context.Context, to, _ string, link string) error {
	exists, err := capturedActionTokenCommitted(ctx, to, link, "PASSWORD_RESET")
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.committed = exists
	m.passwordResetLinks = append(m.passwordResetLinks, link)
	return nil
}

func (m *captureMailer) SendInvitation(_ context.Context, _, _, _, link string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.invitationLinks = append(m.invitationLinks, link)
	return nil
}

func capturedActionTokenCommitted(ctx context.Context, recipient, link, purpose string) (bool, error) {
	parsed, err := url.Parse(link)
	if err != nil {
		return false, err
	}
	fragment, err := url.ParseQuery(parsed.Fragment)
	if err != nil {
		return false, err
	}
	raw := fragment.Get("token")
	separator := strings.IndexByte(raw, '.')
	underscore := strings.IndexByte(raw, '_')
	if underscore < 0 || separator <= underscore+1 {
		return false, nil
	}
	selector := raw[underscore+1 : separator]
	var exists bool
	err = testPool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM auth_action_tokens t
			JOIN users u ON u.id=t.user_id
			WHERE t.id=$1 AND t.purpose=$2 AND u.email_normalized=$3
		)
	`, selector, purpose, strings.ToLower(recipient)).Scan(&exists)
	return exists, err
}

func TestRegistrationFlagsAndVerification(t *testing.T) {
	mail := &captureMailer{}
	verifiedURL, verifiedClient := auxiliaryAuthServer(t, true, true, mail)
	csrfToken := getCSRFTokenAt(t, verifiedClient, verifiedURL)
	payload := map[string]any{
		"email": "verified.task11b@example.com", "password": "outra senha longa segura 2026", "displayName": "Verified Owner",
		"organization": map[string]any{"name": "Empresa Verify E2E", "slug": "empresa-verify-e2e", "timezone": "America/Sao_Paulo", "locale": "pt-BR", "currency": "BRL"},
		"store":        map[string]any{"code": "MATRIZ", "name": "Matriz", "timezone": "America/Sao_Paulo"}, "clientId": "pdv-admin",
	}
	response := authRequestAt(t, verifiedClient, verifiedURL, http.MethodPost, "/auth/register", payload, csrfToken, "")
	if response.StatusCode != http.StatusAccepted {
		t.Fatalf("verified register = %d: %s", response.StatusCode, readBody(response))
	}
	var status struct {
		Status string `json:"status"`
	}
	decodeResponse(t, response, &status)
	if status.Status != "VERIFICATION_REQUIRED" {
		t.Fatalf("status = %q", status.Status)
	}
	mail.mu.Lock()
	link, committed := mail.link, mail.committed
	mail.mu.Unlock()
	if !committed || !strings.Contains(link, "evt_") {
		t.Fatalf("verification mail was not sent after commit (committed=%v)", committed)
	}
	var sessionsCount, hashLength int
	if err := testPool.QueryRow(context.Background(), `SELECT COUNT(*) FROM auth_sessions s JOIN users u ON u.id=s.user_id WHERE u.email_normalized='verified.task11b@example.com'`).Scan(&sessionsCount); err != nil {
		t.Fatal(err)
	}
	if err := testPool.QueryRow(context.Background(), `SELECT OCTET_LENGTH(t.secret_hash) FROM auth_action_tokens t JOIN users u ON u.id=t.user_id WHERE u.email_normalized='verified.task11b@example.com'`).Scan(&hashLength); err != nil {
		t.Fatal(err)
	}
	if sessionsCount != 0 || hashLength != 32 {
		t.Fatalf("verification persistence sessions=%d hash=%d", sessionsCount, hashLength)
	}
	registerLimited := authRequestAt(t, verifiedClient, verifiedURL, http.MethodPost, "/auth/register", payload, csrfToken, "")
	if registerLimited.StatusCode != http.StatusTooManyRequests || registerLimited.Header.Get("Retry-After") == "" {
		t.Fatalf("register rate limit = %d retry=%q", registerLimited.StatusCode, registerLimited.Header.Get("Retry-After"))
	}
	registerLimited.Body.Close()

	disabledURL, disabledClient := auxiliaryAuthServer(t, false, false, mail)
	disabled := authRequestAt(t, disabledClient, disabledURL, http.MethodPost, "/auth/register", payload, "", "")
	if disabled.StatusCode != http.StatusForbidden {
		t.Fatalf("disabled register = %d", disabled.StatusCode)
	}
	if code := responseErrorCode(t, disabled); code != "REGISTRATION_DISABLED" {
		t.Fatalf("disabled code = %q", code)
	}
}

func getCSRFToken(t *testing.T, client *http.Client) string {
	return getCSRFTokenAt(t, client, baseURL)
}

func getCSRFTokenAt(t *testing.T, client *http.Client, serverURL string) string {
	t.Helper()
	resp := authRequestAt(t, client, serverURL, http.MethodGet, "/auth/csrf", nil, "", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("csrf = %d: %s", resp.StatusCode, readBody(resp))
	}
	var body struct {
		CSRFToken string `json:"csrfToken"`
	}
	decodeResponse(t, resp, &body)
	if body.CSRFToken == "" || body.CSRFToken != cookieValueAt(t, client.Jar, serverURL, "pdv_csrf") {
		t.Fatal("csrf body/cookie mismatch")
	}
	return body.CSRFToken
}

func authRequest(t *testing.T, client *http.Client, method, path string, body any, csrfToken, accessToken string) *http.Response {
	return authRequestAt(t, client, baseURL, method, path, body, csrfToken, accessToken)
}

func authRequestAt(t *testing.T, client *http.Client, serverURL, method, path string, body any, csrfToken, accessToken string) *http.Response {
	t.Helper()
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, serverURL+path, reader)
	if err != nil {
		t.Fatal(err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if method != http.MethodGet {
		req.Header.Set("Origin", serverURL)
		req.Header.Set("Sec-Fetch-Site", "same-origin")
	}
	if csrfToken != "" {
		req.Header.Set("X-CSRF-Token", csrfToken)
	}
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func assertAuthCookies(t *testing.T, jar http.CookieJar) {
	t.Helper()
	parsed, _ := url.Parse(baseURL)
	cookies := jar.Cookies(parsed)
	var refresh, csrfCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "pdv_refresh" {
			refresh = c
		}
		if c.Name == "pdv_csrf" {
			csrfCookie = c
		}
	}
	if refresh == nil || csrfCookie == nil {
		t.Fatalf("auth cookies missing: %+v", cookies)
	}
	// CookieJar intentionally does not expose HttpOnly; raw Set-Cookie is covered by cookie manager unit tests.
}

func cookieValue(t *testing.T, jar http.CookieJar, name string) string {
	return cookieValueAt(t, jar, baseURL, name)
}

func cookieValueAt(t *testing.T, jar http.CookieJar, serverURL, name string) string {
	t.Helper()
	parsed, _ := url.Parse(serverURL)
	for _, c := range jar.Cookies(parsed) {
		if c.Name == name {
			return c.Value
		}
	}
	t.Fatalf("cookie %s not found", name)
	return ""
}

func auxiliaryAuthServer(t *testing.T, registrationEnabled, requireVerified bool, authMailer mailer.Mailer) (string, *http.Client) {
	t.Helper()
	listener := mustListen()
	serverURL := "http://" + listener.Addr().String()
	handler := app.New(testDependencies(testStore, testValkey, serverURL, registrationEnabled, requireVerified, authMailer))
	server := app.NewHTTPServer(listener.Addr().String(), handler)
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	})
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	return serverURL, &http.Client{Jar: jar, Timeout: 10 * time.Second}
}

func assertBootstrapRows(t *testing.T, userID, organizationID, storeID string) {
	t.Helper()
	ctx := context.Background()
	var memberships, roles, methods, bindings int
	if err := testPool.QueryRow(ctx, `SELECT COUNT(*) FROM organization_memberships WHERE user_id=$1 AND organization_id=$2 AND status='ACTIVE'`, userID, organizationID).Scan(&memberships); err != nil {
		t.Fatal(err)
	}
	if err := testPool.QueryRow(ctx, `SELECT COUNT(*) FROM roles WHERE organization_id=$1`, organizationID).Scan(&roles); err != nil {
		t.Fatal(err)
	}
	if err := testPool.QueryRow(ctx, `SELECT COUNT(*) FROM payment_methods WHERE organization_id=$1`, organizationID).Scan(&methods); err != nil {
		t.Fatal(err)
	}
	if err := testPool.QueryRow(ctx, `SELECT COUNT(*) FROM store_payment_methods WHERE organization_id=$1 AND store_id=$2`, organizationID, storeID).Scan(&bindings); err != nil {
		t.Fatal(err)
	}
	if memberships != 1 || roles != 7 || methods != 5 || bindings != 5 {
		t.Fatalf("bootstrap counts membership=%d roles=%d methods=%d bindings=%d", memberships, roles, methods, bindings)
	}
}

func prepareAdditionalContexts(t *testing.T, userID, organizationA string) (storeA2, organizationB, storeB1 string) {
	t.Helper()
	ctx := context.Background()
	if err := testPool.QueryRow(ctx, `INSERT INTO stores (organization_id, code, name, timezone, created_by_user_id) VALUES ($1,'FILIAL','Filial','America/Sao_Paulo',$2) RETURNING id`, organizationA, userID).Scan(&storeA2); err != nil {
		t.Fatal(err)
	}
	if err := testPool.QueryRow(ctx, `INSERT INTO organizations (name,slug,created_by_user_id) VALUES ('Empresa B','empresa-b-task11b',$1) RETURNING id`, userID).Scan(&organizationB); err != nil {
		t.Fatal(err)
	}
	if err := testPool.QueryRow(ctx, `INSERT INTO stores (organization_id,code,name,timezone,created_by_user_id) VALUES ($1,'B1','Loja B1','America/Sao_Paulo',$2) RETURNING id`, organizationB, userID).Scan(&storeB1); err != nil {
		t.Fatal(err)
	}
	var membershipB, roleB string
	if err := testPool.QueryRow(ctx, `INSERT INTO organization_memberships (organization_id,user_id,default_store_id,created_by_user_id) VALUES ($1,$2,$3,$2) RETURNING id`, organizationB, userID, storeB1).Scan(&membershipB); err != nil {
		t.Fatal(err)
	}
	if err := testPool.QueryRow(ctx, `INSERT INTO roles (organization_id,key,name,assignment_scope,created_by_membership_id) VALUES ($1,'cashier','Caixa','STORE',$2) RETURNING id`, organizationB, membershipB).Scan(&roleB); err != nil {
		t.Fatal(err)
	}
	if _, err := testPool.Exec(ctx, `INSERT INTO role_scopes (organization_id,role_id,scope_code) VALUES ($1,$2,'catalog.read'),($1,$2,'sales.read'),($1,$2,'sales.create')`, organizationB, roleB); err != nil {
		t.Fatal(err)
	}
	if _, err := testPool.Exec(ctx, `INSERT INTO membership_role_bindings (organization_id,membership_id,role_id,store_id,created_by_membership_id) VALUES ($1,$2,$3,$4,$2)`, organizationB, membershipB, roleB, storeB1); err != nil {
		t.Fatal(err)
	}
	return storeA2, organizationB, storeB1
}

func responseErrorCode(t *testing.T, resp *http.Response) string {
	t.Helper()
	var body errorBody
	decodeResponse(t, resp, &body)
	return body.Error.Code
}
func decodeResponse(t *testing.T, resp *http.Response, target any) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		t.Fatal(err)
	}
}
func readBodyBytes(t *testing.T, resp *http.Response) []byte {
	t.Helper()
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
func readBody(resp *http.Response) string {
	if resp == nil || resp.Body == nil {
		return ""
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return string(data)
}
func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
func noSecrets(value string) bool { return !strings.Contains(value, "rt_") }
