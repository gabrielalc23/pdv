DROP TRIGGER IF EXISTS trg_payments_validate_status_transition ON payments;
DROP TRIGGER IF EXISTS trg_payments_touch_updated_at ON payments;
DROP TABLE IF EXISTS payments;

DROP TRIGGER IF EXISTS trg_store_payment_methods_touch_updated_at
ON store_payment_methods;
DROP TABLE IF EXISTS store_payment_methods;

DROP TRIGGER IF EXISTS trg_payment_methods_touch_updated_at ON payment_methods;
DROP TABLE IF EXISTS payment_methods;

DROP TYPE IF EXISTS payment_status;
DROP TYPE IF EXISTS payment_method_kind;
