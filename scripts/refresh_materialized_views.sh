#!/bin/sh

log_file="/etc/lantern/logs/refresh_materialized_views_logs.txt"

echo "$(date +"%Y-%m-%d %H:%M:%S") - Refreshing and reindexing Lantern materialized views." >> $log_file

docker exec -t lantern-back-end_postgres_1 psql -t -c "REFRESH MATERIALIZED VIEW CONCURRENTLY mv_resource_interactions;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to refresh mv_resource_interactions." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS mv_resource_interactions_uniq;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop mv_resource_interactions_uniq." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE UNIQUE INDEX mv_resource_interactions_uniq ON mv_resource_interactions (vendor_name, fhir_version, resource_type, endpoint_count, operations);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create mv_resource_interactions_uniq." >> $log_file
}

echo "$(date +"%Y-%m-%d %H:%M:%S") - done." >> $log_file