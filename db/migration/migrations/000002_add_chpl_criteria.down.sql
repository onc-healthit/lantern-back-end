BEGIN;

DROP TABLE IF EXISTS certification_criteria;
DROP TABLE IF EXISTS product_criteria;

DROP TRIGGER IF EXISTS set_timestamp_certification_criteria ON certification_criteria;
DROP TRIGGER IF EXISTS set_timestamp_product_criteria ON product_criteria;

COMMIT;