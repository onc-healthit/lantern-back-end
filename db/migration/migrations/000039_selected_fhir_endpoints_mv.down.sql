BEGIN;

DROP MATERIALIZED VIEW IF EXISTS selected_fhir_endpoints_mv CASCADE;

DROP INDEX IF EXISTS idx_selected_fhir_endpoints_mv_unique;

--Drop single column index
DROP INDEX IF EXISTS idx_selected_fhir_endpoints_mv_fhir_version;
DROP INDEX IF EXISTS idx_selected_fhir_endpoints_mv_vendor_name;
DROP INDEX IF EXISTS idx_selected_fhir_endpoints_mv_availability;
DROP INDEX IF EXISTS idx_selected_fhir_endpoints_mv_is_chpl;

COMMIT;