BEGIN;

ALTER TABLE fhir_endpoints_info ADD COLUMN IF NOT EXISTS healthit_product_id INT REFERENCES healthit_products(id) ON DELETE SET NULL; 
ALTER TABLE fhir_endpoints_info_history ADD COLUMN IF NOT EXISTS healthit_product_id INT REFERENCES healthit_products(id) ON DELETE SET NULL;

CREATE OR REPLACE FUNCTION populate_existing_products_info_history() RETURNS VOID as $$
    DECLARE
        t_curs cursor for select id, healthit_product_id from healthit_products_map;
        t_row healthit_products_map%ROWTYPE;
    BEGIN
        FOR t_row in t_curs LOOP
            UPDATE fhir_endpoints_info_history SET healthit_product_id=t_row.healthit_product_id WHERE healthit_mapping_id = t_row.id;
        END LOOP;
    END
$$ LANGUAGE plpgsql;

SELECT populate_existing_products_info_history();

CREATE OR REPLACE FUNCTION populate_existing_products_info() RETURNS VOID as $$
    DECLARE
        t_curs cursor for select id, healthit_product_id from healthit_products_map;
        t_row healthit_products_map%ROWTYPE;
    BEGIN
        FOR t_row in t_curs LOOP
            UPDATE fhir_endpoints_info SET healthit_product_id=t_row.healthit_product_id WHERE healthit_mapping_id = t_row.id;
        END LOOP;
    END
$$ LANGUAGE plpgsql;

SELECT populate_existing_products_info();

DROP TABLE IF EXISTS healthit_products_map;

ALTER TABLE fhir_endpoints_info DROP COLUMN IF EXISTS healthit_mapping_id CASCADE;
ALTER TABLE fhir_endpoints_info_history DROP COLUMN IF EXISTS healthit_mapping_id CASCADE;
 
COMMIT;