BEGIN;

CREATE OR REPLACE FUNCTION delete_data_in_healthit_products() RETURNS VOID as $$
    BEGIN
        DELETE FROM validations WHERE rule_name = 'generalMimeType';
    END;
$$ LANGUAGE plpgsql;

SELECT delete_data_in_healthit_products();

END;