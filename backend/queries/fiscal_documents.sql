-- name: CreateFiscalDocument :one
WITH inserted AS (
    INSERT INTO fiscal_documents (
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
    ON CONFLICT (sale_id) DO NOTHING
    RETURNING
        id,
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
SELECT
    id,
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
FROM inserted
UNION ALL
SELECT
    id,
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
FROM fiscal_documents
WHERE sale_id = sqlc.arg(sale_id)
LIMIT 1;

-- name: GetFiscalDocumentByID :one
SELECT
    id,
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
FROM fiscal_documents
WHERE id = sqlc.arg(id)
LIMIT 1;

-- name: GetFiscalDocumentBySaleID :one
SELECT
    id,
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
FROM fiscal_documents
WHERE sale_id = sqlc.arg(sale_id)
LIMIT 1;

-- name: LockFiscalDocumentByID :one
SELECT
    id,
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
FROM fiscal_documents
WHERE id = sqlc.arg(id)
FOR UPDATE;

-- name: LockFiscalDocumentBySaleID :one
SELECT
    id,
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
FROM fiscal_documents
WHERE sale_id = sqlc.arg(sale_id)
FOR UPDATE;

-- name: MarkFiscalDocumentProcessing :one
UPDATE fiscal_documents
SET status = 'PROCESSING'
WHERE id = sqlc.arg(id)
  AND status = 'PENDING'
RETURNING
    id,
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
    updated_at;

-- name: MarkFiscalDocumentAuthorized :one
UPDATE fiscal_documents
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
WHERE id = sqlc.arg(id)
  AND status IN ('PENDING', 'PROCESSING')
RETURNING
    id,
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
    updated_at;

-- name: MarkFiscalDocumentRejected :one
UPDATE fiscal_documents
SET
    status = 'REJECTED',
    error_code = sqlc.arg(error_code),
    error_message = sqlc.arg(error_message)
WHERE id = sqlc.arg(id)
  AND status IN ('PENDING', 'PROCESSING', 'ERROR')
RETURNING
    id,
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
    updated_at;

-- name: MarkFiscalDocumentError :one
UPDATE fiscal_documents
SET
    status = 'ERROR',
    error_code = sqlc.arg(error_code),
    error_message = sqlc.arg(error_message)
WHERE id = sqlc.arg(id)
  AND status IN ('PENDING', 'PROCESSING', 'REJECTED')
RETURNING
    id,
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
    updated_at;

-- name: MarkFiscalDocumentCancelled :one
UPDATE fiscal_documents
SET
    status = 'CANCELLED',
    cancelled_at = COALESCE(sqlc.narg(cancelled_at), NOW())
WHERE id = sqlc.arg(id)
  AND status IN ('PENDING', 'PROCESSING', 'AUTHORIZED', 'REJECTED', 'ERROR')
RETURNING
    id,
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
    updated_at;

-- name: UpsertFiscalDocumentCallback :one
WITH inserted AS (
    INSERT INTO fiscal_document_callbacks (
        fiscal_document_id,
        provider,
        callback_key,
        payload,
        received_at,
        processed_at
    )
    VALUES (
        sqlc.arg(fiscal_document_id),
        sqlc.arg(provider),
        sqlc.arg(callback_key),
        sqlc.arg(payload),
        COALESCE(sqlc.narg(received_at), NOW()),
        sqlc.narg(processed_at)
    )
    ON CONFLICT (provider, callback_key) DO UPDATE
    SET
        payload = EXCLUDED.payload,
        processed_at = COALESCE(fiscal_document_callbacks.processed_at, EXCLUDED.processed_at),
        updated_at = NOW()
    RETURNING
        id,
        fiscal_document_id,
        provider,
        callback_key,
        payload,
        received_at,
        processed_at,
        created_at,
        updated_at
)
SELECT
    id,
    fiscal_document_id,
    provider,
    callback_key,
    payload,
    received_at,
    processed_at,
    created_at,
    updated_at
FROM inserted
LIMIT 1;

-- name: ListFiscalDocumentCallbacksByDocumentID :many
SELECT
    id,
    fiscal_document_id,
    provider,
    callback_key,
    payload,
    received_at,
    processed_at,
    created_at,
    updated_at
FROM fiscal_document_callbacks
WHERE fiscal_document_id = sqlc.arg(fiscal_document_id)
ORDER BY received_at DESC;

-- name: MarkFiscalDocumentCallbackProcessed :one
UPDATE fiscal_document_callbacks
SET processed_at = COALESCE(sqlc.narg(processed_at), NOW())
WHERE id = sqlc.arg(id)
RETURNING
    id,
    fiscal_document_id,
    provider,
    callback_key,
    payload,
    received_at,
    processed_at,
    created_at,
    updated_at;
