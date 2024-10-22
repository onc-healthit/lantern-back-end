BEGIN;

-- validations Indexes
DROP INDEX IF EXISTS validations_val_res_id_idx;

-- fhir_endpoints Indexes
CREATE INDEX IF NOT EXISTS fhir_endpoint_url_index ON fhir_endpoints (url);

-- fhir_endpoints_info Indexes
DROP INDEX IF EXISTS fhir_endpoints_info_validation_result_id_idx; 

-- fhir_endpoints_info_history Indexes
DROP INDEX IF EXISTS fhir_endpoints_info_history_entered_at_idx;

DROP INDEX IF EXISTS fhir_endpoints_info_history_operation_idx;

DROP INDEX IF EXISTS fhir_endpoints_info_history_requested_fhir_version_idx;

-- healthit_products Indexes
DROP INDEX IF EXISTS healthit_products_certification_status_idx;

DROP INDEX IF EXISTS healthit_products_chpl_id_idx;

-- fhir_endpoint_organizations_map Indexes
DROP INDEX IF EXISTS fhir_endpoint_organizations_map_id_idx;

DROP INDEX IF EXISTS fhir_endpoint_organizations_map_org_database_id_idx;

COMMIT;