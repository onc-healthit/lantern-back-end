BEGIN;

-- Lantern-852
DROP MATERIALIZED VIEW IF EXISTS mv_capstat_sizes_tbl CASCADE;

CREATE MATERIALIZED VIEW mv_capstat_sizes_tbl AS
SELECT
f.url,
pg_column_size(capability_statement::text) AS size,
CASE
    WHEN REGEXP_REPLACE(
            CASE 
            WHEN f.capability_fhir_version = '' THEN 'No Cap Stat'
            ELSE f.capability_fhir_version
            END,
            '-.*', ''
        ) IN (
            'No Cap Stat', '0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2',
            '1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', '3.0.0',
            '3.0.1', '3.0.2', '3.2.0', '3.3.0', '3.5.0', '3.5a.0',
            '4.0.0', '4.0.1'
        )
    THEN REGEXP_REPLACE(
            CASE 
            WHEN f.capability_fhir_version = '' THEN 'No Cap Stat'
            ELSE f.capability_fhir_version
            END,
            '-.*', ''
        )
    ELSE 'Unknown'
END AS fhir_version,
COALESCE(vendors.name, 'Unknown') AS vendor_name
FROM fhir_endpoints_info f
LEFT JOIN vendors ON f.vendor_id = vendors.id
WHERE f.capability_fhir_version != ''
AND f.requested_fhir_version = 'None';

-- Create indexes for mv_capstat_sizes
CREATE UNIQUE INDEX idx_mv_capstat_sizes_uniq ON mv_capstat_sizes_tbl(url);
CREATE INDEX idx_mv_capstat_sizes_fhir ON mv_capstat_sizes_tbl(fhir_version);
CREATE INDEX idx_mv_capstat_sizes_vendor ON mv_capstat_sizes_tbl(vendor_name);

COMMIT;