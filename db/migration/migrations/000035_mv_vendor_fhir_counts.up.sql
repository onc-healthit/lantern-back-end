BEGIN;

DROP MATERIALIZED VIEW IF EXISTS mv_vendor_fhir_counts CASCADE;

CREATE MATERIALIZED VIEW mv_vendor_fhir_counts AS
SELECT 
    COALESCE(v.name, 'Unknown') AS vendor_name,
    CASE
        WHEN e.fhir_version IS NULL OR trim(e.fhir_version) = '' THEN 'No Cap Stat'
        -- Apply the dash rule: if there's a dash, trim after it
        WHEN position('-' in e.fhir_version) > 0 THEN substring(e.fhir_version, 1, position('-' in e.fhir_version) - 1)
        -- If it's not in the valid list, mark as Unknown
        WHEN e.fhir_version NOT IN ('0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2', '1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', '3.0.0', '3.0.1', '3.0.2', '3.2.0', '3.3.0', '3.5.0', '3.5a.0', '4.0.0', '4.0.1') THEN 'Unknown'
        ELSE e.fhir_version
    END AS fhir_version,
    COUNT(DISTINCT e.url) AS n,
    CASE
        WHEN COALESCE(v.name, 'Unknown') = 'Allscripts' THEN 'Allscripts'
        WHEN COALESCE(v.name, 'Unknown') = 'CareEvolution, Inc.' THEN 'CareEvolution'
        WHEN COALESCE(v.name, 'Unknown') = 'Cerner Corporation' THEN 'Cerner'
        WHEN COALESCE(v.name, 'Unknown') = 'Epic Systems Corporation' THEN 'Epic'
        WHEN COALESCE(v.name, 'Unknown') = 'Medical Information Technology, Inc. (MEDITECH)' THEN 'MEDITECH'
        WHEN COALESCE(v.name, 'Unknown') = 'Microsoft Corporation' THEN 'Microsoft'
        WHEN COALESCE(v.name, 'Unknown') = 'Unknown' THEN 'Unknown'
        ELSE COALESCE(v.name, 'Unknown')
    END AS short_name
FROM endpoint_export e
LEFT JOIN vendors v ON e.vendor_name = v.name
GROUP BY 
    COALESCE(v.name, 'Unknown'), 
    CASE
        WHEN e.fhir_version IS NULL OR trim(e.fhir_version) = '' THEN 'No Cap Stat'
        WHEN position('-' in e.fhir_version) > 0 THEN substring(e.fhir_version, 1, position('-' in e.fhir_version) - 1)
        WHEN e.fhir_version NOT IN ('0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2', '1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', '3.0.0', '3.0.1', '3.0.2', '3.2.0', '3.3.0', '3.5.0', '3.5a.0', '4.0.0', '4.0.1') THEN 'Unknown'
        ELSE e.fhir_version
    END,
    CASE
        WHEN COALESCE(v.name, 'Unknown') = 'Allscripts' THEN 'Allscripts'
        WHEN COALESCE(v.name, 'Unknown') = 'CareEvolution, Inc.' THEN 'CareEvolution'
        WHEN COALESCE(v.name, 'Unknown') = 'Cerner Corporation' THEN 'Cerner'
        WHEN COALESCE(v.name, 'Unknown') = 'Epic Systems Corporation' THEN 'Epic'
        WHEN COALESCE(v.name, 'Unknown') = 'Medical Information Technology, Inc. (MEDITECH)' THEN 'MEDITECH'
        WHEN COALESCE(v.name, 'Unknown') = 'Microsoft Corporation' THEN 'Microsoft'
        WHEN COALESCE(v.name, 'Unknown') = 'Unknown' THEN 'Unknown'
        ELSE COALESCE(v.name, 'Unknown')
    END
ORDER BY 
    vendor_name, fhir_version;

-- Add indexes to improve query performance
CREATE INDEX idx_mv_vendor_fhir_counts_vendor ON mv_vendor_fhir_counts(vendor_name);
CREATE INDEX idx_mv_vendor_fhir_counts_fhir ON mv_vendor_fhir_counts(fhir_version);
DROP INDEX IF EXISTS idx_mv_vendor_fhir_counts_unique;
CREATE UNIQUE INDEX idx_mv_vendor_fhir_counts_unique ON mv_vendor_fhir_counts(vendor_name, fhir_version);
COMMIT;