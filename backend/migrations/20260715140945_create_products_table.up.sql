-- Migration: create_products
-- Purpose:
--   Creates the product catalog used by inventory and sales.
--
-- Business rules:
--   - SKU uniquely identifies a product internally.
--   - Barcode is optional, but must be unique when provided.
--   - Monetary values use exact decimal representation.
--   - Products are deactivated instead of physically deleted.
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT uuidv7 (),
    sku VARCHAR(50) NOT NULL,
    barcode VARCHAR(50),
    name VARCHAR(150) NOT NULL,
    price NUMERIC(15, 2) NOT NULL,
    cost NUMERIC(15, 2),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT products_sku_unique UNIQUE (sku),
    CONSTRAINT products_barcode_unique UNIQUE (barcode),
    CONSTRAINT products_sku_not_blank CHECK (BTRIM (sku) <> ''),
    CONSTRAINT products_barcode_not_blank CHECK (
        barcode IS NULL
        OR BTRIM (barcode) <> ''
    ),
    CONSTRAINT products_name_not_blank CHECK (BTRIM (name) <> ''),
    CONSTRAINT products_price_non_negative CHECK (price >= 0),
    CONSTRAINT products_cost_non_negative CHECK (
        cost IS NULL
        OR cost >= 0
    ),
    CONSTRAINT products_updated_at_check CHECK (updated_at >= created_at)
);

CREATE INDEX idx_products_active_name ON products (name)
WHERE
    is_active = TRUE;

CREATE INDEX idx_products_name_trgm ON products USING GIN (name gin_trgm_ops);

COMMENT ON
TABLE products IS 'Stores products available for inventory control and POS sales.';

COMMENT ON COLUMN products.id IS 'Internal UUIDv7 identifier.';

COMMENT ON COLUMN products.sku IS 'Unique internal stock keeping unit.';

COMMENT ON COLUMN products.barcode IS 'Optional unique barcode used by barcode scanners.';

COMMENT ON COLUMN products.name IS 'Product name displayed in administrative and POS interfaces.';

COMMENT ON COLUMN products.price IS 'Current product sale price. Historical prices are preserved in sale_items.';

COMMENT ON COLUMN products.cost IS 'Optional product acquisition cost used for margin calculations.';

COMMENT ON COLUMN products.is_active IS 'Controls whether the product can be used in new sales.';
