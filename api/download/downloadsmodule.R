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
  #write.csv(csvdata, file, row.names = FALSE)
}

# Get organization data and transform to csv
get_organization_csv_data <- function(db_connection, current_vendor = NULL) {
  res <- get_endpoint_list_matches(db_connection, vendor = current_vendor)

  res <- res %>%
    mutate(organization_id = as.integer(organization_id)) %>%
    
    # Left join with deduplicated or collapsed identifiers
    left_join(
      get_org_identifiers_information(db_connection) %>%
        mutate(org_id = as.integer(org_id)) %>%
        group_by(org_id) %>%
        summarise(identifier = paste(unique(identifier), collapse = "\n")),
      by = c("organization_id" = "org_id")
    ) %>%
    
    # Left join with deduplicated or collapsed addresses
    left_join(
      get_org_addresses_information(db_connection) %>%
        mutate(org_id = as.integer(org_id)) %>%
        group_by(org_id) %>%
        summarise(address = paste(unique(address), collapse = "\n")),
      by = c("organization_id" = "org_id")
    ) %>%
    
    left_join(get_org_url_information(db_connection),
          by = c("organization_id" = "org_id")) %>%
    
    mutate(org_url = if_else(str_starts(org_url, "urn:uuid:"), "", org_url)) %>%
    
    select(-organization_id)

  res <- res %>%
    group_by(organization_name) %>%
    summarise(
      identifier = paste(unique(identifier), collapse = "\n"),
      address = paste(unique(address), collapse = "\n"),
      org_url = paste(unique(org_url), collapse = "\n"),
      fhir_version = paste(unique(fhir_version), collapse = "\n"),
      vendor_name = paste(unique(vendor_name), collapse = "\n"),
      .groups = "drop"
    ) %>%
    filter(organization_name != "Unknown") %>%
    mutate(address = toupper(address)) %>%
    arrange(organization_name)

  res
}

get_endpoint_list_matches <- function(db_connection, fhir_version = NULL, vendor = NULL) {
  # Start with base query
  query <- tbl(db_connection, "mv_endpoint_list_organizations")

  # Apply filters in SQL before collecting data
  if (!is.null(fhir_version) && length(fhir_version) > 0) {
    query <- query %>% filter(fhir_version %in% !!fhir_version)
  }

  if (!is.null(vendor)) {
    query <- query %>% filter(vendor_name == !!vendor)
  }

  # Collect the data after applying filters in SQL
  result <- query %>%
    collect() %>%
    tidyr::replace_na(list(organization_name = "Unknown")) %>%
    mutate(organization_name = if_else(organization_name == "", "Unknown", organization_name))

  return(result)
}

get_org_url_information <- function(db_connection) {

  res <- tbl(db_connection,
    sql("SELECT org_id, org_url FROM fhir_endpoint_organization_url")) %>%
    collect()

    res
}

get_org_identifiers_information <- function(db_connection) {

  res <- tbl(db_connection,"fhir_endpoint_organization_identifiers") %>%
    collect()

    res
}

get_org_addresses_information <- function(db_connection) {

  res <- tbl(db_connection,
    sql("SELECT org_id, address FROM fhir_endpoint_organization_addresses")) %>%
    collect()

    res
}