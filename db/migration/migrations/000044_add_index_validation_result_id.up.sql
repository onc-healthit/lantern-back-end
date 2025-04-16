BEGIN;

DROP INDEX IF EXISTS fhir_endpoints_info_history_val_res_idx;

CREATE INDEX fhir_endpoints_info_history_val_res_idx ON fhir_endpoints_info_history (validation_result_id);

COMMIT;