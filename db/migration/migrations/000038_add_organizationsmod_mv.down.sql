BEGIN;

-- Drop all materialized views and their indexes
 DROP MATERIALIZED VIEW IF EXISTS mv_endpoint_list_organizations CASCADE;

COMMIT;