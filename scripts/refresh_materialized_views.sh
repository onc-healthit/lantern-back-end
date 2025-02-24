#!/bin/sh

log_file="/etc/lantern/refresh_materialized_views_logs.txt"
current_datetime=$(date +"%Y-%m-%d %H:%M:%S")

echo "$current_datetime - Refreshing and reindexing Lantern materialized views." >> $log_file

docker exec -t lantern-back-end-postgres-1 psql -t -c "REFRESH MATERIALIZED VIEW CONCURRENTLY mv_http_responses;" -U lantern -d lantern || {
    echo "$current_datetime - Lantern failed to refresh mv_http_responses." >> $log_file
}

docker exec -t lantern-back-end-postgres-1 psql -t -c "REINDEX INDEX CONCURRENTLY mv_http_responses_uniq;" -U lantern -d lantern || {
    echo "$current_datetime - Lantern failed to reindex mv_http_responses_uniq." >> $log_file
}

echo "$current_datetime - done." >> $log_file