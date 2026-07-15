-- Migration: create_fiscal_documents
-- Purpose:
--   Creates fiscal documents associated with completed POS sales.
--
-- Business rules:
--   - Each sale may have only one fiscal document.
--   - Fiscal emission may happen after the sale checkout transaction.
--   - Authorized documents must contain an access key, protocol and issue date.
--   - Cancelled documents preserve their original authorization data.
--   - Fiscal documents cannot exist without a sale.
--
-- Dependencies:
--   - sales

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
    id UUID PRIMARY KEY DEFAULT uuidv7 (),
    sale_id UUID NOT NULL,
    status fiscal_document_status NOT NULL DEFAULT 'PENDING',
    environment fiscal_environment NOT NULL DEFAULT 'HOMOLOGATION',
    document_model SMALLINT NOT NULL DEFAULT 65,
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
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fiscal_documents_sale_fk FOREIGN KEY (sale_id) REFERENCES sales (id) ON DELETE RESTRICT,
    CONSTRAINT fiscal_documents_sale_unique UNIQUE (sale_id),
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
            AND
            access_key IS NOT NULL
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
    CONSTRAINT fiscal_documents_updated_at_check CHECK (updated_at >= created_at)
);

CREATE UNIQUE INDEX idx_fiscal_documents_access_key ON fiscal_documents (access_key)
WHERE
    access_key IS NOT NULL;

CREATE UNIQUE INDEX idx_fiscal_documents_provider_reference ON fiscal_documents (provider, external_reference)
WHERE
    provider IS NOT NULL
    AND external_reference IS NOT NULL;

CREATE INDEX idx_fiscal_documents_status_created_at ON fiscal_documents (status, created_at DESC);

COMMENT ON TYPE fiscal_document_status IS
    'Lifecycle status of a fiscal document emission.';

COMMENT ON TYPE fiscal_environment IS 'Fiscal authority environment used for document emission.';

COMMENT ON
TABLE fiscal_documents IS 'Stores fiscal documents issued for completed POS sales.';

COMMENT ON COLUMN fiscal_documents.sale_id IS 'Sale associated with the fiscal document.';

COMMENT ON COLUMN fiscal_documents.document_model IS 'Fiscal document model. Model 65 represents NFC-e.';

COMMENT ON COLUMN fiscal_documents.access_key IS 'Fiscal document access key returned after authorization.';

COMMENT ON COLUMN fiscal_documents.protocol IS 'Authorization protocol returned by the fiscal authority or provider.';

COMMENT ON COLUMN fiscal_documents.xml IS 'Fiscal document XML generated or returned during emission.';

COMMENT ON COLUMN fiscal_documents.external_reference IS 'Transaction identifier assigned by the fiscal integration provider.';
