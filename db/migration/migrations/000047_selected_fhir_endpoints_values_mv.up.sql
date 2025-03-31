BEGIN;

DROP MATERIALIZED VIEW IF EXISTS selected_fhir_endpoints_values_mv CASCADE;

CREATE MATERIALIZED VIEW selected_fhir_endpoints_values_mv AS
WITH base AS (
    SELECT 
        g.vendor_name AS "Developer",
        g.filter_fhir_version AS "FHIR Version",
        g.url,
        g.fhir_version AS "fhirVersion",
        g.name,
        g.title,
        g.date,
        g.publisher,
        g.description,
        g.purpose,
        g.copyright,
        g.software_name AS "software.name",
        g.software_version AS "software.version",
        g.software_release_date AS "software.releaseDate",
        g.implementation_description AS "implementation.description",
        g.implementation_url AS "implementation.url",
        g.implementation_custodian AS "implementation.custodian"
    FROM get_capstat_values_mv g
),
aggregated AS (
    SELECT 
        "Developer",
        "FHIR Version",
        UNNEST(ARRAY[
            'url', 'fhirVersion', 'name', 'title', 'date', 'publisher', 'description', 'purpose', 'copyright', 
            'software.name', 'software.version', 'software.releaseDate', 
            'implementation.description', 'implementation.url', 'implementation.custodian'
        ]) AS Field,
        UNNEST(ARRAY[
            url, "fhirVersion", name, title, date, publisher, description, purpose, copyright, 
            "software.name", "software.version", "software.releaseDate", 
            "implementation.description", "implementation.url", "implementation.custodian"
        ]) AS field_value
    FROM base
)
SELECT 
    "Developer",
    "FHIR Version",
    Field,
    COALESCE(field_value, '[Empty]') AS field_value,
    COUNT(*)::INT AS "Endpoints"
FROM aggregated
GROUP BY "Developer", "FHIR Version", Field, field_value;


-- Create indexes for performance optimization
CREATE INDEX idx_selected_fhir_endpoints_dev ON selected_fhir_endpoints_values_mv("Developer");
CREATE INDEX idx_selected_fhir_endpoints_fhir_version ON selected_fhir_endpoints_values_mv("FHIR Version");
CREATE INDEX idx_selected_fhir_endpoints_field ON selected_fhir_endpoints_values_mv(Field);
CREATE INDEX idx_selected_fhir_endpoints_field_value ON selected_fhir_endpoints_values_mv(field_value);

-- Create a unique composite index
DROP INDEX IF EXISTS idx_selected_fhir_endpoints_unique;
CREATE UNIQUE INDEX idx_selected_fhir_endpoints_unique ON selected_fhir_endpoints_values_mv("Developer", "FHIR Version", Field, field_value);

COMMIT;