-- name: CreatePaymentForStore :one
WITH existing AS MATERIALIZED (
    SELECT
        payment.id,
        payment.organization_id,
        payment.store_id,
        payment.sale_id,
        payment.payment_method_id,
        payment.idempotency_key,
        payment.status,
        payment.amount,
        payment.received_amount,
        payment.change_amount,
        payment.installments,
        payment.external_reference,
        payment.paid_at,
        payment.cancelled_at,
        payment.created_at,
        payment.updated_at
    FROM payments AS payment
    WHERE payment.organization_id = sqlc.arg(organization_id)
      AND payment.store_id = sqlc.arg(store_id)
      AND payment.idempotency_key = sqlc.arg(idempotency_key)
), inserted AS (
    INSERT INTO payments (
        organization_id,
        store_id,
        sale_id,
        payment_method_id,
        idempotency_key,
        amount,
        received_amount,
        change_amount,
        installments,
        external_reference
    )
    SELECT
        sqlc.arg(organization_id),
        sqlc.arg(store_id),
        sqlc.arg(sale_id),
        sqlc.arg(payment_method_id),
        sqlc.arg(idempotency_key),
        sqlc.arg(amount),
        sqlc.narg(received_amount),
        sqlc.narg(change_amount),
        sqlc.arg(installments),
        sqlc.narg(external_reference)
    FROM store_payment_methods AS binding
    INNER JOIN payment_methods AS method
        ON method.organization_id = binding.organization_id
       AND method.id = binding.payment_method_id
    WHERE binding.organization_id = sqlc.arg(organization_id)
      AND binding.store_id = sqlc.arg(store_id)
      AND binding.payment_method_id = sqlc.arg(payment_method_id)
      AND binding.is_active = TRUE
      AND method.is_active = TRUE
      AND NOT EXISTS (SELECT 1 FROM existing)
    ON CONFLICT (organization_id, store_id, idempotency_key) DO UPDATE
    SET idempotency_key = EXCLUDED.idempotency_key
    RETURNING
        id,
        organization_id,
        store_id,
        sale_id,
        payment_method_id,
        idempotency_key,
        status,
        amount,
        received_amount,
        change_amount,
        installments,
        external_reference,
        paid_at,
        cancelled_at,
        created_at,
        updated_at
)
SELECT * FROM inserted
UNION ALL
SELECT * FROM existing
UNION ALL
SELECT
    payment.id,
    payment.organization_id,
    payment.store_id,
    payment.sale_id,
    payment.payment_method_id,
    payment.idempotency_key,
    payment.status,
    payment.amount,
    payment.received_amount,
    payment.change_amount,
    payment.installments,
    payment.external_reference,
    payment.paid_at,
    payment.cancelled_at,
    payment.created_at,
    payment.updated_at
FROM payments AS payment
WHERE payment.organization_id = sqlc.arg(organization_id)
  AND payment.store_id = sqlc.arg(store_id)
  AND payment.idempotency_key = sqlc.arg(idempotency_key)
  AND NOT EXISTS (SELECT 1 FROM inserted)
  AND NOT EXISTS (SELECT 1 FROM existing)
LIMIT 1;

-- name: GetPaymentByIDForStore :one
SELECT
    payment.id,
    payment.organization_id,
    payment.store_id,
    payment.sale_id,
    payment.payment_method_id,
    payment.idempotency_key,
    payment.status,
    payment.amount,
    payment.received_amount,
    payment.change_amount,
    payment.installments,
    payment.external_reference,
    payment.paid_at,
    payment.cancelled_at,
    payment.created_at,
    payment.updated_at
FROM payments AS payment
WHERE payment.organization_id = sqlc.arg(organization_id)
  AND payment.store_id = sqlc.arg(store_id)
  AND payment.id = sqlc.arg(id)
LIMIT 1;

-- name: GetPaymentByIdempotencyKeyForStore :one
SELECT
    payment.id,
    payment.organization_id,
    payment.store_id,
    payment.sale_id,
    payment.payment_method_id,
    payment.idempotency_key,
    payment.status,
    payment.amount,
    payment.received_amount,
    payment.change_amount,
    payment.installments,
    payment.external_reference,
    payment.paid_at,
    payment.cancelled_at,
    payment.created_at,
    payment.updated_at
FROM payments AS payment
WHERE payment.organization_id = sqlc.arg(organization_id)
  AND payment.store_id = sqlc.arg(store_id)
  AND payment.idempotency_key = sqlc.arg(idempotency_key)
LIMIT 1;

