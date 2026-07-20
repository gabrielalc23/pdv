DROP TRIGGER IF EXISTS trg_fiscal_document_callbacks_touch_updated_at
ON fiscal_document_callbacks;
DROP TABLE IF EXISTS fiscal_document_callbacks;

DROP TRIGGER IF EXISTS trg_fiscal_documents_validate_status_transition
ON fiscal_documents;
DROP TRIGGER IF EXISTS trg_fiscal_documents_touch_updated_at
ON fiscal_documents;
DROP TABLE IF EXISTS fiscal_documents;

DROP TYPE IF EXISTS fiscal_environment;
DROP TYPE IF EXISTS fiscal_document_status;
