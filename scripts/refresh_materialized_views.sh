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

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE UNIQUE INDEX idx_mv_vendor_fhir_counts_unique ON mv_vendor_fhir_counts(vendor_name, fhir_version, sort_order);" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_vendor_fhir_counts_unique." >> $log_file
}

# Add new indexes for mv_vendor_fhir_counts
docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_vendor_fhir_counts_vendor;" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_vendor_fhir_counts_vendor." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_vendor_fhir_counts_vendor ON mv_vendor_fhir_counts(vendor_name);" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_vendor_fhir_counts_vendor." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_vendor_fhir_counts_sort;" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_vendor_fhir_counts_sort." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_vendor_fhir_counts_sort ON mv_vendor_fhir_counts(sort_order);" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_vendor_fhir_counts_sort." >> $log_file
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

# Lantern-852
# Refresh the capstat sizes materialized view
docker exec -t lantern-back-end_postgres_1 psql -t -c "REFRESH MATERIALIZED VIEW CONCURRENTLY mv_capstat_sizes_tbl;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to refresh mv_capstat_sizes_tbl." >> $log_file
}

# Add new indexes for mv_capstat_sizes_tbl
docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_capstat_sizes_uniq;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop mv_capstat_sizes_tbl." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE UNIQUE INDEX idx_mv_capstat_sizes_uniq ON mv_capstat_sizes_tbl(url);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_capstat_sizes_uniq." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_capstat_sizes_vendor;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_capstat_sizes_vendor." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_capstat_sizes_vendor ON mv_capstat_sizes_tbl(vendor_name);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_capstat_sizes_vendor." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_capstat_sizes_fhir;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_capstat_sizes_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_capstat_sizes_fhir ON mv_capstat_sizes_tbl(fhir_version);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_capstat_sizes_fhir." >> $log_file
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

# Refresh and reindex get_capstat_fields_mv
docker exec -t lantern-back-end_postgres_1 psql -t -c "REFRESH MATERIALIZED VIEW CONCURRENTLY get_capstat_fields_mv;" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to refresh get_capstat_fields_mv." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_get_capstat_fields_mv_endpoint_id_field;" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_get_capstat_fields_mv_endpoint_id_field." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE UNIQUE INDEX idx_get_capstat_fields_mv_endpoint_id_field ON get_capstat_fields_mv(endpoint_id, field);" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_get_capstat_fields_mv_endpoint_id_field." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_get_capstat_fields_mv_fhir_version;" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_get_capstat_fields_mv_fhir_version." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_get_capstat_fields_mv_fhir_version ON get_capstat_fields_mv(fhir_version);" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_get_capstat_fields_mv_fhir_version." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_get_capstat_fields_mv_field;" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_get_capstat_fields_mv_field." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_get_capstat_fields_mv_field ON get_capstat_fields_mv(field);" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_get_capstat_fields_mv_field." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_get_capstat_fields_mv_vendor_id;" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_get_capstat_fields_mv_vendor_id." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_get_capstat_fields_mv_vendor_id ON get_capstat_fields_mv(vendor_id);" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_get_capstat_fields_mv_vendor_id." >> $log_file
}

# Refresh and reindex get_value_versions_mv
docker exec -t lantern-back-end_postgres_1 psql -t -c "REFRESH MATERIALIZED VIEW CONCURRENTLY get_value_versions_mv;" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to refresh get_value_versions_mv." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_get_value_versions_mv_field;" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_get_value_versions_mv_field." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE UNIQUE INDEX idx_get_value_versions_mv_field ON get_value_versions_mv(field);" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_get_value_versions_mv_field." >> $log_file
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

# Refresh mv_endpoint_organization_tbl
docker exec -t lantern-back-end_postgres_1 psql -t -c "REFRESH MATERIALIZED VIEW CONCURRENTLY mv_endpoint_organization_tbl;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to refresh mv_endpoint_organization_tbl." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_endpoint_list_org_url;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_endpoint_list_org_url." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE UNIQUE INDEX idx_mv_endpoint_list_org_url ON mv_endpoint_organization_tbl(url);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_endpoint_list_org_url." >> $log_file
}

