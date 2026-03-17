BEGIN;

-- ========================================
-- Migration 000079: Add Data Issues Materialized Views
-- ========================================
-- Purpose: Create materialized views to track data quality issues for developers
-- Tracks:
--   1. Developers with endpoints that have no organization data
--   2. Endpoints sharing list_source URLs
--   3. List_source URL accessibility issues
-- ========================================

-- Drop existing views if they exist (order matters due to dependencies)
DROP MATERIALIZED VIEW IF EXISTS mv_chpl_coverage_summary CASCADE;
DROP MATERIALIZED VIEW IF EXISTS mv_developer_bundle_issues CASCADE;
DROP MATERIALIZED VIEW IF EXISTS mv_developer_data_issues CASCADE;
DROP MATERIALIZED VIEW IF EXISTS mv_latest_endpoint_metadata CASCADE;

-- ========================================
-- MATERIALIZED VIEW: mv_latest_endpoint_metadata
-- ========================================
-- Purpose: Pre-compute the most recent metadata check result per endpoint.
-- This is built as a standalone MV (not an inline CTE) so that
-- mv_developer_data_issues can reference it
-- without re-scanning fhir_endpoints_metadata twice, reducing build time.
-- ========================================

CREATE MATERIALIZED VIEW mv_latest_endpoint_metadata AS
SELECT DISTINCT ON (url, requested_fhir_version)
    url,
    requested_fhir_version,
    http_response
FROM fhir_endpoints_metadata
ORDER BY url, requested_fhir_version, updated_at DESC;

CREATE UNIQUE INDEX idx_mv_latest_endpoint_metadata_url_version
    ON mv_latest_endpoint_metadata(url, requested_fhir_version);

-- ========================================
-- MATERIALIZED VIEW: mv_developer_data_issues
-- ========================================
-- Purpose: Detailed developer-level data issues tracking
-- ========================================

