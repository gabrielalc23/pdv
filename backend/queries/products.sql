-- name: CreateProduct :one
INSERT INTO products (
    sku,
    barcode,
    name,
    category_id,
    price,
    cost,
    is_active
)
VALUES (
    sqlc.arg(sku),
    sqlc.narg(barcode),
    sqlc.arg(name),
    sqlc.narg(category_id),
    sqlc.arg(price),
    sqlc.narg(cost),
    sqlc.arg(is_active)
)
RETURNING
    id,
    sku,
    barcode,
    name,
    category_id,
    price,
    cost,
    is_active,
    created_at,
    updated_at;

-- name: GetProductByID :one
SELECT
    id,
    sku,
    barcode,
    name,
    category_id,
    price,
    cost,
    is_active,
    created_at,
    updated_at
FROM products
WHERE id = sqlc.arg(id)
LIMIT 1;

-- name: GetProductBySKU :one
SELECT
    id,
    sku,
    barcode,
    name,
    category_id,
    price,
    cost,
    is_active,
    created_at,
    updated_at
FROM products
WHERE sku = sqlc.arg(sku)
LIMIT 1;

-- name: GetProductByBarcode :one
SELECT
    id,
    sku,
    barcode,
    name,
    category_id,
    price,
    cost,
    is_active,
    created_at,
    updated_at
FROM products
WHERE barcode = sqlc.arg(barcode)
LIMIT 1;

-- name: ListProducts :many
SELECT
    id,
    sku,
    barcode,
    name,
    category_id,
    price,
    cost,
    is_active,
    created_at,
    updated_at
FROM products
WHERE
    (
        CAST(sqlc.narg(search) AS TEXT) IS NULL
        OR name ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
        OR sku ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
        OR barcode ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
    )
    AND (
        CAST(sqlc.narg(category_id) AS UUID) IS NULL
        OR category_id = CAST(sqlc.narg(category_id) AS UUID)
    )
    AND (NOT sqlc.arg(active_only) OR is_active = TRUE)
ORDER BY name ASC, id ASC
LIMIT sqlc.arg(page_size)
OFFSET sqlc.arg(page_offset);

-- name: CountProducts :one
SELECT COUNT(*)
FROM products
WHERE
    (
        CAST(sqlc.narg(search) AS TEXT) IS NULL
        OR name ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
        OR sku ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
        OR barcode ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
    )
    AND (
        CAST(sqlc.narg(category_id) AS UUID) IS NULL
        OR category_id = CAST(sqlc.narg(category_id) AS UUID)
    )
    AND (NOT sqlc.arg(active_only) OR is_active = TRUE);

-- name: UpdateProduct :one
UPDATE products
SET
    sku = sqlc.arg(sku),
    barcode = sqlc.narg(barcode),
    name = sqlc.arg(name),
    category_id = sqlc.narg(category_id),
    price = sqlc.arg(price),
    cost = sqlc.narg(cost),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING
    id,
    sku,
    barcode,
    name,
    category_id,
    price,
    cost,
    is_active,
    created_at,
    updated_at;

-- name: ActivateProduct :one
UPDATE products
SET
    is_active = TRUE,
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING
    id,
    sku,
    barcode,
    name,
    category_id,
    price,
    cost,
    is_active,
    created_at,
    updated_at;

-- name: DeactivateProduct :one
UPDATE products
SET
    is_active = FALSE,
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING
    id,
    sku,
    barcode,
    name,
    category_id,
    price,
    cost,
    is_active,
    created_at,
    updated_at;
