-- Migration: create_sales
-- Purpose:
--   Creates the sales and sale_items tables used by the POS checkout flow.
--
-- Business rules:
--   - Every sale starts with the OPEN status.
--   - A sale may contain multiple items.
--   - Product data is stored as a snapshot in sale_items.
--   - Monetary values use exact decimal representation.
--   - Item totals are rounded to two decimal places.
--   - Quantities support up to three decimal places.
--
-- Dependencies:
--   - products
CREATE TYPE sale_status AS ENUM ('OPEN', 'COMPLETED', 'CANCELLED');

CREATE TABLE sales (
    id UUID PRIMARY KEY DEFAULT uuidv7 (),
    number BIGSERIAL NOT NULL UNIQUE,
    idempotency_key VARCHAR(255) NOT NULL UNIQUE,
    status sale_status NOT NULL DEFAULT 'OPEN',
    subtotal NUMERIC(15, 2) NOT NULL DEFAULT 0,
    discount NUMERIC(15, 2) NOT NULL DEFAULT 0,
    addition NUMERIC(15, 2) NOT NULL DEFAULT 0,
    total NUMERIC(15, 2) NOT NULL DEFAULT 0,
    opened_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT sales_subtotal_non_negative CHECK (subtotal >= 0),
    CONSTRAINT sales_discount_non_negative CHECK (discount >= 0),
    CONSTRAINT sales_addition_non_negative CHECK (addition >= 0),
    CONSTRAINT sales_total_non_negative CHECK (total >= 0),
    CONSTRAINT sales_discount_not_greater_than_subtotal CHECK (discount <= subtotal),
    CONSTRAINT sales_status_timestamps_consistency CHECK (
        (
            status = 'OPEN'
            AND completed_at IS NULL
            AND cancelled_at IS NULL
        )
        OR (
            status = 'COMPLETED'
            AND completed_at IS NOT NULL
            AND cancelled_at IS NULL
        )
        OR (
            status = 'CANCELLED'
            AND cancelled_at IS NOT NULL
        )
    ),
    CONSTRAINT sales_total_consistency CHECK (
        total = subtotal - discount + addition
    ),
    CONSTRAINT sales_updated_at_check CHECK (updated_at >= created_at)
);

CREATE TABLE sale_items (
    id UUID PRIMARY KEY DEFAULT uuidv7 (),
    sale_id UUID NOT NULL,
    product_id UUID NOT NULL,
    product_name VARCHAR(150) NOT NULL,
    product_sku VARCHAR(50) NOT NULL,
    unit_price NUMERIC(15, 2) NOT NULL,
    quantity NUMERIC(15, 3) NOT NULL,
    discount NUMERIC(15, 2) NOT NULL DEFAULT 0,
    total NUMERIC(15, 2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT sale_items_sale_fk FOREIGN KEY (sale_id) REFERENCES sales (id) ON DELETE CASCADE,
    CONSTRAINT sale_items_product_fk FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE RESTRICT,
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

CREATE INDEX idx_sales_opened_at ON sales (opened_at DESC);

CREATE INDEX idx_sales_status_opened_at ON sales (status, opened_at DESC);

CREATE INDEX idx_sale_items_sale_id ON sale_items (sale_id);

CREATE INDEX idx_sale_items_product_id ON sale_items (product_id);

COMMENT ON TYPE sale_status IS 'Lifecycle status of a POS sale.';

COMMENT ON
TABLE sales IS 'Stores POS sale headers, monetary totals and lifecycle timestamps.';

COMMENT ON COLUMN sales.number IS 'Sequential human-readable number displayed in the POS and receipts.';

COMMENT ON COLUMN sales.subtotal IS 'Sum of sale item totals before sale-level discount and addition.';

COMMENT ON COLUMN sales.discount IS 'Discount applied to the entire sale.';

COMMENT ON COLUMN sales.addition IS 'Additional amount applied to the sale.';

COMMENT ON COLUMN sales.total IS 'Final amount calculated as subtotal minus discount plus addition.';

COMMENT ON
TABLE sale_items IS 'Stores product snapshots and monetary totals for each sale item.';

COMMENT ON COLUMN sale_items.product_name IS 'Product name snapshot preserved at the moment of sale.';

COMMENT ON COLUMN sale_items.product_sku IS 'Product SKU snapshot preserved at the moment of sale.';

COMMENT ON COLUMN sale_items.unit_price IS 'Unit price snapshot used at the moment of sale.';

COMMENT ON COLUMN sale_items.quantity IS 'Quantity sold with support for up to three decimal places.';

COMMENT ON COLUMN sale_items.total IS 'Final item total rounded to two decimal places after discount.';