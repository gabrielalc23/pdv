-- name: CreateSale :one
WITH inserted AS (
    INSERT INTO sales (
        status,
        subtotal,
        discount,
        addition,
        total,
        idempotency_key
    )
    VALUES (
        'OPEN',
        0,
        0,
        0,
        0,
        sqlc.arg(idempotency_key)
    )
    ON CONFLICT (idempotency_key) DO NOTHING
    RETURNING
        id,
        number,
        status,
        subtotal,
        discount,
        addition,
        total,
        opened_at,
        completed_at,
        cancelled_at,
        created_at,
        updated_at,
        idempotency_key
)
SELECT
    id,
    number,
    status,
    subtotal,
    discount,
    addition,
    total,
    opened_at,
    completed_at,
    cancelled_at,
    created_at,
    updated_at,
    idempotency_key
FROM inserted
UNION ALL
SELECT
    id,
    number,
    status,
    subtotal,
    discount,
    addition,
    total,
    opened_at,
    completed_at,
    cancelled_at,
    created_at,
    updated_at,
    idempotency_key
FROM sales
WHERE idempotency_key = sqlc.arg(idempotency_key)
LIMIT 1;

-- name: GetSaleByID :one
SELECT
    id,
    number,
    status,
    subtotal,
    discount,
    addition,
    total,
    opened_at,
    completed_at,
    cancelled_at,
    created_at,
    updated_at,
    idempotency_key
FROM sales
WHERE id = sqlc.arg(id)
LIMIT 1;

-- name: LockSaleByID :one
SELECT
    id,
    number,
    status,
    subtotal,
    discount,
    addition,
    total,
    opened_at,
    completed_at,
    cancelled_at,
    created_at,
    updated_at,
    idempotency_key
FROM sales
WHERE id = sqlc.arg(id)
FOR UPDATE;

-- name: ListSales :many
SELECT
    id,
    number,
    status,
    subtotal,
    discount,
    addition,
    total,
    opened_at,
    completed_at,
    cancelled_at,
    created_at,
    updated_at,
    idempotency_key
FROM sales
WHERE (
    CAST(sqlc.narg(status) AS sale_status) IS NULL
    OR status = CAST(sqlc.narg(status) AS sale_status)
)
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(page_size)
OFFSET sqlc.arg(page_offset);

-- name: CountSales :one
SELECT COUNT(*)
FROM sales
WHERE (
    CAST(sqlc.narg(status) AS sale_status) IS NULL
    OR status = CAST(sqlc.narg(status) AS sale_status)
);

-- name: GetSaleItemByID :one
SELECT
    id,
    sale_id,
    product_id,
    product_name,
    product_sku,
    unit_price,
    quantity,
    discount,
    total,
    created_at
FROM sale_items
WHERE sale_id = sqlc.arg(sale_id)
  AND id = sqlc.arg(id)
LIMIT 1;

-- name: CreateSaleItem :one
INSERT INTO sale_items (
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
    sale_id,
    product_id,
    product_name,
    product_sku,
    unit_price,
    quantity,
    discount,
    total,
    created_at;

-- name: UpdateSaleItem :one
UPDATE sale_items
SET
    quantity = sqlc.arg(quantity),
    discount = sqlc.arg(discount),
    total = sqlc.arg(total)
WHERE sale_id = sqlc.arg(sale_id)
  AND id = sqlc.arg(id)
RETURNING
    id,
    sale_id,
    product_id,
    product_name,
    product_sku,
    unit_price,
    quantity,
    discount,
    total,
    created_at;

-- name: DeleteSaleItem :one
DELETE FROM sale_items
WHERE sale_id = sqlc.arg(sale_id)
  AND id = sqlc.arg(id)
RETURNING
    id,
    sale_id,
    product_id,
    product_name,
    product_sku,
    unit_price,
    quantity,
    discount,
    total,
    created_at;

-- name: ListSaleItemsBySaleID :many
SELECT
    id,
    sale_id,
    product_id,
    product_name,
    product_sku,
    unit_price,
    quantity,
    discount,
    total,
    created_at
FROM sale_items
WHERE sale_id = sqlc.arg(sale_id)
ORDER BY created_at ASC, id ASC;

-- name: RecalculateSaleTotals :one
UPDATE sales
SET
    subtotal = sqlc.arg(subtotal),
    discount = sqlc.arg(discount),
    addition = sqlc.arg(addition),
    total = sqlc.arg(total)
WHERE id = sqlc.arg(id)
  AND status = 'OPEN'
RETURNING
    id,
    number,
    status,
    subtotal,
    discount,
    addition,
    total,
    opened_at,
    completed_at,
    cancelled_at,
    created_at,
    updated_at,
    idempotency_key;

-- name: CancelSale :one
UPDATE sales
SET
    status = 'CANCELLED',
    cancelled_at = NOW()
WHERE id = sqlc.arg(id)
  AND status = 'OPEN'
RETURNING
    id,
    number,
    status,
    subtotal,
    discount,
    addition,
    total,
    opened_at,
    completed_at,
    cancelled_at,
    created_at,
    updated_at,
    idempotency_key;
