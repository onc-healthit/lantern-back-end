BEGIN;

-- Add developer_name column to list_source_info table
-- This enables tracking which developers have empty FHIR bundles
-- (list_sources that return no endpoints)

ALTER TABLE list_source_info
ADD COLUMN IF NOT EXISTS developer_name VARCHAR(500);

-- Create index for faster lookups by developer
CREATE INDEX IF NOT EXISTS idx_list_source_info_developer_name
ON list_source_info(developer_name);

-- Add composite unique constraint on (list_source, developer_name)
-- Multiple developers can share the same list_source URL (e.g., Dev XYZ and Dev ABC both use google.com)
-- This allows the same list_source to appear multiple times with different developers
-- while preventing duplicate (list_source, developer_name) combinations
-- Required for UPSERT (INSERT ... ON CONFLICT) functionality
DO $$
BEGIN
    -- Drop old single-column unique constraint if it exists
    IF EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'list_source_info_list_source_key'
    ) THEN
        ALTER TABLE list_source_info
        DROP CONSTRAINT list_source_info_list_source_key;
    END IF;

    -- Add new composite unique constraint
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'list_source_info_list_source_developer_key'
    ) THEN
        ALTER TABLE list_source_info
        ADD CONSTRAINT list_source_info_list_source_developer_key UNIQUE (list_source, developer_name);
    END IF;
END$$;

COMMIT;
