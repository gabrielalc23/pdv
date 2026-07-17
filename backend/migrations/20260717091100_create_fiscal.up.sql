CREATE TYPE fiscal_document_status AS ENUM (
    'PENDING',
    'PROCESSING',
    'AUTHORIZED',
    'REJECTED',
    'ERROR',
    'CANCELLED'
);

CREATE TYPE fiscal_environment AS ENUM (
    'HOMOLOGATION',
    'PRODUCTION'
);

CREATE TABLE fiscal_documents (
    id UUID DEFAULT uuidv7 () NOT NULL,
    organization_id UUID NOT NULL,
    store_id UUID NOT NULL,
    sale_id UUID NOT NULL,
    status fiscal_document_status DEFAULT 'PENDING' NOT NULL,
    environment fiscal_environment DEFAULT 'HOMOLOGATION' NOT NULL,
    document_model SMALLINT DEFAULT 65 NOT NULL,
    series INTEGER,
    number BIGINT,
    access_key VARCHAR(44),
    protocol VARCHAR(100),
    provider VARCHAR(100),
    external_reference VARCHAR(255),
    xml TEXT,
    error_code VARCHAR(50),
    error_message TEXT,
    issued_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    CONSTRAINT fiscal_documents_pkey PRIMARY KEY (id),
    CONSTRAINT fiscal_documents_organization_store_id_unique UNIQUE (organization_id, store_id, id),
    CONSTRAINT fiscal_documents_sale_unique UNIQUE (
        organization_id,
        store_id,
        sale_id
    ),
    CONSTRAINT fiscal_documents_sale_fk FOREIGN KEY (
        organization_id,
        store_id,
        sale_id
    ) REFERENCES sales (organization_id, store_id, id) ON DELETE RESTRICT,
    CONSTRAINT fiscal_documents_model_nfce CHECK (document_model = 65),
    CONSTRAINT fiscal_documents_series_positive CHECK (
        series IS NULL
        OR series > 0
    ),
    CONSTRAINT fiscal_documents_number_positive CHECK (
        number IS NULL
        OR number > 0
    ),
    CONSTRAINT fiscal_documents_access_key_length CHECK (
        access_key IS NULL
        OR LENGTH(access_key) = 44
    ),
    CONSTRAINT fiscal_documents_protocol_not_blank CHECK (
        protocol IS NULL
        OR BTRIM (protocol) <> ''
    ),
    CONSTRAINT fiscal_documents_provider_not_blank CHECK (
        provider IS NULL
        OR BTRIM (provider) <> ''
    ),
    CONSTRAINT fiscal_documents_external_reference_not_blank CHECK (
        external_reference IS NULL
        OR BTRIM (external_reference) <> ''
    ),
    CONSTRAINT fiscal_documents_error_code_not_blank CHECK (
        error_code IS NULL
        OR BTRIM (error_code) <> ''
    ),
    CONSTRAINT fiscal_documents_error_message_not_blank CHECK (
        error_message IS NULL
        OR BTRIM (error_message) <> ''
    ),
    CONSTRAINT fiscal_documents_authorization_consistency CHECK (
        (
            status = 'PENDING'
            AND access_key IS NULL
            AND protocol IS NULL
            AND issued_at IS NULL
            AND cancelled_at IS NULL
        )
        OR (
            status = 'PROCESSING'
            AND cancelled_at IS NULL
        )
        OR (
            status = 'AUTHORIZED'
            AND access_key IS NOT NULL
            AND protocol IS NOT NULL
            AND issued_at IS NOT NULL
            AND cancelled_at IS NULL
        )
        OR (
            status = 'REJECTED'
            AND cancelled_at IS NULL
        )
        OR (
            status = 'ERROR'
            AND cancelled_at IS NULL
        )
        OR (
            status = 'CANCELLED'
            AND access_key IS NOT NULL
            AND protocol IS NOT NULL
            AND issued_at IS NOT NULL
            AND cancelled_at IS NOT NULL
        )
    ),
    CONSTRAINT fiscal_documents_cancellation_consistency CHECK (
        (
            status = 'CANCELLED'
            AND cancelled_at IS NOT NULL
        )
        OR (
            status <> 'CANCELLED'
            AND cancelled_at IS NULL
        )
    ),
    CONSTRAINT fiscal_documents_issued_at_check CHECK (
        issued_at IS NULL
        OR issued_at >= created_at
    ),
    CONSTRAINT fiscal_documents_cancelled_at_check CHECK (
        cancelled_at IS NULL
        OR cancelled_at >= created_at
    ),
    CONSTRAINT fiscal_documents_updated_at_check CHECK (updated_at >= created_at)
);

