CREATE VIEW receipt_payments AS
SELECT
    payment.organization_id,
    payment.store_id,
    payment.id,
    payment.sale_id,
    payment.payment_method_id,
    method.name AS payment_method_name,
    payment.status,
    payment.amount,
    payment.received_amount,
    payment.change_amount,
    payment.installments,
    payment.external_reference,
    payment.paid_at,
    payment.created_at,
    payment.updated_at
FROM
    payments AS payment
    JOIN payment_methods AS method ON method.organization_id = payment.organization_id
    AND method.id = payment.payment_method_id
WHERE
    payment.status = 'APPROVED';

COMMENT ON VIEW receipt_payments IS 'Approved receipt payments with explicit organization and store context; no refresh is required.';