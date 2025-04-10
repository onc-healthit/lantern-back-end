BEGIN;

-- Create materialized view for endpoint list organizations
DROP MATERIALIZED VIEW IF EXISTS mv_endpoint_list_organizations CASCADE;

CREATE MATERIALIZED VIEW mv_endpoint_list_organizations AS
SELECT DISTINCT
  url,
  UNNEST(COALESCE(NULLIF(processed_names.cleaned_names, '{}'::text[]), ARRAY['Unknown'])) AS organization_name,
  CASE
	WHEN endpoint_export.fhir_version::text = ''::text THEN 'No Cap Stat'::character varying
	ELSE endpoint_export.fhir_version
  END AS fhir_version,
  COALESCE(endpoint_export.vendor_name, 'Unknown'::character varying) AS vendor_name,
  requested_fhir_version
FROM endpoint_export,
LATERAL(
  SELECT
    CASE
      WHEN endpoint_names IS NULL THEN ARRAY['Unknown']
      ELSE ARRAY(
        SELECT btrim(regexp_replace(unnest(string_to_array(regexp_replace(elem.elem::text, '["]', '', 'g'),';')),'\s+',' ','g'))
        	FROM unnest(endpoint_names) elem(elem))
    END AS cleaned_names
) AS processed_names
ORDER BY organization_name;

-- Create indexes for endpoint list organizations materialized view
CREATE UNIQUE INDEX idx_mv_endpoint_list_org_uniq ON mv_endpoint_list_organizations(fhir_version, vendor_name, url, organization_name, requested_fhir_version);
CREATE INDEX idx_mv_endpoint_list_org_fhir ON mv_endpoint_list_organizations(fhir_version);
CREATE INDEX idx_mv_endpoint_list_org_vendor ON mv_endpoint_list_organizations(vendor_name);
CREATE INDEX idx_mv_endpoint_list_org_url ON mv_endpoint_list_organizations(url);

COMMIT;