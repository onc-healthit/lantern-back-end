BEGIN;

DROP MATERIALIZED VIEW IF EXISTS security_endpoints_distinct_mv CASCADE;

CREATE MATERIALIZED VIEW security_endpoints_distinct_mv AS
SELECT DISTINCT
  url_modal AS url,
  condensed_organization_names,
  vendor_name,
  capability_fhir_version,
  tls_version,
  code
FROM selected_security_endpoints_mv;

-- Unique index
CREATE UNIQUE INDEX idx_unique_security_endpoints_distinct_mv ON security_endpoints_distinct_mv (url, condensed_organization_names, vendor_name, capability_fhir_version, tls_version, code);

-- Index to optimize common filter combinations
CREATE INDEX idx_security_endpoints_distinct_filters  ON security_endpoints_distinct_mv(capability_fhir_version, code, vendor_name);

COMMIT;