# Refresh mv_endpoint_export_tbl
docker exec -t lantern-back-end_postgres_1 psql -t -c "REFRESH MATERIALIZED VIEW CONCURRENTLY mv_endpoint_export_tbl;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to refresh mv_endpoint_export_tbl." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_endpoint_export_tbl_unique_id;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_endpoint_export_tbl_unique_id." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE UNIQUE INDEX idx_mv_endpoint_export_tbl_unique_id ON mv_endpoint_export_tbl(mv_id);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_endpoint_export_tbl_unique_id." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_endpoint_export_tbl_vendor;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_endpoint_export_tbl_vendor." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_endpoint_export_tbl_vendor ON mv_endpoint_export_tbl(vendor_name);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_endpoint_export_tbl_vendor." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_endpoint_export_tbl_fhir;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_endpoint_export_tbl_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_endpoint_export_tbl_fhir ON mv_endpoint_export_tbl(fhir_version);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_endpoint_export_tbl_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_endpoint_export_tbl_vendor_fhir;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_endpoint_export_tbl_vendor_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_endpoint_export_tbl_vendor_fhir ON mv_endpoint_export_tbl(vendor_name, fhir_version);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_endpoint_export_tbl_vendor_fhir." >> $log_file
}

# Refresh mv_http_pct
docker exec -t lantern-back-end_postgres_1 psql -t -c "REFRESH MATERIALIZED VIEW CONCURRENTLY mv_http_pct;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to refresh mv_http_pct." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_http_pct_unique_id;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_http_pct_unique_id." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE UNIQUE INDEX idx_mv_http_pct_unique_id ON mv_http_pct(mv_id);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_http_pct_unique_id." >> $log_file
}
docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_http_pct_http_response;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_http_pct_http_response." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_http_pct_http_response ON mv_http_pct(http_response);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_http_pct_http_response." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_http_pct_vendor;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_http_pct_vendor." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_http_pct_vendor ON mv_http_pct(vendor_name);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_http_pct_vendor." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_http_pct_fhir;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_http_pct_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_http_pct_fhir ON mv_http_pct(fhir_version);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_http_pct_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_http_pct_vendor_fhir;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_http_pct_vendor_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_http_pct_vendor_fhir ON mv_http_pct(vendor_name, fhir_version);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_http_pct_vendor_fhir." >> $log_file
}

# Refresh mv_well_known_endpoints
docker exec -t lantern-back-end_postgres_1 psql -t -c "REFRESH MATERIALIZED VIEW CONCURRENTLY mv_well_known_endpoints;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to refresh mv_well_known_endpoints." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_well_known_unique_id;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_well_known_unique_id." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE UNIQUE INDEX idx_mv_well_known_unique_id ON mv_well_known_endpoints(mv_id);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_well_known_unique_id." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_well_known_vendor;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_well_known_vendor." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_well_known_vendor ON mv_well_known_endpoints(vendor_name);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_well_known_vendor." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_well_known_fhir;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_well_known_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_well_known_fhir ON mv_well_known_endpoints(fhir_version);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_well_known_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_well_known_vendor_fhir;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_well_known_vendor_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_well_known_vendor_fhir ON mv_well_known_endpoints(vendor_name, fhir_version);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_well_known_vendor_fhir." >> $log_file
}

# Refresh mv_well_known_no_doc
docker exec -t lantern-back-end_postgres_1 psql -t -c "REFRESH MATERIALIZED VIEW CONCURRENTLY mv_well_known_no_doc;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to refresh mv_well_known_no_doc." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_well_known_no_doc_unique_id;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_well_known_no_doc_unique_id." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE UNIQUE INDEX idx_mv_well_known_no_doc_unique_id ON mv_well_known_no_doc(mv_id);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_well_known_no_doc_unique_id." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_well_known_no_doc_url;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_well_known_no_doc_url." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_well_known_no_doc_url ON mv_well_known_no_doc(url);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_well_known_no_doc_url." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_well_known_no_doc_vendor;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_well_known_no_doc_vendor." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_well_known_no_doc_vendor ON mv_well_known_no_doc(vendor_name);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_well_known_no_doc_vendor." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_well_known_no_doc_fhir;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_well_known_no_doc_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_well_known_no_doc_fhir ON mv_well_known_no_doc(fhir_version);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_well_known_no_doc_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_well_known_no_doc_vendor_fhir;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_well_known_no_doc_vendor_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_well_known_no_doc_vendor_fhir ON mv_well_known_no_doc(vendor_name, fhir_version);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_well_known_no_doc_vendor_fhir." >> $log_file
}

