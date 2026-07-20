CREATE TYPE organization_status AS ENUM (
    'ACTIVE',
    'SUSPENDED',
    'ARCHIVED'
);

CREATE TYPE store_status AS ENUM (
    'ACTIVE',
    'INACTIVE',
    'ARCHIVED'
);

CREATE TABLE organizations (
    id UUID DEFAULT uuidv7 () NOT NULL,
    name VARCHAR(150) NOT NULL,
    slug VARCHAR(120) NOT NULL,
    status organization_status DEFAULT 'ACTIVE' NOT NULL,
    timezone VARCHAR(64) DEFAULT 'America/Sao_Paulo' NOT NULL,
    locale VARCHAR(20) DEFAULT 'pt-BR' NOT NULL,
    currency CHAR(3) DEFAULT 'BRL' NOT NULL,
    authorization_version BIGINT DEFAULT 1 NOT NULL,
    created_by_user_id UUID NOT NULL,
    archived_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    CONSTRAINT organizations_pkey PRIMARY KEY (id),
    CONSTRAINT organizations_slug_unique UNIQUE (slug),
    CONSTRAINT organizations_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES users (id) ON DELETE RESTRICT,
    CONSTRAINT organizations_name_not_blank CHECK (BTRIM (name) <> ''),
    CONSTRAINT organizations_slug_not_blank CHECK (BTRIM (slug) <> ''),
    CONSTRAINT organizations_slug_format CHECK (
        slug ~ '^[a-z0-9]+(?:-[a-z0-9]+)*$'
    ),
    CONSTRAINT organizations_timezone_not_blank CHECK (BTRIM (timezone) <> ''),
    CONSTRAINT organizations_locale_not_blank CHECK (BTRIM (locale) <> ''),
    CONSTRAINT organizations_currency_format CHECK (currency ~ '^[A-Z]{3}$'),
    CONSTRAINT organizations_authorization_version_positive CHECK (authorization_version > 0),
    CONSTRAINT organizations_archived_status_check CHECK (
        (
            status = 'ARCHIVED'
            AND archived_at IS NOT NULL
        )
        OR (
            status <> 'ARCHIVED'
            AND archived_at IS NULL
        )
    ),
    CONSTRAINT organizations_archived_at_check CHECK (
        archived_at IS NULL
        OR archived_at >= created_at
    ),
    CONSTRAINT organizations_updated_at_check CHECK (updated_at >= created_at)
);

CREATE INDEX idx_organizations_status ON organizations (status);

CREATE INDEX idx_organizations_created_by_user_id ON organizations (created_by_user_id);

CREATE TRIGGER trg_organizations_touch_updated_at
BEFORE UPDATE ON organizations
FOR EACH ROW
EXECUTE FUNCTION touch_updated_at();

COMMENT ON
TABLE organizations IS 'Stores organizations that form the application tenant boundary.';

COMMENT ON COLUMN organizations.id IS 'Internal UUIDv7 tenant identifier.';

COMMENT ON COLUMN organizations.name IS 'Organization name displayed in administrative interfaces.';

COMMENT ON COLUMN organizations.slug IS 'Globally unique lowercase URL-safe organization slug.';

COMMENT ON COLUMN organizations.status IS 'Lifecycle status of the organization.';

COMMENT ON COLUMN organizations.timezone IS 'Default IANA timezone for the organization.';

COMMENT ON COLUMN organizations.locale IS 'Default locale used for organization-level presentation.';

COMMENT ON COLUMN organizations.currency IS 'ISO-style uppercase three-letter operating currency.';

COMMENT ON COLUMN organizations.authorization_version IS 'Monotonic version used to invalidate organization authorization.';

COMMENT ON COLUMN organizations.created_by_user_id IS 'Global user that initiated organization creation.';

COMMENT ON COLUMN organizations.archived_at IS 'Timestamp required when the organization is archived.';

CREATE TABLE stores (
    id UUID DEFAULT uuidv7 () NOT NULL,
    organization_id UUID NOT NULL,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(150) NOT NULL,
    status store_status DEFAULT 'ACTIVE' NOT NULL,
    timezone VARCHAR(64) NOT NULL,
    created_by_user_id UUID NOT NULL,
    archived_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    CONSTRAINT stores_pkey PRIMARY KEY (id),
    CONSTRAINT stores_organization_id_id_unique UNIQUE (organization_id, id),
    CONSTRAINT stores_organization_id_code_unique UNIQUE (organization_id, code),
    CONSTRAINT stores_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES organizations (id) ON DELETE RESTRICT,
    CONSTRAINT stores_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES users (id) ON DELETE RESTRICT,
    CONSTRAINT stores_code_not_blank CHECK (BTRIM (code) <> ''),
    CONSTRAINT stores_name_not_blank CHECK (BTRIM (name) <> ''),
    CONSTRAINT stores_timezone_not_blank CHECK (BTRIM (timezone) <> ''),
    CONSTRAINT stores_archived_status_check CHECK (
        (
            status = 'ARCHIVED'
            AND archived_at IS NOT NULL
        )
        OR (
            status <> 'ARCHIVED'
            AND archived_at IS NULL
        )
    ),
    CONSTRAINT stores_archived_at_check CHECK (
        archived_at IS NULL
        OR archived_at >= created_at
    ),
    CONSTRAINT stores_updated_at_check CHECK (updated_at >= created_at)
);

CREATE INDEX idx_stores_organization_status_name ON stores (organization_id, status, name);

CREATE TRIGGER trg_stores_touch_updated_at
BEFORE UPDATE ON stores
FOR EACH ROW
EXECUTE FUNCTION touch_updated_at();

COMMENT ON
TABLE stores IS 'Stores points of sale owned by an organization tenant.';

COMMENT ON COLUMN stores.id IS 'Internal UUIDv7 store identifier.';

COMMENT ON COLUMN stores.organization_id IS 'Organization tenant that owns the store.';

COMMENT ON COLUMN stores.code IS 'Application-normalized uppercase code unique within the organization.';

COMMENT ON COLUMN stores.name IS 'Store name displayed in administrative and POS interfaces.';

COMMENT ON COLUMN stores.status IS 'Lifecycle status of the store.';

COMMENT ON COLUMN stores.timezone IS 'IANA timezone used by the store.';

COMMENT ON COLUMN stores.created_by_user_id IS 'Global user that initiated store creation.';

COMMENT ON COLUMN stores.archived_at IS 'Timestamp required when the store is archived.';