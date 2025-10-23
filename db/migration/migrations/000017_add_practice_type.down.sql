BEGIN;

ALTER TABLE healthit_products DROP COLUMN IF EXISTS practice_type CASCADE; 

COMMIT;