BEGIN;

-- Remove developer_name column from list_source_info
ALTER TABLE list_source_info
DROP COLUMN IF EXISTS developer_name;

-- Note: We don't remove the unique constraint as it may have existed before
-- and other parts of the system may depend on it

COMMIT;
