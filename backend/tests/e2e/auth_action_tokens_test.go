package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"testing"
	"time"

	authmodule "github.com/gabrielalc23/pdv/internal/auth"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	actionTokenOldPassword = "Task 11C old password 2026!"
	actionTokenNewPassword = "Task 11C new password 2027!"
)

func TestAuthActionTokens(t *testing.T) {
	t.Run("CSRF bearer and sensitive header contracts", func(t *testing.T) {
		resetRateLimitState(t)
		serverURL, client := auxiliaryAuthServer(t, true, false, &captureMailer{})
		publicCases := []struct {
			path string
			body map[string]any
		}{
			{path: "/auth/password/forgot", body: map[string]any{"email": "csrf.task11c@example.com"}},
			{path: "/auth/password/reset", body: map[string]any{"token": "prt_invalid", "newPassword": actionTokenNewPassword}},
			{path: "/auth/email/verify", body: map[string]any{"token": "evt_invalid"}},
			{path: "/auth/email/resend-verification", body: map[string]any{"email": "csrf.task11c@example.com"}},
		}
		for _, tt := range publicCases {
			resp := authRequestAt(t, client, serverURL, http.MethodPost, tt.path, tt.body, "", "")
			if resp.StatusCode != http.StatusForbidden || resp.Header.Get("Cache-Control") != "no-store" || resp.Header.Get("Pragma") != "no-cache" {
				t.Fatalf("%s CSRF/header contract status=%d cache=%q pragma=%q", tt.path, resp.StatusCode, resp.Header.Get("Cache-Control"), resp.Header.Get("Pragma"))
			}
			if code := responseErrorCode(t, resp); code != "CSRF_TOKEN_MISSING" {
				t.Fatalf("%s missing CSRF code=%q", tt.path, code)
			}
		}

		for _, tt := range []struct {
			method, path string
			body         map[string]any
		}{
			{method: http.MethodPatch, path: "/me", body: map[string]any{"displayName": "Name"}},
			{method: http.MethodPost, path: "/me/password", body: map[string]any{"currentPassword": actionTokenOldPassword, "newPassword": actionTokenNewPassword}},
		} {
			resp := authRequestAt(t, client, serverURL, tt.method, tt.path, tt.body, "", "")
			if resp.StatusCode != http.StatusUnauthorized || resp.Header.Get("WWW-Authenticate") != "Bearer" {
				t.Fatalf("%s bearer contract status=%d challenge=%q", tt.path, resp.StatusCode, resp.Header.Get("WWW-Authenticate"))
			}
			resp.Body.Close()
		}
	})

	t.Run("unknown forgot is indistinguishable and creates no token or mail", func(t *testing.T) {
		resetRateLimitState(t)
		mail := &captureMailer{}
		serverURL, client := auxiliaryAuthServer(t, true, false, mail)

		var tokensBefore, tokensAfter int
		if err := testPool.QueryRow(context.Background(), `SELECT COUNT(*) FROM auth_action_tokens`).Scan(&tokensBefore); err != nil {
			t.Fatal(err)
		}
		linksBefore, _ := passwordResetLinks(mail)
		acceptedEmailAction(t, client, serverURL, "/auth/password/forgot", "unknown-forgot.task11c@example.com")
		linksAfter, _ := passwordResetLinks(mail)
		if err := testPool.QueryRow(context.Background(), `SELECT COUNT(*) FROM auth_action_tokens`).Scan(&tokensAfter); err != nil {
			t.Fatal(err)
		}
		if tokensAfter != tokensBefore || len(linksAfter) != len(linksBefore) {
			t.Fatalf("unknown forgot persisted token or sent mail: tokens %d->%d mail %d->%d", tokensBefore, tokensAfter, len(linksBefore), len(linksAfter))
		}

		var genericAudits int
		if err := testPool.QueryRow(context.Background(), `
			SELECT COUNT(*)
			FROM security_audit_events
			WHERE event_type='auth.password.reset_requested'
			  AND actor_user_id IS NULL
			  AND outcome='SUCCESS'
			  AND metadata->>'requested_via'='public'
		`).Scan(&genericAudits); err != nil {
			t.Fatal(err)
		}
		if genericAudits < 1 {
			t.Fatal("unknown forgot audit was not recorded")
		}
	})

	t.Run("known forgot sends post-commit link and stores only HMAC", func(t *testing.T) {
		resetRateLimitState(t)
		mail := &captureMailer{}
		serverURL, client := auxiliaryAuthServer(t, true, false, mail)
		auth := registerOwner(t, client, serverURL, "known-forgot.task11c@example.com", "known-forgot-task11c", actionTokenOldPassword)

		acceptedEmailAction(t, client, serverURL, "/auth/password/forgot", auth.User.Email)
		links, committed := passwordResetLinks(mail)
		if len(links) != 1 || !committed {
			t.Fatalf("known forgot mail count=%d committed=%v", len(links), committed)
		}
		rawToken := tokenFromFragment(t, links[0])
		codec := testActionTokenCodec(t)
		parsed, err := codec.Parse(rawToken, authmodule.ActionTokenPurposePasswordReset)
		if err != nil {
			t.Fatal("password reset link contained an invalid token")
		}

		var secretHash []byte
		var consumed bool
		var ttlSeconds int64
		if err := testPool.QueryRow(context.Background(), `
			SELECT secret_hash, consumed_at IS NOT NULL, EXTRACT(EPOCH FROM (expires_at-created_at))::BIGINT
			FROM auth_action_tokens
			WHERE id=$1 AND user_id=$2 AND purpose='PASSWORD_RESET'
		`, parsed.Selector, auth.User.ID).Scan(&secretHash, &consumed, &ttlSeconds); err != nil {
			t.Fatal(err)
		}
		if len(secretHash) != 32 || consumed || ttlSeconds != 1800 || !codec.VerifySecret(parsed.Secret, secretHash) || bytes.Equal(secretHash, []byte(rawToken)) {
			t.Fatalf("password reset token persistence invariant failed: hashLength=%d consumed=%v ttl=%d", len(secretHash), consumed, ttlSeconds)
		}

		var rawTokenColumns, rawAuditReferences, audits int
		if err := testPool.QueryRow(context.Background(), `
			SELECT COUNT(*)
			FROM information_schema.columns
			WHERE table_schema='public' AND table_name='auth_action_tokens'
			  AND column_name IN ('token', 'raw_token', 'secret')
		`).Scan(&rawTokenColumns); err != nil {
			t.Fatal(err)
		}
		if err := testPool.QueryRow(context.Background(), `
			SELECT COUNT(*)
			FROM security_audit_events
			WHERE POSITION($1 IN metadata::text) > 0
		`, rawToken).Scan(&rawAuditReferences); err != nil {
			t.Fatal(err)
		}
		if err := testPool.QueryRow(context.Background(), `
			SELECT COUNT(*)
			FROM security_audit_events
			WHERE event_type='auth.password.reset_requested'
			  AND actor_user_id=$1
			  AND outcome='SUCCESS'
			  AND metadata->>'requested_via'='public'
			  AND COALESCE(metadata->>'email_fingerprint', '') <> ''
		`, auth.User.ID).Scan(&audits); err != nil {
			t.Fatal(err)
		}
		if rawTokenColumns != 0 || rawAuditReferences != 0 || audits != 1 {
			t.Fatalf("known forgot storage/audit mismatch: rawColumns=%d rawAudits=%d audits=%d", rawTokenColumns, rawAuditReferences, audits)
		}
	})

	t.Run("second forgot invalidates the first token", func(t *testing.T) {
		resetRateLimitState(t)
		mail := &captureMailer{}
		serverURL, client := auxiliaryAuthServer(t, true, false, mail)
		auth := registerOwner(t, client, serverURL, "forgot-rotation.task11c@example.com", "forgot-rotation-task11c", actionTokenOldPassword)

		acceptedEmailAction(t, client, serverURL, "/auth/password/forgot", auth.User.Email)
		acceptedEmailAction(t, client, serverURL, "/auth/password/forgot", auth.User.Email)
		links, _ := passwordResetLinks(mail)
		if len(links) != 2 {
			t.Fatalf("password reset mail count=%d", len(links))
		}
		firstRaw := tokenFromFragment(t, links[0])
		secondRaw := tokenFromFragment(t, links[1])
		codec := testActionTokenCodec(t)
		first, firstErr := codec.Parse(firstRaw, authmodule.ActionTokenPurposePasswordReset)
		second, secondErr := codec.Parse(secondRaw, authmodule.ActionTokenPurposePasswordReset)
		if firstErr != nil || secondErr != nil || first.Selector == second.Selector {
			t.Fatal("forgot rotation produced invalid or duplicate selectors")
		}

		var firstConsumed, secondConsumed bool
		if err := testPool.QueryRow(context.Background(), `SELECT consumed_at IS NOT NULL FROM auth_action_tokens WHERE id=$1`, first.Selector).Scan(&firstConsumed); err != nil {
			t.Fatal(err)
		}
		if err := testPool.QueryRow(context.Background(), `SELECT consumed_at IS NOT NULL FROM auth_action_tokens WHERE id=$1`, second.Selector).Scan(&secondConsumed); err != nil {
			t.Fatal(err)
		}
		if !firstConsumed || secondConsumed {
			t.Fatalf("forgot rotation consumption state first=%v second=%v", firstConsumed, secondConsumed)
		}

		csrfToken := getCSRFTokenAt(t, client, serverURL)
		invalidated := authRequestAt(t, client, serverURL, http.MethodPost, "/auth/password/reset", map[string]any{"token": firstRaw, "newPassword": actionTokenNewPassword}, csrfToken, "")
		if invalidated.StatusCode != http.StatusBadRequest {
			t.Fatalf("invalidated reset status=%d", invalidated.StatusCode)
		}
		if code := responseErrorCode(t, invalidated); code != "INVALID_REQUEST" {
			t.Fatalf("invalidated reset code=%q", code)
		}
	})

	t.Run("valid reset changes credentials and revokes every session", func(t *testing.T) {
		resetRateLimitState(t)
		mail := &captureMailer{}
		serverURL, primaryClient := auxiliaryAuthServer(t, true, false, mail)
		email := "valid-reset.task11c@example.com"
		registered := registerOwner(t, primaryClient, serverURL, email, "valid-reset-task11c", actionTokenOldPassword)
		secondaryClient := newHTTPClient(t)
		secondary := loginOwner(t, secondaryClient, serverURL, email, actionTokenOldPassword, "Reset Secondary")

		var versionBefore int64
		if err := testPool.QueryRow(context.Background(), `SELECT password_version FROM users WHERE id=$1`, registered.User.ID).Scan(&versionBefore); err != nil {
			t.Fatal(err)
		}
		publicClient := newHTTPClient(t)
		acceptedEmailAction(t, publicClient, serverURL, "/auth/password/forgot", email)
		links, _ := passwordResetLinks(mail)
		if len(links) != 1 {
			t.Fatalf("password reset mail count=%d", len(links))
		}
		rawToken := tokenFromFragment(t, links[0])
		parsed, err := testActionTokenCodec(t).Parse(rawToken, authmodule.ActionTokenPurposePasswordReset)
		if err != nil {
			t.Fatal("password reset link contained an invalid token")
		}

		csrfToken := getCSRFTokenAt(t, primaryClient, serverURL)
		if !hasCookie(primaryClient.Jar, serverURL, "pdv_refresh") {
			t.Fatal("reset setup did not retain the authenticated refresh cookie")
		}
		reset := authRequestAt(t, primaryClient, serverURL, http.MethodPost, "/auth/password/reset", map[string]any{"token": rawToken, "newPassword": actionTokenNewPassword}, csrfToken, "")
		if reset.StatusCode != http.StatusNoContent {
			t.Fatalf("valid reset status=%d: %s", reset.StatusCode, readBody(reset))
		}
		assertCookiesCleared(t, primaryClient, serverURL, reset)
		reset.Body.Close()

		var versionAfter int64
		var consumed bool
		if err := testPool.QueryRow(context.Background(), `SELECT password_version FROM users WHERE id=$1`, registered.User.ID).Scan(&versionAfter); err != nil {
			t.Fatal(err)
		}
		if err := testPool.QueryRow(context.Background(), `SELECT consumed_at IS NOT NULL FROM auth_action_tokens WHERE id=$1`, parsed.Selector).Scan(&consumed); err != nil {
			t.Fatal(err)
		}
		stats := sessionStatsForUser(t, registered.User.ID)
		if versionAfter != versionBefore+1 || !consumed || stats.totalSessions != 2 || stats.activeSessions != 0 || stats.totalRefresh != 2 || stats.unrevokedRefresh != 0 || stats.expectedReasons != 2 {
			t.Fatalf("reset persistence mismatch: version=%d->%d consumed=%v stats=%+v", versionBefore, versionAfter, consumed, stats)
		}

		for name, oldAuth := range map[string]authResponse{"registered": registered, "secondary": secondary} {
			oldMe := authRequestAt(t, newHTTPClient(t), serverURL, http.MethodGet, "/me", nil, "", oldAuth.AccessToken)
			if oldMe.StatusCode != http.StatusUnauthorized {
				t.Fatalf("%s old access status=%d", name, oldMe.StatusCode)
			}
			oldMe.Body.Close()
		}

		csrfToken = getCSRFTokenAt(t, primaryClient, serverURL)
		reused := authRequestAt(t, primaryClient, serverURL, http.MethodPost, "/auth/password/reset", map[string]any{"token": rawToken, "newPassword": "Task 11C another password 2028!"}, csrfToken, "")
		if reused.StatusCode != http.StatusBadRequest {
			t.Fatalf("reused reset status=%d", reused.StatusCode)
		}
		if code := responseErrorCode(t, reused); code != "INVALID_REQUEST" {
			t.Fatalf("reused reset code=%q", code)
		}

		loginClient := newHTTPClient(t)
		csrfToken = getCSRFTokenAt(t, loginClient, serverURL)
		oldLogin := authRequestAt(t, loginClient, serverURL, http.MethodPost, "/auth/login", loginPayload(email, actionTokenOldPassword, "Old Reset Password"), csrfToken, "")
		if oldLogin.StatusCode != http.StatusUnauthorized {
			t.Fatalf("old password login status=%d", oldLogin.StatusCode)
		}
		if code := responseErrorCode(t, oldLogin); code != "INVALID_CREDENTIALS" {
			t.Fatalf("old password login code=%q", code)
		}
		newLogin := loginOwner(t, loginClient, serverURL, email, actionTokenNewPassword, "New Reset Password")
		if newLogin.AccessToken == "" {
			t.Fatal("new password login did not issue access")
		}

		var completionAudits int
		if err := testPool.QueryRow(context.Background(), `SELECT COUNT(*) FROM security_audit_events WHERE actor_user_id=$1 AND event_type='auth.password.reset_completed' AND outcome='SUCCESS'`, registered.User.ID).Scan(&completionAudits); err != nil {
			t.Fatal(err)
		}
		if completionAudits != 1 {
			t.Fatalf("password reset completion audits=%d", completionAudits)
		}
	})

	t.Run("patch me changes only display name and get me reflects it", func(t *testing.T) {
		resetRateLimitState(t)
		serverURL, client := auxiliaryAuthServer(t, true, false, &captureMailer{})
		email := "profile.task11c@example.com"
		registered := registerOwner(t, client, serverURL, email, "profile-task11c", actionTokenOldPassword)

		type persistedIdentity struct {
			email, normalized, displayName, status, verifiedAt, passwordHash string
			passwordVersion                                                  int64
		}
		loadIdentity := func() persistedIdentity {
			t.Helper()
			var value persistedIdentity
			if err := testPool.QueryRow(context.Background(), `
				SELECT u.email, u.email_normalized, u.display_name, u.status::text,
				       COALESCE(u.email_verified_at::text, ''), u.password_version, p.password_hash
				FROM users u JOIN user_passwords p ON p.user_id=u.id
				WHERE u.id=$1
			`, registered.User.ID).Scan(&value.email, &value.normalized, &value.displayName, &value.status, &value.verifiedAt, &value.passwordVersion, &value.passwordHash); err != nil {
				t.Fatal(err)
			}
			return value
		}
		before := loadIdentity()

		updated := authRequestAt(t, client, serverURL, http.MethodPatch, "/me", map[string]any{"displayName": "  Task 11C Updated Owner  "}, "", registered.AccessToken)
		if updated.StatusCode != http.StatusOK {
			t.Fatalf("patch me status=%d: %s", updated.StatusCode, readBody(updated))
		}
		var updatedUser struct {
			ID, Email, DisplayName string
			EmailVerified          bool `json:"emailVerified"`
		}
		decodeResponse(t, updated, &updatedUser)
		if updatedUser.ID != registered.User.ID || updatedUser.Email != email || updatedUser.DisplayName != "Task 11C Updated Owner" || !updatedUser.EmailVerified {
			t.Fatalf("patch me response=%+v", updatedUser)
		}

		me := authRequestAt(t, client, serverURL, http.MethodGet, "/me", nil, "", registered.AccessToken)
		if me.StatusCode != http.StatusOK {
			t.Fatalf("get me status=%d: %s", me.StatusCode, readBody(me))
		}
		var meBody struct {
			User struct {
				ID, Email, DisplayName string
				EmailVerified          bool `json:"emailVerified"`
			} `json:"user"`
		}
		decodeResponse(t, me, &meBody)
		if meBody.User.ID != registered.User.ID || meBody.User.Email != email || meBody.User.DisplayName != "Task 11C Updated Owner" || !meBody.User.EmailVerified {
			t.Fatalf("get me user=%+v", meBody.User)
		}

		after := loadIdentity()
		if after.displayName != "Task 11C Updated Owner" || before.email != after.email || before.normalized != after.normalized || before.status != after.status || before.verifiedAt != after.verifiedAt || before.passwordVersion != after.passwordVersion || before.passwordHash != after.passwordHash {
			t.Fatal("patch me changed a persisted field other than display name")
		}
	})

	t.Run("authenticated password change verifies current password and revokes every session", func(t *testing.T) {
		resetRateLimitState(t)
		serverURL, primaryClient := auxiliaryAuthServer(t, true, false, &captureMailer{})
		email := "password-change.task11c@example.com"
		registered := registerOwner(t, primaryClient, serverURL, email, "password-change-task11c", actionTokenOldPassword)
		secondaryClient := newHTTPClient(t)
		secondary := loginOwner(t, secondaryClient, serverURL, email, actionTokenOldPassword, "Password Secondary")

		var versionBefore int64
		if err := testPool.QueryRow(context.Background(), `SELECT password_version FROM users WHERE id=$1`, registered.User.ID).Scan(&versionBefore); err != nil {
			t.Fatal(err)
		}
		wrong := authRequestAt(t, primaryClient, serverURL, http.MethodPost, "/me/password", map[string]any{"currentPassword": "Task 11C definitely wrong 2026!", "newPassword": actionTokenNewPassword}, "", registered.AccessToken)
		if wrong.StatusCode != http.StatusUnauthorized {
			t.Fatalf("wrong current password status=%d", wrong.StatusCode)
		}
		if code := responseErrorCode(t, wrong); code != "INVALID_CREDENTIALS" {
			t.Fatalf("wrong current password code=%q", code)
		}
		var versionAfterWrong int64
		if err := testPool.QueryRow(context.Background(), `SELECT password_version FROM users WHERE id=$1`, registered.User.ID).Scan(&versionAfterWrong); err != nil {
			t.Fatal(err)
		}
		wrongStats := sessionStatsForUser(t, registered.User.ID)
		if versionAfterWrong != versionBefore || wrongStats.activeSessions != 2 || !hasCookie(primaryClient.Jar, serverURL, "pdv_refresh") {
			t.Fatalf("wrong current password changed state: version=%d->%d stats=%+v", versionBefore, versionAfterWrong, wrongStats)
		}

		changed := authRequestAt(t, primaryClient, serverURL, http.MethodPost, "/me/password", map[string]any{"currentPassword": actionTokenOldPassword, "newPassword": actionTokenNewPassword}, "", registered.AccessToken)
		if changed.StatusCode != http.StatusNoContent {
			t.Fatalf("password change status=%d: %s", changed.StatusCode, readBody(changed))
		}
		assertCookiesCleared(t, primaryClient, serverURL, changed)
		changed.Body.Close()

		var versionAfter int64
		if err := testPool.QueryRow(context.Background(), `SELECT password_version FROM users WHERE id=$1`, registered.User.ID).Scan(&versionAfter); err != nil {
			t.Fatal(err)
		}
		stats := sessionStatsForUserWithReason(t, registered.User.ID, "password_changed")
		if versionAfter != versionBefore+1 || stats.totalSessions != 2 || stats.activeSessions != 0 || stats.totalRefresh != 2 || stats.unrevokedRefresh != 0 || stats.expectedReasons != 2 {
			t.Fatalf("password change persistence mismatch: version=%d->%d stats=%+v", versionBefore, versionAfter, stats)
		}

		for name, oldAuth := range map[string]authResponse{"registered": registered, "secondary": secondary} {
			oldMe := authRequestAt(t, newHTTPClient(t), serverURL, http.MethodGet, "/me", nil, "", oldAuth.AccessToken)
			if oldMe.StatusCode != http.StatusUnauthorized {
				t.Fatalf("%s old access status=%d", name, oldMe.StatusCode)
			}
			oldMe.Body.Close()
		}

		loginClient := newHTTPClient(t)
		csrfToken := getCSRFTokenAt(t, loginClient, serverURL)
		oldLogin := authRequestAt(t, loginClient, serverURL, http.MethodPost, "/auth/login", loginPayload(email, actionTokenOldPassword, "Old Changed Password"), csrfToken, "")
		if oldLogin.StatusCode != http.StatusUnauthorized {
			t.Fatalf("old changed password login status=%d", oldLogin.StatusCode)
		}
		if code := responseErrorCode(t, oldLogin); code != "INVALID_CREDENTIALS" {
			t.Fatalf("old changed password login code=%q", code)
		}
		newLogin := loginOwner(t, loginClient, serverURL, email, actionTokenNewPassword, "New Changed Password")
		if newLogin.AccessToken == "" {
			t.Fatal("changed password login did not issue access")
		}

		var audits int
		if err := testPool.QueryRow(context.Background(), `SELECT COUNT(*) FROM security_audit_events WHERE actor_user_id=$1 AND event_type='auth.password.changed' AND outcome='SUCCESS'`, registered.User.ID).Scan(&audits); err != nil {
			t.Fatal(err)
		}
		if audits != 1 {
			t.Fatalf("password changed audits=%d", audits)
		}
	})

	t.Run("verification resend rotates token and verification is idempotent", func(t *testing.T) {
		resetRateLimitState(t)
		mail := &captureMailer{}
		serverURL, client := auxiliaryAuthServer(t, true, true, mail)
		email := "verification.task11c@example.com"
		registerWithVerification(t, client, serverURL, email, "verification-task11c", actionTokenOldPassword)

		verificationLinks, committed := verificationEmailLinks(mail)
		if len(verificationLinks) != 1 || !committed {
			t.Fatalf("registration verification mail count=%d committed=%v", len(verificationLinks), committed)
		}
		oldRaw := tokenFromFragment(t, verificationLinks[0])
		codec := testActionTokenCodec(t)
		oldParsed, err := codec.Parse(oldRaw, authmodule.ActionTokenPurposeEmailVerification)
		if err != nil {
			t.Fatal("registration verification link contained an invalid token")
		}

		acceptedEmailAction(t, client, serverURL, "/auth/email/resend-verification", email)
		verificationLinks, committed = verificationEmailLinks(mail)
		if len(verificationLinks) != 2 || !committed {
			t.Fatalf("resend verification mail count=%d committed=%v", len(verificationLinks), committed)
		}
		newRaw := tokenFromFragment(t, verificationLinks[1])
		newParsed, err := codec.Parse(newRaw, authmodule.ActionTokenPurposeEmailVerification)
		if err != nil || oldParsed.Selector == newParsed.Selector {
			t.Fatal("resend verification produced an invalid or duplicate selector")
		}

		var oldConsumed, newConsumed bool
		var verificationTTL int64
		if err := testPool.QueryRow(context.Background(), `SELECT consumed_at IS NOT NULL FROM auth_action_tokens WHERE id=$1`, oldParsed.Selector).Scan(&oldConsumed); err != nil {
			t.Fatal(err)
		}
		if err := testPool.QueryRow(context.Background(), `SELECT consumed_at IS NOT NULL, EXTRACT(EPOCH FROM (expires_at-created_at))::BIGINT FROM auth_action_tokens WHERE id=$1`, newParsed.Selector).Scan(&newConsumed, &verificationTTL); err != nil {
			t.Fatal(err)
		}
		if !oldConsumed || newConsumed || verificationTTL != 86400 {
			t.Fatalf("verification rotation state old=%v new=%v ttl=%d", oldConsumed, newConsumed, verificationTTL)
		}

		csrfToken := getCSRFTokenAt(t, client, serverURL)
		oldVerify := authRequestAt(t, client, serverURL, http.MethodPost, "/auth/email/verify", map[string]any{"token": oldRaw}, csrfToken, "")
		if oldVerify.StatusCode != http.StatusBadRequest {
			t.Fatalf("old verification token status=%d", oldVerify.StatusCode)
		}
		if code := responseErrorCode(t, oldVerify); code != "INVALID_REQUEST" {
			t.Fatalf("old verification token code=%q", code)
		}

		for attempt := 1; attempt <= 2; attempt++ {
			csrfToken = getCSRFTokenAt(t, client, serverURL)
			verified := authRequestAt(t, client, serverURL, http.MethodPost, "/auth/email/verify", map[string]any{"token": newRaw}, csrfToken, "")
			if verified.StatusCode != http.StatusNoContent {
				t.Fatalf("verification attempt %d status=%d: %s", attempt, verified.StatusCode, readBody(verified))
			}
			verified.Body.Close()
		}

		var verifiedAt bool
		var audits int
		if err := testPool.QueryRow(context.Background(), `SELECT email_verified_at IS NOT NULL FROM users WHERE email_normalized=$1`, email).Scan(&verifiedAt); err != nil {
			t.Fatal(err)
		}
		if err := testPool.QueryRow(context.Background(), `
			SELECT COUNT(*)
			FROM security_audit_events e
			JOIN users u ON u.id=e.actor_user_id
			WHERE u.email_normalized=$1 AND e.event_type='auth.email.verified' AND e.outcome='SUCCESS'
		`, email).Scan(&audits); err != nil {
			t.Fatal(err)
		}
		if !verifiedAt || audits != 1 {
			t.Fatalf("verification persistence verified=%v audits=%d", verifiedAt, audits)
		}
		verificationLinks, _ = verificationEmailLinks(mail)
		if len(verificationLinks) != 2 {
			t.Fatalf("unexpected verification mail count after verify=%d", len(verificationLinks))
		}
		loggedIn := loginOwner(t, newHTTPClient(t), serverURL, email, actionTokenOldPassword, "Verified Login")
		if !loggedIn.User.EmailVerified {
			t.Fatal("verified login did not reflect verified email")
		}
	})

	t.Run("expired reset token returns gone", func(t *testing.T) {
		resetRateLimitState(t)
		serverURL, client := auxiliaryAuthServer(t, true, false, &captureMailer{})
		registered := registerOwner(t, client, serverURL, "expired-reset.task11c@example.com", "expired-reset-task11c", actionTokenOldPassword)
		userID := parseUUID(t, registered.User.ID)
		now := time.Now().UTC()
		rawToken, selector := createDirectResetToken(t, userID, now.Add(-2*time.Hour), now.Add(-time.Hour))

		var versionBefore int64
		if err := testPool.QueryRow(context.Background(), `SELECT password_version FROM users WHERE id=$1`, userID).Scan(&versionBefore); err != nil {
			t.Fatal(err)
		}
		csrfToken := getCSRFTokenAt(t, client, serverURL)
		expired := authRequestAt(t, client, serverURL, http.MethodPost, "/auth/password/reset", map[string]any{"token": rawToken, "newPassword": actionTokenNewPassword}, csrfToken, "")
		if expired.StatusCode != http.StatusGone {
			t.Fatalf("expired reset status=%d", expired.StatusCode)
		}
		if code := responseErrorCode(t, expired); code != "ACTION_TOKEN_EXPIRED" {
			t.Fatalf("expired reset code=%q", code)
		}

		var versionAfter int64
		var consumed bool
		if err := testPool.QueryRow(context.Background(), `SELECT password_version FROM users WHERE id=$1`, userID).Scan(&versionAfter); err != nil {
			t.Fatal(err)
		}
		if err := testPool.QueryRow(context.Background(), `SELECT consumed_at IS NOT NULL FROM auth_action_tokens WHERE id=$1`, selector).Scan(&consumed); err != nil {
			t.Fatal(err)
		}
		if versionAfter != versionBefore || consumed {
			t.Fatalf("expired reset changed state: version=%d->%d consumed=%v", versionBefore, versionAfter, consumed)
		}
	})

	t.Run("concurrent reset consumes once without deadlock", func(t *testing.T) {
		resetRateLimitState(t)
		serverURL, primaryClient := auxiliaryAuthServer(t, true, false, &captureMailer{})
		email := "concurrent-reset.task11c@example.com"
		registered := registerOwner(t, primaryClient, serverURL, email, "concurrent-reset-task11c", actionTokenOldPassword)
		secondaryClient := newHTTPClient(t)
		secondary := loginOwner(t, secondaryClient, serverURL, email, actionTokenOldPassword, "Concurrent Secondary")
		userID := parseUUID(t, registered.User.ID)
		now := time.Now().UTC()
		rawToken, selector := createDirectResetToken(t, userID, now, now.Add(30*time.Minute))

		var versionBefore int64
		if err := testPool.QueryRow(context.Background(), `SELECT password_version FROM users WHERE id=$1`, userID).Scan(&versionBefore); err != nil {
			t.Fatal(err)
		}
		raceClients := []*http.Client{newHTTPClient(t), newHTTPClient(t)}
		csrfTokens := []string{
			getCSRFTokenAt(t, raceClients[0], serverURL),
			getCSRFTokenAt(t, raceClients[1], serverURL),
		}
		payload, err := json.Marshal(map[string]any{"token": rawToken, "newPassword": actionTokenNewPassword})
		if err != nil {
			t.Fatal(err)
		}
		type result struct {
			status int
			err    error
		}
		ready := make(chan struct{}, 2)
		start := make(chan struct{})
		results := make(chan result, 2)
		for i := range 2 {
			go func(client *http.Client, csrfToken string) {
				ready <- struct{}{}
				<-start
				ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
				defer cancel()
				req, requestErr := http.NewRequestWithContext(ctx, http.MethodPost, serverURL+"/auth/password/reset", bytes.NewReader(payload))
				if requestErr != nil {
					results <- result{err: requestErr}
					return
				}
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Origin", serverURL)
				req.Header.Set("Sec-Fetch-Site", "same-origin")
				req.Header.Set("X-CSRF-Token", csrfToken)
				resp, requestErr := client.Do(req)
				if requestErr != nil {
					results <- result{err: requestErr}
					return
				}
				resp.Body.Close()
				results <- result{status: resp.StatusCode}
			}(raceClients[i], csrfTokens[i])
		}
		<-ready
		<-ready
		close(start)

		statuses := map[int]int{}
		timer := time.NewTimer(10 * time.Second)
		defer timer.Stop()
		for range 2 {
			select {
			case outcome := <-results:
				if outcome.err != nil {
					t.Fatalf("concurrent reset request failed: %v", outcome.err)
				}
				statuses[outcome.status]++
			case <-timer.C:
				t.Fatal("concurrent reset requests deadlocked")
			}
		}
		if statuses[http.StatusNoContent] != 1 || statuses[http.StatusBadRequest] != 1 || len(statuses) != 2 {
			t.Fatalf("concurrent reset statuses=%v", statuses)
		}

		var versionAfter int64
		var consumed bool
		if err := testPool.QueryRow(context.Background(), `SELECT password_version FROM users WHERE id=$1`, userID).Scan(&versionAfter); err != nil {
			t.Fatal(err)
		}
		if err := testPool.QueryRow(context.Background(), `SELECT consumed_at IS NOT NULL FROM auth_action_tokens WHERE id=$1`, selector).Scan(&consumed); err != nil {
			t.Fatal(err)
		}
		stats := sessionStatsForUser(t, registered.User.ID)
		var audits int
		if err := testPool.QueryRow(context.Background(), `SELECT COUNT(*) FROM security_audit_events WHERE actor_user_id=$1 AND event_type='auth.password.reset_completed' AND outcome='SUCCESS'`, userID).Scan(&audits); err != nil {
			t.Fatal(err)
		}
		if versionAfter != versionBefore+1 || !consumed || audits != 1 || stats.totalSessions != 2 || stats.activeSessions != 0 || stats.totalRefresh != 2 || stats.unrevokedRefresh != 0 || stats.expectedReasons != 2 {
			t.Fatalf("concurrent reset persistence mismatch: version=%d->%d consumed=%v audits=%d stats=%+v", versionBefore, versionAfter, consumed, audits, stats)
		}
		for name, oldAuth := range map[string]authResponse{"registered": registered, "secondary": secondary} {
			oldMe := authRequestAt(t, newHTTPClient(t), serverURL, http.MethodGet, "/me", nil, "", oldAuth.AccessToken)
			if oldMe.StatusCode != http.StatusUnauthorized {
				t.Fatalf("%s concurrent-reset old access status=%d", name, oldMe.StatusCode)
			}
			oldMe.Body.Close()
		}
	})

	t.Run("forgot and resend enforce email and IP limits", func(t *testing.T) {
		mail := &captureMailer{}
		serverURL, client := auxiliaryAuthServer(t, true, true, mail)

		for _, endpoint := range []struct {
			name, path, emailPrefix string
		}{
			{name: "forgot", path: "/auth/password/forgot", emailPrefix: "forgot-rate"},
			{name: "resend", path: "/auth/email/resend-verification", emailPrefix: "resend-rate"},
		} {
			t.Run(endpoint.name+" email", func(t *testing.T) {
				resetRateLimitState(t)
				csrfToken := getCSRFTokenAt(t, client, serverURL)
				email := endpoint.emailPrefix + "-email.task11c@example.com"
				for attempt := 1; attempt <= 3; attempt++ {
					resp := authRequestAt(t, client, serverURL, http.MethodPost, endpoint.path, map[string]any{"email": email}, csrfToken, "")
					if resp.StatusCode != http.StatusAccepted {
						t.Fatalf("%s email attempt %d status=%d: %s", endpoint.name, attempt, resp.StatusCode, readBody(resp))
					}
					resp.Body.Close()
				}
				limited := authRequestAt(t, client, serverURL, http.MethodPost, endpoint.path, map[string]any{"email": email}, csrfToken, "")
				assertRateLimited(t, endpoint.name+" email", limited)
			})

			t.Run(endpoint.name+" IP", func(t *testing.T) {
				resetRateLimitState(t)
				csrfToken := getCSRFTokenAt(t, client, serverURL)
				for attempt := range 10 {
					email := fmt.Sprintf("%s-ip-%02d.task11c@example.com", endpoint.emailPrefix, attempt)
					resp := authRequestAt(t, client, serverURL, http.MethodPost, endpoint.path, map[string]any{"email": email}, csrfToken, "")
					if resp.StatusCode != http.StatusAccepted {
						t.Fatalf("%s IP attempt %d status=%d: %s", endpoint.name, attempt+1, resp.StatusCode, readBody(resp))
					}
					resp.Body.Close()
				}
				limitedEmail := endpoint.emailPrefix + "-ip-limited.task11c@example.com"
				limited := authRequestAt(t, client, serverURL, http.MethodPost, endpoint.path, map[string]any{"email": limitedEmail}, csrfToken, "")
				assertRateLimited(t, endpoint.name+" IP", limited)
			})
		}

		passwordLinks, _ := passwordResetLinks(mail)
		verificationLinks, _ := verificationEmailLinks(mail)
		if len(passwordLinks) != 0 || len(verificationLinks) != 0 {
			t.Fatalf("unknown rate-limit requests sent mail: password=%d verification=%d", len(passwordLinks), len(verificationLinks))
		}
	})
}

