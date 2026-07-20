-- name: CreateAuthSession :one
INSERT INTO auth_sessions (
    user_id,
    client_id,
    device_name,
    user_agent,
    ip_address,
    context_kind,
    current_organization_id,
    current_membership_id,
    current_store_id,
    idle_expires_at,
    absolute_expires_at
)
VALUES (
    sqlc.arg(user_id),
    sqlc.arg(client_id),
    sqlc.narg(device_name),
    sqlc.narg(user_agent),
    sqlc.narg(ip_address),
    sqlc.arg(context_kind),
    sqlc.narg(current_organization_id),
    sqlc.narg(current_membership_id),
    sqlc.narg(current_store_id),
    sqlc.arg(idle_expires_at),
    sqlc.arg(absolute_expires_at)
)
RETURNING
    id,
    user_id,
    status,
    client_id,
    device_name,
    user_agent,
    ip_address,
    context_kind,
    current_organization_id,
    current_membership_id,
    current_store_id,
    idle_expires_at,
    absolute_expires_at,
    last_seen_at,
    revoked_at,
    revoke_reason,
    created_at,
    updated_at;

-- name: GetAuthSessionForUpdate :one
SELECT
    s.id,
    s.user_id,
    s.status,
    s.client_id,
    s.device_name,
    s.user_agent,
    s.ip_address,
    s.context_kind,
    s.current_organization_id,
    s.current_membership_id,
    s.current_store_id,
    s.idle_expires_at,
    s.absolute_expires_at,
    s.last_seen_at,
    s.revoked_at,
    s.revoke_reason,
    s.created_at,
    s.updated_at,
    u.status AS user_status,
    u.password_version,
    o.status AS organization_status,
    o.authorization_version AS organization_authorization_version,
    m.status AS membership_status,
    m.authorization_version AS membership_authorization_version,
    st.status AS store_status
FROM auth_sessions AS s
JOIN users AS u ON u.id = s.user_id
LEFT JOIN organizations AS o ON o.id = s.current_organization_id
LEFT JOIN organization_memberships AS m
  ON m.organization_id = s.current_organization_id
 AND m.id = s.current_membership_id
 AND m.user_id = s.user_id
LEFT JOIN stores AS st
  ON st.organization_id = s.current_organization_id
 AND st.id = s.current_store_id
WHERE s.id = sqlc.arg(session_id)
FOR UPDATE OF s;

-- name: GetAuthSessionState :one
SELECT
    s.id,
    s.user_id,
    s.status,
    s.client_id,
    s.context_kind,
    s.current_organization_id,
    s.current_membership_id,
    s.current_store_id,
    s.idle_expires_at,
    s.absolute_expires_at,
    s.last_seen_at,
    u.status AS user_status,
    u.password_version,
    o.status AS organization_status,
    o.authorization_version AS organization_authorization_version,
    m.status AS membership_status,
    m.authorization_version AS membership_authorization_version,
    st.status AS store_status
FROM auth_sessions AS s
JOIN users AS u ON u.id = s.user_id
LEFT JOIN organizations AS o ON o.id = s.current_organization_id
LEFT JOIN organization_memberships AS m
  ON m.organization_id = s.current_organization_id
 AND m.id = s.current_membership_id
 AND m.user_id = s.user_id
LEFT JOIN stores AS st
  ON st.organization_id = s.current_organization_id
 AND st.id = s.current_store_id
WHERE s.id = sqlc.arg(session_id);

-- name: GetAuthSessionView :one
SELECT
    s.id,
    s.user_id,
    s.status,
    s.client_id,
    s.device_name,
    s.context_kind,
    s.current_organization_id,
    s.current_membership_id,
    s.current_store_id,
    s.idle_expires_at,
    s.absolute_expires_at,
    s.last_seen_at,
    s.created_at,
    u.email,
    u.display_name,
    u.status AS user_status,
    u.email_verified_at,
    u.password_version,
    o.name AS organization_name,
    o.slug AS organization_slug,
    o.status AS organization_status,
    o.authorization_version AS organization_authorization_version,
    m.status AS membership_status,
    m.authorization_version AS membership_authorization_version,
    st.code AS store_code,
    st.name AS store_name,
    st.status AS store_status
