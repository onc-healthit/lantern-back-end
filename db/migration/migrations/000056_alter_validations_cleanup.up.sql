BEGIN;

DROP INDEX IF EXISTS fhir_endpoints_info_history_val_res_idx;
CREATE INDEX fhir_endpoints_info_history_val_res_idx ON fhir_endpoints_info_history (validation_result_id);

-- First drop the existing constraint
ALTER TABLE validations
DROP CONSTRAINT IF EXISTS fk_validations_validation_results;

-- Then add the new constraint with CASCADE
ALTER TABLE validations
ADD CONSTRAINT fk_validations_validation_results
FOREIGN KEY (validation_result_id) 
REFERENCES validation_results(id)
ON DELETE CASCADE;

COMMIT;