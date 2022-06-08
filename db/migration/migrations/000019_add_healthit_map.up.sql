BEGIN;

CREATE TABLE IF NOT EXISTS healthit_products_map (
    id SERIAL,
    healthit_product_id INT REFERENCES healthit_products(id) ON DELETE SET NULL
);

ALTER TABLE fhir_endpoints_info ADD COLUMN IF NOT EXISTS healthit_mapping_id INT;
ALTER TABLE fhir_endpoints_info_history ADD COLUMN IF NOT EXISTS healthit_mapping_id INT;

CREATE OR REPLACE FUNCTION populate_endpoints_products_info_history() RETURNS VOID as $$
    DECLARE
        t_curs cursor for select healthit_product_id from fhir_endpoints_info_history where healthit_product_id is not NULL;
        t_row fhir_endpoints_info_history%ROWTYPE;
        j INTEGER;
    BEGIN
        FOR t_row in t_curs LOOP
            INSERT INTO healthit_products_map (healthit_product_id) VALUES (t_row.healthit_product_id);
            SELECT currval(pg_get_serial_sequence('healthit_products_map','id')) INTO j;
            UPDATE fhir_endpoints_info_history SET healthit_mapping_id = j WHERE current of t_curs; 
        END LOOP;
    END;
$$ LANGUAGE plpgsql;
SELECT populate_endpoints_products_info_history();

CREATE OR REPLACE FUNCTION populate_endpoints_products_info() RETURNS VOID as $$
    DECLARE
        t_curs cursor for select healthit_product_id from fhir_endpoints_info where healthit_product_id is not NULL;
        t_row fhir_endpoints_info%ROWTYPE;
        j INTEGER;
    BEGIN
        FOR t_row in t_curs LOOP
            INSERT INTO healthit_products_map (healthit_product_id) VALUES (t_row.healthit_product_id);
            SELECT currval(pg_get_serial_sequence('healthit_products_map','id')) INTO j;
            UPDATE fhir_endpoints_info SET healthit_mapping_id = j WHERE current of t_curs; 
        END LOOP;
    END;
$$ LANGUAGE plpgsql;
SELECT populate_endpoints_products_info();

ALTER TABLE fhir_endpoints_info_history DROP COLUMN IF EXISTS healthit_product_id CASCADE;
ALTER TABLE fhir_endpoints_info DROP COLUMN IF EXISTS healthit_product_id CASCADE;

COMMIT;
