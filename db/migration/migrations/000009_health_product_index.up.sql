BEGIN;

CREATE INDEX IF NOT EXISTS healthit_product_name_version_idx ON healthit_products (name, version);

COMMIT;