CREATE TYPE payment_method_kind AS ENUM (
    'CASH',
    'PIX',
    'DEBIT_CARD',
    'CREDIT_CARD',
    'VOUCHER',
    'STORE_CREDIT',
    'OTHER'
);

CREATE TABLE payment_methods (
    id UUID PRIMARY KEY DEFAULT uuidv7 (),
    code VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    kind payment_method_kind NOT NULL,
    provider VARCHAR(100),
    allows_change BOOLEAN NOT NULL DEFAULT FALSE,
    requires_external_reference BOOLEAN NOT NULL DEFAULT FALSE,
    allows_installments BOOLEAN NOT NULL DEFAULT FALSE,
    max_installments SMALLINT NOT NULL DEFAULT 1,
    fee_percentage NUMERIC(7, 4) NOT NULL DEFAULT 0,
    settlement_days INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT payment_methods_code_unique UNIQUE (code),
    CONSTRAINT payment_methods_code_not_blank CHECK (BTRIM (code) <> ''),
    CONSTRAINT payment_methods_name_not_blank CHECK (BTRIM (name) <> ''),
    CONSTRAINT payment_methods_fee_non_negative CHECK (fee_percentage >= 0),
    CONSTRAINT payment_methods_settlement_days_non_negative CHECK (settlement_days >= 0),
    CONSTRAINT payment_methods_sort_order_non_negative CHECK (sort_order >= 0),
    CONSTRAINT payment_methods_max_installments_positive CHECK (max_installments >= 1),
    CONSTRAINT payment_methods_change_only_cash CHECK (
        allows_change = FALSE
        OR kind = 'CASH'
    ),
    CONSTRAINT payment_methods_installments_consistency CHECK (
        (
            allows_installments = TRUE
            AND kind = 'CREDIT_CARD'
            AND max_installments > 1
        )
        OR (
            allows_installments = FALSE
            AND max_installments = 1
        )
    ),
    CONSTRAINT payment_methods_external_reference_provider_consistency CHECK (
        requires_external_reference = FALSE
        OR (
            provider IS NOT NULL
            AND BTRIM(provider) <> ''
        )
    ),
    CONSTRAINT payment_methods_updated_at_check CHECK (updated_at >= created_at)
);

CREATE INDEX idx_payment_methods_active_sort ON payment_methods (is_active, sort_order);

CREATE INDEX idx_payment_methods_kind ON payment_methods (kind);
