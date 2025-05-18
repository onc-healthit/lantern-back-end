BEGIN;

-- Drop all materialized views and their indexes
 DROP MATERIALIZED VIEW IF EXISTS mv_capstat_sizes_tbl CASCADE;

COMMIT;