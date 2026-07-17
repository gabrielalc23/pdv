CREATE TYPE sale_status AS ENUM (
    'OPEN',
    'COMPLETED',
    'CANCELLED'
);

CREATE FUNCTION validate_status_transition()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    old_status TEXT := OLD.status::TEXT;
    new_status TEXT := NEW.status::TEXT;
BEGIN
    IF old_status = new_status THEN
        RETURN NEW;
    END IF;

    IF TG_TABLE_NAME = 'sales' THEN
        IF old_status = 'OPEN' AND new_status IN ('COMPLETED', 'CANCELLED') THEN
            RETURN NEW;
        END IF;

        IF old_status = 'COMPLETED' AND new_status = 'CANCELLED' THEN
            RETURN NEW;
        END IF;

        RAISE EXCEPTION 'invalid sale status transition: % -> %', old_status, new_status
            USING ERRCODE = '23514',
                  CONSTRAINT = 'sales_status_transition_check';
    ELSIF TG_TABLE_NAME = 'payments' THEN
        IF old_status = 'PENDING' AND new_status IN ('APPROVED', 'DECLINED', 'CANCELLED') THEN
            RETURN NEW;
        END IF;

        IF old_status = 'APPROVED' AND new_status = 'CANCELLED' THEN
            RETURN NEW;
        END IF;

        RAISE EXCEPTION 'invalid payment status transition: % -> %', old_status, new_status
            USING ERRCODE = '23514',
                  CONSTRAINT = 'payments_status_transition_check';
    ELSIF TG_TABLE_NAME = 'fiscal_documents' THEN
        IF old_status = 'PENDING' AND new_status IN ('PROCESSING', 'AUTHORIZED', 'REJECTED', 'ERROR', 'CANCELLED') THEN
            RETURN NEW;
        END IF;

        IF old_status = 'PROCESSING' AND new_status IN ('AUTHORIZED', 'REJECTED', 'ERROR', 'CANCELLED') THEN
            RETURN NEW;
        END IF;

        IF old_status = 'AUTHORIZED' AND new_status = 'CANCELLED' THEN
            RETURN NEW;
        END IF;

        IF old_status = 'REJECTED' AND new_status IN ('PROCESSING', 'ERROR', 'CANCELLED') THEN
            RETURN NEW;
        END IF;

        IF old_status = 'ERROR' AND new_status IN ('PROCESSING', 'REJECTED', 'CANCELLED') THEN
            RETURN NEW;
        END IF;

        RAISE EXCEPTION 'invalid fiscal document status transition: % -> %', old_status, new_status
            USING ERRCODE = '23514',
                  CONSTRAINT = 'fiscal_documents_status_transition_check';
    END IF;

    RETURN NEW;
END;
$$;

CREATE TABLE sale_number_counters (
    organization_id UUID NOT NULL,
    store_id UUID NOT NULL,
    next_number BIGINT DEFAULT 1 NOT NULL,
    CONSTRAINT sale_number_counters_pkey PRIMARY KEY (organization_id, store_id),
    CONSTRAINT sale_number_counters_store_fkey FOREIGN KEY (organization_id, store_id) REFERENCES stores (organization_id, id) ON DELETE RESTRICT,
    CONSTRAINT sale_number_counters_next_number_positive CHECK (next_number > 0)
);

COMMENT ON
TABLE sale_number_counters IS 'Per-store counters used to allocate sale numbers transactionally.';

CREATE TABLE sales (
    id UUID DEFAULT uuidv7 () NOT NULL,
    organization_id UUID NOT NULL,
    store_id UUID NOT NULL,
    number BIGINT NOT NULL,
    idempotency_key VARCHAR(255) NOT NULL,
    status sale_status DEFAULT 'OPEN' NOT NULL,
    subtotal NUMERIC(15, 2) DEFAULT 0 NOT NULL,
    discount NUMERIC(15, 2) DEFAULT 0 NOT NULL,
    addition NUMERIC(15, 2) DEFAULT 0 NOT NULL,
    total NUMERIC(15, 2) DEFAULT 0 NOT NULL,
    opened_by_membership_id UUID NOT NULL,
    completed_by_membership_id UUID,
    cancelled_by_membership_id UUID,
    opened_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    completed_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    CONSTRAINT sales_pkey PRIMARY KEY (id),
    CONSTRAINT sales_organization_store_id_unique UNIQUE (organization_id, store_id, id),
    CONSTRAINT sales_organization_store_number_unique UNIQUE (
        organization_id,
        store_id,
        number
    ),
    CONSTRAINT sales_organization_store_idempotency_unique UNIQUE (
        organization_id,
        store_id,
        idempotency_key
    ),
    CONSTRAINT sales_store_fkey FOREIGN KEY (organization_id, store_id) REFERENCES stores (organization_id, id) ON DELETE RESTRICT,
    CONSTRAINT sales_opened_by_membership_fkey FOREIGN KEY (
        organization_id,
        opened_by_membership_id
    ) REFERENCES organization_memberships (organization_id, id) ON DELETE RESTRICT,
    CONSTRAINT sales_completed_by_membership_fkey FOREIGN KEY (
        organization_id,
        completed_by_membership_id
    ) REFERENCES organization_memberships (organization_id, id) ON DELETE RESTRICT,
    CONSTRAINT sales_cancelled_by_membership_fkey FOREIGN KEY (
        organization_id,
        cancelled_by_membership_id
    ) REFERENCES organization_memberships (organization_id, id) ON DELETE RESTRICT,
    CONSTRAINT sales_number_positive CHECK (number > 0),
    CONSTRAINT sales_idempotency_key_not_blank CHECK (BTRIM (idempotency_key) <> ''),
    CONSTRAINT sales_subtotal_non_negative CHECK (subtotal >= 0),
    CONSTRAINT sales_discount_non_negative CHECK (discount >= 0),
    CONSTRAINT sales_addition_non_negative CHECK (addition >= 0),
    CONSTRAINT sales_total_non_negative CHECK (total >= 0),
    CONSTRAINT sales_discount_not_greater_than_subtotal CHECK (discount <= subtotal),
    CONSTRAINT sales_status_timestamps_consistency CHECK (
        (
            status = 'OPEN'
            AND completed_at IS NULL
            AND completed_by_membership_id IS NULL
            AND cancelled_at IS NULL
            AND cancelled_by_membership_id IS NULL
        )
        OR (
            status = 'COMPLETED'
            AND completed_at IS NOT NULL
            AND completed_by_membership_id IS NOT NULL
            AND cancelled_at IS NULL
            AND cancelled_by_membership_id IS NULL
        )
        OR (
            status = 'CANCELLED'
            AND cancelled_at IS NOT NULL
            AND cancelled_by_membership_id IS NOT NULL
            AND (
                (
                    completed_at IS NULL
                    AND completed_by_membership_id IS NULL
                )
                OR (
                    completed_at IS NOT NULL
                    AND completed_by_membership_id IS NOT NULL
                )
            )
        )
    ),
    CONSTRAINT sales_total_consistency CHECK (
        total = subtotal - discount + addition
    ),
    CONSTRAINT sales_opened_at_check CHECK (opened_at >= created_at),
    CONSTRAINT sales_completed_at_check CHECK (
        completed_at IS NULL
        OR completed_at >= opened_at
    ),
    CONSTRAINT sales_cancelled_at_check CHECK (
        cancelled_at IS NULL
        OR cancelled_at >= opened_at
    ),
    CONSTRAINT sales_updated_at_check CHECK (updated_at >= created_at)
);

