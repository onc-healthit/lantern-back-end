BEGIN;

CREATE INDEX healthit_product_name_version_idx ON healthit_products (name, version);

COMMIT;