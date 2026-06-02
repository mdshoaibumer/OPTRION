-- 000024_create_event_outbox.down.sql
DROP FUNCTION IF EXISTS purge_processed_outbox_events(INT);
DROP FUNCTION IF EXISTS create_outbox_partition();
DROP TABLE IF EXISTS event_outbox CASCADE;
