-- name: CreateFiscalDocumentForStore :one
WITH inserted AS (
    INSERT INTO fiscal_documents (
        organization_id,
        store_id,
        sale_id,
        series,
        number,
        access_key,
        protocol,
        provider,
        external_reference,
        xml,
        error_code,
        error_message,
        issued_at,
        cancelled_at
    )
    VALUES (
        sqlc.arg(organization_id),
        sqlc.arg(store_id),
        sqlc.arg(sale_id),
        sqlc.narg(series),
        sqlc.narg(number),
        sqlc.narg(access_key),
        sqlc.narg(protocol),
        sqlc.narg(provider),
        sqlc.narg(external_reference),
        sqlc.narg(xml),
        sqlc.narg(error_code),
        sqlc.narg(error_message),
        sqlc.narg(issued_at),
        sqlc.narg(cancelled_at)
    )
    ON CONFLICT (organization_id, store_id, sale_id) DO UPDATE
    SET sale_id = EXCLUDED.sale_id
    RETURNING
        id,
        organization_id,
        store_id,
        sale_id,
        status,
        environment,
        document_model,
        series,
        number,
        access_key,
        protocol,
        provider,
        external_reference,
        xml,
        error_code,
        error_message,
        issued_at,
        cancelled_at,
        created_at,
        updated_at
)
SELECT * FROM inserted
UNION ALL
SELECT
    document.id,
    document.organization_id,
    document.store_id,
    document.sale_id,
    document.status,
    document.environment,
    document.document_model,
    document.series,
    document.number,
    document.access_key,
    document.protocol,
    document.provider,
    document.external_reference,
    document.xml,
    document.error_code,
    document.error_message,
    document.issued_at,
    document.cancelled_at,
    document.created_at,
    document.updated_at
FROM fiscal_documents AS document
WHERE document.organization_id = sqlc.arg(organization_id)
  AND document.store_id = sqlc.arg(store_id)
  AND document.sale_id = sqlc.arg(sale_id)
LIMIT 1;

-- name: GetFiscalDocumentByIDForStore :one
SELECT
    document.id,
    document.organization_id,
    document.store_id,
    document.sale_id,
    document.status,
    document.environment,
    document.document_model,
    document.series,
    document.number,
    document.access_key,
    document.protocol,
    document.provider,
    document.external_reference,
    document.xml,
    document.error_code,
    document.error_message,
    document.issued_at,
    document.cancelled_at,
    document.created_at,
    document.updated_at
FROM fiscal_documents AS document
WHERE document.organization_id = sqlc.arg(organization_id)
  AND document.store_id = sqlc.arg(store_id)
  AND document.id = sqlc.arg(id)
LIMIT 1;

-- name: GetFiscalDocumentBySaleIDForStore :one
SELECT
    document.id,
    document.organization_id,
    document.store_id,
    document.sale_id,
    document.status,
    document.environment,
    document.document_model,
    document.series,
    document.number,
    document.access_key,
    document.protocol,
    document.provider,
    document.external_reference,
    document.xml,
    document.error_code,
    document.error_message,
    document.issued_at,
    document.cancelled_at,
    document.created_at,
    document.updated_at
FROM fiscal_documents AS document
WHERE document.organization_id = sqlc.arg(organization_id)
  AND document.store_id = sqlc.arg(store_id)
  AND document.sale_id = sqlc.arg(sale_id)
LIMIT 1;

-- name: LockFiscalDocumentByIDForStore :one
SELECT
    document.id,
    document.organization_id,
    document.store_id,
    document.sale_id,
    document.status,
    document.environment,
    document.document_model,
    document.series,
    document.number,
    document.access_key,
    document.protocol,
    document.provider,
    document.external_reference,
    document.xml,
    document.error_code,
    document.error_message,
    document.issued_at,
    document.cancelled_at,
    document.created_at,
    document.updated_at
FROM fiscal_documents AS document
WHERE document.organization_id = sqlc.arg(organization_id)
  AND document.store_id = sqlc.arg(store_id)
  AND document.id = sqlc.arg(id)
FOR UPDATE;

-- name: LockFiscalDocumentBySaleIDForStore :one
SELECT
    document.id,
    document.organization_id,
    document.store_id,
    document.sale_id,
    document.status,
    document.environment,
    document.document_model,
    document.series,
    document.number,
    document.access_key,
    document.protocol,
    document.provider,
    document.external_reference,
    document.xml,
    document.error_code,
    document.error_message,
    document.issued_at,
    document.cancelled_at,
    document.created_at,
    document.updated_at
