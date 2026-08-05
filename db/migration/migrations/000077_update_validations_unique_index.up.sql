BEGIN;

DROP INDEX IF EXISTS mv_validation_results_plot_unique_idx;
CREATE UNIQUE INDEX mv_validation_results_plot_unique_idx 
ON mv_validation_results_plot(url, fhir_version, vendor_name, rule_name, valid, expected, actual, comment, reference);

DROP INDEX IF EXISTS mv_validation_failures_unique_idx;
CREATE UNIQUE INDEX mv_validation_failures_unique_idx ON mv_validation_failures (url, fhir_version, vendor_name, rule_name, expected, actual, reference);

COMMIT;
