source("downloadsmodule.R")

#* @apiTitle Download Daily FHIR Endpoints report

#* Echo Download Daily FHIR Endpoints report
#* @get /daily/download
#* @param developer Filter by developer name (optional)
#* @param fhir_version Comma-separated list of FHIR versions to filter (optional)
#* @param source Filter by source: 'CHPL', 'State Medicaid', 'Payer', 'Other' (optional)
#* @description Download a csv containing daily endpoint data
function(req, res, developer = NULL, fhir_version = NULL, source = NULL) {
    err <- block_unknown_query_params(
      req, res,
      allowed_params = c("developer", "fhir_version", "source")
    )
    if (!is.null(err)) return(err)

    # Normalize and parse fhir_versions
    fhir_versions_vec <- parse_csv_param(fhir_version)
    
    # Validate Parameters
    fv <- validate_fhir_versions(res, fhir_versions_vec, valid_fhir_versions)
    if (!isTRUE(fv$ok)) return(list(error = fv$error))
    filtered_fhir_versions <- fv$value

    src <- validate_one_of(res, source, c("CHPL", "State Medicaid", "Payer", "Other"), "source")
    if (!isTRUE(src$ok)) return(list(error = src$error))
    source_filter <- src$value

    dv <- validate_developer(res, db_connection, developer)
    if (!isTRUE(dv$ok)) return(list(error = dv$error))

    # Log filters for debugging
    message("Endpoint API Filters - Developer: ", developer, 
          ", FHIR Versions: ", paste(fhir_versions_vec, collapse = ", "),
          ", Source: ", source_filter)

    safe_developer <- sanitize_token(developer)
    safe_fhir <- if (!is.null(filtered_fhir_versions)) {
      paste0("fhir_", sanitize_compact(paste(filtered_fhir_versions, collapse = "_")))
    } else NULL
    safe_source <- if (!is.null(source_filter)) {
      paste0("source_", sanitize_compact(source_filter))
    } else NULL

    st <- format(Sys.time(), "%Y-%m-%d")
    filename <- build_csv_filename(
        prefix = "fhir_endpoints",
        parts = c(safe_developer, safe_fhir, safe_source),
        date_str = st
      )
 
    set_csv_headers(res, filename)

    endpoints_data <- get_endpoints_csv_data(
      db_connection,
      developer = developer,
      fhir_versions = filtered_fhir_versions,
      source = source_filter
    )

    write_and_stream_csv(res, filename, endpoints_data)  
}

#* @get /organizations/v1
#* @param developer Filter by developer name (optional)
#* @param organization_detail Filter by data presence: 'present' or 'absent' (optional)
#* @param identifier Filter by exact identifier value (optional)
#* @param fhir_version Comma-separated list of FHIR versions to filter (optional)
#* @description Download a CSV file containing daily organization data
function(req, res, developer = NULL, organization_detail = NULL, identifier = NULL, fhir_version = NULL) {
  err <- block_unknown_query_params(
    req, res,
    allowed_params = c("developer", "organization_detail", "identifier", "fhir_version")
  )
  if (!is.null(err)) return(err)
  
  # Normalize and parse fhir_versions
  fhir_versions_vec <- parse_csv_param(fhir_version)

  # Validate Parameters
  fv <- validate_fhir_versions(res, fhir_versions_vec, valid_fhir_versions)
  if (!isTRUE(fv$ok)) return(list(error = fv$error))
  filtered_fhir_versions <- fv$value

  od <- validate_organization_detail(res, organization_detail)
  if (!isTRUE(od$ok)) return(list(error = od$error))
  organization_detail_flag <- od$value

  dv <- validate_developer(res, db_connection, developer)
  if (!isTRUE(dv$ok)) return(list(error = dv$error))

  # Log filters for debugging
  message("Organization API Filters - Developer: ", developer, 
        ", Organization Detail: ", organization_detail_flag, 
        ", Identifier: ", identifier, 
        ", FHIR Versions: ", paste(fhir_versions_vec, collapse = ", "))

  safe_developer <- sanitize_token(developer)
  safe_identifier <- if (!is.null(identifier)) paste0("id_", sanitize_compact(identifier)) else NULL
  safe_org_detail <- if (!is.null(organization_detail_flag)) paste0("organization_detail_", organization_detail_flag) else NULL
  safe_fhir <- if (!is.null(filtered_fhir_versions)) {
    paste0("fhir_", sanitize_compact(paste(filtered_fhir_versions, collapse = "_")))
  } else NULL

  st <- format(Sys.time(), "%Y-%m-%d")
  filename <- build_csv_filename(
    prefix = "fhir_endpoint_organizations",
    parts = c(safe_developer, safe_identifier, safe_org_detail, safe_fhir),
    date_str = st
  )

  set_csv_headers(res, filename)

  org_data <- get_organization_csv_data(
    db_connection,
    developer = developer,
    organization_detail = organization_detail_flag,
    identifier = identifier,
    fhir_versions = filtered_fhir_versions
  )
  write_and_stream_csv(res, filename, org_data)
}

