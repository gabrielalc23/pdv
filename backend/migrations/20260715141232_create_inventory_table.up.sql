CREATE TYPE inventory_movement_type AS ENUM (
    'PURCHASE',
    'SALE',
    'SALE_CANCELLATION',
    'ADJUSTMENT_IN',
    'ADJUSTMENT_OUT'
);

CREATE TABLE inventory (
    product_id UUID PRIMARY KEY REFERENCES products (id),
    quantity NUMERIC(15, 3) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT inventory_quantity_non_negative CHECK (quantity >= 0),
    CONSTRAINT inventory_updated_at_check CHECK (updated_at >= created_at)
);

CREATE TABLE inventory_movements (
    id UUID PRIMARY KEY DEFAULT uuidv7 (),
    product_id UUID NOT NULL REFERENCES products (id),
    movement_type inventory_movement_type NOT NULL,
    quantity NUMERIC(15, 3) NOT NULL,
    previous_quantity NUMERIC(15, 3) NOT NULL,
    current_quantity NUMERIC(15, 3) NOT NULL,
    reason TEXT,
    reference_type VARCHAR(30) NOT NULL,
    reference_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT inventory_movements_quantity_positive CHECK (quantity > 0),
    CONSTRAINT inventory_movements_reference_not_blank CHECK (BTRIM(reference_type) <> ''),
    CONSTRAINT inventory_movements_reference_unique UNIQUE (
        product_id,
        movement_type,
        reference_type,
        reference_id
    )
);

CREATE INDEX idx_inventory_movements_product ON inventory_movements (product_id, created_at DESC);

CREATE INDEX idx_inventory_movements_reference ON inventory_movements (reference_type, reference_id);
