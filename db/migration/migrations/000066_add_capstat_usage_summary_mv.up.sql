BEGIN;

DROP MATERIALIZED VIEW IF EXISTS capstat_usage_summary_mv CASCADE;

CREATE MATERIALIZED VIEW capstat_usage_summary_mv AS
SELECT 
  field,
  "FHIR Version",
  "Developer",
  is_used,
  SUM("Endpoints") AS count
FROM selected_fhir_endpoints_values_mv
GROUP BY field, "FHIR Version", "Developer", is_used;

CREATE INDEX idx_usage_summary_filters ON capstat_usage_summary_mv(field, "FHIR Version", "Developer", is_used);

COMMIT;