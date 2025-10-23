BEGIN;

DROP TABLE IF EXISTS list_source_info;
CREATE TABLE IF NOT EXISTS list_source_info (
    list_source             VARCHAR(500),
    is_chpl                    BOOLEAN
);

COMMIT;