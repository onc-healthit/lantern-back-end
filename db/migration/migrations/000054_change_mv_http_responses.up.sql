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
            WHEN m.http_response = 100 THEN 'Continue'
            WHEN m.http_response = 101 THEN 'Switching Protocols'
            WHEN m.http_response = 102 THEN 'Processing'
            WHEN m.http_response = 103 THEN 'Early Hints'
            WHEN m.http_response = 200 THEN 'OK'
            WHEN m.http_response = 201 THEN 'Created'
            WHEN m.http_response = 202 THEN 'Accepted'
            WHEN m.http_response = 203 THEN 'Non-Authoritative Information'
            WHEN m.http_response = 204 THEN 'No Content'
            WHEN m.http_response = 205 THEN 'Reset Content'
            WHEN m.http_response = 206 THEN 'Partial Content'
            WHEN m.http_response = 207 THEN 'Multi-Status'
            WHEN m.http_response = 208 THEN 'Already Reported'
            WHEN m.http_response = 226 THEN 'IM Used'
            WHEN m.http_response = 300 THEN 'Multiple Choices'
            WHEN m.http_response = 301 THEN 'Moved Permanently'
            WHEN m.http_response = 302 THEN 'Found'
            WHEN m.http_response = 303 THEN 'See Other'
            WHEN m.http_response = 304 THEN 'Not Modified'
            WHEN m.http_response = 305 THEN 'Use Proxy'
            WHEN m.http_response = 306 THEN 'Switch Proxy'
            WHEN m.http_response = 307 THEN 'Temporary Redirect'
            WHEN m.http_response = 308 THEN 'Permanent Redirect'
            WHEN m.http_response = 400 THEN 'Bad Request'
            WHEN m.http_response = 401 THEN 'Unauthorized'
            WHEN m.http_response = 402 THEN 'Payment Required'
            WHEN m.http_response = 403 THEN 'Forbidden'
            WHEN m.http_response = 404 THEN 'Not Found'
            WHEN m.http_response = 405 THEN 'Method Not Allowed'
            WHEN m.http_response = 406 THEN 'Not Acceptable'
            WHEN m.http_response = 407 THEN 'Proxy Authentication Required'
            WHEN m.http_response = 408 THEN 'Request Timeout'
            WHEN m.http_response = 409 THEN 'Conflict'
            WHEN m.http_response = 410 THEN 'Gone'
            WHEN m.http_response = 411 THEN 'Length Required'
            WHEN m.http_response = 412 THEN 'Precondition Failed'
            WHEN m.http_response = 413 THEN 'Payload Too Large'
            WHEN m.http_response = 414 THEN 'Request URI Too Long'
            WHEN m.http_response = 415 THEN 'Unsupported Media Type'
            WHEN m.http_response = 416 THEN 'Requested Range Not Satisfiable'
            WHEN m.http_response = 417 THEN 'Expectation Failed'
            WHEN m.http_response = 418 THEN 'I''m a teapot'
            WHEN m.http_response = 421 THEN 'Misdirected Request'
            WHEN m.http_response = 422 THEN 'Unprocessable Entity'
            WHEN m.http_response = 423 THEN 'Locked'
            WHEN m.http_response = 424 THEN 'Failed Dependency'
            WHEN m.http_response = 425 THEN 'Too Early'
            WHEN m.http_response = 426 THEN 'Upgrade Required'
            WHEN m.http_response = 428 THEN 'Precondition Required'
            WHEN m.http_response = 429 THEN 'Too Many Requests'
            WHEN m.http_response = 431 THEN 'Request Header Fields Too Large'
            WHEN m.http_response = 451 THEN 'Unavailable for Legal Reasons'
            WHEN m.http_response = 500 THEN 'Internal Server Error'
            WHEN m.http_response = 501 THEN 'Not Implemented'
            WHEN m.http_response = 502 THEN 'Bad Gateway'
            WHEN m.http_response = 503 THEN 'Service Unavailable'
            WHEN m.http_response = 504 THEN 'Gateway Timeout'
            WHEN m.http_response = 505 THEN 'HTTP Version Not Supported'
            WHEN m.http_response = 506 THEN 'Variant Also Negotiates'
            WHEN m.http_response = 507 THEN 'Insufficient Storage'
            WHEN m.http_response = 508 THEN 'Loop Detected'
            WHEN m.http_response = 509 THEN 'Bandwidth Limit Exceeded'
            WHEN m.http_response = 510 THEN 'Not Extended'
            WHEN m.http_response = 511 THEN 'Network Authentication Required'
            ELSE 'Other'
        END AS code_label,
        -- Cast the count to an integer
        COUNT(DISTINCT f.url)::INTEGER AS count_endpoints
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
        -- Cast the sum to an integer to ensure it's not displayed as a decimal
        SUM(count_endpoints)::INTEGER AS count_endpoints
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

DROP INDEX IF EXISTS mv_http_responses_vendor_name_idx;

CREATE INDEX mv_http_responses_vendor_name_idx
  ON mv_http_responses (vendor_name);

COMMIT;