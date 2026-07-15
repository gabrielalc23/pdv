-- name: ListCatalogProducts :many
SELECT
    p.id,
    p.sku,
    p.barcode,
    p.name,
    p.price,
    COALESCE(i.quantity, CAST(0 AS NUMERIC(15, 3))) AS quantity,
    p.is_active,
    COALESCE(i.quantity, CAST(0 AS NUMERIC(15, 3))) > 0 AS in_stock,
    p.created_at,
    p.updated_at
FROM products p
LEFT JOIN inventory i ON i.product_id = p.id
WHERE
    (
        CAST(sqlc.narg(search) AS TEXT) IS NULL
        OR p.name ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
        OR p.sku ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
        OR p.barcode ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
    )
    AND (NOT CAST(sqlc.arg(active_only) AS BOOLEAN) OR p.is_active = TRUE)
    AND (
        NOT CAST(sqlc.arg(in_stock_only) AS BOOLEAN)
        OR COALESCE(i.quantity, CAST(0 AS NUMERIC(15, 3))) > 0
    )
ORDER BY p.name ASC, p.id ASC
LIMIT sqlc.arg(page_size)
OFFSET sqlc.arg(page_offset);

-- name: CountCatalogProducts :one
SELECT COUNT(*)
FROM products p
LEFT JOIN inventory i ON i.product_id = p.id
WHERE
    (
        CAST(sqlc.narg(search) AS TEXT) IS NULL
        OR p.name ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
        OR p.sku ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
        OR p.barcode ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
    )
    AND (NOT CAST(sqlc.arg(active_only) AS BOOLEAN) OR p.is_active = TRUE)
    AND (
        NOT CAST(sqlc.arg(in_stock_only) AS BOOLEAN)
        OR COALESCE(i.quantity, CAST(0 AS NUMERIC(15, 3))) > 0
    );

-- name: GetCatalogProductByID :one
SELECT
    p.id,
    p.sku,
    p.barcode,
    p.name,
    p.price,
    COALESCE(i.quantity, CAST(0 AS NUMERIC(15, 3))) AS quantity,
    p.is_active,
    COALESCE(i.quantity, CAST(0 AS NUMERIC(15, 3))) > 0 AS in_stock,
    p.created_at,
    p.updated_at
FROM products p
LEFT JOIN inventory i ON i.product_id = p.id
WHERE p.id = sqlc.arg(id)
LIMIT 1;

-- name: GetCatalogProductByBarcode :one
SELECT
    p.id,
    p.sku,
    p.barcode,
    p.name,
    p.price,
    COALESCE(i.quantity, CAST(0 AS NUMERIC(15, 3))) AS quantity,
    p.is_active,
    COALESCE(i.quantity, CAST(0 AS NUMERIC(15, 3))) > 0 AS in_stock,
    p.created_at,
    p.updated_at
FROM products p
LEFT JOIN inventory i ON i.product_id = p.id
WHERE p.barcode = sqlc.arg(barcode)
LIMIT 1;
