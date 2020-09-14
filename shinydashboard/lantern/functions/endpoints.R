# Functions to compute metrics on endpoints
library(purrr)

# Package that makes it easier to work with dates and times for getting avg response times
library(lubridate)


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
    left_join(endpoint_export_tbl %>%
        distinct(url, vendor_name, fhir_version, tls_version, mime_types, http_response, supported_resources), by = c("url" = "url")) %>%
        mutate(updated = as.Date(updated_at)) %>%
        select(url, organization_names, updated, vendor_name, fhir_version, tls_version, mime_types, http_response, supported_resources) %>%
        left_join(app$http_response_code_tbl %>% select(code, label),
          by = c("http_response" = "code")) %>%
          mutate(status = paste(http_response, "-", label)) %>%
          distinct(url, .keep_all = TRUE)
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
    "http_non200" = max((curr_tally %>% filter(http_response != 200)) %>% pull(n), 0),
    "http_404" = max((curr_tally %>% filter(http_response == 404)) %>% pull(n), 0),
    "http_503" = max((curr_tally %>% filter(http_response == 503)) %>% pull(n), 0)
  )
}

# get the date of the most recently updated fhir_endpoint
get_endpoint_last_updated <- function(db_tables) {
  as.character.Date(app_data$last_updated)
}

