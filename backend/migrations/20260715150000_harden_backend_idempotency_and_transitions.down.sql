DROP TRIGGER IF EXISTS trg_fiscal_document_callbacks_touch_updated_at ON fiscal_document_callbacks;
DROP TRIGGER IF EXISTS trg_fiscal_documents_validate_status_transition ON fiscal_documents;
DROP TRIGGER IF EXISTS trg_fiscal_documents_touch_updated_at ON fiscal_documents;
DROP TRIGGER IF EXISTS trg_payments_validate_status_transition ON payments;
DROP TRIGGER IF EXISTS trg_payments_touch_updated_at ON payments;
DROP TRIGGER IF EXISTS trg_payment_methods_touch_updated_at ON payment_methods;
DROP TRIGGER IF EXISTS trg_sales_validate_status_transition ON sales;
DROP TRIGGER IF EXISTS trg_sales_touch_updated_at ON sales;
DROP TRIGGER IF EXISTS trg_inventory_touch_updated_at ON inventory;
DROP TRIGGER IF EXISTS trg_products_touch_updated_at ON products;

DROP INDEX IF EXISTS idx_fiscal_document_callbacks_document_received;
DROP TABLE IF EXISTS fiscal_document_callbacks;

DROP FUNCTION IF EXISTS validate_status_transition();
DROP FUNCTION IF EXISTS touch_updated_at();
