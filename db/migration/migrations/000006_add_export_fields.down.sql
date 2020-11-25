BEGIN;

DROP VIEW IF EXISTS endpoint_export;

CREATE or REPLACE VIEW endpoint_export AS
SELECT endpts.url, endpts.list_source, vendors.name as vendor_name, endpts.organization_names AS endpoint_names, endpts_info.tls_version, endpts_info.mime_types, endpts_info.http_response, endpts_info.capability_statement->>'fhirVersion' AS FHIR_VERSION, endpts_info.capability_statement->>'publisher' AS PUBLISHER, endpts_info.capability_statement->'software'->'name' AS SOFTWARE_NAME, endpts_info.capability_statement->'software'->'version' AS SOFTWARE_VERSION, endpts_info.capability_statement->'software'->'releaseDate' AS SOFTWARE_RELEASEDATE, orgs.name AS ORGANIZATION_NAME, orgs.secondary_name AS ORGANIZATION_SECONDARY_NAME, orgs.taxonomy, orgs.Location->>'state' AS STATE, orgs.Location->>'zipcode' AS ZIPCODE, links.confidence AS MATCH_SCORE, endpts_info.supported_resources
FROM endpoint_organization AS links
RIGHT JOIN fhir_endpoints AS endpts ON links.url = endpts.url
LEFT JOIN fhir_endpoints_info AS endpts_info ON endpts.url = endpts_info.url
LEFT JOIN vendors ON endpts_info.vendor_id = vendors.id
LEFT JOIN npi_organizations AS orgs ON links.organization_npi_id = orgs.npi_id;

COMMIT;