-- name: CreateSaleForStore :one
WITH existing AS MATERIALIZED (
    SELECT
        sale.id,
        sale.organization_id,
        sale.store_id,
        sale.number,
        sale.idempotency_key,
        sale.status,
        sale.subtotal,
        sale.discount,
        sale.addition,
        sale.total,
        sale.opened_by_membership_id,
        sale.completed_by_membership_id,
        sale.cancelled_by_membership_id,
        sale.opened_at,
        sale.completed_at,
        sale.cancelled_at,
        sale.created_at,
        sale.updated_at
    FROM sales AS sale
    WHERE sale.organization_id = sqlc.arg(organization_id)
      AND sale.store_id = sqlc.arg(store_id)
      AND sale.idempotency_key = sqlc.arg(idempotency_key)
), allocated_number AS (
    INSERT INTO sale_number_counters (
        organization_id,
        store_id,
        next_number
    )
    SELECT
        sqlc.arg(organization_id),
        sqlc.arg(store_id),
        2
    WHERE NOT EXISTS (SELECT 1 FROM existing)
    ON CONFLICT (organization_id, store_id) DO UPDATE
    SET next_number = sale_number_counters.next_number + 1
    RETURNING next_number - 1 AS number
), inserted AS (
    INSERT INTO sales (
        organization_id,
        store_id,
        number,
        idempotency_key,
        opened_by_membership_id
    )
    SELECT
        sqlc.arg(organization_id),
        sqlc.arg(store_id),
        allocated_number.number,
        sqlc.arg(idempotency_key),
        sqlc.arg(opened_by_membership_id)
    FROM allocated_number
    ON CONFLICT (organization_id, store_id, idempotency_key) DO UPDATE
    SET idempotency_key = EXCLUDED.idempotency_key
    RETURNING
        id,
        organization_id,
        store_id,
        number,
        idempotency_key,
        status,
        subtotal,
        discount,
        addition,
        total,
        opened_by_membership_id,
        completed_by_membership_id,
        cancelled_by_membership_id,
        opened_at,
        completed_at,
        cancelled_at,
        created_at,
        updated_at
)
SELECT * FROM inserted
UNION ALL
SELECT * FROM existing
UNION ALL
SELECT
    sale.id,
    sale.organization_id,
    sale.store_id,
    sale.number,
    sale.idempotency_key,
    sale.status,
    sale.subtotal,
    sale.discount,
    sale.addition,
    sale.total,
    sale.opened_by_membership_id,
    sale.completed_by_membership_id,
    sale.cancelled_by_membership_id,
    sale.opened_at,
    sale.completed_at,
    sale.cancelled_at,
    sale.created_at,
    sale.updated_at
FROM sales AS sale
WHERE sale.organization_id = sqlc.arg(organization_id)
  AND sale.store_id = sqlc.arg(store_id)
  AND sale.idempotency_key = sqlc.arg(idempotency_key)
  AND NOT EXISTS (SELECT 1 FROM inserted)
  AND NOT EXISTS (SELECT 1 FROM existing)
LIMIT 1;

-- name: GetSaleByIDForStore :one
SELECT
    sale.id,
    sale.organization_id,
    sale.store_id,
    sale.number,
    sale.idempotency_key,
    sale.status,
    sale.subtotal,
    sale.discount,
    sale.addition,
    sale.total,
    sale.opened_by_membership_id,
    sale.completed_by_membership_id,
    sale.cancelled_by_membership_id,
    sale.opened_at,
    sale.completed_at,
    sale.cancelled_at,
    sale.created_at,
    sale.updated_at
FROM sales AS sale
WHERE sale.organization_id = sqlc.arg(organization_id)
  AND sale.store_id = sqlc.arg(store_id)
  AND sale.id = sqlc.arg(id)
LIMIT 1;

-- name: GetSaleByIdempotencyKeyForStore :one
SELECT
    sale.id,
    sale.organization_id,
    sale.store_id,
    sale.number,
    sale.idempotency_key,
    sale.status,
    sale.subtotal,
    sale.discount,
    sale.addition,
    sale.total,
    sale.opened_by_membership_id,
    sale.completed_by_membership_id,
    sale.cancelled_by_membership_id,
    sale.opened_at,
    sale.completed_at,
    sale.cancelled_at,
    sale.created_at,
    sale.updated_at
