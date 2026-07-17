-- name: CreateInvitation :one
INSERT INTO organization_invitations (
    organization_id,
    email,
    email_normalized,
    secret_hash,
    expires_at,
    invited_by_membership_id
)
VALUES (
    sqlc.arg(organization_id),
    sqlc.arg(email),
    sqlc.arg(email_normalized),
    sqlc.arg(secret_hash),
    sqlc.arg(expires_at),
    sqlc.arg(invited_by_membership_id)
)
RETURNING
    id,
    organization_id,
    email,
    email_normalized,
    status,
    secret_hash,
    expires_at,
    accepted_at,
    revoked_at,
    invited_by_membership_id,
    accepted_by_membership_id,
    created_at,
    updated_at;

-- name: GetInvitationForUpdate :one
SELECT
    id,
    organization_id,
    email,
    email_normalized,
    status,
    secret_hash,
    expires_at,
    accepted_at,
    revoked_at,
    invited_by_membership_id,
    accepted_by_membership_id,
    created_at,
    updated_at
FROM organization_invitations
WHERE id = sqlc.arg(id)
FOR UPDATE;

-- name: CreateInvitationRoleBindings :many
INSERT INTO invitation_role_bindings (
    organization_id,
    invitation_id,
    role_id,
    store_id
)
SELECT
    sqlc.arg(organization_id),
    sqlc.arg(invitation_id),
    requested.role_id,
    requested.store_id
FROM jsonb_to_recordset(CAST(sqlc.arg(assignments) AS JSONB))
    AS requested(role_id UUID, store_id UUID)
ON CONFLICT DO NOTHING
RETURNING
    id,
    organization_id,
    invitation_id,
    role_id,
    store_id,
    created_at;

-- name: ListInvitationRoleBindings :many
SELECT
    b.id,
    b.organization_id,
    b.invitation_id,
    b.role_id,
    b.store_id,
    b.created_at,
    r.key AS role_key,
    r.assignment_scope,
    s.code AS store_code,
    s.name AS store_name
FROM invitation_role_bindings AS b
JOIN roles AS r
  ON r.organization_id = b.organization_id
 AND r.id = b.role_id
LEFT JOIN stores AS s
  ON s.organization_id = b.organization_id
 AND s.id = b.store_id
WHERE b.organization_id = sqlc.arg(organization_id)
  AND b.invitation_id = sqlc.arg(invitation_id)
ORDER BY r.assignment_scope ASC, r.key ASC, s.name ASC NULLS FIRST, b.id ASC;

-- name: AcceptInvitation :one
UPDATE organization_invitations
SET
    status = 'ACCEPTED',
    accepted_at = NOW(),
    accepted_by_membership_id = sqlc.arg(accepted_by_membership_id)
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(invitation_id)
  AND status = 'PENDING'
  AND expires_at > NOW()
RETURNING
    id,
    organization_id,
    status,
    accepted_at,
    accepted_by_membership_id,
    updated_at;

-- name: RevokeInvitation :one
UPDATE organization_invitations
SET
    status = 'REVOKED',
    revoked_at = NOW()
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(invitation_id)
  AND status = 'PENDING'
RETURNING
    id,
    organization_id,
    status,
    revoked_at,
    updated_at;
