-- 000025_create_rate_limit_table.down.sql
DROP FUNCTION IF EXISTS cleanup_rate_limit_windows(INT);
DROP FUNCTION IF EXISTS check_rate_limit(VARCHAR, INT, INT);
DROP TABLE IF EXISTS rate_limit_windows;