FROM fiscal_documents AS document
WHERE document.organization_id = sqlc.arg(organization_id)
  AND document.store_id = sqlc.arg(store_id)
  AND document.sale_id = sqlc.arg(sale_id)
FOR UPDATE;

-- name: MarkFiscalDocumentProcessingForStore :one
UPDATE fiscal_documents AS document
SET status = 'PROCESSING'
WHERE document.organization_id = sqlc.arg(organization_id)
  AND document.store_id = sqlc.arg(store_id)
  AND document.id = sqlc.arg(id)
  AND document.status IN ('PENDING', 'REJECTED', 'ERROR')
RETURNING
    document.id,
    document.organization_id,
    document.store_id,
    document.sale_id,
    document.status,
    document.environment,
    document.document_model,
    document.series,
    document.number,
    document.access_key,
    document.protocol,
    document.provider,
    document.external_reference,
    document.xml,
    document.error_code,
    document.error_message,
    document.issued_at,
    document.cancelled_at,
    document.created_at,
    document.updated_at;

-- name: MarkFiscalDocumentAuthorizedForStore :one
UPDATE fiscal_documents AS document
SET
    status = 'AUTHORIZED',
    access_key = sqlc.arg(access_key),
    protocol = sqlc.arg(protocol),
    provider = sqlc.arg(provider),
    external_reference = sqlc.narg(external_reference),
    xml = sqlc.narg(xml),
    issued_at = COALESCE(sqlc.narg(issued_at), NOW()),
    error_code = NULL,
    error_message = NULL
WHERE document.organization_id = sqlc.arg(organization_id)
  AND document.store_id = sqlc.arg(store_id)
  AND document.id = sqlc.arg(id)
  AND document.status IN ('PENDING', 'PROCESSING')
RETURNING
    document.id,
    document.organization_id,
    document.store_id,
    document.sale_id,
    document.status,
    document.environment,
    document.document_model,
    document.series,
    document.number,
    document.access_key,
    document.protocol,
    document.provider,
    document.external_reference,
    document.xml,
    document.error_code,
    document.error_message,
    document.issued_at,
    document.cancelled_at,
    document.created_at,
    document.updated_at;

-- name: MarkFiscalDocumentRejectedForStore :one
UPDATE fiscal_documents AS document
SET
    status = 'REJECTED',
    error_code = sqlc.arg(error_code),
    error_message = sqlc.arg(error_message)
WHERE document.organization_id = sqlc.arg(organization_id)
  AND document.store_id = sqlc.arg(store_id)
  AND document.id = sqlc.arg(id)
  AND document.status IN ('PENDING', 'PROCESSING', 'ERROR')
RETURNING
    document.id,
    document.organization_id,
    document.store_id,
    document.sale_id,
    document.status,
    document.environment,
    document.document_model,
    document.series,
    document.number,
    document.access_key,
    document.protocol,
    document.provider,
    document.external_reference,
    document.xml,
    document.error_code,
    document.error_message,
    document.issued_at,
    document.cancelled_at,
    document.created_at,
    document.updated_at;

-- name: MarkFiscalDocumentErrorForStore :one
UPDATE fiscal_documents AS document
SET
    status = 'ERROR',
    error_code = sqlc.arg(error_code),
    error_message = sqlc.arg(error_message)
WHERE document.organization_id = sqlc.arg(organization_id)
  AND document.store_id = sqlc.arg(store_id)
  AND document.id = sqlc.arg(id)
  AND document.status IN ('PENDING', 'PROCESSING', 'REJECTED')
RETURNING
    document.id,
    document.organization_id,
    document.store_id,
    document.sale_id,
    document.status,
    document.environment,
    document.document_model,
    document.series,
    document.number,
    document.access_key,
    document.protocol,
    document.provider,
    document.external_reference,
    document.xml,
    document.error_code,
    document.error_message,
    document.issued_at,
    document.cancelled_at,
    document.created_at,
    document.updated_at;

-- name: MarkFiscalDocumentCancelledForStore :one
UPDATE fiscal_documents AS document
SET
    status = 'CANCELLED',
    cancelled_at = COALESCE(sqlc.narg(cancelled_at), NOW())
