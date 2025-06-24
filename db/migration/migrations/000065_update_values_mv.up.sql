BEGIN;

DROP MATERIALIZED VIEW IF EXISTS get_value_versions_mv CASCADE;

CREATE MATERIALIZED VIEW get_value_versions_mv AS
SELECT 
    field,
    array_agg(DISTINCT fhir_version ORDER BY fhir_version) AS fhir_versions
FROM 
    get_capstat_fields_mv
GROUP BY 
    field;

-- Create a unique composite index
DROP INDEX IF EXISTS idx_get_value_versions_mv_field;
CREATE UNIQUE INDEX idx_get_value_versions_mv_field ON get_value_versions_mv(field);


DROP MATERIALIZED VIEW IF EXISTS selected_fhir_endpoints_values_mv CASCADE;

CREATE MATERIALIZED VIEW selected_fhir_endpoints_values_mv AS
WITH base_data AS (
    -- Start with the capstat values data
    SELECT 
        g.vendor_name AS "Developer",
        g.filter_fhir_version AS "FHIR Version",
        g.fhir_version AS "fhirVersion",
        g.software_name AS "software.name",
        g.software_version AS "software.version",
        g.software_release_date AS "software.releaseDate",
        g.implementation_description AS "implementation.description",
        g.implementation_url AS "implementation.url",
        g.implementation_custodian AS "implementation.custodian",
        -- All other fields from capability statement
        g.url,
        g.version,
        g.name,
        g.title,
        g.date,
        g.publisher,
        g.description,
        g.purpose,
        g.copyright,
        g.endpoint_id
    FROM get_capstat_values_mv g
),
-- Create a cross join of all possible field combinations
field_combinations AS (
    SELECT 
        b."Developer",
        b."FHIR Version",
        v.field,
        UNNEST(v.fhir_versions) AS field_version,
        -- Create a lateral join to get the value for each field
        CASE 
            WHEN v.field = 'url' THEN b.url
            WHEN v.field = 'version' THEN b.version
            WHEN v.field = 'name' THEN b.name
            WHEN v.field = 'title' THEN b.title
            WHEN v.field = 'date' THEN b.date
            WHEN v.field = 'publisher' THEN b.publisher
            WHEN v.field = 'description' THEN b.description
            WHEN v.field = 'purpose' THEN b.purpose
            WHEN v.field = 'copyright' THEN b.copyright
            WHEN v.field = 'software.name' THEN b."software.name"
            WHEN v.field = 'software.version' THEN b."software.version"
            WHEN v.field = 'software.releaseDate' THEN b."software.releaseDate"
            WHEN v.field = 'implementation.description' THEN b."implementation.description"
            WHEN v.field = 'implementation.url' THEN b."implementation.url"
            WHEN v.field = 'implementation.custodian' THEN b."implementation.custodian"
            WHEN v.field = 'fhirVersion' THEN b."fhirVersion"
			ELSE NULL
        END AS field_value,
        b.endpoint_id
    FROM base_data b
    CROSS JOIN get_value_versions_mv v
    WHERE b."FHIR Version" IN (SELECT UNNEST(v.fhir_versions) FROM get_value_versions_mv WHERE field = v.field)
)
-- Final aggregation
SELECT 
    "Developer",
    "FHIR Version",
    field,
    CASE WHEN COALESCE(field_value, '[Empty]') = '[Empty]' THEN 'no' ELSE 'yes' END AS is_used,
    COALESCE(field_value, '[Empty]') AS field_value,
    COUNT(DISTINCT endpoint_id)::INT AS "Endpoints"  -- Explicitly cast to INT
FROM field_combinations
GROUP BY "Developer", "FHIR Version", field, field_value
ORDER BY "Developer", "FHIR Version", field, field_value;


-- Create indexes for performance optimization
CREATE INDEX idx_selected_fhir_endpoints_dev ON selected_fhir_endpoints_values_mv("Developer");
CREATE INDEX idx_selected_fhir_endpoints_fhir_version ON selected_fhir_endpoints_values_mv("FHIR Version");
CREATE INDEX idx_selected_fhir_endpoints_field ON selected_fhir_endpoints_values_mv(Field);
CREATE INDEX idx_selected_fhir_endpoints_field_value ON selected_fhir_endpoints_values_mv(field_value);
CREATE INDEX idx_selected_fhir_endpoints_is_used ON selected_fhir_endpoints_values_mv(is_used);
CREATE INDEX idx_summary_query ON selected_fhir_endpoints_values_mv (field, "FHIR Version", "Developer", is_used);

-- Create a unique composite index
DROP INDEX IF EXISTS idx_selected_fhir_endpoints_unique;
CREATE UNIQUE INDEX idx_selected_fhir_endpoints_unique ON selected_fhir_endpoints_values_mv("Developer", "FHIR Version", Field, field_value);

COMMIT;