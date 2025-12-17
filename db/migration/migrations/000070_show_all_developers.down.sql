BEGIN;

ALTER TABLE fhir_endpoints_info DROP CONSTRAINT IF EXISTS fhir_endpoints_info_unique;

ALTER TABLE fhir_endpoints_info ADD CONSTRAINT fhir_endpoints_info_unique UNIQUE (url, requested_fhir_version);

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

CREATE UNIQUE INDEX idx_mv_response_tally_http_code ON mv_response_tally(http_200);

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

CREATE UNIQUE INDEX mv_http_responses_uniq
  ON mv_http_responses (aggregation_date, vendor_name, http_code);

CREATE INDEX mv_http_responses_vendor_name_idx
  ON mv_http_responses (vendor_name);

DROP MATERIALIZED VIEW IF EXISTS fhir_endpoint_comb_mv CASCADE;

CREATE MATERIALIZED VIEW fhir_endpoint_comb_mv AS 
SELECT 
    ROW_NUMBER() OVER () AS id,
    t.url,
    t.endpoint_names,
    t.info_created,
    t.info_updated,
    t.list_source,
    t.vendor_name,
    t.capability_fhir_version,
    t.fhir_version,
    t.format,
    t.http_response,
    t.response_time_seconds,
    t.smart_http_response,
    t.errors,
    t.availability,
    t.kind,
    t.requested_fhir_version,
    t.is_chpl,
    t.status,
    t.cap_stat_exists
FROM (
    SELECT DISTINCT ON (e.url, e.vendor_name, e.fhir_version, e.http_response, e.requested_fhir_version)
        e.url,
        e.endpoint_names,
        e.info_created,
        e.info_updated,
        e.list_source,
        e.vendor_name,
        e.capability_fhir_version,
        e.fhir_version,
        e.format,
        e.http_response,
        e.response_time_seconds,
        e.smart_http_response,
        e.errors,
        e.availability,
        e.kind,
        e.requested_fhir_version,
        lsi.is_chpl,
        CASE 
            WHEN e.http_response = 200 THEN CONCAT('Success: ', e.http_response, ' - ', r.code_label)
            WHEN e.http_response IS NULL OR e.http_response = 0 THEN 'Failure: 0 - NA'
            ELSE CONCAT('Failure: ', e.http_response, ' - ', r.code_label)
        END AS status,
        LOWER(CASE 
            WHEN e.kind != 'instance' THEN 'true*'::TEXT  
            ELSE e.cap_stat_exists::TEXT
        END) AS cap_stat_exists
    FROM endpoint_export_mv e
    LEFT JOIN mv_http_responses r ON e.http_response = r.http_code
    LEFT JOIN list_source_info lsi ON e.list_source = lsi.list_source
    ORDER BY e.url, e.vendor_name, e.fhir_version, e.http_response, e.requested_fhir_version
) t;

--Unique index for refreshing the MV concurrently
CREATE UNIQUE INDEX fhir_endpoint_comb_mv_unique_idx ON fhir_endpoint_comb_mv (id, url, list_source);

DROP MATERIALIZED VIEW IF EXISTS selected_fhir_endpoints_mv CASCADE;