CREATE UNIQUE INDEX idx_fiscal_documents_access_key ON fiscal_documents (access_key)
WHERE
    access_key IS NOT NULL;

CREATE UNIQUE INDEX idx_fiscal_documents_provider_reference ON fiscal_documents (provider, external_reference)
WHERE
    provider IS NOT NULL
    AND external_reference IS NOT NULL;

CREATE INDEX idx_fiscal_documents_status_created_at ON fiscal_documents (
    organization_id,
    store_id,
    status,
    created_at DESC
);

CREATE TRIGGER trg_fiscal_documents_touch_updated_at
BEFORE UPDATE ON fiscal_documents
FOR EACH ROW
EXECUTE FUNCTION touch_updated_at();

CREATE TRIGGER trg_fiscal_documents_validate_status_transition
BEFORE UPDATE OF status ON fiscal_documents
FOR EACH ROW
EXECUTE FUNCTION validate_status_transition();

COMMENT ON
TABLE fiscal_documents IS 'Store-scoped fiscal documents associated with POS sales.';

COMMENT ON COLUMN fiscal_documents.access_key IS 'Globally unique fiscal access key when authorized.';

COMMENT ON COLUMN fiscal_documents.external_reference IS 'Provider transaction identifier.';

CREATE TABLE fiscal_document_callbacks (
    id UUID DEFAULT uuidv7 () NOT NULL,
    organization_id UUID NOT NULL,
    store_id UUID NOT NULL,
    fiscal_document_id UUID NOT NULL,
    provider VARCHAR(100) NOT NULL,
    callback_key VARCHAR(255) NOT NULL,
    payload TEXT NOT NULL,
    received_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    CONSTRAINT fiscal_document_callbacks_pkey PRIMARY KEY (id),
    CONSTRAINT fiscal_document_callbacks_document_fkey FOREIGN KEY (
        organization_id,
        store_id,
        fiscal_document_id
    ) REFERENCES fiscal_documents (organization_id, store_id, id) ON DELETE CASCADE,
    CONSTRAINT fiscal_document_callbacks_provider_key_unique UNIQUE (provider, callback_key),
    CONSTRAINT fiscal_document_callbacks_provider_not_blank CHECK (BTRIM (provider) <> ''),
    CONSTRAINT fiscal_document_callbacks_callback_key_not_blank CHECK (BTRIM (callback_key) <> ''),
    CONSTRAINT fiscal_document_callbacks_payload_not_blank CHECK (BTRIM (payload) <> ''),
    CONSTRAINT fiscal_document_callbacks_processed_at_check CHECK (
        processed_at IS NULL
        OR processed_at >= received_at
    ),
    CONSTRAINT fiscal_document_callbacks_updated_at_check CHECK (updated_at >= created_at)
);

CREATE INDEX idx_fiscal_document_callbacks_document_received ON fiscal_document_callbacks (
    organization_id,
    store_id,
    fiscal_document_id,
    received_at DESC
);

CREATE TRIGGER trg_fiscal_document_callbacks_touch_updated_at
BEFORE UPDATE ON fiscal_document_callbacks
FOR EACH ROW
EXECUTE FUNCTION touch_updated_at();

COMMENT ON
TABLE fiscal_document_callbacks IS 'Provider callbacks deduplicated globally by provider and callback key.';