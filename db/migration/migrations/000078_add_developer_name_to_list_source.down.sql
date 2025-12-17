BEGIN;

-- Remove composite unique constraint
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'list_source_info_list_source_developer_key'
    ) THEN
        ALTER TABLE list_source_info
        DROP CONSTRAINT list_source_info_list_source_developer_key;
    END IF;
END$$;

-- Remove index
DROP INDEX IF EXISTS idx_list_source_info_developer_name;

-- Remove developer_name column from list_source_info
ALTER TABLE list_source_info
DROP COLUMN IF EXISTS developer_name;

COMMIT;