# Refresh mv_smart_response_capabilities
docker exec -t lantern-back-end_postgres_1 psql -t -c "REFRESH MATERIALIZED VIEW CONCURRENTLY mv_smart_response_capabilities;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to refresh mv_smart_response_capabilities." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_smart_response_capabilities_unique_id;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_smart_response_capabilities_unique_id." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE UNIQUE INDEX idx_mv_smart_response_capabilities_unique_id ON mv_smart_response_capabilities(mv_id);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_smart_response_capabilities_unique_id." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_smart_response_capabilities_id;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_smart_response_capabilities_id." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_smart_response_capabilities_id ON mv_smart_response_capabilities(id);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_smart_response_capabilities_id." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_smart_response_capabilities_vendor;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_smart_response_capabilities_vendor." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_smart_response_capabilities_vendor ON mv_smart_response_capabilities(vendor_name);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_smart_response_capabilities_vendor." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_smart_response_capabilities_fhir;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_smart_response_capabilities_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_smart_response_capabilities_fhir ON mv_smart_response_capabilities(fhir_version);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_smart_response_capabilities_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_smart_response_capabilities_capability;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_smart_response_capabilities_capability." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_smart_response_capabilities_capability ON mv_smart_response_capabilities(capability);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_smart_response_capabilities_capability." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_smart_response_capabilities_vendor_fhir;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_smart_response_capabilities_vendor_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_smart_response_capabilities_vendor_fhir ON mv_smart_response_capabilities(vendor_name, fhir_version);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_smart_response_capabilities_vendor_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_smart_response_capabilities_capability_fhir;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_smart_response_capabilities_capability_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_smart_response_capabilities_capability_fhir ON mv_smart_response_capabilities(capability, fhir_version);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_smart_response_capabilities_capability_fhir." >> $log_file
}

# Refresh mv_selected_endpoints
docker exec -t lantern-back-end_postgres_1 psql -t -c "REFRESH MATERIALIZED VIEW CONCURRENTLY mv_selected_endpoints;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to refresh mv_selected_endpoints." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_selected_endpoints_unique_id;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_selected_endpoints_unique_id." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE UNIQUE INDEX idx_mv_selected_endpoints_unique_id ON mv_selected_endpoints(mv_id);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_selected_endpoints_unique_id." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_selected_endpoints_vendor;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_selected_endpoints_vendor." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_selected_endpoints_vendor ON mv_selected_endpoints(vendor_name);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_selected_endpoints_vendor." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_selected_endpoints_fhir;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_selected_endpoints_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_selected_endpoints_fhir ON mv_selected_endpoints(capability_fhir_version);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_selected_endpoints_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_selected_endpoints_vendor_fhir;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_selected_endpoints_vendor_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_selected_endpoints_vendor_fhir ON mv_selected_endpoints(vendor_name, capability_fhir_version);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_selected_endpoints_vendor_fhir." >> $log_file
}

