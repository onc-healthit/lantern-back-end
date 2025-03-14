#!/bin/sh

log_file="/etc/lantern/logs/refresh_dashboard_views_logs.txt"
echo "$(date +"%Y-%m-%d %H:%M:%S") - Refreshing Lantern dashboard materialized views." >> $log_file

# First check if views exist and create them if needed, then create unique indexes
docker exec -t lantern-back-end_postgres_1 psql -t -c "
DO \$\$
BEGIN
    -- Check if mv_endpoint_totals exists
    IF EXISTS (SELECT 1 FROM pg_matviews WHERE matviewname = 'mv_endpoint_totals') THEN
        -- Create unique index if it doesn't exist
        IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_mv_endpoint_totals_unique') THEN
            EXECUTE 'CREATE UNIQUE INDEX idx_mv_endpoint_totals_unique ON mv_endpoint_totals(aggregation_date)';
        END IF;
        
        -- Refresh the view
        EXECUTE 'REFRESH MATERIALIZED VIEW CONCURRENTLY mv_endpoint_totals';
    ELSE
        RAISE NOTICE 'mv_endpoint_totals does not exist, skipping refresh';
    END IF;
    
    -- Check if mv_response_tally exists
    IF EXISTS (SELECT 1 FROM pg_matviews WHERE matviewname = 'mv_response_tally') THEN
        -- Create unique index if it doesn't exist
        IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_mv_response_tally_unique') THEN
            EXECUTE 'CREATE UNIQUE INDEX idx_mv_response_tally_unique ON mv_response_tally(http_200)';
        END IF;
        
        -- Refresh the view
        EXECUTE 'REFRESH MATERIALIZED VIEW CONCURRENTLY mv_response_tally';
    ELSE
        RAISE NOTICE 'mv_response_tally does not exist, skipping refresh';
    END IF;
    
    -- Check if mv_vendor_fhir_counts exists
    IF EXISTS (SELECT 1 FROM pg_matviews WHERE matviewname = 'mv_vendor_fhir_counts') THEN
        -- Create unique index if it doesn't exist
        IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_mv_vendor_fhir_counts_unique') THEN
            EXECUTE 'CREATE UNIQUE INDEX idx_mv_vendor_fhir_counts_unique ON mv_vendor_fhir_counts(vendor_name, fhir_version)';
        END IF;
        
        -- Refresh the view
        EXECUTE 'REFRESH MATERIALIZED VIEW CONCURRENTLY mv_vendor_fhir_counts';
    ELSE
        RAISE NOTICE 'mv_vendor_fhir_counts does not exist, skipping refresh';
    END IF;
END
\$\$;
" -U lantern -d lantern 2>> $log_file

echo "$(date +"%Y-%m-%d %H:%M:%S") - done." >> $log_file