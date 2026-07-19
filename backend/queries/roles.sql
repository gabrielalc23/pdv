-- name: CreateRole :one
INSERT INTO
    roles (
        organization_id,
        key,
        name,
        description,
        assignment_scope,
        is_system,
        is_mutable,
        created_by_membership_id
    )
VALUES (
        sqlc.arg (organization_id),
        sqlc.arg (key),
        sqlc.arg (name),
        sqlc.narg (description),
        sqlc.arg (assignment_scope),
        sqlc.arg (is_system),
        sqlc.arg (is_mutable),
        sqlc.narg (created_by_membership_id)
    ) RETURNING id,
    organization_id,
    key,
    name,
    description,
    assignment_scope,
    is_system,
    is_mutable,
    is_active,
    created_by_membership_id,
    created_at,
     updated_at;

-- name: ListPermissionScopes :many
SELECT
    code,
    resource,
    action,
    scope_level,
    description,
    is_assignable,
    created_at
FROM permission_scopes
ORDER BY code ASC;

-- name: ReplaceRoleScopes :many
WITH deleted_scopes AS (
    DELETE FROM role_scopes
    WHERE organization_id = sqlc.arg(organization_id)
      AND role_id = sqlc.arg(role_id)
      AND scope_code <> ALL(CAST(sqlc.arg(scope_codes) AS TEXT[]))
), requested_scopes AS (
    SELECT DISTINCT UNNEST(CAST(sqlc.arg(scope_codes) AS TEXT[])) AS scope_code
)
INSERT INTO role_scopes (
    organization_id,
    role_id,
    scope_code
)
SELECT
    sqlc.arg(organization_id),
    sqlc.arg(role_id),
    requested_scopes.scope_code
FROM requested_scopes
ORDER BY requested_scopes.scope_code
ON CONFLICT (organization_id, role_id, scope_code) DO NOTHING
RETURNING
    organization_id,
    role_id,
    scope_code,
    created_at;

-- name: LockRoleForScopeChange :one
SELECT
    id,
    organization_id,
    is_mutable
FROM roles
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(role_id)
  AND is_mutable
FOR UPDATE;

-- name: GetRoleForOrganization :one
SELECT
    id,
    organization_id,
    key,
    name,
    description,
    assignment_scope,
    is_system,
    is_mutable,
    is_active,
    created_by_membership_id,
    created_at,
    updated_at
FROM roles
WHERE
    organization_id = sqlc.arg (organization_id)
    AND id = sqlc.arg (role_id);

-- name: ListRolesWithScopes :many
SELECT
    r.id,
    r.organization_id,
    r.key,
    r.name,
    r.description,
    r.assignment_scope,
    r.is_system,
    r.is_mutable,
    r.is_active,
    r.created_by_membership_id,
    r.created_at,
    r.updated_at,
    ARRAY(
        SELECT role_scope.scope_code::TEXT
        FROM role_scopes AS role_scope
        WHERE role_scope.organization_id = r.organization_id
          AND role_scope.role_id = r.id
        ORDER BY role_scope.scope_code
    )::TEXT[] AS scope_codes
FROM roles AS r
WHERE r.organization_id = sqlc.arg(organization_id)
ORDER BY r.name ASC, r.id ASC;

-- name: ResolveEffectiveScopes :one
WITH valid_context AS (
    SELECT
        m.organization_id,
        m.id AS membership_id,
        CAST(sqlc.arg(context_kind) AS auth_context_kind) AS context_kind,
        CAST(sqlc.narg(store_id) AS UUID) AS store_id
    FROM organization_memberships AS m
    JOIN users AS u ON u.id = m.user_id
    JOIN organizations AS o ON o.id = m.organization_id
    LEFT JOIN stores AS s
      ON s.organization_id = m.organization_id
     AND s.id = CAST(sqlc.narg(store_id) AS UUID)
    WHERE m.organization_id = sqlc.arg(organization_id)
      AND m.id = sqlc.arg(membership_id)
      AND m.user_id = sqlc.arg(user_id)
      AND u.status = 'ACTIVE'
      AND o.status = 'ACTIVE'
      AND m.status = 'ACTIVE'
      AND (
          (
              CAST(sqlc.arg(context_kind) AS auth_context_kind) = 'ORGANIZATION'
              AND CAST(sqlc.narg(store_id) AS UUID) IS NULL
          )
          OR (
              CAST(sqlc.arg(context_kind) AS auth_context_kind) = 'STORE'
              AND CAST(sqlc.narg(store_id) AS UUID) IS NOT NULL
              AND s.status = 'ACTIVE'
          )
      )
), effective_roles AS (
    SELECT DISTINCT
        r.organization_id,
        r.id,
        r.key
    FROM valid_context AS c
    JOIN membership_role_bindings AS b
      ON b.organization_id = c.organization_id
     AND b.membership_id = c.membership_id
    JOIN roles AS r
      ON r.organization_id = b.organization_id
     AND r.id = b.role_id
    WHERE r.is_active
      AND (b.expires_at IS NULL OR b.expires_at > NOW())
      AND (
          (
              c.context_kind = 'ORGANIZATION'
              AND r.assignment_scope = 'ORGANIZATION'
              AND b.store_id IS NULL
          )
          OR (
              c.context_kind = 'STORE'
              AND (
                  (
                      r.assignment_scope = 'ORGANIZATION'
                      AND b.store_id IS NULL
                  )
                  OR (
                      r.assignment_scope = 'STORE'
                      AND b.store_id = c.store_id
                  )
              )
          )
      )
)
SELECT
    ARRAY(
        SELECT r.key
        FROM effective_roles AS r
        ORDER BY r.key
    )::TEXT[] AS role_keys,
    ARRAY(
        SELECT DISTINCT rs.scope_code
        FROM effective_roles AS r
        JOIN role_scopes AS rs
          ON rs.organization_id = r.organization_id
         AND rs.role_id = r.id
        JOIN permission_scopes AS ps ON ps.code = rs.scope_code
        WHERE c.context_kind = 'STORE'
           OR ps.scope_level = 'ORGANIZATION'
        ORDER BY rs.scope_code
    )::TEXT[] AS scope_codes
