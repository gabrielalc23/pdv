CREATE TABLE categories (
    id UUID DEFAULT uuidv7 () NOT NULL,
    organization_id UUID NOT NULL,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(120) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    CONSTRAINT categories_pkey PRIMARY KEY (id),
    CONSTRAINT categories_organization_id_id_unique UNIQUE (organization_id, id),
    CONSTRAINT categories_organization_id_name_unique UNIQUE (organization_id, name),
    CONSTRAINT categories_organization_id_slug_unique UNIQUE (organization_id, slug),
    CONSTRAINT categories_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES organizations (id) ON DELETE RESTRICT,
    CONSTRAINT categories_name_not_blank CHECK (BTRIM (name) <> ''),
    CONSTRAINT categories_slug_not_blank CHECK (BTRIM (slug) <> ''),
    CONSTRAINT categories_updated_at_check CHECK (updated_at >= created_at)
);

CREATE INDEX idx_categories_active_name ON categories (organization_id, name)
WHERE
    is_active = TRUE;

CREATE TRIGGER trg_categories_touch_updated_at
BEFORE UPDATE ON categories
FOR EACH ROW
EXECUTE FUNCTION touch_updated_at();

COMMENT ON
TABLE categories IS 'Organization-owned product catalog categories.';

COMMENT ON COLUMN categories.organization_id IS 'Organization tenant that owns the category.';

COMMENT ON COLUMN categories.slug IS 'Category slug unique within the organization.';

CREATE TABLE products (
    id UUID DEFAULT uuidv7 () NOT NULL,
    organization_id UUID NOT NULL,
    sku VARCHAR(50) NOT NULL,
    barcode VARCHAR(50),
    name VARCHAR(150) NOT NULL,
    category_id UUID,
    price NUMERIC(15, 2) NOT NULL,
    cost NUMERIC(15, 2),
    is_active BOOLEAN DEFAULT TRUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    CONSTRAINT products_pkey PRIMARY KEY (id),
    CONSTRAINT products_organization_id_id_unique UNIQUE (organization_id, id),
    CONSTRAINT products_organization_id_sku_unique UNIQUE (organization_id, sku),
    CONSTRAINT products_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES organizations (id) ON DELETE RESTRICT,
    CONSTRAINT products_category_fkey FOREIGN KEY (organization_id, category_id) REFERENCES categories (organization_id, id) ON DELETE SET NULL (category_id),
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

CREATE UNIQUE INDEX idx_products_organization_barcode_unique ON products (organization_id, barcode)
WHERE
    barcode IS NOT NULL;

CREATE INDEX idx_products_active_name ON products (organization_id, name)
WHERE
    is_active = TRUE;

CREATE INDEX idx_products_category_id ON products (organization_id, category_id);

CREATE INDEX idx_products_name_trgm ON products USING GIN (name gin_trgm_ops);

CREATE TRIGGER trg_products_touch_updated_at
BEFORE UPDATE ON products
FOR EACH ROW
EXECUTE FUNCTION touch_updated_at();

COMMENT ON
TABLE products IS 'Organization-owned products available to stores in the tenant.';

COMMENT ON COLUMN products.organization_id IS 'Organization tenant that owns the product.';

COMMENT ON COLUMN products.sku IS 'Stock keeping unit unique within the organization.';

COMMENT ON COLUMN products.barcode IS 'Optional barcode unique within the organization.';

COMMENT ON COLUMN products.category_id IS 'Optional category from the same organization.';

COMMENT ON COLUMN products.price IS 'Current sale price; historical values are preserved in sale items.';