BEGIN;

DROP MATERIALIZED VIEW IF EXISTS endpoint_export_mv CASCADE;

DROP INDEX IF EXISTS endpoint_export_mv_unique_idx;

COMMIT;