func newHTTPClient(t *testing.T) *http.Client {
	t.Helper()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	return &http.Client{Jar: jar, Timeout: 10 * time.Second}
}

func registerOwner(t *testing.T, client *http.Client, serverURL, email, slug, password string) authResponse {
	t.Helper()
	csrfToken := getCSRFTokenAt(t, client, serverURL)
	resp := authRequestAt(t, client, serverURL, http.MethodPost, "/auth/register", ownerRegisterPayload(email, slug, password), csrfToken, "")
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register %s status=%d: %s", email, resp.StatusCode, readBody(resp))
	}
	var result authResponse
	decodeResponse(t, resp, &result)
	if result.AccessToken == "" || result.User.ID == "" || result.User.Email != email {
		t.Fatalf("register %s returned incomplete auth response", email)
	}
	return result
}

func registerWithVerification(t *testing.T, client *http.Client, serverURL, email, slug, password string) {
	t.Helper()
	csrfToken := getCSRFTokenAt(t, client, serverURL)
	resp := authRequestAt(t, client, serverURL, http.MethodPost, "/auth/register", ownerRegisterPayload(email, slug, password), csrfToken, "")
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("verification-required register %s status=%d: %s", email, resp.StatusCode, readBody(resp))
	}
	var body struct {
		Status string `json:"status"`
	}
	decodeResponse(t, resp, &body)
	if body.Status != "VERIFICATION_REQUIRED" {
		t.Fatalf("verification-required register %s status body=%q", email, body.Status)
	}
}

