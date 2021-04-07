BEGIN;

DROP TRIGGER IF EXISTS add_fhir_endpoint_info_history_trigger ON fhir_endpoints_info;
DROP VIEW IF EXISTS endpoint_export;
DROP VIEW IF EXISTS org_mapping;

DROP INDEX IF EXISTS fhir_version_idx;

ALTER TABLE fhir_endpoints_info 
ADD COLUMN requested_fhir_version VARCHAR(500);

ALTER TABLE fhir_endpoints_info_history 
ADD COLUMN requested_fhir_version VARCHAR(500);

ALTER TABLE fhir_endpoints_info 
ADD COLUMN capability_fhir_version VARCHAR(500);

ALTER TABLE fhir_endpoints_info_history 
ADD COLUMN capability_fhir_version VARCHAR(500);

ALTER TABLE fhir_endpoints_info DROP CONSTRAINT fhir_endpoints_info_url_key;
ALTER TABLE fhir_endpoints_info ADD CONSTRAINT fhir_endpoints_info_unique UNIQUE(url, requested_fhir_version);

CREATE OR REPLACE FUNCTION populate_capability_fhir_version_info() RETURNS VOID as $$
    DECLARE
        t_curs cursor for select capability_statement from fhir_endpoints_info;
        t_row fhir_endpoints_info%ROWTYPE;
        capStatVersion VARCHAR(500);
    BEGIN
        FOR t_row in t_curs LOOP
            SELECT cast(coalesce(nullif(t_row.capability_statement->>'fhirVersion',NULL),'') as varchar(500)) INTO capStatVersion;
            UPDATE fhir_endpoints_info SET requested_fhir_version = '', capability_fhir_version = capStatVersion WHERE current of t_curs; 
        END LOOP;
    END;
$$ LANGUAGE plpgsql;

SELECT populate_capability_fhir_version_info();

CREATE OR REPLACE FUNCTION populate_capability_fhir_version_info_history() RETURNS VOID as $$
    DECLARE
        t_curs cursor for select capability_statement from fhir_endpoints_info_history;
        t_row fhir_endpoints_info%ROWTYPE;
        capStatVersion VARCHAR(500);
    BEGIN
        FOR t_row in t_curs LOOP
            SELECT cast(coalesce(nullif(t_row.capability_statement->>'fhirVersion',NULL),'') as varchar(500)) INTO capStatVersion;
            UPDATE fhir_endpoints_info_history SET requested_fhir_version = '', capability_fhir_version = capStatVersion WHERE current of t_curs; 
        END LOOP;
    END;
$$ LANGUAGE plpgsql;

SELECT populate_capability_fhir_version_info_history();

-- captures history for the fhir_endpoint_info table
CREATE TRIGGER add_fhir_endpoint_info_history_trigger
AFTER INSERT OR UPDATE OR DELETE on fhir_endpoints_info
FOR EACH ROW
WHEN (current_setting('metadata.setting', 't') IS NULL OR current_setting('metadata.setting', 't') = 'FALSE')
EXECUTE PROCEDURE add_fhir_endpoint_info_history();

CREATE or REPLACE VIEW endpoint_export AS
SELECT endpts.url, endpts.list_source, endpts.organization_names AS endpoint_names,
    vendors.name as vendor_name,
    endpts_info.tls_version, endpts_info.mime_types, endpts_metadata.http_response,
    endpts_metadata.response_time_seconds, endpts_metadata.smart_http_response, endpts_metadata.errors,
    endpts_info.capability_fhir_version AS FHIR_VERSION,
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
LEFT JOIN fhir_endpoints_metadata AS endpts_metadata ON endpts_info.metadata_id = endpts_metadata.id
LEFT JOIN vendors ON endpts_info.vendor_id = vendors.id
LEFT JOIN npi_organizations AS orgs ON links.organization_npi_id = orgs.npi_id
WHERE endpts_info.requested_fhir_version = '';

CREATE INDEX capability_fhir_version_idx ON fhir_endpoints_info (capability_fhir_version);
CREATE INDEX requested_fhir_version_idx ON fhir_endpoints_info (requested_fhir_version);

COMMIT;