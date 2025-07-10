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

#* @get /daily/download-orgs
#* @description Download a CSV file containing daily organization data
function(res) {
  res$setHeader("Content-Type", "text/csv")
  res$setHeader("Content-Disposition", "attachment; filename=fhir_endpoint_organizations.csv")

  st <- format(Sys.time(), "%Y-%m-%d")
  filename <- paste0("fhir_endpoint_organizations_", st, ".csv")
  if (!file.exists(filename)) {
    org_data <- get_organization_csv_data(db_connection)
    write.csv(org_data, file = filename, row.names = FALSE)
  }
  include_file(filename, res, content_type = "text/csv")
}