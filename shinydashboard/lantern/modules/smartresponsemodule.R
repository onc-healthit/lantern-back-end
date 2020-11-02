# SMART-on-FHIR Well-known URI responses

smartresponsemodule_UI <- function(id) {

  ns <- NS(id)

  tagList(
    p("This is the SMART-on-FHIR Core Capabilities response page. FHIR endpoints
      requiring authorization shall provide a JSON document at the endpoint URL with ",
      code("/.well-known/smart-configuration"), "appended to the end of the base URL."
    ),
    fluidRow(
      column(width = 6,
             tableOutput(ns("well_known_summary_table")),
             tableOutput(ns("smart_vendor_table"))
      ),
      column(width = 6,
             tableOutput(ns("smart_capability_count_table")))
    ),
    h3("Endpoints by Well Known URI support"),
    p("This is the list of endpoints which have returned a valid SMART Core Capabilities JSON document at the", code("/.well-known/smart-configuration"), " URI."),
    DT::dataTableOutput(ns("well_known_endpoints"))
  )
}

smartresponsemodule <- function(
  input,
  output,
  session,
  sel_fhir_version,
  sel_vendor
) {

  ns <- session$ns

  # IN PROGRESS - need to get the correct query with smart_http_response and smart_response columns
  # can we show a summary table of how many endpoints supporting /.well-known/smart-configuration ?
  output$smart_capability_count_table <- renderTable(
    get_smart_response_capability_count(isolate(app_data$smart_response_capabilities()))
  )

  output$smart_vendor_table <- renderTable({
    isolate(app_data$smart_response_capabilities()) %>%
      distinct(id, .keep_all = TRUE) %>%
      group_by(id, fhir_version, vendor_name) %>%
      count() %>%
      group_by(fhir_version, vendor_name) %>%
      count(wt = n) %>%
      select("FHIR Version" = fhir_version, "Developer" = vendor_name, "Endpoints" = n)
  })

  output$well_known_summary_table <- renderTable(
    isolate(app_data$well_known_endpoint_counts())
  )

  selected_endpoints <- reactive({
    res <- isolate(app_data$well_known_endpoints_tbl())
    req(sel_fhir_version(), sel_vendor())
    if (sel_fhir_version() != ui_special_values$ALL_FHIR_VERSIONS) {
      res <- res %>% filter(fhir_version == sel_fhir_version())
    }
    if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
      res <- res %>% filter(vendor_name == sel_vendor())
    }
    res
  })

  output$well_known_endpoints <-  DT::renderDataTable({
    datatable(selected_endpoints(),
              colnames = c("URL", "Organization", "Developer", "FHIR Version"),
              rownames = FALSE,
              options = list(scrollX = TRUE)
    )
  })

}
