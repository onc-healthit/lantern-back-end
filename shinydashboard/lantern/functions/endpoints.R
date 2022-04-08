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
  indexed <- db_tables$fhir_endpoints_info %>% filter(requested_fhir_version == "None") %>% distinct(url) %>% count() %>% pull(n)
  fhir_endpoint_totals <- list(
    "all_endpoints"     = all,
    "indexed_endpoints" = indexed,
    "nonindexed_endpoints" = max(all - indexed, 0)
  )
}

# create a join to get more detailed table of fhir_endpoint information
get_fhir_endpoints_tbl <- function() {
  ret_tbl <- endpoint_export_tbl %>%
    distinct(url, vendor_name, fhir_version, http_response, requested_fhir_version, .keep_all = TRUE) %>%
    select(url, endpoint_names, info_created, info_updated, list_source, vendor_name, capability_fhir_version, fhir_version, format, http_response, response_time_seconds, smart_http_response, errors, availability, cap_stat_exists, kind) %>%
    left_join(app$http_response_code_tbl %>% select(code, label),
      by = c("http_response" = "code")) %>%
      mutate(status = if_else(http_response == 200, paste("Success:", http_response, "-", label), paste("Failure:", http_response, "-", label))) %>%
      mutate(cap_stat_exists = tolower(as.character(cap_stat_exists))) %>%
      mutate(cap_stat_exists = case_when(
        kind != "instance" ~ "true*",
        TRUE ~ cap_stat_exists
      ))
}

# get the endpoint tally by http_response received
get_response_tally_list <- function(db_tables) {
  curr_tally <- db_tables$fhir_endpoints_info %>%
    filter(requested_fhir_version == "None") %>%
    select(metadata_id) %>%
    left_join(db_tables$fhir_endpoints_metadata %>% select(http_response, id),
      by = c("metadata_id" = "id")) %>%
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
  as.character.Date(isolate(app_data$last_updated()))
}