# Refresh mv_endpoint_export_tbl
docker exec -t lantern-back-end_postgres_1 psql -t -c "REFRESH MATERIALIZED VIEW CONCURRENTLY mv_endpoint_export_tbl;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to refresh mv_endpoint_export_tbl." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_endpoint_export_tbl_unique_id;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_endpoint_export_tbl_unique_id." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE UNIQUE INDEX idx_mv_endpoint_export_tbl_unique_id ON mv_endpoint_export_tbl(mv_id);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_endpoint_export_tbl_unique_id." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_endpoint_export_tbl_vendor;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_endpoint_export_tbl_vendor." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_endpoint_export_tbl_vendor ON mv_endpoint_export_tbl(vendor_name);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_endpoint_export_tbl_vendor." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_endpoint_export_tbl_fhir;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_endpoint_export_tbl_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_endpoint_export_tbl_fhir ON mv_endpoint_export_tbl(fhir_version);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_endpoint_export_tbl_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_endpoint_export_tbl_vendor_fhir;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_endpoint_export_tbl_vendor_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_endpoint_export_tbl_vendor_fhir ON mv_endpoint_export_tbl(vendor_name, fhir_version);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_endpoint_export_tbl_vendor_fhir." >> $log_file
}

# Refresh mv_http_pct
docker exec -t lantern-back-end_postgres_1 psql -t -c "REFRESH MATERIALIZED VIEW CONCURRENTLY mv_http_pct;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to refresh mv_http_pct." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_http_pct_unique_id;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_http_pct_unique_id." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE UNIQUE INDEX idx_mv_http_pct_unique_id ON mv_http_pct(mv_id);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_http_pct_unique_id." >> $log_file
}
docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_http_pct_http_response;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_http_pct_http_response." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_http_pct_http_response ON mv_http_pct(http_response);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_http_pct_http_response." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_http_pct_vendor;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_http_pct_vendor." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_http_pct_vendor ON mv_http_pct(vendor_name);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_http_pct_vendor." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_http_pct_fhir;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_http_pct_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_http_pct_fhir ON mv_http_pct(fhir_version);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_http_pct_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_http_pct_vendor_fhir;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_http_pct_vendor_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_http_pct_vendor_fhir ON mv_http_pct(vendor_name, fhir_version);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_http_pct_vendor_fhir." >> $log_file
}

# Refresh mv_well_known_endpoints
docker exec -t lantern-back-end_postgres_1 psql -t -c "REFRESH MATERIALIZED VIEW CONCURRENTLY mv_well_known_endpoints;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to refresh mv_well_known_endpoints." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_well_known_unique_id;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_well_known_unique_id." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE UNIQUE INDEX idx_mv_well_known_unique_id ON mv_well_known_endpoints(mv_id);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_well_known_unique_id." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_well_known_vendor;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_well_known_vendor." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_well_known_vendor ON mv_well_known_endpoints(vendor_name);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_well_known_vendor." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_well_known_fhir;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_well_known_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_well_known_fhir ON mv_well_known_endpoints(fhir_version);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_well_known_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_well_known_vendor_fhir;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_well_known_vendor_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_well_known_vendor_fhir ON mv_well_known_endpoints(vendor_name, fhir_version);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_well_known_vendor_fhir." >> $log_file
}

# Refresh mv_well_known_no_doc
docker exec -t lantern-back-end_postgres_1 psql -t -c "REFRESH MATERIALIZED VIEW CONCURRENTLY mv_well_known_no_doc;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to refresh mv_well_known_no_doc." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_well_known_no_doc_unique_id;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_well_known_no_doc_unique_id." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE UNIQUE INDEX idx_mv_well_known_no_doc_unique_id ON mv_well_known_no_doc(mv_id);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_well_known_no_doc_unique_id." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_well_known_no_doc_url;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_well_known_no_doc_url." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_well_known_no_doc_url ON mv_well_known_no_doc(url);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_well_known_no_doc_url." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_well_known_no_doc_vendor;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_well_known_no_doc_vendor." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_well_known_no_doc_vendor ON mv_well_known_no_doc(vendor_name);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_well_known_no_doc_vendor." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_well_known_no_doc_fhir;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_well_known_no_doc_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_well_known_no_doc_fhir ON mv_well_known_no_doc(fhir_version);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_well_known_no_doc_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_well_known_no_doc_vendor_fhir;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_well_known_no_doc_vendor_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_well_known_no_doc_vendor_fhir ON mv_well_known_no_doc(vendor_name, fhir_version);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_well_known_no_doc_vendor_fhir." >> $log_file
}

