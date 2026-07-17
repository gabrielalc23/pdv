CREATE TYPE payment_method_kind AS ENUM (
    'CASH',
    'PIX',
    'DEBIT_CARD',
    'CREDIT_CARD',
    'VOUCHER',
    'STORE_CREDIT',
    'OTHER'
);

CREATE TYPE payment_status AS ENUM (
    'PENDING',
    'APPROVED',
    'DECLINED',
    'CANCELLED'
);

CREATE TABLE payment_methods (
    id UUID DEFAULT uuidv7() NOT NULL,
    organization_id UUID NOT NULL,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    kind payment_method_kind NOT NULL,
    provider VARCHAR(100),
    allows_change BOOLEAN DEFAULT FALSE NOT NULL,
    requires_external_reference BOOLEAN DEFAULT FALSE NOT NULL,
    allows_installments BOOLEAN DEFAULT FALSE NOT NULL,
    max_installments SMALLINT DEFAULT 1 NOT NULL,
    fee_percentage NUMERIC(7, 4) DEFAULT 0 NOT NULL,
    settlement_days INTEGER DEFAULT 0 NOT NULL,
    is_active BOOLEAN DEFAULT TRUE NOT NULL,
    sort_order INTEGER DEFAULT 0 NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    CONSTRAINT payment_methods_pkey PRIMARY KEY (id),
    CONSTRAINT payment_methods_organization_id_id_unique
        UNIQUE (organization_id, id),
    CONSTRAINT payment_methods_organization_id_code_unique
        UNIQUE (organization_id, code),
    CONSTRAINT payment_methods_organization_id_fkey
        FOREIGN KEY (organization_id)
        REFERENCES organizations (id) ON DELETE RESTRICT,
    CONSTRAINT payment_methods_code_not_blank CHECK (BTRIM(code) <> ''),
    CONSTRAINT payment_methods_name_not_blank CHECK (BTRIM(name) <> ''),
    CONSTRAINT payment_methods_fee_non_negative CHECK (fee_percentage >= 0),
    CONSTRAINT payment_methods_settlement_days_non_negative
        CHECK (settlement_days >= 0),
    CONSTRAINT payment_methods_sort_order_non_negative CHECK (sort_order >= 0),
    CONSTRAINT payment_methods_max_installments_positive
        CHECK (max_installments >= 1),
    CONSTRAINT payment_methods_change_only_cash CHECK (
        NOT allows_change OR kind = 'CASH'
    ),
    CONSTRAINT payment_methods_installments_consistency CHECK (
        (
            allows_installments
            AND kind = 'CREDIT_CARD'
            AND max_installments > 1
        )
        OR (NOT allows_installments AND max_installments = 1)
    ),
    CONSTRAINT payment_methods_external_reference_provider_consistency CHECK (
        NOT requires_external_reference
        OR (provider IS NOT NULL AND BTRIM(provider) <> '')
    ),
    CONSTRAINT payment_methods_updated_at_check CHECK (updated_at >= created_at)
);

CREATE INDEX idx_payment_methods_active_sort
ON payment_methods (organization_id, is_active, sort_order);

CREATE INDEX idx_payment_methods_kind
ON payment_methods (organization_id, kind);

CREATE TRIGGER trg_payment_methods_touch_updated_at
BEFORE UPDATE ON payment_methods
FOR EACH ROW
EXECUTE FUNCTION touch_updated_at();

COMMENT ON TABLE payment_methods IS 'Organization-owned payment method configuration.';
COMMENT ON COLUMN payment_methods.organization_id IS 'Organization tenant that configures the method.';
COMMENT ON COLUMN payment_methods.code IS 'Payment method code unique within the organization.';

CREATE TABLE store_payment_methods (
    organization_id UUID NOT NULL,
    store_id UUID NOT NULL,
    payment_method_id UUID NOT NULL,
    is_active BOOLEAN DEFAULT TRUE NOT NULL,
    sort_order INTEGER DEFAULT 0 NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    CONSTRAINT store_payment_methods_pkey
        PRIMARY KEY (organization_id, store_id, payment_method_id),
    CONSTRAINT store_payment_methods_store_fkey
        FOREIGN KEY (organization_id, store_id)
        REFERENCES stores (organization_id, id) ON DELETE RESTRICT,
    CONSTRAINT store_payment_methods_method_fkey
        FOREIGN KEY (organization_id, payment_method_id)
        REFERENCES payment_methods (organization_id, id) ON DELETE RESTRICT,
    CONSTRAINT store_payment_methods_sort_order_non_negative
        CHECK (sort_order >= 0),
    CONSTRAINT store_payment_methods_updated_at_check
        CHECK (updated_at >= created_at)
);

