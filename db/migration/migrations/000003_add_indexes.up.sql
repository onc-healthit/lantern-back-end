BEGIN;

CREATE INDEX IF NOT EXISTS fhir_endpoints_url_idx ON fhir_endpoints (url);
CREATE INDEX IF NOT EXISTS fhir_endpoints_info_url_idx ON fhir_endpoints_info (url);
CREATE INDEX IF NOT EXISTS fhir_endpoints_info_history_url_idx ON fhir_endpoints_info_history (url);
CREATE INDEX IF NOT EXISTS  endpoint_organization_url_idx ON endpoint_organization (url);

CREATE INDEX IF NOT EXISTS vendor_id_idx ON vendors (id);
CREATE INDEX IF NOT EXISTS fhir_endpoints_info_vendor_id_idx ON fhir_endpoints_info (vendor_id);
CREATE INDEX IF NOT EXISTS  fhir_endpoints_info_history_vendor_id_idx ON fhir_endpoints_info_history (vendor_id);

CREATE INDEX IF NOT EXISTS npi_organizations_npi_id_idx ON npi_organizations (npi_id);
CREATE INDEX IF NOT EXISTS endpoint_organization_npi_id_idx ON endpoint_organization (organization_npi_id);

COMMIT;