-- 000026_create_processed_events.down.sql
DROP FUNCTION IF EXISTS purge_processed_events(INT);
DROP TABLE IF EXISTS processed_events;
