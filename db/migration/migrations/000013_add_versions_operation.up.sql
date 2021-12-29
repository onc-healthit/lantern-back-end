BEGIN;

DROP TRIGGER IF EXISTS add_fhir_endpoint_info_history_trigger ON fhir_endpoints_info;
DROP VIEW IF EXISTS endpoint_export;
DROP VIEW IF EXISTS org_mapping;

DROP INDEX IF EXISTS fhir_version_idx;

ALTER TABLE fhir_endpoints ADD COLUMN IF NOT EXISTS versions_response JSONB;

ALTER TABLE fhir_endpoints_info 
ADD COLUMN IF NOT EXISTS requested_fhir_version VARCHAR(500);

ALTER TABLE fhir_endpoints_info_history 
ADD COLUMN IF NOT EXISTS requested_fhir_version VARCHAR(500);

ALTER TABLE fhir_endpoints_info 
ADD COLUMN IF NOT EXISTS capability_fhir_version VARCHAR(500);

ALTER TABLE fhir_endpoints_info_history 
ADD COLUMN IF NOT EXISTS capability_fhir_version VARCHAR(500);

ALTER TABLE fhir_endpoints_info DROP CONSTRAINT IF EXISTS fhir_endpoints_info_url_key;
ALTER TABLE fhir_endpoints_info DROP CONSTRAINT IF EXISTS fhir_endpoints_info_unique;
ALTER TABLE fhir_endpoints_info ADD CONSTRAINT fhir_endpoints_info_unique UNIQUE(url, requested_fhir_version);

ALTER TABLE fhir_endpoints_metadata ADD COLUMN IF NOT EXISTS requested_fhir_version VARCHAR(500) DEFAULT 'None';
ALTER TABLE fhir_endpoints_availability ADD COLUMN IF NOT EXISTS requested_fhir_version VARCHAR(500) DEFAULT 'None';

CREATE OR REPLACE FUNCTION populate_capability_fhir_version_info() RETURNS VOID as $$
    DECLARE
        t_curs cursor for select capability_statement from fhir_endpoints_info;
        t_row fhir_endpoints_info%ROWTYPE;
        capStatVersion VARCHAR(500);
    BEGIN
        FOR t_row in t_curs LOOP
            SELECT cast(coalesce(nullif(t_row.capability_statement->>'fhirVersion',NULL),'') as varchar(500)) INTO capStatVersion;
            UPDATE fhir_endpoints_info SET requested_fhir_version = 'None', capability_fhir_version = capStatVersion WHERE current of t_curs;
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
            UPDATE fhir_endpoints_info_history SET requested_fhir_version = 'None', capability_fhir_version = capStatVersion WHERE current of t_curs;
        END LOOP;
    END;
$$ LANGUAGE plpgsql;

SELECT populate_capability_fhir_version_info_history();

CREATE OR REPLACE FUNCTION update_fhir_endpoint_availability_info() RETURNS TRIGGER AS $fhir_endpoints_availability$
    DECLARE
        okay_count       bigint;
        all_count        bigint;
    BEGIN
        --
        -- Create or update a row in fhir_endpoint_availabilty with new total http and 200 http count 
        -- when an endpoint is inserted or updated in fhir_endpoint_info. Also calculate new 
        -- endpoint availability precentage
        SELECT http_200_count, http_all_count INTO okay_count, all_count FROM fhir_endpoints_availability WHERE url = NEW.url AND requested_fhir_version = NEW.requested_fhir_version;
        IF  NOT FOUND THEN
            IF NEW.http_response = 200 THEN
                INSERT INTO fhir_endpoints_availability(url, http_200_count, http_all_count, requested_fhir_version) VALUES (NEW.url, 1, 1, NEW.requested_fhir_version);
                NEW.availability = 1.00;
                RETURN NEW;
            ELSE
                INSERT INTO fhir_endpoints_availability(url, http_200_count, http_all_count, requested_fhir_version) VALUES (NEW.url, 0, 1, NEW.requested_fhir_version);
                NEW.availability = 0.00;
                RETURN NEW;
            END IF;
        ELSE
            IF NEW.http_response = 200 THEN
                UPDATE fhir_endpoints_availability SET http_200_count = okay_count + 1.0, http_all_count = all_count + 1.0 WHERE url = NEW.url AND requested_fhir_version = NEW.requested_fhir_version;
                NEW.availability := (okay_count + 1.0) / (all_count + 1.0);
                RETURN NEW;
            ELSE
                UPDATE fhir_endpoints_availability SET http_all_count = all_count + 1.0 WHERE url = NEW.url AND requested_fhir_version = NEW.requested_fhir_version;
                NEW.availability := (okay_count) / (all_count + 1.0);
                RETURN NEW;
            END IF;
        END IF;
    END;
$fhir_endpoints_availability$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS add_fhir_endpoint_info_history_trigger ON fhir_endpoints_info;
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
    endpts_info.requested_fhir_version,
    orgs.name AS ORGANIZATION_NAME, orgs.secondary_name AS ORGANIZATION_SECONDARY_NAME,
    orgs.taxonomy, orgs.Location->>'state' AS STATE, orgs.Location->>'zipcode' AS ZIPCODE,
    links.confidence AS MATCH_SCORE, endpts_metadata.availability
FROM endpoint_organization AS links
RIGHT JOIN fhir_endpoints AS endpts ON links.url = endpts.url
LEFT JOIN fhir_endpoints_info AS endpts_info ON endpts.url = endpts_info.url
LEFT JOIN fhir_endpoints_metadata AS endpts_metadata ON endpts_info.metadata_id = endpts_metadata.id
LEFT JOIN vendors ON endpts_info.vendor_id = vendors.id
LEFT JOIN npi_organizations AS orgs ON links.organization_npi_id = orgs.npi_id;

CREATE INDEX IF NOT EXISTS capability_fhir_version_idx ON fhir_endpoints_info (capability_fhir_version);
CREATE INDEX IF NOT EXISTS requested_fhir_version_idx ON fhir_endpoints_info (requested_fhir_version);

COMMIT;