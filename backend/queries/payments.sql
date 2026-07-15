-- name: CreatePayment :one
WITH inserted AS (
    INSERT INTO payments (
        sale_id,
        payment_method_id,
        amount,
        received_amount,
        change_amount,
        installments,
        external_reference,
        idempotency_key
    )
    VALUES (
        sqlc.arg(sale_id),
        sqlc.arg(payment_method_id),
        sqlc.arg(amount),
        sqlc.narg(received_amount),
        sqlc.narg(change_amount),
        sqlc.arg(installments),
        sqlc.narg(external_reference),
        sqlc.arg(idempotency_key)
    )
    ON CONFLICT (idempotency_key) DO NOTHING
    RETURNING
        id,
        sale_id,
        payment_method_id,
        status,
        amount,
        received_amount,
        change_amount,
        installments,
        external_reference,
        paid_at,
        cancelled_at,
        created_at,
        updated_at,
        idempotency_key
)
SELECT
    id,
    sale_id,
    payment_method_id,
    status,
    amount,
    received_amount,
    change_amount,
    installments,
    external_reference,
    paid_at,
    cancelled_at,
    created_at,
    updated_at,
    idempotency_key
FROM inserted
UNION ALL
SELECT
    id,
    sale_id,
    payment_method_id,
    status,
    amount,
    received_amount,
    change_amount,
    installments,
    external_reference,
    paid_at,
    cancelled_at,
    created_at,
    updated_at,
    idempotency_key
FROM payments
WHERE idempotency_key = sqlc.arg(idempotency_key)
LIMIT 1;

-- name: GetPaymentByID :one
SELECT
    id,
    sale_id,
    payment_method_id,
    status,
    amount,
    received_amount,
    change_amount,
    installments,
    external_reference,
    paid_at,
    cancelled_at,
    created_at,
    updated_at,
    idempotency_key
FROM payments
WHERE id = sqlc.arg(id)
LIMIT 1;

-- name: GetPaymentByIdempotencyKey :one
SELECT
    id,
    sale_id,
    payment_method_id,
    status,
    amount,
    received_amount,
    change_amount,
    installments,
    external_reference,
    paid_at,
    cancelled_at,
    created_at,
    updated_at,
    idempotency_key
FROM payments
WHERE idempotency_key = sqlc.arg(idempotency_key)
LIMIT 1;

-- name: GetPaymentByExternalReference :one
SELECT
    id,
    sale_id,
    payment_method_id,
    status,
    amount,
    received_amount,
    change_amount,
    installments,
    external_reference,
    paid_at,
    cancelled_at,
    created_at,
    updated_at,
    idempotency_key
FROM payments
WHERE payment_method_id = sqlc.arg(payment_method_id)
  AND external_reference = sqlc.arg(external_reference)
LIMIT 1;

-- name: ListPaymentsBySaleID :many
SELECT
    id,
    sale_id,
    payment_method_id,
    status,
    amount,
    received_amount,
    change_amount,
    installments,
    external_reference,
    paid_at,
    cancelled_at,
    created_at,
    updated_at,
    idempotency_key
FROM payments
WHERE sale_id = sqlc.arg(sale_id)
ORDER BY created_at, id;

-- name: LockPaymentByID :one
SELECT
    id,
    sale_id,
    payment_method_id,
    status,
    amount,
    received_amount,
    change_amount,
    installments,
    external_reference,
    paid_at,
    cancelled_at,
    created_at,
    updated_at,
    idempotency_key
FROM payments
WHERE id = sqlc.arg(id)
FOR UPDATE;

-- name: ApprovePayment :one
UPDATE payments
SET
    status = 'APPROVED',
    paid_at = NOW(),
    received_amount = COALESCE(sqlc.narg(received_amount), received_amount),
    change_amount = COALESCE(sqlc.narg(change_amount), change_amount)
WHERE id = sqlc.arg(id)
  AND status = 'PENDING'
RETURNING
    id,
    sale_id,
    payment_method_id,
    status,
    amount,
    received_amount,
    change_amount,
    installments,
    external_reference,
    paid_at,
    cancelled_at,
    created_at,
    updated_at,
    idempotency_key;

-- name: DeclinePayment :one
UPDATE payments
SET
    status = 'DECLINED'
WHERE id = sqlc.arg(id)
  AND status = 'PENDING'
RETURNING
    id,
    sale_id,
    payment_method_id,
    status,
    amount,
    received_amount,
    change_amount,
    installments,
    external_reference,
    paid_at,
    cancelled_at,
    created_at,
    updated_at,
    idempotency_key;

-- name: CancelPayment :one
UPDATE payments
SET
    status = 'CANCELLED',
    cancelled_at = NOW()
WHERE id = sqlc.arg(id)
  AND status IN ('PENDING', 'APPROVED', 'DECLINED')
RETURNING
    id,
    sale_id,
    payment_method_id,
    status,
    amount,
    received_amount,
    change_amount,
    installments,
    external_reference,
    paid_at,
    cancelled_at,
    created_at,
    updated_at,
    idempotency_key;
