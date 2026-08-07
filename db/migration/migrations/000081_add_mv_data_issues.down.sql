BEGIN;

-- Drop materialized views in reverse order
DROP MATERIALIZED VIEW IF EXISTS mv_chpl_coverage_summary CASCADE;
DROP MATERIALIZED VIEW IF EXISTS mv_developer_bundle_issues CASCADE;
DROP MATERIALIZED VIEW IF EXISTS mv_developer_data_issues CASCADE;

COMMIT;