FROM auth_sessions AS s
JOIN users AS u ON u.id = s.user_id
LEFT JOIN organizations AS o ON o.id = s.current_organization_id
LEFT JOIN organization_memberships AS m
  ON m.organization_id = s.current_organization_id
 AND m.id = s.current_membership_id
 AND m.user_id = s.user_id
LEFT JOIN stores AS st
  ON st.organization_id = s.current_organization_id
 AND st.id = s.current_store_id
WHERE s.id = sqlc.arg(session_id);

-- name: GetLastActiveSessionContextForClient :one
SELECT
    id,
    context_kind,
    current_organization_id,
    current_membership_id,
    current_store_id
FROM auth_sessions
WHERE user_id = sqlc.arg(user_id)
  AND client_id = sqlc.arg(client_id)
  AND status = 'ACTIVE'
  AND idle_expires_at > NOW()
  AND absolute_expires_at > NOW()
ORDER BY last_seen_at DESC, id DESC
LIMIT 1;

-- name: UpdateSessionContext :one
UPDATE auth_sessions
SET
    context_kind = sqlc.arg(context_kind),
    current_organization_id = sqlc.narg(current_organization_id),
    current_membership_id = sqlc.narg(current_membership_id),
    current_store_id = sqlc.narg(current_store_id),
    last_seen_at = NOW()
WHERE id = sqlc.arg(session_id)
  AND user_id = sqlc.arg(user_id)
  AND status = 'ACTIVE'
  AND idle_expires_at > NOW()
  AND absolute_expires_at > NOW()
RETURNING
    id,
    user_id,
    status,
    context_kind,
    current_organization_id,
    current_membership_id,
    current_store_id,
    idle_expires_at,
    absolute_expires_at,
    updated_at;

-- name: TouchSession :one
UPDATE auth_sessions
SET
    last_seen_at = NOW(),
    idle_expires_at = LEAST(sqlc.arg(idle_expires_at), absolute_expires_at)
WHERE id = sqlc.arg(session_id)
  AND user_id = sqlc.arg(user_id)
  AND status = 'ACTIVE'
  AND idle_expires_at > NOW()
  AND absolute_expires_at > NOW()
RETURNING
    id,
    user_id,
    idle_expires_at,
    absolute_expires_at,
    last_seen_at,
    updated_at;

-- name: RevokeSession :one
UPDATE auth_sessions
SET
    status = 'REVOKED',
    revoked_at = NOW(),
    revoke_reason = sqlc.arg(revoke_reason)
WHERE id = sqlc.arg(session_id)
  AND user_id = sqlc.arg(user_id)
  AND status = 'ACTIVE'
RETURNING
    id,
    user_id,
    status,
    current_organization_id,
    current_membership_id,
    current_store_id,
    revoked_at,
    revoke_reason,
    updated_at;

-- name: MarkSessionCompromised :one
UPDATE auth_sessions
SET
    status = 'COMPROMISED',
    revoked_at = NOW(),
    revoke_reason = sqlc.arg(revoke_reason)
WHERE id = sqlc.arg(session_id)
  AND status = 'ACTIVE'
RETURNING
    id,
    user_id,
    status,
    current_organization_id,
    current_membership_id,
    current_store_id,
    revoked_at,
    revoke_reason,
    updated_at;

-- name: RevokeAllUserSessions :many
UPDATE auth_sessions
SET
    status = 'REVOKED',
    revoked_at = NOW(),
    revoke_reason = sqlc.arg(revoke_reason)
WHERE user_id = sqlc.arg(user_id)
  AND status = 'ACTIVE'
RETURNING
    id,
    user_id,
    current_organization_id,
    current_membership_id,
    current_store_id,
    status,
    revoked_at,
    revoke_reason;

-- name: RevokeAllActiveUserSessions :many
UPDATE auth_sessions
SET
    status = 'REVOKED',
    revoked_at = NOW(),
    revoke_reason = sqlc.arg(revoke_reason)
WHERE user_id = sqlc.arg(user_id)
  AND status = 'ACTIVE'
RETURNING id;

-- name: ExpireSessions :many
UPDATE auth_sessions
SET
    status = 'EXPIRED',
    revoked_at = NOW(),
    revoke_reason = 'session_expired'
WHERE status = 'ACTIVE'
  AND (
      idle_expires_at <= NOW()
      OR absolute_expires_at <= NOW()
  )
RETURNING
    id,
    user_id,
    current_organization_id,
    current_membership_id,
    current_store_id,
    status,
    revoked_at,
    revoke_reason;

