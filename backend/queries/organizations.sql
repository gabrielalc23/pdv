-- name: CreateOrganization :one
INSERT INTO
    organizations (
        name,
        slug,
        timezone,
        locale,
        currency,
        created_by_user_id
    )
VALUES (
        sqlc.arg (name),
        sqlc.arg (slug),
        sqlc.arg (timezone),
        sqlc.arg (locale),
        sqlc.arg (currency),
        sqlc.arg (created_by_user_id)
    ) RETURNING id,
    name,
    slug,
    status,
    timezone,
    locale,
    currency,
    authorization_version,
    created_by_user_id,
    archived_at,
    created_at,
    updated_at;

-- name: IncrementOrganizationAuthorizationVersion :one
UPDATE organizations
SET
    authorization_version = authorization_version + 1
WHERE
    id = sqlc.arg (organization_id) RETURNING id AS organization_id,
    authorization_version,
    updated_at;

-- name: GetOrganizationAuthorizationVersion :one
SELECT
    id AS organization_id,
    status,
    authorization_version
FROM organizations
WHERE
    id = sqlc.arg (organization_id);

-- name: LockOrganizationForOwnerChange :one
SELECT
    id AS organization_id,
    status,
    authorization_version
FROM organizations
WHERE
    id = sqlc.arg (organization_id) FOR
UPDATE;