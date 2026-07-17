-- name: CreateCategoryForOrganization :one
INSERT INTO categories (
    organization_id,
    name,
    slug,
    is_active
)
VALUES (
    sqlc.arg(organization_id),
    sqlc.arg(name),
    sqlc.arg(slug),
    TRUE
)
RETURNING
    organization_id,
    id,
    name,
    slug,
    is_active,
    created_at,
    updated_at;

-- name: GetCategoryByIDForOrganization :one
SELECT
    organization_id,
    id,
    name,
    slug,
    is_active,
    created_at,
    updated_at
FROM categories
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
LIMIT 1;

-- name: GetCategoryByNameForOrganization :one
SELECT
    organization_id,
    id,
    name,
    slug,
    is_active,
    created_at,
    updated_at
FROM categories
WHERE organization_id = sqlc.arg(organization_id)
  AND name = sqlc.arg(name)
LIMIT 1;

-- name: GetCategoryBySlugForOrganization :one
SELECT
    organization_id,
    id,
    name,
    slug,
    is_active,
    created_at,
    updated_at
FROM categories
WHERE organization_id = sqlc.arg(organization_id)
  AND slug = sqlc.arg(slug)
LIMIT 1;

-- name: ListCategoriesForOrganization :many
SELECT
    organization_id,
    id,
    name,
    slug,
    is_active,
    created_at,
    updated_at
FROM categories
WHERE organization_id = sqlc.arg(organization_id)
  AND (
      CAST(sqlc.narg(search) AS TEXT) IS NULL
      OR name ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
  )
  AND (NOT CAST(sqlc.arg(active_only) AS BOOLEAN) OR is_active = TRUE)
ORDER BY name ASC, id ASC;

-- name: UpdateCategoryForOrganization :one
UPDATE categories
SET
    name = sqlc.arg(name),
    slug = sqlc.arg(slug),
    updated_at = NOW()
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
RETURNING
    organization_id,
    id,
    name,
    slug,
    is_active,
    created_at,
    updated_at;

-- name: ActivateCategoryForOrganization :one
UPDATE categories
SET
    is_active = TRUE,
    updated_at = NOW()
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
RETURNING
    organization_id,
    id,
    name,
    slug,
    is_active,
    created_at,
    updated_at;

-- name: DeactivateCategoryForOrganization :one
UPDATE categories
SET
    is_active = FALSE,
    updated_at = NOW()
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
RETURNING
    organization_id,
    id,
    name,
    slug,
    is_active,
    created_at,
    updated_at;
