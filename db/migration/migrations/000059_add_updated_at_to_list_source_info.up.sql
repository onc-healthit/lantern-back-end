BEGIN;

-- Add updated_at column to list_source_info table
ALTER TABLE list_source_info ADD COLUMN updated_at TIMESTAMPTZ DEFAULT NOW();

-- Update existing records to have current timestamp
UPDATE list_source_info SET updated_at = NOW();

-- Add index for performance on cleanup queries
CREATE INDEX idx_list_source_info_updated_at ON list_source_info(updated_at);

COMMIT;