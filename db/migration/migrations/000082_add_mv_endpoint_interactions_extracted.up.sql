BEGIN;

-- Create materialized view for mv_endpoint_interactions_extracted
DROP MATERIALIZED VIEW IF EXISTS mv_endpoint_interactions_extracted CASCADE;

CREATE MATERIALIZED VIEW mv_endpoint_interactions_extracted AS
SELECT DISTINCT
    f.id AS endpoint_id,
    f.url,
    f.vendor_id,
    COALESCE(v.name, 'Unknown') AS vendor_name,
    CASE
        WHEN f.capability_fhir_version = '' THEN 'No Cap Stat'
        WHEN position('-' in f.capability_fhir_version) > 0 THEN
            substring(f.capability_fhir_version from 1 for position('-' in f.capability_fhir_version) - 1)
        ELSE f.capability_fhir_version
    END AS fhir_version,
    json_array_elements(f.capability_statement #> '{rest,0,resource}') ->> 'type' AS resource_type,
    json_array_elements(
        json_array_elements(f.capability_statement #> '{rest,0,resource}') -> 'interaction'
    ) ->> 'code' AS interaction_code
FROM fhir_endpoints_info f
LEFT JOIN vendors v ON f.vendor_id = v.id
WHERE f.capability_statement IS NOT NULL
ORDER BY f.id;

-- Create indexes for mv_endpoint_interactions_extracted
CREATE UNIQUE INDEX idx_mv_endpoint_interactions_extracted_unique ON mv_endpoint_interactions_extracted (endpoint_id, resource_type, interaction_code);
CREATE INDEX idx_mv_endpoint_interactions_extracted_vendor ON mv_endpoint_interactions_extracted (vendor_name);
CREATE INDEX idx_mv_endpoint_interactions_extracted_fhir ON mv_endpoint_interactions_extracted (fhir_version);
CREATE INDEX idx_mv_endpoint_interactions_extracted_url ON mv_endpoint_interactions_extracted (url);

COMMIT;