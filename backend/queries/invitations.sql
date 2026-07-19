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

-- name: ExpirePendingInvitationsForEmail :many
UPDATE organization_invitations
SET status='EXPIRED'
WHERE organization_id=sqlc.arg(organization_id)
  AND email_normalized=sqlc.arg(email_normalized)
  AND status='PENDING' AND expires_at <= NOW()
RETURNING id;

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

-- name: GetInvitationForOrganization :one
SELECT id, organization_id, email, email_normalized, status, secret_hash, expires_at,
       accepted_at, revoked_at, invited_by_membership_id, accepted_by_membership_id,
       created_at, updated_at
FROM organization_invitations
WHERE organization_id=sqlc.arg(organization_id) AND id=sqlc.arg(invitation_id);

-- name: ListInvitations :many
SELECT id, organization_id, email, email_normalized,
       CASE WHEN status='PENDING' AND expires_at <= NOW() THEN 'EXPIRED'::invitation_status ELSE status END AS status,
       secret_hash, expires_at, accepted_at, revoked_at, invited_by_membership_id,
       accepted_by_membership_id, created_at, updated_at
FROM organization_invitations
WHERE organization_id=sqlc.arg(organization_id)
  AND (CAST(sqlc.narg(status) AS invitation_status) IS NULL
       OR (CASE WHEN status='PENDING' AND expires_at <= NOW() THEN 'EXPIRED'::invitation_status ELSE status END)=CAST(sqlc.narg(status) AS invitation_status))
  AND (CAST(sqlc.narg(email) AS TEXT) IS NULL OR email_normalized ILIKE '%'||CAST(sqlc.narg(email) AS TEXT)||'%')
  AND (CAST(sqlc.narg(created_from) AS TIMESTAMPTZ) IS NULL OR created_at >= CAST(sqlc.narg(created_from) AS TIMESTAMPTZ))
  AND (CAST(sqlc.narg(created_to) AS TIMESTAMPTZ) IS NULL OR created_at < CAST(sqlc.narg(created_to) AS TIMESTAMPTZ))
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(page_size) OFFSET sqlc.arg(page_offset);

-- name: CountInvitations :one
SELECT COUNT(*) FROM organization_invitations
WHERE organization_id=sqlc.arg(organization_id)
  AND (CAST(sqlc.narg(status) AS invitation_status) IS NULL
       OR (CASE WHEN status='PENDING' AND expires_at <= NOW() THEN 'EXPIRED'::invitation_status ELSE status END)=CAST(sqlc.narg(status) AS invitation_status))
  AND (CAST(sqlc.narg(email) AS TEXT) IS NULL OR email_normalized ILIKE '%'||CAST(sqlc.narg(email) AS TEXT)||'%')
  AND (CAST(sqlc.narg(created_from) AS TIMESTAMPTZ) IS NULL OR created_at >= CAST(sqlc.narg(created_from) AS TIMESTAMPTZ))
  AND (CAST(sqlc.narg(created_to) AS TIMESTAMPTZ) IS NULL OR created_at < CAST(sqlc.narg(created_to) AS TIMESTAMPTZ));

-- name: RotateInvitationSecret :one
UPDATE organization_invitations
SET secret_hash=sqlc.arg(secret_hash), expires_at=sqlc.arg(expires_at), status='PENDING',
    accepted_at=NULL, revoked_at=NULL
WHERE organization_id=sqlc.arg(organization_id) AND id=sqlc.arg(invitation_id)
  AND (status='PENDING' OR status='EXPIRED')
RETURNING id, organization_id, email, email_normalized, status, secret_hash, expires_at,
          accepted_at, revoked_at, invited_by_membership_id, accepted_by_membership_id,
          created_at, updated_at;

-- name: ExpireInvitation :one
UPDATE organization_invitations SET status='EXPIRED'
WHERE id=sqlc.arg(invitation_id) AND status='PENDING' AND expires_at <= NOW()
RETURNING id;

-- name: CreateMembershipBindingsFromInvitation :many
INSERT INTO membership_role_bindings
  (organization_id, membership_id, role_id, store_id, created_by_membership_id)
SELECT b.organization_id, sqlc.arg(membership_id), b.role_id, b.store_id,
       i.invited_by_membership_id
FROM invitation_role_bindings b
JOIN organization_invitations i ON i.organization_id=b.organization_id AND i.id=b.invitation_id
WHERE b.organization_id=sqlc.arg(organization_id) AND b.invitation_id=sqlc.arg(invitation_id)
ON CONFLICT DO NOTHING
RETURNING id, organization_id, membership_id, role_id, store_id,
          created_by_membership_id, expires_at, created_at;

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
