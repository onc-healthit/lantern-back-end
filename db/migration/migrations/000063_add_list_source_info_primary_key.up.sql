BEGIN;

-- Add primary key constraint to list_source_info table
ALTER TABLE list_source_info ADD CONSTRAINT list_source_info_pkey PRIMARY KEY (list_source);

COMMIT;