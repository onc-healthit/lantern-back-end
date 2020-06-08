library(DT)
endpointsmodule_UI <- function(id) {
  
  ns <- NS(id)
  
  tagList(
    h1("endpoints table"),
    DT::dataTableOutput(ns("endpoints_table"))
  )
}

endpointsmodule <- function(
  input, 
  output, 
  session
){
  ns <- session$ns
  output$endpoints_table <- DT::renderDataTable(get_fhir_endpoints_tbl(db_tables))

}
