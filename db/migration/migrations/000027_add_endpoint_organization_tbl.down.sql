BEGIN;

DROP TABLE IF EXISTS fhir_endpoint_organizations;
DROP TABLE IF EXISTS fhir_endpoint_organizations_map;
DROP VIEW IF EXISTS endpoint_export;
DROP VIEW IF EXISTS organization_location;
DROP TRIGGER IF EXISTS set_timestamp_fhir_endpoint_organizations;

ALTER TABLE fhir_endpoints DROP COLUMN IF EXISTS org_database_map_id CASCADE;
ALTER TABLE fhir_endpoints ADD COLUMN IF NOT EXISTS organization_names VARCHAR(500)[]; 
ALTER TABLE fhir_endpoints ADD COLUMN IF NOT EXISTS npi_ids VARCHAR(500)[]; 

CREATE or REPLACE VIEW endpoint_export AS
SELECT endpts.url, endpts.list_source, endpts.organization_names AS endpoint_names,
    vendors.name as vendor_name,
    endpts_info.tls_version, endpts_info.mime_types, endpts_metadata.http_response,
    endpts_metadata.response_time_seconds, endpts_metadata.smart_http_response, endpts_metadata.errors,
    EXISTS (SELECT 1 FROM fhir_endpoints_info WHERE capability_statement::jsonb != 'null' AND endpts.url = fhir_endpoints_info.url) as CAP_STAT_EXISTS,
    endpts_info.capability_fhir_version AS FHIR_VERSION,
    endpts_info.capability_statement->>'publisher' AS PUBLISHER,
    endpts_info.capability_statement->'software'->'name' AS SOFTWARE_NAME,
    endpts_info.capability_statement->'software'->'version' AS SOFTWARE_VERSION,
    endpts_info.capability_statement->'software'->'releaseDate' AS SOFTWARE_RELEASEDATE,
    endpts_info.capability_statement->'format' AS FORMAT,
    endpts_info.capability_statement->>'kind' AS KIND,
    endpts_info.updated_at AS INFO_UPDATED, endpts_info.created_at AS INFO_CREATED,
    endpts_info.requested_fhir_version, endpts_metadata.availability,
    list_source_info.is_chpl
FROM fhir_endpoints AS endpts
LEFT JOIN fhir_endpoints_info AS endpts_info ON endpts.url = endpts_info.url
LEFT JOIN fhir_endpoints_metadata AS endpts_metadata ON endpts_info.metadata_id = endpts_metadata.id
LEFT JOIN vendors ON endpts_info.vendor_id = vendors.id
LEFT JOIN list_source_info ON endpts.list_source = list_source_info.list_source;

CREATE or REPLACE VIEW organization_location AS
    SELECT endpts.url, endpts.organization_names AS endpoint_names, endpts_info.capability_fhir_version AS FHIR_VERSION, 
    endpts_info.requested_fhir_version, vendors.name as vendor_name, orgs.name AS ORGANIZATION_NAME, orgs.secondary_name AS ORGANIZATION_SECONDARY_NAME,
    orgs.taxonomy, orgs.Location->>'state' AS STATE, orgs.Location->>'zipcode' AS ZIPCODE, orgs.npi_id as NPI_ID,
    links.confidence AS MATCH_SCORE
FROM endpoint_organization AS links
RIGHT JOIN fhir_endpoints AS endpts ON links.url = endpts.url
LEFT JOIN fhir_endpoints_info AS endpts_info ON endpts.url = endpts_info.url   
LEFT JOIN vendors ON endpts_info.vendor_id = vendors.id
LEFT JOIN npi_organizations AS orgs ON links.organization_npi_id = orgs.npi_id
WHERE links.confidence > .97 AND orgs.Location->>'zipcode' IS NOT null;

DROP INDEX IF EXISTS fhir_endpoint_organizations_id;
DROP INDEX IF EXISTS fhir_endpoint_organizations_name;
DROP INDEX IF EXISTS fhir_endpoint_organizations_zipcode;
DROP INDEX IF EXISTS fhir_endpoint_organizations_npi_id;

COMMIT;