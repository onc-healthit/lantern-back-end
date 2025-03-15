BEGIN;

DROP MATERIALIZED VIEW IF EXISTS mv_resource_interactions CASCADE;

DROP INDEX IF EXISTS mv_resource_interactions_uniq;

DROP INDEX IF EXISTS mv_resource_interactions_vendor_name_idx;

DROP INDEX IF EXISTS mv_resource_interactions_fhir_version_idx;

DROP INDEX IF EXISTS mv_resource_interactions_resource_type_idx;

DROP INDEX IF EXISTS mv_resource_interactions_operations_idx;

COMMIT;