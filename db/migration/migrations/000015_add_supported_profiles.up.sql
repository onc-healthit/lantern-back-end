ALTER TABLE fhir_endpoints_info ADD COLUMN IF NOT EXISTS supported_profiles JSONB;
ALTER TABLE fhir_endpoints_info_history ADD COLUMN IF NOT EXISTS supported_profiles JSONB;