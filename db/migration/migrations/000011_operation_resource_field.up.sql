BEGIN;

ALTER TABLE fhir_endpoints_info ADD COLUMN operation_resource JSONB;
ALTER TABLE fhir_endpoints_info_history ADD COLUMN operation_resource JSONB;

COMMIT;