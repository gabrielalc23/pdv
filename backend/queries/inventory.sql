-- name: GetInventoryByProductIDForStore :one
SELECT
    organization_id,
    store_id,
    product_id,
    quantity,
    created_at,
    updated_at
FROM inventory
WHERE organization_id = sqlc.arg(organization_id)
  AND store_id = sqlc.arg(store_id)
  AND product_id = sqlc.arg(product_id)
LIMIT 1;

-- name: ListInventoryForStore :many
SELECT
    i.organization_id,
    i.store_id,
    p.id AS product_id,
    p.sku,
    p.barcode,
    p.name,
    p.is_active,
    i.quantity,
    i.created_at,
    i.updated_at
FROM inventory i
INNER JOIN products p
        ON p.organization_id = i.organization_id
       AND p.id = i.product_id
WHERE i.organization_id = sqlc.arg(organization_id)
  AND i.store_id = sqlc.arg(store_id)
  AND (
      CAST(sqlc.narg(search) AS TEXT) IS NULL
      OR p.name ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
      OR p.sku ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
      OR p.barcode ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
  )
  AND (NOT CAST(sqlc.arg(active_only) AS BOOLEAN) OR p.is_active = TRUE)
ORDER BY p.name ASC, p.id ASC
LIMIT sqlc.arg(page_size)
OFFSET sqlc.arg(page_offset);

-- name: CountInventoryForStore :one
SELECT COUNT(*)
FROM inventory i
INNER JOIN products p
        ON p.organization_id = i.organization_id
       AND p.id = i.product_id
WHERE i.organization_id = sqlc.arg(organization_id)
  AND i.store_id = sqlc.arg(store_id)
  AND (
      CAST(sqlc.narg(search) AS TEXT) IS NULL
      OR p.name ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
      OR p.sku ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
      OR p.barcode ILIKE '%' || CAST(sqlc.narg(search) AS TEXT) || '%'
  )
  AND (NOT CAST(sqlc.arg(active_only) AS BOOLEAN) OR p.is_active = TRUE);

-- name: IncreaseInventoryForStore :one
INSERT INTO inventory (
    organization_id,
    store_id,
    product_id,
    quantity
)
SELECT
    sqlc.arg(organization_id),
    sqlc.arg(store_id),
    sqlc.arg(product_id),
    sqlc.arg(quantity)
WHERE sqlc.arg(quantity)::NUMERIC > 0
ON CONFLICT (organization_id, store_id, product_id)
DO UPDATE SET
    quantity = inventory.quantity + EXCLUDED.quantity,
    updated_at = NOW()
RETURNING
    organization_id,
    store_id,
    product_id,
    CAST(quantity - sqlc.arg(quantity) AS NUMERIC(15, 3)) AS previous_quantity,
    quantity AS current_quantity,
    created_at,
    updated_at;

-- name: DecreaseInventoryForStore :one
UPDATE inventory
SET
    quantity = quantity - sqlc.arg(quantity),
    updated_at = NOW()
WHERE organization_id = sqlc.arg(organization_id)
  AND store_id = sqlc.arg(store_id)
  AND product_id = sqlc.arg(product_id)
  AND sqlc.arg(quantity)::NUMERIC > 0
  AND quantity >= sqlc.arg(quantity)
RETURNING
    organization_id,
    store_id,
    product_id,
    CAST(quantity + sqlc.arg(quantity) AS NUMERIC(15, 3)) AS previous_quantity,
    quantity AS current_quantity,
    created_at,
    updated_at;

-- name: CreateInventoryMovementForStore :one
INSERT INTO inventory_movements (
    organization_id,
    store_id,
    product_id,
    actor_membership_id,
    movement_type,
    quantity,
    previous_quantity,
    current_quantity,
    reason,
    reference_type,
    reference_id
)
VALUES (
    sqlc.arg(organization_id),
    sqlc.arg(store_id),
    sqlc.arg(product_id),
    sqlc.arg(actor_membership_id),
    sqlc.arg(movement_type),
    sqlc.arg(quantity),
    sqlc.arg(previous_quantity),
    sqlc.arg(current_quantity),
    sqlc.narg(reason),
    sqlc.arg(reference_type),
    sqlc.arg(reference_id)
)
RETURNING
    organization_id,
    store_id,
    id,
    product_id,
    actor_membership_id,
    movement_type,
    quantity,
    previous_quantity,
    current_quantity,
    reason,
    reference_type,
    reference_id,
    created_at;

-- name: GetInventoryMovementByReferenceForStore :one
SELECT
    organization_id,
    store_id,
    id,
    product_id,
    actor_membership_id,
    movement_type,
    quantity,
    previous_quantity,
    current_quantity,
    reason,
    reference_type,
    reference_id,
    created_at
FROM inventory_movements
WHERE organization_id = sqlc.arg(organization_id)
  AND store_id = sqlc.arg(store_id)
  AND product_id = sqlc.arg(product_id)
  AND movement_type = sqlc.arg(movement_type)
  AND reference_type = sqlc.arg(reference_type)
  AND reference_id = sqlc.arg(reference_id)
LIMIT 1;

-- name: ListInventoryMovementsByProductIDForStore :many
SELECT
    organization_id,
    store_id,
    id,
    product_id,
    actor_membership_id,
    movement_type,
    quantity,
    previous_quantity,
    current_quantity,
    reason,
    reference_type,
    reference_id,
    created_at
FROM inventory_movements
WHERE organization_id = sqlc.arg(organization_id)
  AND store_id = sqlc.arg(store_id)
  AND product_id = sqlc.arg(product_id)
  AND (
      CAST(sqlc.narg(movement_type_filter) AS inventory_movement_type) IS NULL
      OR movement_type = CAST(sqlc.narg(movement_type_filter) AS inventory_movement_type)
  )
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(page_size)
OFFSET sqlc.arg(page_offset);

-- name: CountInventoryMovementsByProductIDForStore :one
SELECT COUNT(*)
FROM inventory_movements
WHERE organization_id = sqlc.arg(organization_id)
  AND store_id = sqlc.arg(store_id)
  AND product_id = sqlc.arg(product_id)
  AND (
      CAST(sqlc.narg(movement_type_filter) AS inventory_movement_type) IS NULL
      OR movement_type = CAST(sqlc.narg(movement_type_filter) AS inventory_movement_type)
  );
