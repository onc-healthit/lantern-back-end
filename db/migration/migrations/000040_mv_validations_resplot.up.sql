BEGIN;
-- Materialized view for validation results plot
DROP MATERIALIZED VIEW IF EXISTS mv_validation_results_plot CASCADE;

CREATE MATERIALIZED VIEW mv_validation_results_plot AS
SELECT 
    COALESCE(vendors.name, 'Unknown') as vendor_name,
    CASE 
        WHEN capability_fhir_version = '' THEN 'No Cap Stat'
        WHEN position('-' in capability_fhir_version) > 0 THEN substring(capability_fhir_version, 1, position('-' in capability_fhir_version) - 1)
        WHEN capability_fhir_version NOT IN ('0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2', '1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', '3.0.0', '3.0.1', '3.0.2', '3.2.0', '3.3.0', '3.5.0', '3.5a.0', '4.0.0', '4.0.1') THEN 'Unknown'
        ELSE capability_fhir_version
    END AS fhir_version,
    v.rule_name,
    v.valid,
    v.reference,
    COUNT(*) as count
FROM validations v
JOIN fhir_endpoints_info f ON v.validation_result_id = f.validation_result_id
LEFT JOIN vendors ON f.vendor_id = vendors.id
WHERE f.requested_fhir_version = 'None'
AND v.rule_name IS NOT NULL
GROUP BY 
    vendors.name, 
    f.capability_fhir_version,
    v.rule_name, 
    v.valid, 
    v.reference;

CREATE UNIQUE INDEX mv_validation_results_plot_unique_idx 
ON mv_validation_results_plot(vendor_name, fhir_version, rule_name, valid, reference);

CREATE INDEX mv_validation_results_plot_vendor_idx ON mv_validation_results_plot(vendor_name);
CREATE INDEX mv_validation_results_plot_fhir_idx ON mv_validation_results_plot(fhir_version);
CREATE INDEX mv_validation_results_plot_rule_idx ON mv_validation_results_plot(rule_name);
CREATE INDEX mv_validation_results_plot_valid_idx ON mv_validation_results_plot(valid);
CREATE INDEX mv_validation_results_plot_reference_idx ON mv_validation_results_plot(reference);
COMMIT;