func ownerRegisterPayload(email, slug, password string) map[string]any {
	return map[string]any{
		"email": email, "password": password, "displayName": "Task 11C Owner",
		"organization": map[string]any{"name": "Task 11C " + slug, "slug": slug, "timezone": "America/Sao_Paulo", "locale": "pt-BR", "currency": "BRL"},
		"store":        map[string]any{"code": "MATRIZ", "name": "Matriz", "timezone": "America/Sao_Paulo"},
		"clientId":     "pdv-admin", "deviceName": "Task 11C Browser",
	}
}

func loginPayload(email, password, deviceName string) map[string]any {
	return map[string]any{"email": email, "password": password, "clientId": "pdv-admin", "deviceName": deviceName}
}

func loginOwner(t *testing.T, client *http.Client, serverURL, email, password, deviceName string) authResponse {
	t.Helper()
	csrfToken := getCSRFTokenAt(t, client, serverURL)
	resp := authRequestAt(t, client, serverURL, http.MethodPost, "/auth/login", loginPayload(email, password, deviceName), csrfToken, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login %s status=%d: %s", email, resp.StatusCode, readBody(resp))
	}
	var result authResponse
	decodeResponse(t, resp, &result)
	if result.AccessToken == "" || result.Session.ID == "" {
		t.Fatalf("login %s returned incomplete auth response", email)
	}
	return result
}

