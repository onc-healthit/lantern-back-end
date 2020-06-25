# Functions to compute metrics on endpoints
library(purrr)

# Will need scalable solution for creating short names from Vendor names for UI
vendor_short_names <- data.frame(
  vendor_name = c("Allscripts", "CareEvolution, Inc.", "Cerner Corporation", "Epic Systems Corporation", "Medical Information Technology, Inc. (MEDITECH)", "Unknown"),
  short_name = c("Allscripts", "CareEvolution", "Cerner", "Epic", "MEDITECH", "Unknown"),
  stringsAsFactors = FALSE)

# Get Endpoint Totals
# Return list of counts of:
# - all registered endpoints
# - indexed endpoints that have been queried
# - non-indexed endpoints yet to be queried
get_endpoint_totals_list <- function(db_tables) {
  all <- db_tables$fhir_endpoints %>% distinct(url) %>% count() %>% pull(n)
  indexed <- db_tables$fhir_endpoints_info %>% distinct(url) %>% count() %>% pull(n)
  fhir_endpoint_totals <- list(
    "all_endpoints"     = all,
    "indexed_endpoints" = indexed,
    "nonindexed_endpoints" = max(all - indexed, 0)
  )
}

# create a join to get more detailed table of fhir_endpoint information
get_fhir_endpoints_tbl <- function(db_tables) {
  db_tables$fhir_endpoints %>%
    collect() %>%
    distinct(url, .keep_all=TRUE) %>%
    left_join(endpoint_export_tbl %>%
          distinct(url, vendor_name, fhir_version, tls_version, mime_types, http_response),
        by = c("url" = "url")) %>%
    mutate(updated = as.Date(updated_at)) %>%
    select(url, organization_names, updated, vendor_name, fhir_version, tls_version, mime_types, http_response) %>%
    left_join(http_response_code_tbl %>% select(code, label),
              by = c("http_response" = "code")) %>%
    mutate(status = paste(http_response, "-", label))
}

# get the endpoint tally by http_response received
get_response_tally_list <- function(db_tables) {
  curr_tally <- db_tables$fhir_endpoints_info %>%
    select(http_response) %>%
    group_by(http_response) %>%
    tally()

  # Get the list of most recent HTTP responses when requesting the capability statement from the
  # fhir_endpoints
  list(
    "http_200" = max((curr_tally %>% filter(http_response == 200)) %>% pull(n), 0),
    "http_404" = max((curr_tally %>% filter(http_response == 404)) %>% pull(n), 0),
    "http_503" = max((curr_tally %>% filter(http_response == 503)) %>% pull(n), 0)
  )
}

# get the date of the most recently updated fhir_endpoint
get_endpoint_last_updated <- function(db_tables) {
  as.character.Date(db_tables$fhir_endpoints_info %>% arrange(desc(updated_at)) %>% head(1) %>% pull(updated_at))
}

# Compute the percentage of each response code for all responses received
get_http_response_summary_tbl <- function(db_tables) {
  db_tables$fhir_endpoints_info_history %>%
    select(id, http_response) %>%
    mutate(code = as.character(http_response)) %>%
    group_by(id, code, http_response) %>%
    summarise(Percentage = n()) %>%
    ungroup() %>%
    group_by(id) %>%
    mutate(Percentage = Percentage / sum(Percentage, na.rm = TRUE) * 100) %>%
    ungroup() %>%
    collect()
}

# Get the count of endpoints by vendor
get_fhir_version_vendor_count <- function(endpoint_tbl) {
  endpoint_tbl %>%
    distinct(vendor_name, url, fhir_version) %>%
    group_by(vendor_name, fhir_version) %>%
    tally() %>%
    ungroup() %>%
    select(vendor_name, fhir_version, n) %>%
    left_join(vendor_short_names)
}

get_fhir_version_factors <- function(endpoint_tbl) {
    mutate(endpoint_tbl,
           vendor_f = as.factor(vendor_name),
           fhir_f = as.factor(fhir_version)
    )
}

# Get the list of distinct fhir versions for use in filtering
get_fhir_version_list <- function(endpoint_tbl) {
  fhir_version_list <- list(
    "All Versions" = ui_special_values$ALL_FHIR_VERSIONS
  )
  fh <- endpoint_tbl %>%
    distinct(fhir_version) %>%
    split(.$fhir_version) %>%
    purrr::map(~ .$fhir_version)
  fhir_version_list <- c(fhir_version_list, fh)
}

# Get the list of distinct vendor names for use in filtering
get_vendor_list <- function(endpoint_export_tbl) {
  vendor_list <- list(
    "All Vendors" = ui_special_values$ALL_VENDORS
  )

  vl <- endpoint_export_tbl %>%
           distinct(vendor_name) %>%
           arrange(vendor_name) %>%
           split(.$vendor_name) %>%
           purrr::map(~ .$vendor_name)

  vendor_list <- c(vendor_list, vl)
}

get_fhir_resources_tbl <- function(db_tables) {
  fei <- db_tables$fhir_endpoints_info %>% collect() %>% head(100)
  res <- fei %>%
    purrr::pmap_dfr(function(...) {
      current <- tibble(...)
      cs <- jsonlite::fromJSON(fei %>% filter(id==current$id) %>% .$capability_statement) 
      if (!is.null(cs)) {
        resources <- purrr::pluck(cs,"rest","resource",1)
        type_df <- as_tibble(resources$type) %>% rename(type=value)
        type_df$endpoint_id <- current$id
        type_df$vendor_id <- current$vendor_id
        type_df$fhir_version <- cs$fhirVersion
        type_df
      } else {
        NULL
      }
    })
}

