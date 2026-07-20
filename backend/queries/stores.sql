-- name: CreateStore :one
INSERT INTO stores (
    organization_id,
    code,
    name,
    timezone,
    created_by_user_id
)
VALUES (
    sqlc.arg(organization_id),
    sqlc.arg(code),
    sqlc.arg(name),
    sqlc.arg(timezone),
    sqlc.arg(created_by_user_id)
)
RETURNING
    id,
    organization_id,
    code,
    name,
    status,
    timezone,
    created_by_user_id,
    archived_at,
    created_at,
    updated_at;

-- name: GetStoreForOrganization :one
SELECT
    id,
    organization_id,
    code,
    name,
    status,
    timezone,
    created_by_user_id,
    archived_at,
    created_at,
    updated_at
FROM stores
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(store_id);

-- name: ListStoresForMembership :many
SELECT
    s.id,
    s.organization_id,
    s.code,
    s.name,
    s.status,
    s.timezone,
    s.created_by_user_id,
    s.archived_at,
    s.created_at,
    s.updated_at
FROM organization_memberships AS m
JOIN organizations AS o ON o.id = m.organization_id
JOIN stores AS s ON s.organization_id = m.organization_id
WHERE m.organization_id = sqlc.arg(organization_id)
  AND m.id = sqlc.arg(membership_id)
  AND m.status = 'ACTIVE'
  AND o.status = 'ACTIVE'
  AND s.status = 'ACTIVE'
  AND (
      EXISTS (
          SELECT 1
          FROM membership_role_bindings AS b
          JOIN roles AS r
            ON r.organization_id = b.organization_id
           AND r.id = b.role_id
          JOIN role_scopes AS rs
            ON rs.organization_id = r.organization_id
           AND rs.role_id = r.id
          JOIN permission_scopes AS ps ON ps.code = rs.scope_code
          WHERE b.organization_id = m.organization_id
            AND b.membership_id = m.id
            AND b.store_id IS NULL
            AND r.assignment_scope = 'ORGANIZATION'
            AND r.is_active
            AND ps.scope_level = 'STORE'
            AND (b.expires_at IS NULL OR b.expires_at > NOW())
      )
      OR EXISTS (
          SELECT 1
          FROM membership_role_bindings AS b
          JOIN roles AS r
            ON r.organization_id = b.organization_id
           AND r.id = b.role_id
          JOIN role_scopes AS rs
            ON rs.organization_id = r.organization_id
           AND rs.role_id = r.id
          JOIN permission_scopes AS ps ON ps.code = rs.scope_code
          WHERE b.organization_id = m.organization_id
            AND b.membership_id = m.id
            AND b.store_id = s.id
            AND r.assignment_scope = 'STORE'
            AND r.is_active
            AND ps.scope_level = 'STORE'
            AND (b.expires_at IS NULL OR b.expires_at > NOW())
      )
  )
ORDER BY s.name ASC, s.id ASC;

-- name: ListStoresForOrganization :many
SELECT id, organization_id, code, name, status, timezone, created_by_user_id,
       archived_at, created_at, updated_at
FROM stores
WHERE organization_id = sqlc.arg(organization_id)
  AND (CAST(sqlc.narg(status) AS store_status) IS NULL OR status = CAST(sqlc.narg(status) AS store_status))
  AND (CAST(sqlc.narg(search) AS TEXT) IS NULL
       OR name ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
       OR code ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%')
ORDER BY name ASC, id ASC
LIMIT sqlc.arg(page_size) OFFSET sqlc.arg(page_offset);

-- name: CountStoresForOrganization :one
SELECT COUNT(*)
FROM stores
WHERE organization_id = sqlc.arg(organization_id)
  AND (CAST(sqlc.narg(status) AS store_status) IS NULL OR status = CAST(sqlc.narg(status) AS store_status))
  AND (CAST(sqlc.narg(search) AS TEXT) IS NULL
       OR name ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
       OR code ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%');

-- name: LockStoreForStatusChange :one
SELECT id, organization_id, code, name, status, timezone, created_by_user_id,
       archived_at, created_at, updated_at
FROM stores
WHERE organization_id = sqlc.arg(organization_id) AND id = sqlc.arg(store_id)
FOR UPDATE;

-- name: CountActiveStores :one
SELECT COUNT(*) FROM stores
WHERE organization_id = sqlc.arg(organization_id) AND status = 'ACTIVE';

-- name: HasOpenSalesForStore :one
SELECT EXISTS (
  SELECT 1 FROM sales
  WHERE organization_id = sqlc.arg(organization_id)
    AND store_id = sqlc.arg(store_id)
    AND status = 'OPEN'
);

-- name: UpdateStore :one
UPDATE stores
SET
    code = sqlc.arg(code),
    name = sqlc.arg(name),
    timezone = sqlc.arg(timezone)
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(store_id)
RETURNING
    id,
    organization_id,
    code,
    name,
    status,
    timezone,
    created_by_user_id,
    archived_at,
    created_at,
    updated_at;

-- name: UpdateStoreStatus :one
UPDATE stores
SET
    status = sqlc.arg(status),
    archived_at = CASE
        WHEN sqlc.arg(status)::store_status = 'ARCHIVED' THEN NOW()
        ELSE NULL
    END
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(store_id)
RETURNING
    id,
    organization_id,
    code,
    name,
    status,
    timezone,
    created_by_user_id,
    archived_at,
    created_at,
    updated_at;
