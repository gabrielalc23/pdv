package e2e

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
)

const administrationPassword = "Task 12 secure password 2026!"

func TestAdministration(t *testing.T) {
	resetRateLimitState(t)
	mail := &captureMailer{}
	serverURL, client := auxiliaryAuthServer(t, true, false, mail)
	owner := registerOwner(t, client, serverURL, "owner.task12@example.com", "owner-task12", administrationPassword)
	if owner.Context.Organization == nil || owner.Context.Store == nil || owner.Context.MembershipID == nil {
		t.Fatalf("owner registration returned incomplete tenant context: %+v", owner.Context)
	}
	organizationID := owner.Context.Organization.ID
	membershipID := *owner.Context.MembershipID

	t.Run("authenticated organization list and current", func(t *testing.T) {
		listResponse := authRequestAt(t, client, serverURL, http.MethodGet, "/me/organizations", nil, "", owner.AccessToken)
		if listResponse.StatusCode != http.StatusOK {
			t.Fatalf("list organizations status=%d: %s", listResponse.StatusCode, readBody(listResponse))
		}
		var list struct {
			Data []struct {
				MembershipID string `json:"membershipId"`
				Organization struct {
					ID   string `json:"id"`
					Slug string `json:"slug"`
				} `json:"organization"`
			} `json:"data"`
		}
		decodeResponse(t, listResponse, &list)
		if len(list.Data) != 1 || list.Data[0].MembershipID != membershipID || list.Data[0].Organization.ID != organizationID || list.Data[0].Organization.Slug != "owner-task12" {
			t.Fatalf("unexpected organization list: %+v", list.Data)
		}

		currentResponse := authRequestAt(t, client, serverURL, http.MethodGet, "/v1/organizations/current", nil, "", owner.AccessToken)
		if currentResponse.StatusCode != http.StatusOK {
			t.Fatalf("current organization status=%d: %s", currentResponse.StatusCode, readBody(currentResponse))
		}
		var current struct {
			ID     string `json:"id"`
			Slug   string `json:"slug"`
			Status string `json:"status"`
		}
		decodeResponse(t, currentResponse, &current)
		if current.ID != organizationID || current.Slug != "owner-task12" || current.Status != "ACTIVE" {
			t.Fatalf("unexpected current organization: %+v", current)
		}
	})

	t.Run("admin route requires access token", func(t *testing.T) {
		response := authRequestAt(t, newHTTPClient(t), serverURL, http.MethodGet, "/me/organizations", nil, "", "")
		assertErrorResponse(t, response, http.StatusUnauthorized, "ACCESS_TOKEN_MISSING")
		if response.Header.Get("WWW-Authenticate") != "Bearer" {
			t.Fatalf("unauthorized route challenge=%q", response.Header.Get("WWW-Authenticate"))
		}
	})

	t.Run("cross-tenant membership is not found", func(t *testing.T) {
		if organizationID == testOrgID {
			t.Fatal("cross-tenant fixture unexpectedly belongs to the owner organization")
		}
		response := authRequestAt(t, client, serverURL, http.MethodGet, "/v1/members/"+testMembershipID, nil, "", owner.AccessToken)
		assertErrorResponse(t, response, http.StatusNotFound, "MEMBERSHIP_NOT_FOUND")
	})

	t.Run("disabled tenant creation rejects create without persistence", func(t *testing.T) {
		var before, after int
		if err := testPool.QueryRow(context.Background(), `SELECT COUNT(*) FROM organizations WHERE created_by_user_id=$1`, owner.User.ID).Scan(&before); err != nil {
			t.Fatal(err)
		}
		response := authRequestAt(t, client, serverURL, http.MethodPost, "/v1/organizations", map[string]any{
			"organization": map[string]any{
				"name": "Disabled Task 12", "slug": "disabled-task12", "timezone": "America/Sao_Paulo", "locale": "pt-BR", "currency": "BRL",
			},
			"store": map[string]any{"code": "MATRIZ", "name": "Matriz", "timezone": "America/Sao_Paulo"},
		}, "", owner.AccessToken)
		assertErrorResponse(t, response, http.StatusForbidden, "TENANT_CREATION_DISABLED")
		if err := testPool.QueryRow(context.Background(), `SELECT COUNT(*) FROM organizations WHERE created_by_user_id=$1`, owner.User.ID).Scan(&after); err != nil {
			t.Fatal(err)
		}
		if after != before {
			t.Fatalf("disabled tenant creation changed organization count: %d -> %d", before, after)
		}
	})

	t.Run("system owner role cannot be deactivated", func(t *testing.T) {
		listResponse := authRequestAt(t, client, serverURL, http.MethodGet, "/v1/roles", nil, "", owner.AccessToken)
		if listResponse.StatusCode != http.StatusOK {
			t.Fatalf("list roles status=%d: %s", listResponse.StatusCode, readBody(listResponse))
		}
		var roles struct {
			Data []struct {
				ID       string `json:"id"`
				Key      string `json:"key"`
				IsSystem bool   `json:"isSystem"`
				IsActive bool   `json:"isActive"`
			} `json:"data"`
		}
		decodeResponse(t, listResponse, &roles)
		var ownerRoleID string
		for _, role := range roles.Data {
			if role.Key == "owner" && role.IsSystem && role.IsActive {
				ownerRoleID = role.ID
				break
			}
		}
		if ownerRoleID == "" {
			t.Fatalf("active system owner role missing from response: %+v", roles.Data)
		}

		response := authRequestAt(t, client, serverURL, http.MethodPost, "/v1/roles/"+ownerRoleID+"/deactivate", nil, "", owner.AccessToken)
		assertErrorResponse(t, response, http.StatusConflict, "SYSTEM_ROLE_IMMUTABLE")
		var active bool
		if err := testPool.QueryRow(context.Background(), `SELECT is_active FROM roles WHERE organization_id=$1 AND id=$2`, organizationID, ownerRoleID).Scan(&active); err != nil {
			t.Fatal(err)
		}
		if !active {
			t.Fatal("system owner role was deactivated despite the invariant")
		}
	})

	t.Run("new user accepts invitation into authenticated organization session", func(t *testing.T) {
		var cashierRoleID string
		if err := testPool.QueryRow(context.Background(), `SELECT id FROM roles WHERE organization_id=$1 AND key='cashier'`, organizationID).Scan(&cashierRoleID); err != nil {
			t.Fatal(err)
		}
		invitedEmail := "cashier.task12@example.com"
		response := authRequestAt(t, client, serverURL, http.MethodPost, "/v1/invitations", map[string]any{
			"email":       invitedEmail,
			"assignments": []map[string]any{{"roleId": cashierRoleID, "storeId": owner.Context.Store.ID}},
		}, "", owner.AccessToken)
		if response.StatusCode != http.StatusCreated {
			t.Fatalf("create invitation status=%d: %s", response.StatusCode, readBody(response))
		}
		response.Body.Close()

		mail.mu.Lock()
		if len(mail.invitationLinks) != 1 {
			mail.mu.Unlock()
			t.Fatalf("invitation links=%d, want 1", len(mail.invitationLinks))
		}
		link := mail.invitationLinks[0]
		mail.mu.Unlock()
		rawToken := tokenFromFragment(t, link)

		inviteeClient := newHTTPClient(t)
		csrfToken := getCSRFTokenAt(t, inviteeClient, serverURL)
		accepted := authRequestAt(t, inviteeClient, serverURL, http.MethodPost, "/auth/invitations/accept", map[string]any{
			"token": rawToken, "displayName": "Task 12 Cashier", "password": administrationPassword,
			"clientId": "pdv-admin", "deviceName": "Task 12 E2E",
		}, csrfToken, "")
		if accepted.StatusCode != http.StatusOK {
			t.Fatalf("accept invitation status=%d: %s", accepted.StatusCode, readBody(accepted))
		}
		body := readBody(accepted)
		if bytes.Contains([]byte(body), []byte(rawToken)) || bytes.Contains([]byte(body), []byte("secretHash")) {
			t.Fatal("acceptance response exposed invitation secret")
		}
		var authenticated authResponse
		decodeJSONBytes(t, body, &authenticated)
		if authenticated.AccessToken == "" || authenticated.Context.Organization == nil || authenticated.Context.Organization.ID != organizationID {
			t.Fatalf("unexpected accepted context: %+v", authenticated.Context)
		}
		storeContextResponse := authRequestAt(t, inviteeClient, serverURL, http.MethodPost, "/auth/context", map[string]any{
			"organizationId": organizationID, "storeId": owner.Context.Store.ID,
		}, "", authenticated.AccessToken)
		if storeContextResponse.StatusCode != http.StatusOK {
			t.Fatalf("switch accepted user to store status=%d: %s", storeContextResponse.StatusCode, readBody(storeContextResponse))
		}
		var storeContext authResponse
		decodeResponse(t, storeContextResponse, &storeContext)
		if storeContext.Context.Store == nil || storeContext.Context.Store.ID != owner.Context.Store.ID || !contains(storeContext.Context.Roles, "cashier") {
			t.Fatalf("unexpected accepted store context: %+v", storeContext.Context)
		}

		var status string
		var storedHash []byte
		if err := testPool.QueryRow(context.Background(), `SELECT status, secret_hash FROM organization_invitations WHERE organization_id=$1 AND email_normalized=$2`, organizationID, invitedEmail).Scan(&status, &storedHash); err != nil {
			t.Fatal(err)
		}
		if status != "ACCEPTED" || len(storedHash) != 32 || bytes.Equal(storedHash, []byte(rawToken)) {
			t.Fatalf("invitation persistence status=%s hash_length=%d", status, len(storedHash))
		}
	})

	t.Run("missing scope returns forbidden", func(t *testing.T) {
		limitedOrganizationID, limitedStoreID := createLimitedOrgContext(t, owner.User.ID)
		contextResponse := authRequestAt(t, client, serverURL, http.MethodPost, "/auth/context", map[string]any{
			"organizationId": limitedOrganizationID,
			"storeId":        limitedStoreID,
		}, "", owner.AccessToken)
		if contextResponse.StatusCode != http.StatusOK {
			t.Fatalf("switch to limited context status=%d: %s", contextResponse.StatusCode, readBody(contextResponse))
		}
		var limited authResponse
		decodeResponse(t, contextResponse, &limited)
		if limited.Context.Store == nil || limited.Context.Store.ID != limitedStoreID || !contains(limited.Context.Scopes, "catalog.read") || contains(limited.Context.Scopes, "roles.read") {
			t.Fatalf("unexpected limited authorization context: %+v", limited.Context)
		}

		response := authRequestAt(t, client, serverURL, http.MethodGet, "/v1/roles", nil, "", limited.AccessToken)
		assertErrorResponse(t, response, http.StatusForbidden, "INSUFFICIENT_SCOPE")
	})
}

