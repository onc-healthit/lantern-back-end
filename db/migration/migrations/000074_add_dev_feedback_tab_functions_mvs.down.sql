BEGIN;

DROP FUNCTION IF EXISTS validate_npi_luhn(TEXT);

DROP FUNCTION IF EXISTS is_address_like(TEXT);

DROP FUNCTION IF EXISTS validate_identifier_value(TEXT, TEXT);

DROP FUNCTION IF EXISTS is_valid_organization_name(TEXT);

DROP FUNCTION IF EXISTS is_valid_organization_address(TEXT);

DROP MATERIALIZED VIEW IF EXISTS mv_organization_quality CASCADE;

DROP MATERIALIZED VIEW IF EXISTS mv_organization_quality_summary CASCADE;

DROP MATERIALIZED VIEW IF EXISTS mv_organization_identifier_summary CASCADE;

COMMIT;