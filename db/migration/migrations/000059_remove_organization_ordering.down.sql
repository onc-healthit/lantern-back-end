BEGIN;

-- Drop the materialized view without ordering
DROP MATERIALIZED VIEW IF EXISTS mv_endpoint_list_organizations CASCADE;

-- Recreate with ORDER BY clause (original version)
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_endpoint_list_organizations
AS
SELECT DISTINCT
    endpoint_export.url,
    COALESCE(
        NULLIF(
            btrim(
                regexp_replace(name_id.cleaned_name, '\s+', ' ', 'g')
            ), 
        ''), 
    'Unknown') AS organization_name,
    
    COALESCE(
        name_id.cleaned_id::text,
        'Unknown'
    ) AS organization_id,
    
    CASE
        WHEN endpoint_export.fhir_version::text = ''::text THEN 'No Cap Stat'::character varying
        ELSE endpoint_export.fhir_version
    END AS fhir_version,
    
    COALESCE(endpoint_export.vendor_name, 'Unknown'::character varying) AS vendor_name

FROM
    endpoint_export
LEFT JOIN LATERAL (
    SELECT
        name_elem AS cleaned_name,
        id_elem AS cleaned_id
    FROM
        unnest(endpoint_export.endpoint_names, endpoint_export.endpoint_ids) AS u(name_elem, id_elem)
) AS name_id ON TRUE

ORDER BY
    organization_name
WITH DATA;

 -- Create indexes for endpoint list organizations materialized view
CREATE UNIQUE INDEX idx_mv_endpoint_list_org_uniq ON mv_endpoint_list_organizations(fhir_version, vendor_name, url, organization_name, organization_id);
CREATE INDEX idx_mv_endpoint_list_org_fhir ON mv_endpoint_list_organizations(fhir_version);
CREATE INDEX idx_mv_endpoint_list_org_vendor ON mv_endpoint_list_organizations(vendor_name);
CREATE INDEX idx_mv_endpoint_list_org_url ON mv_endpoint_list_organizations(url);

COMMIT;