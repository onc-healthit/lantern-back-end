BEGIN;

-- Drop indexes
DROP INDEX IF EXISTS idx_payer_info_edi_id;
DROP INDEX IF EXISTS idx_payer_info_url;
DROP INDEX IF EXISTS idx_payer_endpoints_url;
DROP INDEX IF EXISTS idx_payer_endpoints_payer_id;
DROP INDEX IF EXISTS idx_payers_email;

-- Drop triggers
DROP TRIGGER IF EXISTS set_timestamp_payer_info ON payer_info;
DROP TRIGGER IF EXISTS set_timestamp_payer_endpoints ON payer_endpoints;
DROP TRIGGER IF EXISTS set_timestamp_payers ON payers;

-- Drop tables (in reverse order due to foreign key constraints)
DROP TABLE IF EXISTS payer_info;
DROP TABLE IF EXISTS payer_endpoints;
DROP TABLE IF EXISTS payers;

COMMIT;