BEGIN;

DROP MATERIALIZED VIEW IF EXISTS mv_validation_results_plot CASCADE;

CREATE MATERIALIZED VIEW mv_validation_results_plot AS
SELECT DISTINCT t.url,
t.fhir_version,
t.vendor_name,
t.rule_name,
t.valid,
t.expected,
t.actual,
t.comment,
t.reference
FROM (SELECT DISTINCT ON (f.url, f.requested_fhir_version, v.validation_result_id, v.rule_name, f.vendor_id)
        COALESCE(vendors.name, 'Unknown'::character varying) AS vendor_name,
        f.url,
            CASE
                WHEN f.capability_fhir_version::text = ''::text THEN 'No Cap Stat'::character varying
                WHEN "position"(f.capability_fhir_version::text, '-'::text) > 0 THEN "substring"(f.capability_fhir_version::text, 1, "position"(f.capability_fhir_version::text, '-'::text) - 1)::character varying
                WHEN f.capability_fhir_version::text <> ALL (ARRAY['0.4.0'::character varying, '0.4'::character varying, '0.5.0'::character varying, '0.5'::character varying, '1.0.0'::character varying, '1.0'::character varying, '1'::character varying, '1.0.1'::character varying, '1.0.2'::character varying, '1.1.0'::character varying, '1.1'::character varying, '1.2.0'::character varying, '1.2'::character varying, '1.4.0'::character varying, '1.4'::character varying, '1.6.0'::character varying, '1.6'::character varying, '1.8.0'::character varying, '1.8'::character varying, '3.0.0'::character varying, '3.0'::character varying, '3'::character varying, '3.0.1'::character varying, '3.0.2'::character varying, '3.2.0'::character varying, '3.2'::character varying, '3.3.0'::character varying, '3.3'::character varying, '3.5.0'::character varying, '3.5'::character varying, '3.5a.0'::character varying, '4.0.0'::character varying, '4.0'::character varying, '4'::character varying, '4.0.1'::character varying, '4.1.0'::character varying, '4.1'::character varying, '4.3.0'::character varying, '4.3'::character varying, '4.2.0'::character varying, '4.2'::character varying, '4.4.0'::character varying, '4.4'::character varying, '4.5.0'::character varying, '4.5'::character varying, '4.6.0'::character varying, '4.6'::character varying, '5.0.0'::character varying, '5.0'::character varying, '5'::character varying]::text[]) THEN 'Unknown'::character varying
                ELSE f.capability_fhir_version
            END AS fhir_version,
        v.rule_name,
        v.valid,
        v.expected,
        v.actual,
        v.comment,
        v.reference,
        v.validation_result_id AS id,
        f.requested_fhir_version
        FROM fhir_endpoints_info f
            JOIN validations v ON f.validation_result_id = v.validation_result_id
            LEFT JOIN vendors ON f.vendor_id = vendors.id
        ORDER BY f.url, f.requested_fhir_version, v.validation_result_id, v.rule_name, f.vendor_id) t;

CREATE UNIQUE INDEX mv_validation_results_plot_unique_idx 
ON mv_validation_results_plot(url, fhir_version, vendor_name, rule_name, valid, expected, actual);

CREATE INDEX mv_validation_results_plot_vendor_idx ON mv_validation_results_plot(vendor_name);
CREATE INDEX mv_validation_results_plot_fhir_idx ON mv_validation_results_plot(fhir_version);
CREATE INDEX mv_validation_results_plot_rule_idx ON mv_validation_results_plot(rule_name);
CREATE INDEX mv_validation_results_plot_valid_idx ON mv_validation_results_plot(valid);
CREATE INDEX mv_validation_results_plot_reference_idx ON mv_validation_results_plot(reference);

DROP MATERIALIZED VIEW IF EXISTS mv_validation_failures CASCADE;

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