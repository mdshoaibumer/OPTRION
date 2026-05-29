-- 000003_create_environments.down.sql
DROP TRIGGER IF EXISTS trg_environments_updated_at ON environments;
DROP TABLE IF EXISTS environments;
