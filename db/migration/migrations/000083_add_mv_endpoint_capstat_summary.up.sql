BEGIN;

-- Create materialized view for mv_endpoint_capstat_summary
DROP MATERIALIZED VIEW IF EXISTS mv_endpoint_capstat_summary CASCADE;

CREATE MATERIALIZED VIEW mv_endpoint_capstat_summary AS
SELECT 
    e.url,
    e.vendor_name,
    e.fhir_version,
    COALESCE(r.resource_count, 0) AS resources,
    COALESCE(s.search_param_count, 0) AS search_params,
    COALESCE(i.interaction_count, 0) AS interactions
FROM selected_fhir_endpoints_mv e
LEFT JOIN (
    SELECT 
        f.url,
        COUNT(DISTINCT type) AS resource_count
    FROM mv_endpoint_resource_types f
    GROUP BY f.url
) r ON e.url = r.url
LEFT JOIN (
    SELECT 
        f.url,
        COUNT(DISTINCT (resource_type || ':' || search_param)) AS search_param_count
    FROM mv_endpoint_search_params f
    GROUP BY f.url
) s ON e.url = s.url
LEFT JOIN (
    SELECT 
        f.url,
        COUNT(DISTINCT (resource_type || ':' || interaction_code)) AS interaction_count
    FROM mv_endpoint_interactions_extracted f
    GROUP BY f.url
) i ON e.url = i.url
WHERE e.cap_stat_exists = 'true'
ORDER BY e.vendor_name, e.fhir_version;

-- Create indexes for mv_endpoint_capstat_summary
CREATE INDEX idx_mv_endpoint_capstat_summary_url ON mv_endpoint_capstat_summary(url);
CREATE INDEX idx_mv_endpoint_capstat_summary_vendor ON mv_endpoint_capstat_summary(vendor_name);
CREATE INDEX idx_mv_endpoint_capstat_summary_fhir ON mv_endpoint_capstat_summary(fhir_version);

COMMIT;