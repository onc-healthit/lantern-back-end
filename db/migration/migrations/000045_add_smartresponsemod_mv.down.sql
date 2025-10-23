BEGIN;

-- Drop all materialized views and their indexes
DROP MATERIALIZED VIEW IF EXISTS mv_endpoint_organization_tbl CASCADE;
DROP MATERIALIZED VIEW IF EXISTS mv_endpoint_export_tbl CASCADE;
DROP MATERIALIZED VIEW IF EXISTS mv_http_pct CASCADE;
DROP MATERIALIZED VIEW IF EXISTS mv_well_known_endpoints CASCADE;
DROP MATERIALIZED VIEW IF EXISTS mv_well_known_no_doc CASCADE;
DROP MATERIALIZED VIEW IF EXISTS mv_smart_response_capabilities CASCADE;
DROP MATERIALIZED VIEW IF EXISTS mv_selected_endpoints CASCADE;

COMMIT;