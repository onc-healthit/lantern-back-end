BEGIN;

DROP MATERIALIZED VIEW IF EXISTS mv_http_responses CASCADE;

CREATE MATERIALIZED VIEW mv_http_responses AS
WITH response_by_vendor AS (
    SELECT
        CASE
            WHEN v.name IS NULL OR v.name = '' THEN 'Unknown'
            ELSE v.name
        END AS vendor_name,
        m.http_response AS http_code,
        CASE 
            WHEN m.http_response = 200 THEN 'OK'
            WHEN m.http_response = 301 THEN 'Moved Permanently'
            WHEN m.http_response = 302 THEN 'Found'
            WHEN m.http_response = 400 THEN 'Bad Request'
            WHEN m.http_response = 401 THEN 'Unauthorized'
            WHEN m.http_response = 403 THEN 'Forbidden'
            WHEN m.http_response = 404 THEN 'Not Found'
            WHEN m.http_response = 500 THEN 'Internal Server Error'
            WHEN m.http_response = 503 THEN 'Service Unavailable'
            ELSE 'Other'
        END AS code_label,
        COUNT(DISTINCT f.url) AS count_endpoints
    FROM fhir_endpoints_info f
    LEFT JOIN vendors v
           ON f.vendor_id = v.id
    LEFT JOIN fhir_endpoints_metadata m
           ON f.metadata_id = m.id
    WHERE m.http_response IS NOT NULL
      AND f.requested_fhir_version = 'None'
    GROUP BY v.name, m.http_response
),
response_all_devs AS (
    SELECT
        'ALL_DEVELOPERS' AS vendor_name,
        http_code,
        code_label,
        SUM(count_endpoints) AS count_endpoints
    FROM response_by_vendor
    GROUP BY http_code, code_label
)
SELECT 
    now() AS aggregation_date,
    vendor_name,
    http_code,
    code_label,
    count_endpoints
FROM response_by_vendor

UNION ALL

SELECT
    now() AS aggregation_date,
    vendor_name,
    http_code,
    code_label,
    count_endpoints
FROM response_all_devs;

DROP INDEX IF EXISTS mv_http_responses_uniq;

CREATE UNIQUE INDEX mv_http_responses_uniq
  ON mv_http_responses (aggregation_date, vendor_name, http_code);

COMMIT;