CREATE MATERIALIZED VIEW mv_developer_data_issues AS
WITH
-- Get all unique vendors from fhir_endpoints_info AND shared_list_sources
-- Includes: (1) all FHIR endpoint vendors, (2) CHPL developers with empty bundles,
-- (3) ALL developers sharing a list_source (catches vendors not in fhir_endpoints_info)
all_vendors AS (
    SELECT DISTINCT COALESCE(v.name, 'Unknown') as vendor_name
    FROM fhir_endpoints_info fei
    LEFT JOIN vendors v ON fei.vendor_id = v.id
    WHERE fei.requested_fhir_version = 'None'

    UNION

    -- Include developers with empty bundles (from shared_list_sources table)
    -- NOTE: This only includes CHPL developers from the CHPL CSV
    SELECT DISTINCT sls.developer_name as vendor_name
    FROM shared_list_sources sls
    LEFT JOIN fhir_endpoints fe ON REPLACE(REPLACE(sls.list_source, 'u0026', '&'), '%26', '&') = REPLACE(REPLACE(fe.list_source, 'u0026', '&'), '%26', '&')
    GROUP BY sls.developer_name, sls.list_source
    HAVING COUNT(fe.url) = 0

    UNION

    -- Include ALL developers sharing a list_source with at least one other developer
    -- This catches developers (e.g. CareCloud Inc., Meridian) who appear in shared_list_sources
    -- but have no endpoints in fhir_endpoints_info
    SELECT DISTINCT sls.developer_name as vendor_name
    FROM shared_list_sources sls
    WHERE sls.list_source IN (
        SELECT list_source
        FROM shared_list_sources
        GROUP BY list_source
        HAVING COUNT(DISTINCT developer_name) > 1
    )

    UNION

    -- Include ALL CHPL developers by developer_name (catches name-mismatch developers
    -- who have endpoints in fhir_endpoints_info but their vendor name differs from CHPL name)
    SELECT DISTINCT developer_name as vendor_name
    FROM shared_list_sources
),
-- Total endpoints per vendor
vendor_endpoints AS (
    SELECT
        COALESCE(v.name, 'Unknown') as vendor_name,
        COUNT(DISTINCT fei.url) as total_endpoints
    FROM fhir_endpoints_info fei
    LEFT JOIN vendors v ON fei.vendor_id = v.id
    WHERE fei.requested_fhir_version = 'None'
    GROUP BY v.name
),
-- Endpoints with organization data
vendor_endpoints_with_data AS (
    SELECT
        COALESCE(v.name, 'Unknown') as vendor_name,
        COUNT(DISTINCT sfem.url) as endpoints_with_data
    FROM selected_fhir_endpoints_mv sfem
    LEFT JOIN fhir_endpoints_info fei ON sfem.url = fei.url AND fei.requested_fhir_version = 'None'
    LEFT JOIN vendors v ON fei.vendor_id = v.id
    WHERE
        sfem.endpoint_names IS NOT NULL
        AND sfem.endpoint_names != ''
        AND TRIM(sfem.endpoint_names) != ''
        AND sfem.requested_fhir_version = 'None'
    GROUP BY v.name
),
-- Endpoints with no organization data
vendor_no_org_data AS (
    SELECT
        COALESCE(v.name, 'Unknown') as vendor_name,
        COUNT(DISTINCT sfem.url) as no_org_data_endpoints
    FROM selected_fhir_endpoints_mv sfem
    LEFT JOIN fhir_endpoints_info fei ON sfem.url = fei.url AND fei.requested_fhir_version = 'None'
    LEFT JOIN vendors v ON fei.vendor_id = v.id
    WHERE
        (sfem.endpoint_names IS NULL OR sfem.endpoint_names = '' OR TRIM(sfem.endpoint_names) = '')
        AND sfem.requested_fhir_version = 'None'
    GROUP BY v.name
),
-- Accessible endpoints per vendor
-- Uses mv_latest_endpoint_metadata to get current state per endpoint (avoids double-counting)
vendor_accessible_endpoints AS (
    SELECT
        COALESCE(v.name, 'Unknown') as vendor_name,
        COUNT(DISTINCT lm.url) as accessible_endpoints
    FROM mv_latest_endpoint_metadata lm
    INNER JOIN fhir_endpoints_info fei ON lm.url = fei.url AND lm.requested_fhir_version = fei.requested_fhir_version
    LEFT JOIN vendors v ON fei.vendor_id = v.id
    WHERE
        lm.http_response = 200
        AND lm.requested_fhir_version = 'None'
    GROUP BY v.name
),
-- Inaccessible endpoints per vendor
-- Uses mv_latest_endpoint_metadata; counts http_response >= 400 AND http_response = 0
-- http_response = 0 means connection failed (no HTTP response received from server)
vendor_inaccessible_endpoints AS (
    SELECT
        COALESCE(v.name, 'Unknown') as vendor_name,
        COUNT(DISTINCT lm.url) as inaccessible_endpoints
    FROM mv_latest_endpoint_metadata lm
    INNER JOIN fhir_endpoints_info fei ON lm.url = fei.url AND lm.requested_fhir_version = fei.requested_fhir_version
    LEFT JOIN vendors v ON fei.vendor_id = v.id
    WHERE
        lm.http_response IS NOT NULL
        AND (lm.http_response >= 400 OR lm.http_response = 0)
        AND lm.requested_fhir_version = 'None'
    GROUP BY v.name
),
-- Organization count per vendor from fhir_endpoint_organizations
vendor_organizations AS (
    SELECT
        COALESCE(v.name, 'Unknown') as vendor_name,
        COUNT(DISTINCT feo.organization_name) as organization_count
    FROM fhir_endpoint_organizations feo
    INNER JOIN fhir_endpoint_organizations_map feom ON feo.id = feom.org_database_id
    INNER JOIN fhir_endpoints fe ON feom.id = fe.id
    INNER JOIN fhir_endpoints_info fei ON fe.url = fei.url
    LEFT JOIN vendors v ON fei.vendor_id = v.id
    WHERE
        feo.organization_name IS NOT NULL
        AND feo.organization_name != ''
        AND fei.requested_fhir_version = 'None'
    GROUP BY v.name
),
-- Developers with empty bundles (list_sources returning no endpoints)
-- Uses shared_list_sources table
-- NOTE: This is CHPL-only because shared_list_sources only contains CHPL developers from CSV
developers_empty_bundles AS (
    SELECT DISTINCT
        sls.developer_name as vendor_name
    FROM shared_list_sources sls
    LEFT JOIN fhir_endpoints fe ON REPLACE(REPLACE(sls.list_source, 'u0026', '&'), '%26', '&') = REPLACE(REPLACE(fe.list_source, 'u0026', '&'), '%26', '&')
    GROUP BY sls.developer_name, sls.list_source
    HAVING COUNT(fe.url) = 0
),
-- Developers/vendors sharing list sources (from shared_list_sources table)
vendors_sharing_list_sources AS (
    SELECT DISTINCT
        sls.developer_name as vendor_name
    FROM shared_list_sources sls
    WHERE sls.list_source IN (
        SELECT list_source
        FROM shared_list_sources
        GROUP BY list_source
        HAVING COUNT(DISTINCT developer_name) > 1
    )
),
-- CHPL developers: any developer with at least one entry in shared_list_sources
chpl_developers AS (
    SELECT DISTINCT developer_name as vendor_name
    FROM shared_list_sources
),
-- Developers whose set of FHIR endpoint URLs exactly matches another developer's set
-- Logic: two developers share FHIR endpoints if their endpoint URL sets are identical
developers_sharing_fhir_endpoints AS (
    WITH dev_endpoint_sets AS (
        SELECT
            sls.developer_name,
            ARRAY_AGG(DISTINCT fe.url ORDER BY fe.url) AS endpoint_set
        FROM shared_list_sources sls
        INNER JOIN fhir_endpoints fe ON sls.list_source = fe.list_source
        GROUP BY sls.developer_name
    )
    SELECT DISTINCT d1.developer_name AS vendor_name
    FROM dev_endpoint_sets d1
    JOIN dev_endpoint_sets d2
        ON d1.developer_name != d2.developer_name
        AND d1.endpoint_set = d2.endpoint_set
)
SELECT
    av.vendor_name,
    COALESCE(ve.total_endpoints, 0) as total_endpoints,
    COALESCE(vewd.endpoints_with_data, 0) as endpoints_with_org_data,
    COALESCE(vnod.no_org_data_endpoints, 0) as no_org_data_endpoints,
    COALESCE(vae.accessible_endpoints, 0) as accessible_endpoints,
    COALESCE(vie.inaccessible_endpoints, 0) as inaccessible_endpoints,
    COALESCE(vo.organization_count, 0) as organization_count,
    CASE
        WHEN COALESCE(ve.total_endpoints, 0) = 0 THEN 0
        ELSE ROUND((COALESCE(vewd.endpoints_with_data, 0)::numeric / ve.total_endpoints::numeric) * 100, 1)
    END as data_completeness_percentage,
    CASE
        WHEN deb.vendor_name IS NOT NULL THEN TRUE
        ELSE FALSE
    END as has_empty_bundle,
    CASE
        WHEN vsls.vendor_name IS NOT NULL THEN TRUE
        ELSE FALSE
    END as shares_list_source,
    CASE
        WHEN cd.vendor_name IS NOT NULL THEN TRUE
        ELSE FALSE
    END as is_chpl_developer,
    CASE
        WHEN dsfe.vendor_name IS NOT NULL THEN TRUE
        ELSE FALSE
    END as shares_fhir_endpoints
