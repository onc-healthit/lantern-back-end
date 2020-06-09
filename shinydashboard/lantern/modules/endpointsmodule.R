library(DT)
library(purrr)

endpointsmodule_UI <- function(id) {
  
  fhir_version_list <- get_fhir_version_list(endpoint_export_tbl)
  vendor_list <- get_vendor_list(endpoint_export_tbl)
  
  ns <- NS(id)
  
  tagList(
    fluidRow(
      column(width=4,
             textOutput(ns("endpoint_count"))
      ),
      column(width=4,
             selectInput(
               inputId = ns("fhir_version"),
               label = "FHIR Version:",
               choices = fhir_version_list,
               selected = 99,
               size = length(fhir_version_list),
               selectize = FALSE)
      ),
      column(width=4,
             selectInput(
               inputId = ns("vendor"),
               label = "Vendor:",
               choices = vendor_list,
               selected = 99,
               size = length(vendor_list),
               selectize = FALSE)
      )
    ),
    DT::dataTableOutput(ns("endpoints_table"))
  )
}

endpointsmodule <- function(
  input, 
  output, 
  session
){
  ns <- session$ns

  output$endpoint_count <- renderText({paste("Matching Endpoints:",nrow(selected_fhir_endpoints()))})
  
  selected_fhir_endpoints <- reactive({
    res <- get_fhir_endpoints_tbl(db_tables)
    req(input$fhir_version,input$vendor)
    if (input$fhir_version != G$ALL_FHIR_VERSIONS) res <- res %>% filter(fhir_version == input$fhir_version)
    if (input$vendor != G$ALL_VENDORS) res <- res %>% filter(vendor_name == input$vendor)
    res
  })

  output$endpoints_table <- DT::renderDataTable(selected_fhir_endpoints())

}
