-- name: CreatePaymentMethodForOrganization :one
INSERT INTO payment_methods (
    organization_id,
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
    sqlc.arg(organization_id),
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
    organization_id,
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

-- name: GetPaymentMethodByIDForOrganization :one
SELECT
    method.id,
    method.organization_id,
    method.code,
    method.name,
    method.kind,
    method.provider,
    method.allows_change,
    method.requires_external_reference,
    method.allows_installments,
    method.max_installments,
    method.fee_percentage,
    method.settlement_days,
    method.is_active,
    method.sort_order,
    method.created_at,
    method.updated_at
FROM payment_methods AS method
WHERE method.organization_id = sqlc.arg(organization_id)
  AND method.id = sqlc.arg(id)
LIMIT 1;

-- name: GetPaymentMethodByCodeForOrganization :one
SELECT
    method.id,
    method.organization_id,
    method.code,
    method.name,
    method.kind,
    method.provider,
    method.allows_change,
    method.requires_external_reference,
    method.allows_installments,
    method.max_installments,
    method.fee_percentage,
    method.settlement_days,
    method.is_active,
    method.sort_order,
    method.created_at,
    method.updated_at
FROM payment_methods AS method
WHERE method.organization_id = sqlc.arg(organization_id)
  AND method.code = sqlc.arg(code)
LIMIT 1;

-- name: ListPaymentMethodsForOrganization :many
SELECT
    method.id,
    method.organization_id,
    method.code,
    method.name,
    method.kind,
    method.provider,
    method.allows_change,
    method.requires_external_reference,
    method.allows_installments,
    method.max_installments,
    method.fee_percentage,
    method.settlement_days,
    method.is_active,
    method.sort_order,
    method.created_at,
    method.updated_at
FROM payment_methods AS method
WHERE method.organization_id = sqlc.arg(organization_id)
ORDER BY method.sort_order, method.name, method.id;

-- name: CountPaymentMethodsForOrganization :one
SELECT COUNT(*)
FROM payment_methods AS method
WHERE method.organization_id = sqlc.arg(organization_id);

-- name: ListActivePaymentMethodsForOrganization :many
SELECT
    method.id,
    method.organization_id,
    method.code,
    method.name,
    method.kind,
    method.provider,
    method.allows_change,
    method.requires_external_reference,
    method.allows_installments,
    method.max_installments,
    method.fee_percentage,
    method.settlement_days,
    method.is_active,
    method.sort_order,
    method.created_at,
    method.updated_at
FROM payment_methods AS method
WHERE method.organization_id = sqlc.arg(organization_id)
  AND method.is_active = TRUE
ORDER BY method.sort_order, method.name, method.id;

-- name: CountActivePaymentMethodsForOrganization :one
SELECT COUNT(*)
FROM payment_methods AS method
WHERE method.organization_id = sqlc.arg(organization_id)
  AND method.is_active = TRUE;

-- name: GetOperationalPaymentMethodByIDForStore :one
SELECT
    method.id,
    method.organization_id,
    binding.store_id,
    method.code,
    method.name,
    method.kind,
    method.provider,
    method.allows_change,
    method.requires_external_reference,
    method.allows_installments,
    method.max_installments,
    method.fee_percentage,
    method.settlement_days,
    method.is_active,
    binding.sort_order AS store_sort_order,
    method.created_at,
    method.updated_at
FROM store_payment_methods AS binding
INNER JOIN payment_methods AS method
    ON method.organization_id = binding.organization_id
   AND method.id = binding.payment_method_id
WHERE binding.organization_id = sqlc.arg(organization_id)
  AND binding.store_id = sqlc.arg(store_id)
  AND binding.payment_method_id = sqlc.arg(id)
  AND binding.is_active = TRUE
  AND method.is_active = TRUE
LIMIT 1;

-- name: ListOperationalPaymentMethodsForStore :many
SELECT
    method.id,
    method.organization_id,
    binding.store_id,
    method.code,
    method.name,
    method.kind,
    method.provider,
    method.allows_change,
    method.requires_external_reference,
    method.allows_installments,
    method.max_installments,
    method.fee_percentage,
    method.settlement_days,
    method.is_active,
    binding.sort_order AS store_sort_order,
    method.created_at,
    method.updated_at
FROM store_payment_methods AS binding
INNER JOIN payment_methods AS method
    ON method.organization_id = binding.organization_id
   AND method.id = binding.payment_method_id
WHERE binding.organization_id = sqlc.arg(organization_id)
  AND binding.store_id = sqlc.arg(store_id)
  AND binding.is_active = TRUE
  AND method.is_active = TRUE
ORDER BY binding.sort_order, method.name, method.id;

-- name: CountOperationalPaymentMethodsForStore :one
SELECT COUNT(*)
FROM store_payment_methods AS binding
INNER JOIN payment_methods AS method
    ON method.organization_id = binding.organization_id
   AND method.id = binding.payment_method_id
WHERE binding.organization_id = sqlc.arg(organization_id)
  AND binding.store_id = sqlc.arg(store_id)
  AND binding.is_active = TRUE
  AND method.is_active = TRUE;

-- name: UpdatePaymentMethodForOrganization :one
UPDATE payment_methods AS method
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
WHERE method.organization_id = sqlc.arg(organization_id)
  AND method.id = sqlc.arg(id)
RETURNING
    method.id,
    method.organization_id,
    method.code,
    method.name,
    method.kind,
    method.provider,
    method.allows_change,
    method.requires_external_reference,
    method.allows_installments,
    method.max_installments,
    method.fee_percentage,
    method.settlement_days,
    method.is_active,
    method.sort_order,
    method.created_at,
    method.updated_at;

-- name: ActivatePaymentMethodForOrganization :one
UPDATE payment_methods AS method
SET is_active = TRUE
WHERE method.organization_id = sqlc.arg(organization_id)
  AND method.id = sqlc.arg(id)
RETURNING
    method.id,
    method.organization_id,
    method.code,
    method.name,
    method.kind,
    method.provider,
    method.allows_change,
    method.requires_external_reference,
    method.allows_installments,
    method.max_installments,
    method.fee_percentage,
    method.settlement_days,
    method.is_active,
    method.sort_order,
    method.created_at,
    method.updated_at;

-- name: DeactivatePaymentMethodForOrganization :one
UPDATE payment_methods AS method
SET is_active = FALSE
WHERE method.organization_id = sqlc.arg(organization_id)
  AND method.id = sqlc.arg(id)
RETURNING
    method.id,
    method.organization_id,
    method.code,
    method.name,
    method.kind,
    method.provider,
    method.allows_change,
    method.requires_external_reference,
    method.allows_installments,
    method.max_installments,
    method.fee_percentage,
    method.settlement_days,
    method.is_active,
    method.sort_order,
    method.created_at,
    method.updated_at;
