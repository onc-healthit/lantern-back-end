# Functions to compute metrics on endpoints
library(purrr)

# Package that makes it easier to work with dates and times for getting avg response times
library(lubridate)

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

get_endpoint_organizations <- function(db_connection) {
    res <- tbl(db_connection,
    sql("SELECT DISTINCT url, UNNEST(endpoint_names) as endpoint_names_list FROM endpoint_export ORDER BY endpoint_names_list")) %>%
    collect() %>%
    group_by(url) %>%
    summarise(endpoint_names_list = list(endpoint_names_list))
    res
}

get_endpoint_organization_list <- function(endpoint) {
    res <- tbl(db_connection,
    sql(paste0("SELECT url, UNNEST(endpoint_names) as endpoint_names_list FROM endpoint_export WHERE url = '", endpoint, "' ORDER BY endpoint_names_list"))) %>%
    collect() %>%
    group_by(url) %>%
    summarise(endpoint_names_list = list(endpoint_names_list)) %>%
    mutate(endpoint_names_list = gsub("^c\\(|\\)$", "", endpoint_names_list)) %>%
    mutate(endpoint_names_list = gsub("(\", )", "\";", as.character(endpoint_names_list))) %>%
    mutate(endpoint_names_list = gsub("\"", "", endpoint_names_list))

    res$endpoint_names_list
}

