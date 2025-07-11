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
        write.csv(download_data(db_tables), file=filename, row.names=FALSE)
    }
    include_file(filename, res, content_type = "text/csv")
}

#* @get /organizations/v1
#* @param vendor Filter by vendor name (optional)
#* @description Download a CSV file containing daily organization data
function(res, vendor=NULL) {

  # Only check vendor if it's provided
  if (!is.null(vendor)) {
    # Check against vendors table
    all_vendors <- tbl(db_connection, "vendors") %>%
      select(name) %>%
      distinct() %>%
      collect() %>%
      pull(name)

    if (!(vendor %in% all_vendors)) {
      res$status <- 400
      return(list(
        error = paste0("Vendor '", vendor, "' not found in the CHPL-certified vendor list. ",
                       "Please check for typos or verify the exact vendor name.")
      ))
    }
  }

  res$setHeader("Content-Type", "text/csv")

  st <- format(Sys.time(), "%Y-%m-%d")

  # Sanitize vendor name for safe filenames
  safe_vendor <- if (!is.null(vendor)) {
    gsub("[^A-Za-z0-9_]+", "_", vendor)  # replace spaces/special chars with underscores
  } else {
    NULL
  }

  filename <- if (!is.null(safe_vendor)) {
    paste0("fhir_endpoint_organizations_", safe_vendor, "_", st, ".csv")
  } else {
    paste0("fhir_endpoint_organizations_", st, ".csv")
  }

  res$setHeader("Content-Disposition", paste0("attachment; filename=", filename))

  if (!file.exists(filename)) {
    org_data <- get_organization_csv_data(db_connection, vendor)
    write.csv(org_data, file = filename, row.names = FALSE)
  }
  include_file(filename, res, content_type = "text/csv")
}