-- name: GetPaymentByExternalReferenceForStore :one
SELECT
    payment.id,
    payment.organization_id,
    payment.store_id,
    payment.sale_id,
    payment.payment_method_id,
    payment.idempotency_key,
    payment.status,
    payment.amount,
    payment.received_amount,
    payment.change_amount,
    payment.installments,
    payment.external_reference,
    payment.paid_at,
    payment.cancelled_at,
    payment.created_at,
    payment.updated_at
FROM payments AS payment
WHERE payment.organization_id = sqlc.arg(organization_id)
  AND payment.store_id = sqlc.arg(store_id)
  AND payment.payment_method_id = sqlc.arg(payment_method_id)
  AND payment.external_reference = sqlc.arg(external_reference)
LIMIT 1;

-- name: ListPaymentsBySaleIDForStore :many
SELECT
    payment.id,
    payment.organization_id,
    payment.store_id,
    payment.sale_id,
    payment.payment_method_id,
    payment.idempotency_key,
    payment.status,
    payment.amount,
    payment.received_amount,
    payment.change_amount,
    payment.installments,
    payment.external_reference,
    payment.paid_at,
    payment.cancelled_at,
    payment.created_at,
    payment.updated_at
FROM payments AS payment
WHERE payment.organization_id = sqlc.arg(organization_id)
  AND payment.store_id = sqlc.arg(store_id)
  AND payment.sale_id = sqlc.arg(sale_id)
ORDER BY payment.created_at, payment.id;

-- name: CountPaymentsBySaleIDForStore :one
SELECT COUNT(*)
FROM payments AS payment
WHERE payment.organization_id = sqlc.arg(organization_id)
  AND payment.store_id = sqlc.arg(store_id)
  AND payment.sale_id = sqlc.arg(sale_id);

-- name: LockPaymentByIDForStore :one
SELECT
    payment.id,
    payment.organization_id,
    payment.store_id,
    payment.sale_id,
    payment.payment_method_id,
    payment.idempotency_key,
    payment.status,
    payment.amount,
    payment.received_amount,
    payment.change_amount,
    payment.installments,
    payment.external_reference,
    payment.paid_at,
    payment.cancelled_at,
    payment.created_at,
    payment.updated_at
FROM payments AS payment
WHERE payment.organization_id = sqlc.arg(organization_id)
  AND payment.store_id = sqlc.arg(store_id)
  AND payment.id = sqlc.arg(id)
FOR UPDATE;

-- name: ApprovePaymentForStore :one
UPDATE payments AS payment
SET
    status = 'APPROVED',
    paid_at = NOW(),
    received_amount = COALESCE(sqlc.narg(received_amount), payment.received_amount),
    change_amount = COALESCE(sqlc.narg(change_amount), payment.change_amount)
WHERE payment.organization_id = sqlc.arg(organization_id)
  AND payment.store_id = sqlc.arg(store_id)
  AND payment.id = sqlc.arg(id)
  AND payment.status = 'PENDING'
RETURNING
    payment.id,
    payment.organization_id,
    payment.store_id,
    payment.sale_id,
    payment.payment_method_id,
    payment.idempotency_key,
    payment.status,
    payment.amount,
    payment.received_amount,
    payment.change_amount,
    payment.installments,
    payment.external_reference,
    payment.paid_at,
    payment.cancelled_at,
    payment.created_at,
    payment.updated_at;

-- name: DeclinePaymentForStore :one
UPDATE payments AS payment
SET status = 'DECLINED'
WHERE payment.organization_id = sqlc.arg(organization_id)
  AND payment.store_id = sqlc.arg(store_id)
  AND payment.id = sqlc.arg(id)
  AND payment.status = 'PENDING'
RETURNING
    payment.id,
    payment.organization_id,
    payment.store_id,
    payment.sale_id,
    payment.payment_method_id,
    payment.idempotency_key,
    payment.status,
    payment.amount,
    payment.received_amount,
    payment.change_amount,
    payment.installments,
    payment.external_reference,
    payment.paid_at,
    payment.cancelled_at,
    payment.created_at,
    payment.updated_at;

-- name: CancelPaymentForStore :one
UPDATE payments AS payment
SET
    status = 'CANCELLED',
    cancelled_at = NOW()
WHERE payment.organization_id = sqlc.arg(organization_id)
  AND payment.store_id = sqlc.arg(store_id)
  AND payment.id = sqlc.arg(id)
  AND payment.status IN ('PENDING', 'APPROVED')
RETURNING
    payment.id,
    payment.organization_id,
    payment.store_id,
    payment.sale_id,
    payment.payment_method_id,
    payment.idempotency_key,
    payment.status,
    payment.amount,
    payment.received_amount,
    payment.change_amount,
    payment.installments,
    payment.external_reference,
    payment.paid_at,
    payment.cancelled_at,
    payment.created_at,
    payment.updated_at;
