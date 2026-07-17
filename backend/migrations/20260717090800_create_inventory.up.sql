CREATE TYPE inventory_movement_type AS ENUM (
    'PURCHASE',
    'SALE',
    'SALE_CANCELLATION',
    'ADJUSTMENT_IN',
    'ADJUSTMENT_OUT'
);

CREATE TABLE inventory (
    organization_id UUID NOT NULL,
    store_id UUID NOT NULL,
    product_id UUID NOT NULL,
    quantity NUMERIC(15, 3) DEFAULT 0 NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    CONSTRAINT inventory_pkey PRIMARY KEY (
        organization_id,
        store_id,
        product_id
    ),
    CONSTRAINT inventory_store_fkey FOREIGN KEY (organization_id, store_id) REFERENCES stores (organization_id, id) ON DELETE RESTRICT,
    CONSTRAINT inventory_product_fkey FOREIGN KEY (organization_id, product_id) REFERENCES products (organization_id, id) ON DELETE RESTRICT,
    CONSTRAINT inventory_quantity_non_negative CHECK (quantity >= 0),
    CONSTRAINT inventory_updated_at_check CHECK (updated_at >= created_at)
);

CREATE TRIGGER trg_inventory_touch_updated_at
BEFORE UPDATE ON inventory
FOR EACH ROW
EXECUTE FUNCTION touch_updated_at();

COMMENT ON
TABLE inventory IS 'Current product quantity for one organization store.';

COMMENT ON COLUMN inventory.organization_id IS 'Organization tenant shared by the store and product.';

COMMENT ON COLUMN inventory.store_id IS 'Store whose stock is represented.';

COMMENT ON COLUMN inventory.product_id IS 'Organization product held by the store.';

CREATE TABLE inventory_movements (
    id UUID DEFAULT uuidv7 () NOT NULL,
    organization_id UUID NOT NULL,
    store_id UUID NOT NULL,
    product_id UUID NOT NULL,
    actor_membership_id UUID NOT NULL,
    movement_type inventory_movement_type NOT NULL,
    quantity NUMERIC(15, 3) NOT NULL,
    previous_quantity NUMERIC(15, 3) NOT NULL,
    current_quantity NUMERIC(15, 3) NOT NULL,
    reason TEXT,
    reference_type VARCHAR(30) NOT NULL,
    reference_id UUID NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    CONSTRAINT inventory_movements_pkey PRIMARY KEY (id),
    CONSTRAINT inventory_movements_store_fkey FOREIGN KEY (organization_id, store_id) REFERENCES stores (organization_id, id) ON DELETE RESTRICT,
    CONSTRAINT inventory_movements_product_fkey FOREIGN KEY (organization_id, product_id) REFERENCES products (organization_id, id) ON DELETE RESTRICT,
    CONSTRAINT inventory_movements_actor_membership_fkey FOREIGN KEY (
        organization_id,
        actor_membership_id
    ) REFERENCES organization_memberships (organization_id, id) ON DELETE RESTRICT,
    CONSTRAINT inventory_movements_quantity_positive CHECK (quantity > 0),
    CONSTRAINT inventory_movements_reference_not_blank CHECK (BTRIM (reference_type) <> ''),
    CONSTRAINT inventory_movements_reference_unique UNIQUE (
        organization_id,
        store_id,
        product_id,
        movement_type,
        reference_type,
        reference_id
    )
);

CREATE INDEX idx_inventory_movements_product ON inventory_movements (
    organization_id,
    store_id,
    product_id,
    created_at DESC
);

CREATE INDEX idx_inventory_movements_reference ON inventory_movements (
    organization_id,
    store_id,
    reference_type,
    reference_id
);

COMMENT ON
TABLE inventory_movements IS 'Append-style inventory changes scoped to one organization and store.';

COMMENT ON COLUMN inventory_movements.actor_membership_id IS 'Organization membership that executed the stock operation.';

COMMENT ON COLUMN inventory_movements.reference_id IS 'Idempotency reference interpreted together with tenant, store, product, type, and reference type.';