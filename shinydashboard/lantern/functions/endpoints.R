# Functions to compute metrics on endpoints
library(purrr)

# Package that makes it easier to work with dates and times for getting avg response times # nolint
library(lubridate)

library(glue)

get_endpoint_organization_list <- function(endpoint) {
  splitString <- strsplit(endpoint, "&&")[[1]]
    endpoint_name <- splitString[1]
    vendor_name <- splitString[2]
  
  query <- glue_sql("SELECT url, UNNEST(endpoint_names) as endpoint_names_list FROM endpoint_export WHERE url = {endpoint_name} AND vendor_name = {vendor_name} ORDER BY endpoint_names_list", endpoint_name = endpoint_name, vendor_name = vendor_name, .con = db_connection)
  res <- tbl(db_connection, sql(query)) %>%
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
# Get Endpoint Totals
# Return list of counts of:
# - all registered endpoints
# - indexed endpoints that have been queried
# - non-indexed endpoints yet to be queried
get_endpoint_totals_list <- function(db_tables) {
  # head(1) %>% collect() pushes the row limit down to SQL (LIMIT 1) instead of materializing the
  # whole mv_endpoint_totals view into R via as.data.frame() before keeping only the first row.
  totals_data <- db_tables$mv_endpoint_totals %>%
    head(1) %>%
    collect()
  
  fhir_endpoint_totals <- list(
    "all_endpoints"     = totals_data$all_endpoints,
    "indexed_endpoints" = totals_data$indexed_endpoints,
    "nonindexed_endpoints" = totals_data$nonindexed_endpoints
  )
  
  return(fhir_endpoint_totals)
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


get_endpoint_last_updated <- function(db_tables) {
  last_updated <- db_tables$mv_endpoint_totals %>%
    head(1) %>%
    collect() %>%
    pull(last_updated)
  
  as.character.Date(last_updated)
}


# LANTERN-831
get_http_response_tbl <- function(vendor_name) {
  query <- glue_sql("SELECT http_code, code_label, count_endpoints FROM mv_http_responses WHERE vendor_name = {vendor_name} ORDER BY http_code", .con = db_connection)

  res <- tbl(db_connection,
    sql(query)) %>%
    collect()
    res
}

# LANTERN-831
get_http_response_tbl_all <- function() {
  res <- tbl(db_connection,
    sql("SELECT http_code, code_label, count_endpoints FROM mv_http_responses WHERE vendor_name = 'ALL_DEVELOPERS' ORDER BY http_code")) %>%
    collect()
  res
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
  mutate(fhir_version = normalize_fhir_version(fhir_version)) %>%
  distinct(fhir_version) %>%
  split(.$fhir_version) %>%
  purrr::map(~ .$fhir_version)
}

get_distinct_fhir_version_list <- function(endpoint_export_tbl) {
  res <- endpoint_export_tbl %>%
  filter(fhir_version != "No Cap Stat") %>%
  distinct(fhir_version) %>%
  mutate(fhir_version = normalize_fhir_version(fhir_version)) %>%
  distinct(fhir_version) %>%
  split(.$fhir_version) %>%
  purrr::map(~ .$fhir_version)
}

# Get the list of distinct fhir versions for use in filtering
get_fhir_version_list <- function(endpoint_export_tbl, no_cap_stat) {
  fhir_version_list <- list()

  res <- endpoint_export_tbl %>%
  distinct(fhir_version) %>%
  mutate(fhir_version = normalize_fhir_version(fhir_version)) %>%
  distinct(fhir_version)

  res <- res %>% mutate(fhir_version_name = case_when(
  fhir_version %in% dstu2 ~ "DSTU2",
  fhir_version %in% stu3 ~ "STU3",
  fhir_version %in% r4 ~ "R4",
  fhir_version %in% r4b ~ "R4B",   
  fhir_version %in% r5 ~ "R5",
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

  r4bVals <- res %>%
    filter(fhir_version_name == "R4B") %>%
    select(fhir_version) %>%
    split(.$fhir_version) %>%
    purrr::map(~ .$fhir_version)

  r5Vals <- res %>%
    filter(fhir_version_name == "R5") %>%
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

  if (length(r4bVals) > 0) {
    r4bList <- list("R4B" = r4bVals)
    fhir_version_list <- c(fhir_version_list, r4bList)
  }

  if (length(r5Vals) > 0) {
    r5List <- list("R5" = r5Vals)
    fhir_version_list <- c(fhir_version_list, r5List)
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

# Shared WHERE-clause text builder for get_fhir_resource_by_op() and get_fhir_resource_by_op_count()
# below -- both filter mv_resource_interactions identically, differing only in whether they return
# the grouped (resource_type, fhir_version) rows or a count of them.
build_fhir_resource_by_op_where_clause <- function(operations_vec, search_query) {
  clause <- "WHERE fhir_version IN ({fhir_versions_vec*})
            AND resource_type IN ({resource_types_vec*})"

  # Add a filter for operations if they are selected
  if (length(operations_vec) >= 1) {
    clause <- paste0(clause, " AND operations @> ARRAY[{operations_vec*}]")
  }

  # Always filter on vendor_name: mv_resource_interactions already holds pre-aggregated rows whose
  # vendor_name is 'All Developers' (see dbsetup.sql), so skipping this filter for the "All
  # Developers" selection would sum those rows together with every per-vendor row.
  clause <- paste0(clause, " AND vendor_name = {vendor_name}")

  # Add search filter if present
  if (!is.null(search_query) && search_query != "") {
    clause <- paste0(clause, " AND (resource_type ILIKE {pattern} OR fhir_version ILIKE {pattern})")
  }

  clause
}

# Return the endpoint counts for selected FHIR resources, operations, fhir version and vendor name
get_fhir_resource_by_op <- function(db_connection, operations_vec, fhir_versions_vec, resource_types_vec, vendor_name, page_size = -1, offset = -1, search_query = NULL) {
  # pattern is referenced (via NSE) by build_fhir_resource_by_op_where_clause()'s {pattern}
  # placeholder when glue_sql() below evaluates it in this function's frame.
  pattern <- if (!is.null(search_query) && search_query != "") paste0("%", search_query, "%") else NULL

  query_str <- paste0(
    "SELECT resource_type as type, fhir_version, SUM(endpoint_count) as n
            FROM mv_resource_interactions
            ",
    build_fhir_resource_by_op_where_clause(operations_vec, search_query),
    " GROUP BY (resource_type, fhir_version)
            ORDER BY resource_type, fhir_version"
  )

  if (page_size > -1 && offset > -1) {
    query_str <- paste0(query_str, " LIMIT ", page_size, " OFFSET ", offset)
  }

  query <- glue_sql(query_str, .con = db_connection)

  res <- tbl(db_connection, sql(query)) %>%
  collect()

  res
}

# Count of the grouped (resource_type, fhir_version) rows get_fhir_resource_by_op() above would
# return for the same filters, used for the resource tab's total_pages calculation instead of
# materializing the full grouped result (via page_size = -1, offset = -1) and nrow()-ing it.
get_fhir_resource_by_op_count <- function(db_connection, operations_vec, fhir_versions_vec, resource_types_vec, vendor_name, search_query = NULL) {
  pattern <- if (!is.null(search_query) && search_query != "") paste0("%", search_query, "%") else NULL

  query_str <- paste0(
    "SELECT COUNT(*) as count FROM (SELECT resource_type, fhir_version FROM mv_resource_interactions ",
    build_fhir_resource_by_op_where_clause(operations_vec, search_query),
    " GROUP BY (resource_type, fhir_version)) grouped"
  )

  query <- glue_sql(query_str, .con = db_connection)

  tbl(db_connection, sql(query)) %>% collect() %>% pull(count)
}

# Expands operation_resource (a JSONB map of operation -> array of resource types) into
# Operation/Resource rows in a single query, using jsonb_each() to get the operation keys joined
# with jsonb_array_elements_text() to expand each key's resource array -- instead of one query to
# list the operation keys followed by a separate round trip per operation (N+1) to expand each
# one's resources.
get_endpoint_resources <- function(db_connection, endpointURL, requestedFhirVersion, vendorName) {
  query <- glue_sql("SELECT kv.key AS \"Operation\", elem AS \"Resource\"
         FROM fhir_endpoints_info f, vendors v,
              jsonb_each(f.operation_resource::jsonb) AS kv(key, value),
              jsonb_array_elements_text(kv.value) AS elem
         WHERE f.url = {endpointURL} AND f.requested_fhir_version = {requestedFhirVersion}
         AND v.name = {vendorName} AND v.id = f.vendor_id",
    endpointURL = endpointURL, requestedFhirVersion = requestedFhirVersion, vendorName = vendorName,
    .con = db_connection)
  res <- tbl(db_connection, sql(query)) %>%
  collect()

  res
}

get_capstat_fields <- function(db_connection) {
  res <- tbl(db_connection, sql("SELECT * FROM get_capstat_fields_mv")) %>% 
    collect()
  return(res)
}

# Returns both required/optional fields and extensions in one query (the "extension" column
# distinguishes them) -- callers that need just one or the other should filter this result in R
# rather than calling this function twice, since the underlying query is identical either way and
# the split was previously happening only after collect() anyway.
get_endpoint_capstat_fields <- function(db_connection, endpointURL, requestedFhirVersion, vendorName) {
  query <- glue_sql("SELECT
      f.url,
      json_array_elements(f.included_fields::json) ->> 'Field' as field,
      json_array_elements(f.included_fields::json) ->> 'Exists' as exist,
      json_array_elements(f.included_fields::json) ->> 'Extension' as extension
      from fhir_endpoints_info f, vendors v
      WHERE f.url = {endpointURL} AND f.requested_fhir_version = {requestedFhirVersion}
      AND v.name = {vendorName} AND f.vendor_id = v.id AND included_fields::text <> 'null'",
    endpointURL = endpointURL, requestedFhirVersion = requestedFhirVersion, vendorName = vendorName,
    .con = db_connection)
  res <- tbl(db_connection, sql(query)) %>%
    collect()
}

get_supported_profiles <- function(db_connection) {
  res <- tbl(db_connection, "endpoint_supported_profiles_mv") %>% collect()
}

get_endpoint_supported_profiles <- function(db_connection, endpointURL, requestedFhirVersion, vendorName) {
    query <- glue_sql("SELECT
      json_array_elements(f.supported_profiles::json) ->> 'ProfileURL' as profileurl,
      json_array_elements(f.supported_profiles::json) ->> 'ProfileName' as profilename,
      json_array_elements(f.supported_profiles::json) ->> 'Resource' as resource
      from fhir_endpoints_info f, vendors v
      WHERE f.supported_profiles != 'null' AND f.url = {endpointURL} AND f.requested_fhir_version = {requestedFhirVersion}
      AND v.name = {vendorName} AND f.vendor_id = v.id",
      endpointURL = endpointURL, requestedFhirVersion = requestedFhirVersion, vendorName = vendorName,
      .con = db_connection)
    res <- tbl(db_connection, sql(query)) %>%
    collect()

    res
}

get_org_active_information <- function(db_connection, org_id) {

  res <- tbl(db_connection,
    sql("SELECT org_id, active FROM fhir_endpoint_organization_active")) %>%
    filter(org_id == !!org_id) %>%
    collect()

    res
}

get_org_url_information <- function(db_connection) {

  res <- tbl(db_connection,
    sql("SELECT org_id, org_url FROM fhir_endpoint_organization_url")) %>%
    collect()

    res
}

get_org_identifiers_information <- function(db_connection, org_id) {

  res <- tbl(db_connection,"fhir_endpoint_organization_identifiers") %>%
    filter(org_id == !!org_id) %>%
    collect()

    res
}

get_org_addresses_information <- function(db_connection, org_id) {

  res <- tbl(db_connection,
    sql("SELECT org_id, address FROM fhir_endpoint_organization_addresses")) %>%
    filter(org_id == !!org_id) %>%
    collect()

    res
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
  # `date` is a trusted, app-generated SQL fragment (either a numeric offset or the literal
  # expression "maxdate.maximum" -- see get_range()), so it is interpolated as-is like
  # qry_interval_seconds; endpointURL/requestedFhirVersion come from client-controlled input
  # (current_endpoint()) and are safely quoted rather than pasted directly into the query string.
  url_q <- DBI::dbQuoteString(db_connection, endpointURL)
  version_q <- DBI::dbQuoteString(db_connection, requestedFhirVersion)
  all_endpoints_response_time <- as_tibble(
    tbl(db_connection,
        sql(paste0("SELECT date.datetime AS time, response_time_seconds as response
                    FROM (SELECT floor(extract(epoch from updated_at)/", qry_interval_seconds, ")*", qry_interval_seconds, " AS datetime, response_time_seconds FROM fhir_endpoints_metadata WHERE response_time_seconds > 0 AND url = ", url_q, " AND requested_fhir_version = ", version_q, ") as date,
                    (SELECT max(floor(extract(epoch from updated_at)/", qry_interval_seconds, ")*", qry_interval_seconds, ") AS maximum FROM fhir_endpoints_metadata WHERE url = ", url_q, " AND requested_fhir_version = ", version_q, ") as maxdate
                    WHERE date.datetime between (maxdate.maximum-", date, ") AND maxdate.maximum
                    ORDER BY time"))
        )
    ) %>%
    mutate(date = as_datetime(time)) %>%
    select(date, response)
}


get_endpoint_http_over_time <- function(db_connection, date, endpointURL, requestedFhirVersion) {
  url_q <- DBI::dbQuoteString(db_connection, endpointURL)
  version_q <- DBI::dbQuoteString(db_connection, requestedFhirVersion)
  endpoint_http_over_time <- as_tibble(
    tbl(db_connection,
        sql(paste0("SELECT http_responses.http_response AS http_response, http_responses.datetime AS time
                    FROM (SELECT http_response, floor(extract(epoch from updated_at)) AS datetime FROM fhir_endpoints_metadata WHERE url = ", url_q, " AND requested_fhir_version = ", version_q, ") as http_responses,
                    (SELECT max(floor(extract(epoch from updated_at))) AS maximum FROM fhir_endpoints_metadata WHERE url = ", url_q, " AND requested_fhir_version = ", version_q, ") as maxdate
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
  res <- tbl(db_connection, "mv_get_security_endpoints") %>%
    collect()
  return(res)
}

get_endpoint_smart_response_capabilities <- function(db_connection, endpointURL, requestedFhirVersion, vendorName) {
  query <- glue_sql("SELECT
      json_array_elements_text((f.smart_response->'capabilities')::json) as capability
    FROM fhir_endpoints_info f
    LEFT JOIN fhir_endpoints_metadata m on f.metadata_id = m.id
    LEFT JOIN vendors v on f.vendor_id = v.id
    WHERE f.metadata_id = m.id AND f.url = {endpointURL} AND f.requested_fhir_version = {requestedFhirVersion}
    AND m.smart_http_response=200 AND v.name = {vendorName}",
    endpointURL = endpointURL, requestedFhirVersion = requestedFhirVersion, vendorName = vendorName,
    .con = db_connection)
  res <- tbl(db_connection, sql(query)) %>%
    collect()
  res
}

get_endpoint_products <- function(db_connection, endpointURL, requestedFhirVersion, vendor_name) {
  message("Inside get_endpoint_products")
  query <- glue_sql("SELECT
        f.url, h.name, h.version, h.api_url, h.certification_status, h.certification_date, h.certification_edition,
        h.chpl_id, h.last_modified_in_chpl  FROM fhir_endpoints_info f, healthit_products h, healthit_products_map hm, vendors v WHERE f.healthit_mapping_id = hm.id AND
        hm.healthit_product_id = h.id AND f.healthit_mapping_id IS NOT NULL AND f.url = {endpointURL} AND f.requested_fhir_version = {requestedFhirVersion}
        AND f.vendor_id = v.id AND v.name = {vendor_name}",
    endpointURL = endpointURL, requestedFhirVersion = requestedFhirVersion, vendor_name = vendor_name,
    .con = db_connection)
  res <- tbl(db_connection, sql(query)) %>%
        collect() %>%
    select(name, version, chpl_id, api_url, certification_status, certification_edition, certification_date, last_modified_in_chpl)
  res
}

# Get counts of authorization types supported by FHIR Version
get_auth_type_count <- function(db_connection) {
  res <- tbl(db_connection, "mv_auth_type_count") %>%
    arrange(`FHIR Version`, Code) %>%
    collect()
  return(res)
}

# Get count of endpoints which have NOT returned a valid capability statement
get_no_cap_statement_count <- function(db_connection) {
  res <- tbl(db_connection,
             sql("select count(*) from fhir_endpoints_info where jsonb_typeof(capability_statement::jsonb) <> 'object' AND requested_fhir_version = 'None'")
  ) %>% pull(count)
}

# Return a summary table of information about endpoint security statements
get_endpoint_security_counts <- function(db_connection) {
  # Simply query the materialized view which contains all the pre-calculated data
  res <- tbl(db_connection, "mv_endpoint_security_counts") %>%
    collect()
  return(res)
}

get_response_tally_list <- function(db_tables) {
  response_tally <- db_tables$mv_response_tally %>%
                    head(1) %>%
                    collect()
  
  return(response_tally)
}


get_endpoint_implementation_guide <- function(db_connection, endpointURL, requestedFhirVersion, vendorName) {
  query <- glue_sql("SELECT
          json_array_elements(f.capability_statement::json#>'{{implementationGuide}}') as implementation_guide
          FROM fhir_endpoints_info f, vendors v
          WHERE f.url = {endpointURL} AND f.requested_fhir_version = {requestedFhirVersion}
          AND v.name = {vendorName} AND f.vendor_id = v.id",
    endpointURL = endpointURL, requestedFhirVersion = requestedFhirVersion, vendorName = vendorName,
    .con = db_connection)
  res <- tbl(db_connection, sql(query)) %>%
    collect()

  res
}

get_endpoint_list_matches <- function(db_connection, fhir_version = NULL, vendor = NULL) {
  # Start with base query
  query <- tbl(db_connection, "mv_endpoint_list_organizations")

  # Apply filters in SQL before collecting data
  if (!is.null(fhir_version) && length(fhir_version) > 0) {
    query <- query %>% filter(fhir_version %in% !!fhir_version)
  }

  if (!is.null(vendor) && !is.na(vendor) && vendor != ui_special_values$ALL_DEVELOPERS) {
    query <- query %>% filter(vendor_name == !!vendor)
  }

  # Collect the data after applying filters in SQL
  result <- query %>%
    collect() %>%
    tidyr::replace_na(list(organization_name = "Unknown")) %>%
    mutate(organization_name = if_else(organization_name == "", "Unknown", organization_name))

  return(result)
}


get_capability_and_smart_response <- function(db_connection, endpointURL, requestedFhirVersion) {
  query <- glue_sql("SELECT capability_statement, smart_response FROM fhir_endpoints_info WHERE
          url = {endpointURL} AND requested_fhir_version = {requestedFhirVersion} LIMIT 1",
    endpointURL = endpointURL, requestedFhirVersion = requestedFhirVersion, .con = db_connection)
  res <- tbl(db_connection, sql(query)) %>%
    collect()
  res

}

# The export rows for one endpoint, filtered in SQL instead of in an in-memory copy of the whole
# export. A missing (NA) version or vendor matches no rows, as it did with the in-memory filter.
get_endpoint_export_rows <- function(endpointURL, requestedFhirVersion, vendorName) {
  db_tables$endpoint_export_mv %>%
    filter(url == !!endpointURL,
           requested_fhir_version == !!requestedFhirVersion,
           vendor_name == !!vendorName) %>%
    collect()
}

get_details_page_metrics <- function(endpointURL, requestedFhirVersion, vendorName) {
  res <- get_endpoint_export_rows(endpointURL, requestedFhirVersion, vendorName) %>%
    distinct(url, http_response, smart_http_response, errors, cap_stat_exists, availability) %>%
    mutate(status = if_else(http_response == 200, "ACTIVE", "INACTIVE")) %>%
    mutate(errors = if_else(errors == "", "None", errors)) %>%
    mutate(availability = availability * 100) %>%
    left_join(http_response_codes %>% select(code, label),
          by = c("http_response" = "code")) %>%
      mutate(http_response = if_else(http_response == 200, paste(http_response, "-", label), paste(http_response, "-", label))) %>%
      left_join(http_response_codes %>% select(code, label),
          by = c("smart_http_response" = "code")) %>%
          mutate(smart_http_response = if_else(smart_http_response == 200, paste(smart_http_response, "-", label.y), paste(smart_http_response, "-", label.y)))
  res

}

get_details_page_info <- function(endpointURL, requestedFhirVersion, vendorName, db_connection) {
    export_rows <- get_endpoint_export_rows(endpointURL, requestedFhirVersion, vendorName)

    res <- export_rows %>%
          distinct(url, fhir_version, vendor_name, software_name, software_version, software_releasedate, format, info_created, info_updated)

    resListSource <- export_rows %>%
          distinct(list_source)

    security_query <- glue_sql("SELECT
            json_array_elements(json_array_elements(capability_statement::json#>'{{rest,0,security,service}}')->'coding')::json->>'code' as security
            FROM fhir_endpoints_info
            WHERE url = {endpointURL} AND requested_fhir_version = {requestedFhirVersion} LIMIT 1",
      endpointURL = endpointURL, requestedFhirVersion = requestedFhirVersion, .con = db_connection)
    resSecurity <-  tbl(db_connection, sql(security_query)) %>%
    collect()

    supported_versions_query <- glue_sql("SELECT
            DISTINCT versions_response->'Response'->>'versions' as supported_versions, versions_response->'Response'->>'default' as default_version
            FROM fhir_endpoints
            WHERE url = {endpointURL}",
      endpointURL = endpointURL, .con = db_connection)
    resSupportedVersions <- tbl(db_connection, sql(supported_versions_query)) %>%
    collect() %>%
    mutate(supported_versions = gsub("\\[\"|\"\\]", "", as.character(supported_versions))) %>%
    mutate(default_version = gsub("\"|\"", "", as.character(default_version)))

    res$list_source <- paste0(resListSource$list_source, collapse = "\n")

    # Replace with SMA Provider Directory if it matches specific values
    sma_values <- c("1up (Gainwell)", "Acentra", "CNSI Provider One", 
                    "Conduent", "Edifecs", "Not Available", "Safhir from Onyx",
                    "Salesforce/MiHIN", "State Developed")

    if (res$list_source %in% sma_values) {
      res$list_source <- "State Medicaid Agency (SMA) Provider Directory"
    }

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

# Process-wide cache for lists derived from the endpoint export (dropdown choices). Each entry is
# stamped with get_endpoint_last_updated(), which scripts/refresh_materialized_views.sh changes
# last. A read compares the stamp (one one-row query) and rebuilds only when it has changed, so a
# refresh is picked up on the next use without any scheduled re-fetch. If a rebuild fails, the last
# good value (or `fallback`) is returned so a database error can't blank a filter bar.
shared_data_cache <- new.env(parent = emptyenv())

cached_by_last_updated <- function(name, build, fallback = list()) {
  stamp <- tryCatch(get_endpoint_last_updated(db_tables), error = function(e) NULL)
  if (length(stamp) != 1) {
    stamp <- NULL
  }
  entry <- shared_data_cache[[name]]
  if (!is.null(entry) && (is.null(stamp) || identical(entry$stamp, stamp))) {
    return(entry$value)
  }
  value <- tryCatch(build(), error = function(e) {
    message("Error rebuilding ", name, ": ", conditionMessage(e))
    NULL
  })
  if (is.null(value)) {
    return(if (is.null(entry)) fallback else entry$value)
  }
  if (!is.null(stamp)) {
    shared_data_cache[[name]] <- list(stamp = stamp, value = value)
  }
  value
}

# The distinct (fhir_version, vendor_name) pairs in the endpoint export. This is all the dropdown
# list builders above need, so the whole export is no longer pulled into memory for them.
fetch_export_filter_values <- function() {
  db_tables$endpoint_export_mv %>%
    distinct(fhir_version, vendor_name) %>%
    collect()
}

shared_vendor_list <- function() {
  cached_by_last_updated("vendor_list",
    function() get_vendor_list(fetch_export_filter_values()),
    fallback = list("All Developers" = ui_special_values$ALL_DEVELOPERS))
}

shared_fhir_version_list <- function(no_cap_stat) {
  cached_by_last_updated(paste0("fhir_version_list_", no_cap_stat),
    function() get_fhir_version_list(fetch_export_filter_values(), no_cap_stat))
}

shared_distinct_fhir_version_list <- function(no_cap_stat) {
  cached_by_last_updated(paste0("distinct_fhir_version_list_", no_cap_stat),
    function() {
      if (no_cap_stat) {
        get_distinct_fhir_version_list_no_capstat(fetch_export_filter_values())
      } else {
        get_distinct_fhir_version_list(fetch_export_filter_values())
      }
    })
}

shared_security_code_list <- function() {
  cached_by_last_updated("security_code_list",
    function() {
      tbl(db_connection, "mv_get_security_endpoints") %>% distinct(code) %>% collect() %>% pull(code)
    },
    fallback = character(0))
}
