BEGIN;

DROP TABLE IF EXISTS fhir_endpoints_availability;

ALTER TABLE fhir_endpoints_info DROP COLUMN IF EXISTS availability CASCADE; 

DROP FUNCTION IF EXISTS update_fhir_endpoint_availability_info();
DROP TRIGGER IF EXISTS update_fhir_endpoint_availability_trigger ON fhir_endpoints_info;
DROP TRIGGER IF EXISTS set_timestamp_fhir_endpoint_availability ON fhir_endpoints_availability;

COMMIT;