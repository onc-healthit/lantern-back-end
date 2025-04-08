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
SELECT fhir_version, url, expected, actual, vendor_name, rule_name, reference
FROM mv_validation_results_plot
WHERE valid = 'false';

CREATE UNIQUE INDEX mv_validation_failures_unique_idx ON mv_validation_failures(url, fhir_version, vendor_name, rule_name);
CREATE INDEX mv_validation_failures_url_idx ON mv_validation_failures(url);
CREATE INDEX mv_validation_failures_fhir_version_idx ON mv_validation_failures(fhir_version);
CREATE INDEX mv_validation_failures_vendor_name_idx ON mv_validation_failures(vendor_name);
CREATE INDEX mv_validation_failures_rule_name_idx ON mv_validation_failures(rule_name);
CREATE INDEX mv_validation_failures_reference_idx ON mv_validation_failures(reference);
COMMIT;