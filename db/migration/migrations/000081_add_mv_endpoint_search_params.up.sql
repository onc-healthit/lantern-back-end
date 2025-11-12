BEGIN;

-- Create materialized view for mv_endpoint_search_params
DROP MATERIALIZED VIEW IF EXISTS mv_endpoint_search_params CASCADE;

CREATE MATERIALIZED VIEW mv_endpoint_search_params AS
SELECT DISTINCT
    f.id AS endpoint_id,
    f.url,
    f.vendor_id,
    COALESCE(v.name, 'Unknown') AS vendor_name,
    CASE 
        WHEN f.capability_fhir_version = '' THEN 'No Cap Stat'
        WHEN position('-' in f.capability_fhir_version) > 0 THEN substring(f.capability_fhir_version from 1 for position('-' in f.capability_fhir_version) - 1)
        WHEN f.capability_fhir_version IN (
            '0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2', '1.1.0', '1.2.0', '1.4.0',
            '1.6.0', '1.8.0', '3.0.0', '3.0.1', '3.0.2', '3.2.0', '3.3.0', '3.5.0',
            '3.5a.0', '4.0.0', '4.0.1'
        ) THEN f.capability_fhir_version
        ELSE 'Unknown'
    END AS fhir_version,
    json_array_elements(f.capability_statement::json #> '{rest,0,resource}') ->> 'type' AS resource_type,
    json_array_elements(
        json_array_elements(f.capability_statement::json #> '{rest,0,resource}') -> 'searchParam'
    ) ->> 'name' AS search_param
FROM fhir_endpoints_info f
LEFT JOIN vendors v ON f.vendor_id = v.id
WHERE f.requested_fhir_version = 'None'
ORDER BY resource_type, search_param;

-- Create indexes for mv_endpoint_search_params
CREATE UNIQUE INDEX idx_mv_endpoint_search_params_unique ON mv_endpoint_search_params(endpoint_id, resource_type, search_param);
CREATE INDEX idx_mv_endpoint_search_params_url ON mv_endpoint_search_params(url);
CREATE INDEX idx_mv_endpoint_search_params_vendor ON mv_endpoint_search_params(vendor_name);
CREATE INDEX idx_mv_endpoint_search_params_fhir ON mv_endpoint_search_params(fhir_version);
CREATE INDEX idx_mv_endpoint_search_params_resource ON mv_endpoint_search_params(resource_type);
CREATE INDEX idx_mv_endpoint_search_params_param ON mv_endpoint_search_params(search_param);

COMMIT;