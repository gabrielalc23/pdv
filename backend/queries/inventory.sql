-- name: GetInventoryByProductID :one
SELECT
    product_id,
    quantity,
    created_at,
    updated_at
FROM inventory
WHERE product_id = sqlc.arg(product_id)
LIMIT 1;

-- name: ListInventory :many
SELECT
    p.id AS product_id,
    p.sku,
    p.barcode,
    p.name,
    p.is_active,
    i.quantity,
    i.created_at,
    i.updated_at
FROM inventory i
INNER JOIN products p ON p.id = i.product_id
WHERE
    (
        CAST(sqlc.narg(search) AS TEXT) IS NULL
        OR p.name ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
        OR p.sku ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
        OR p.barcode ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
    )
    AND (NOT sqlc.arg(active_only) OR p.is_active = TRUE)
ORDER BY p.name ASC, p.id ASC
LIMIT sqlc.arg(page_size)
OFFSET sqlc.arg(page_offset);

-- name: CountInventory :one
SELECT COUNT(*)
FROM inventory i
INNER JOIN products p ON p.id = i.product_id
WHERE
    (
        CAST(sqlc.narg(search) AS TEXT) IS NULL
        OR p.name ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
        OR p.sku ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
        OR p.barcode ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
    )
    AND (NOT sqlc.arg(active_only) OR p.is_active = TRUE);

-- name: IncreaseInventory :one
INSERT INTO inventory (
    product_id,
    quantity
)
VALUES (
    sqlc.arg(product_id),
    sqlc.arg(quantity)
)
ON CONFLICT (product_id)
DO UPDATE SET
    quantity = inventory.quantity + EXCLUDED.quantity,
    updated_at = NOW()
RETURNING
    product_id,
    CAST(quantity - sqlc.arg(quantity) AS NUMERIC(15, 3)) AS previous_quantity,
    quantity AS current_quantity,
    created_at,
    updated_at;

-- name: DecreaseInventory :one
UPDATE inventory
SET
    quantity = quantity - sqlc.arg(quantity),
    updated_at = NOW()
WHERE product_id = sqlc.arg(product_id)
  AND quantity >= sqlc.arg(quantity)
RETURNING
    product_id,
    CAST(quantity + sqlc.arg(quantity) AS NUMERIC(15, 3)) AS previous_quantity,
    quantity AS current_quantity,
    created_at,
    updated_at;

-- name: CreateInventoryMovement :one
INSERT INTO inventory_movements (
    product_id,
    movement_type,
    quantity,
    previous_quantity,
    current_quantity,
    reason,
    reference_type,
    reference_id
)
VALUES (
    sqlc.arg(product_id),
    sqlc.arg(movement_type),
    sqlc.arg(quantity),
    sqlc.arg(previous_quantity),
    sqlc.arg(current_quantity),
    sqlc.narg(reason),
    sqlc.arg(reference_type),
    sqlc.arg(reference_id)
)
RETURNING
    id,
    product_id,
    movement_type,
    quantity,
    previous_quantity,
    current_quantity,
    reason,
    reference_type,
    reference_id,
    created_at;

-- name: GetInventoryMovementByReference :one
SELECT
    id,
    product_id,
    movement_type,
    quantity,
    previous_quantity,
    current_quantity,
    reason,
    reference_type,
    reference_id,
    created_at
FROM inventory_movements
WHERE product_id = sqlc.arg(product_id)
  AND movement_type = sqlc.arg(movement_type)
  AND reference_type = sqlc.arg(reference_type)
  AND reference_id = sqlc.arg(reference_id)
LIMIT 1;

-- name: ListInventoryMovementsByProductID :many
SELECT
    id,
    product_id,
    movement_type,
    quantity,
    previous_quantity,
    current_quantity,
    reason,
    reference_type,
    reference_id,
    created_at
FROM inventory_movements
WHERE product_id = sqlc.arg(product_id)
  AND (
      CAST(sqlc.narg(movement_type_filter) AS inventory_movement_type) IS NULL
      OR movement_type = CAST(sqlc.narg(movement_type_filter) AS inventory_movement_type)
  )
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(page_size)
OFFSET sqlc.arg(page_offset);

-- name: CountInventoryMovementsByProductID :one
SELECT COUNT(*)
FROM inventory_movements
WHERE product_id = sqlc.arg(product_id)
  AND (
      CAST(sqlc.narg(movement_type_filter) AS inventory_movement_type) IS NULL
      OR movement_type = CAST(sqlc.narg(movement_type_filter) AS inventory_movement_type)
  );