FROM all_vendors av
LEFT JOIN vendor_endpoints ve ON av.vendor_name = ve.vendor_name
LEFT JOIN vendor_endpoints_with_data vewd ON av.vendor_name = vewd.vendor_name
LEFT JOIN vendor_no_org_data vnod ON av.vendor_name = vnod.vendor_name
LEFT JOIN vendor_accessible_endpoints vae ON av.vendor_name = vae.vendor_name
LEFT JOIN vendor_inaccessible_endpoints vie ON av.vendor_name = vie.vendor_name
LEFT JOIN vendor_organizations vo ON av.vendor_name = vo.vendor_name
LEFT JOIN developers_empty_bundles deb ON av.vendor_name = deb.vendor_name
LEFT JOIN vendors_sharing_list_sources vsls ON av.vendor_name = vsls.vendor_name
LEFT JOIN chpl_developers cd ON av.vendor_name = cd.vendor_name
LEFT JOIN developers_sharing_fhir_endpoints dsfe ON av.vendor_name = dsfe.vendor_name
-- Removed: WHERE COALESCE(ve.total_endpoints, 0) > 0
-- This was filtering out developers with empty bundles who have 0 endpoints
ORDER BY
    CASE
        WHEN COALESCE(vnod.no_org_data_endpoints, 0) = COALESCE(ve.total_endpoints, 0)
             AND COALESCE(ve.total_endpoints, 0) > 0 THEN 1  -- Critical: All endpoints have no org data
        WHEN COALESCE(vnod.no_org_data_endpoints, 0) > 0 THEN 2  -- Warning: Some endpoints have no org data
        ELSE 3  -- OK: All endpoints have org data
    END,
    av.vendor_name;

