BEGIN;

-- Drop all materialized views and their indexes
DROP MATERIALIZED VIEW IF EXISTS mv_npi_organization_matches CASCADE;
DROP MATERIALIZED VIEW IF EXISTS mv_endpoint_list_organizations CASCADE;
DROP MATERIALIZED VIEW IF EXISTS mv_endpoint_locations CASCADE;

COMMIT;