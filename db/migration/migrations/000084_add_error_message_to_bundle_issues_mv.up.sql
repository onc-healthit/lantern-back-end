BEGIN;

-- ========================================
-- Migration 000084: Add error_message and sharing_group_id to mv_developer_bundle_issues
-- ========================================
-- Purpose:
--   error_message: Surface the most recent error from endpoint_query_errors per bundle URL,
--     displayed as the "Comments" column in the Developer Feedback tab.
--   sharing_group_id: Assign the same integer to all developers whose resolved FHIR endpoint
--     URL sets are identical, enabling visual peer-grouping in the UI without adding visible
--     columns. Developers with no sharing peer get NULL.
-- Depends on: mv_developer_data_issues (migration 000081),
--             endpoint_query_errors (migration 000083)
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
        COUNT(DISTINCT feo.id) AS organization_count
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
),
-- Most recent error message per bundle URL from endpoint_query_errors
latest_query_errors AS (
    SELECT DISTINCT ON (list_source)
        list_source,
        error_message
    FROM endpoint_query_errors
    ORDER BY list_source, queried_at DESC NULLS LAST
),
-- Endpoint URL sets per developer (used to identify sharing groups)
dev_endpoint_sets AS (
    SELECT
        sls.developer_name,
        ARRAY_AGG(DISTINCT fe.url ORDER BY fe.url) AS endpoint_set
    FROM shared_list_sources sls
    INNER JOIN fhir_endpoints fe ON sls.list_source = fe.list_source
    GROUP BY sls.developer_name
),
-- Developers that share their endpoint set with at least one other developer
sharing_devs AS (
    SELECT DISTINCT d1.developer_name
    FROM dev_endpoint_sets d1
    JOIN dev_endpoint_sets d2
        ON d1.developer_name != d2.developer_name
        AND d1.endpoint_set = d2.endpoint_set
),
-- Assign a stable integer group ID per unique endpoint set (only for sharing devs)
endpoint_group_ids AS (
    SELECT
        developer_name,
        DENSE_RANK() OVER (ORDER BY endpoint_set) AS sharing_group_id
    FROM dev_endpoint_sets
    WHERE developer_name IN (SELECT developer_name FROM sharing_devs)
)
SELECT
    sls.developer_name,
    sls.list_source,
    COALESCE(bte.total_endpoints, 0)            AS total_endpoints,
    COALESCE(bewd.endpoints_with_org_data, 0)   AS endpoints_with_org_data,
    COALESCE(bnod.no_org_data_endpoints, 0)     AS no_org_data_endpoints,
    COALESCE(bo.organization_count, 0)           AS organization_count,
    CASE WHEN COALESCE(bte.total_endpoints, 0) = 0
         THEN TRUE ELSE FALSE END                AS has_empty_bundle,
    CASE WHEN su.list_source IS NOT NULL
         THEN TRUE ELSE FALSE END                AS shares_list_source,
    COALESCE(dsf.shares_fhir_endpoints, FALSE)   AS shares_fhir_endpoints,
    TRUE                                         AS is_chpl_developer,
    COALESCE(lqe.error_message, 'N/A')           AS error_message,
    eg.sharing_group_id
FROM shared_list_sources sls
LEFT JOIN bundle_total_endpoints     bte  ON sls.list_source = bte.list_source
LEFT JOIN bundle_endpoints_with_data bewd ON sls.list_source = bewd.list_source
LEFT JOIN bundle_no_org_data          bnod ON sls.list_source = bnod.list_source
LEFT JOIN bundle_organizations        bo   ON sls.list_source = bo.list_source
LEFT JOIN shared_urls                 su   ON sls.list_source = su.list_source
LEFT JOIN dev_shares_fhir             dsf  ON sls.developer_name = dsf.vendor_name
LEFT JOIN latest_query_errors         lqe  ON sls.list_source = lqe.list_source
LEFT JOIN endpoint_group_ids          eg   ON sls.developer_name = eg.developer_name
ORDER BY sls.developer_name, sls.list_source;

CREATE UNIQUE INDEX idx_mv_developer_bundle_issues_unique
    ON mv_developer_bundle_issues(developer_name, list_source);
CREATE INDEX idx_mv_developer_bundle_issues_developer
    ON mv_developer_bundle_issues(developer_name);
CREATE INDEX idx_mv_developer_bundle_issues_list_source
    ON mv_developer_bundle_issues(list_source);

COMMIT;
