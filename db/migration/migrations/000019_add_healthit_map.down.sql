BEGIN;

DROP TABLE IF EXISTS healthit_products_map;

ALTER TABLE fhir_endpoints_info DROP COLUMN IF EXISTS healthit_mapping_id CASCADE;
ALTER TABLE fhir_endpoints_info ADD COLUMN IF NOT EXISTS healthit_product_id INT REFERENCES healthit_products(id) ON DELETE SET NULL; 

ALTER TABLE fhir_endpoints_info_history DROP COLUMN IF EXISTS healthit_mapping_id CASCADE;
ALTER TABLE fhir_endpoints_info_history ADD COLUMN IF NOT EXISTS healthit_product_id INT REFERENCES healthit_products(id) ON DELETE SET NULL; 

COMMIT;