WHERE document.organization_id = sqlc.arg(organization_id)
  AND document.store_id = sqlc.arg(store_id)
  AND document.id = sqlc.arg(id)
  AND document.status IN ('PENDING', 'PROCESSING', 'AUTHORIZED', 'REJECTED', 'ERROR')
  AND document.access_key IS NOT NULL
  AND document.protocol IS NOT NULL
  AND document.issued_at IS NOT NULL
RETURNING
    document.id,
    document.organization_id,
    document.store_id,
    document.sale_id,
    document.status,
    document.environment,
    document.document_model,
    document.series,
    document.number,
    document.access_key,
    document.protocol,
    document.provider,
    document.external_reference,
    document.xml,
    document.error_code,
    document.error_message,
    document.issued_at,
    document.cancelled_at,
    document.created_at,
    document.updated_at;

-- name: UpsertFiscalDocumentCallbackForStore :one
INSERT INTO fiscal_document_callbacks (
    organization_id,
    store_id,
    fiscal_document_id,
    provider,
    callback_key,
    payload,
    received_at,
    processed_at
)
VALUES (
    sqlc.arg(organization_id),
    sqlc.arg(store_id),
    sqlc.arg(fiscal_document_id),
    sqlc.arg(provider),
    sqlc.arg(callback_key),
    sqlc.arg(payload),
        COALESCE(CAST(sqlc.narg(received_at) AS TIMESTAMPTZ), NOW()),
    sqlc.narg(processed_at)
)
ON CONFLICT (provider, callback_key) DO UPDATE
SET
    payload = EXCLUDED.payload,
    processed_at = COALESCE(fiscal_document_callbacks.processed_at, EXCLUDED.processed_at)
WHERE fiscal_document_callbacks.organization_id = EXCLUDED.organization_id
  AND fiscal_document_callbacks.store_id = EXCLUDED.store_id
  AND fiscal_document_callbacks.fiscal_document_id = EXCLUDED.fiscal_document_id
RETURNING
    id,
    organization_id,
    store_id,
    fiscal_document_id,
    provider,
    callback_key,
    payload,
    received_at,
    processed_at,
    created_at,
    updated_at;

-- name: ListFiscalDocumentCallbacksByDocumentIDForStore :many
SELECT
    callback.id,
    callback.organization_id,
    callback.store_id,
    callback.fiscal_document_id,
    callback.provider,
    callback.callback_key,
    callback.payload,
    callback.received_at,
    callback.processed_at,
    callback.created_at,
    callback.updated_at
FROM fiscal_document_callbacks AS callback
WHERE callback.organization_id = sqlc.arg(organization_id)
  AND callback.store_id = sqlc.arg(store_id)
  AND callback.fiscal_document_id = sqlc.arg(fiscal_document_id)
ORDER BY callback.received_at DESC, callback.id DESC;

-- name: CountFiscalDocumentCallbacksByDocumentIDForStore :one
SELECT COUNT(*)
FROM fiscal_document_callbacks AS callback
WHERE callback.organization_id = sqlc.arg(organization_id)
  AND callback.store_id = sqlc.arg(store_id)
  AND callback.fiscal_document_id = sqlc.arg(fiscal_document_id);

-- name: LockFiscalDocumentCallbackByIDForStore :one
SELECT
    callback.id,
    callback.organization_id,
    callback.store_id,
    callback.fiscal_document_id,
    callback.provider,
    callback.callback_key,
    callback.payload,
    callback.received_at,
    callback.processed_at,
    callback.created_at,
    callback.updated_at
FROM fiscal_document_callbacks AS callback
WHERE callback.organization_id = sqlc.arg(organization_id)
  AND callback.store_id = sqlc.arg(store_id)
  AND callback.id = sqlc.arg(id)
FOR UPDATE;

-- name: MarkFiscalDocumentCallbackProcessedForStore :one
UPDATE fiscal_document_callbacks AS callback
SET processed_at = COALESCE(sqlc.narg(processed_at), NOW())
WHERE callback.organization_id = sqlc.arg(organization_id)
  AND callback.store_id = sqlc.arg(store_id)
  AND callback.id = sqlc.arg(id)
RETURNING
    callback.id,
    callback.organization_id,
    callback.store_id,
    callback.fiscal_document_id,
    callback.provider,
    callback.callback_key,
    callback.payload,
    callback.received_at,
    callback.processed_at,
    callback.created_at,
    callback.updated_at;
