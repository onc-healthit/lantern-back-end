source("../common/db_connection.R")

# Get endpoints data and transform to csv
get_endpoints_csv_data <- function(db_connection, developer = NULL, fhir_versions = NULL, identifier = NULL, source = NULL) {
  query <- "
    SELECT * 
    FROM selected_fhir_endpoints_mv
    WHERE TRUE"

  params <- list()

  # Developer filter
  if (!is.null(developer)) {
    query <- paste0(query, " AND vendor_name = {developer}")
    params$developer <- developer
  }

  # FHIR version filter
  if (!is.null(fhir_versions)) {
    query <- paste0(query, " AND fhir_version IN ({fhir_versions*})")
    params$fhir_versions <- fhir_versions
  }

  # Source (is_chpl) filter
  if (!is.null(source)) {
    query <- paste0(query, " AND is_chpl = {source}")
    params$source <- source
  }

  # Finalize
  query <- paste0(query, "
    ORDER BY vendor_name, list_source, url, requested_fhir_version")

  # Build SQL safely
  if (length(params) > 0) {
    sql_query <- do.call(glue::glue_sql, c(list(query, .con = db_connection), params))
  } else {
    sql_query <- glue::glue_sql(query, .con = db_connection)
  }

  df <- DBI::dbGetQuery(db_connection, sql_query)

  # Format for csv
  df <- df %>%
    select(-id, -status, -availability, -fhir_version, -urlModal, -condensed_endpoint_names) %>%
    rowwise() %>%
    mutate(endpoint_names = ifelse(length(strsplit(endpoint_names, ";")[[1]]) > 100, paste0("Subset of Organizations, see Lantern Website for full list:", paste0(head(strsplit(endpoint_names, ";")[[1]], 100), collapse = ";")), endpoint_names),
            info_created = format(info_created, "%m/%d/%y %H:%M"),
            info_updated = format(info_updated, "%m/%d/%y %H:%M"),
            list_source = ifelse(list_source %in% c("1up (Gainwell)", "Acentra", "CNSI Provider One", 
                  "Conduent", "Edifecs", "Not Available", "Safhir from Onyx",
                  "Salesforce/MiHIN", "State Developed"), 
                  "State Medicaid Agency (SMA) Provider Directory", 
                  list_source)) %>%
    ungroup() %>%
    rename(api_information_source_name = endpoint_names, api_developer_name = vendor_name) %>%
    rename(created_at = info_created, updated = info_updated) %>%
    rename(http_response_time_second = response_time_seconds) %>%
    rename(source = is_chpl)

  return(df)
}

# Get organization data and transform to csv
get_organization_csv_data <- function(db_connection, developer = NULL, fhir_versions = NULL, identifier = NULL, organization_detail = NULL, source = NULL) {
  query <- "
    WITH base_data AS (
      SELECT
        organization_name,
        identifier_types_csv as identifier_type,
        identifier_values_csv as identifier_value,
        addresses_csv as address,
        endpoint_urls_csv as url,
        fhir_versions_array,
        vendor_names_array,
        is_chpl_array
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

  # Source (is_chpl) filter
  if (!is.null(source)) {
    query <- paste0(query, " AND is_chpl_array && ARRAY[{source}]")
    params$source <- source
  }

  # Identifier filter
  if (!is.null(identifier)) {
    query <- paste0(query, " AND {identifier_exact} = ANY(string_to_array(identifier_values_csv, E'\\n'))")
    params$identifier_exact <- paste0(identifier)
  }

  # Organization Detail filter
  if (!is.null(organization_detail) && organization_detail == "present") {
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
      string_agg(DISTINCT vendor_name, E'\\n') AS api_developer_name,
      string_agg(DISTINCT source_value, E'\\n') AS source
    FROM base_data bd
    CROSS JOIN LATERAL unnest(bd.fhir_versions_array) AS fhir_version
    CROSS JOIN LATERAL unnest(bd.vendor_names_array) AS vendor_name
    CROSS JOIN LATERAL unnest(bd.is_chpl_array) AS source_value
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

  # Re-apply source filter to unnest output
  if (!is.null(source)) {
    query <- paste0(query, " AND source_value = {source_display}")
    params$source_display <- source
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
      address = ifelse(is.na(address), "", address),
      source = ifelse(is.na(source), "", source)
    )

  return(df)
}