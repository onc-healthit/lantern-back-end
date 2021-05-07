BEGIN;

ALTER TABLE fhir_endpoints_info DROP COLUMN IF EXISTS validation_result_id CASCADE;
ALTER TABLE fhir_endpoints_info_history DROP COLUMN IF EXISTS validation_result_id CASCADE;

ALTER TABLE fhir_endpoints_info ADD COLUMN validation JSONB;
ALTER TABLE fhir_endpoints_info_history ADD COLUMN validation JSONB;

DROP TABLE IF EXISTS validations;
DROP TABLE IF EXISTS validation_results;

COMMIT;
