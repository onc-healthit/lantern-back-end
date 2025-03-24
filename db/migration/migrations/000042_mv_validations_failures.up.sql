BEGIN;
-- Materialized view for validation failures
DROP INDEX IF EXISTS mv_validation_failures_unique_idx;
DROP INDEX IF EXISTS mv_validation_failures_url_idx;
DROP INDEX IF EXISTS mv_validation_failures_fhir_version_idx;
DROP INDEX IF EXISTS mv_validation_failures_vendor_name_idx;
DROP INDEX IF EXISTS mv_validation_failures_rule_name_idx;
DROP INDEX IF EXISTS mv_validation_failures_reference_idx;
DROP MATERIALIZED VIEW IF EXISTS mv_validation_failures;

CREATE MATERIALIZED VIEW mv_validation_failures AS
SELECT 
    ROW_NUMBER() OVER (ORDER BY f.url, v.rule_name) as id,
    f.url,
    CASE 
        WHEN capability_fhir_version = '' THEN 'No Cap Stat'
        WHEN position('-' in capability_fhir_version) > 0 THEN substring(capability_fhir_version, 1, position('-' in capability_fhir_version) - 1)
        WHEN capability_fhir_version NOT IN ('0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2', '1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', '3.0.0', '3.0.1', '3.0.2', '3.2.0', '3.3.0', '3.5.0', '3.5a.0', '4.0.0', '4.0.1') THEN 'Unknown'
        ELSE capability_fhir_version
    END AS fhir_version,
    COALESCE(vendors.name, 'Unknown') as vendor_name,
    v.rule_name,
    v.reference,
    v.expected,
    v.actual
FROM validations v
JOIN fhir_endpoints_info f ON v.validation_result_id = f.validation_result_id
LEFT JOIN vendors ON f.vendor_id = vendors.id
WHERE f.requested_fhir_version = 'None'
AND v.valid = FALSE
AND v.rule_name IS NOT NULL;

CREATE UNIQUE INDEX mv_validation_failures_unique_idx ON mv_validation_failures(id);
CREATE INDEX mv_validation_failures_url_idx ON mv_validation_failures(url);
CREATE INDEX mv_validation_failures_fhir_version_idx ON mv_validation_failures(fhir_version);
CREATE INDEX mv_validation_failures_vendor_name_idx ON mv_validation_failures(vendor_name);
CREATE INDEX mv_validation_failures_rule_name_idx ON mv_validation_failures(rule_name);
CREATE INDEX mv_validation_failures_reference_idx ON mv_validation_failures(reference);
COMMIT;