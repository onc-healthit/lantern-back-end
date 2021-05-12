BEGIN;
DROP VIEW IF EXISTS endpoint_export;
DROP INDEX IF EXISTS capability_fhir_version_idx;
DROP INDEX IF EXISTS requested_fhir_version_idx;

CREATE OR REPLACE FUNCTION delete_requested_version_entries() RETURNS VOID as $$
    BEGIN
        DELETE FROM fhir_endpoints_info WHERE requested_fhir_version != '';
        DELETE FROM fhir_endpoints_info_history WHERE requested_fhir_version != '';
    END;
$$ LANGUAGE plpgsql;

SELECT delete_requested_version_entries();

ALTER TABLE fhir_endpoints_info DROP COLUMN IF EXISTS requested_fhir_version; 
ALTER TABLE fhir_endpoints_info_history DROP COLUMN IF EXISTS requested_fhir_version; 
ALTER TABLE fhir_endpoints_info DROP COLUMN IF EXISTS capability_fhir_version;
ALTER TABLE fhir_endpoints_info_history DROP COLUMN IF EXISTS capability_fhir_version;

ALTER TABLE fhir_endpoints_info ADD UNIQUE (url);

CREATE or REPLACE VIEW org_mapping AS
SELECT endpts.url, endpts.list_source, vendors.name as vendor_name, endpts.organization_names AS endpoint_names, orgs.name AS ORGANIZATION_NAME, orgs.secondary_name AS ORGANIZATION_SECONDARY_NAME, orgs.taxonomy, orgs.Location->>'state' AS STATE, orgs.Location->>'zipcode' AS ZIPCODE, links.confidence AS MATCH_SCORE
FROM endpoint_organization AS links
LEFT JOIN fhir_endpoints AS endpts ON links.url = endpts.url
LEFT JOIN fhir_endpoints_info AS endpts_info ON endpts.url = endpts_info.url
LEFT JOIN vendors ON endpts_info.vendor_id = vendors.id
LEFT JOIN npi_organizations AS orgs ON links.organization_npi_id = orgs.npi_id;

CREATE or REPLACE VIEW endpoint_export AS
SELECT endpts.url, endpts.list_source, endpts.organization_names AS endpoint_names,
    vendors.name as vendor_name,
    endpts_info.tls_version, endpts_info.mime_types, endpts_metadata.http_response,
    endpts_metadata.response_time_seconds, endpts_metadata.smart_http_response, endpts_metadata.errors,
    endpts_info.capability_statement->>'fhirVersion' AS FHIR_VERSION,
    endpts_info.capability_statement->>'publisher' AS PUBLISHER,
    endpts_info.capability_statement->'software'->'name' AS SOFTWARE_NAME,
    endpts_info.capability_statement->'software'->'version' AS SOFTWARE_VERSION,
    endpts_info.capability_statement->'software'->'releaseDate' AS SOFTWARE_RELEASEDATE,
    endpts_info.updated_at AS INFO_UPDATED, endpts_info.created_at AS INFO_CREATED,
    orgs.name AS ORGANIZATION_NAME, orgs.secondary_name AS ORGANIZATION_SECONDARY_NAME,
    orgs.taxonomy, orgs.Location->>'state' AS STATE, orgs.Location->>'zipcode' AS ZIPCODE,
    links.confidence AS MATCH_SCORE, endpts_info.operation_resource,
    endpts_metadata.availability
FROM endpoint_organization AS links
RIGHT JOIN fhir_endpoints AS endpts ON links.url = endpts.url
LEFT JOIN fhir_endpoints_info AS endpts_info ON endpts.url = endpts_info.url
LEFT JOIN fhir_endpoints_metadata AS endpts_metadata ON endpts_info.metadata_id = endpts_metadata.id
LEFT JOIN vendors ON endpts_info.vendor_id = vendors.id
LEFT JOIN npi_organizations AS orgs ON links.organization_npi_id = orgs.npi_id;

CREATE INDEX fhir_version_idx ON fhir_endpoints_info ((capability_statement->>'fhirVersion'));
ALTER TABLE fhir_endpoints_info DROP CONSTRAINT fhir_endpoints_info_unique;

COMMIT;