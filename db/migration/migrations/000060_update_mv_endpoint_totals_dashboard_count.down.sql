BEGIN;

-- Drop the main materialized view with CASCADE to automatically handle all dependencies
DROP MATERIALIZED VIEW IF EXISTS mv_endpoint_totals CASCADE;

-- Recreate mv_endpoint_totals with original logic
CREATE MATERIALIZED VIEW mv_endpoint_totals AS
WITH latest_metadata AS (
    SELECT max(updated_at) AS last_updated
    FROM fhir_endpoints_metadata
), 
totals AS (
    SELECT 
        -- Original logic: count unique URLs only
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

-- Recreate the unique index
CREATE UNIQUE INDEX idx_mv_endpoint_totals_date ON mv_endpoint_totals(aggregation_date);

-- Recreate the dependent materialized view with original logic
CREATE MATERIALIZED VIEW mv_endpoint_security_counts AS
WITH 
-- Get total indexed endpoints from mv_endpoint_totals
total_endpoints AS (
  SELECT 
    'Total Indexed Endpoints' AS status,
    all_endpoints::integer AS endpoints,
    1 AS sort_order
  FROM mv_endpoint_totals
  ORDER BY aggregation_date DESC
  LIMIT 1
),
-- Get HTTP 200 responses from mv_response_tally
http_200_endpoints AS (
  SELECT 
    'Endpoints with successful response (HTTP 200)' AS status,
    http_200::integer AS endpoints,
    2 AS sort_order
  FROM mv_response_tally
  LIMIT 1
),
-- Get non-200 responses from mv_response_tally
http_non200_endpoints AS (
  SELECT 
    'Endpoints with unsuccessful response' AS status,
    http_non200::integer AS endpoints,
    3 AS sort_order
  FROM mv_response_tally
  LIMIT 1
),
-- Get count of endpoints without valid capability statement
no_cap_statement AS (
  SELECT 
    'Endpoints without valid CapabilityStatement / Conformance Resource' AS status,
    COUNT(*)::integer AS endpoints,
    4 AS sort_order
  FROM fhir_endpoints_info 
  WHERE jsonb_typeof(capability_statement::jsonb) <> 'object' 
    AND requested_fhir_version = 'None'
),
-- Get count of endpoints with valid security resource
security_endpoints AS (
  SELECT 
    'Endpoints with valid security resource' AS status,
    COUNT(DISTINCT id)::integer AS endpoints,
    5 AS sort_order
  FROM mv_get_security_endpoints
),
-- Combine all results
combined_results AS (
  SELECT status, endpoints, sort_order FROM total_endpoints
  UNION ALL
  SELECT status, endpoints, sort_order FROM http_200_endpoints
  UNION ALL
  SELECT status, endpoints, sort_order FROM http_non200_endpoints
  UNION ALL
  SELECT status, endpoints, sort_order FROM no_cap_statement
  UNION ALL
  SELECT status, endpoints, sort_order FROM security_endpoints
)
-- Final select with ordering
SELECT 
  status AS "Status",
  endpoints AS "Endpoints"
FROM combined_results
ORDER BY sort_order;

-- Create the unique index for mv_endpoint_security_counts
CREATE UNIQUE INDEX idx_mv_endpoint_security_counts ON mv_endpoint_security_counts("Status");

COMMIT;