-- Create indexes for faster queries
CREATE INDEX idx_mv_developer_data_issues_vendor ON mv_developer_data_issues(vendor_name);
CREATE INDEX idx_mv_developer_data_issues_no_org_data ON mv_developer_data_issues(no_org_data_endpoints);

-- ========================================
-- MATERIALIZED VIEW: mv_developer_bundle_issues
-- ========================================
-- Purpose: One row per (developer_name, list_source) pair from shared_list_sources.
-- All endpoint/org counts are computed per bundle URL, so a developer with multiple
-- list_sources appears as multiple rows with accurate per-URL counts.
-- This replaces the ambiguous per-developer has_empty_bundle boolean in
-- mv_developer_data_issues, which fired TRUE if ANY one URL had 0 endpoints.
-- shares_fhir_endpoints is inherently developer-level and is joined from
-- mv_developer_data_issues — it correctly repeats across all bundle rows for
-- the same developer.
-- ========================================

DROP MATERIALIZED VIEW IF EXISTS mv_developer_bundle_issues;

CREATE MATERIALIZED VIEW mv_developer_bundle_issues AS
WITH
-- Total endpoints per bundle URL
bundle_total_endpoints AS (
    SELECT
        fe.list_source,
        COUNT(DISTINCT fe.url) AS total_endpoints
    FROM fhir_endpoints fe
    GROUP BY fe.list_source
),
-- Endpoints with org data per bundle URL (endpoint_names populated)
bundle_endpoints_with_data AS (
    SELECT
        sfem.list_source,
        COUNT(DISTINCT sfem.url) AS endpoints_with_org_data
    FROM selected_fhir_endpoints_mv sfem
    WHERE
        sfem.endpoint_names IS NOT NULL
        AND sfem.endpoint_names != ''
        AND TRIM(sfem.endpoint_names) != ''
        AND sfem.requested_fhir_version = 'None'
    GROUP BY sfem.list_source
),
-- Endpoints with NO org data per bundle URL
bundle_no_org_data AS (
    SELECT
        sfem.list_source,
        COUNT(DISTINCT sfem.url) AS no_org_data_endpoints
    FROM selected_fhir_endpoints_mv sfem
    WHERE
        (sfem.endpoint_names IS NULL OR sfem.endpoint_names = '' OR TRIM(sfem.endpoint_names) = '')
        AND sfem.requested_fhir_version = 'None'
    GROUP BY sfem.list_source
),
-- Organization count per bundle URL
bundle_organizations AS (
    SELECT
        fe.list_source,
        COUNT(DISTINCT feo.organization_name) AS organization_count
    FROM fhir_endpoint_organizations feo
    INNER JOIN fhir_endpoint_organizations_map feom ON feo.id = feom.org_database_id
    INNER JOIN fhir_endpoints fe ON feom.id = fe.id
    INNER JOIN fhir_endpoints_info fei ON fe.url = fei.url
    WHERE
        feo.organization_name IS NOT NULL
        AND feo.organization_name != ''
        AND fei.requested_fhir_version = 'None'
    GROUP BY fe.list_source
),
-- Bundle URLs shared by more than one developer
shared_urls AS (
    SELECT list_source
    FROM shared_list_sources
    GROUP BY list_source
    HAVING COUNT(DISTINCT developer_name) > 1
),
-- Developer-level shares_fhir_endpoints flag (inherently developer-level)
dev_shares_fhir AS (
    SELECT vendor_name, shares_fhir_endpoints
    FROM mv_developer_data_issues
)
SELECT
    sls.developer_name,
    sls.list_source,
    COALESCE(bte.total_endpoints, 0)           AS total_endpoints,
    COALESCE(bewd.endpoints_with_org_data, 0)  AS endpoints_with_org_data,
    COALESCE(bnod.no_org_data_endpoints, 0)    AS no_org_data_endpoints,
    COALESCE(bo.organization_count, 0)          AS organization_count,
    CASE WHEN COALESCE(bte.total_endpoints, 0) = 0
         THEN TRUE ELSE FALSE END               AS has_empty_bundle,
    CASE WHEN su.list_source IS NOT NULL
         THEN TRUE ELSE FALSE END               AS shares_list_source,
    COALESCE(dsf.shares_fhir_endpoints, FALSE)  AS shares_fhir_endpoints,
    TRUE                                        AS is_chpl_developer
