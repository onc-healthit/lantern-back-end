BEGIN;

ALTER TABLE validations 
DROP CONSTRAINT validations_validation_result_id_fkey;

ALTER TABLE validations 
ADD CONSTRAINT validations_validation_result_id_fkey
FOREIGN KEY (validation_result_id) 
REFERENCES validation_results(id) 
ON DELETE CASCADE;


COMMIT;