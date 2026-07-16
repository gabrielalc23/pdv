-- name: ListReceiptPaymentsBySaleID :many
SELECT
    id,
    sale_id,
    payment_method_id,
    payment_method_name,
    status,
    amount,
    received_amount,
    change_amount,
    installments,
    external_reference,
    paid_at,
    created_at,
    updated_at
FROM mv_receipt_payments
WHERE sale_id = $1
ORDER BY created_at, id;
