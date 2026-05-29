-- 000001_create_tenants.down.sql
DROP TRIGGER IF EXISTS trg_tenants_updated_at ON tenants;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS tenants;
