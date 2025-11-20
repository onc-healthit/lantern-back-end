BEGIN;

-- Add updated_at column to list_source_info table
ALTER TABLE list_source_info ADD COLUMN updated_at TIMESTAMPTZ DEFAULT NOW();

-- Update existing records to have current timestamp (to be run on prod just once during initial set up)
UPDATE list_source_info SET updated_at = NOW();

-- Add index for performance on cleanup queries
CREATE INDEX idx_list_source_info_updated_at ON list_source_info(updated_at);

-- Add missing indexes needed for efficient cleanup operations
CREATE INDEX idx_fhir_endpoints_metadata_url ON fhir_endpoints_metadata(url);
CREATE INDEX idx_fhir_endpoints_availability_url ON fhir_endpoints_availability(url);

COMMIT;