BEGIN;

DROP MATERIALIZED VIEW IF EXISTS mv_vendor_fhir_counts CASCADE;
DROP INDEX IF EXISTS idx_mv_vendor_fhir_counts_vendor;
DROP INDEX IF EXISTS idx_mv_vendor_fhir_counts_fhir;
DROP INDEX IF EXISTS idx_mv_vendor_fhir_counts_unique;
DROP INDEX IF EXISTS idx_mv_vendor_fhir_counts_sort;

CREATE MATERIALIZED VIEW mv_vendor_fhir_counts AS
WITH vendor_totals AS (
    -- Calculate total endpoints for each vendor
    SELECT 
        COALESCE(v.name, 'Unknown') AS vendor_name,
        COUNT(DISTINCT e.url) AS total_endpoints
    FROM endpoint_export e
    LEFT JOIN vendors v ON e.vendor_name = v.name
    GROUP BY COALESCE(v.name, 'Unknown')
),
vendor_rank AS (
    -- Rank vendors by total endpoints and determine top 10
    SELECT 
        vendor_name,
        total_endpoints,
        RANK() OVER (ORDER BY total_endpoints DESC) AS rank
    FROM vendor_totals
),
endpoint_counts_base AS (
    -- Get counts by vendor and FHIR version
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
        -- Flag whether this vendor should be part of "Others"
        CASE 
            WHEN r.rank <= 10 THEN false
            WHEN t.total_endpoints < 50 THEN true
            ELSE false
        END AS is_other
    FROM endpoint_export e
    LEFT JOIN vendors v ON e.vendor_name = v.name
    JOIN vendor_totals t ON COALESCE(v.name, 'Unknown') = t.vendor_name
    JOIN vendor_rank r ON t.vendor_name = r.vendor_name
    GROUP BY 
        COALESCE(v.name, 'Unknown'), 
        CASE
            WHEN e.fhir_version IS NULL OR trim(e.fhir_version) = '' THEN 'No Cap Stat'
            WHEN position('-' in e.fhir_version) > 0 THEN substring(e.fhir_version, 1, position('-' in e.fhir_version) - 1)
            WHEN e.fhir_version NOT IN ('0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2', '1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', '3.0.0', '3.0.1', '3.0.2', '3.2.0', '3.3.0', '3.5.0', '3.5a.0', '4.0.0', '4.0.1') THEN 'Unknown'
            ELSE e.fhir_version
        END,
        r.rank,
        t.total_endpoints
),
others_aggregated AS (
    -- Aggregate "others" into one row per FHIR version
    SELECT 
        'Others' AS vendor_name,
        fhir_version,
        SUM(n) AS n,
        true AS is_other
    FROM endpoint_counts_base
    WHERE is_other = true
    GROUP BY fhir_version
),
non_others AS (
    -- Keep non-others as is
    SELECT
        vendor_name,
        fhir_version,
        n,
        is_other
    FROM endpoint_counts_base
    WHERE is_other = false
),
combined AS (
    -- Combine both sets
    SELECT * FROM others_aggregated
    UNION ALL
    SELECT * FROM non_others
)
SELECT 
    c.vendor_name,
    c.fhir_version,
    c.n,
    -- Add a sort order field
    CASE
        WHEN r.rank IS NOT NULL AND r.rank <= 10 THEN r.rank
        WHEN c.vendor_name = 'Others' THEN 9999  -- Others go at the bottom
        ELSE 1000 + COALESCE(r.rank, 0)  -- Keep remaining larger vendors in order
    END AS sort_order
FROM combined c
LEFT JOIN vendor_rank r ON c.vendor_name = r.vendor_name
ORDER BY sort_order, vendor_name, fhir_version;

-- Add indexes to improve query performance
CREATE INDEX idx_mv_vendor_fhir_counts_vendor ON mv_vendor_fhir_counts(vendor_name);
CREATE INDEX idx_mv_vendor_fhir_counts_fhir ON mv_vendor_fhir_counts(fhir_version);
CREATE INDEX idx_mv_vendor_fhir_counts_sort ON mv_vendor_fhir_counts(sort_order);

-- Create a unique index for concurrent refresh
-- Since vendor_name and fhir_version alone aren't guaranteed to be unique anymore (due to "Others"),
-- we include sort_order to ensure uniqueness
CREATE UNIQUE INDEX idx_mv_vendor_fhir_counts_unique ON mv_vendor_fhir_counts(vendor_name, fhir_version, sort_order);

COMMIT;