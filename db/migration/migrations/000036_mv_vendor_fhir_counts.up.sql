BEGIN;

DROP MATERIALIZED VIEW IF EXISTS mv_vendor_fhir_counts;

CREATE MATERIALIZED VIEW mv_vendor_fhir_counts AS
WITH developer_counts AS (
    SELECT 
        v_1.name AS vendor_name,
        sum(count(e_1.url)) OVER (PARTITION BY v_1.name) AS developer_count
    FROM endpoint_export e_1
    LEFT JOIN vendors v_1 
        ON e_1.vendor_name::text = v_1.name::text
    GROUP BY v_1.name
)
SELECT 
    COALESCE(v.name, 'Unknown'::character varying) AS vendor_name,
    COALESCE(NULLIF(btrim(e.fhir_version::text), ''::text), 'Unknown'::text) AS fhir_version,
    count(e.url)::integer AS n,
    COALESCE(
        CASE
            WHEN v.name::text = 'Allscripts' THEN 'Allscripts'
            WHEN v.name::text = 'CareEvolution, Inc.' THEN 'CareEvolution'
            WHEN v.name::text = 'Cerner Corporation' THEN 'Cerner'
            WHEN v.name::text = 'Epic Systems Corporation' THEN 'Epic'
            WHEN v.name::text = 'Medical Information Technology, Inc. (MEDITECH)' THEN 'MEDITECH'
            WHEN v.name::text = 'Microsoft Corporation' THEN 'Microsoft'
            WHEN v.name::text = 'NA' THEN 'Unknown'
            ELSE v.name
        END, 'Unknown'::character varying
    ) AS short_name,
    COALESCE(dc.developer_count, 0::numeric) AS developer_count,
    COALESCE(
        concat(
            round(
                COALESCE(count(e.url)::numeric / NULLIF(dc.developer_count, 0::numeric) * 100::numeric, 0::numeric),
                0
            ), '%'
        ), '0%'::text
    ) AS percentage
FROM endpoint_export e
LEFT JOIN vendors v 
    ON e.vendor_name::text = v.name::text
LEFT JOIN developer_counts dc 
    ON v.name::text = dc.vendor_name::text
GROUP BY v.name, e.fhir_version, dc.developer_count
ORDER BY 
    COALESCE(v.name, 'Unknown'::character varying), 
    COALESCE(e.fhir_version, 'Unknown'::character varying);

COMMIT;