func acceptedEmailAction(t *testing.T, client *http.Client, serverURL, path, email string) {
	t.Helper()
	csrfToken := getCSRFTokenAt(t, client, serverURL)
	resp := authRequestAt(t, client, serverURL, http.MethodPost, path, map[string]any{"email": email}, csrfToken, "")
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("%s for %s status=%d: %s", path, email, resp.StatusCode, readBody(resp))
	}
	var body struct {
		Status string `json:"status"`
	}
	decodeResponse(t, resp, &body)
	if body.Status != "ACCEPTED" {
		t.Fatalf("%s for %s response status=%q", path, email, body.Status)
	}
}

func tokenFromFragment(t *testing.T, link string) string {
	t.Helper()
	parsed, err := url.Parse(link)
	if err != nil || parsed.RawQuery != "" || parsed.Fragment == "" {
		t.Fatal("mail link did not contain an isolated token fragment")
	}
	fragment, err := url.ParseQuery(parsed.Fragment)
	if err != nil {
		t.Fatal("mail link token fragment was malformed")
	}
	values, ok := fragment["token"]
	if !ok || len(values) != 1 || values[0] == "" || len(fragment) != 1 {
		t.Fatal("mail link did not contain exactly one fragment token")
	}
	return values[0]
}

func passwordResetLinks(mail *captureMailer) ([]string, bool) {
	mail.mu.Lock()
	defer mail.mu.Unlock()
	return append([]string(nil), mail.passwordResetLinks...), mail.committed
}

