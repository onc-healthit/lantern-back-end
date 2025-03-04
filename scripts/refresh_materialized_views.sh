#!/bin/sh
log_file="/etc/lantern/refresh_materialized_views_logs.txt"
echo "$(date +"%Y-%m-%d %H:%M:%S") - Refreshing and reindexing Lantern materialized views." >> $log_file

docker exec -t lantern-back-end_postgres_1 psql -t -c "REFRESH MATERIALIZED VIEW mv_contact_information;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to refresh mv_contact_information." >> $log_file
}

# Reindex to keep performance optimal
docker exec -t lantern-back-end_postgres_1 psql -t -c "REINDEX INDEX mv_contact_information_uniq;" -U lantern -d lantern
docker exec -t lantern-back-end_postgres_1 psql -t -c "REINDEX INDEX mv_contact_information_url_idx;" -U lantern -d lantern
docker exec -t lantern-back-end_postgres_1 psql -t -c "REINDEX INDEX mv_contact_information_fhir_version_idx;" -U lantern -d lantern
docker exec -t lantern-back-end_postgres_1 psql -t -c "REINDEX INDEX mv_contact_information_vendor_name_idx;" -U lantern -d lantern
docker exec -t lantern-back-end_postgres_1 psql -t -c "REINDEX INDEX mv_contact_information_has_contact_idx;" -U lantern -d lantern
docker exec -t lantern-back-end_postgres_1 psql -t -c "REINDEX INDEX mv_contact_information_contact_rank_idx;" -U lantern -d lantern

echo "$(date +"%Y-%m-%d %H:%M:%S") - done." >> $log_file