FROM sales AS sale
WHERE sale.organization_id = sqlc.arg(organization_id)
  AND sale.store_id = sqlc.arg(store_id)
  AND sale.idempotency_key = sqlc.arg(idempotency_key)
LIMIT 1;

-- name: LockSaleByIDForStore :one
SELECT
    sale.id,
    sale.organization_id,
    sale.store_id,
    sale.number,
    sale.idempotency_key,
    sale.status,
    sale.subtotal,
    sale.discount,
    sale.addition,
    sale.total,
    sale.opened_by_membership_id,
    sale.completed_by_membership_id,
    sale.cancelled_by_membership_id,
    sale.opened_at,
    sale.completed_at,
    sale.cancelled_at,
    sale.created_at,
    sale.updated_at
FROM sales AS sale
WHERE sale.organization_id = sqlc.arg(organization_id)
  AND sale.store_id = sqlc.arg(store_id)
  AND sale.id = sqlc.arg(id)
FOR UPDATE;

-- name: ListSalesForStore :many
SELECT
    sale.id,
    sale.organization_id,
    sale.store_id,
    sale.number,
    sale.idempotency_key,
    sale.status,
    sale.subtotal,
    sale.discount,
    sale.addition,
    sale.total,
    sale.opened_by_membership_id,
    sale.completed_by_membership_id,
    sale.cancelled_by_membership_id,
    sale.opened_at,
    sale.completed_at,
    sale.cancelled_at,
    sale.created_at,
    sale.updated_at
FROM sales AS sale
WHERE sale.organization_id = sqlc.arg(organization_id)
  AND sale.store_id = sqlc.arg(store_id)
  AND (
      CAST(sqlc.narg(status) AS sale_status) IS NULL
      OR sale.status = CAST(sqlc.narg(status) AS sale_status)
  )
ORDER BY sale.created_at DESC, sale.id DESC
LIMIT sqlc.arg(page_size)
OFFSET sqlc.arg(page_offset);

-- name: CountSalesForStore :one
SELECT COUNT(*)
FROM sales AS sale
WHERE sale.organization_id = sqlc.arg(organization_id)
  AND sale.store_id = sqlc.arg(store_id)
  AND (
      CAST(sqlc.narg(status) AS sale_status) IS NULL
      OR sale.status = CAST(sqlc.narg(status) AS sale_status)
  );

-- name: GetSaleItemByIDForStore :one
SELECT
    item.id,
    item.organization_id,
    item.store_id,
    item.sale_id,
    item.product_id,
    item.product_name,
    item.product_sku,
    item.unit_price,
    item.quantity,
    item.discount,
    item.total,
    item.created_at
FROM sale_items AS item
WHERE item.organization_id = sqlc.arg(organization_id)
  AND item.store_id = sqlc.arg(store_id)
  AND item.sale_id = sqlc.arg(sale_id)
  AND item.id = sqlc.arg(id)
LIMIT 1;

-- name: CreateSaleItemForStore :one
INSERT INTO sale_items (
    organization_id,
    store_id,
    sale_id,
    product_id,
    product_name,
    product_sku,
    unit_price,
    quantity,
    discount,
    total
)
VALUES (
    sqlc.arg(organization_id),
    sqlc.arg(store_id),
    sqlc.arg(sale_id),
    sqlc.arg(product_id),
    sqlc.arg(product_name),
    sqlc.arg(product_sku),
    sqlc.arg(unit_price),
    sqlc.arg(quantity),
    sqlc.arg(discount),
    sqlc.arg(total)
)
RETURNING
    id,
    organization_id,
    store_id,
    sale_id,
    product_id,
    product_name,
    product_sku,
    unit_price,
    quantity,
    discount,
    total,
    created_at;

-- name: UpdateSaleItemForStore :one
UPDATE sale_items AS item
SET
    quantity = sqlc.arg(quantity),
    discount = sqlc.arg(discount),
    total = sqlc.arg(total)
WHERE item.organization_id = sqlc.arg(organization_id)
  AND item.store_id = sqlc.arg(store_id)
  AND item.sale_id = sqlc.arg(sale_id)
  AND item.id = sqlc.arg(id)
RETURNING
    item.id,
    item.organization_id,
    item.store_id,
    item.sale_id,
    item.product_id,
    item.product_name,
    item.product_sku,
    item.unit_price,
    item.quantity,
    item.discount,
    item.total,
    item.created_at;

