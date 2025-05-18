BEGIN;

DROP INDEX IF EXISTS fhir_endpoints_info_history_val_res_idx;

ALTER TABLE validations
DROP CONSTRAINT IF EXISTS fk_validations_validation_results;

COMMIT;