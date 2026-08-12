BEGIN;

-- Restore mv_developer_bundle_issues to its migration 000081 state (no error_message column)

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
-- Preserve each raw organization's bundle relationship, then consolidate using the
-- same organization-detail fields as mv_organizations_final.
bundle_organization_groups AS (
    SELECT DISTINCT
        fe.list_source,
        moa.organization_name,
        moa.identifier_types_html,
        moa.identifier_values_html,
        moa.addresses_html,
        moa.org_urls_html,
        moa.fhir_versions_html,
        moa.vendor_names_html,
        moa.is_chpl_html,
        moa.identifier_types_csv,
        moa.identifier_values_csv,
        moa.addresses_csv,
        moa.org_urls_csv,
        moa.fhir_versions_csv,
        moa.vendor_names_csv,
        moa.is_chpl_csv
    FROM mv_organizations_aggregated moa
    INNER JOIN fhir_endpoint_organizations_map feom ON moa.org_id = feom.org_database_id
    INNER JOIN fhir_endpoints fe ON feom.id = fe.id
),
bundle_organizations AS (
    SELECT
        list_source,
        COUNT(*) AS organization_count
    FROM bundle_organization_groups
    GROUP BY list_source
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
    COALESCE(bte.total_endpoints, 0)            AS total_endpoints,
    COALESCE(bewd.endpoints_with_org_data, 0)   AS endpoints_with_org_data,
    COALESCE(bnod.no_org_data_endpoints, 0)     AS no_org_data_endpoints,
    COALESCE(bo.organization_count, 0)           AS organization_count,
    CASE WHEN COALESCE(bte.total_endpoints, 0) = 0
         THEN TRUE ELSE FALSE END                AS has_empty_bundle,
    CASE WHEN su.list_source IS NOT NULL
         THEN TRUE ELSE FALSE END                AS shares_list_source,
    COALESCE(dsf.shares_fhir_endpoints, FALSE)   AS shares_fhir_endpoints,
    TRUE                                         AS is_chpl_developer
FROM shared_list_sources sls
LEFT JOIN bundle_total_endpoints     bte  ON sls.list_source = bte.list_source
LEFT JOIN bundle_endpoints_with_data bewd ON sls.list_source = bewd.list_source
LEFT JOIN bundle_no_org_data          bnod ON sls.list_source = bnod.list_source
LEFT JOIN bundle_organizations        bo   ON sls.list_source = bo.list_source
LEFT JOIN shared_urls                 su   ON sls.list_source = su.list_source
LEFT JOIN dev_shares_fhir             dsf  ON sls.developer_name = dsf.vendor_name
ORDER BY sls.developer_name, sls.list_source;

CREATE UNIQUE INDEX idx_mv_developer_bundle_issues_unique
    ON mv_developer_bundle_issues(developer_name, list_source);
CREATE INDEX idx_mv_developer_bundle_issues_developer
    ON mv_developer_bundle_issues(developer_name);
CREATE INDEX idx_mv_developer_bundle_issues_list_source
    ON mv_developer_bundle_issues(list_source);

COMMIT;
