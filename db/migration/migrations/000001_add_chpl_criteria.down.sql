BEGIN;

DROP TABLE IF EXISTS certification_criteria;
DROP TABLE IF EXISTS product_criteria;

DROP TRIGGER IF EXISTS set_timestamp_certification_criteria ON certification_criteria;
DROP TRIGGER IF EXISTS set_timestamp_product_criteria ON product_criteria;

CREATE OR REPLACE FUNCTION delete_data_in_healthit_products() RETURNS VOID as $$
    BEGIN
        DELETE FROM healthit_products;
    END;
$$ LANGUAGE plpgsql;

SELECT delete_data_in_healthit_products();

COMMIT;