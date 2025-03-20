BEGIN;

DROP MATERIALIZED VIEW IF EXISTS fhir_endpoint_comb_mv CASCADE;

CREATE MATERIALIZED VIEW fhir_endpoint_comb_mv AS
WITH enriched_endpoints AS (
    SELECT 
        e.*,
        COALESCE(r.code_label, 'Other') AS code_label, 
        CASE 
            WHEN e.http_response = 200 THEN CONCAT('Success: ', e.http_response, ' - ', r.code_label)
            ELSE CONCAT('Failure: ', e.http_response, ' - ', r.code_label)
        END AS status,
        LOWER(CASE 
            WHEN e.kind != 'instance' THEN 'true*'::TEXT  
            ELSE e.cap_stat_exists::TEXT
        END) AS cap_stat_exists_transformed
    FROM endpoint_export_mv e
    LEFT JOIN mv_http_responses r 
        ON e.http_response = r.http_code
)
SELECT 
    ROW_NUMBER() OVER () AS id,
    e.url,
    e.endpoint_names,
    e.info_created,
    e.info_updated,
    e.list_source,
    e.vendor_name,
    e.capability_fhir_version,
    e.fhir_version, 
    e.format,
    e.http_response,
    e.response_time_seconds,
    e.smart_http_response,
    e.errors,
    e.availability,
    e.kind,
    e.requested_fhir_version,
    lsi.is_chpl,
    e.status,
    e.cap_stat_exists_transformed AS cap_stat_exists 
FROM enriched_endpoints e
LEFT JOIN list_source_info lsi 
    ON e.list_source = lsi.list_source;

--Unique index for refreshing the MV concurrently
CREATE UNIQUE INDEX fhir_endpoint_comb_mv_unique_idx ON fhir_endpoint_comb_mv (id, url, list_source);

COMMIT;