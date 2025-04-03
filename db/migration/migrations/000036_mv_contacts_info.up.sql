BEGIN;

DROP MATERIALIZED VIEW IF EXISTS mv_contacts_info CASCADE;

CREATE MATERIALIZED VIEW mv_contacts_info AS
WITH contact_info_extracted AS (
  SELECT DISTINCT
    url,
    json_array_elements((capability_statement->>'contact')::json)->>'name' as contact_name,
    json_array_elements((json_array_elements((capability_statement->>'contact')::json)->>'telecom')::json)->>'system' as contact_type,
    json_array_elements((json_array_elements((capability_statement->>'contact')::json)->>'telecom')::json)->>'value' as contact_value,
    CAST(NULLIF(json_array_elements((json_array_elements((capability_statement->>'contact')::json)->>'telecom')::json)->>'rank', '') AS INTEGER) as contact_preference
  FROM fhir_endpoints_info
  WHERE capability_statement::jsonb != 'null' AND requested_fhir_version = 'None'
),
endpoint_details AS (
  SELECT DISTINCT -- Added DISTINCT to eliminate potential duplication
    url,
    -- Fix for handling Unknown vendor - make sure empty or NULL is replaced with 'Unknown'
    CASE 
      WHEN vendor_name IS NULL OR vendor_name = '' THEN 'Unknown' 
      ELSE vendor_name 
    END AS vendor_name,
    CASE 
      WHEN fhir_version = '' OR fhir_version IS NULL THEN 'No Cap Stat'
      WHEN position('-' in fhir_version) > 0 THEN substring(fhir_version from 1 for position('-' in fhir_version) - 1)
      WHEN fhir_version NOT IN ('0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2', '1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', '3.0.0', '3.0.1', '3.0.2', '3.2.0', '3.3.0', '3.5.0', '3.5a.0', '4.0.0', '4.0.1') THEN 'Unknown'
      ELSE fhir_version
    END AS fhir_version,
    requested_fhir_version
  FROM endpoint_export
  WHERE requested_fhir_version = 'None'
),
endpoint_names_grouped AS (
  SELECT 
    url, 
    string_agg(DISTINCT endpoint_names_list, ';') AS endpoint_names -- Added DISTINCT to avoid duplications
  FROM (
    SELECT DISTINCT url, UNNEST(endpoint_names) as endpoint_names_list 
    FROM endpoint_export 
    WHERE requested_fhir_version = 'None'
    ORDER BY endpoint_names_list
  ) AS unnested
  GROUP BY url
),
-- First, get URLs with contact info
urls_with_contacts AS (
  SELECT DISTINCT url
  FROM contact_info_extracted
),
-- Then, get URLs without contact info
urls_without_contacts AS (
  SELECT DISTINCT e.url
  FROM endpoint_details e
  LEFT JOIN urls_with_contacts c ON e.url = c.url
  WHERE c.url IS NULL
),
-- Combine contact data
joined_with_contacts AS (
  SELECT 
    e.url,
    e.vendor_name,
    e.fhir_version,
    eng.endpoint_names,
    e.requested_fhir_version,
    c.contact_name,
    c.contact_type,
    c.contact_value,
    COALESCE(c.contact_preference, 999) AS contact_preference,
    TRUE AS has_contact,
    MD5(CONCAT(
      e.url, 
      COALESCE(c.contact_name, ''), 
      COALESCE(c.contact_type, ''), 
      COALESCE(c.contact_value, ''),
      COALESCE(c.contact_preference::text, '999'),
      COALESCE(random()::text, '')  -- Add randomness to handle duplicates
    )) AS unique_hash
  FROM 
    endpoint_details e
  INNER JOIN 
    urls_with_contacts uc ON e.url = uc.url
  LEFT JOIN 
    endpoint_names_grouped eng ON e.url = eng.url
  LEFT JOIN 
    contact_info_extracted c ON e.url = c.url
),
-- Handle URLs without contacts
joined_without_contacts AS (
  SELECT 
    e.url,
    e.vendor_name,
    e.fhir_version,
    eng.endpoint_names,
    e.requested_fhir_version,
    NULL AS contact_name,
    NULL AS contact_type,
    NULL AS contact_value,
    999 AS contact_preference,
    FALSE AS has_contact,
    MD5(CONCAT(
      e.url, 
      'no_contact',
      COALESCE(random()::text, '')  -- Add randomness to handle duplicates
    )) AS unique_hash
  FROM 
    endpoint_details e
  INNER JOIN 
    urls_without_contacts nc ON e.url = nc.url
  LEFT JOIN 
    endpoint_names_grouped eng ON e.url = eng.url
)
-- Combine both sets
SELECT * FROM joined_with_contacts
UNION ALL
SELECT * FROM joined_without_contacts
ORDER BY 
  url, 
  contact_preference;

-- Create unique index for concurrent refresh
DROP INDEX IF EXISTS idx_mv_contacts_info_unique;
CREATE UNIQUE INDEX idx_mv_contacts_info_unique ON mv_contacts_info(unique_hash);

-- Create indexes to improve query performance
DROP INDEX IF EXISTS idx_mv_contacts_info_url;
CREATE INDEX idx_mv_contacts_info_url ON mv_contacts_info(url);

DROP INDEX IF EXISTS idx_mv_contacts_info_fhir_version;
CREATE INDEX idx_mv_contacts_info_fhir_version ON mv_contacts_info(fhir_version);

DROP INDEX IF EXISTS idx_mv_contacts_info_vendor_name;
CREATE INDEX idx_mv_contacts_info_vendor_name ON mv_contacts_info(vendor_name);

DROP INDEX IF EXISTS idx_mv_contacts_info_has_contact;
CREATE INDEX idx_mv_contacts_info_has_contact ON mv_contacts_info(has_contact);

DROP INDEX IF EXISTS idx_mv_contacts_info_contact_preference;
CREATE INDEX idx_mv_contacts_info_contact_preference ON mv_contacts_info(contact_preference);

COMMIT;