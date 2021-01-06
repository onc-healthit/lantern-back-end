BEGIN;

DROP VIEW IF EXISTS endpoint_export;

DROP TRIGGER IF EXISTS update_fhir_endpoint_availability_trigger ON fhir_endpoints_info;
DROP TRIGGER IF EXISTS update_fhir_endpoint_availability_trigger ON fhir_endpoints_metadata;
DROP TRIGGER IF EXISTS set_timestamp_fhir_endpoints_metadata ON fhir_endpoints_metadata;

ALTER TABLE fhir_endpoints_info 
ADD COLUMN http_response INTEGER, 
ADD COLUMN availability DECIMAL(5,4), 
ADD COLUMN errors VARCHAR(500), 
ADD COLUMN response_time_seconds DECIMAL(7,4), 
ADD COLUMN smart_http_response INTEGER;

ALTER TABLE fhir_endpoints_info_history 
ADD COLUMN http_response INTEGER, 
ADD COLUMN availability DECIMAL(5,4), 
ADD COLUMN errors VARCHAR(500), 
ADD COLUMN response_time_seconds DECIMAL(7,4), 
ADD COLUMN smart_http_response INTEGER;

ALTER TABLE fhir_endpoints_info
DISABLE TRIGGER add_fhir_endpoint_info_history_trigger;


CREATE OR REPLACE FUNCTION populate_existing_tables_endpoints_info() RETURNS VOID as $$
    DECLARE
        i RECORD;
    BEGIN
        FOR i IN SELECT DISTINCT fhir_endpoints_metadata.id, fhir_endpoints_metadata.http_response, fhir_endpoints_metadata.availability, fhir_endpoints_metadata.errors, fhir_endpoints_metadata.response_time_seconds, fhir_endpoints_metadata.smart_http_response FROM fhir_endpoints_metadata
        LOOP
            UPDATE fhir_endpoints_info SET http_response=i.http_response, availability=i.availability, errors=i.errors, response_time_seconds=i.response_time_seconds, smart_http_response=i.smart_http_response WHERE metadata_id = i.id;
            UPDATE fhir_endpoints_info_history SET http_response=i.http_response, availability=i.availability, errors=i.errors, response_time_seconds=i.response_time_seconds, smart_http_response=i.smart_http_response WHERE metadata_id = i.id;
        END LOOP;
    END
$$ LANGUAGE plpgsql;

SELECT populate_existing_tables_endpoints_info();

ALTER TABLE fhir_endpoints_info 
DROP COLUMN metadata_id;

ALTER TABLE fhir_endpoints_info_history
DROP COLUMN metadata_id;

ALTER TABLE fhir_endpoints_info
ENABLE TRIGGER add_fhir_endpoint_info_history_trigger;

DROP TABLE IF EXISTS fhir_endpoints_metadata;

CREATE or REPLACE VIEW endpoint_export AS
SELECT endpts.url, endpts.list_source, endpts.organization_names AS endpoint_names,
    vendors.name as vendor_name,
    endpts_info.tls_version, endpts_info.mime_types, endpts_info.http_response,
    endpts_info.response_time_seconds, endpts_info.smart_http_response, endpts_info.errors,
    endpts_info.capability_statement->>'fhirVersion' AS FHIR_VERSION,
    endpts_info.capability_statement->>'publisher' AS PUBLISHER,
    endpts_info.capability_statement->'software'->'name' AS SOFTWARE_NAME,
    endpts_info.capability_statement->'software'->'version' AS SOFTWARE_VERSION,
    endpts_info.capability_statement->'software'->'releaseDate' AS SOFTWARE_RELEASEDATE,
    endpts_info.updated_at AS INFO_UPDATED, endpts_info.created_at AS INFO_CREATED,
    orgs.name AS ORGANIZATION_NAME, orgs.secondary_name AS ORGANIZATION_SECONDARY_NAME,
    orgs.taxonomy, orgs.Location->>'state' AS STATE, orgs.Location->>'zipcode' AS ZIPCODE,
    links.confidence AS MATCH_SCORE, endpts_info.supported_resources,
    endpts_info.availability
FROM endpoint_organization AS links
RIGHT JOIN fhir_endpoints AS endpts ON links.url = endpts.url
LEFT JOIN fhir_endpoints_info AS endpts_info ON endpts.url = endpts_info.url
LEFT JOIN vendors ON endpts_info.vendor_id = vendors.id
LEFT JOIN npi_organizations AS orgs ON links.organization_npi_id = orgs.npi_id;

-- increments total number of times http status returned for endpoint 
CREATE TRIGGER update_fhir_endpoint_availability_trigger
BEFORE INSERT OR UPDATE on fhir_endpoints_info
FOR EACH ROW
EXECUTE PROCEDURE update_fhir_endpoint_availability_info();

COMMIT;