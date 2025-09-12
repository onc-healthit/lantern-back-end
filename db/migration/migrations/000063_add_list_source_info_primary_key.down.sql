BEGIN;

-- Remove primary key constraint from list_source_info table
ALTER TABLE list_source_info DROP CONSTRAINT IF EXISTS list_source_info_pkey;

COMMIT;