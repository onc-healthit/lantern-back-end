BEGIN;

-- Remove the index first
DROP INDEX IF EXISTS idx_list_source_info_updated_at;

-- Remove the updated_at column
ALTER TABLE list_source_info DROP COLUMN IF EXISTS updated_at;

COMMIT;