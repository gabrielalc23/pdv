-- name: CreatePaymentMethod :one
INSERT INTO payment_methods (
    code,
    name,
    kind,
    provider,
    allows_change,
    requires_external_reference,
    allows_installments,
    max_installments,
    fee_percentage,
    settlement_days,
    is_active,
    sort_order
)
VALUES (
    sqlc.arg(code),
    sqlc.arg(name),
    sqlc.arg(kind),
    sqlc.narg(provider),
    sqlc.arg(allows_change),
    sqlc.arg(requires_external_reference),
    sqlc.arg(allows_installments),
    sqlc.arg(max_installments),
    sqlc.arg(fee_percentage),
    sqlc.arg(settlement_days),
    sqlc.arg(is_active),
    sqlc.arg(sort_order)
)
RETURNING
    id,
    code,
    name,
    kind,
    provider,
    allows_change,
    requires_external_reference,
    allows_installments,
    max_installments,
    fee_percentage,
    settlement_days,
    is_active,
    sort_order,
    created_at,
    updated_at;

-- name: GetPaymentMethodByID :one
SELECT
    id,
    code,
    name,
    kind,
    provider,
    allows_change,
    requires_external_reference,
    allows_installments,
    max_installments,
    fee_percentage,
    settlement_days,
    is_active,
    sort_order,
    created_at,
    updated_at
FROM payment_methods
WHERE id = sqlc.arg(id)
LIMIT 1;

-- name: GetPaymentMethodByCode :one
SELECT
    id,
    code,
    name,
    kind,
    provider,
    allows_change,
    requires_external_reference,
    allows_installments,
    max_installments,
    fee_percentage,
    settlement_days,
    is_active,
    sort_order,
    created_at,
    updated_at
FROM payment_methods
WHERE code = sqlc.arg(code)
LIMIT 1;

-- name: ListPaymentMethods :many
SELECT
    id,
    code,
    name,
    kind,
    provider,
    allows_change,
    requires_external_reference,
    allows_installments,
    max_installments,
    fee_percentage,
    settlement_days,
    is_active,
    sort_order,
    created_at,
    updated_at
FROM payment_methods
ORDER BY sort_order, name;

-- name: ListActivePaymentMethods :many
SELECT
    id,
    code,
    name,
    kind,
    provider,
    allows_change,
    requires_external_reference,
    allows_installments,
    max_installments,
    fee_percentage,
    settlement_days,
    is_active,
    sort_order,
    created_at,
    updated_at
FROM payment_methods
WHERE is_active = TRUE
ORDER BY sort_order, name;

-- name: UpdatePaymentMethod :one
UPDATE payment_methods
SET
    code = sqlc.arg(code),
    name = sqlc.arg(name),
    kind = sqlc.arg(kind),
    provider = sqlc.narg(provider),
    allows_change = sqlc.arg(allows_change),
    requires_external_reference = sqlc.arg(requires_external_reference),
    allows_installments = sqlc.arg(allows_installments),
    max_installments = sqlc.arg(max_installments),
    fee_percentage = sqlc.arg(fee_percentage),
    settlement_days = sqlc.arg(settlement_days),
    is_active = sqlc.arg(is_active),
    sort_order = sqlc.arg(sort_order)
WHERE id = sqlc.arg(id)
RETURNING
    id,
    code,
    name,
    kind,
    provider,
    allows_change,
    requires_external_reference,
    allows_installments,
    max_installments,
    fee_percentage,
    settlement_days,
    is_active,
    sort_order,
    created_at,
    updated_at;

-- name: ActivatePaymentMethod :one
UPDATE payment_methods
SET is_active = TRUE
WHERE id = sqlc.arg(id)
RETURNING
    id,
    code,
    name,
    kind,
    provider,
    allows_change,
    requires_external_reference,
    allows_installments,
    max_installments,
    fee_percentage,
    settlement_days,
    is_active,
    sort_order,
    created_at,
    updated_at;

-- name: DeactivatePaymentMethod :one
UPDATE payment_methods
SET is_active = FALSE
WHERE id = sqlc.arg(id)
RETURNING
    id,
    code,
    name,
    kind,
    provider,
    allows_change,
    requires_external_reference,
    allows_installments,
    max_installments,
    fee_percentage,
    settlement_days,
    is_active,
    sort_order,
    created_at,
    updated_at;
