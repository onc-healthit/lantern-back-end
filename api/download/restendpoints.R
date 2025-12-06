source("downloadsmodule.R")

#* @apiTitle Download Daily FHIR Endpoints report

#* Echo Download Daily FHIR Endpoints report
#* @get /daily/download
#* @description Download a csv containing daily endpoint data
function(res) {
    res$setHeader("Content-Type", "text/csv")
    res$setHeader("Content-Disposition", "attachment; filename=fhir_endpoints.csv")
    st <- format(Sys.time(), "%Y-%m-%d")
    filename <- paste("fhir_endpoints_", st, ".csv", sep = "")
    if (!file.exists(filename)) {
        write.csv(download_data(), file=filename, row.names=FALSE)
    }
    include_file(filename, res, content_type = "text/csv")
}

#* @get /organizations/v1
#* @param developer Filter by developer name (optional)
#* @param organization_detail Filter by data presence: 'present' or 'absent' (optional)
#* @param identifier Filter by exact identifier value (optional)
#* @param fhir_version Comma-separated list of FHIR versions to filter (optional)
#* @description Download a CSV file containing daily organization data
function(res, developer = NULL, organization_detail = NULL, identifier = NULL, fhir_version = NULL) {
  # Normalize and parse fhir_versions
  fhir_versions_vec <- if (!is.null(fhir_version)) {
    strsplit(fhir_version, ",")[[1]] %>% trimws()
  } else {
    NULL
  }

  # Validate provided FHIR versions against known valid ones
  filtered_fhir_versions <- if (!is.null(fhir_versions_vec)) {
    valid <- fhir_versions_vec[fhir_versions_vec %in% valid_fhir_versions]
    invalid <- setdiff(fhir_versions_vec, valid)

    if (length(valid) == 0) {
      res$status <- 400
      return(list(
        error = paste0("None of the provided FHIR versions are valid. Accepted values include: ",
                      paste(sort(valid_fhir_versions), collapse = ", "))
      ))
    }

    if (length(invalid) > 0) {
      message("Ignoring invalid FHIR versions: ", paste(invalid, collapse = ", "))
    }

    valid
  } else {
    NULL
  }

  # Validate organization_detail flag
  organization_detail_flag <- NULL
  if (!is.null(organization_detail)) {
    organization_detail <- tolower(organization_detail)
    if (organization_detail == "present") {
      organization_detail_flag <- organization_detail
    } else {
      res$status <- 400
      return(list(error = "Invalid value for 'organization_detail'. Only 'present' is supported."))
    }
  }

  # Log filters for debugging
  message("Organization API Filters - Developer: ", developer, 
        ", Organization Detail: ", organization_detail_flag, 
        ", Identifier: ", identifier, 
        ", FHIR Versions: ", paste(fhir_versions_vec, collapse = ", "))

  # Only check developer if it's provided
  if (!is.null(developer)) {
    # Check against vendors table
    all_vendors <- tbl(db_connection, "vendors") %>%
      select(name) %>%
      distinct() %>%
      collect() %>%
      pull(name)

    if (!(developer %in% all_vendors)) {
      res$status <- 400
      return(list(
        error = paste0("Developer '", developer, "' not found in the CHPL-certified vendor list. ",
                       "Please check for typos or verify the exact developer name.")
      ))
    }
  }

  res$setHeader("Content-Type", "text/csv")

  st <- format(Sys.time(), "%Y-%m-%d")

  # Sanitize the filters for safe filenames
  safe_developer <- if (!is.null(developer)) {
    gsub("[^A-Za-z0-9_]+", "_", developer)  # replace spaces/special chars with underscores
  } else {
    NULL
  }
  safe_identifier <- if (!is.null(identifier)) {
    paste0("id_", gsub("[^A-Za-z0-9]", "", identifier))
  } else NULL

  safe_organization_detail <- if (!is.null(organization_detail_flag)) paste0("organization_detail_", organization_detail_flag) else NULL

  safe_fhir <- if (!is.null(filtered_fhir_versions)) {
    paste0("fhir_", gsub("[^A-Za-z0-9]", "", paste(filtered_fhir_versions, collapse = "_")))
  } else NULL

  filename_parts <- c("fhir_endpoint_organizations", safe_developer, safe_identifier, safe_organization_detail, safe_fhir, st)
  filename <- paste0(paste(na.omit(filename_parts), collapse = "_"), ".csv")  

  res$setHeader("Content-Disposition", paste0("attachment; filename=", filename))

  if (!file.exists(filename)) {
    org_data <- get_organization_csv_data(
      db_connection,
      developer = developer,
      organization_detail = organization_detail_flag,
      identifier = identifier,
      fhir_versions = filtered_fhir_versions
    )
    write.csv(org_data, file = filename, row.names = FALSE)
  }
  include_file(filename, res, content_type = "text/csv")
}