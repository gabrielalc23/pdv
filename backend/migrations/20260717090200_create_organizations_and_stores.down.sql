DROP TRIGGER IF EXISTS trg_stores_touch_updated_at ON stores;
DROP TABLE IF EXISTS stores;

DROP TRIGGER IF EXISTS trg_organizations_touch_updated_at ON organizations;
DROP TABLE IF EXISTS organizations;

DROP TYPE IF EXISTS store_status;
DROP TYPE IF EXISTS organization_status;
