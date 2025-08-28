source("../common/db_connection.R")

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

# Downloadable csv of selected dataset
download_data <- function() {
  csvdata <- get_fhir_endpoints_tbl() %>%
      select(-status, -availability, -fhir_version) %>%
      rowwise() %>%
      mutate(endpoint_names = ifelse(length(strsplit(endpoint_names, ";")[[1]]) > 100, paste0("Subset of Organizations, see Lantern Website for full list:", paste0(head(strsplit(endpoint_names, ";")[[1]], 100), collapse = ";")), endpoint_names),
             info_created = format(info_created, "%m/%d/%y %H:%M"),
             info_updated = format(info_updated, "%m/%d/%y %H:%M")) %>%
      rename(api_information_source_name = endpoint_names, certified_api_developer_name = vendor_name) %>%
      rename(created_at = info_created, updated = info_updated) %>%
      rename(http_response_time_second = response_time_seconds)
}


# Get organization data and transform to csv
get_organization_csv_data <- function(db_connection, developer = NULL, fhir_versions = NULL, identifier = NULL, hti1 = NULL) {
  query <- "
    WITH base_data AS (
      SELECT
        organization_name,
        identifier_types_html as identifier_type,
        identifier_values_html as identifier_value,
        addresses_csv as address,
        endpoint_urls_csv as url,
        fhir_versions_array,
        vendor_names_array
      FROM mv_organizations_final
      WHERE TRUE"

  params <- list()

  # Developer filter
  if (!is.null(developer)) {
    query <- paste0(query, " AND vendor_names_array && ARRAY[{developer*}]")
    params$developer <- developer
  }

  # FHIR version filter
  if (!is.null(fhir_versions)) {
    query <- paste0(query, " AND fhir_versions_array && ARRAY[{fhir_versions*}]")
    params$fhir_versions <- fhir_versions
  }

  # Identifier filter
  if (!is.null(identifier)) {
    query <- paste0(query, " AND identifier_values_csv = {identifier_exact}")
    params$identifier_exact <- paste0(identifier)
  }

  # HTI-1 filter
  if (!is.null(hti1) && hti1 == "present") {
    query <- paste0(query, " AND ((identifier_values_csv IS NOT NULL AND identifier_values_csv <> '')",
                           " OR (addresses_csv IS NOT NULL AND addresses_csv <> ''))")
  }

  # Continue query with CROSS JOINs and filtering for output
  query <- paste0(query, "
    )
    SELECT
      organization_name,
      identifier_type,
      identifier_value,
      address,
      url AS fhir_endpoint_url,
      string_agg(DISTINCT fhir_version, E'\\n') AS fhir_version,
      string_agg(DISTINCT vendor_name, E'\\n') AS vendor_name
    FROM base_data bd
    CROSS JOIN LATERAL unnest(bd.fhir_versions_array) AS fhir_version
    CROSS JOIN LATERAL unnest(bd.vendor_names_array) AS vendor_name
    WHERE 1=1")

  # Re-apply developer and fhir_version filters to unnest output
  if (!is.null(developer)) {
    query <- paste0(query, " AND vendor_name = ANY(ARRAY[{developer_display*}])")
    params$developer_display <- developer
  }

  if (!is.null(fhir_versions)) {
    query <- paste0(query, " AND fhir_version = ANY(ARRAY[{fhir_versions_display*}])")
    params$fhir_versions_display <- fhir_versions
  }

  # Finalize
  query <- paste0(query, "
    GROUP BY organization_name, identifier_type, identifier_value, address, fhir_endpoint_url
    ORDER BY organization_name")

  # Build SQL safely
  if (length(params) > 0) {
    sql_query <- do.call(glue::glue_sql, c(list(query, .con = db_connection), params))
  } else {
    sql_query <- glue::glue_sql(query, .con = db_connection)
  }

  df <- DBI::dbGetQuery(db_connection, sql_query)

  # Clean output
  df <- df %>%
    mutate(
      identifier_type = ifelse(is.na(identifier_type), "", identifier_type),
      identifier_value = ifelse(is.na(identifier_value), "", identifier_value),
      address = ifelse(is.na(address), "", address)
    )

  return(df)
}