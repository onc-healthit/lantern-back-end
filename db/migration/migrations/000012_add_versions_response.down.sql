BEGIN;

ALTER TABLE fhir_endpoints DROP COLUMN IF EXISTS versions_response; 

COMMIT;