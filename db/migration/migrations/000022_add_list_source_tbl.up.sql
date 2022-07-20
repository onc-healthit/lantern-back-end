BEGIN;

CREATE OR REPLACE FUNCTION create_list_source_table() RETURNS VOID as $$
    BEGIN
        CREATE TABLE IF NOT EXISTS list_source_info (
            list_source             VARCHAR(500),
            is_chpl                 BOOLEAN
);
    END;
$$ LANGUAGE plpgsql;

SELECT create_list_source_table();

END;