BEGIN;

DROP MATERIALIZED VIEW IF EXISTS response_tally_mv;

CREATE MATERIALIZED VIEW mv_endpoint_totals AS
WITH latest_metadata AS (
    SELECT max(fhir_endpoints_metadata.updated_at) AS last_updated
    FROM fhir_endpoints_metadata
), 
totals AS (
    SELECT 
        (SELECT count(DISTINCT fhir_endpoints.url) FROM fhir_endpoints) AS all_endpoints,
        (SELECT count(DISTINCT fhir_endpoints_info.url) 
         FROM fhir_endpoints_info 
         WHERE fhir_endpoints_info.requested_fhir_version IS NULL) AS indexed_endpoints
)
SELECT 
    now() AS aggregation_date,
    totals.all_endpoints,
    totals.indexed_endpoints,
    totals.all_endpoints - totals.indexed_endpoints AS nonindexed_endpoints,
    (SELECT latest_metadata.last_updated FROM latest_metadata) AS last_updated
FROM totals;

COMMIT