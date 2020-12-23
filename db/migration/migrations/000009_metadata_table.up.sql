BEGIN;

DROP VIEW IF EXISTS endpoint_export;

DROP TABLE IF EXISTS fhir_endpoints_metadata;

CREATE TABLE IF NOT EXISTS fhir_endpoints_metadata AS
    SELECT id, url, http_response, availability, errors, response_time_seconds, smart_http_response, created_at, updated_at   
    FROM fhir_endpoints_info;

ALTER TABLE fhir_endpoints_info 
DROP COLUMN http_response, 
DROP COLUMN availability, 
DROP COLUMN errors, 
DROP COLUMN response_time_seconds, 
DROP COLUMN smart_http_response;

ALTER TABLE fhir_endpoints_info_history 
DROP COLUMN http_response, 
DROP COLUMN availability, 
DROP COLUMN errors, 
DROP COLUMN response_time_seconds, 
DROP COLUMN smart_http_response;


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
    links.confidence AS MATCH_SCORE, endpts_info.supported_resources,
    endpts_metadata.availability
FROM endpoint_organization AS links
RIGHT JOIN fhir_endpoints AS endpts ON links.url = endpts.url
LEFT JOIN fhir_endpoints_info AS endpts_info ON endpts.url = endpts_info.url
LEFT JOIN fhir_endpoints_metadata AS endpts_metadata ON endpts.url = endpts_metadata.url
LEFT JOIN vendors ON endpts_info.vendor_id = vendors.id
LEFT JOIN npi_organizations AS orgs ON links.organization_npi_id = orgs.npi_id;

DROP TRIGGER IF EXISTS update_fhir_endpoint_availability_trigger ON fhir_endpoints_info;
DROP TRIGGER IF EXISTS update_fhir_endpoint_availability_trigger ON fhir_endpoints_metadata;
DROP TRIGGER IF EXISTS set_timestamp_fhir_endpoints_metadata ON fhir_endpoints_metadata;

-- increments total number of times http status returned for endpoint 
CREATE TRIGGER update_fhir_endpoint_availability_trigger
BEFORE INSERT OR UPDATE on fhir_endpoints_metadata
FOR EACH ROW
EXECUTE PROCEDURE update_fhir_endpoint_availability_info();

CREATE TRIGGER set_timestamp_fhir_endpoints_metadata
BEFORE UPDATE ON fhir_endpoints_metadata
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

COMMIT;
