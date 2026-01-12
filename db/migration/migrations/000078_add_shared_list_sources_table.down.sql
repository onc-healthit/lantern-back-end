BEGIN;

-- Drop indexes
DROP INDEX IF EXISTS idx_shared_list_sources_updated_at;
DROP INDEX IF EXISTS idx_shared_list_sources_list_source;

-- Drop table
DROP TABLE IF EXISTS shared_list_sources CASCADE;

COMMIT;