# Refresh mv_smart_response_capabilities
docker exec -t lantern-back-end_postgres_1 psql -t -c "REFRESH MATERIALIZED VIEW CONCURRENTLY mv_smart_response_capabilities;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to refresh mv_smart_response_capabilities." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_smart_response_capabilities_unique_id;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_smart_response_capabilities_unique_id." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE UNIQUE INDEX idx_mv_smart_response_capabilities_unique_id ON mv_smart_response_capabilities(mv_id);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_smart_response_capabilities_unique_id." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_smart_response_capabilities_id;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_smart_response_capabilities_id." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_smart_response_capabilities_id ON mv_smart_response_capabilities(id);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_smart_response_capabilities_id." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_smart_response_capabilities_vendor;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_smart_response_capabilities_vendor." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_smart_response_capabilities_vendor ON mv_smart_response_capabilities(vendor_name);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_smart_response_capabilities_vendor." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_smart_response_capabilities_fhir;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_smart_response_capabilities_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_smart_response_capabilities_fhir ON mv_smart_response_capabilities(fhir_version);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_smart_response_capabilities_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_smart_response_capabilities_capability;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_smart_response_capabilities_capability." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_smart_response_capabilities_capability ON mv_smart_response_capabilities(capability);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_smart_response_capabilities_capability." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_smart_response_capabilities_vendor_fhir;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_smart_response_capabilities_vendor_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_smart_response_capabilities_vendor_fhir ON mv_smart_response_capabilities(vendor_name, fhir_version);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_smart_response_capabilities_vendor_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_smart_response_capabilities_capability_fhir;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_smart_response_capabilities_capability_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_smart_response_capabilities_capability_fhir ON mv_smart_response_capabilities(capability, fhir_version);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_smart_response_capabilities_capability_fhir." >> $log_file
}

# Refresh mv_selected_endpoints
docker exec -t lantern-back-end_postgres_1 psql -t -c "REFRESH MATERIALIZED VIEW CONCURRENTLY mv_selected_endpoints;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to refresh mv_selected_endpoints." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_selected_endpoints_unique_id;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_selected_endpoints_unique_id." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE UNIQUE INDEX idx_mv_selected_endpoints_unique_id ON mv_selected_endpoints(mv_id);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_selected_endpoints_unique_id." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_selected_endpoints_vendor;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_selected_endpoints_vendor." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_selected_endpoints_vendor ON mv_selected_endpoints(vendor_name);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_selected_endpoints_vendor." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_selected_endpoints_fhir;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_selected_endpoints_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_selected_endpoints_fhir ON mv_selected_endpoints(capability_fhir_version);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_selected_endpoints_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_selected_endpoints_vendor_fhir;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_selected_endpoints_vendor_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_selected_endpoints_vendor_fhir ON mv_selected_endpoints(vendor_name, capability_fhir_version);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_selected_endpoints_vendor_fhir." >> $log_file
}

# Refresh security_endpoints_mv
docker exec -t lantern-back-end_postgres_1 psql -t -c "REFRESH MATERIALIZED VIEW CONCURRENTLY security_endpoints_mv;" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to refresh security_endpoints_mv." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_unique_security_endpoints;" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_unique_security_endpoints." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE UNIQUE INDEX idx_unique_security_endpoints ON security_endpoints_mv (id, url, vendor_name, code);" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_unique_security_endpoints." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_security_endpoints_url;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_security_endpoints_url." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_security_endpoints_url ON security_endpoints_mv (url);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_security_endpoints_url." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_security_endpoints_fhir_version;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_security_endpoints_fhir_version." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_security_endpoints_fhir_version ON security_endpoints_mv (fhir_version_final);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_security_endpoints_fhir_version." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_security_endpoints_vendor_name;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_security_endpoints_vendor_name." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_security_endpoints_vendor_name ON security_endpoints_mv (vendor_name);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_security_endpoints_vendor_name." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_security_endpoints_code;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_security_endpoints_code." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_security_endpoints_code ON security_endpoints_mv (code);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_security_endpoints_code." >> $log_file
}

