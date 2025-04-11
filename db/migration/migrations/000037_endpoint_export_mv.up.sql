BEGIN;

DROP MATERIALIZED VIEW IF EXISTS endpoint_export_mv CASCADE;

CREATE MATERIALIZED VIEW endpoint_export_mv AS
WITH endpoint_organizations AS (
    SELECT DISTINCT url, UNNEST(endpoint_names) AS endpoint_name
    FROM endpoint_export
),
grouped_organizations AS (
    SELECT url, 
           STRING_AGG(endpoint_name, '; ') AS endpoint_names 
    FROM endpoint_organizations
    WHERE endpoint_name IS NOT NULL AND endpoint_name <> 'NULL'
    GROUP BY url
),
processed_versions AS (
    SELECT 
        e.*,
        -- Step 1: Replace empty fhir_version with "No Cap Stat"
        CASE 
            WHEN e.fhir_version = '' THEN 'No Cap Stat'
            ELSE e.fhir_version
        END AS capability_fhir_version,
        -- Step 2: Extract version before "-" if present
        CASE 
            WHEN e.fhir_version = '' THEN 'No Cap Stat'
            WHEN POSITION('-' IN e.fhir_version) > 0 THEN SPLIT_PART(e.fhir_version, '-', 1)
            ELSE e.fhir_version
        END AS fhir_version_raw
    FROM endpoint_export e
)
SELECT 
    p.url, 
    p.list_source, 
    COALESCE(NULLIF(p.vendor_name, ''), 'Unknown') AS vendor_name,
    p.capability_fhir_version,
    -- Step 3: Use the fixed list of valid FHIR versions 
    CASE 
        WHEN p.capability_fhir_version = 'No Cap Stat' THEN 'No Cap Stat'  -- Ensure "No Cap Stat" is preserved
        WHEN p.fhir_version_raw IN ('No Cap Stat', '0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2', 
                                  '1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', '3.0.0', 
                                  '3.0.1', '3.0.2', '3.2.0', '3.3.0', '3.5.0', '3.5a.0', 
                                  '4.0.0', '4.0.1')
            THEN p.fhir_version_raw
        ELSE 'Unknown'  
    END AS fhir_version,
    p.tls_version,
    p.mime_types,
    p.http_response,
    p.response_time_seconds,
    p.smart_http_response,
    p.errors,
    p.cap_stat_exists,
    p.publisher,
    p.software_name,
    p.software_version,
    p.software_releasedate,
    REGEXP_REPLACE(p.format::TEXT, '[\[\]"]', '', 'g') AS format, 
    p.kind,
    p.info_updated,
    p.info_created,
    p.requested_fhir_version,
    p.availability,
    lsi.is_chpl,
    COALESCE(g.endpoint_names, '') AS endpoint_names
FROM processed_versions p
LEFT JOIN list_source_info lsi 
    ON p.list_source = lsi.list_source
LEFT JOIN grouped_organizations g 
    ON p.url = g.url;

-- Unique Index for refeshing the MV concurrently 
DROP INDEX IF EXISTS endpoint_export_mv_unique_idx;
CREATE UNIQUE INDEX endpoint_export_mv_unique_idx ON endpoint_export_mv (url, list_source, vendor_name, fhir_version, info_updated);


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
DROP INDEX IF EXISTS fhir_endpoint_comb_mv_unique_idx;
CREATE UNIQUE INDEX fhir_endpoint_comb_mv_unique_idx ON fhir_endpoint_comb_mv (id, url, list_source);


DROP MATERIALIZED VIEW IF EXISTS selected_fhir_endpoints_mv CASCADE;

-- Create the modified materialized view with an id column
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
DROP INDEX IF EXISTS idx_selected_fhir_endpoints_mv_unique;
CREATE UNIQUE INDEX idx_selected_fhir_endpoints_mv_unique ON selected_fhir_endpoints_mv(id, url, requested_fhir_version);

-- Create single column indexes to improve filtering performance
CREATE INDEX idx_selected_fhir_endpoints_mv_fhir_version ON selected_fhir_endpoints_mv(fhir_version);
CREATE INDEX idx_selected_fhir_endpoints_mv_vendor_name ON selected_fhir_endpoints_mv(vendor_name);
CREATE INDEX idx_selected_fhir_endpoints_mv_availability ON selected_fhir_endpoints_mv(availability);
CREATE INDEX idx_selected_fhir_endpoints_mv_is_chpl ON selected_fhir_endpoints_mv(is_chpl);

COMMIT;
