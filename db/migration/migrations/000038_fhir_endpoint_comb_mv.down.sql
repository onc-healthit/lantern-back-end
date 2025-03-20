BEGIN;

DROP MATERIALIZED VIEW IF EXISTS fhir_endpoint_comb_mv;

DROP INDEX IF EXISTS fhir_endpoint_comb_mv_unique_idx;

COMMIT;