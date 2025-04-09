BEGIN;

-- Create materialized view for endpoint list organizations
DROP MATERIALIZED VIEW IF EXISTS mv_endpoint_list_organizations CASCADE;

CREATE MATERIALIZED VIEW mv_endpoint_list_organizations AS
SELECT DISTINCT
    url,

    -- Handle NULL and empty values in endpoint_names, trim spaces properly
    UNNEST(
        CASE 
            WHEN endpoint_names IS NULL THEN ARRAY['Unknown']  -- Convert NULL to "Unknown"
            ELSE ARRAY(
                SELECT TRIM(REGEXP_REPLACE(UNNEST(string_to_array(REGEXP_REPLACE(elem, '["]', '', 'g'), ';')), '\s+', ' ', 'g'))
                FROM UNNEST(endpoint_names) AS elem
            )
        END
    ) AS organization_name,

    -- Replace empty fhir_version values with "No Cap Stat"
    CASE 
        WHEN fhir_version = '' THEN 'No Cap Stat'
        ELSE fhir_version
    END AS fhir_version,

    -- Replace NULL vendor_name with "Unknown"
    COALESCE(vendor_name, 'Unknown') AS vendor_name,

    requested_fhir_version
FROM endpoint_export
ORDER BY organization_name, url;

-- Create indexes for endpoint list organizations materialized view
CREATE UNIQUE INDEX idx_mv_endpoint_list_org_uniq ON mv_endpoint_list_organizations(fhir_version, vendor_name, url, organization_name, requested_fhir_version);
CREATE INDEX idx_mv_endpoint_list_org_fhir ON mv_endpoint_list_organizations(fhir_version);
CREATE INDEX idx_mv_endpoint_list_org_vendor ON mv_endpoint_list_organizations(vendor_name);
CREATE INDEX idx_mv_endpoint_list_org_url ON mv_endpoint_list_organizations(url);

COMMIT;