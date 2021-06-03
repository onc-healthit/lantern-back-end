BEGIN;

ALTER TABLE fhir_endpoints ADD COLUMN versions_response JSONB;

COMMIT;