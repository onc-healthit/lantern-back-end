source("downloadsmodule.R")

#* @apiTitle Download Daily FHIR Endpoints report
#* Echo Download Daily FHIR Endpoints report
#* @get /api/download
function(res) {
    res$setHeader("Content-Type", "text/csv")
    res$setHeader("Content-Disposition", "attachment; filename=fhir_endpoints.csv")
    write.csv(get_fhir_endpoints_tbl(db_tables), file='fhir_endpoints.csv', row.names=FALSE)
    include_file('fhir_endpoints.csv', res, content_type = "text/csv")
}