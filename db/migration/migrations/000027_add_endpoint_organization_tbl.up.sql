BEGIN;

DROP TABLE IF EXISTS fhir_endpoint_organizations;
DROP TABLE IF EXISTS fhir_endpoint_organizations_map;
DROP VIEW IF EXISTS endpoint_export;
DROP VIEW IF EXISTS organization_location;
DROP VIEW IF EXISTS joined_export_tables;
DROP TRIGGER IF EXISTS set_timestamp_fhir_endpoint_organizations;

ALTER TABLE fhir_endpoints DROP COLUMN IF EXISTS organization_names CASCADE; 
ALTER TABLE fhir_endpoints DROP COLUMN IF EXISTS npi_ids CASCADE; 

CREATE TABLE fhir_endpoint_organizations (
    id                      SERIAL PRIMARY KEY,
    organization_name       VARCHAR(500),
    organization_zipcode    VARCHAR(500),
    organization_npi_id    VARCHAR(500),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE fhir_endpoint_organizations_map (
    id INT REFERENCES fhir_endpoints(id) ON DELETE SET NULL,
    org_database_id INT REFERENCES fhir_endpoint_organizations(id) ON DELETE SET NULL
);

CREATE or REPLACE VIEW joined_export_tables AS
SELECT endpts.url, endpts.list_source, endpt_orgnames.organization_names AS endpoint_names,
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
    endpts_info.requested_fhir_version, endpts_metadata.availability
FROM fhir_endpoints AS endpts
LEFT JOIN fhir_endpoints_info AS endpts_info ON endpts.url = endpts_info.url
LEFT JOIN fhir_endpoints_metadata AS endpts_metadata ON endpts_info.metadata_id = endpts_metadata.id
LEFT JOIN vendors ON endpts_info.vendor_id = vendors.id
LEFT JOIN (SELECT fom.id as id, array_agg(fo.organization_name) as organization_names 
FROM fhir_endpoints AS fe, fhir_endpoint_organizations_map AS fom, fhir_endpoint_organizations AS fo
WHERE fe.id = fom.id AND fom.org_database_id = fo.id
GROUP BY fom.id) as endpt_orgnames ON endpts.id = endpt_orgnames.id;

CREATE or REPLACE VIEW endpoint_export AS
SELECT export_tables.url, export_tables.list_source, export_tables.endpoint_names,
    export_tables.vendor_name,
    export_tables.tls_version, export_tables.mime_types, export_tables.http_response,
    export_tables.response_time_seconds, export_tables.smart_http_response, export_tables.errors,
    EXISTS (SELECT 1 FROM fhir_endpoints_info WHERE capability_statement::jsonb != 'null' AND export_tables.url = fhir_endpoints_info.url) as CAP_STAT_EXISTS,
    export_tables.fhir_version,
    export_tables.publisher,
    export_tables.software_name,
    export_tables.software_version,
    export_tables.software_releasedate,
    export_tables.format,
    export_tables.kind,
    export_tables.info_updated,
    export_tables.info_created,
    export_tables.requested_fhir_version,
    export_tables.availability,
    list_source_info.is_chpl
FROM joined_export_tables AS export_tables
LEFT JOIN list_source_info ON export_tables.list_source = list_source_info.list_source;

CREATE or REPLACE VIEW organization_location AS
    SELECT export_tables.url, export_tables.endpoint_names, export_tables.fhir_version,
    export_tables.requested_fhir_version, export_tables.vendor_name, orgs.name AS ORGANIZATION_NAME, orgs.secondary_name AS ORGANIZATION_SECONDARY_NAME,
    orgs.taxonomy, orgs.Location->>'state' AS STATE, orgs.Location->>'zipcode' AS ZIPCODE, orgs.npi_id as NPI_ID,
    links.confidence AS MATCH_SCORE
FROM endpoint_organization AS links
RIGHT JOIN joined_export_tables AS export_tables ON links.url = export_tables.url
LEFT JOIN npi_organizations AS orgs ON links.organization_npi_id = orgs.npi_id
WHERE links.confidence > .97 AND orgs.Location->>'zipcode' IS NOT null;

CREATE TRIGGER set_timestamp_fhir_endpoint_organizations
BEFORE UPDATE ON fhir_endpoint_organizations
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

CREATE INDEX IF NOT EXISTS fhir_endpoint_organizations_id ON fhir_endpoint_organizations (id);
CREATE INDEX IF NOT EXISTS fhir_endpoint_organizations_name ON fhir_endpoint_organizations (organization_name);
CREATE INDEX IF NOT EXISTS fhir_endpoint_organizations_zipcode ON fhir_endpoint_organizations (organization_zipcode);
CREATE INDEX IF NOT EXISTS fhir_endpoint_organizations_npi_id ON fhir_endpoint_organizations (organization_npi_id);

COMMIT;