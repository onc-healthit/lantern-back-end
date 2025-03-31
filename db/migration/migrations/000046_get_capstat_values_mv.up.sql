BEGIN;

DROP MATERIALIZED VIEW IF EXISTS get_capstat_values_mv CASCADE;

CREATE MATERIALIZED VIEW get_capstat_values_mv AS
WITH valid_fhir_versions AS (
    -- Dynamically extract all distinct FHIR versions from the dataset
    SELECT DISTINCT 
        CASE 
            WHEN capability_fhir_version LIKE '%-%' THEN SPLIT_PART(capability_fhir_version, '-', 1)
            ELSE capability_fhir_version
        END AS version
    FROM fhir_endpoints_info
    WHERE capability_fhir_version IS NOT NULL
)
SELECT 
    f.id AS endpoint_id,
    f.vendor_id,
    COALESCE(vendors.name, 'Unknown') AS vendor_name,
    CASE 
        WHEN f.capability_fhir_version = '' THEN 'No Cap Stat' 
        ELSE f.capability_fhir_version 
    END AS fhir_version,
    -- Extract the major version dynamically
    CASE 
        WHEN f.capability_fhir_version = '' THEN 'No Cap Stat'
        WHEN f.capability_fhir_version LIKE '%-%' THEN SPLIT_PART(f.capability_fhir_version, '-', 1)
        ELSE f.capability_fhir_version 
    END AS raw_filter_fhir_version,
    -- Check dynamically against extracted valid FHIR versions
    CASE 
        WHEN f.capability_fhir_version = '' THEN 'No Cap Stat'
        WHEN (
            CASE 
                WHEN f.capability_fhir_version LIKE '%-%' THEN SPLIT_PART(f.capability_fhir_version, '-', 1)
                ELSE f.capability_fhir_version 
            END
        ) IN (SELECT version FROM valid_fhir_versions) 
        THEN (
            CASE 
                WHEN f.capability_fhir_version LIKE '%-%' THEN SPLIT_PART(f.capability_fhir_version, '-', 1)
                ELSE f.capability_fhir_version 
            END
        ) 
        ELSE 'Unknown' 
    END AS filter_fhir_version,
    f.capability_statement->>'url' AS url,
    f.capability_statement->>'version' AS version,
    f.capability_statement->>'name' AS name,
    f.capability_statement->>'title' AS title,
    f.capability_statement->>'date' AS date,
    f.capability_statement->>'publisher' AS publisher,
    f.capability_statement->>'description' AS description,
    f.capability_statement->>'purpose' AS purpose,
    f.capability_statement->>'copyright' AS copyright,
    f.capability_statement->'software'->>'name' AS software_name,
    f.capability_statement->'software'->>'version' AS software_version,
    f.capability_statement->'software'->>'releaseDate' AS software_release_date,
    f.capability_statement->'implementation'->>'description' AS implementation_description,
    f.capability_statement->'implementation'->>'url' AS implementation_url,
    f.capability_statement->'implementation'->>'custodian' AS implementation_custodian
FROM fhir_endpoints_info f
LEFT JOIN vendors ON f.vendor_id = vendors.id
WHERE f.capability_statement::jsonb IS NOT NULL
AND f.requested_fhir_version = 'None';

-- Create indexes for performance optimization
CREATE INDEX idx_get_capstat_values_mv_endpoint_id ON get_capstat_values_mv(endpoint_id);
CREATE INDEX idx_get_capstat_values_mv_vendor_id ON get_capstat_values_mv(vendor_id);
CREATE INDEX idx_get_capstat_values_mv_filter_fhir_version ON get_capstat_values_mv(filter_fhir_version);
CREATE INDEX idx_get_capstat_values_mv_vendor_name ON get_capstat_values_mv(vendor_name);

-- Create a unique composite index
DROP INDEX IF EXISTS idx_get_capstat_values_mv_unique;
CREATE UNIQUE INDEX idx_get_capstat_values_mv_unique ON get_capstat_values_mv(endpoint_id, vendor_id, filter_fhir_version);

COMMIT;