BEGIN;

DROP MATERIALIZED VIEW IF EXISTS security_endpoints_mv CASCADE;

CREATE MATERIALIZED VIEW security_endpoints_mv AS
SELECT 
    ROW_NUMBER() OVER () AS id,
    e.url,
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
    codes.code,
    CASE 
        -- First transform empty to "No Cap Stat"
        WHEN e.fhir_version = '' THEN 'No Cap Stat'
        -- Then handle version with dash
        WHEN e.fhir_version LIKE '%-%' THEN 
            CASE 
                WHEN SPLIT_PART(e.fhir_version, '-', 1) IN ('0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2', '1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', '3.0.0', '3.0.1', '3.0.2', '3.2.0', '3.3.0', '3.5.0', '3.5a.0', '4.0.0', '4.0.1') 
                THEN SPLIT_PART(e.fhir_version, '-', 1)
                ELSE 'Unknown'
            END
        -- Handle regular versions
        WHEN e.fhir_version IN ('0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2', '1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', '3.0.0', '3.0.1', '3.0.2', '3.2.0', '3.3.0', '3.5.0', '3.5a.0', '4.0.0', '4.0.1') 
        THEN e.fhir_version
        ELSE 'Unknown'
    END AS fhir_version_final
FROM endpoint_export e
JOIN fhir_endpoints_info f ON e.url = f.url
JOIN LATERAL (
    SELECT json_array_elements(json_array_elements(f.capability_statement::json#>'{rest,0,security,service}')->'coding')::json->>'code' AS code
) codes ON true
WHERE f.requested_fhir_version = 'None';

--indexing 
CREATE INDEX idx_security_endpoints_url ON security_endpoints_mv (url);
CREATE INDEX idx_security_endpoints_fhir_version ON security_endpoints_mv (fhir_version_final);
CREATE INDEX idx_security_endpoints_vendor_name ON security_endpoints_mv (vendor_name);
CREATE INDEX idx_security_endpoints_code ON security_endpoints_mv (code);

--unique index 
DROP INDEX IF EXISTS idx_unique_security_endpoints;
CREATE UNIQUE INDEX idx_unique_security_endpoints ON security_endpoints_mv (id, url, vendor_name, code);


DROP MATERIALIZED VIEW IF EXISTS selected_security_endpoints_mv CASCADE;

CREATE MATERIALIZED VIEW selected_security_endpoints_mv AS
SELECT 
    se.id,
    se.url,
    se.organization_names,
    se.vendor_name,
    se.capability_fhir_version,
    se.fhir_version_final AS fhir_version,
    se.tls_version,
    se.code,
    -- Create the condensed_organization_names with the modal link for endpoints with more than 5 organizations
    CASE 
        WHEN se.organization_names IS NOT NULL AND 
             array_length(string_to_array(se.organization_names, ';'), 1) > 5 
        THEN 
            CONCAT(
                array_to_string(
                    ARRAY(
                        SELECT unnest(string_to_array(se.organization_names, ';')) 
                        LIMIT 5
                    ), 
                    '; '
                ),
                '; <a class="lantern-url" tabindex="0" aria-label="Press enter to open a pop up modal containing the endpoint''s entire list of API information source names." onkeydown="javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)" onclick="Shiny.setInputValue(''show_details'',''', 
                se.url, 
                ''',{priority: ''event''});"> Click For More... </a>'
            )
        ELSE 
            se.organization_names 
    END AS condensed_organization_names,
    
    -- Create the URL with modal functionality
    CONCAT(
        '<a class="lantern-url" tabindex="0" aria-label="Press enter to open a pop up modal containing additional information for this endpoint." onkeydown="javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)" onclick="Shiny.setInputValue(''endpoint_popup'',''', 
        se.url, 
        '&&None'',{priority: ''event''});">', 
        se.url, 
        '</a>'
    ) AS url_modal
FROM 
    security_endpoints_mv se;

-- Add indexing for better performance
CREATE INDEX idx_selected_security_endpoints_fhir_version ON selected_security_endpoints_mv (fhir_version);
CREATE INDEX idx_selected_security_endpoints_vendor_name ON selected_security_endpoints_mv (vendor_name);
CREATE INDEX idx_selected_security_endpoints_code ON selected_security_endpoints_mv (code);

-- Create a unique composite index
DROP INDEX IF EXISTS idx_unique_selected_security_endpoints;
CREATE UNIQUE INDEX idx_unique_selected_security_endpoints ON selected_security_endpoints_mv (id, url, code);

COMMIT;