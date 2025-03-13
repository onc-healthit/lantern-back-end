BEGIN;

DROP MATERIALIZED VIEW IF EXISTS mv_resource_interactions CASCADE;

DROP INDEX IF EXISTS mv_resource_interactions_uniq;

COMMIT;