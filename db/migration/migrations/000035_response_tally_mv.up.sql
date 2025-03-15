BEGIN;

DROP MATERIALIZED VIEW IF EXISTS mv_response_tally CASCADE;

CREATE MATERIALIZED VIEW mv_response_tally AS
WITH response_counts AS (
    SELECT 
        fem.http_response,
        count(*) AS response_count
    FROM fhir_endpoints_info fei
    JOIN fhir_endpoints_metadata fem 
        ON fei.metadata_id = fem.id
    WHERE fei.requested_fhir_version = 'None'
    GROUP BY fem.http_response
)
SELECT 
    COALESCE(SUM(
        CASE 
            WHEN http_response = 200 THEN response_count 
            ELSE 0 
        END), 0) AS http_200,
    COALESCE(SUM(
        CASE 
            WHEN http_response <> 200 THEN response_count 
            ELSE 0 
        END), 0) AS http_non200,
    COALESCE(SUM(
        CASE 
            WHEN http_response = 404 THEN response_count 
            ELSE 0 
        END), 0) AS http_404,
    COALESCE(SUM(
        CASE 
            WHEN http_response = 503 THEN response_count 
            ELSE 0 
        END), 0) AS http_503
FROM response_counts;

DROP INDEX IF EXISTS idx_mv_response_tally_http_code;
CREATE UNIQUE INDEX idx_mv_response_tally_http_code ON mv_response_tally(http_200);

COMMIT;