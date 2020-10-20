BEGIN;

DROP INDEX IF EXISTS fhir_version_idx;
DROP INDEX IF EXISTS implementation_guide_idx;
DROP INDEX IF EXISTS field_idx;
DROP INDEX IF EXISTS exists_idx;
DROP INDEX IF EXISTS extension_idx;

DROP INDEX IF EXISTS resource_type_idx;

DROP INDEX IF EXISTS capstat_url_idx;
DROP INDEX IF EXISTS capstat_version_idx;
DROP INDEX IF EXISTS capstat_name_idx;
DROP INDEX IF EXISTS capstat_title_idx;
DROP INDEX IF EXISTS capstat_date_idx;
DROP INDEX IF EXISTS capstat_publisher_idx;
DROP INDEX IF EXISTS capstat_description_idx;
DROP INDEX IF EXISTS capstat_purpose_idx;
DROP INDEX IF EXISTS capstat_copyright_idx;

DROP INDEX IF EXISTS capstat_software_name_idx;
DROP INDEX IF EXISTS capstat_software_version_idx;
DROP INDEX IF EXISTS capstat_software_releaseDate_idx;
DROP INDEX IF EXISTS capstat_implementation_description_idx;
DROP INDEX IF EXISTS capstat_implementation_url_idx;
DROP INDEX IF EXISTS capstat_implementation_custodian_idx;

DROP INDEX IF EXISTS security_code_idx;
DROP INDEX IF EXISTS security_service_idx;

DROP INDEX IF EXISTS smart_capabilities_idx;

DROP INDEX IF EXISTS location_zipcode_idx;

COMMIT;