BEGIN;

DROP MATERIALIZED VIEW IF EXISTS mv_get_security_endpoints CASCADE;

CREATE MATERIALIZED VIEW mv_get_security_endpoints AS
SELECT
  f.id,
  f.vendor_id,
  COALESCE(v.name, 'Unknown') AS name,
  CASE 
    WHEN capability_fhir_version = '' THEN 'No Cap Stat'
    WHEN position('-' in capability_fhir_version) > 0 THEN 
      CASE
        WHEN substring(capability_fhir_version, 1, position('-' in capability_fhir_version) - 1) IN 
            ('0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2', '1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', 
             '3.0.0', '3.0.1', '3.0.2', '3.2.0', '3.3.0', '3.5.0', '3.5a.0', '4.0.0', '4.0.1', 'No Cap Stat')
        THEN substring(capability_fhir_version, 1, position('-' in capability_fhir_version) - 1)
        ELSE 'Unknown'
      END
    WHEN capability_fhir_version IN 
        ('0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2', '1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', 
         '3.0.0', '3.0.1', '3.0.2', '3.2.0', '3.3.0', '3.5.0', '3.5a.0', '4.0.0', '4.0.1', 'No Cap Stat')
    THEN capability_fhir_version
    ELSE 'Unknown'
  END AS fhir_version,
  json_array_elements(json_array_elements(capability_statement::json#>'{rest,0,security,service}')->'coding')::json->>'code' AS code,
  json_array_elements(capability_statement::json#>'{rest,0,security}' -> 'service')::json ->> 'text' AS text
FROM fhir_endpoints_info f 
LEFT JOIN vendors v ON f.vendor_id = v.id
WHERE requested_fhir_version = 'None';

-- Create indexes for performance
CREATE UNIQUE INDEX idx_mv_get_security_endpoints ON mv_get_security_endpoints(id, code);
CREATE INDEX idx_mv_get_security_endpoints_name ON mv_get_security_endpoints(name);
CREATE INDEX idx_mv_get_security_endpoints_fhir ON mv_get_security_endpoints(fhir_version);

DROP MATERIALIZED VIEW IF EXISTS mv_auth_type_count CASCADE;

CREATE MATERIALIZED VIEW mv_auth_type_count AS
WITH endpoints_by_version AS (
  -- Get total count of distinct IDs per FHIR version
  SELECT 
    fhir_version,
    COUNT(DISTINCT id) AS tc
  FROM 
    mv_get_security_endpoints
  GROUP BY 
    fhir_version
),
endpoints_by_version_code AS (
  -- Count endpoints for each code within each FHIR version
  SELECT 
    s.fhir_version,
    s.code,
    e.tc,
    COUNT(DISTINCT s.id) AS endpoints
  FROM 
    mv_get_security_endpoints s
  JOIN 
    endpoints_by_version e ON s.fhir_version = e.fhir_version
  GROUP BY 
    s.fhir_version, s.code, e.tc
)
-- Calculate final results with percentages
SELECT 
  code AS "Code",
  fhir_version AS "FHIR Version",
  endpoints::integer AS "Endpoints",
  ROUND(endpoints::numeric * 100 / tc)::integer || '%' AS "Percent"
FROM 
  endpoints_by_version_code
ORDER BY 
  "FHIR Version",  
  "Code"; 

-- Create indexes for performance
CREATE UNIQUE INDEX idx_mv_auth_type_count ON mv_auth_type_count("Code", "FHIR Version");
CREATE INDEX idx_mv_auth_type_count_fhir ON mv_auth_type_count("FHIR Version");
CREATE INDEX idx_mv_auth_type_count_endpoints ON mv_auth_type_count("Endpoints");

DROP MATERIALIZED VIEW IF EXISTS mv_endpoint_security_counts CASCADE;

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

-- Create a unique index
CREATE UNIQUE INDEX idx_mv_endpoint_security_counts ON mv_endpoint_security_counts("Status");

COMMIT;