-- name: DeleteSaleItemForStore :one
DELETE FROM sale_items AS item
WHERE item.organization_id = sqlc.arg(organization_id)
  AND item.store_id = sqlc.arg(store_id)
  AND item.sale_id = sqlc.arg(sale_id)
  AND item.id = sqlc.arg(id)
RETURNING
    item.id,
    item.organization_id,
    item.store_id,
    item.sale_id,
    item.product_id,
    item.product_name,
    item.product_sku,
    item.unit_price,
    item.quantity,
    item.discount,
    item.total,
    item.created_at;

-- name: ListSaleItemsBySaleIDForStore :many
SELECT
    item.id,
    item.organization_id,
    item.store_id,
    item.sale_id,
    item.product_id,
    item.product_name,
    item.product_sku,
    item.unit_price,
    item.quantity,
    item.discount,
    item.total,
    item.created_at
FROM sale_items AS item
WHERE item.organization_id = sqlc.arg(organization_id)
  AND item.store_id = sqlc.arg(store_id)
  AND item.sale_id = sqlc.arg(sale_id)
ORDER BY item.created_at ASC, item.id ASC;

-- name: CountSaleItemsBySaleIDForStore :one
SELECT COUNT(*)
FROM sale_items AS item
WHERE item.organization_id = sqlc.arg(organization_id)
  AND item.store_id = sqlc.arg(store_id)
  AND item.sale_id = sqlc.arg(sale_id);

-- name: RecalculateSaleTotalsForStore :one
UPDATE sales AS sale
SET
    subtotal = sqlc.arg(subtotal),
    discount = sqlc.arg(discount),
    addition = sqlc.arg(addition),
    total = sqlc.arg(total)
WHERE sale.organization_id = sqlc.arg(organization_id)
  AND sale.store_id = sqlc.arg(store_id)
  AND sale.id = sqlc.arg(id)
  AND sale.status = 'OPEN'
RETURNING
    sale.id,
    sale.organization_id,
    sale.store_id,
    sale.number,
    sale.idempotency_key,
    sale.status,
    sale.subtotal,
    sale.discount,
    sale.addition,
    sale.total,
    sale.opened_by_membership_id,
    sale.completed_by_membership_id,
    sale.cancelled_by_membership_id,
    sale.opened_at,
    sale.completed_at,
    sale.cancelled_at,
    sale.created_at,
    sale.updated_at;

-- name: CompleteSaleForStore :one
UPDATE sales AS sale
SET
    status = 'COMPLETED',
    completed_by_membership_id = sqlc.arg(completed_by_membership_id),
    completed_at = NOW()
WHERE sale.organization_id = sqlc.arg(organization_id)
  AND sale.store_id = sqlc.arg(store_id)
  AND sale.id = sqlc.arg(id)
  AND sale.status = 'OPEN'
RETURNING
    sale.id,
    sale.organization_id,
    sale.store_id,
    sale.number,
    sale.idempotency_key,
    sale.status,
    sale.subtotal,
    sale.discount,
    sale.addition,
    sale.total,
    sale.opened_by_membership_id,
    sale.completed_by_membership_id,
    sale.cancelled_by_membership_id,
    sale.opened_at,
    sale.completed_at,
    sale.cancelled_at,
    sale.created_at,
    sale.updated_at;

-- name: CancelSaleForStore :one
UPDATE sales AS sale
SET
    status = 'CANCELLED',
    cancelled_by_membership_id = sqlc.arg(cancelled_by_membership_id),
    cancelled_at = NOW()
WHERE sale.organization_id = sqlc.arg(organization_id)
  AND sale.store_id = sqlc.arg(store_id)
  AND sale.id = sqlc.arg(id)
  AND sale.status IN ('OPEN', 'COMPLETED')
RETURNING
    sale.id,
    sale.organization_id,
    sale.store_id,
    sale.number,
    sale.idempotency_key,
    sale.status,
    sale.subtotal,
    sale.discount,
    sale.addition,
    sale.total,
    sale.opened_by_membership_id,
    sale.completed_by_membership_id,
    sale.cancelled_by_membership_id,
    sale.opened_at,
    sale.completed_at,
    sale.cancelled_at,
    sale.created_at,
    sale.updated_at;
