BEGIN;

DROP VIEW IF EXISTS endpoint_export;

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

COMMIT;