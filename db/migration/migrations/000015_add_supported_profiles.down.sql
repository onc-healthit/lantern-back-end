BEGIN;

ALTER TABLE fhir_endpoints_info DROP COLUMN IF EXISTS supported_profiles CASCADE; 
ALTER TABLE fhir_endpoints_info_history DROP COLUMN IF EXISTS supported_profiles CASCADE;

COMMIT;