# Compute the percentage of each response code for all responses received
get_http_response_summary_tbl <- function(db_tables) {
  db_tables$fhir_endpoints_info_history %>%
    collect() %>%
    left_join(endpoint_export_tbl %>%
      select(url, vendor_name), by = c("url" = "url")) %>%
      select(url, id, http_response, vendor_name) %>%
      mutate(code = as.character(http_response)) %>%
      group_by(id, url, code, http_response, vendor_name) %>%
      summarise(Percentage = n()) %>%
      ungroup() %>%
      group_by(id) %>%
      mutate(Percentage = Percentage / sum(Percentage, na.rm = TRUE) * 100) %>%
      ungroup() %>%
      collect() %>%
      tidyr::replace_na(list(vendor_name = "Unknown"))
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
get_fhir_version_list <- function(endpoint_export_tbl) {
  fhir_version_list <- list(
    "All Versions" = ui_special_values$ALL_FHIR_VERSIONS
  )
  fh <- endpoint_export_tbl %>%
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

# Return list of FHIR Resource Types by endpoint_id, type, fhir_version and vendor
get_fhir_resource_types <- function(db_connection) {
  res <- tbl(db_connection,
    sql("SELECT f.id as endpoint_id,
      vendor_id,
      vendors.name as vendor_name,
      capability_statement->>'fhirVersion' as fhir_version,
      json_array_elements(capability_statement::json#>'{rest,0,resource}') ->> 'type' as type
      from fhir_endpoints_info f
      LEFT JOIN vendors on f.vendor_id = vendors.id
      ORDER BY type")) %>%
    collect() %>%
    tidyr::replace_na(list(vendor_name = "Unknown"))
}

# Return list of FHIR Resources
get_resource_list <- function(endpoint_tbl) {
  rl <- endpoint_tbl %>%
           distinct(type) %>%
           arrange(type) %>%
           split(.$type) %>%
           purrr::map(~ .$type)
  return(rl)
}

get_capstat_fields <- function(db_connection) {
  res <- tbl(db_connection,
    sql("SELECT f.id as endpoint_id,
      vendor_id,
      vendors.name as vendor_name,
      capability_statement->>'fhirVersion' as fhir_version,
      json_array_elements(included_fields::json) ->> 'Field' as field,
      json_array_elements(included_fields::json) ->> 'Exists' as exist,
      json_array_elements(included_fields::json) ->> 'Extension' as extension
      from fhir_endpoints_info f
      LEFT JOIN vendors on f.vendor_id = vendors.id
      WHERE included_fields != 'null'
      ORDER BY field")) %>%
    collect() %>%
    tidyr::replace_na(list(vendor_name = "Unknown"))
}

# Summarize count of resource types by type, fhir_version
get_fhir_resource_count <- function(fhir_resources_tbl) {
  res <- fhir_resources_tbl %>%
    group_by(type, fhir_version) %>%
    count() %>%
    rename(Resource = type, Endpoints = n)
}

get_capstat_fields_count <- function(capstat_fields_tbl, extensionBool) {
  res <- capstat_fields_tbl %>%
    group_by(field, exist, fhir_version, extension) %>%
    count() %>%
    filter(exist == "true") %>%
    filter(extension == extensionBool) %>%
    ungroup() %>%
    select(-exist) %>%
    select(-extension) %>%
    rename(Fields = field, Endpoints = n)
}

get_capstat_fields_list <- function(capstat_fields_tbl) {
  res <- capstat_fields_tbl %>%
    group_by(field) %>%
    filter(extension == "false") %>%
    count() %>%
    select(field)
}

get_capstat_extensions_list <- function(capstat_fields_tbl) {
  res <- capstat_fields_tbl %>%
    group_by(field) %>%
    filter(extension == "true") %>%
    count() %>%
    select(field)
}

get_avg_response_time <- function(db_connection, date) {
  # get time series of response time metrics for all endpoints
  # groups response time averages by 23 hour intervals and shows data for a range of 30 days
  all_endpoints_response_time <- as_tibble(
    tbl(db_connection,
        sql(paste0("SELECT date.datetime AS time, date.average AS avg
                    FROM (SELECT floor(extract(epoch from fhir_endpoints_info_history.entered_at)/82800)*82800 AS datetime, AVG(fhir_endpoints_info_history.response_time_seconds) as average FROM fhir_endpoints_info_history GROUP BY datetime) as date,
                    (SELECT max(floor(extract(epoch from fhir_endpoints_info_history.entered_at)/82800)*82800) AS maximum FROM fhir_endpoints_info_history) as maxdate
                    WHERE date.datetime between (maxdate.maximum-", date, ") AND maxdate.maximum
                    GROUP BY time, average
                    ORDER BY time"))
        )
    ) %>%
    mutate(date = as_datetime(time)) %>%
    select(date, avg)

  # convert to xts format for use in dygraph
  xts(x = all_endpoints_response_time$avg,
      order.by = all_endpoints_response_time$date
  )
}

# get tibble of endpoints which include a security service attribute
# in their capability statement, each service coding as a row
get_security_endpoints <- function(db_connection) {
  res <- tbl(db_connection,
    sql("SELECT
          f.id,
          f.vendor_id,
          v.name,
          capability_statement->>'fhirVersion' as fhir_version,
          json_array_elements(json_array_elements(capability_statement::json#>'{rest,0,security,service}')->'coding')::json->>'code' as code,
          json_array_elements(capability_statement::json#>'{rest,0,security}' -> 'service')::json ->> 'text' as text
        FROM fhir_endpoints_info f, vendors v
        WHERE f.vendor_id = v.id")) %>%
    collect() %>%
    tidyr::replace_na(list(vendor_name = "Unknown")) %>%
    tidyr::replace_na(list(fhir_version = "Unknown"))
}

# get tibble of endpoints which include a security service attribute
# in their capability statement, each service coding as a row
# for display in table of endpoints, with organization name and URL
get_security_endpoints_tbl <- function(db_connection) {
  res <- tbl(db_connection,
    sql("SELECT
          e.url,
          e.organization_names,
          v.name as vendor_name,
          capability_statement->>'fhirVersion' as fhir_version,
          f.tls_version,
          json_array_elements(json_array_elements(capability_statement::json#>'{rest,0,security,service}')->'coding')::json->>'code' as code
        FROM fhir_endpoints_info f, vendors v, fhir_endpoints e
        WHERE f.vendor_id = v.id
        AND e.id = f.id")) %>%
    collect() %>%
    tidyr::replace_na(list(vendor_name = "Unknown")) %>%
    tidyr::replace_na(list(fhir_version = "Unknown"))
}

# Get list of SMART Core Capabilities supported by endpoints returning http 200
get_smart_response_capabilities <- function(db_connection) {
  res <- tbl(db_connection,
    sql("SELECT
      f.id,
      f.smart_http_response,
      v.name as vendor_name,
      f.capability_statement->>'fhirVersion' as fhir_version,
      json_array_elements_text((smart_response->'capabilities')::json) as capability
    FROM fhir_endpoints_info f, vendors v
    WHERE vendor_id = v.id
    AND smart_http_response=200")) %>%
    collect() %>%
    tidyr::replace_na(list(vendor_name = "Unknown")) %>%
    tidyr::replace_na(list(fhir_version = "Unknown"))
}

# Summarize the count of capabilities reported in SMART Core Capabilities JSON doc
get_smart_response_capability_count <- function(endpoints_tbl) {
  res <- endpoints_tbl %>%
    group_by(fhir_version, capability) %>%
    count() %>%
    rename("FHIR Version" = fhir_version, Capability = capability, Endpoints = n)
  res
}

# Query fhir endpoints and return list of endpoints that have
# returned a valid JSON document at /.well-known/smart-configuration
# This implies a smart_http_response of 200.
#
get_well_known_endpoints_tbl <- function(db_connection) {
  res <- tbl(db_connection,
    sql("SELECT e.url, e.organization_names, v.name as vendor_name,
      f.capability_statement->>'fhirVersion' as fhir_version
    FROM fhir_endpoints_info f
    LEFT JOIN vendors v on f.vendor_id = v.id
    LEFT JOIN fhir_endpoints e
    ON f.id = e.id
    WHERE f.smart_http_response = 200
    AND jsonb_typeof(f.smart_response) = 'object'")) %>%
    collect() %>%
    tidyr::replace_na(list(vendor_name = "Unknown")) %>%
    tidyr::replace_na(list(fhir_version = "Unknown"))
}

# get count of well known endpoints returning http 200 (but not
# checking if valid SMART core capability doc returned)
get_well_known_endpoints_count <- function(db_connection) {
  res <- tbl(db_connection,
      sql("SELECT count(*) from fhir_endpoints_info
          WHERE smart_http_response = 200")) %>%
      collect() %>%
      pull(count)
  as.integer(res)
}

# Find any endpoints which have returned a smart_http_response of 200
# at the well known endpoint url /.well-known/smart-configuration
# but did NOT return a valid JSON document when queried
get_well_known_endpoints_no_doc <- function(db_connection) {
  res <- tbl(db_connection,
    sql("SELECT f.id, e.url, f.vendor_id, e.organization_names, v.name as vendor_name,
      f.capability_statement->>'fhirVersion' as fhir_version,
      f.smart_http_response,
      f.smart_response
    FROM fhir_endpoints_info f
    LEFT JOIN vendors v on f.vendor_id = v.id
    LEFT JOIN fhir_endpoints e
    ON f.id = e.id
    WHERE f.smart_http_response = 200
    AND jsonb_typeof(f.smart_response) <> 'object'")) %>%
    collect() %>%
    tidyr::replace_na(list(vendor_name = "Unknown")) %>%
    tidyr::replace_na(list(fhir_version = "Unknown"))
}

# Return a summary table of information about endpoint security statements
get_well_known_endpoint_counts <- function(db_connection) {
  res <- tribble(
    ~Status, ~Endpoints,
    "Total Indexed Endpoints", as.integer(app_data$fhir_endpoint_totals$all_endpoints),
    "Endpoints with successful response (HTTP 200)", as.integer(app_data$response_tally$http_200),
    "Well Known URI Endpoints with successful response (HTTP 200)", get_well_known_endpoints_count(db_connection),
    "Well Known URI Endpoints with valid response JSON document", as.integer(nrow(app_data$well_known_endpoints_tbl)),
    "Well Known URI Endpoints without valid response JSON document", as.integer(nrow(app_data$well_known_endpoints_no_doc))
  )
}

# Get counts of authorization types supported by FHIR Version
get_auth_type_count <- function(security_endpoints) {
  security_endpoints %>%
    group_by(fhir_version) %>%
    mutate(tc = n_distinct(id)) %>%
    group_by(fhir_version, code, tc) %>%
    count(name = "Endpoints") %>%
    mutate(Percent = percent(Endpoints / tc))  %>%
    ungroup() %>%
    select("Code" = code, "FHIR Version" = fhir_version, Endpoints, Percent)
}

# Get count of endpoints which have NOT returned a valid capability statement
get_no_cap_statement_count <- function(db_connection) {
  res <- tbl(db_connection,
             sql("select count(*) from fhir_endpoints_info where jsonb_typeof(capability_statement) <> 'object'")
  ) %>% pull(count)
}

# Return a summary table of information about endpoint security statements
get_endpoint_security_counts <- function(db_connection) {
  res <- tribble(
    ~Status, ~Endpoints,
    "Total Indexed Endpoints", as.integer(app_data$fhir_endpoint_totals$all_endpoints),
    "Endpoints with successful response (HTTP 200)", as.integer(app_data$response_tally$http_200),
    "Endpoints with unsuccessful response", as.integer(app_data$response_tally$http_non200),
    "Endpoints without valid capability statement", as.integer(get_no_cap_statement_count(db_connection)),
    "Endpoints with valid security resource", as.integer(nrow(app_data$security_endpoints %>% distinct(id)))
  )
}

get_organization_locations <- function(db_connection) {
  res <- tbl(db_connection,
      sql("SELECT id, name, left(location->>'zipcode',5) as zipcode from npi_organizations")
  ) %>%
    collect() %>%
    left_join(app$zip_to_zcta, by = c("zipcode" = "zipcode")) %>%
    filter(!is.na(lng), !is.na(lat))
  res
}

get_endpoint_locations <- function(db_connection) {
  res <- tbl(db_connection,
    sql("SELECT
          distinct(url),
          endpoint_names[1] as endpoint_name,
          organization_name,
          fhir_version,
          vendor_name,
          match_score,
          left(zipcode,5) as zipcode
        FROM endpoint_export where zipcode is NOT NULL AND match_score > .97 ")
    ) %>%
    collect() %>%
    left_join(app$zip_to_zcta, by = c("zipcode" = "zipcode")) %>%
    filter(!is.na(lng), !is.na(lat)) %>%
    tidyr::replace_na(list(vendor_name = "Unknown")) %>%
    tidyr::replace_na(list(fhir_version = "Unknown"))
  res
}
