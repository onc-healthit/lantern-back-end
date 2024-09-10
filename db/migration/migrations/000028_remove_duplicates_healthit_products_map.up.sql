BEGIN;

CREATE OR REPLACE FUNCTION delete_duplicate_data_in_healthit_products_map() RETURNS VOID as $$
    BEGIN
        WITH CTE AS (
    SELECT 
        ctid,
        ROW_NUMBER() OVER(PARTITION BY id, healthit_product_id ORDER BY id) AS DuplicateCount
    FROM 
        public.healthit_products_map
)
DELETE FROM public.healthit_products_map
WHERE ctid IN (
    SELECT ctid
    FROM CTE
    WHERE DuplicateCount > 1
);
    END;
$$ LANGUAGE plpgsql;

SELECT delete_duplicate_data_in_healthit_products_map();

ALTER TABLE IF EXISTS public.healthit_products_map ADD CONSTRAINT unique_id_healthit_product UNIQUE (id, healthit_product_id);

COMMIT;