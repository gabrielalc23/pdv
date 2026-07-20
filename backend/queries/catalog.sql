-- name: ListCatalogProductsForStore :many
SELECT
    p.organization_id,
    s.id AS store_id,
    p.id,
    p.sku,
    p.barcode,
    p.name,
    p.category_id,
    c.name AS category_name,
    p.price,
    COALESCE(i.quantity, CAST(0 AS NUMERIC(15, 3))) AS quantity,
    p.is_active,
    COALESCE(i.quantity, CAST(0 AS NUMERIC(15, 3))) > 0 AS in_stock,
    p.created_at,
    p.updated_at
FROM products p
INNER JOIN stores s
        ON s.organization_id = p.organization_id
       AND s.id = sqlc.arg(store_id)
LEFT JOIN inventory i
       ON i.organization_id = p.organization_id
      AND i.store_id = s.id
      AND i.product_id = p.id
LEFT JOIN categories c
       ON c.organization_id = p.organization_id
      AND c.id = p.category_id
WHERE p.organization_id = sqlc.arg(organization_id)
  AND (
      CAST(sqlc.narg(search) AS TEXT) IS NULL
      OR p.name ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
      OR p.sku ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
      OR p.barcode ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
  )
  AND (
      CAST(sqlc.narg(category_id) AS UUID) IS NULL
      OR p.category_id = CAST(sqlc.narg(category_id) AS UUID)
  )
  AND (NOT CAST(sqlc.arg(active_only) AS BOOLEAN) OR p.is_active = TRUE)
  AND (
      NOT CAST(sqlc.arg(in_stock_only) AS BOOLEAN)
      OR COALESCE(i.quantity, CAST(0 AS NUMERIC(15, 3))) > 0
  )
ORDER BY p.name ASC, p.id ASC
LIMIT sqlc.arg(page_size)
OFFSET sqlc.arg(page_offset);

-- name: CountCatalogProductsForStore :one
SELECT COUNT(*)
FROM products p
INNER JOIN stores s
        ON s.organization_id = p.organization_id
       AND s.id = sqlc.arg(store_id)
LEFT JOIN inventory i
       ON i.organization_id = p.organization_id
      AND i.store_id = s.id
      AND i.product_id = p.id
LEFT JOIN categories c
       ON c.organization_id = p.organization_id
      AND c.id = p.category_id
WHERE p.organization_id = sqlc.arg(organization_id)
  AND (
      CAST(sqlc.narg(search) AS TEXT) IS NULL
      OR p.name ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
      OR p.sku ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
      OR p.barcode ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
  )
  AND (
      CAST(sqlc.narg(category_id) AS UUID) IS NULL
      OR p.category_id = CAST(sqlc.narg(category_id) AS UUID)
  )
  AND (NOT CAST(sqlc.arg(active_only) AS BOOLEAN) OR p.is_active = TRUE)
  AND (
      NOT CAST(sqlc.arg(in_stock_only) AS BOOLEAN)
      OR COALESCE(i.quantity, CAST(0 AS NUMERIC(15, 3))) > 0
  );

-- name: GetCatalogProductByIDForStore :one
SELECT
    p.organization_id,
    s.id AS store_id,
    p.id,
    p.sku,
    p.barcode,
    p.name,
    p.category_id,
    c.name AS category_name,
    p.price,
    COALESCE(i.quantity, CAST(0 AS NUMERIC(15, 3))) AS quantity,
    p.is_active,
    COALESCE(i.quantity, CAST(0 AS NUMERIC(15, 3))) > 0 AS in_stock,
    p.created_at,
    p.updated_at
FROM products p
INNER JOIN stores s
        ON s.organization_id = p.organization_id
       AND s.id = sqlc.arg(store_id)
LEFT JOIN inventory i
       ON i.organization_id = p.organization_id
      AND i.store_id = s.id
      AND i.product_id = p.id
LEFT JOIN categories c
       ON c.organization_id = p.organization_id
      AND c.id = p.category_id
WHERE p.organization_id = sqlc.arg(organization_id)
  AND p.id = sqlc.arg(id)
LIMIT 1;

-- name: GetCatalogProductByBarcodeForStore :one
SELECT
    p.organization_id,
    s.id AS store_id,
    p.id,
    p.sku,
    p.barcode,
    p.name,
    p.category_id,
    c.name AS category_name,
    p.price,
    COALESCE(i.quantity, CAST(0 AS NUMERIC(15, 3))) AS quantity,
    p.is_active,
    COALESCE(i.quantity, CAST(0 AS NUMERIC(15, 3))) > 0 AS in_stock,
    p.created_at,
    p.updated_at
FROM products p
INNER JOIN stores s
        ON s.organization_id = p.organization_id
       AND s.id = sqlc.arg(store_id)
LEFT JOIN inventory i
       ON i.organization_id = p.organization_id
      AND i.store_id = s.id
      AND i.product_id = p.id
LEFT JOIN categories c
       ON c.organization_id = p.organization_id
      AND c.id = p.category_id
WHERE p.organization_id = sqlc.arg(organization_id)
  AND p.barcode = sqlc.arg(barcode)
LIMIT 1;
