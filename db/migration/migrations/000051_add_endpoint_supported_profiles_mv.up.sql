BEGIN;

DROP MATERIALIZED VIEW IF EXISTS endpoint_supported_profiles_mv CASCADE;

CREATE MATERIALIZED VIEW endpoint_supported_profiles_mv AS
SELECT
  f.id AS endpoint_id,
  f.url,
  f.vendor_id,
  COALESCE(vendors.name, 'Unknown') AS vendor_name,
  CASE
    WHEN f.capability_fhir_version = '' THEN 'No Cap Stat'
    WHEN split_part(f.capability_fhir_version, '-', 1) = ANY (
      ARRAY[
        'No Cap Stat', '0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2', '1.1.0',
        '1.2.0', '1.4.0', '1.6.0', '1.8.0', '3.0.0', '3.0.1', '3.0.2',
        '3.2.0', '3.3.0', '3.5.0', '3.5a.0', '4.0.0', '4.0.1'
      ]
    ) THEN split_part(f.capability_fhir_version, '-', 1)
    ELSE 'Unknown'
  END AS fhir_version,
  sp.value ->> 'Resource' AS resource,
  sp.value ->> 'ProfileURL' AS profileurl,
  sp.value ->> 'ProfileName' AS profilename
FROM
  fhir_endpoints_info f
LEFT JOIN
  vendors ON f.vendor_id = vendors.id
CROSS JOIN LATERAL
  json_array_elements(f.supported_profiles::json) sp(value)
WHERE
  f.supported_profiles::text <> 'null'
  AND f.requested_fhir_version = 'None';

COMMIT;