# Will need scalable solution for creating short names from Vendor names for UI
vendor_short_names <- data.frame(
  vendor_name = c("Allscripts", "CareEvolution, Inc.", "Cerner Corporation", "Epic Systems Corporation", "Medical Information Technology, Inc. (MEDITECH)", "Microsoft Corporation", "Unknown"),
  short_name = c("Allscripts", "CareEvolution", "Cerner", "Epic", "MEDITECH", "Microsoft", "Unknown"),
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
  res <- tbl(db_connection,
    sql("SELECT url, endpoint_names, info_created, info_updated, list_source, 
                vendor_name, capability_fhir_version, fhir_version, format, 
                http_response, response_time_seconds, smart_http_response, errors, 
                kind, availability, requested_fhir_version, is_chpl,
                cap_stat_exists, status 
         FROM fhir_endpoint_comb_mv
         ORDER BY vendor_name, list_source, url, requested_fhir_version")) %>%
    collect()
  
  res
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
    left_join(endpoint_export_tbl() %>%
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
  tbl <- endpoint_tbl %>%
    distinct(vendor_name, url, fhir_version) %>%
    group_by(vendor_name, fhir_version) %>%
    tally() %>%
    ungroup() %>%
    select(vendor_name, fhir_version, n) %>%
    left_join(vendor_short_names) %>%
    mutate(short_name = ifelse(is.na(short_name), vendor_name, short_name))

    tbl
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

get_endpoint_resource_by_op <- function(db_connection, endpointURL, requestedFhirVersion, field) {
  res <- tbl(db_connection,
    sql(paste0("SELECT
      jsonb_array_elements_text(operation_resource->'", field, "') as type
      from fhir_endpoints_info
      WHERE url = '", endpointURL, "' AND requested_fhir_version = '", requestedFhirVersion, "'"))) %>%
    collect()
  res
}

get_endpoint_resources <- function(db_connection, endpointURL, requestedFhirVersion) {
  res <- tbl(db_connection,
    sql(paste0("SELECT jsonb_object_keys(operation_resource::jsonb) as operations
         FROM fhir_endpoints_info WHERE url = '", endpointURL, "' AND requested_fhir_version = '", requestedFhirVersion, "'"
    ))
  ) %>%
  collect()

  op_list <- as.list(res$operations)
  table <- data.frame(matrix(ncol = 2, nrow = 0))
  colnames(table) <- c("Operation", "Resource")

  if (length(op_list) > 0) {
    for (op in op_list) {
      resources <- isolate(get_endpoint_resource_by_op(db_connection, endpointURL, requestedFhirVersion, op))
      newTable <- data.frame("Operation" = c(op), "Resource" = c(resources$type))
      table <- rbind(table, newTable)
    }
  }
  table
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

get_endpoint_capstat_fields <- function(db_connection, endpointURL, requestedFhirVersion, extensionBool) {
  res <- tbl(db_connection,
    sql(paste0("SELECT
      url,
      json_array_elements(included_fields::json) ->> 'Field' as field,
      json_array_elements(included_fields::json) ->> 'Exists' as exist,
      json_array_elements(included_fields::json) ->> 'Extension' as extension
      from fhir_endpoints_info f
      WHERE url = '", endpointURL, "' AND requested_fhir_version = '", requestedFhirVersion, "'"
    ))
  ) %>%
    collect() %>%
    filter(extension == extensionBool) %>%
    select(field, exist)
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

get_endpoint_supported_profiles <- function(db_connection, endpointURL, requestedFhirVersion) {
    res <- tbl(db_connection,
    sql(paste0("SELECT
      json_array_elements(supported_profiles::json) ->> 'ProfileURL' as profileurl,
      json_array_elements(supported_profiles::json) ->> 'ProfileName' as profilename,
      json_array_elements(supported_profiles::json) ->> 'Resource' as resource
      from fhir_endpoints_info f
      WHERE supported_profiles != 'null' AND url = '", endpointURL, "' AND requested_fhir_version = '", requestedFhirVersion, "'"))) %>%
    collect()

    res
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

# get contact information
get_contact_information <- function(db_connection) {

  contacts_tbl <- tbl(db_connection,
    sql("SELECT DISTINCT
				  url,
				  json_array_elements((capability_statement->>'contact')::json)->>'name' as contact_name,
        	json_array_elements((json_array_elements((capability_statement->>'contact')::json)->>'telecom')::json)->>'system' as contact_type,
          json_array_elements((json_array_elements((capability_statement->>'contact')::json)->>'telecom')::json)->>'value' as contact_value,
          json_array_elements((json_array_elements((capability_statement->>'contact')::json)->>'telecom')::json)->>'rank' as contact_preference
          FROM fhir_endpoints_info
          WHERE capability_statement::jsonb != 'null' AND requested_fhir_version = 'None'")) %>%
    collect()


    res <- endpoint_export_tbl() %>%
        distinct(url, vendor_name, fhir_version, endpoint_names, .keep_all = TRUE) %>%
        select(url, vendor_name, fhir_version, endpoint_names, requested_fhir_version) %>%
        filter(requested_fhir_version == "None") %>%
        left_join(contacts_tbl, by = c("url" = "url"))

    res
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

get_endpoint_response_time <- function(db_connection, date, endpointURL, requestedFhirVersion) {
  # get time series of response time metrics for all endpoints
  # groups response time averages by 23 hour intervals and shows data for a range of 30 days
  all_endpoints_response_time <- as_tibble(
    tbl(db_connection,
        sql(paste0("SELECT date.datetime AS time, response_time_seconds as response
                    FROM (SELECT floor(extract(epoch from updated_at)/", qry_interval_seconds, ")*", qry_interval_seconds, " AS datetime, response_time_seconds FROM fhir_endpoints_metadata WHERE response_time_seconds > 0 AND url = '", endpointURL, "' AND requested_fhir_version = '", requestedFhirVersion, "') as date,
                    (SELECT max(floor(extract(epoch from updated_at)/", qry_interval_seconds, ")*", qry_interval_seconds, ") AS maximum FROM fhir_endpoints_metadata WHERE url = '", endpointURL, "' AND requested_fhir_version = '", requestedFhirVersion, "') as maxdate
                    WHERE date.datetime between (maxdate.maximum-", date, ") AND maxdate.maximum
                    ORDER BY time"))
        )
    ) %>%
    mutate(date = as_datetime(time)) %>%
    select(date, response)
}


get_endpoint_http_over_time <- function(db_connection, date, endpointURL, requestedFhirVersion) {
  endpoint_http_over_time <- as_tibble(
    tbl(db_connection,
        sql(paste0("SELECT http_responses.http_response AS http_response, http_responses.datetime AS time
                    FROM (SELECT http_response, floor(extract(epoch from updated_at)) AS datetime FROM fhir_endpoints_metadata WHERE url = '", endpointURL, "' AND requested_fhir_version = '", requestedFhirVersion, "') as http_responses,
                    (SELECT max(floor(extract(epoch from updated_at))) AS maximum FROM fhir_endpoints_metadata WHERE url = '", endpointURL, "' AND requested_fhir_version = '", requestedFhirVersion, "') as maxdate
                    WHERE http_responses.datetime between (maxdate.maximum-", date, ") AND maxdate.maximum
                    ORDER BY time"))
        )
    ) %>%
    mutate(date = as_datetime(time)) %>%
    select(date, http_response)
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
            a.endpoint_names as organization_names,
            a.vendor_name,
            a.capability_fhir_version,
            a.tls_version,
            a.code
        FROM
          (SELECT e.url,
            e.endpoint_names,
            e.fhir_version as capability_fhir_version,
            e.tls_version,
            e.vendor_name,
            json_array_elements(json_array_elements(f.capability_statement::json#>'{rest,0,security,service}')->'coding')::json->>'code' as code
          FROM endpoint_export e,fhir_endpoints_info f
          WHERE e.url = f.url AND f.requested_fhir_version = 'None') a")) %>%
    collect() %>%
    tidyr::replace_na(list(vendor_name = "Unknown")) %>%
    mutate(capability_fhir_version = if_else(capability_fhir_version == "", "No Cap Stat", capability_fhir_version)) %>%
    mutate(fhir_version = if_else(grepl("-", capability_fhir_version, fixed = TRUE), sub("-.*", "", capability_fhir_version), capability_fhir_version)) %>%
    mutate(fhir_version = if_else(fhir_version %in% valid_fhir_versions, fhir_version, "Unknown")) %>%
    mutate(organization_names = gsub("(\\{|\\})", "", as.character(organization_names))) %>%
    mutate(organization_names = gsub("(\",\")", "; ", as.character(organization_names))) %>%
    mutate(organization_names = gsub("(\")", "", as.character(organization_names)))
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

get_endpoint_smart_response_capabilities <- function(db_connection, endpointURL, requestedFhirVersion) {
  res <- tbl(db_connection,
    sql(paste0("SELECT
      json_array_elements_text((smart_response->'capabilities')::json) as capability
    FROM fhir_endpoints_info f
    LEFT JOIN fhir_endpoints_metadata m on f.metadata_id = m.id
    WHERE f.metadata_id = m.id AND f.url = '", endpointURL, "' AND f.requested_fhir_version = '", requestedFhirVersion, "'
    AND m.smart_http_response=200"))) %>%
    collect()
  res
}

get_endpoint_products <- function(db_connection, endpointURL, requestedFhirVersion) {
  res <- tbl(db_connection,
    sql(paste0("SELECT
        f.url, h.name, h.version, h.api_url, h.certification_status, h.certification_date, h.certification_edition,
        h.chpl_id, h.last_modified_in_chpl  FROM fhir_endpoints_info f, healthit_products h, healthit_products_map hm WHERE f.healthit_mapping_id = hm.id AND
        hm.healthit_product_id = h.id AND f.healthit_mapping_id IS NOT NULL AND f.url = '", endpointURL, "' AND f.requested_fhir_version = '", requestedFhirVersion, "'"))) %>%
        collect() %>%
    select(name, version, chpl_id, api_url, certification_status, certification_edition, certification_date, last_modified_in_chpl)
  res
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
    sql("SELECT e.url, e.endpoint_names as organization_names, e.vendor_name,
      e.fhir_version as capability_fhir_version
    FROM endpoint_export e
    LEFT JOIN fhir_endpoints_info f
    LEFT JOIN fhir_endpoints_metadata m on f.metadata_id = m.id
    LEFT JOIN vendors v on f.vendor_id = v.id
    ON e.url = f.url
    WHERE m.smart_http_response = 200 AND f.requested_fhir_version = 'None'
    AND jsonb_typeof(f.smart_response::jsonb) = 'object'")) %>%
    collect() %>%
    tidyr::replace_na(list(vendor_name = "Unknown")) %>%
    mutate(capability_fhir_version = if_else(capability_fhir_version == "", "No Cap Stat", capability_fhir_version)) %>%
    mutate(fhir_version = if_else(grepl("-", capability_fhir_version, fixed = TRUE), sub("-.*", "", capability_fhir_version), capability_fhir_version)) %>%
    mutate(fhir_version = if_else(fhir_version %in% valid_fhir_versions, fhir_version, "Unknown")) %>%
    mutate(organization_names = gsub("(\\{|\\})", "", as.character(organization_names))) %>%
    mutate(organization_names = gsub("(\",\")", "; ", as.character(organization_names))) %>%
    mutate(organization_names = gsub("(\")", "", as.character(organization_names)))
}

# Find any endpoints which have returned a smart_http_response of 200
# at the well known endpoint url /.well-known/smart-configuration
# but did NOT return a valid JSON document when queried
get_well_known_endpoints_no_doc <- function(db_connection) {
  res <- tbl(db_connection,
    sql("SELECT f.id, e.url, f.vendor_id, e.endpoint_names as organization_names, e.vendor_name,
      e.fhir_version,
      m.smart_http_response,
      f.smart_response
    FROM endpoint_export e
    LEFT JOIN fhir_endpoints_info f
    LEFT JOIN fhir_endpoints_metadata m on f.metadata_id = m.id
    ON e.url = f.url
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
    left_join(app$zip_to_zcta(), by = c("zipcode" = "zipcode")) %>%
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
          left(zipcode,5) as zipcode,
          npi_id
        FROM organization_location")
    ) %>%
    collect() %>%
    left_join(app$zip_to_zcta(), by = c("zipcode" = "zipcode")) %>%
    filter(!is.na(lng), !is.na(lat)) %>%
    tidyr::replace_na(list(vendor_name = "Unknown")) %>%
    mutate(fhir_version = if_else(fhir_version == "", "No Cap Stat", fhir_version)) %>%
    mutate(fhir_version = if_else(grepl("-", fhir_version, fixed = TRUE), sub("-.*", "", fhir_version), fhir_version)) %>%
    mutate(fhir_version = if_else(fhir_version %in% valid_fhir_versions, fhir_version, "Unknown"))
  res
}

get_single_endpoint_locations <- function(db_connection, endpointURL, requestedFhirVersion) {
  res <- tbl(db_connection,
    sql(paste0("SELECT
          url,
          organization_name,
          npi_id,
          match_score,
          left(zipcode,5) as zipcode
        FROM organization_location where url = '", endpointURL, "' AND requested_fhir_version = '", requestedFhirVersion, "'"))
    ) %>%
    collect() %>%
    left_join(app$zip_to_zcta(), by = c("zipcode" = "zipcode")) %>%
    filter(!is.na(lng), !is.na(lat)) %>%
    distinct(organization_name, match_score, zipcode, lat, lng, npi_id)
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

get_endpoint_implementation_guide <- function(db_connection, endpointURL, requestedFhirVersion) {
  res <- tbl(db_connection,
    sql(paste0("SELECT
          json_array_elements(capability_statement::json#>'{implementationGuide}') as implementation_guide
          FROM fhir_endpoints_info f
          WHERE url = '", endpointURL, "' AND requested_fhir_version = '", requestedFhirVersion, "'"))) %>%
    collect()

  res
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
          validations.validation_result_id as id,
          requested_fhir_version
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
    el <- endpoint_export_tbl() %>%
          separate_rows(endpoint_names, sep = ";") %>%
          select(url, endpoint_names, fhir_version, vendor_name, requested_fhir_version) %>%
          rename(organization_name = endpoint_names) %>%
          tidyr::replace_na(list(organization_name = "Unknown")) %>%
          mutate(organization_name = if_else(organization_name == "", "Unknown", organization_name))
    el
}


get_capability_and_smart_response <- function(db_connection, endpointURL, requestedFhirVersion) {
  res <- tbl(db_connection,
    sql(paste0("SELECT capability_statement, smart_response FROM fhir_endpoints_info WHERE
          url = '", endpointURL, "' AND requested_fhir_version = '", requestedFhirVersion, "'"))
   ) %>%
    collect()
  res

}

get_details_page_metrics <- function(endpointURL, requestedFhirVersion) {
  res <- endpoint_export_tbl() %>%
    filter(url == endpointURL) %>%
    filter(requested_fhir_version == requestedFhirVersion) %>%
    distinct(url, http_response, smart_http_response, errors, cap_stat_exists, availability) %>%
    mutate(status = if_else(http_response == 200, "ACTIVE", "INACTIVE")) %>%
    mutate(errors = if_else(errors == "", "None", errors)) %>%
    mutate(availability = availability * 100) %>%
    left_join(app$http_response_code_tbl() %>% select(code, label),
          by = c("http_response" = "code")) %>%
      mutate(http_response = if_else(http_response == 200, paste(http_response, "-", label), paste(http_response, "-", label))) %>%
      left_join(app$http_response_code_tbl() %>% select(code, label),
          by = c("smart_http_response" = "code")) %>%
          mutate(smart_http_response = if_else(smart_http_response == 200, paste(smart_http_response, "-", label.y), paste(smart_http_response, "-", label.y)))
  res

}

get_details_page_info <- function(endpointURL, requestedFhirVersion, db_connection) {
    res <- endpoint_export_tbl() %>%
          filter(url == endpointURL) %>%
          filter(requested_fhir_version == requestedFhirVersion) %>%
          distinct(url, fhir_version, vendor_name, software_name, software_version, software_releasedate, format, info_created, info_updated)

    resListSource <- endpoint_export_tbl() %>%
          filter(url == endpointURL) %>%
          filter(requested_fhir_version == requestedFhirVersion) %>%
          distinct(list_source)

    resSecurity <-  tbl(db_connection,
        sql(paste0("SELECT
            json_array_elements(json_array_elements(capability_statement::json#>'{rest,0,security,service}')->'coding')::json->>'code' as security
            FROM fhir_endpoints_info
            WHERE url = '", endpointURL, "' AND requested_fhir_version = '", requestedFhirVersion, "'"))) %>%
    collect()

    resSupportedVersions <- tbl(db_connection,
        sql(paste0("SELECT
            DISTINCT versions_response->'Response'->>'versions' as supported_versions, versions_response->'Response'->>'default' as default_version
            FROM fhir_endpoints
            WHERE url = '", endpointURL, "'"))) %>%
    collect() %>%
    mutate(supported_versions = gsub("\\[\"|\"\\]", "", as.character(supported_versions))) %>%
    mutate(default_version = gsub("\"|\"", "", as.character(default_version)))

    res$list_source <- paste0(resListSource$list_source, collapse = "\n")
    res$security <- paste0(resSecurity$security, collapse = ",")
    res$supported_versions <- resSupportedVersions$supported_versions
    res$default_version <- resSupportedVersions$default_version

    res <- res %>%
    mutate(vendor_name = if_else(vendor_name == "Unknown", "Not Available", vendor_name)) %>%
    mutate(fhir_version = if_else(fhir_version == "No Cap Stat", "Not Available", fhir_version)) %>%
    mutate(security = if_else(security == "", "Not Available", security)) %>%
    tidyr::replace_na(list(software_name = "Not Available", software_version = "Not Available", software_releasedate = "Not Available", format = "Not Available", supported_versions = "Not Available", default_version = "Not Available")) %>%
    mutate(software_name = gsub("\"", "", as.character(software_name))) %>%
    mutate(software_version = gsub("\"", "", as.character(software_version))) %>%
    mutate(software_releasedate = gsub("\"", "", as.character(software_releasedate)))

    res
}