func verificationEmailLinks(mail *captureMailer) ([]string, bool) {
	mail.mu.Lock()
	defer mail.mu.Unlock()
	return append([]string(nil), mail.verificationLinks...), mail.committed
}

func testActionTokenCodec(t *testing.T) authmodule.ActionTokenCodec {
	t.Helper()
	codec, err := authmodule.NewActionTokenCodec([]byte("abcdef0123456789abcdef0123456789"))
	if err != nil {
		t.Fatal(err)
	}
	return codec
}

func parseUUID(t *testing.T, value string) pgtype.UUID {
	t.Helper()
	var id pgtype.UUID
	if err := id.Scan(value); err != nil || !id.Valid {
		t.Fatalf("invalid persisted UUID %q", value)
	}
	return id
}

func createDirectResetToken(t *testing.T, userID pgtype.UUID, createdAt, expiresAt time.Time) (string, pgtype.UUID) {
	t.Helper()
	var selector pgtype.UUID
	if err := testPool.QueryRow(context.Background(), `SELECT uuidv7()`).Scan(&selector); err != nil {
		t.Fatal(err)
	}
	rawToken, secretHash, err := testActionTokenCodec(t).Generate(authmodule.ActionTokenPurposePasswordReset, selector)
	if err != nil {
		t.Fatal("could not generate direct password reset token")
	}
	if _, err := testPool.Exec(context.Background(), `
		INSERT INTO auth_action_tokens (id, user_id, purpose, secret_hash, expires_at, created_at)
		VALUES ($1, $2, 'PASSWORD_RESET', $3, $4, $5)
	`, selector, userID, secretHash, expiresAt, createdAt); err != nil {
		t.Fatal(err)
	}
	return rawToken, selector
}

