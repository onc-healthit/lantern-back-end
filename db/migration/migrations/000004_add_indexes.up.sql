BEGIN;

CREATE INDEX IF NOT EXISTS vendor_name_idx ON vendors (name);
CREATE INDEX IF NOT EXISTS fhir_version_idx ON fhir_endpoints_info ((capability_statement->>'fhirVersion'));
CREATE INDEX IF NOT EXISTS implementation_guide_idx ON fhir_endpoints_info ((capability_statement->>'implementationGuide'));
CREATE INDEX IF NOT EXISTS field_idx ON fhir_endpoints_info ((included_fields->> 'Field'));
CREATE INDEX IF NOT EXISTS exists_idx ON fhir_endpoints_info ((included_fields->> 'Exists'));
CREATE INDEX IF NOT EXISTS extension_idx ON fhir_endpoints_info ((included_fields->> 'Extension'));

CREATE INDEX IF NOT EXISTS resource_type_idx ON fhir_endpoints_info (((capability_statement::json#>'{rest,0,resource}') ->> 'type'));

CREATE INDEX IF NOT EXISTS capstat_url_idx ON fhir_endpoints_info ((capability_statement->>'url'));
CREATE INDEX IF NOT EXISTS capstat_version_idx ON fhir_endpoints_info ((capability_statement->>'version'));
CREATE INDEX IF NOT EXISTS capstat_name_idx ON fhir_endpoints_info ((capability_statement->>'name'));
CREATE INDEX IF NOT EXISTS capstat_title_idx ON fhir_endpoints_info ((capability_statement->>'title'));
CREATE INDEX IF NOT EXISTS capstat_date_idx ON fhir_endpoints_info ((capability_statement->>'date'));
CREATE INDEX IF NOT EXISTS capstat_publisher_idx ON fhir_endpoints_info ((capability_statement->>'publisher'));
CREATE INDEX IF NOT EXISTS capstat_description_idx ON fhir_endpoints_info ((capability_statement->>'description'));
CREATE INDEX IF NOT EXISTS capstat_purpose_idx ON fhir_endpoints_info ((capability_statement->>'purpose'));
CREATE INDEX IF NOT EXISTS capstat_copyright_idx ON fhir_endpoints_info ((capability_statement->>'copyright'));

CREATE INDEX IF NOT EXISTS capstat_software_name_idx ON fhir_endpoints_info ((capability_statement->'software'->>'name'));
CREATE INDEX IF NOT EXISTS capstat_software_version_idx ON fhir_endpoints_info ((capability_statement->'software'->>'version'));
CREATE INDEX IF NOT EXISTS capstat_software_releaseDate_idx ON fhir_endpoints_info ((capability_statement->'software'->>'releaseDate'));
CREATE INDEX IF NOT EXISTS capstat_implementation_description_idx ON fhir_endpoints_info ((capability_statement->'implementation'->>'description'));
CREATE INDEX IF NOT EXISTS capstat_implementation_url_idx ON fhir_endpoints_info ((capability_statement->'implementation'->>'url'));
CREATE INDEX IF NOT EXISTS capstat_implementation_custodian_idx ON fhir_endpoints_info ((capability_statement->'implementation'->>'custodian'));

CREATE INDEX IF NOT EXISTS security_code_idx ON fhir_endpoints_info ((capability_statement::json#>'{rest,0,security,service}'->'coding'->>'code'));
CREATE INDEX IF NOT EXISTS security_service_idx ON fhir_endpoints_info ((capability_statement::json#>'{rest,0,security}' -> 'service' ->> 'text'));

CREATE INDEX IF NOT EXISTS smart_capabilities_idx ON fhir_endpoints_info ((smart_response->'capabilities'));

CREATE INDEX IF NOT EXISTS location_zipcode_idx ON npi_organizations ((location->>'zipcode'));

COMMIT;