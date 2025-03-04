BEGIN;
DROP MATERIALIZED VIEW IF EXISTS mv_contact_information CASCADE;
CREATE MATERIALIZED VIEW mv_contact_information AS
WITH contact_data AS (
  -- Get contact information from JSON
  SELECT
    f.url,
    f.requested_fhir_version,
    COALESCE(v.name, 'Unknown') AS vendor_name,
    CASE WHEN f.capability_fhir_version = '' THEN 'No Cap Stat'
         WHEN f.capability_fhir_version SIMILAR TO '[0-9]+\.[0-9]+\.[0-9]+-.*' THEN SUBSTRING(f.capability_fhir_version FROM 1 FOR POSITION('-' IN f.capability_fhir_version)-1)
         ELSE f.capability_fhir_version
    END AS fhir_version,
    e.endpoint_names,
    contact_obj->>'name' AS contact_name,
    telecom_obj->>'system' AS contact_type,
    telecom_obj->>'value' AS contact_value,
    COALESCE((telecom_obj->>'rank')::integer, 999) AS contact_preference
  FROM fhir_endpoints_info f
  LEFT JOIN vendors v ON f.vendor_id = v.id
  LEFT JOIN endpoint_export e ON f.url = e.url AND f.requested_fhir_version = e.requested_fhir_version
  LEFT JOIN LATERAL jsonb_array_elements(f.capability_statement::jsonb->'contact') contact_obj
    ON f.capability_statement::jsonb != 'null'
  LEFT JOIN LATERAL jsonb_array_elements(contact_obj->'telecom') telecom_obj
    ON TRUE
  WHERE f.requested_fhir_version = 'None'
),
endpoints_with_metrics AS (
  -- Calculate metrics and prepare for final view
  SELECT
    cd.url,
    cd.requested_fhir_version,
    cd.vendor_name,
    cd.fhir_version,
    cd.endpoint_names,
    cd.contact_name,
    cd.contact_type,
    cd.contact_value,
    cd.contact_preference,
    -- Pre-process endpoint names for display (handling as text)
    -- Pre-process endpoint names for display (handling as text and removing braces/quotes)
CASE 
  WHEN cd.endpoint_names IS NULL THEN NULL
  WHEN cd.endpoint_names::text = '' THEN NULL
  ELSE
    -- Remove curly braces and quotes
    REGEXP_REPLACE(
      REGEXP_REPLACE(
        REGEXP_REPLACE(
          CASE 
            -- Count semicolons to determine if there are more than 5 entries
            WHEN (LENGTH(cd.endpoint_names::text) - LENGTH(REPLACE(cd.endpoint_names::text, ';', ''))) / LENGTH(';') >= 5 THEN
              -- Take portion up to the 5th semicolon and add "[more]"
              SUBSTRING(
                cd.endpoint_names::text, 
                1, 
                COALESCE(NULLIF(STRPOS(
                  SUBSTRING(
                    cd.endpoint_names::text,
                    COALESCE(NULLIF(STRPOS(
                      SUBSTRING(
                        cd.endpoint_names::text,
                        COALESCE(NULLIF(STRPOS(
                          SUBSTRING(
                            cd.endpoint_names::text,
                            COALESCE(NULLIF(STRPOS(cd.endpoint_names::text, ';'), 0), 0) + 1
                          ), 
                          ';'
                        ), 0), 0) + 1
                      ), 
                      ';'
                    ), 0), 0) + 1
                  ), 
                  ';'
                ), 0), LENGTH(cd.endpoint_names::text))
              ) || ' [more]'
            ELSE cd.endpoint_names::text
          END,
          '\\{|\\}', '', 'g'  -- Remove curly braces
        ),
        '"', '', 'g'  -- Remove double quotes
      ),
      '\\\\', '', 'g'  -- Remove escape backslashes
    )
END AS condensed_endpoint_names,
    -- Calculate other metrics
    COUNT(*) OVER (PARTITION BY cd.url) AS num_contacts,
    CASE 
      WHEN cd.contact_name IS NOT NULL OR cd.contact_type IS NOT NULL OR cd.contact_value IS NOT NULL 
      THEN TRUE ELSE FALSE 
    END AS has_contact,
    ROW_NUMBER() OVER (PARTITION BY cd.url ORDER BY cd.contact_preference) AS contact_rank
  FROM contact_data cd
)
SELECT *
FROM endpoints_with_metrics;

-- Create necessary indexes
CREATE UNIQUE INDEX mv_contact_information_uniq
  ON mv_contact_information (url, requested_fhir_version, COALESCE(contact_rank, -1));
CREATE INDEX mv_contact_information_fhir_version_idx 
  ON mv_contact_information (fhir_version);
CREATE INDEX mv_contact_information_vendor_name_idx 
  ON mv_contact_information (vendor_name);
CREATE INDEX mv_contact_information_has_contact_idx 
  ON mv_contact_information (has_contact);

COMMIT;