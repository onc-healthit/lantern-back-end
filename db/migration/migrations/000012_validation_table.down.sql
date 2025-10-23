BEGIN;

ALTER TABLE fhir_endpoints_info DROP COLUMN IF EXISTS validation_result_id CASCADE;
ALTER TABLE fhir_endpoints_info_history DROP COLUMN IF EXISTS validation_result_id CASCADE;

ALTER TABLE fhir_endpoints_info ADD COLUMN IF NOT EXISTS validation JSONB;
ALTER TABLE fhir_endpoints_info_history ADD COLUMN IF NOT EXISTS validation JSONB;

DROP TABLE IF EXISTS validations;
DROP TABLE IF EXISTS validation_results;

COMMIT;