FROM valid_context AS c;

-- name: CreateRoleBinding :one
WITH
    organization_binding AS (
        INSERT INTO
            membership_role_bindings (
                organization_id,
                membership_id,
                role_id,
                store_id,
                created_by_membership_id,
                expires_at
            )
        SELECT sqlc.arg (organization_id), sqlc.arg (membership_id), sqlc.arg (role_id), NULL, sqlc.arg (created_by_membership_id), CAST(
                sqlc.narg (expires_at) AS TIMESTAMPTZ
            )
        WHERE
            CAST(sqlc.narg (store_id) AS UUID) IS NULL ON CONFLICT (
                organization_id,
                membership_id,
                role_id
            )
        WHERE
            store_id IS NULL DO
        UPDATE
        SET
            expires_at = EXCLUDED.expires_at,
            created_by_membership_id = EXCLUDED.created_by_membership_id RETURNING id,
            organization_id,
            membership_id,
            role_id,
            store_id,
            created_by_membership_id,
            expires_at,
            created_at
    ),
    store_binding AS (
        INSERT INTO
            membership_role_bindings (
                organization_id,
                membership_id,
                role_id,
                store_id,
                created_by_membership_id,
                expires_at
            )
        SELECT sqlc.arg (organization_id), sqlc.arg (membership_id), sqlc.arg (role_id), CAST(sqlc.narg (store_id) AS UUID), sqlc.arg (created_by_membership_id), CAST(
                sqlc.narg (expires_at) AS TIMESTAMPTZ
            )
        WHERE
            CAST(sqlc.narg (store_id) AS UUID) IS NOT NULL ON CONFLICT (
                organization_id,
                membership_id,
                role_id,
                store_id
            )
        WHERE
            store_id IS NOT NULL DO
        UPDATE
        SET
            expires_at = EXCLUDED.expires_at,
            created_by_membership_id = EXCLUDED.created_by_membership_id RETURNING id,
            organization_id,
            membership_id,
            role_id,
            store_id,
            created_by_membership_id,
            expires_at,
            created_at
    )
SELECT *
FROM organization_binding
UNION ALL
SELECT *
FROM store_binding;

-- name: DeleteRoleBinding :one
DELETE FROM membership_role_bindings
WHERE
    organization_id = sqlc.arg (organization_id)
    AND id = sqlc.arg (binding_id) RETURNING id,
    organization_id,
    membership_id,
    role_id,
    store_id,
    created_by_membership_id,
    expires_at,
    created_at;

-- name: ListMemberRoleBindings :many
SELECT
    b.id,
    b.organization_id,
    b.membership_id,
    b.role_id,
    b.store_id,
    b.created_by_membership_id,
    b.expires_at,
    b.created_at,
    r.key AS role_key,
    r.name AS role_name,
    r.assignment_scope,
    r.is_system,
    r.is_active,
    s.code AS store_code,
    s.name AS store_name
FROM
    membership_role_bindings AS b
    JOIN roles AS r ON r.organization_id = b.organization_id
    AND r.id = b.role_id
    LEFT JOIN stores AS s ON s.organization_id = b.organization_id
    AND s.id = b.store_id
WHERE
    b.organization_id = sqlc.arg (organization_id)
    AND b.membership_id = sqlc.arg (membership_id)
ORDER BY r.assignment_scope ASC, r.name ASC, s.name ASC NULLS FIRST, b.id ASC;

-- name: CountActiveOwnersForUpdate :one
WITH locked_owners AS (
    SELECT m.id
    FROM organization_memberships AS m
    JOIN membership_role_bindings AS b
      ON b.organization_id = m.organization_id
     AND b.membership_id = m.id
    JOIN roles AS r
      ON r.organization_id = b.organization_id
     AND r.id = b.role_id
    WHERE m.organization_id = sqlc.arg(organization_id)
      AND m.status = 'ACTIVE'
      AND r.key = 'owner'
      AND r.is_system
      AND r.is_active
      AND b.store_id IS NULL
      AND (b.expires_at IS NULL OR b.expires_at > NOW())
    FOR UPDATE OF m, b
)
SELECT COUNT(DISTINCT id)::BIGINT AS owner_count
FROM locked_owners;
