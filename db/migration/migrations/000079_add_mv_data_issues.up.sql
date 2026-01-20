BEGIN;

-- ========================================
-- Migration 000077: Add Data Issues Materialized Views
-- ========================================
-- Purpose: Create materialized views to track data quality issues for developers
-- Tracks:
--   1. Developers with endpoints that have no organization data
--   2. Endpoints sharing list_source URLs
--   3. List_source URL accessibility issues
-- ========================================

-- Drop existing views if they exist
DROP MATERIALIZED VIEW IF EXISTS mv_developer_data_issues CASCADE;
DROP MATERIALIZED VIEW IF EXISTS mv_data_issues_summary CASCADE;

-- ========================================
-- MATERIALIZED VIEW: mv_data_issues_summary
-- ========================================
-- Purpose: System-wide summary statistics for data issues
-- ========================================

CREATE MATERIALIZED VIEW mv_data_issues_summary AS
WITH
-- Developers with endpoints that have no organization data
developers_with_no_org_data AS (
    SELECT
        COALESCE(v.name, 'Unknown') as vendor_name,
        COUNT(DISTINCT sfem.url) as no_org_data_endpoint_count
    FROM selected_fhir_endpoints_mv sfem
    LEFT JOIN fhir_endpoints_info fei ON sfem.url = fei.url AND fei.requested_fhir_version = 'None'
    LEFT JOIN vendors v ON fei.vendor_id = v.id
    WHERE
        (sfem.endpoint_names IS NULL OR sfem.endpoint_names = '' OR TRIM(sfem.endpoint_names) = '')
        AND sfem.requested_fhir_version = 'None'
    GROUP BY v.name
    HAVING COUNT(DISTINCT sfem.url) > 0
),
-- Count of endpoints with no organization data
endpoints_with_no_org_data AS (
    SELECT COUNT(DISTINCT sfem.url) as count
    FROM selected_fhir_endpoints_mv sfem
    WHERE
        (sfem.endpoint_names IS NULL OR sfem.endpoint_names = '' OR TRIM(sfem.endpoint_names) = '')
        AND sfem.requested_fhir_version = 'None'
),
-- Count of developers sharing list sources with at least one other developer (from shared_list_sources table)
developers_sharing_list_sources_count AS (
    SELECT COUNT(DISTINCT developer_name) as count
    FROM shared_list_sources
    WHERE list_source IN (
        SELECT list_source
        FROM shared_list_sources
        GROUP BY list_source
        HAVING COUNT(DISTINCT developer_name) > 1
    )
),
-- Inaccessible list_source URLs (HTTP errors)
-- Only counts list_sources where ALL endpoints are inaccessible (HTTP >= 400)
inaccessible_list_sources AS (
    SELECT
        fe.list_source
    FROM fhir_endpoints fe
    INNER JOIN fhir_endpoints_metadata fem ON fe.url = fem.url
    WHERE
        fe.list_source IS NOT NULL
        AND fe.list_source != ''
        AND fem.http_response IS NOT NULL
        AND fem.requested_fhir_version = 'None'
    GROUP BY fe.list_source
    -- Only include list_sources where ALL endpoints have HTTP response >= 400
    HAVING COUNT(*) = COUNT(CASE WHEN fem.http_response >= 400 THEN 1 END)
        AND COUNT(CASE WHEN fem.http_response >= 400 THEN 1 END) > 0
),
-- Endpoints from inaccessible list_sources
endpoints_with_inaccessible_list_sources AS (
    SELECT
        COUNT(DISTINCT fe.url) as count
    FROM fhir_endpoints fe
    INNER JOIN inaccessible_list_sources ils ON fe.list_source = ils.list_source
),
-- Developers with empty FHIR bundles (list_sources with no endpoints)
-- Uses shared_list_sources table to find developers whose list_source returns no endpoints
-- NOTE: This is CHPL-only because shared_list_sources only contains CHPL developers from CSV
developers_with_empty_bundles AS (
    SELECT
        sls.developer_name,
        sls.list_source
    FROM shared_list_sources sls
    LEFT JOIN fhir_endpoints fe ON sls.list_source = fe.list_source
    GROUP BY sls.developer_name, sls.list_source
    HAVING COUNT(fe.url) = 0
)
SELECT
    (SELECT COUNT(*) FROM developers_with_no_org_data) as developers_with_no_org_data_count,
    (SELECT count FROM endpoints_with_no_org_data) as endpoints_with_no_org_data_count,
    (SELECT count FROM developers_sharing_list_sources_count) as shared_list_sources_count,
    (SELECT count FROM developers_sharing_list_sources_count) as developers_sharing_list_sources_count,
    (SELECT COUNT(*) FROM inaccessible_list_sources) as inaccessible_list_sources_count,
    (SELECT count FROM endpoints_with_inaccessible_list_sources) as endpoints_with_inaccessible_list_sources_count,
    (SELECT COUNT(DISTINCT developer_name) FROM developers_with_empty_bundles) as developers_with_empty_bundles_count;

