BEGIN;

DROP MATERIALIZED VIEW IF EXISTS mv_organizations_aggregated CASCADE;

CREATE MATERIALIZED VIEW mv_organizations_aggregated AS
WITH base_filtered_data AS (
    -- Step 1: Replicate the exact R filtering logic from get_endpoint_list_matches
    SELECT 
        mv.organization_name,
        mv.organization_id,
        mv.url,
        mv.fhir_version,
        mv.vendor_name
    FROM mv_endpoint_list_organizations mv
),
processed_data AS (
    -- Step 2: Apply the R mutate logic
    SELECT DISTINCT
        -- Replicate tidyr::replace_na(list(organization_name = "Unknown"))
        CASE 
            WHEN organization_name IS NULL OR organization_name = '' THEN 'Unknown'
            ELSE organization_name
        END AS organization_name,
        -- Replicate mutate(organization_id = as.integer(organization_id))
        CASE 
            WHEN organization_id IS NULL OR organization_id = '' OR organization_id = 'Unknown' THEN NULL
            WHEN organization_id ~ '^[0-9]+$' THEN organization_id::integer
            ELSE NULL
        END as org_id,
        url,
        -- Replicate the consistent FHIR version processing
        CASE 
            WHEN fhir_version = '' OR fhir_version IS NULL THEN 'No Cap Stat'
            WHEN position('-' in fhir_version) > 0 THEN 
                CASE
                    WHEN substring(fhir_version, 1, position('-' in fhir_version) - 1) IN 
                        ('0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2', '1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', 
                         '3.0.0', '3.0.1', '3.0.2', '3.2.0', '3.3.0', '3.5.0', '3.5a.0', '4.0.0', '4.0.1', 'No Cap Stat')
                    THEN substring(fhir_version, 1, position('-' in fhir_version) - 1)
                    ELSE 'Unknown'
                END
            WHEN fhir_version IN 
                ('0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2', '1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', 
                 '3.0.0', '3.0.1', '3.0.2', '3.2.0', '3.3.0', '3.5.0', '3.5a.0', '4.0.0', '4.0.1', 'No Cap Stat')
            THEN fhir_version
            ELSE 'Unknown'
        END AS fhir_version,
        vendor_name
    FROM base_filtered_data
),
-- Step 3: Get organization IDs per organization name 
org_ids_per_name AS (
    SELECT 
        organization_name,
        array_agg(DISTINCT org_id) FILTER (WHERE org_id IS NOT NULL) as org_ids_array
    FROM processed_data
    GROUP BY organization_name
),
-- Step 4: Get DISTINCT identifiers per organization (avoiding double aggregation)
identifiers_raw AS (
    SELECT DISTINCT
        opn.organization_name,
        fei.identifier
    FROM org_ids_per_name opn
    LEFT JOIN fhir_endpoint_organization_identifiers fei 
        ON (opn.org_ids_array IS NOT NULL 
            AND array_length(opn.org_ids_array, 1) > 0 
            AND fei.org_id = ANY(opn.org_ids_array))
    WHERE fei.identifier IS NOT NULL
),
identifiers_agg AS (
    SELECT 
        organization_name,
        string_agg(identifier, '<br/>') as identifiers_html,
        string_agg(identifier, E'\n') as identifiers_csv
    FROM identifiers_raw
    GROUP BY organization_name
),
-- Step 5: Get DISTINCT addresses per organization (avoiding double aggregation)
addresses_raw AS (
    SELECT DISTINCT
        opn.organization_name,
        UPPER(fea.address) as address
    FROM org_ids_per_name opn
    LEFT JOIN fhir_endpoint_organization_addresses fea 
        ON (opn.org_ids_array IS NOT NULL 
            AND array_length(opn.org_ids_array, 1) > 0 
            AND fea.org_id = ANY(opn.org_ids_array))
    WHERE fea.address IS NOT NULL
),
addresses_agg AS (
    SELECT 
        organization_name,
        string_agg(address, '<br/>') as addresses_html,
        string_agg(address, E'\n') as addresses_csv
    FROM addresses_raw
    GROUP BY organization_name
),
-- Step 6: Get DISTINCT org URLs per organization with urn:uuid filtering
urls_raw AS (
    SELECT DISTINCT
        opn.organization_name,
        -- FIXED: Apply the urn:uuid filtering 
        CASE 
            WHEN feou.org_url LIKE 'urn:uuid:%' THEN ''
            ELSE feou.org_url
        END as org_url
    FROM org_ids_per_name opn
    LEFT JOIN fhir_endpoint_organization_url feou 
        ON (opn.org_ids_array IS NOT NULL 
            AND array_length(opn.org_ids_array, 1) > 0 
            AND feou.org_id = ANY(opn.org_ids_array)
            AND feou.org_url IS NOT NULL 
            AND feou.org_url != '')
    WHERE feou.org_url IS NOT NULL AND feou.org_url != ''
),
urls_agg AS (
    SELECT 
        organization_name,
        string_agg(org_url, '<br/>') FILTER (WHERE org_url != '') as org_urls_html,
        string_agg(org_url, E'\n') FILTER (WHERE org_url != '') as org_urls_csv
    FROM urls_raw
    GROUP BY organization_name
),
-- Step 7: Replicate the R group_by/summarise logic 
endpoint_data_agg AS (
    SELECT 
        organization_name,
        -- HTML formatted endpoint URLs 
        string_agg(
            DISTINCT '<a class="lantern-url" tabindex="0" aria-label="Press enter to open a pop up modal containing additional information for this endpoint." onkeydown="javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)" onclick="Shiny.setInputValue(''endpoint_popup'',&quot;' || url || '&quot,{priority: ''event''});"> ' || url || '</a>',
            '<br/>'
        ) as endpoint_urls_html,
        string_agg(DISTINCT url, E'\n') as endpoint_urls_csv,
        string_agg(DISTINCT fhir_version, '<br/>') as fhir_versions_html,
        string_agg(DISTINCT fhir_version, E'\n') as fhir_versions_csv,
        string_agg(DISTINCT vendor_name, '<br/>') as vendor_names_html,
        string_agg(DISTINCT vendor_name, E'\n') as vendor_names_csv,
        -- Arrays for filtering (exactly as R code)
        ARRAY(SELECT DISTINCT unnest(array_agg(fhir_version))::text ORDER BY unnest)::text[] as fhir_versions_array,
        ARRAY(SELECT DISTINCT unnest(array_agg(vendor_name))::text ORDER BY unnest)::text[] as vendor_names_array,
        ARRAY(SELECT DISTINCT unnest(array_agg(url))::text ORDER BY unnest)::text[] as urls_array
    FROM processed_data
    GROUP BY organization_name
)
-- Step 8: Final select with the exact R filter logic
SELECT 
    eda.organization_name,
    -- For HTML display (pagination)
    COALESCE(ia.identifiers_html, '') as identifiers_html,
    COALESCE(aa.addresses_html, '') as addresses_html,
	eda.endpoint_urls_html,
    COALESCE(ua.org_urls_html, '') as org_urls_html,
    eda.fhir_versions_html,
    eda.vendor_names_html,
    
    -- For CSV export  
    COALESCE(ia.identifiers_csv, '') as identifiers_csv,
    COALESCE(aa.addresses_csv, '') as addresses_csv,
    eda.endpoint_urls_csv,
	COALESCE(ua.org_urls_csv, '') as org_urls_csv,
    eda.fhir_versions_csv,
    eda.vendor_names_csv,
    
    -- Arrays for filtering 
    eda.fhir_versions_array,
    eda.vendor_names_array,
    eda.urls_array
    
FROM endpoint_data_agg eda
LEFT JOIN identifiers_agg ia ON eda.organization_name = ia.organization_name
LEFT JOIN addresses_agg aa ON eda.organization_name = aa.organization_name  
LEFT JOIN urls_agg ua ON eda.organization_name = ua.organization_name
WHERE eda.organization_name != 'Unknown';

-- Create indexes for performance 
CREATE UNIQUE INDEX idx_mv_orgs_agg_name ON mv_organizations_aggregated(organization_name);
CREATE INDEX idx_mv_orgs_agg_fhir_versions ON mv_organizations_aggregated USING GIN(fhir_versions_array);
CREATE INDEX idx_mv_orgs_agg_vendor_names ON mv_organizations_aggregated USING GIN(vendor_names_array);
CREATE INDEX idx_mv_orgs_agg_urls ON mv_organizations_aggregated USING GIN(urls_array);

COMMIT;