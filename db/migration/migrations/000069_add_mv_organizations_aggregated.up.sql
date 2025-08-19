BEGIN;

DROP MATERIALIZED VIEW IF EXISTS mv_organizations_aggregated CASCADE;

CREATE MATERIALIZED VIEW mv_organizations_aggregated AS
WITH base_filtered_data AS (
    -- Step 1: Get the source data from mv_endpoint_list_organizations
    SELECT 
        mv.organization_name,
        mv.organization_id,
        mv.url,
        mv.fhir_version,
        mv.vendor_name
    FROM mv_endpoint_list_organizations mv
),
processed_data AS (
    -- Step 2: Apply the R mutate logic including uppercase conversion
    SELECT DISTINCT
        -- Replicate tidyr::replace_na(list(organization_name = "Unknown")) + UPPER()
        CASE 
            WHEN organization_name IS NULL OR organization_name = '' THEN 'UNKNOWN'
            ELSE UPPER(organization_name)
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
    WHERE organization_id IS NOT NULL AND organization_id != '' AND organization_id != 'Unknown'
),
-- Step 3: Get DISTINCT identifiers per organization ID
identifiers_raw AS (
    SELECT DISTINCT
        pd.org_id,
        fei.identifier
    FROM processed_data pd
    LEFT JOIN fhir_endpoint_organization_identifiers fei ON pd.org_id = fei.org_id
    WHERE fei.identifier IS NOT NULL
),
identifiers_agg AS (
    SELECT 
        org_id,
        string_agg(identifier, '<br/>') as identifiers_html,
        -- Truncate at complete lines to prevent CSV corruption
        CASE 
            WHEN LENGTH(string_agg(identifier, E'\n')) <= 32765 
            THEN string_agg(identifier, E'\n')
            ELSE 
                -- Find the last complete line within 32765 chars
                LEFT(
                    string_agg(identifier, E'\n'), 
                    GREATEST(
                        1,
                        CASE 
                            WHEN POSITION(E'\n' IN REVERSE(LEFT(string_agg(identifier, E'\n'), 32765))) > 0
                            THEN 32765 - POSITION(E'\n' IN REVERSE(LEFT(string_agg(identifier, E'\n'), 32765))) + 1
                            ELSE 32765
                        END
                    )
                )
        END as identifiers_csv
    FROM identifiers_raw
    GROUP BY org_id
),
-- Step 4: Get DISTINCT addresses per organization ID
addresses_raw AS (
    SELECT DISTINCT
        pd.org_id,
        UPPER(fea.address) as address
    FROM processed_data pd
    LEFT JOIN fhir_endpoint_organization_addresses fea ON pd.org_id = fea.org_id
    WHERE fea.address IS NOT NULL
),
addresses_agg AS (
    SELECT 
        org_id,
        string_agg(address, '<br/>') as addresses_html,
        -- Truncate at complete lines to prevent CSV corruption
        CASE 
            WHEN LENGTH(string_agg(address, E'\n')) <= 32765 
            THEN string_agg(address, E'\n')
            ELSE 
                -- Find the last complete line within 32765 chars
                LEFT(
                    string_agg(address, E'\n'), 
                    GREATEST(
                        1,
                        CASE 
                            WHEN POSITION(E'\n' IN REVERSE(LEFT(string_agg(address, E'\n'), 32765))) > 0
                            THEN 32765 - POSITION(E'\n' IN REVERSE(LEFT(string_agg(address, E'\n'), 32765))) + 1
                            ELSE 32765
                        END
                    )
                )
        END as addresses_csv
    FROM addresses_raw
    GROUP BY org_id
),
-- Step 5: Get DISTINCT org URLs per organization ID with urn:uuid filtering
urls_raw AS (
    SELECT DISTINCT
        pd.org_id,
        -- Apply the urn:uuid filtering 
        CASE 
            WHEN feou.org_url LIKE 'urn:uuid:%' THEN ''
            ELSE feou.org_url
        END as org_url
    FROM processed_data pd
    LEFT JOIN fhir_endpoint_organization_url feou ON pd.org_id = feou.org_id
    WHERE feou.org_url IS NOT NULL AND feou.org_url != ''
),
urls_agg AS (
    SELECT 
        org_id,
        string_agg(org_url, '<br/>') FILTER (WHERE org_url != '') as org_urls_html,
        -- Truncate at complete lines to prevent CSV corruption
        CASE 
            WHEN LENGTH(string_agg(org_url, E'\n') FILTER (WHERE org_url != '')) <= 32765 
            THEN string_agg(org_url, E'\n') FILTER (WHERE org_url != '')
            ELSE 
                -- Find the last complete line within 32765 chars
                LEFT(
                    string_agg(org_url, E'\n') FILTER (WHERE org_url != ''), 
                    GREATEST(
                        1,
                        CASE 
                            WHEN POSITION(E'\n' IN REVERSE(LEFT(string_agg(org_url, E'\n') FILTER (WHERE org_url != ''), 32765))) > 0
                            THEN 32765 - POSITION(E'\n' IN REVERSE(LEFT(string_agg(org_url, E'\n') FILTER (WHERE org_url != ''), 32765))) + 1
                            ELSE 32765
                        END
                    )
                )
        END as org_urls_csv
    FROM urls_raw
    GROUP BY org_id
),
-- Step 6: Group by organization ID 
endpoint_data_agg AS (
    SELECT 
        org_id,
        -- Use any organization name for this org_id (they should all be the same after UPPER conversion)
        MAX(organization_name) as organization_name,
        -- HTML formatted endpoint URLs
        string_agg(
            DISTINCT '<a class="lantern-url" tabindex="0" aria-label="Press enter to open a pop up modal containing additional information for this endpoint." onkeydown="javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)" onclick="Shiny.setInputValue(''endpoint_popup'',&quot;' || url || '&quot,{priority: ''event''});"> ' || url || '</a>',
            '<br/>'
        ) as endpoint_urls_html,
        -- Truncate at complete lines to prevent CSV corruption
        CASE 
            WHEN LENGTH(string_agg(DISTINCT url, E'\n')) <= 32765 
            THEN string_agg(DISTINCT url, E'\n')
            ELSE 
                -- Find the last complete line within 32765 chars
                LEFT(
                    string_agg(DISTINCT url, E'\n'), 
                    GREATEST(
                        1,
                        CASE 
                            WHEN POSITION(E'\n' IN REVERSE(LEFT(string_agg(DISTINCT url, E'\n'), 32765))) > 0
                            THEN 32765 - POSITION(E'\n' IN REVERSE(LEFT(string_agg(DISTINCT url, E'\n'), 32765))) + 1
                            ELSE 32765
                        END
                    )
                )
        END as endpoint_urls_csv,
        string_agg(DISTINCT fhir_version, '<br/>') as fhir_versions_html,
        string_agg(DISTINCT fhir_version, E'\n') as fhir_versions_csv,
        string_agg(DISTINCT vendor_name, '<br/>') as vendor_names_html,
        string_agg(DISTINCT vendor_name, E'\n') as vendor_names_csv,
        -- Arrays for filtering (exactly as original code)
        ARRAY(SELECT DISTINCT unnest(array_agg(fhir_version))::text ORDER BY unnest)::text[] as fhir_versions_array,
        ARRAY(SELECT DISTINCT unnest(array_agg(vendor_name))::text ORDER BY unnest)::text[] as vendor_names_array,
        ARRAY(SELECT DISTINCT unnest(array_agg(url))::text ORDER BY unnest)::text[] as urls_array
    FROM processed_data
    GROUP BY org_id  -- KEY CHANGE: Group by org_id instead of organization_name
)
-- Step 7: Final select with JOINs to get all related data per organization ID
SELECT 
    eda.organization_name,
    eda.org_id,
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
LEFT JOIN identifiers_agg ia ON eda.org_id = ia.org_id
LEFT JOIN addresses_agg aa ON eda.org_id = aa.org_id  
LEFT JOIN urls_agg ua ON eda.org_id = ua.org_id
WHERE eda.organization_name != 'UNKNOWN'
ORDER BY eda.organization_name, eda.org_id;

-- Create indexes for performance 
CREATE UNIQUE INDEX idx_mv_orgs_agg_org_id ON mv_organizations_aggregated(org_id);
CREATE INDEX idx_mv_orgs_agg_name ON mv_organizations_aggregated(organization_name);
CREATE INDEX idx_mv_orgs_agg_fhir_versions ON mv_organizations_aggregated USING GIN(fhir_versions_array);
CREATE INDEX idx_mv_orgs_agg_vendor_names ON mv_organizations_aggregated USING GIN(vendor_names_array);
CREATE INDEX idx_mv_orgs_agg_urls ON mv_organizations_aggregated USING GIN(urls_array);

COMMIT;