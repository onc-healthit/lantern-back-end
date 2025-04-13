BEGIN; 
-- Materialized view for validation details 
DROP INDEX IF EXISTS mv_validation_details_unique_idx; 
DROP MATERIALIZED VIEW IF EXISTS mv_validation_details;

CREATE MATERIALIZED VIEW mv_validation_details AS 
WITH validation_data AS ( 
    SELECT 
        COALESCE(vendors.name, 'Unknown') as vendor_name, 
        CASE 
            WHEN capability_fhir_version = '' THEN 'No Cap Stat' 
            WHEN position('-' in capability_fhir_version) > 0 THEN substring(capability_fhir_version, 1, position('-' in capability_fhir_version) - 1) 
            WHEN capability_fhir_version NOT IN ('0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2', '1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', '3.0.0', '3.0.1', '3.0.2', '3.2.0', '3.3.0', '3.5.0', '3.5a.0', '4.0.0', '4.0.1') THEN 'Unknown' 
            ELSE capability_fhir_version 
        END AS fhir_version, 
        rule_name, 
        reference 
    FROM validations v 
    JOIN fhir_endpoints_info f ON v.validation_result_id = f.validation_result_id 
    LEFT JOIN vendors on f.vendor_id = vendors.id 
    WHERE f.requested_fhir_version = 'None'
    AND v.rule_name IS NOT NULL 
),
mapped_versions AS (
    SELECT DISTINCT
        rule_name,
        CASE 
            WHEN fhir_version IN ('0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2') THEN 'DSTU2' 
            WHEN fhir_version IN ('1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', '3.0.0', '3.0.1', '3.0.2') THEN 'STU3' 
            WHEN fhir_version IN ('3.2.0', '3.3.0', '3.5.0', '3.5a.0', '4.0.0', '4.0.1') THEN 'R4' 
            ELSE fhir_version
        END AS version_name,
        -- Add a sort order to maintain the original ordering
        CASE
            WHEN fhir_version IN ('0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2') THEN 1 -- DSTU2
            WHEN fhir_version IN ('1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', '3.0.0', '3.0.1', '3.0.2') THEN 2 -- STU3
            WHEN fhir_version IN ('3.2.0', '3.3.0', '3.5.0', '3.5a.0', '4.0.0', '4.0.1') THEN 3 -- R4
            ELSE 4 -- Others
        END AS sort_order
    FROM validation_data
    WHERE fhir_version != 'Unknown' AND fhir_version != 'No Cap Stat'
),
validation_versions AS (
    SELECT
        rule_name,
        STRING_AGG(version_name, ', ' ORDER BY sort_order) as fhir_version_names
    FROM (
        SELECT DISTINCT rule_name, version_name, sort_order
        FROM mapped_versions
    ) AS unique_versions
    GROUP BY rule_name
)
SELECT 
    vd.rule_name, 
    COALESCE(vv.fhir_version_names, '') as fhir_version_names 
FROM ( 
    SELECT DISTINCT rule_name 
    FROM validation_data 
) vd 
LEFT JOIN validation_versions vv ON vd.rule_name = vv.rule_name 
ORDER BY vd.rule_name;

CREATE UNIQUE INDEX mv_validation_details_unique_idx ON mv_validation_details(rule_name); 
COMMIT;