CREATE INDEX idx_sales_opened_at ON sales (
    organization_id,
    store_id,
    opened_at DESC
);

CREATE INDEX idx_sales_status_opened_at ON sales (
    organization_id,
    store_id,
    status,
    opened_at DESC
);

CREATE TRIGGER trg_sales_touch_updated_at
BEFORE UPDATE ON sales
FOR EACH ROW
EXECUTE FUNCTION touch_updated_at();

CREATE TRIGGER trg_sales_validate_status_transition
BEFORE UPDATE OF status ON sales
FOR EACH ROW
EXECUTE FUNCTION validate_status_transition();

COMMENT ON
TABLE sales IS 'Store-scoped POS sale headers and lifecycle actors.';

COMMENT ON COLUMN sales.number IS 'Human-readable sequential number allocated independently per store.';

COMMENT ON COLUMN sales.idempotency_key IS 'Request idempotency key unique within one store.';

COMMENT ON COLUMN sales.opened_by_membership_id IS 'Organization membership that opened the sale.';

CREATE TABLE sale_items (
    id UUID DEFAULT uuidv7 () NOT NULL,
    organization_id UUID NOT NULL,
    store_id UUID NOT NULL,
    sale_id UUID NOT NULL,
    product_id UUID NOT NULL,
    product_name VARCHAR(150) NOT NULL,
    product_sku VARCHAR(50) NOT NULL,
    unit_price NUMERIC(15, 2) NOT NULL,
    quantity NUMERIC(15, 3) NOT NULL,
    discount NUMERIC(15, 2) DEFAULT 0 NOT NULL,
    total NUMERIC(15, 2) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    CONSTRAINT sale_items_pkey PRIMARY KEY (id),
    CONSTRAINT sale_items_sale_fk FOREIGN KEY (
        organization_id,
        store_id,
        sale_id
    ) REFERENCES sales (organization_id, store_id, id) ON DELETE CASCADE,
    CONSTRAINT sale_items_product_fk FOREIGN KEY (organization_id, product_id) REFERENCES products (organization_id, id) ON DELETE RESTRICT,
    CONSTRAINT sale_items_product_name_not_blank CHECK (BTRIM (product_name) <> ''),
    CONSTRAINT sale_items_product_sku_not_blank CHECK (BTRIM (product_sku) <> ''),
    CONSTRAINT sale_items_unit_price_non_negative CHECK (unit_price >= 0),
    CONSTRAINT sale_items_quantity_positive CHECK (quantity > 0),
    CONSTRAINT sale_items_discount_non_negative CHECK (discount >= 0),
    CONSTRAINT sale_items_total_non_negative CHECK (total >= 0),
    CONSTRAINT sale_items_discount_not_greater_than_gross CHECK (
        discount <= ROUND(unit_price * quantity, 2)
    ),
    CONSTRAINT sale_items_total_consistency CHECK (
        total = ROUND(
            (unit_price * quantity) - discount,
            2
        )
    )
);

CREATE INDEX idx_sale_items_sale_id ON sale_items (
    organization_id,
    store_id,
    sale_id
);

CREATE INDEX idx_sale_items_product_id ON sale_items (
    organization_id,
    store_id,
    product_id
);

COMMENT ON
TABLE sale_items IS 'Tenant- and store-scoped product snapshots for sale items.';

COMMENT ON COLUMN sale_items.product_name IS 'Product name snapshot at the moment of sale.';

COMMENT ON COLUMN sale_items.product_sku IS 'Product SKU snapshot at the moment of sale.';