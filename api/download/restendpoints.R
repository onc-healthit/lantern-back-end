source("downloadsmodule.R")

#* @apiTitle Download Daily FHIR Endpoints report
#* Echo Download Daily FHIR Endpoints report
#* @get /daily/download
function(res) {
    res$setHeader("Content-Type", "text/csv")
    res$setHeader("Content-Disposition", "attachment; filename=fhir_endpoints.csv")
    st <- format(Sys.time(), "%Y-%m-%d")
    filename <- paste("fhir_endpoints_", st, ".csv", sep = "")
    print(filename)
    if (!file.exists(filename)) {
        print("Not")
        write.csv(get_fhir_endpoints_tbl(db_tables), file=filename, row.names=FALSE)
    }
    include_file(filename, res, content_type = "text/csv")
}