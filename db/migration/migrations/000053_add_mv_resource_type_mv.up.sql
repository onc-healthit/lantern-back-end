BEGIN;

-- Create materialized view for resource types
DROP MATERIALIZED VIEW IF EXISTS mv_endpoint_resource_types CASCADE;
CREATE MATERIALIZED VIEW mv_endpoint_resource_types AS
SELECT 
    f.id AS endpoint_id,
    f.vendor_id,
    COALESCE(vendors.name, 'Unknown') AS vendor_name,
    CASE 
        WHEN f.capability_fhir_version = '' THEN 'No Cap Stat'
        WHEN position('-' in f.capability_fhir_version) > 0 THEN substring(f.capability_fhir_version from 1 for position('-' in f.capability_fhir_version) - 1)
        WHEN f.capability_fhir_version IN ('0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2', '1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', '3.0.0', '3.0.1', '3.0.2', '3.2.0', '3.3.0', '3.5.0', '3.5a.0', '4.0.0', '4.0.1') 
            THEN f.capability_fhir_version
        ELSE 'Unknown'
    END AS fhir_version,
    json_array_elements(capability_statement::json#>'{rest,0,resource}') ->> 'type' AS type
FROM fhir_endpoints_info f
LEFT JOIN vendors ON f.vendor_id = vendors.id
WHERE f.requested_fhir_version = 'None'
ORDER BY type;

-- Create indexes for better performance
CREATE UNIQUE INDEX idx_mv_endpoint_resource_types_unique ON mv_endpoint_resource_types(endpoint_id, vendor_id, fhir_version, type);
CREATE INDEX idx_mv_endpoint_resource_types_vendor ON mv_endpoint_resource_types(vendor_name);
CREATE INDEX idx_mv_endpoint_resource_types_fhir ON mv_endpoint_resource_types(fhir_version);
CREATE INDEX idx_mv_endpoint_resource_types_type ON mv_endpoint_resource_types(type);

COMMIT;
