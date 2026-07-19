-- name: CreateMembership :one
INSERT INTO organization_memberships (
    organization_id,
    user_id,
    default_store_id,
    created_by_user_id
)
VALUES (
    sqlc.arg(organization_id),
    sqlc.arg(user_id),
    sqlc.narg(default_store_id),
    sqlc.arg(created_by_user_id)
)
RETURNING
    id,
    organization_id,
    user_id,
    status,
    default_store_id,
    authorization_version,
    joined_at,
    suspended_at,
    removed_at,
    created_by_user_id,
    created_at,
    updated_at;

-- name: GetActiveMembership :one
SELECT
    m.id,
    m.organization_id,
    m.user_id,
    m.status,
    m.default_store_id,
    m.authorization_version,
    m.joined_at,
    m.suspended_at,
    m.removed_at,
    m.created_by_user_id,
    m.created_at,
    m.updated_at
FROM organization_memberships AS m
JOIN organizations AS o ON o.id = m.organization_id
JOIN users AS u ON u.id = m.user_id
WHERE m.organization_id = sqlc.arg(organization_id)
  AND m.user_id = sqlc.arg(user_id)
  AND m.status = 'ACTIVE'
  AND o.status = 'ACTIVE'
  AND u.status = 'ACTIVE'
LIMIT 1;

-- name: GetMembershipContextForUser :one
SELECT
    m.id,
    m.organization_id,
    m.user_id,
    m.status,
    m.default_store_id,
    m.authorization_version,
    m.joined_at,
    m.suspended_at,
    m.removed_at,
    m.created_by_user_id,
    m.created_at,
    m.updated_at,
    o.name AS organization_name,
    o.slug AS organization_slug,
    o.status AS organization_status,
    o.authorization_version AS organization_authorization_version
FROM organization_memberships AS m
JOIN organizations AS o ON o.id = m.organization_id
WHERE m.organization_id = sqlc.arg(organization_id)
  AND m.user_id = sqlc.arg(user_id)
  AND m.status <> 'REMOVED'
ORDER BY m.joined_at ASC, m.id ASC
LIMIT 1;

-- name: GetMembershipForUpdate :one
SELECT
    id,
    organization_id,
    user_id,
    status,
    default_store_id,
    authorization_version,
    joined_at,
    suspended_at,
    removed_at,
    created_by_user_id,
    created_at,
    updated_at
FROM organization_memberships
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(membership_id)
FOR UPDATE;

-- name: GetMembershipForOrganization :one
SELECT m.id, m.organization_id, m.user_id, m.status, m.default_store_id,
       m.authorization_version, m.joined_at, m.suspended_at, m.removed_at,
       m.created_by_user_id, m.created_at, m.updated_at,
       u.email, u.display_name, u.status AS user_status, u.email_verified_at,
       s.code AS default_store_code, s.name AS default_store_name, s.status AS default_store_status
FROM organization_memberships AS m
JOIN users AS u ON u.id = m.user_id
LEFT JOIN stores AS s ON s.organization_id = m.organization_id AND s.id = m.default_store_id
WHERE m.organization_id = sqlc.arg(organization_id) AND m.id = sqlc.arg(membership_id);

-- name: GetLatestMembershipForUserInOrganization :one
SELECT id, organization_id, user_id, status, default_store_id, authorization_version,
       joined_at, suspended_at, removed_at, created_by_user_id, created_at, updated_at
FROM organization_memberships
WHERE organization_id = sqlc.arg(organization_id) AND user_id = sqlc.arg(user_id)
ORDER BY created_at DESC, id DESC
LIMIT 1;

-- name: IncrementMembershipAuthorizationVersion :one
UPDATE organization_memberships
SET authorization_version = authorization_version + 1
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(membership_id)
RETURNING
    id,
    organization_id,
    authorization_version,
    updated_at;

-- name: ListMemberships :many
SELECT
    m.id,
    m.organization_id,
    m.user_id,
    m.status,
    m.default_store_id,
    m.authorization_version,
    m.joined_at,
    m.suspended_at,
    m.removed_at,
    m.created_by_user_id,
    m.created_at,
    m.updated_at,
    u.email,
    u.email_normalized,
    u.display_name,
    u.status AS user_status,
    s.code AS default_store_code,
    s.name AS default_store_name
FROM organization_memberships AS m
JOIN users AS u ON u.id = m.user_id
LEFT JOIN stores AS s
  ON s.organization_id = m.organization_id
 AND s.id = m.default_store_id
WHERE m.organization_id = sqlc.arg(organization_id)
  AND (
      CAST(sqlc.narg(status) AS membership_status) IS NULL
      OR m.status = CAST(sqlc.narg(status) AS membership_status)
  )
  AND (
      CAST(sqlc.narg(search) AS TEXT) IS NULL
      OR u.display_name ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
      OR u.email_normalized ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
  )
ORDER BY m.created_at DESC, m.id DESC
LIMIT sqlc.arg(page_size)
OFFSET sqlc.arg(page_offset);

-- name: CountMemberships :one
SELECT COUNT(*)
FROM organization_memberships AS m
JOIN users AS u ON u.id = m.user_id
WHERE m.organization_id = sqlc.arg(organization_id)
  AND (
      CAST(sqlc.narg(status) AS membership_status) IS NULL
      OR m.status = CAST(sqlc.narg(status) AS membership_status)
  )
  AND (
      CAST(sqlc.narg(search) AS TEXT) IS NULL
      OR u.display_name ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
      OR u.email_normalized ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
  );

-- name: UpdateMembershipStatus :one
UPDATE organization_memberships
SET
    status = sqlc.arg(status),
    suspended_at = CASE
        WHEN sqlc.arg(status)::membership_status = 'SUSPENDED' THEN NOW()
        ELSE NULL
    END,
    removed_at = CASE
        WHEN sqlc.arg(status)::membership_status = 'REMOVED' THEN NOW()
        ELSE NULL
    END,
    authorization_version = authorization_version + 1
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(membership_id)
RETURNING
    id,
    organization_id,
    user_id,
    status,
    default_store_id,
    authorization_version,
    joined_at,
    suspended_at,
    removed_at,
    created_by_user_id,
    created_at,
    updated_at;
