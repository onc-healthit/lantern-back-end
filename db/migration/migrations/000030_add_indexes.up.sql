BEGIN;

-- validations Indexes
CREATE INDEX IF NOT EXISTS validations_val_res_id_idx ON validations (validation_result_id);

-- fhir_endpoints Indexes
DROP INDEX IF EXISTS fhir_endpoint_url_index;

-- fhir_endpoints_info Indexes
CREATE INDEX IF NOT EXISTS fhir_endpoints_info_validation_result_id_idx ON fhir_endpoints_info (validation_result_id); 

-- fhir_endpoints_info_history Indexes
CREATE INDEX IF NOT EXISTS fhir_endpoints_info_history_entered_at_idx ON fhir_endpoints_info_history (entered_at);

CREATE INDEX IF NOT EXISTS fhir_endpoints_info_history_operation_idx ON fhir_endpoints_info_history (operation);

CREATE INDEX IF NOT EXISTS fhir_endpoints_info_history_requested_fhir_version_idx ON fhir_endpoints_info_history (requested_fhir_version);

-- healthit_products Indexes
CREATE INDEX IF NOT EXISTS healthit_products_certification_status_idx ON healthit_products (certification_status);

CREATE INDEX IF NOT EXISTS healthit_products_chpl_id_idx ON healthit_products (chpl_id);

-- fhir_endpoint_organizations_map Indexes
CREATE INDEX IF NOT EXISTS fhir_endpoint_organizations_map_id_idx ON fhir_endpoint_organizations_map (id);

CREATE INDEX IF NOT EXISTS fhir_endpoint_organizations_map_org_database_id_idx ON fhir_endpoint_organizations_map (org_database_id);

COMMIT;