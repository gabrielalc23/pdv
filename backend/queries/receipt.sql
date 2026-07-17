-- name: ListReceiptPaymentsBySaleIDForStore :many
SELECT
    receipt.organization_id,
    receipt.store_id,
    receipt.id,
    receipt.sale_id,
    receipt.payment_method_id,
    receipt.payment_method_name,
    receipt.status,
    receipt.amount,
    receipt.received_amount,
    receipt.change_amount,
    receipt.installments,
    receipt.external_reference,
    receipt.paid_at,
    receipt.created_at,
    receipt.updated_at
FROM receipt_payments AS receipt
WHERE receipt.organization_id = sqlc.arg(organization_id)
  AND receipt.store_id = sqlc.arg(store_id)
  AND receipt.sale_id = sqlc.arg(sale_id)
ORDER BY receipt.created_at, receipt.id;

-- name: CountReceiptPaymentsBySaleIDForStore :one
SELECT COUNT(*)
FROM receipt_payments AS receipt
WHERE receipt.organization_id = sqlc.arg(organization_id)
  AND receipt.store_id = sqlc.arg(store_id)
  AND receipt.sale_id = sqlc.arg(sale_id);
