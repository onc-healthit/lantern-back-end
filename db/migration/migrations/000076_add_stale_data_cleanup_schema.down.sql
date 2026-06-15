BEGIN;

ALTER TABLE list_source_info DROP CONSTRAINT IF EXISTS list_source_info_pkey;
DROP INDEX IF EXISTS idx_list_source_info_updated_at;
DROP INDEX IF EXISTS idx_fhir_endpoints_metadata_url;
DROP INDEX IF EXISTS idx_fhir_endpoints_availability_url;
DROP INDEX IF EXISTS idx_fhir_endpoints_list_source;
ALTER TABLE list_source_info DROP COLUMN IF EXISTS updated_at;

COMMIT;
