source("../common/db_connection.R")

# create a join to get more detailed table of fhir_endpoint information
get_fhir_endpoints_tbl <- function() {
  res <- tbl(db_connection,
    sql("SELECT url, endpoint_names, info_created, info_updated, list_source, 
                vendor_name, capability_fhir_version, fhir_version, format, 
                http_response, response_time_seconds, smart_http_response, errors, 
                availability, cap_stat_exists, kind, 
                requested_fhir_version, is_chpl, status 
         FROM fhir_endpoint_comb_mv")) %>%
    collect()
  
  res
}

# Downloadable csv of selected dataset
download_data <- function() {
  csvdata <- get_fhir_endpoints_tbl() %>%
      select(-status, -availability, -fhir_version) %>%
      rowwise() %>%
      mutate(endpoint_names = ifelse(length(strsplit(endpoint_names, ";")[[1]]) > 100, paste0("Subset of Organizations, see Lantern Website for full list:", paste0(head(strsplit(endpoint_names, ";")[[1]], 100), collapse = ";")), endpoint_names)) %>%
      rename(api_information_source_name = endpoint_names, certified_api_developer_name = vendor_name) %>%
      rename(created_at = info_created, updated = info_updated) %>%
      rename(http_response_time_second = response_time_seconds)
}