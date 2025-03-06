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

# Refresh the endpoint list organizations materialized view
docker exec -t lantern-back-end_postgres_1 psql -t -c "REFRESH MATERIALIZED VIEW mv_endpoint_list_organizations;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to refresh mv_endpoint_list_organizations." >> $log_file
}

# Refresh the endpoint locations materialized view
docker exec -t lantern-back-end_postgres_1 psql -t -c "REFRESH MATERIALIZED VIEW mv_endpoint_locations;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to refresh mv_endpoint_locations." >> $log_file
}

# Reindex the endpoint list organizations indexes
docker exec -t lantern-back-end_postgres_1 psql -t -c "REINDEX INDEX idx_mv_endpoint_list_org_fhir;" -U lantern -d lantern
docker exec -t lantern-back-end_postgres_1 psql -t -c "REINDEX INDEX idx_mv_endpoint_list_org_vendor;" -U lantern -d lantern
docker exec -t lantern-back-end_postgres_1 psql -t -c "REINDEX INDEX idx_mv_endpoint_list_org_url;" -U lantern -d lantern

# Reindex the endpoint locations indexes
docker exec -t lantern-back-end_postgres_1 psql -t -c "REINDEX INDEX idx_mv_endpoint_loc_fhir;" -U lantern -d lantern
docker exec -t lantern-back-end_postgres_1 psql -t -c "REINDEX INDEX idx_mv_endpoint_loc_vendor;" -U lantern -d lantern
docker exec -t lantern-back-end_postgres_1 psql -t -c "REINDEX INDEX idx_mv_endpoint_loc_url;" -U lantern -d lantern