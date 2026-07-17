-- Migration: create_categories
-- Purpose: Stores the product categories used by the catalog and POS.
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT uuidv7 (),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(120) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT categories_name_unique UNIQUE (name),
    CONSTRAINT categories_slug_unique UNIQUE (slug),
    CONSTRAINT categories_name_not_blank CHECK (BTRIM (name) <> ''),
    CONSTRAINT categories_slug_not_blank CHECK (BTRIM (slug) <> ''),
    CONSTRAINT categories_updated_at_check CHECK (updated_at >= created_at)
);

CREATE INDEX idx_categories_active_name ON categories (name)
WHERE
    is_active = TRUE;

COMMENT ON TABLE categories IS 'Product categories used to organize the catalog.';
