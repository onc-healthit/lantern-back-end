BEGIN;

ALTER TABLE IF EXISTS public.healthit_products_map DROP CONSTRAINT IF EXISTS unique_id_healthit_product;


COMMIT;