-- name: CreateRefreshToken :one
INSERT INTO auth_refresh_tokens (
    id,
    session_id,
    parent_token_id,
    secret_hash,
    expires_at
)
VALUES (
    sqlc.arg(id),
    sqlc.arg(session_id),
    sqlc.narg(parent_token_id),
    sqlc.arg(secret_hash),
    sqlc.arg(expires_at)
)
RETURNING
    id,
    session_id,
    parent_token_id,
    replaced_by_token_id,
    secret_hash,
    expires_at,
    consumed_at,
    revoked_at,
    created_at;

-- name: GetRefreshTokenForUpdate :one
SELECT
    id,
    session_id,
    parent_token_id,
    replaced_by_token_id,
    secret_hash,
    expires_at,
    consumed_at,
    revoked_at,
    created_at
FROM auth_refresh_tokens
WHERE id = sqlc.arg(id)
FOR UPDATE;

-- name: ConsumeAndReplaceRefreshToken :one
UPDATE auth_refresh_tokens
SET
    consumed_at = NOW(),
    replaced_by_token_id = sqlc.arg(replaced_by_token_id)
WHERE session_id = sqlc.arg(session_id)
  AND id = sqlc.arg(id)
  AND consumed_at IS NULL
  AND revoked_at IS NULL
  AND expires_at > NOW()
RETURNING
    id,
    session_id,
    parent_token_id,
    replaced_by_token_id,
    expires_at,
    consumed_at,
    revoked_at,
    created_at;

-- name: RevokeSessionRefreshTokens :many
UPDATE auth_refresh_tokens
SET revoked_at = NOW()
WHERE session_id = sqlc.arg(session_id)
  AND revoked_at IS NULL
RETURNING
    id,
    session_id,
    parent_token_id,
    replaced_by_token_id,
    expires_at,
    consumed_at,
    revoked_at,
    created_at;

-- name: RevokeAllUserRefreshTokens :execrows
UPDATE auth_refresh_tokens AS t
SET revoked_at = NOW()
WHERE t.revoked_at IS NULL
  AND EXISTS (
      SELECT 1
      FROM auth_sessions AS s
      WHERE s.id = t.session_id
        AND s.user_id = sqlc.arg(user_id)
  );

-- name: ListAllUserSessionIDs :many
SELECT id
FROM auth_sessions
WHERE user_id = sqlc.arg(user_id)
ORDER BY id;

-- name: ListSessionIDsForMembership :many
SELECT id FROM auth_sessions
WHERE current_organization_id=sqlc.arg(organization_id)
  AND current_membership_id=sqlc.arg(membership_id)
ORDER BY id;

-- name: ListSessionIDsForOrganization :many
SELECT id FROM auth_sessions
WHERE current_organization_id=sqlc.arg(organization_id)
ORDER BY id;

-- name: ListSessionIDsForStore :many
SELECT id FROM auth_sessions
WHERE current_organization_id=sqlc.arg(organization_id)
  AND current_store_id=sqlc.arg(store_id)
ORDER BY id;

-- name: RevokeOrganizationSessions :many
UPDATE auth_sessions
SET status='REVOKED', revoked_at=NOW(), revoke_reason=sqlc.arg(revoke_reason)
WHERE current_organization_id=sqlc.arg(organization_id) AND status='ACTIVE'
RETURNING id;

-- name: ListUserSessions :many
SELECT
    id,
    user_id,
    status,
    client_id,
    device_name,
    user_agent,
    ip_address,
    context_kind,
    current_organization_id,
    current_membership_id,
    current_store_id,
    idle_expires_at,
    absolute_expires_at,
    last_seen_at,
    revoked_at,
    revoke_reason,
    created_at,
    updated_at
FROM auth_sessions
WHERE user_id = sqlc.arg(user_id)
ORDER BY last_seen_at DESC;

-- name: GetAuthSessionByID :one
SELECT
    id,
    user_id,
    status,
    client_id,
    device_name,
    user_agent,
    ip_address,
    context_kind,
    current_organization_id,
    current_membership_id,
    current_store_id,
    idle_expires_at,
    absolute_expires_at,
    last_seen_at,
    revoked_at,
    revoke_reason,
    created_at,
    updated_at
FROM auth_sessions
WHERE id = sqlc.arg(id);
