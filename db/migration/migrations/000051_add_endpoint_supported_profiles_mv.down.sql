BEGIN;

DROP INDEX IF EXISTS endpoint_supported_profiles_mv_uidx;
DROP INDEX IF EXISTS idx_profiles_fhir_version;
DROP INDEX IF EXISTS idx_profiles_vendor_name;
DROP INDEX IF EXISTS idx_profiles_profileurl;

DROP MATERIALIZED VIEW IF EXISTS endpoint_supported_profiles_mv CASCADE;

COMMIT;
