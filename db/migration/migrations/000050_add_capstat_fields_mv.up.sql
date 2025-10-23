BEGIN;

-- Create materialized view for capstat_fields
DROP MATERIALIZED VIEW IF EXISTS mv_capstat_fields CASCADE;
CREATE MATERIALIZED VIEW mv_capstat_fields AS 

SELECT 
  f.id AS endpoint_id,
  f.vendor_id,
  COALESCE(vendors.name, 'Unknown') AS vendor_name,
  CASE
    WHEN REGEXP_REPLACE(f.capability_fhir_version, '-.*', '') IN (
      'No Cap Stat', '0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2',
      '1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', '3.0.0',
      '3.0.1', '3.0.2', '3.2.0', '3.3.0', '3.5.0', '3.5a.0',
      '4.0.0', '4.0.1'
    )
    THEN REGEXP_REPLACE(f.capability_fhir_version, '-.*', '')
    ELSE 'Unknown'
  END AS fhir_version,
  json_array_elements(f.included_fields::json) ->> 'Field' AS field,
  json_array_elements(f.included_fields::json) ->> 'Exists' AS exist,
  json_array_elements(f.included_fields::json) ->> 'Extension' AS extension
FROM fhir_endpoints_info f
LEFT JOIN vendors ON f.vendor_id = vendors.id
WHERE f.included_fields != 'null'
  AND f.requested_fhir_version = 'None'
ORDER BY endpoint_id;

-- Create indexes for mv_capstat_fields
CREATE UNIQUE INDEX idx_mv_capstat_fields_unique ON mv_capstat_fields(endpoint_id, fhir_version, field);
CREATE INDEX idx_mv_capstat_fields_vendor ON mv_capstat_fields(vendor_name);
CREATE INDEX idx_mv_capstat_fields_fhir ON mv_capstat_fields(fhir_version);

-- Create materialized view for capstat_fields_text
DROP MATERIALIZED VIEW IF EXISTS mv_capstat_values_fields CASCADE;
CREATE MATERIALIZED VIEW mv_capstat_values_fields AS 

WITH base AS (
  -- Start with the materialized view filtered on extension
  SELECT *
  FROM mv_capstat_fields
  WHERE extension = 'false'
),
with_version AS (
  -- Add a computed column for the FHIR version name based on your groups
  SELECT
    field,
    fhir_version,
    CASE
      WHEN fhir_version IN ('0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2') THEN 'DSTU2'
      WHEN fhir_version IN ('1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', '3.0.0', '3.0.1', '3.0.2') THEN 'STU3'
      WHEN fhir_version IN ('3.2.0', '3.3.0', '3.5.0', '3.5a.0', '4.0.0', '4.0.1') THEN 'R4'
      ELSE 'DSTU2'
    END AS fhir_version_name
  FROM base
),
fvn AS (
  -- For each field, get the unique FHIR version names as a comma‐separated list
  SELECT 
    field, 
    string_agg(DISTINCT fhir_version_name, ', ') AS fhir_version_names
  FROM with_version
  GROUP BY field
)
SELECT DISTINCT 
  fhir_version,
  with_version.field || ' (' || fvn.fhir_version_names || ')' AS field_version
FROM with_version
LEFT JOIN fvn ON with_version.field = fvn.field
ORDER BY field_version;

-- Create indexes for mv_capstat_values_fields
CREATE UNIQUE INDEX idx_mv_capstat_values_fields_unique ON mv_capstat_values_fields(fhir_version, field_version);
CREATE INDEX idx_mv_capstat_values_fields_field_version ON mv_capstat_values_fields(field_version);
CREATE INDEX idx_mv_capstat_values_fields_fhir ON mv_capstat_values_fields(fhir_version);

-- Create materialized view for capstat_extension_text
DROP MATERIALIZED VIEW IF EXISTS mv_capstat_values_extension CASCADE;
CREATE MATERIALIZED VIEW mv_capstat_values_extension AS 

WITH base AS (
  -- Start with the materialized view filtered on extension
  SELECT *
  FROM mv_capstat_fields
  WHERE extension = 'true'
),
with_version AS (
  -- Add a computed column for the FHIR version name based on your groups
  SELECT
    field,
    fhir_version,
    CASE
      WHEN fhir_version IN ('0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2') THEN 'DSTU2'
      WHEN fhir_version IN ('1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', '3.0.0', '3.0.1', '3.0.2') THEN 'STU3'
      WHEN fhir_version IN ('3.2.0', '3.3.0', '3.5.0', '3.5a.0', '4.0.0', '4.0.1') THEN 'R4'
      ELSE 'DSTU2'
    END AS fhir_version_name
  FROM base
),
fvn AS (
  -- For each field, get the unique FHIR version names as a comma‐separated list
  SELECT 
    field, 
    string_agg(DISTINCT fhir_version_name, ', ') AS fhir_version_names
  FROM with_version
  GROUP BY field
)
SELECT DISTINCT 
  fhir_version,
  with_version.field || ' (' || fvn.fhir_version_names || ')' AS field_version
FROM with_version
LEFT JOIN fvn ON with_version.field = fvn.field
ORDER BY field_version;

-- Create indexes for mv_capstat_values_extension
CREATE UNIQUE INDEX idx_mv_capstat_values_extension_unique ON mv_capstat_values_extension(fhir_version, field_version);
CREATE INDEX idx_mv_capstat_values_extension_field_version ON mv_capstat_values_extension(field_version);
CREATE INDEX idx_mv_capstat_values_extension_fhir ON mv_capstat_values_extension(fhir_version);

COMMIT;