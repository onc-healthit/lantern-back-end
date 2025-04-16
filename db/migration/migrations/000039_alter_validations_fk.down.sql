BEGIN;

ALTER TABLE validations 
DROP CONSTRAINT validations_validation_result_id_fkey;

COMMIT;