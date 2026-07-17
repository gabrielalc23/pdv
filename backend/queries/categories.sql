-- name: CreateCategory :one
INSERT INTO categories (name, slug, is_active)
VALUES (sqlc.arg(name), sqlc.arg(slug), TRUE)
RETURNING id, name, slug, is_active, created_at, updated_at;

-- name: GetCategoryByID :one
SELECT id, name, slug, is_active, created_at, updated_at
FROM categories
WHERE id = sqlc.arg(id)
LIMIT 1;

-- name: GetCategoryByName :one
SELECT id, name, slug, is_active, created_at, updated_at
FROM categories
WHERE name = sqlc.arg(name)
LIMIT 1;

-- name: GetCategoryBySlug :one
SELECT id, name, slug, is_active, created_at, updated_at
FROM categories
WHERE slug = sqlc.arg(slug)
LIMIT 1;

-- name: ListCategories :many
SELECT id, name, slug, is_active, created_at, updated_at
FROM categories
WHERE
    (
        CAST(sqlc.narg(search) AS TEXT) IS NULL
        OR name ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
    )
    AND (NOT CAST(sqlc.arg(active_only) AS BOOLEAN) OR is_active = TRUE)
ORDER BY name ASC, id ASC;

-- name: UpdateCategory :one
UPDATE categories
SET
    name = sqlc.arg(name),
    slug = sqlc.arg(slug),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING id, name, slug, is_active, created_at, updated_at;

-- name: ActivateCategory :one
UPDATE categories
SET is_active = TRUE, updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING id, name, slug, is_active, created_at, updated_at;

-- name: DeactivateCategory :one
UPDATE categories
SET is_active = FALSE, updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING id, name, slug, is_active, created_at, updated_at;