# Compute the percentage of each response code for all responses received
get_http_response_summary_tbl <- function(db_tables) {
  db_tables$fhir_endpoints_info %>%
    collect() %>%
    filter(requested_fhir_version == "None") %>%
    left_join(endpoint_export_tbl %>%
      select(url, vendor_name, http_response, fhir_version), by = c("url" = "url")) %>%
      select(url, id, http_response, vendor_name, fhir_version) %>%
      mutate(code = as.character(http_response)) %>%
      group_by(id, url, code, http_response, vendor_name, fhir_version) %>%
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

get_distinct_fhir_version_list_no_capstat <- function(endpoint_export_tbl) {
  res <- endpoint_export_tbl %>%
  distinct(fhir_version) %>%
  split(.$fhir_version) %>%
  purrr::map(~ .$fhir_version)
}

get_distinct_fhir_version_list <- function(endpoint_export_tbl) {
  res <- endpoint_export_tbl %>%
  filter(fhir_version != "No Cap Stat") %>%
  distinct(fhir_version) %>%
  split(.$fhir_version) %>%
  purrr::map(~ .$fhir_version)
}

# Get the list of distinct fhir versions for use in filtering
get_fhir_version_list <- function(endpoint_export_tbl, no_cap_stat) {
  fhir_version_list <- list()

  res <- endpoint_export_tbl %>%
  distinct(fhir_version)

  res <- res %>% mutate(fhir_version_name = case_when(
  fhir_version %in% dstu2 ~ "DSTU2",
  fhir_version %in% stu3 ~ "STU3",
  fhir_version %in% r4 ~ "R4",
  fhir_version == "Unknown" ~ "Unknown",
  TRUE ~ "No Cap Stat"
  ))

  dstu2Vals <- res %>%
    filter(fhir_version_name == "DSTU2") %>%
    select(fhir_version) %>%
    split(.$fhir_version) %>%
    purrr::map(~ .$fhir_version)

  stu3Vals <- res %>%
    filter(fhir_version_name == "STU3") %>%
    select(fhir_version) %>%
    split(.$fhir_version) %>%
    purrr::map(~ .$fhir_version)

  r4Vals <- res %>%
    filter(fhir_version_name == "R4") %>%
    select(fhir_version) %>%
    split(.$fhir_version) %>%
    purrr::map(~ .$fhir_version)

  unknownVals <- res %>%
    filter(fhir_version_name == "Unknown") %>%
    select(fhir_version) %>%
    split(.$fhir_version) %>%
    purrr::map(~ .$fhir_version)

  noVals <- res %>%
    filter(fhir_version_name == "No Cap Stat") %>%
    select(fhir_version) %>%
    split(.$fhir_version) %>%
    purrr::map(~ .$fhir_version)

  if (length(dstu2Vals) > 0) {
    dstu2List <- list("DSTU2" = dstu2Vals)
    fhir_version_list <- c(fhir_version_list, dstu2List)
  }

  if (length(stu3Vals) > 0) {
    stu3List <- list("STU3" = stu3Vals)
    fhir_version_list <- c(fhir_version_list, stu3List)
  }

  if (length(r4Vals) > 0) {
    r4List <- list("R4" = r4Vals)
    fhir_version_list <- c(fhir_version_list, r4List)
  }

  if (length(unknownVals) > 0) {
    if (length(noVals) > 0 && no_cap_stat == TRUE) {
      otherList <- list("Other" = c(unknownVals, noVals))
      fhir_version_list <- c(fhir_version_list, otherList)
    } else {
      otherList <- list("Other" = unknownVals)
      fhir_version_list <- c(fhir_version_list, otherList)
    }
  } else if (length(noVals) > 0 && no_cap_stat == TRUE) {
      otherList <- list("Other" = noVals)
      fhir_version_list <- c(fhir_version_list, otherList)
  }

  fhir_version_list
}

# Get the list of distinct vendor names for use in filtering
get_vendor_list <- function(endpoint_export_tbl) {
  vendor_list <- list(
    "All Developers" = ui_special_values$ALL_DEVELOPERS
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
      capability_fhir_version as fhir_version,
      json_array_elements(capability_statement::json#>'{rest,0,resource}') ->> 'type' as type
      from fhir_endpoints_info f
      LEFT JOIN vendors on f.vendor_id = vendors.id
      WHERE requested_fhir_version = 'None'
      ORDER BY type")) %>%
    collect() %>%
    tidyr::replace_na(list(vendor_name = "Unknown")) %>%
    mutate(fhir_version = if_else(fhir_version == "", "No Cap Stat", fhir_version)) %>%
    mutate(fhir_version = if_else(grepl("-", fhir_version, fixed = TRUE), sub("-.*", "", fhir_version), fhir_version)) %>%
    mutate(fhir_version = if_else(fhir_version %in% valid_fhir_versions, fhir_version, "Unknown"))
}

# Return list of resources for the given operation field by
# endpoint_id, vendor and fhir_version
get_fhir_resource_by_op <- function(db_connection, field) {
  res <- tbl(db_connection,
    sql(paste0("SELECT f.id as endpoint_id,
      vendor_id,
      vendors.name as vendor_name,
      capability_fhir_version as fhir_version,
      operation_resource->>'", field, "' as type
      from fhir_endpoints_info f
      LEFT JOIN vendors on f.vendor_id = vendors.id
      WHERE requested_fhir_version = 'None'"))) %>%
    collect() %>%
    tidyr::replace_na(list(vendor_name = "Unknown")) %>%
    mutate(fhir_version = if_else(fhir_version == "", "No Cap Stat", fhir_version)) %>%
    mutate(fhir_version = if_else(grepl("-", fhir_version, fixed = TRUE), sub("-.*", "", fhir_version), fhir_version)) %>%
    mutate(fhir_version = if_else(fhir_version %in% valid_fhir_versions, fhir_version, "Unknown"))
}

get_capstat_fields <- function(db_connection) {
  res <- tbl(db_connection,
    sql("SELECT f.id as endpoint_id,
      vendor_id,
      vendors.name as vendor_name,
      capability_fhir_version as fhir_version,
      json_array_elements(included_fields::json) ->> 'Field' as field,
      json_array_elements(included_fields::json) ->> 'Exists' as exist,
      json_array_elements(included_fields::json) ->> 'Extension' as extension
      from fhir_endpoints_info f
      LEFT JOIN vendors on f.vendor_id = vendors.id
      WHERE included_fields != 'null' AND requested_fhir_version = 'None'
      ORDER BY field")) %>%
    collect() %>%
    tidyr::replace_na(list(vendor_name = "Unknown")) %>%
    mutate(fhir_version = if_else(grepl("-", fhir_version, fixed = TRUE), sub("-.*", "", fhir_version), fhir_version)) %>%
    mutate(fhir_version = if_else(fhir_version %in% valid_fhir_versions, fhir_version, "Unknown"))
}

get_supported_profiles <- function(db_connection) {
  res <- tbl(db_connection,
    sql("SELECT f.id as endpoint_id,
      f.url,
      vendor_id,
      vendors.name as vendor_name,
      capability_fhir_version as fhir_version,
      json_array_elements(supported_profiles::json) ->> 'Resource' as resource,
      json_array_elements(supported_profiles::json) ->> 'ProfileURL' as profileurl,
      json_array_elements(supported_profiles::json) ->> 'ProfileName' as profilename
      from fhir_endpoints_info f
      LEFT JOIN vendors on f.vendor_id = vendors.id
      WHERE supported_profiles != 'null' AND requested_fhir_version = 'None'")) %>%
    collect() %>%
    tidyr::replace_na(list(vendor_name = "Unknown")) %>%
    mutate(fhir_version = if_else(fhir_version == "", "No Cap Stat", fhir_version)) %>%
    mutate(fhir_version = if_else(grepl("-", fhir_version, fixed = TRUE), sub("-.*", "", fhir_version), fhir_version)) %>%
    mutate(fhir_version = if_else(fhir_version %in% valid_fhir_versions, fhir_version, "Unknown"))
}

# Summarize count of implementation guides by implementation_guide, fhir_version
get_implementation_guide_count <- function(fhir_resources_tbl) {
  res <- fhir_resources_tbl %>%
    group_by(implementation_guide, fhir_version) %>%
    filter(implementation_guide != "None") %>%
    count() %>%
    rename(Implementation = implementation_guide, Endpoints = n)
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

# get values from specific fields we're interested in displaying
# get two fhir version fields, one for fhir version filter and one for field filter
# this is necessary when choosing fhir version as the field value as the selected field’s column gets renamed to field_value when selected
get_capstat_values <- function(db_connection) {
  res <- tbl(db_connection,
    sql("SELECT f.id as endpoint_id,
      vendor_id,
      vendors.name as vendor_name,
      capability_fhir_version as fhir_version,
      capability_fhir_version as filter_fhir_version,
      capability_statement->>'url' as url,
      capability_statement->>'version' as version,
      capability_statement->>'name' as name,
      capability_statement->>'title' as title,
      capability_statement->>'date' as date,
      capability_statement->>'publisher' as publisher,
      capability_statement->>'description' as description,
      capability_statement->>'purpose' as purpose,
      capability_statement->>'copyright' as copyright,
      capability_statement->'software'->>'name' as software_name,
      capability_statement->'software'->>'version' as software_version,
      capability_statement->'software'->>'releaseDate' as software_release_date,
      capability_statement->'implementation'->>'description' as implementation_description,
      capability_statement->'implementation'->>'url' as implementation_url,
      capability_statement->'implementation'->>'custodian' as implementation_custodian
      from fhir_endpoints_info f
      LEFT JOIN vendors on f.vendor_id = vendors.id
      WHERE capability_statement::jsonb != 'null' AND requested_fhir_version = 'None'")) %>%
    collect() %>%
    tidyr::replace_na(list(vendor_name = "Unknown")) %>%
    mutate(fhir_version = if_else(fhir_version == "", "No Cap Stat", fhir_version)) %>%
    mutate(filter_fhir_version = if_else(grepl("-", filter_fhir_version, fixed = TRUE), sub("-.*", "", filter_fhir_version), filter_fhir_version)) %>%
    mutate(filter_fhir_version = if_else(filter_fhir_version %in% valid_fhir_versions, filter_fhir_version, "Unknown"))
}

get_capstat_values_list <- function(capstat_values_tbl) {
  res <- capstat_values_tbl
}

get_avg_response_time <- function(db_connection, date) {
  # get time series of response time metrics for all endpoints
  # groups response time averages by 23 hour intervals and shows data for a range of 30 days
  all_endpoints_response_time <- as_tibble(
    tbl(db_connection,
        sql(paste0("SELECT date.datetime AS time, date.average AS avg, date.maximum AS max, date.minimum AS min
                    FROM (SELECT floor(extract(epoch from updated_at)/", qry_interval_seconds, ")*", qry_interval_seconds, " AS datetime, ROUND(AVG(response_time_seconds), 4) as average, MAX(response_time_seconds) as maximum, MIN(response_time_seconds) as minimum FROM fhir_endpoints_metadata WHERE response_time_seconds > 0 AND requested_fhir_version = 'None' GROUP BY datetime) as date,
                    (SELECT max(floor(extract(epoch from updated_at)/", qry_interval_seconds, ")*", qry_interval_seconds, ") AS maximum FROM fhir_endpoints_metadata WHERE requested_fhir_version = 'None') as maxdate
                    WHERE date.datetime between (maxdate.maximum-", date, ") AND maxdate.maximum
                    GROUP BY time, average, date.maximum, date.minimum
                    ORDER BY time"))
        )
    ) %>%
    mutate(date = as_datetime(time)) %>%
    select(date, avg, max, min)
}

# get tibble of endpoints which include a security service attribute
# in their capability statement, each service coding as a row
get_security_endpoints <- function(db_connection) {
  res <- tbl(db_connection,
    sql("SELECT
          f.id,
          f.vendor_id,
          v.name,
          capability_fhir_version as fhir_version,
          json_array_elements(json_array_elements(capability_statement::json#>'{rest,0,security,service}')->'coding')::json->>'code' as code,
          json_array_elements(capability_statement::json#>'{rest,0,security}' -> 'service')::json ->> 'text' as text
        FROM fhir_endpoints_info f LEFT JOIN vendors v
        ON f.vendor_id = v.id
        WHERE requested_fhir_version = 'None'")) %>%
    collect() %>%
    tidyr::replace_na(list(vendor_name = "Unknown")) %>%
    mutate(fhir_version = if_else(fhir_version == "", "No Cap Stat", fhir_version)) %>%
    mutate(fhir_version = if_else(grepl("-", fhir_version, fixed = TRUE), sub("-.*", "", fhir_version), fhir_version)) %>%
    mutate(fhir_version = if_else(fhir_version %in% valid_fhir_versions, fhir_version, "Unknown"))

}

# get tibble of endpoints which include a security service attribute
# in their capability statement, each service coding as a row
# for display in table of endpoints, with organization name and URL
get_security_endpoints_tbl <- function(db_connection) {
  res <- tbl(db_connection,
    sql("SELECT a.url,
            a.organization_names,
            b.vendor_name,
            a.capability_fhir_version,
            a.tls_version,
            a.code
        FROM
          (SELECT e.url,
            e.organization_names,
            capability_fhir_version as capability_fhir_version,
            f.tls_version,
            f.vendor_id,
            json_array_elements(json_array_elements(capability_statement::json#>'{rest,0,security,service}')->'coding')::json->>'code' as code
          FROM fhir_endpoints_info f,fhir_endpoints e
          WHERE e.url = f.url AND requested_fhir_version = 'None') a
        LEFT JOIN (SELECT v.name as vendor_name, v.id FROM vendors v) b
        ON a.vendor_id = b.id")) %>%
    collect() %>%
    tidyr::replace_na(list(vendor_name = "Unknown")) %>%
    mutate(capability_fhir_version = if_else(capability_fhir_version == "", "No Cap Stat", capability_fhir_version)) %>%
    mutate(fhir_version = if_else(grepl("-", capability_fhir_version, fixed = TRUE), sub("-.*", "", capability_fhir_version), capability_fhir_version)) %>%
    mutate(fhir_version = if_else(fhir_version %in% valid_fhir_versions, fhir_version, "Unknown"))
}

# Get list of SMART Core Capabilities supported by endpoints returning http 200
get_smart_response_capabilities <- function(db_connection) {
  res <- tbl(db_connection,
    sql("SELECT
      f.id,
      m.smart_http_response,
      v.name as vendor_name,
      f.capability_fhir_version as fhir_version,
      json_array_elements_text((smart_response->'capabilities')::json) as capability
    FROM fhir_endpoints_info f
    LEFT JOIN vendors v ON f.vendor_id = v.id
    LEFT JOIN fhir_endpoints_metadata m on f.metadata_id = m.id
    WHERE vendor_id = v.id AND f.metadata_id = m.id AND f.requested_fhir_version = 'None'
    AND m.smart_http_response=200")) %>%
    collect() %>%
    tidyr::replace_na(list(vendor_name = "Unknown")) %>%
    mutate(fhir_version = if_else(fhir_version == "", "No Cap Stat", fhir_version)) %>%
    mutate(fhir_version = if_else(grepl("-", fhir_version, fixed = TRUE), sub("-.*", "", fhir_version), fhir_version)) %>%
    mutate(fhir_version = if_else(fhir_version %in% valid_fhir_versions, fhir_version, "Unknown"))
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
      f.capability_fhir_version as capability_fhir_version
    FROM fhir_endpoints_info f
    LEFT JOIN fhir_endpoints_metadata m on f.metadata_id = m.id
    LEFT JOIN vendors v on f.vendor_id = v.id
    LEFT JOIN fhir_endpoints e
    ON f.url = e.url
    WHERE m.smart_http_response = 200 AND f.requested_fhir_version = 'None'
    AND jsonb_typeof(f.smart_response::jsonb) = 'object'")) %>%
    collect() %>%
    tidyr::replace_na(list(vendor_name = "Unknown")) %>%
    mutate(capability_fhir_version = if_else(capability_fhir_version == "", "No Cap Stat", capability_fhir_version)) %>%
    mutate(fhir_version = if_else(grepl("-", capability_fhir_version, fixed = TRUE), sub("-.*", "", capability_fhir_version), capability_fhir_version)) %>%
    mutate(fhir_version = if_else(fhir_version %in% valid_fhir_versions, fhir_version, "Unknown"))
}

# Find any endpoints which have returned a smart_http_response of 200
# at the well known endpoint url /.well-known/smart-configuration
# but did NOT return a valid JSON document when queried
get_well_known_endpoints_no_doc <- function(db_connection) {
  res <- tbl(db_connection,
    sql("SELECT f.id, e.url, f.vendor_id, e.organization_names, v.name as vendor_name,
      f.capability_fhir_version as fhir_version,
      m.smart_http_response,
      f.smart_response
    FROM fhir_endpoints_info f
    LEFT JOIN fhir_endpoints_metadata m on f.metadata_id = m.id
    LEFT JOIN vendors v on f.vendor_id = v.id
    LEFT JOIN fhir_endpoints e
    ON f.url = e.url
    WHERE m.smart_http_response = 200 AND f.requested_fhir_version = 'None'
    AND jsonb_typeof(f.smart_response::jsonb) <> 'object'")) %>%
    collect() %>%
    tidyr::replace_na(list(vendor_name = "Unknown")) %>%
    mutate(fhir_version = if_else(fhir_version == "", "No Cap Stat", fhir_version)) %>%
    mutate(fhir_version = if_else(grepl("-", fhir_version, fixed = TRUE), sub("-.*", "", fhir_version), fhir_version)) %>%
    mutate(fhir_version = if_else(fhir_version %in% valid_fhir_versions, fhir_version, "Unknown"))
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
             sql("select count(*) from fhir_endpoints_info where jsonb_typeof(capability_statement::jsonb) <> 'object' AND requested_fhir_version = 'None'")
  ) %>% pull(count)
}

# Return a summary table of information about endpoint security statements
get_endpoint_security_counts <- function(db_connection) {
  res <- tribble(
    ~Status, ~Endpoints,
    "Total Indexed Endpoints", as.integer(isolate(app_data$fhir_endpoint_totals()$all_endpoints)),
    "Endpoints with successful response (HTTP 200)", as.integer(isolate(app_data$response_tally()$http_200)),
    "Endpoints with unsuccessful response", as.integer(isolate(app_data$response_tally()$http_non200)),
    "Endpoints without valid CapabilityStatement / Conformance Resource", as.integer(get_no_cap_statement_count(db_connection)),
    "Endpoints with valid security resource", as.integer(nrow(isolate(app_data$security_endpoints()) %>% distinct(id)))
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
    mutate(fhir_version = if_else(fhir_version == "", "No Cap Stat", fhir_version)) %>%
    mutate(fhir_version = if_else(grepl("-", fhir_version, fixed = TRUE), sub("-.*", "", fhir_version), fhir_version)) %>%
    mutate(fhir_version = if_else(fhir_version %in% valid_fhir_versions, fhir_version, "Unknown"))
  res
}
# get implementation guides stored in capability statement
get_implementation_guide <- function(db_connection) {
  res <- tbl(db_connection,
    sql("SELECT
          f.url as url,
          capability_fhir_version as fhir_version,
          json_array_elements(capability_statement::json#>'{implementationGuide}') as implementation_guide,
          vendors.name as vendor_name
          FROM fhir_endpoints_info f
          LEFT JOIN vendors on f.vendor_id = vendors.id
          WHERE requested_fhir_version = 'None'")) %>%
    collect() %>%
    tidyr::replace_na(list(vendor_name = "Unknown")) %>%
    mutate(fhir_version = if_else(fhir_version == "", "No Cap Stat", fhir_version)) %>%
    tidyr::replace_na(list(implementation_guide = "None")) %>%
    mutate(fhir_version = if_else(grepl("-", fhir_version, fixed = TRUE), sub("-.*", "", fhir_version), fhir_version)) %>%
    mutate(fhir_version = if_else(fhir_version %in% valid_fhir_versions, fhir_version, "Unknown"))
}

get_cap_stat_sizes <- function(db_connection) {
  res <- tbl(db_connection,
    sql("SELECT
          f.url as url,
          pg_column_size(capability_statement::text) as size,
          capability_fhir_version as fhir_version,
          vendors.name as vendor_name
          FROM fhir_endpoints_info f
          LEFT JOIN vendors on f.vendor_id = vendors.id WHERE capability_fhir_version != ''
          AND requested_fhir_version = 'None'")) %>%
    collect() %>%
    tidyr::replace_na(list(vendor_name = "Unknown")) %>%
    mutate(fhir_version = if_else(fhir_version == "", "No Cap Stat", fhir_version)) %>%
    mutate(fhir_version = if_else(grepl("-", fhir_version, fixed = TRUE), sub("-.*", "", fhir_version), fhir_version)) %>%
    mutate(fhir_version = if_else(fhir_version %in% valid_fhir_versions, fhir_version, "Unknown"))
}

get_validation_results <- function(db_connection) {
  res <- tbl(db_connection,
    sql("SELECT vendors.name as vendor_name,
          f.url as url,
          capability_fhir_version as fhir_version,
          rule_name,
          valid,
          expected,
          actual,
          comment,
          reference,
          validations.validation_result_id as id
        FROM fhir_endpoints_info f
          LEFT JOIN vendors on f.vendor_id = vendors.id
          INNER JOIN validations on f.validation_result_id = validations.validation_result_id
        ORDER BY validations.validation_result_id, rule_name")) %>%
    collect() %>%
    tidyr::replace_na(list(vendor_name = "Unknown")) %>%
    mutate(fhir_version = if_else(fhir_version == "", "No Cap Stat", fhir_version)) %>%
    mutate(fhir_version = if_else(grepl("-", fhir_version, fixed = TRUE), sub("-.*", "", fhir_version), fhir_version)) %>%
    mutate(fhir_version = if_else(fhir_version %in% valid_fhir_versions, fhir_version, "Unknown"))
}

get_endpoint_list_matches <- function() {
    el <- endpoint_export_tbl %>%
          unnest(endpoint_names) %>%
          select(url, endpoint_names, fhir_version, vendor_name) %>%
          rename(organization_name = endpoint_names) %>%
          tidyr::replace_na(list(organization_name = "Unknown"))
    el
}

get_npi_organization_matches <- function() {
  nl <- endpoint_export_tbl %>%
          select(url, organization_name, organization_secondary_name, npi_id, fhir_version, vendor_name, match_score, zipcode) %>%
          mutate(match_score = match_score*100)  %>%
          filter(match_score >= 97) %>%
          tidyr::replace_na(list(organization_name = "Unknown", organization_secondary_name = "Unknown", npi_id = "Unknown", zipcode = "Unknown")) %>%
          mutate(organization_secondary_name = if_else(organization_secondary_name == "", "Unknown", organization_secondary_name))
  nl
}

database_fetcher <- reactive({
  app_data$fhir_endpoint_totals(get_endpoint_totals_list(db_tables))

  app_data$response_tally(get_response_tally_list(db_tables))

  app_data$http_pct(get_http_response_summary_tbl(db_tables))

  app_data$vendor_count_tbl(get_fhir_version_vendor_count(endpoint_export_tbl))

  app_data$endpoint_resource_types(get_fhir_resource_types(db_connection))

  app_data$capstat_fields(get_capstat_fields(db_connection))

  app_data$supported_profiles(get_supported_profiles(db_connection))

  app_data$capstat_values(get_capstat_values(db_connection))

  app_data$last_updated(now("UTC"))

  app_data$security_endpoints(get_security_endpoints(db_connection))

  app_data$security_endpoints_tbl(get_security_endpoints_tbl(db_connection))

  app_data$auth_type_counts(get_auth_type_count(isolate(app_data$security_endpoints())))

  app_data$security_code_list(isolate(app_data$security_endpoints()) %>%
    distinct(code) %>%
    pull(code))

  app_data$smart_response_capabilities(get_smart_response_capabilities(db_connection))

  app_data$well_known_endpoints_tbl(get_well_known_endpoints_tbl(db_connection))

  app_data$well_known_endpoints_no_doc(get_well_known_endpoints_no_doc(db_connection))

  app_data$endpoint_security_counts(get_endpoint_security_counts(db_connection))

  app_data$implementation_guide(get_implementation_guide(db_connection))

  app_data$endpoint_locations(get_endpoint_locations(db_connection))

  app_data$capstat_sizes_tbl(get_cap_stat_sizes(db_connection))

  app_data$validation_tbl(get_validation_results(db_connection))

})
