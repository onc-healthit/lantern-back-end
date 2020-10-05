BEGIN;

DROP INDEX IF EXISTS fhir_endpoints_url_idx;
DROP INDEX IF EXISTS fhir_endpoints_info_url_idx;
DROP INDEX IF EXISTS fhir_endpoints_info_history_url_idx;
DROP INDEX IF EXISTS endpoint_organization_url_idx;

DROP INDEX IF EXISTS vendor_id_idx;
DROP INDEX IF EXISTS fhir_endpoints_info_vendor_id_idx;
DROP INDEX IF EXISTS fhir_endpoints_info_history_vendor_id_idx;

DROP INDEX IF EXISTS npi_organizations_npi_id_idx;
DROP INDEX IF EXISTS endpoint_organization_npi_id_idx;


COMMIT;