-- Create index for faster refresh
CREATE INDEX idx_mv_data_issues_summary ON mv_data_issues_summary(developers_with_no_org_data_count);

-- ========================================
-- MATERIALIZED VIEW: mv_developer_data_issues
-- ========================================
-- Purpose: Detailed developer-level data issues tracking
-- ========================================

CREATE MATERIALIZED VIEW mv_developer_data_issues AS
WITH
-- Get all unique vendors from fhir_endpoints_info AND shared_list_sources (to include empty bundle developers)
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
    LEFT JOIN fhir_endpoints fe ON sls.list_source = fe.list_source
    GROUP BY sls.developer_name, sls.list_source
    HAVING COUNT(fe.url) = 0
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
vendor_accessible_endpoints AS (
    SELECT
        COALESCE(v.name, 'Unknown') as vendor_name,
        COUNT(DISTINCT fem.url) as accessible_endpoints
    FROM fhir_endpoints_metadata fem
    INNER JOIN fhir_endpoints_info fei ON fem.url = fei.url AND fem.requested_fhir_version = fei.requested_fhir_version
    LEFT JOIN vendors v ON fei.vendor_id = v.id
    WHERE
        fem.http_response = 200
        AND fem.requested_fhir_version = 'None'
    GROUP BY v.name
),
-- Inaccessible endpoints per vendor
vendor_inaccessible_endpoints AS (
    SELECT
        COALESCE(v.name, 'Unknown') as vendor_name,
        COUNT(DISTINCT fem.url) as inaccessible_endpoints
    FROM fhir_endpoints_metadata fem
    INNER JOIN fhir_endpoints_info fei ON fem.url = fei.url AND fem.requested_fhir_version = fei.requested_fhir_version
    LEFT JOIN vendors v ON fei.vendor_id = v.id
    WHERE
        fem.http_response IS NOT NULL
        AND fem.http_response >= 400
        AND fem.requested_fhir_version = 'None'
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
    LEFT JOIN fhir_endpoints fe ON sls.list_source = fe.list_source
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
    END as shares_list_source
FROM all_vendors av
LEFT JOIN vendor_endpoints ve ON av.vendor_name = ve.vendor_name
LEFT JOIN vendor_endpoints_with_data vewd ON av.vendor_name = vewd.vendor_name
LEFT JOIN vendor_no_org_data vnod ON av.vendor_name = vnod.vendor_name
LEFT JOIN vendor_accessible_endpoints vae ON av.vendor_name = vae.vendor_name
LEFT JOIN vendor_inaccessible_endpoints vie ON av.vendor_name = vie.vendor_name
LEFT JOIN vendor_organizations vo ON av.vendor_name = vo.vendor_name
LEFT JOIN developers_empty_bundles deb ON av.vendor_name = deb.vendor_name
LEFT JOIN vendors_sharing_list_sources vsls ON av.vendor_name = vsls.vendor_name
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

COMMIT;
