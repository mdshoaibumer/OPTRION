-- 000023_add_api_key_rotation.down.sql
-- Remove API key rotation support.

DROP INDEX IF EXISTS idx_api_keys_grace_period;
DROP INDEX IF EXISTS idx_api_keys_rotated_from;

ALTER TABLE api_keys DROP COLUMN IF EXISTS rotated_from;
ALTER TABLE api_keys DROP COLUMN IF EXISTS rotated_at;
ALTER TABLE api_keys DROP COLUMN IF EXISTS grace_expires_at;