# Shared API Helpers

# Parse comma-separated fhir_version query param, trim whitespace
parse_csv_param <- function(x) {
  if (is.null(x)) return(NULL)

  out <- strsplit(x, ",")[[1]]
  out <- trimws(out)

  # remove empty tokens
  out <- out[out != ""]

  if (length(out) == 0) NULL else out
}

# Validate fhir versions against valid_fhir_versions; return vector of valid versions or NULL
validate_fhir_versions <- function(res, fhir_versions_vec, valid_fhir_versions) {
  if (is.null(fhir_versions_vec)) {
      return(list(ok = TRUE, value = NULL, error = NULL))
  }

  valid <- fhir_versions_vec[fhir_versions_vec %in% valid_fhir_versions]
  invalid <- setdiff(fhir_versions_vec, valid)

  if (length(valid) == 0) {
    res$status <- 400
    return(list(
      ok = FALSE,
      value = NULL,
      error = paste0(
        "None of the provided FHIR versions are valid. Accepted values include: ",
        paste(sort(valid_fhir_versions), collapse = ", ")
      )
    ))
  }

  if (length(invalid) > 0) {
    message("Ignoring invalid FHIR versions: ", paste(invalid, collapse = ", "))
  }

  list(ok = TRUE, value = valid, error = NULL)
}

# Validate developer against vendors table
validate_developer <- function(res, db_connection, developer) {
  if (is.null(developer)) return(list(ok = TRUE, error = NULL))

  all_vendors <- tbl(db_connection, "vendors") %>%
    select(name) %>%
    distinct() %>%
    collect() %>%
    pull(name)

  if (!(developer %in% all_vendors)) {
    res$status <- 400
    return(list(
      ok = FALSE,
      error = paste0(
        "Developer '", developer, "' not found in the CHPL-certified vendor list. ",
        "Please check for typos or verify the exact developer name."
      )
    ))
  }

  list(ok = TRUE, error = NULL)
}

# Generic "validate one-of" helper for simple enum params
validate_one_of <- function(res, value, allowed, param_name) {
  if (is.null(value)) return(list(ok = TRUE, value = NULL, error = NULL))

  if (!(value %in% allowed)) {
    res$status <- 400
    return(list(
      ok = FALSE,
      value = NULL,
      error = paste0(
        "Invalid value for '", param_name, "'. Accepted values are: ",
        paste(allowed, collapse = ", ")
      )
    ))
  }

  list(ok = TRUE, value = value, error = NULL)
}

validate_organization_detail <- function(res, organization_detail) {
  if (is.null(organization_detail)) return(list(ok = TRUE, value = NULL, error = NULL))

  od <- tolower(organization_detail)
  if (od != "present") {
    res$status <- 400
    return(list(ok = FALSE, value = NULL,
                error = "Invalid value for 'organization_detail'. Only 'present' is supported."))
  }

  list(ok = TRUE, value = od, error = NULL)
}

# Sanitize helpers for filename parts
sanitize_token <- function(x) {
  if (is.null(x)) return(NULL)
  # replace spaces/special chars with underscores
  gsub("[^A-Za-z0-9_]+", "_", x)
}

sanitize_compact <- function(x) {
  if (is.null(x)) return(NULL)
  # remove non-alphanumeric characters (for fhir versions/ids)
  gsub("[^A-Za-z0-9]", "", x)
}

build_csv_filename <- function(prefix, parts, date_str) {
  filename_parts <- c(prefix, parts, date_str)
  paste0(paste(na.omit(filename_parts), collapse = "_"), ".csv")
}

set_csv_headers <- function(res, filename) {
  res$setHeader("Content-Type", "text/csv")
  res$setHeader("Content-Disposition", paste0("attachment; filename=", filename))
}

write_and_stream_csv <- function(res, filename, data) {
  if (!file.exists(filename)) {
    write.csv(data, file = filename, row.names = FALSE)
  }
  include_file(filename, res, content_type = "text/csv")
}

block_unknown_query_params <- function(req, res, allowed_params) {
  query_names <- names(req$argsQuery)
  query_names <- unique(query_names)

  unknown_params <- setdiff(query_names, allowed_params)

  if (length(unknown_params) > 0) {
    res$status <- 400
    return(list(
      error = paste0(
        "Invalid query parameter(s): ",
        paste(sort(unknown_params), collapse = ", "),
        ". Supported parameters are: ",
        paste(allowed_params, collapse = ", "),
        "."
      )
    ))
  }

  NULL
}