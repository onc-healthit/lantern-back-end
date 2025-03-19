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

    -- Step 3: Validate against the correct set of FHIR versions
    CASE 
        WHEN p.capability_fhir_version = 'No Cap Stat' THEN 'No Cap Stat'  -- Ensure "No Cap Stat" is preserved
        WHEN p.fhir_version_raw IN ('1.0.2', '1.4.0', '3.0.1', '3.0.2', '4.0.0', '4.0.1', '4.3.0', '5.0.0') 
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

COMMIT;