type sessionStats struct {
	totalSessions, activeSessions, totalRefresh, unrevokedRefresh, expectedReasons int
}

func sessionStatsForUser(t *testing.T, userID string) sessionStats {
	t.Helper()
	return sessionStatsForUserWithReason(t, userID, "password_reset")
}

func sessionStatsForUserWithReason(t *testing.T, userID, reason string) sessionStats {
	t.Helper()
	var stats sessionStats
	if err := testPool.QueryRow(context.Background(), `
		SELECT
			(SELECT COUNT(*) FROM auth_sessions WHERE user_id=$1),
			(SELECT COUNT(*) FROM auth_sessions WHERE user_id=$1 AND status='ACTIVE'),
			(SELECT COUNT(*) FROM auth_refresh_tokens r JOIN auth_sessions s ON s.id=r.session_id WHERE s.user_id=$1),
			(SELECT COUNT(*) FROM auth_refresh_tokens r JOIN auth_sessions s ON s.id=r.session_id WHERE s.user_id=$1 AND r.revoked_at IS NULL),
			(SELECT COUNT(*) FROM auth_sessions WHERE user_id=$1 AND status='REVOKED' AND revoke_reason=$2)
	`, userID, reason).Scan(&stats.totalSessions, &stats.activeSessions, &stats.totalRefresh, &stats.unrevokedRefresh, &stats.expectedReasons); err != nil {
		t.Fatal(err)
	}
	return stats
}