CREATE TRIGGER trg_store_payment_methods_touch_updated_at
BEFORE UPDATE ON store_payment_methods
FOR EACH ROW
EXECUTE FUNCTION touch_updated_at();

COMMENT ON TABLE store_payment_methods IS 'Enables and orders organization payment methods for one store.';

CREATE TABLE payments (
    id UUID DEFAULT uuidv7() NOT NULL,
    organization_id UUID NOT NULL,
    store_id UUID NOT NULL,
    sale_id UUID NOT NULL,
    payment_method_id UUID NOT NULL,
    idempotency_key VARCHAR(255) NOT NULL,
    status payment_status DEFAULT 'PENDING' NOT NULL,
    amount NUMERIC(15, 2) NOT NULL,
    received_amount NUMERIC(15, 2),
    change_amount NUMERIC(15, 2),
    installments SMALLINT DEFAULT 1 NOT NULL,
    external_reference VARCHAR(255),
    paid_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    CONSTRAINT payments_pkey PRIMARY KEY (id),
    CONSTRAINT payments_organization_store_id_unique
        UNIQUE (organization_id, store_id, id),
    CONSTRAINT payments_organization_store_idempotency_unique
        UNIQUE (organization_id, store_id, idempotency_key),
    CONSTRAINT payments_sale_fk
        FOREIGN KEY (organization_id, store_id, sale_id)
        REFERENCES sales (organization_id, store_id, id) ON DELETE RESTRICT,
    CONSTRAINT payments_method_fk
        FOREIGN KEY (organization_id, payment_method_id)
        REFERENCES payment_methods (organization_id, id) ON DELETE RESTRICT,
    CONSTRAINT payments_store_method_fkey
        FOREIGN KEY (organization_id, store_id, payment_method_id)
        REFERENCES store_payment_methods (
            organization_id,
            store_id,
            payment_method_id
        ) ON DELETE RESTRICT,
    CONSTRAINT payments_idempotency_key_not_blank
        CHECK (BTRIM(idempotency_key) <> ''),
    CONSTRAINT payments_amount_positive CHECK (amount > 0),
    CONSTRAINT payments_received_amount_non_negative CHECK (
        received_amount IS NULL OR received_amount >= 0
    ),
    CONSTRAINT payments_change_amount_non_negative CHECK (
        change_amount IS NULL OR change_amount >= 0
    ),
    CONSTRAINT payments_installments_positive CHECK (installments >= 1),
    CONSTRAINT payments_external_reference_not_blank CHECK (
        external_reference IS NULL OR BTRIM(external_reference) <> ''
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
        OR (status = 'CANCELLED' AND cancelled_at IS NOT NULL)
    ),
    CONSTRAINT payments_paid_at_check CHECK (
        paid_at IS NULL OR paid_at >= created_at
    ),
    CONSTRAINT payments_cancelled_at_check CHECK (
        cancelled_at IS NULL OR cancelled_at >= created_at
    ),
    CONSTRAINT payments_updated_at_check CHECK (updated_at >= created_at)
);

CREATE UNIQUE INDEX idx_payments_method_external_reference
ON payments (organization_id, payment_method_id, external_reference)
WHERE external_reference IS NOT NULL;

CREATE INDEX idx_payments_sale_id
ON payments (organization_id, store_id, sale_id);

CREATE INDEX idx_payments_status_created_at
ON payments (organization_id, store_id, status, created_at DESC);

CREATE INDEX idx_payments_method_id
ON payments (organization_id, payment_method_id);

CREATE TRIGGER trg_payments_touch_updated_at
BEFORE UPDATE ON payments
FOR EACH ROW
EXECUTE FUNCTION touch_updated_at();

CREATE TRIGGER trg_payments_validate_status_transition
BEFORE UPDATE OF status ON payments
FOR EACH ROW
EXECUTE FUNCTION validate_status_transition();

COMMENT ON TABLE payments IS 'Store-scoped payment transactions associated with POS sales.';
COMMENT ON COLUMN payments.idempotency_key IS 'Request idempotency key unique within one store.';
COMMENT ON COLUMN payments.external_reference IS 'Provider reference unique within the organization payment method.';