func decodeJSONBytes(t *testing.T, body string, target any) {
	t.Helper()
	response := &http.Response{Body: io.NopCloser(bytes.NewBufferString(body))}
	decodeResponse(t, response, target)
}

func createLimitedOrgContext(t *testing.T, userID string) (organizationID, storeID string) {
	t.Helper()
	ctx := context.Background()
	if err := testPool.QueryRow(ctx, `
		INSERT INTO organizations (name, slug, created_by_user_id)
		VALUES ('Task 12 Limited Org', 'limited-task12', $1)
		RETURNING id
	`, userID).Scan(&organizationID); err != nil {
		t.Fatal(err)
	}
	if err := testPool.QueryRow(ctx, `
		INSERT INTO stores (organization_id, code, name, timezone, created_by_user_id)
		VALUES ($1, 'LIMITED', 'Limited Store', 'America/Sao_Paulo', $2)
		RETURNING id
	`, organizationID, userID).Scan(&storeID); err != nil {
		t.Fatal(err)
	}
	var membershipID, roleID string
	if err := testPool.QueryRow(ctx, `
		INSERT INTO organization_memberships (organization_id, user_id, default_store_id, created_by_user_id)
		VALUES ($1, $2, $3, $2)
		RETURNING id
	`, organizationID, userID, storeID).Scan(&membershipID); err != nil {
		t.Fatal(err)
	}
	if err := testPool.QueryRow(ctx, `
		INSERT INTO roles (organization_id, key, name, assignment_scope, created_by_membership_id)
		VALUES ($1, 'task12_cashier', 'Task 12 Cashier', 'STORE', $2)
		RETURNING id
	`, organizationID, membershipID).Scan(&roleID); err != nil {
		t.Fatal(err)
	}
	if _, err := testPool.Exec(ctx, `INSERT INTO role_scopes (organization_id, role_id, scope_code) VALUES ($1, $2, 'catalog.read')`, organizationID, roleID); err != nil {
		t.Fatal(err)
	}
	if _, err := testPool.Exec(ctx, `
		INSERT INTO membership_role_bindings (organization_id, membership_id, role_id, store_id, created_by_membership_id)
		VALUES ($1, $2, $3, $4, $2)
	`, organizationID, membershipID, roleID, storeID); err != nil {
		t.Fatal(err)
	}
	return organizationID, storeID
}

func assertErrorResponse(t *testing.T, response *http.Response, status int, code string) {
	t.Helper()
	if response.StatusCode != status {
		t.Fatalf("response status=%d, want %d: %s", response.StatusCode, status, readBody(response))
	}
	if actual := responseErrorCode(t, response); actual != code {
		t.Fatalf("response error code=%q, want %q", actual, code)
	}
}
