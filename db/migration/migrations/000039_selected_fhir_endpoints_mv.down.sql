BEGIN;

DROP MATERIALIZED VIEW IF EXISTS selected_fhir_endpoints_mv CASCADE;

DROP INDEX IF EXISTS idx_selected_fhir_endpoints_mv_unique;

COMMIT;