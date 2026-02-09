BEGIN;

DROP MATERIALIZED VIEW IF EXISTS mv_organizations_final CASCADE;

CREATE MATERIALIZED VIEW mv_organizations_final AS
SELECT 
    ROW_NUMBER() OVER (ORDER BY organization_name) as org_id,
    organization_name,
    identifier_types_html,
    identifier_values_html,
    addresses_html,
    org_urls_html,
    -- Combine endpoint URLs where everything else matches
    string_agg(DISTINCT endpoint_urls_html, '<br/>') as endpoint_urls_html,
    fhir_versions_html,
    vendor_names_html,
    
    -- CSV versions
    identifier_types_csv,
    identifier_values_csv,
    addresses_csv,
    org_urls_csv,
    string_agg(DISTINCT endpoint_urls_csv, E'\n') as endpoint_urls_csv,
    fhir_versions_csv,
    vendor_names_csv,
    
    -- Arrays for filtering (combine from all matching rows) - FIXED: Use |||| delimiter instead of comma
    ARRAY(SELECT DISTINCT elem FROM unnest(string_to_array(string_agg(array_to_string(fhir_versions_array, '||||'), '||||'), '||||')) AS elem ORDER BY elem) as fhir_versions_array,
    ARRAY(SELECT DISTINCT elem FROM unnest(string_to_array(string_agg(array_to_string(vendor_names_array, '||||'), '||||'), '||||')) AS elem ORDER BY elem) as vendor_names_array,
    ARRAY(SELECT DISTINCT elem FROM unnest(string_to_array(string_agg(array_to_string(urls_array, '||||'), '||||'), '||||')) AS elem ORDER BY elem) as urls_array
    
FROM mv_organizations_aggregated
GROUP BY 
    organization_name,
    identifier_types_html,
    identifier_values_html,
    addresses_html,
    org_urls_html,
    fhir_versions_html,
    vendor_names_html,
    identifier_types_csv,
    identifier_values_csv,
    addresses_csv,
    org_urls_csv,
    fhir_versions_csv,
    vendor_names_csv
ORDER BY organization_name;

-- Create indexes for performance
CREATE UNIQUE INDEX idx_mv_orgs_final_org_id ON mv_organizations_final(org_id);
CREATE INDEX idx_mv_orgs_final_name ON mv_organizations_final(organization_name);
CREATE INDEX idx_mv_orgs_final_fhir_versions ON mv_organizations_final USING GIN(fhir_versions_array);
CREATE INDEX idx_mv_orgs_final_vendor_names ON mv_organizations_final USING GIN(vendor_names_array);
CREATE INDEX idx_mv_orgs_final_urls ON mv_organizations_final USING GIN(urls_array);

COMMIT;