# Refresh selected_security_endpoints_mv
docker exec -t lantern-back-end_postgres_1 psql -t -c "REFRESH MATERIALIZED VIEW CONCURRENTLY selected_security_endpoints_mv;" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to refresh selected_security_endpoints_mv." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_unique_selected_security_endpoints;" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_unique_selected_security_endpoints." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE UNIQUE INDEX idx_unique_selected_security_endpoints ON selected_security_endpoints_mv (id, url, code);" -U lantern -d lantern || { 
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_unique_selected_security_endpoints." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_selected_security_endpoints_fhir_version;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_selected_security_endpoints_fhir_version." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_selected_security_endpoints_fhir_version ON selected_security_endpoints_mv (fhir_version);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_selected_security_endpoints_fhir_version." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_selected_security_endpoints_vendor_name;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_selected_security_endpoints_vendor_name." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_selected_security_endpoints_vendor_name ON selected_security_endpoints_mv (vendor_name);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_selected_security_endpoints_vendor_name." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_selected_security_endpoints_code;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_selected_security_endpoints_code." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_selected_security_endpoints_code ON selected_security_endpoints_mv (code);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_selected_security_endpoints_code." >> $log_file
}

# Refresh mv_validation_results_plot
docker exec -t lantern-back-end_postgres_1 psql -t -c "REFRESH MATERIALIZED VIEW CONCURRENTLY mv_validation_results_plot;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to refresh mv_validation_results_plot." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS mv_validation_results_plot_unique_idx;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop mv_validation_results_plot_unique_idx." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE UNIQUE INDEX mv_validation_results_plot_unique_idx ON mv_validation_results_plot(url, fhir_version, vendor_name, rule_name, valid, expected, actual);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create mv_validation_results_plot_unique_idx." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS mv_validation_results_plot_vendor_idx;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop mv_validation_results_plot_vendor_idx." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX mv_validation_results_plot_vendor_idx ON mv_validation_results_plot(vendor_name);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create mv_validation_results_plot_vendor_idx." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS mv_validation_results_plot_fhir_idx;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop mv_validation_results_plot_fhir_idx." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX mv_validation_results_plot_fhir_idx ON mv_validation_results_plot(fhir_version);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create mv_validation_results_plot_fhir_idx." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS mv_validation_results_plot_rule_idx;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop mv_validation_results_plot_rule_idx." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX mv_validation_results_plot_rule_idx ON mv_validation_results_plot(rule_name);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create mv_validation_results_plot_rule_idx." >> $log_file
}

# Refresh mv_validation_details
docker exec -t lantern-back-end_postgres_1 psql -t -c "REFRESH MATERIALIZED VIEW CONCURRENTLY mv_validation_details;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to refresh mv_validation_details." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS mv_validation_details_unique_idx;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop mv_validation_details_unique_idx." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE UNIQUE INDEX mv_validation_details_unique_idx ON mv_validation_details(rule_name);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create mv_validation_details_unique_idx." >> $log_file
}

# Refresh mv_validation_failures
docker exec -t lantern-back-end_postgres_1 psql -t -c "REFRESH MATERIALIZED VIEW CONCURRENTLY mv_validation_failures;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to refresh mv_validation_failures." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS mv_validation_failures_unique_idx;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop mv_validation_failures_unique_idx." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE UNIQUE INDEX mv_validation_failures_unique_idx ON mv_validation_failures(url, fhir_version, vendor_name, rule_name);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create mv_validation_failures_unique_idx." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS mv_validation_failures_url_idx;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop mv_validation_failures_url_idx." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX mv_validation_failures_url_idx ON mv_validation_failures(url);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create mv_validation_failures_url_idx." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS mv_validation_failures_rule_name_idx;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop mv_validation_failures_rule_name_idx." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX mv_validation_failures_rule_name_idx ON mv_validation_failures(rule_name);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create mv_validation_failures_rule_name_idx." >> $log_file
}

