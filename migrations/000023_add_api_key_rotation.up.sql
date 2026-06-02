-- 000023_add_api_key_rotation.up.sql
-- Support API key rotation: create new key, grace period for old key, then revoke.

-- Add rotation tracking columns
ALTER TABLE api_keys ADD COLUMN rotated_from UUID REFERENCES api_keys(id);
ALTER TABLE api_keys ADD COLUMN rotated_at TIMESTAMPTZ;
ALTER TABLE api_keys ADD COLUMN grace_expires_at TIMESTAMPTZ;

-- Index for finding keys in grace period
CREATE INDEX idx_api_keys_grace_period ON api_keys(grace_expires_at) 
    WHERE grace_expires_at IS NOT NULL AND status = 'active';

-- Index for rotation chain tracking
CREATE INDEX idx_api_keys_rotated_from ON api_keys(rotated_from)
    WHERE rotated_from IS NOT NULL;

COMMENT ON COLUMN api_keys.rotated_from IS 'ID of the API key this key was rotated from';
COMMENT ON COLUMN api_keys.rotated_at IS 'When the old key was rotated (set on old key)';
COMMENT ON COLUMN api_keys.grace_expires_at IS 'Old key remains valid until this time, then auto-revokes';