func assertCookiesCleared(t *testing.T, client *http.Client, serverURL string, resp *http.Response) {
	t.Helper()
	deleted := map[string]bool{"pdv_refresh": false, "pdv_csrf": false}
	for _, cookie := range resp.Cookies() {
		if _, ok := deleted[cookie.Name]; ok && cookie.Value == "" && cookie.MaxAge < 0 {
			deleted[cookie.Name] = true
		}
	}
	if !deleted["pdv_refresh"] || !deleted["pdv_csrf"] {
		t.Fatalf("auth cookie deletion headers missing: %+v", deleted)
	}
	if hasCookie(client.Jar, serverURL, "pdv_refresh") || hasCookie(client.Jar, serverURL, "pdv_csrf") {
		t.Fatal("auth cookies remained in cookie jar after credential change")
	}
}

func hasCookie(jar http.CookieJar, serverURL, name string) bool {
	parsed, err := url.Parse(serverURL)
	if err != nil {
		return false
	}
	for _, cookie := range jar.Cookies(parsed) {
		if cookie.Name == name {
			return true
		}
	}
	return false
}

func assertRateLimited(t *testing.T, name string, resp *http.Response) {
	t.Helper()
	if resp.StatusCode != http.StatusTooManyRequests || strings.TrimSpace(resp.Header.Get("Retry-After")) == "" {
		t.Fatalf("%s rate limit status=%d retry-after=%q: %s", name, resp.StatusCode, resp.Header.Get("Retry-After"), readBody(resp))
	}
	if code := responseErrorCode(t, resp); code != "RATE_LIMITED" {
		t.Fatalf("%s rate limit code=%q", name, code)
	}
}

func resetRateLimitState(t *testing.T) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := testValkey.Do(ctx, testValkey.B().Flushdb().Build()); err != nil {
		t.Fatalf("flush isolated Task 11C Valkey DB: %v", err)
	}
}
