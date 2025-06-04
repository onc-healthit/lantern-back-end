BEGIN;

DROP MATERIALIZED VIEW IF EXISTS mv_endpoint_totals;
DROP INDEX IF EXISTS idx_mv_endpoint_totals_date;

CREATE MATERIALIZED VIEW mv_endpoint_totals AS
WITH latest_metadata AS (
    SELECT max(updated_at) AS last_updated
    FROM fhir_endpoints_metadata
), 
totals AS (
    SELECT 
        -- Count (url, fhir_version) combinations to match Endpoints tab logic
        (SELECT count(*) FROM (SELECT DISTINCT url, fhir_version FROM selected_fhir_endpoints_mv) AS combinations) AS all_endpoints,
        (SELECT count(*) FROM (SELECT DISTINCT fei.url, fei.capability_fhir_version 
        FROM fhir_endpoints_info fei
        WHERE fei.requested_fhir_version = 'None') AS combinations) AS indexed_endpoints
)
SELECT 
    now() AS aggregation_date,
    totals.all_endpoints,
    totals.indexed_endpoints,
    greatest(totals.all_endpoints - totals.indexed_endpoints, 0) AS nonindexed_endpoints,
    (SELECT latest_metadata.last_updated FROM latest_metadata) AS last_updated
FROM totals;

-- Recreate the unique index
CREATE UNIQUE INDEX idx_mv_endpoint_totals_date ON mv_endpoint_totals(aggregation_date);

COMMIT;