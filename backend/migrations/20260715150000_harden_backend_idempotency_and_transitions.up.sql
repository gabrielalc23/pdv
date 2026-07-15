-- Migration: harden_backend_idempotency_and_transitions
-- Purpose:
--   Adds database triggers and callback storage for idempotency and

CREATE OR REPLACE FUNCTION touch_updated_at()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION validate_status_transition()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    old_status TEXT := OLD.status::TEXT;
    new_status TEXT := NEW.status::TEXT;
BEGIN
    IF old_status = new_status THEN
        RETURN NEW;
    END IF;

    IF TG_TABLE_NAME = 'sales' THEN
        IF old_status = 'OPEN' AND new_status IN ('COMPLETED', 'CANCELLED') THEN
            RETURN NEW;
        END IF;

        IF old_status = 'COMPLETED' AND new_status = 'CANCELLED' THEN
            RETURN NEW;
        END IF;

        RAISE EXCEPTION 'invalid sale status transition: % -> %', old_status, new_status
            USING ERRCODE = '23514';
    ELSIF TG_TABLE_NAME = 'payments' THEN
        IF old_status = 'PENDING' AND new_status IN ('APPROVED', 'DECLINED', 'CANCELLED') THEN
            RETURN NEW;
        END IF;

        IF old_status = 'APPROVED' AND new_status = 'CANCELLED' THEN
            RETURN NEW;
        END IF;

        RAISE EXCEPTION 'invalid payment status transition: % -> %', old_status, new_status
            USING ERRCODE = '23514';
    ELSIF TG_TABLE_NAME = 'fiscal_documents' THEN
        IF old_status = 'PENDING' AND new_status IN ('PROCESSING', 'AUTHORIZED', 'REJECTED', 'ERROR', 'CANCELLED') THEN
            RETURN NEW;
        END IF;

        IF old_status = 'PROCESSING' AND new_status IN ('AUTHORIZED', 'REJECTED', 'ERROR', 'CANCELLED') THEN
            RETURN NEW;
        END IF;

        IF old_status = 'AUTHORIZED' AND new_status = 'CANCELLED' THEN
            RETURN NEW;
        END IF;

        IF old_status = 'REJECTED' AND new_status IN ('PROCESSING', 'ERROR', 'CANCELLED') THEN
            RETURN NEW;
        END IF;

        IF old_status = 'ERROR' AND new_status IN ('PROCESSING', 'REJECTED', 'CANCELLED') THEN
            RETURN NEW;
        END IF;

        RAISE EXCEPTION 'invalid fiscal document status transition: % -> %', old_status, new_status
            USING ERRCODE = '23514';
    END IF;

    RETURN NEW;
END;
$$;

CREATE TABLE fiscal_document_callbacks (
    id UUID PRIMARY KEY DEFAULT uuidv7 (),
    fiscal_document_id UUID NOT NULL REFERENCES fiscal_documents (id) ON DELETE CASCADE,
    provider VARCHAR(100) NOT NULL,
    callback_key VARCHAR(255) NOT NULL,
    payload TEXT NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fiscal_document_callbacks_provider_not_blank CHECK (BTRIM (provider) <> ''),
    CONSTRAINT fiscal_document_callbacks_callback_key_not_blank CHECK (BTRIM (callback_key) <> ''),
    CONSTRAINT fiscal_document_callbacks_provider_key_unique UNIQUE (provider, callback_key)
);

CREATE INDEX idx_fiscal_document_callbacks_document_received ON fiscal_document_callbacks (
    fiscal_document_id,
    received_at DESC
);

CREATE TRIGGER trg_products_touch_updated_at
BEFORE UPDATE ON products
FOR EACH ROW
EXECUTE FUNCTION touch_updated_at();

CREATE TRIGGER trg_inventory_touch_updated_at
BEFORE UPDATE ON inventory
FOR EACH ROW
EXECUTE FUNCTION touch_updated_at();

CREATE TRIGGER trg_sales_touch_updated_at
BEFORE UPDATE ON sales
FOR EACH ROW
EXECUTE FUNCTION touch_updated_at();

CREATE TRIGGER trg_sales_validate_status_transition
BEFORE UPDATE OF status ON sales
FOR EACH ROW
EXECUTE FUNCTION validate_status_transition();

CREATE TRIGGER trg_payment_methods_touch_updated_at
BEFORE UPDATE ON payment_methods
FOR EACH ROW
EXECUTE FUNCTION touch_updated_at();

CREATE TRIGGER trg_payments_touch_updated_at
BEFORE UPDATE ON payments
FOR EACH ROW
EXECUTE FUNCTION touch_updated_at();

CREATE TRIGGER trg_payments_validate_status_transition
BEFORE UPDATE OF status ON payments
FOR EACH ROW
EXECUTE FUNCTION validate_status_transition();

CREATE TRIGGER trg_fiscal_documents_touch_updated_at
BEFORE UPDATE ON fiscal_documents
FOR EACH ROW
EXECUTE FUNCTION touch_updated_at();

CREATE TRIGGER trg_fiscal_documents_validate_status_transition
BEFORE UPDATE OF status ON fiscal_documents
FOR EACH ROW
EXECUTE FUNCTION validate_status_transition();

CREATE TRIGGER trg_fiscal_document_callbacks_touch_updated_at
BEFORE UPDATE ON fiscal_document_callbacks
FOR EACH ROW
EXECUTE FUNCTION touch_updated_at();