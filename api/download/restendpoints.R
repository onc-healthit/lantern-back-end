source("downloadsmodule.R")

#* @apiTitle Download Daily FHIR Endpoints report
#* Echo Download Daily FHIR Endpoints report
#* @get /daily/download
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