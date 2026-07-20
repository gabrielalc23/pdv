DROP TABLE IF EXISTS inventory_movements;

DROP TRIGGER IF EXISTS trg_inventory_touch_updated_at ON inventory;
DROP TABLE IF EXISTS inventory;

DROP TYPE IF EXISTS inventory_movement_type;