# Refresh mv_validation_results_plot
docker exec -t lantern-back-end_postgres_1 psql -t -c "REFRESH MATERIALIZED VIEW CONCURRENTLY mv_validation_results_plot;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to refresh mv_validation_results_plot." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS mv_validation_results_plot_unique_idx;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop mv_validation_results_plot_unique_idx." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE UNIQUE INDEX mv_validation_results_plot_unique_idx ON mv_validation_results_plot(url, fhir_version, vendor_name, rule_name, valid, expected, actual);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create mv_validation_results_plot_unique_idx." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS mv_validation_results_plot_vendor_idx;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop mv_validation_results_plot_vendor_idx." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX mv_validation_results_plot_vendor_idx ON mv_validation_results_plot(vendor_name);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create mv_validation_results_plot_vendor_idx." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS mv_validation_results_plot_fhir_idx;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop mv_validation_results_plot_fhir_idx." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX mv_validation_results_plot_fhir_idx ON mv_validation_results_plot(fhir_version);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create mv_validation_results_plot_fhir_idx." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS mv_validation_results_plot_rule_idx;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop mv_validation_results_plot_rule_idx." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX mv_validation_results_plot_rule_idx ON mv_validation_results_plot(rule_name);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create mv_validation_results_plot_rule_idx." >> $log_file
}

# Refresh mv_validation_details
docker exec -t lantern-back-end_postgres_1 psql -t -c "REFRESH MATERIALIZED VIEW CONCURRENTLY mv_validation_details;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to refresh mv_validation_details." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS mv_validation_details_unique_idx;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop mv_validation_details_unique_idx." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE UNIQUE INDEX mv_validation_details_unique_idx ON mv_validation_details(rule_name);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create mv_validation_details_unique_idx." >> $log_file
}

# Refresh mv_validation_failures
docker exec -t lantern-back-end_postgres_1 psql -t -c "REFRESH MATERIALIZED VIEW CONCURRENTLY mv_validation_failures;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to refresh mv_validation_failures." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS mv_validation_failures_unique_idx;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop mv_validation_failures_unique_idx." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE UNIQUE INDEX mv_validation_failures_unique_idx ON mv_validation_failures(url, fhir_version, vendor_name, rule_name);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create mv_validation_failures_unique_idx." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS mv_validation_failures_url_idx;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop mv_validation_failures_url_idx." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX mv_validation_failures_url_idx ON mv_validation_failures(url);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create mv_validation_failures_url_idx." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS mv_validation_failures_rule_name_idx;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop mv_validation_failures_rule_name_idx." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX mv_validation_failures_rule_name_idx ON mv_validation_failures(rule_name);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create mv_validation_failures_rule_name_idx." >> $log_file
}

# Lantern-839
# Refresh the endpoint list organizations materialized view
docker exec -t lantern-back-end_postgres_1 psql -t -c "REFRESH MATERIALIZED VIEW CONCURRENTLY mv_endpoint_list_organizations;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to refresh mv_endpoint_list_organizations." >> $log_file
}

# Add new indexes for mv_endpoint_list_organizations
docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_endpoint_list_org_uniq;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop mv_endpoint_list_organizations." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE UNIQUE INDEX idx_mv_endpoint_list_org_uniq ON mv_endpoint_list_organizations(fhir_version, vendor_name, url, organization_name, requested_fhir_version);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create mv_endpoint_list_organizations." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_endpoint_list_org_vendor;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_endpoint_list_org_vendor." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_endpoint_list_org_vendor ON mv_endpoint_list_organizations(vendor_name);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_endpoint_list_org_vendor." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_endpoint_list_org_fhir;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_endpoint_list_org_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_endpoint_list_org_fhir ON mv_endpoint_list_organizations(fhir_version);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_endpoint_list_org_fhir." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "DROP INDEX IF EXISTS idx_mv_endpoint_list_org_url;" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to drop idx_mv_endpoint_list_org_url." >> $log_file
}

docker exec -t lantern-back-end_postgres_1 psql -t -c "CREATE INDEX idx_mv_endpoint_list_org_url ON mv_endpoint_list_organizations(url);" -U lantern -d lantern || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to create idx_mv_endpoint_list_org_url." >> $log_file

echo "$(date +"%Y-%m-%d %H:%M:%S") - done." >> $log_file