CREATE MATERIALIZED VIEW selected_fhir_endpoints_mv AS
SELECT 
    ROW_NUMBER() OVER () AS id,  -- Generate a unique sequential ID
    e.url,
    e.endpoint_names,
    e.info_created,
    e.info_updated,
    e.list_source,
    e.vendor_name,
    e.capability_fhir_version,
    e.fhir_version,
    e.format,
    e.http_response,
    e.response_time_seconds,
    e.smart_http_response,
    e.errors,
    e.availability * 100 AS availability,
    e.kind,
    e.requested_fhir_version,
    lsi.is_chpl,
    e.status,
    e.cap_stat_exists,
    
    -- Generate URL modal link
    CONCAT('<a class="lantern-url" tabindex="0" aria-label="Press enter to open a pop-up modal containing additional information for this endpoint." 
            onkeydown="javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)" 
            onclick="Shiny.setInputValue(''endpoint_popup'',''', e.url, '&&', e.requested_fhir_version, ''',{priority: ''event''});">', e.url, '</a>') 
    AS "urlModal",

    -- Generate Condensed Endpoint Names
    CASE 
        WHEN e.endpoint_names IS NOT NULL 
             AND array_length(string_to_array(e.endpoint_names, ';'), 1) > 5
        THEN CONCAT(
            array_to_string(ARRAY(SELECT unnest(string_to_array(e.endpoint_names, ';')) LIMIT 5), '; '),
            '; <a class="lantern-url" tabindex="0" aria-label="Press enter to open a pop-up modal containing the endpoint''s entire list of API information source names." 
                onkeydown="javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)" 
                onclick="Shiny.setInputValue(''show_details'',''', e.url, ''',{priority: ''event''});"> Click For More... </a>'
        )
        ELSE e.endpoint_names
    END AS condensed_endpoint_names

FROM fhir_endpoint_comb_mv e
LEFT JOIN list_source_info lsi 
    ON e.list_source = lsi.list_source;

-- Create a unique composite index including the new id column
CREATE UNIQUE INDEX idx_selected_fhir_endpoints_mv_unique ON selected_fhir_endpoints_mv(id, url, requested_fhir_version);

-- Create single column indexes to improve filtering performance
CREATE INDEX idx_selected_fhir_endpoints_mv_fhir_version ON selected_fhir_endpoints_mv(fhir_version);
CREATE INDEX idx_selected_fhir_endpoints_mv_vendor_name ON selected_fhir_endpoints_mv(vendor_name);
CREATE INDEX idx_selected_fhir_endpoints_mv_availability ON selected_fhir_endpoints_mv(availability);
CREATE INDEX idx_selected_fhir_endpoints_mv_is_chpl ON selected_fhir_endpoints_mv(is_chpl);

DROP MATERIALIZED VIEW IF EXISTS mv_endpoint_totals CASCADE;

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

CREATE UNIQUE INDEX idx_mv_endpoint_totals_date ON mv_endpoint_totals(aggregation_date);

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

DROP MATERIALIZED VIEW IF EXISTS mv_resource_interactions CASCADE;

CREATE MATERIALIZED VIEW mv_resource_interactions AS
WITH expanded_resources AS (
  SELECT
    f.id AS endpoint_id,
    COALESCE(v.name, 'Unknown') AS vendor_name,
    CASE WHEN f.capability_fhir_version = '' THEN 'No Cap Stat'
         ELSE f.capability_fhir_version
    END AS fhir_version,

    -- Extract resource type from the JSONB structure
    resource_elem->>'type' AS resource_type,

    -- Extract individual operation names (this expands into multiple rows)
    COALESCE(interaction_elem->>'code', 'not specified') AS operation_name

  FROM fhir_endpoints_info f
  LEFT JOIN vendors v ON f.vendor_id = v.id

  -- Expand the "resource" array
  LEFT JOIN LATERAL json_array_elements((f.capability_statement->'rest')->0->'resource') resource_elem
    ON TRUE

	-- Expand the "interaction" array within each resource
  LEFT JOIN LATERAL json_array_elements(resource_elem->'interaction') interaction_elem
    ON TRUE
	
  WHERE f.requested_fhir_version = 'None'
),
aggregated_operations AS (
  SELECT
    vendor_name,
    fhir_version,
    resource_type,
	COUNT(DISTINCT endpoint_id) AS endpoint_count,
    -- Aggregate operations into an array
    ARRAY_AGG(DISTINCT operation_name) AS operations

  FROM expanded_resources
  GROUP BY vendor_name, fhir_version, resource_type
)
SELECT *
FROM aggregated_operations;

CREATE UNIQUE INDEX mv_resource_interactions_uniq
  ON mv_resource_interactions (
    vendor_name,
    fhir_version,
    resource_type,
    endpoint_count,
    operations
  );

CREATE INDEX mv_resource_interactions_vendor_name_idx
  ON mv_resource_interactions (vendor_name);

CREATE INDEX mv_resource_interactions_fhir_version_idx
  ON mv_resource_interactions (fhir_version);

CREATE INDEX mv_resource_interactions_resource_type_idx
  ON mv_resource_interactions (resource_type);

CREATE INDEX mv_resource_interactions_operations_idx
  ON mv_resource_interactions USING GIN (operations);

DROP INDEX IF EXISTS idx_mv_capstat_sizes_uniq;

CREATE UNIQUE INDEX idx_mv_capstat_sizes_uniq ON mv_capstat_sizes_tbl(url);

DROP MATERIALIZED VIEW IF EXISTS mv_validation_results_plot CASCADE;

CREATE MATERIALIZED VIEW mv_validation_results_plot AS
SELECT DISTINCT t.url,
t.fhir_version,
t.vendor_name,
t.rule_name,
t.valid,
t.expected,
t.actual,
t.comment,
t.reference
FROM ( SELECT COALESCE(vendors.name, 'Unknown'::character varying) AS vendor_name,
        f.url,
            CASE
                WHEN f.capability_fhir_version::text = ''::text THEN 'No Cap Stat'::character varying
                WHEN "position"(f.capability_fhir_version::text, '-'::text) > 0 THEN "substring"(f.capability_fhir_version::text, 1, "position"(f.capability_fhir_version::text, '-'::text) - 1)::character varying
                WHEN f.capability_fhir_version::text <> ALL (ARRAY['0.4.0'::character varying, '0.5.0'::character varying, '1.0.0'::character varying, '1.0.1'::character varying, '1.0.2'::character varying, '1.1.0'::character varying, '1.2.0'::character varying, '1.4.0'::character varying, '1.6.0'::character varying, '1.8.0'::character varying, '3.0.0'::character varying, '3.0.1'::character varying, '3.0.2'::character varying, '3.2.0'::character varying, '3.3.0'::character varying, '3.5.0'::character varying, '3.5a.0'::character varying, '4.0.0'::character varying, '4.0.1'::character varying]::text[]) THEN 'Unknown'::character varying
                ELSE f.capability_fhir_version
            END AS fhir_version,
        v.rule_name,
        v.valid,
        v.expected,
        v.actual,
        v.comment,
        v.reference,
        v.validation_result_id AS id,
        f.requested_fhir_version
        FROM fhir_endpoints_info f
            JOIN validations v ON f.validation_result_id = v.validation_result_id
            LEFT JOIN vendors ON f.vendor_id = vendors.id
        ORDER BY v.validation_result_id, v.rule_name) t;

CREATE UNIQUE INDEX mv_validation_results_plot_unique_idx 
ON mv_validation_results_plot(url, fhir_version, vendor_name, rule_name, valid, expected, actual);

CREATE INDEX mv_validation_results_plot_vendor_idx ON mv_validation_results_plot(vendor_name);
CREATE INDEX mv_validation_results_plot_fhir_idx ON mv_validation_results_plot(fhir_version);
CREATE INDEX mv_validation_results_plot_rule_idx ON mv_validation_results_plot(rule_name);
CREATE INDEX mv_validation_results_plot_valid_idx ON mv_validation_results_plot(valid);
CREATE INDEX mv_validation_results_plot_reference_idx ON mv_validation_results_plot(reference);

CREATE MATERIALIZED VIEW mv_validation_failures AS
SELECT fhir_version, url, expected, actual, vendor_name, rule_name, reference
FROM mv_validation_results_plot
WHERE valid = 'false';

CREATE UNIQUE INDEX mv_validation_failures_unique_idx ON mv_validation_failures(url, fhir_version, vendor_name, rule_name);
CREATE INDEX mv_validation_failures_url_idx ON mv_validation_failures(url);
CREATE INDEX mv_validation_failures_fhir_version_idx ON mv_validation_failures(fhir_version);
CREATE INDEX mv_validation_failures_vendor_name_idx ON mv_validation_failures(vendor_name);
CREATE INDEX mv_validation_failures_rule_name_idx ON mv_validation_failures(rule_name);
CREATE INDEX mv_validation_failures_reference_idx ON mv_validation_failures(reference);

DROP MATERIALIZED VIEW IF EXISTS security_endpoints_mv CASCADE;

CREATE MATERIALIZED VIEW security_endpoints_mv AS
SELECT 
    ROW_NUMBER() OVER () AS id,
    e.url,
    REPLACE(
        REPLACE(
            REPLACE(
                REPLACE(e.endpoint_names::TEXT, '{', ''), 
                '}', ''
            ), 
            '","', '; '
        ),
        '"', ''
    ) AS organization_names,
    COALESCE(e.vendor_name, 'Unknown') AS vendor_name,
    CASE 
        WHEN e.fhir_version = '' THEN 'No Cap Stat'
        ELSE e.fhir_version 
    END AS capability_fhir_version,
    e.tls_version,
    codes.code,
    CASE 
        -- First transform empty to "No Cap Stat"
        WHEN e.fhir_version = '' THEN 'No Cap Stat'
        -- Then handle version with dash
        WHEN e.fhir_version LIKE '%-%' THEN 
            CASE 
                WHEN SPLIT_PART(e.fhir_version, '-', 1) IN ('0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2', '1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', '3.0.0', '3.0.1', '3.0.2', '3.2.0', '3.3.0', '3.5.0', '3.5a.0', '4.0.0', '4.0.1') 
                THEN SPLIT_PART(e.fhir_version, '-', 1)
                ELSE 'Unknown'
            END
        -- Handle regular versions
        WHEN e.fhir_version IN ('0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2', '1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', '3.0.0', '3.0.1', '3.0.2', '3.2.0', '3.3.0', '3.5.0', '3.5a.0', '4.0.0', '4.0.1') 
        THEN e.fhir_version
        ELSE 'Unknown'
    END AS fhir_version_final
FROM endpoint_export e
JOIN fhir_endpoints_info f ON e.url = f.url
JOIN LATERAL (
    SELECT json_array_elements(json_array_elements(f.capability_statement::json#>'{rest,0,security,service}')->'coding')::json->>'code' AS code
) codes ON true
WHERE f.requested_fhir_version = 'None';

--indexing 
CREATE INDEX idx_security_endpoints_url ON security_endpoints_mv (url);
CREATE INDEX idx_security_endpoints_fhir_version ON security_endpoints_mv (fhir_version_final);
CREATE INDEX idx_security_endpoints_vendor_name ON security_endpoints_mv (vendor_name);
CREATE INDEX idx_security_endpoints_code ON security_endpoints_mv (code);
--unique index
CREATE UNIQUE INDEX idx_unique_security_endpoints ON security_endpoints_mv (id, url, vendor_name, code);

CREATE MATERIALIZED VIEW selected_security_endpoints_mv AS
SELECT 
    se.id,
    se.url,
    se.organization_names,
    se.vendor_name,
    se.capability_fhir_version,
    se.fhir_version_final AS fhir_version,
    se.tls_version,
    se.code,
    -- Create the condensed_organization_names with the modal link for endpoints with more than 5 organizations
    CASE 
        WHEN se.organization_names IS NOT NULL AND 
             array_length(string_to_array(se.organization_names, ';'), 1) > 5 
        THEN 
            CONCAT(
                array_to_string(
                    ARRAY(
                        SELECT unnest(string_to_array(se.organization_names, ';')) 
                        LIMIT 5
                    ), 
                    '; '
                ),
                '; <a class="lantern-url" tabindex="0" aria-label="Press enter to open a pop up modal containing the endpoint''s entire list of API information source names." onkeydown="javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)" onclick="Shiny.setInputValue(''show_details'',''', 
                se.url, 
                ''',{priority: ''event''});"> Click For More... </a>'
            )
        ELSE 
            se.organization_names 
    END AS condensed_organization_names,
    
    -- Create the URL with modal functionality
    CONCAT(
        '<a class="lantern-url" tabindex="0" aria-label="Press enter to open a pop up modal containing additional information for this endpoint." onkeydown="javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)" onclick="Shiny.setInputValue(''endpoint_popup'',''', 
        se.url, 
        '&&None'',{priority: ''event''});">', 
        se.url, 
        '</a>'
    ) AS url_modal
FROM 
    security_endpoints_mv se;

-- Add indexing for better performance
CREATE INDEX idx_selected_security_endpoints_fhir_version ON selected_security_endpoints_mv (fhir_version);
CREATE INDEX idx_selected_security_endpoints_vendor_name ON selected_security_endpoints_mv (vendor_name);
CREATE INDEX idx_selected_security_endpoints_code ON selected_security_endpoints_mv (code);
-- Create a unique composite index
CREATE UNIQUE INDEX idx_unique_selected_security_endpoints ON selected_security_endpoints_mv (id, url, code);

DROP MATERIALIZED VIEW IF EXISTS security_endpoints_distinct_mv CASCADE;
CREATE MATERIALIZED VIEW security_endpoints_distinct_mv AS
SELECT DISTINCT
  url_modal AS url,
  condensed_organization_names,
  vendor_name,
  capability_fhir_version,
  tls_version,
  code
FROM selected_security_endpoints_mv;

-- Create indexes for security_endpoints_distinct_mv
CREATE UNIQUE INDEX idx_unique_security_endpoints_distinct_mv ON security_endpoints_distinct_mv (url, condensed_organization_names, vendor_name, capability_fhir_version, tls_version, code);
CREATE INDEX idx_security_endpoints_distinct_filters  ON security_endpoints_distinct_mv(capability_fhir_version, code, vendor_name);

DROP MATERIALIZED VIEW IF EXISTS mv_http_pct CASCADE;

CREATE MATERIALIZED VIEW mv_http_pct AS
WITH grouped AS (
  SELECT
    f.id,
    f.url,
    e.http_response,
    e.vendor_name,
    e.fhir_version,
    CAST(e.http_response AS text) AS code,
    COUNT(*) AS cnt
  FROM fhir_endpoints_info f
  LEFT JOIN mv_endpoint_export_tbl e ON f.url = e.url
  WHERE f.requested_fhir_version = 'None'
  GROUP BY f.id, f.url, e.http_response, e.vendor_name, e.fhir_version
)
SELECT
  row_number() OVER () AS mv_id,
  id,
  url,
  http_response,
  COALESCE(vendor_name, 'Unknown') AS vendor_name,
  fhir_version,
  code,
  cnt * 100.0 / SUM(cnt) OVER (PARTITION BY id) AS Percentage
FROM grouped;

-- Create indexes for mv_http_pct
CREATE UNIQUE INDEX idx_mv_http_pct_unique_id ON mv_http_pct(mv_id);
CREATE INDEX idx_mv_http_pct_http_response ON mv_http_pct (http_response);
CREATE INDEX idx_mv_http_pct_vendor ON mv_http_pct (vendor_name);
CREATE INDEX idx_mv_http_pct_fhir ON mv_http_pct (fhir_version);
CREATE INDEX idx_mv_http_pct_vendor_fhir ON mv_http_pct (vendor_name, fhir_version);

DROP MATERIALIZED VIEW IF EXISTS mv_well_known_endpoints CASCADE;
CREATE MATERIALIZED VIEW mv_well_known_endpoints AS

WITH base AS (
         SELECT e.url,
            array_to_string(e.endpoint_names, ';'::text) AS organization_names,
            COALESCE(e.vendor_name, 'Unknown'::character varying) AS vendor_name,
                CASE
                    WHEN e.fhir_version::text = ''::text THEN 'No Cap Stat'::character varying
                    ELSE e.fhir_version
                END AS capability_fhir_version
           FROM endpoint_export e
             LEFT JOIN fhir_endpoints_info f ON e.url::text = f.url::text
             LEFT JOIN fhir_endpoints_metadata m ON f.metadata_id = m.id
             LEFT JOIN vendors v ON f.vendor_id = v.id
          WHERE m.smart_http_response = 200 AND f.requested_fhir_version::text = 'None'::text AND jsonb_typeof(f.smart_response::jsonb) = 'object'::text
        )
 SELECT 
   	row_number() OVER () AS mv_id,
	base.url,
    regexp_replace(regexp_replace(regexp_replace(base.organization_names, '[{}]'::text, ''::text, 'g'::text), '","'::text, '; '::text, 'g'::text), '"'::text, ''::text, 'g'::text) AS organization_names,
    base.vendor_name,
    base.capability_fhir_version,
        CASE
            WHEN
            CASE
                WHEN base.capability_fhir_version::text ~~ '%-%'::text THEN split_part(base.capability_fhir_version::text, '-'::text, 1)::character varying
                ELSE base.capability_fhir_version
            END::text = ANY (ARRAY['No Cap Stat'::character varying, '0.4.0'::character varying, '0.5.0'::character varying, '1.0.0'::character varying, '1.0.1'::character varying, '1.0.2'::character varying, '1.1.0'::character varying, '1.2.0'::character varying, '1.4.0'::character varying, '1.6.0'::character varying, '1.8.0'::character varying, '3.0.0'::character varying, '3.0.1'::character varying, '3.0.2'::character varying, '3.2.0'::character varying, '3.3.0'::character varying, '3.5.0'::character varying, '3.5a.0'::character varying, '4.0.0'::character varying, '4.0.1'::character varying]::text[]) THEN
            CASE
                WHEN base.capability_fhir_version::text ~~ '%-%'::text THEN split_part(base.capability_fhir_version::text, '-'::text, 1)::character varying
                ELSE base.capability_fhir_version
            END
            ELSE 'Unknown'::character varying
        END AS fhir_version
   FROM base;

-- Create indexes for mv_well_known_endpoints
CREATE UNIQUE INDEX idx_mv_well_known_unique_id ON mv_well_known_endpoints(mv_id);
CREATE INDEX idx_mv_well_known_vendor ON mv_well_known_endpoints(vendor_name);
CREATE INDEX idx_mv_well_known_fhir ON mv_well_known_endpoints(fhir_version);
CREATE INDEX idx_mv_well_known_vendor_fhir ON mv_well_known_endpoints(vendor_name, fhir_version);

DROP MATERIALIZED VIEW IF EXISTS mv_selected_endpoints CASCADE;
CREATE MATERIALIZED VIEW mv_selected_endpoints AS
WITH original AS (
 SELECT
 	DISTINCT mv_well_known_endpoints.url,
        CASE
            WHEN mv_well_known_endpoints.organization_names IS NULL OR mv_well_known_endpoints.organization_names = ''::text THEN mv_well_known_endpoints.organization_names
            ELSE
            CASE
                WHEN cardinality(string_to_array(mv_well_known_endpoints.organization_names, ';'::text)) > 5 THEN (((array_to_string(( SELECT array_agg(t.elem) AS array_agg
                   FROM unnest(string_to_array(mv_well_known_endpoints.organization_names, ';'::text)) WITH ORDINALITY t(elem, ord)
                  WHERE t.ord <= 5), ';'::text) || '; '::text) || '<a class="lantern-url" tabindex="0" aria-label="Press enter to open a pop up modal containing the endpoint''s entire list of API information source names." onkeydown="javascript:(function(event) { if (event.keyCode === 13){event.target.click();}})(event)" onclick="Shiny.setInputValue(''show_details'','''::text) || mv_well_known_endpoints.url::text) || ''',{priority: ''event''});"> Click For More... </a>'::text
                ELSE mv_well_known_endpoints.organization_names
            END
        END AS condensed_organization_names,
    mv_well_known_endpoints.vendor_name,
    mv_well_known_endpoints.capability_fhir_version
 FROM mv_well_known_endpoints)
 SELECT 
   row_number() OVER (ORDER BY url) AS mv_id,
   *
 FROM original;

-- Create indexes for mv_selected_endpoints
CREATE UNIQUE INDEX idx_mv_selected_endpoints_unique_id ON mv_selected_endpoints(mv_id);
CREATE INDEX idx_mv_selected_endpoints_vendor ON mv_selected_endpoints(vendor_name);
CREATE INDEX idx_mv_selected_endpoints_fhir ON mv_selected_endpoints(capability_fhir_version);
CREATE INDEX idx_mv_selected_endpoints_vendor_fhir ON mv_selected_endpoints(vendor_name, capability_fhir_version);

COMMIT;