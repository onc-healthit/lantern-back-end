BEGIN;

DROP MATERIALIZED VIEW IF EXISTS mv_endpoint_totals CASCADE;

CREATE MATERIALIZED VIEW mv_endpoint_totals AS
WITH latest_metadata AS (
    SELECT max(updated_at) AS last_updated
    FROM fhir_endpoints_metadata
), 
totals AS (
    SELECT 
        (SELECT count(DISTINCT url) FROM fhir_endpoints) AS all_endpoints,
        (SELECT count(DISTINCT url) 
         FROM fhir_endpoints_info 
         WHERE requested_fhir_version = 'None') AS indexed_endpoints
)
SELECT 
    now() AS aggregation_date,
    totals.all_endpoints,
    totals.indexed_endpoints,
    greatest(totals.all_endpoints - totals.indexed_endpoints, 0) AS nonindexed_endpoints,
    (SELECT latest_metadata.last_updated FROM latest_metadata) AS last_updated
FROM totals;

DROP INDEX IF EXISTS idx_mv_endpoint_totals_date;
CREATE UNIQUE INDEX idx_mv_endpoint_totals_date ON mv_endpoint_totals(aggregation_date);

COMMIT;