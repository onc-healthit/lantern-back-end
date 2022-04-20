BEGIN;

ALTER TABLE fhir_endpoints_info ADD COLUMN IF NOT EXISTS supported_profiles JSONB DEFAULT 'null';
ALTER TABLE fhir_endpoints_info_history ADD COLUMN IF NOT EXISTS supported_profiles JSONB DEFAULT 'null';

COMMIT;