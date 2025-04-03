BEGIN;

DROP MATERIALIZED VIEW IF EXISTS security_endpoints_mv CASCADE;

CREATE MATERIALIZED VIEW security_endpoints_mv AS
WITH valid_fhir_versions AS (
    -- Dynamically extract all distinct valid FHIR versions from the dataset
    SELECT DISTINCT 
        CASE 
            WHEN fhir_version LIKE '%-%' THEN SPLIT_PART(fhir_version, '-', 1)
            ELSE fhir_version
        END AS version
    FROM endpoint_export
    WHERE fhir_version IS NOT NULL AND fhir_version != ''
)
SELECT 
    ROW_NUMBER() OVER () AS id,  -- Generate a unique sequential ID
	e.url,
    -- Completely remove ALL double quotes, matching the gsub operations in R
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
    -- Extract the code exactly like in the original query
    json_array_elements(json_array_elements(f.capability_statement::json#>'{rest,0,security,service}')->'coding')::json->>'code' AS code,
    -- Dynamically check against valid_fhir_versions
    CASE 
        WHEN (
            CASE 
                WHEN e.fhir_version LIKE '%-%' THEN SPLIT_PART(e.fhir_version, '-', 1)
                ELSE e.fhir_version 
            END
        ) IN (SELECT version FROM valid_fhir_versions) 
        THEN (
            CASE 
                WHEN e.fhir_version LIKE '%-%' THEN SPLIT_PART(e.fhir_version, '-', 1)
                ELSE e.fhir_version 
            END
        ) 
        ELSE 'Unknown' 
    END AS fhir_version_final
FROM endpoint_export e
JOIN fhir_endpoints_info f ON e.url = f.url
WHERE f.requested_fhir_version = 'None';

--indexing 
CREATE INDEX idx_security_endpoints_url ON security_endpoints_mv (url);
CREATE INDEX idx_security_endpoints_fhir_version ON security_endpoints_mv (fhir_version_final);
CREATE INDEX idx_security_endpoints_vendor_name ON security_endpoints_mv (vendor_name);
CREATE INDEX idx_security_endpoints_code ON security_endpoints_mv (code);

--unique index 
DROP INDEX IF EXISTS idx_unique_security_endpoints;
CREATE UNIQUE INDEX idx_unique_security_endpoints ON security_endpoints_mv (id, url, vendor_name, code);

COMMIT;