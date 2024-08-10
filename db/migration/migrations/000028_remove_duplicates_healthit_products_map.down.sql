BEGIN;

ALTER TABLE public.healthit_products_map DROP CONSTRAINT unique_id_healthit_product;


COMMIT;