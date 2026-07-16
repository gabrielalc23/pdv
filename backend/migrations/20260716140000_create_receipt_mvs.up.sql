CREATE MATERIALIZED VIEW mv_receipt_payments AS
SELECT
    p.id,
    p.sale_id,
    p.payment_method_id,
    pm.name AS payment_method_name,
    p.status,
    p.amount,
    p.received_amount,
    p.change_amount,
    p.installments,
    p.external_reference,
    p.paid_at,
    p.created_at,
    p.updated_at
FROM payments p
JOIN payment_methods pm ON pm.id = p.payment_method_id
WHERE p.status = 'APPROVED';

CREATE UNIQUE INDEX idx_mv_receipt_payments_pk ON mv_receipt_payments (id);
CREATE INDEX idx_mv_receipt_payments_sale ON mv_receipt_payments (sale_id);
