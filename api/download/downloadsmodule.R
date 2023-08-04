source("../common/db_connection.R")

get_endpoint_organizations <- function(db_connection) {
    res <- tbl(db_connection,
    sql("SELECT DISTINCT url, UNNEST(endpoint_names) as endpoint_names_list FROM endpoint_export ORDER BY endpoint_names_list")) %>%
    collect() %>%
    group_by(url) %>%
    summarise(endpoint_names_list = list(endpoint_names_list))
    res
}

# Get the Endpoint export table and clean up for UI
get_endpoint_export_tbl <- function(db_tables) {

endpoint_organization_tbl <- get_endpoint_organizations(db_connection)

endpoint_export_tbl <- db_tables$endpoint_export %>%
  collect() %>%
  mutate(vendor_name = na_if(vendor_name, "")) %>%
  tidyr::replace_na(list(vendor_name = "Unknown")) %>%
  mutate(fhir_version = if_else(fhir_version == "", "No Cap Stat", fhir_version)) %>%
  rename(capability_fhir_version = fhir_version) %>%
  mutate(fhir_version = if_else(grepl("-", capability_fhir_version, fixed = TRUE), sub("-.*", "", capability_fhir_version), capability_fhir_version)) %>%
  mutate(fhir_version = if_else(fhir_version %in% valid_fhir_versions, fhir_version, "Unknown")) %>%
  left_join(endpoint_organization_tbl) %>%
  mutate(endpoint_names = gsub("^c\\(|\\)$", "", as.character(endpoint_names_list))) %>%
  mutate(endpoint_names = gsub("(\", )", "\"; ", as.character(endpoint_names))) %>%
  mutate(endpoint_names = gsub("NULL", "", as.character(endpoint_names))) %>%
  mutate(endpoint_names = gsub("(\")", "", as.character(endpoint_names))) %>%
  mutate(format = gsub("(\"|\"|\\[|\\])", "", as.character(format)))
  endpoint_export_tbl
}

get_fhir_endpoints_tbl <- function(db_tables) {
  ret_tbl <- get_endpoint_export_tbl(db_tables) %>%
    distinct(url, vendor_name, fhir_version, http_response, requested_fhir_version, .keep_all = TRUE) %>%
    select(url, endpoint_names, info_created, info_updated, list_source, vendor_name, capability_fhir_version, fhir_version, format, http_response, response_time_seconds, smart_http_response, errors, availability, cap_stat_exists, kind, requested_fhir_version, is_chpl) %>%
    left_join(http_response_code_tbl() %>% select(code, label),
      by = c("http_response" = "code")) %>%
      mutate(status = if_else(http_response == 200, paste("Success:", http_response, "-", label), paste("Failure:", http_response, "-", label))) %>%
      mutate(cap_stat_exists = tolower(as.character(cap_stat_exists))) %>%
      mutate(cap_stat_exists = case_when(
        kind != "instance" ~ "true*",
        TRUE ~ cap_stat_exists
      ))
 }

# Downloadable csv of selected dataset
download_data <- function(db_tables) {
  csvdata <- get_fhir_endpoints_tbl(db_tables) %>%
    select(-label, -status, -availability, -fhir_version) %>%
    rowwise() %>%
    mutate(endpoint_names = ifelse(length(strsplit(endpoint_names, ";")[[1]]) > 100, paste0("Subset of Organizations, see Lantern Website for full list:", paste0(head(strsplit(endpoint_names, ";")[[1]], 100), collapse = ";")), endpoint_names)) %>%
    rename(api_information_source_name = endpoint_names, certified_api_developer_name = vendor_name) %>%
    rename(created_at = info_created, updated = info_updated) %>%
    rename(http_response_time_second = response_time_seconds)
  write.csv(csvdata, file, row.names = FALSE)
}