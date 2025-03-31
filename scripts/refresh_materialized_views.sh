#!/bin/sh

log_file="/etc/lantern/logs/refresh_materialized_views_logs.txt"

echo "$(date +"%Y-%m-%d %H:%M:%S") - Refreshing and reindexing Lantern materialized views." >> $log_file

# Refresh mv_vendor_fhir_counts
docker exec -t lantern-back-end_postgres_1 psql -t -c "REFRESH MATERIALIZED VIEW CONCURRENTLY mv_vendor_fhir_counts;" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to refresh mv_vendor_fhir_counts." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_vendor_fhir_counts_unique;" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_vendor_fhir_counts_unique." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE UNIQUE INDEX idx_mv_vendor_fhir_counts_unique ON mv_vendor_fhir_counts(vendor_name, fhir_version);" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_vendor_fhir_counts_unique." >> $log_file
}

# Add new indexes for mv_vendor_fhir_counts
docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_vendor_fhir_counts_vendor;" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_vendor_fhir_counts_vendor." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_vendor_fhir_counts_vendor ON mv_vendor_fhir_counts(vendor_name);" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_vendor_fhir_counts_vendor." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_vendor_fhir_counts_fhir;" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_vendor_fhir_counts_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_vendor_fhir_counts_fhir ON mv_vendor_fhir_counts(fhir_version);" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_vendor_fhir_counts_fhir." >> $log_file
}

# Refresh mv_response_tally
docker exec -t lantern-back-end_postgres_1 psql -t -c "REFRESH MATERIALIZED VIEW CONCURRENTLY mv_response_tally;" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to refresh mv_response_tally." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_response_tally_http_code;" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_response_tally_http_code." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE UNIQUE INDEX idx_mv_response_tally_http_code ON mv_response_tally(http_200);" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_response_tally_http_code." >> $log_file
}

# Refresh mv_endpoint_totals
docker exec -t lantern-back-end_postgres_1 psql -t -c "REFRESH MATERIALIZED VIEW CONCURRENTLY mv_endpoint_totals;" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to refresh mv_endpoint_totals." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_endpoint_totals_date;" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_endpoint_totals_date." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE UNIQUE INDEX idx_mv_endpoint_totals_date ON mv_endpoint_totals(aggregation_date);" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_endpoint_totals_date." >> $log_file
}



docker exec -t lantern-back-end_postgres_1 psql -t -c "REFRESH MATERIALIZED VIEW CONCURRENTLY mv_http_responses;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to refresh mv_http_responses." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS mv_http_responses_uniq;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop mv_http_responses_uniq." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE UNIQUE INDEX mv_http_responses_uniq ON mv_http_responses (aggregation_date, vendor_name, http_code);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create mv_http_responses_uniq." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS mv_http_responses_vendor_name_idx;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop mv_http_responses_vendor_name_idx." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX mv_http_responses_vendor_name_idx ON mv_http_responses (vendor_name);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create mv_http_responses_vendor_name_idx." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "REFRESH MATERIALIZED VIEW CONCURRENTLY mv_resource_interactions;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to refresh mv_resource_interactions." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS mv_resource_interactions_uniq;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop mv_resource_interactions_uniq." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE UNIQUE INDEX mv_resource_interactions_uniq ON mv_resource_interactions (vendor_name, fhir_version, resource_type, endpoint_count, operations);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create mv_resource_interactions_uniq." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS mv_resource_interactions_vendor_name_idx;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop mv_resource_interactions_vendor_name_idx." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX mv_resource_interactions_vendor_name_idx ON mv_resource_interactions (vendor_name);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create mv_resource_interactions_vendor_name_idx." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS mv_resource_interactions_fhir_version_idx;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop mv_resource_interactions_fhir_version_idx." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX mv_resource_interactions_fhir_version_idx ON mv_resource_interactions (fhir_version);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create mv_resource_interactions_fhir_version_idx." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS mv_resource_interactions_resource_type_idx;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop mv_resource_interactions_resource_type_idx." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX mv_resource_interactions_resource_type_idx ON mv_resource_interactions (resource_type);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create mv_resource_interactions_resource_type_idx." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS mv_resource_interactions_operations_idx;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop mv_resource_interactions_operations_idx." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX mv_resource_interactions_operations_idx ON mv_resource_interactions USING GIN (operations);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create mv_resource_interactions_operations_idx." >> $log_file
}

# Refresh and reindex get_capstat_values_mv
docker exec -t lantern-back-end_postgres_1 psql -t -c "REFRESH MATERIALIZED VIEW CONCURRENTLY get_capstat_values_mv;" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to refresh get_capstat_values_mv." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_get_capstat_values_mv_unique;" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_get_capstat_values_mv_unique." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE UNIQUE INDEX idx_get_capstat_values_mv_unique ON get_capstat_values_mv(endpoint_id, vendor_id, filter_fhir_version);" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_get_capstat_values_mv_unique." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_get_capstat_values_mv_endpoint_id;" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_get_capstat_values_mv_endpoint_id." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_get_capstat_values_mv_endpoint_id ON get_capstat_values_mv(endpoint_id);" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_get_capstat_values_mv_endpoint_id." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_get_capstat_values_mv_vendor_id;" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_get_capstat_values_mv_vendor_id." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_get_capstat_values_mv_vendor_id ON get_capstat_values_mv(vendor_id);" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_get_capstat_values_mv_vendor_id." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_get_capstat_values_mv_filter_fhir_version;" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_get_capstat_values_mv_filter_fhir_version." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_get_capstat_values_mv_filter_fhir_version ON get_capstat_values_mv(filter_fhir_version);" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_get_capstat_values_mv_filter_fhir_version." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_get_capstat_values_mv_vendor_name;" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_get_capstat_values_mv_vendor_name." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_get_capstat_values_mv_vendor_name ON get_capstat_values_mv(vendor_name);" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_get_capstat_values_mv_vendor_name." >> $log_file
}

# Refresh and reindex selected_fhir_endpoints_values_mv
docker exec -t lantern-back-end_postgres_1 psql -t -c "REFRESH MATERIALIZED VIEW CONCURRENTLY selected_fhir_endpoints_values_mv;" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to refresh selected_fhir_endpoints_values_mv." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_selected_fhir_endpoints_unique;" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_selected_fhir_endpoints_unique." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE UNIQUE INDEX idx_selected_fhir_endpoints_unique ON selected_fhir_endpoints_values_mv("Developer", "FHIR Version", Field, field_value);" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_selected_fhir_endpoints_unique." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_selected_fhir_endpoints_dev;" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_selected_fhir_endpoints_dev." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_selected_fhir_endpoints_dev ON selected_fhir_endpoints_values_mv("Developer");" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_selected_fhir_endpoints_dev." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_selected_fhir_endpoints_fhir_version;" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_selected_fhir_endpoints_fhir_version." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_selected_fhir_endpoints_fhir_version ON selected_fhir_endpoints_values_mv("FHIR Version");" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_selected_fhir_endpoints_fhir_version." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_selected_fhir_endpoints_field;" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_selected_fhir_endpoints_field." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_selected_fhir_endpoints_field ON selected_fhir_endpoints_values_mv(Field);" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_selected_fhir_endpoints_field." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_selected_fhir_endpoints_field_value;" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_selected_fhir_endpoints_field_value." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_selected_fhir_endpoints_field_value ON selected_fhir_endpoints_values_mv(field_value);" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_selected_fhir_endpoints_field_value." >> $log_file
}

echo "$(date +"%Y-%m-%d %H:%M:%S") - done." >> $log_file