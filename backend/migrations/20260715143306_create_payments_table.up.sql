-- Migration: create_payments
-- Purpose:
--   Creates payment transactions recorded during the POS checkout flow.
--
-- Business rules:
--   - A sale may contain multiple payments.
--   - Payment records cannot exist without a sale and a payment method.
--   - Monetary values use exact decimal representation.
--   - External transaction references cannot be duplicated within the
--     same configured payment method.
--   - Cash-specific rules, installment limits and payment totals are
--     validated by the application.
--
-- Dependencies:
--   - sales
--   - payment_methods

CREATE TYPE payment_status AS ENUM (
    'PENDING',
    'APPROVED',
    'DECLINED',
    'CANCELLED'
);

CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT uuidv7 (),
    sale_id UUID NOT NULL,
    payment_method_id UUID NOT NULL,
    idempotency_key VARCHAR(255) NOT NULL UNIQUE,
    status payment_status NOT NULL DEFAULT 'PENDING',
    amount NUMERIC(15, 2) NOT NULL,
    received_amount NUMERIC(15, 2),
    change_amount NUMERIC(15, 2),
    installments SMALLINT NOT NULL DEFAULT 1,
    external_reference VARCHAR(255),
    paid_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT payments_sale_fk FOREIGN KEY (sale_id) REFERENCES sales (id) ON DELETE RESTRICT,
    CONSTRAINT payments_method_fk FOREIGN KEY (payment_method_id) REFERENCES payment_methods (id) ON DELETE RESTRICT,
    CONSTRAINT payments_amount_positive CHECK (amount > 0),
    CONSTRAINT payments_received_amount_non_negative CHECK (
        received_amount IS NULL
        OR received_amount >= 0
    ),
    CONSTRAINT payments_change_amount_non_negative CHECK (
        change_amount IS NULL
        OR change_amount >= 0
    ),
    CONSTRAINT payments_installments_positive CHECK (installments >= 1),
    CONSTRAINT payments_external_reference_not_blank CHECK (
        external_reference IS NULL
        OR BTRIM (external_reference) <> ''
    ),
    CONSTRAINT payments_status_timestamps_consistency CHECK (
        (
            status = 'PENDING'
            AND paid_at IS NULL
            AND cancelled_at IS NULL
        )
        OR (
            status = 'APPROVED'
            AND paid_at IS NOT NULL
            AND cancelled_at IS NULL
        )
        OR (
            status = 'DECLINED'
            AND paid_at IS NULL
            AND cancelled_at IS NULL
        )
        OR (
            status = 'CANCELLED'
            AND cancelled_at IS NOT NULL
        )
    ),
    CONSTRAINT payments_updated_at_check CHECK (updated_at >= created_at)
);

CREATE INDEX idx_payments_sale_id ON payments (sale_id);

CREATE INDEX idx_payments_status_created_at ON payments (status, created_at DESC);

CREATE INDEX idx_payments_method_id ON payments (payment_method_id);

CREATE UNIQUE INDEX idx_payments_method_external_reference ON payments (
    payment_method_id,
    external_reference
)
WHERE
    external_reference IS NOT NULL;

COMMENT ON TYPE payment_status IS
    'Lifecycle status of a payment transaction.';

COMMENT ON
TABLE payments IS 'Stores individual payment transactions associated with POS sales.';

COMMENT ON COLUMN payments.sale_id IS 'Sale to which the payment is applied.';

COMMENT ON COLUMN payments.payment_method_id IS 'Configured payment method used by the transaction.';

COMMENT ON COLUMN payments.amount IS 'Amount applied to the sale by this payment.';

COMMENT ON COLUMN payments.received_amount IS 'Amount received from the customer when the payment method supports change.';

COMMENT ON COLUMN payments.change_amount IS 'Amount returned to the customer as change.';

COMMENT ON COLUMN payments.installments IS 'Number of installments selected for the payment.';

COMMENT ON COLUMN payments.external_reference IS 'Transaction identifier supplied by an external payment provider.';

COMMENT ON COLUMN payments.paid_at IS 'Timestamp at which the payment was approved.';

COMMENT ON COLUMN payments.cancelled_at IS 'Timestamp at which the payment was cancelled.';
