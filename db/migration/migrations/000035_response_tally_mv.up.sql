BEGIN;

CREATE MATERIALIZED VIEW response_tally_mv AS
WITH subquery AS (
    SELECT 
        fem.http_response,
        count(*) AS response_count
    FROM fhir_endpoints_info fei
    JOIN fhir_endpoints_metadata fem 
        ON fei.metadata_id = fem.id
    WHERE fei.requested_fhir_version::text = 'None'::text
    GROUP BY fem.http_response
)
SELECT 
    COALESCE(SUM(
        CASE 
            WHEN subquery.http_response = 200 THEN subquery.response_count 
            ELSE 0::bigint 
        END), 0::numeric) AS http_200,
    COALESCE(SUM(
        CASE 
            WHEN subquery.http_response <> 200 THEN subquery.response_count 
            ELSE 0::bigint 
        END), 0::numeric) AS http_non200,
    COALESCE(SUM(
        CASE 
            WHEN subquery.http_response = 404 THEN subquery.response_count 
            ELSE 0::bigint 
        END), 0::numeric) AS http_404,
    COALESCE(SUM(
        CASE 
            WHEN subquery.http_response = 503 THEN subquery.response_count 
            ELSE 0::bigint 
        END), 0::numeric) AS http_503
FROM subquery;


COMMIT