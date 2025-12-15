BEGIN;

-- Add developer_name column to list_source_info table
-- This enables tracking which developers have empty FHIR bundles
-- (list_sources that return no endpoints)

ALTER TABLE list_source_info
ADD COLUMN IF NOT EXISTS developer_name VARCHAR(500);

-- Create index for faster lookups by developer
CREATE INDEX IF NOT EXISTS idx_list_source_info_developer_name
ON list_source_info(developer_name);

-- Add unique constraint on list_source if it doesn't exist
-- This is required for UPSERT (INSERT ... ON CONFLICT) functionality
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'list_source_info_list_source_key'
    ) THEN
        ALTER TABLE list_source_info
        ADD CONSTRAINT list_source_info_list_source_key UNIQUE (list_source);
    END IF;
END$$;

COMMIT;
