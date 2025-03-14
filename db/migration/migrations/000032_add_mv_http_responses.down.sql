BEGIN;

DROP MATERIALIZED VIEW IF EXISTS mv_http_responses CASCADE;

DROP INDEX IF EXISTS mv_http_responses_uniq;

DROP INDEX IF EXISTS mv_http_responses_vendor_name_idx;

COMMIT;