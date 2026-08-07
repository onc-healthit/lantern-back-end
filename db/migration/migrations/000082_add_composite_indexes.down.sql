BEGIN;

DROP INDEX IF EXISTS idx_fhir_endpoints_info_url_reqver;
DROP INDEX IF EXISTS idx_fhir_endpoints_availability_url_reqver;

COMMIT;
