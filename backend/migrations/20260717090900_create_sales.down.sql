DROP TABLE IF EXISTS sale_items;

DROP TRIGGER IF EXISTS trg_sales_validate_status_transition ON sales;
DROP TRIGGER IF EXISTS trg_sales_touch_updated_at ON sales;
DROP TABLE IF EXISTS sales;

DROP TABLE IF EXISTS sale_number_counters;

DROP FUNCTION IF EXISTS validate_status_transition();
DROP TYPE IF EXISTS sale_status;