FROM shared_list_sources sls
LEFT JOIN bundle_total_endpoints    bte  ON sls.list_source = bte.list_source
LEFT JOIN bundle_endpoints_with_data bewd ON sls.list_source = bewd.list_source
LEFT JOIN bundle_no_org_data         bnod ON sls.list_source = bnod.list_source
LEFT JOIN bundle_organizations       bo   ON sls.list_source = bo.list_source
LEFT JOIN shared_urls                su   ON sls.list_source = su.list_source
LEFT JOIN dev_shares_fhir            dsf  ON sls.developer_name = dsf.vendor_name
ORDER BY sls.developer_name, sls.list_source;

CREATE UNIQUE INDEX idx_mv_developer_bundle_issues_unique
    ON mv_developer_bundle_issues(developer_name, list_source);
CREATE INDEX idx_mv_developer_bundle_issues_developer
    ON mv_developer_bundle_issues(developer_name);
CREATE INDEX idx_mv_developer_bundle_issues_list_source
    ON mv_developer_bundle_issues(list_source);

-- ========================================
-- MATERIALIZED VIEW: mv_chpl_coverage_summary
-- ========================================
-- Purpose: Pre-compute CHPL coverage counts and last-fetch timestamp for the
--          Developer Feedback tab Coverage Overview card and "CHPL data last fetched" label.
-- Single-row MV — refreshed on the same schedule as other developer MVs.
-- ========================================
CREATE MATERIALIZED VIEW mv_chpl_coverage_summary AS
SELECT
    MAX(sls.updated_at)                                                                   AS last_updated,
    COUNT(DISTINCT sls.developer_name)                                                    AS chpl_dev_count,
    COUNT(DISTINCT sls.list_source)                                                       AS chpl_bundle_count,
    COUNT(DISTINCT CASE WHEN lsi.is_chpl = 'CHPL' THEN lsi.list_source END)              AS lantern_chpl_bundle_count,
    COUNT(DISTINCT CASE WHEN lsi.is_chpl = 'CHPL' AND v.name IS NOT NULL THEN v.name END) AS lantern_chpl_dev_count
FROM shared_list_sources sls
LEFT JOIN list_source_info lsi    ON sls.list_source = lsi.list_source
LEFT JOIN fhir_endpoints fe       ON lsi.list_source = fe.list_source
LEFT JOIN fhir_endpoints_info fei ON fe.url = fei.url AND fei.requested_fhir_version = 'None'
LEFT JOIN vendors v               ON fei.vendor_id = v.id;

-- Unique index on constant expression enables REFRESH CONCURRENTLY for this single-row MV
CREATE UNIQUE INDEX idx_mv_chpl_coverage_summary_unique
    ON mv_chpl_coverage_summary((1));

COMMIT;
