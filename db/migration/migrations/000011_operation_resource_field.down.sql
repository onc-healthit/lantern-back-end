BEGIN;

ALTER TABLE fhir_endpoints_info DROP COLUMN IF EXISTS operation_resource CASCADE;
ALTER TABLE fhir_endpoints_info_history DROP COLUMN IF EXISTS operation_resource CASCADE;

COMMIT;