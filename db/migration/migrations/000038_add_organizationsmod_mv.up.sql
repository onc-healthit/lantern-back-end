BEGIN;

-- Create materialized view for NPI organization matches
DROP MATERIALIZED VIEW IF EXISTS mv_npi_organization_matches CASCADE;
CREATE MATERIALIZED VIEW mv_npi_organization_matches AS
SELECT 
    url,
    requested_fhir_version,
    vendor_name,
    fhir_version,
    organization_name,
    organization_secondary_name,
    npi_id,
    zipcode,
    match_score * 100 AS match_score,
    CASE 
        WHEN match_score * 100 = 100 THEN '100'
        WHEN match_score * 100 >= 99 AND match_score * 100 < 100 THEN '99-100'
        WHEN match_score * 100 >= 98 AND match_score * 100 < 99 THEN '98-100'
        WHEN match_score * 100 >= 97 AND match_score * 100 < 98 THEN '97-100'
        ELSE 'below_97'
    END AS confidence_range
FROM 
    organization_location
WHERE 
    match_score >= 0.97
ORDER BY 
    organization_name, url;

-- Create indexes for NPI organization matches materialized view
CREATE INDEX idx_mv_npi_org_fhir ON mv_npi_organization_matches(fhir_version);
CREATE INDEX idx_mv_npi_org_vendor ON mv_npi_organization_matches(vendor_name);
CREATE INDEX idx_mv_npi_org_confidence ON mv_npi_organization_matches(confidence_range);
CREATE INDEX idx_mv_npi_org_url ON mv_npi_organization_matches(url);
CREATE INDEX idx_mv_npi_org_requested_fhir ON mv_npi_organization_matches(requested_fhir_version);

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

-- Create materialized view for endpoint locations
DROP MATERIALIZED VIEW IF EXISTS mv_endpoint_locations CASCADE;
CREATE MATERIALIZED VIEW mv_endpoint_locations AS
SELECT
    ol.url,
    ol.endpoint_names[1] as endpoint_name,
    ol.organization_name,
    ol.fhir_version,
    ol.vendor_name,
    ol.match_score,
    left(ol.zipcode, 5) as zipcode,
    ol.npi_id,
    ol.requested_fhir_version
FROM 
    organization_location ol;

-- Create indexes for endpoint locations materialized view
CREATE INDEX idx_mv_endpoint_loc_fhir ON mv_endpoint_locations(fhir_version);
CREATE INDEX idx_mv_endpoint_loc_vendor ON mv_endpoint_locations(vendor_name);
CREATE INDEX idx_mv_endpoint_loc_url ON mv_endpoint_locations(url);

COMMIT;