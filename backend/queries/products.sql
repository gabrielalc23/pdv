-- name: CreateProductForOrganization :one
INSERT INTO products (
    organization_id,
    sku,
    barcode,
    name,
    category_id,
    price,
    cost,
    is_active
)
VALUES (
    sqlc.arg(organization_id),
    sqlc.arg(sku),
    sqlc.narg(barcode),
    sqlc.arg(name),
    sqlc.narg(category_id),
    sqlc.arg(price),
    sqlc.narg(cost),
    sqlc.arg(is_active)
)
RETURNING
    organization_id,
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

-- name: GetProductByIDForOrganization :one
SELECT
    organization_id,
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
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
LIMIT 1;

-- name: GetProductByIDForStore :one
SELECT
    p.organization_id,
    s.id AS store_id,
    p.id,
    p.sku,
    p.barcode,
    p.name,
    p.category_id,
    p.price,
    p.cost,
    p.is_active,
    p.created_at,
    p.updated_at
FROM products p
INNER JOIN stores s
        ON s.organization_id = p.organization_id
       AND s.id = sqlc.arg(store_id)
WHERE p.organization_id = sqlc.arg(organization_id)
  AND p.id = sqlc.arg(id)
LIMIT 1;

-- name: GetProductBySKUForOrganization :one
SELECT
    organization_id,
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
WHERE organization_id = sqlc.arg(organization_id)
  AND sku = sqlc.arg(sku)
LIMIT 1;

-- name: GetProductByBarcodeForOrganization :one
SELECT
    organization_id,
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
WHERE organization_id = sqlc.arg(organization_id)
  AND barcode = sqlc.arg(barcode)
LIMIT 1;

-- name: ListProductsForOrganization :many
SELECT
    organization_id,
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
WHERE organization_id = sqlc.arg(organization_id)
  AND (
      CAST(sqlc.narg(search) AS TEXT) IS NULL
      OR name ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
      OR sku ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
      OR barcode ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
  )
  AND (
      CAST(sqlc.narg(category_id) AS UUID) IS NULL
      OR category_id = CAST(sqlc.narg(category_id) AS UUID)
  )
  AND (NOT CAST(sqlc.arg(active_only) AS BOOLEAN) OR is_active = TRUE)
ORDER BY name ASC, id ASC
LIMIT sqlc.arg(page_size)
OFFSET sqlc.arg(page_offset);

-- name: CountProductsForOrganization :one
SELECT COUNT(*)
FROM products
WHERE organization_id = sqlc.arg(organization_id)
  AND (
      CAST(sqlc.narg(search) AS TEXT) IS NULL
      OR name ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
      OR sku ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
      OR barcode ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
  )
  AND (
      CAST(sqlc.narg(category_id) AS UUID) IS NULL
      OR category_id = CAST(sqlc.narg(category_id) AS UUID)
  )
  AND (NOT CAST(sqlc.arg(active_only) AS BOOLEAN) OR is_active = TRUE);

-- name: UpdateProductForOrganization :one
UPDATE products
SET
    sku = sqlc.arg(sku),
    barcode = sqlc.narg(barcode),
    name = sqlc.arg(name),
    category_id = sqlc.narg(category_id),
    price = sqlc.arg(price),
    cost = sqlc.narg(cost),
    updated_at = NOW()
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
RETURNING
    organization_id,
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

-- name: ActivateProductForOrganization :one
UPDATE products
SET
    is_active = TRUE,
    updated_at = NOW()
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
RETURNING
    organization_id,
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

-- name: DeactivateProductForOrganization :one
UPDATE products
SET
    is_active = FALSE,
    updated_at = NOW()
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
RETURNING
    organization_id,
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
