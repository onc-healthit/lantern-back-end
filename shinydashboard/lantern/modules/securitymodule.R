# Security Module

securitymodule_UI <- function(id) {

  ns <- NS(id)

  tagList(
    p("This is the list of security authorization types reported by the CapabilityStatement / Conformance Resources from the endpoints."),
    fluidRow(
      column(width = 6,
             tableOutput(ns("endpoint_summary_table"))
      ),
      column(width = 6,
             tableOutput(ns("auth_type_count_table"))
      )
    ),
    h3("Endpoints by Authorization Type"),
    div(
      uiOutput("show_security_filter"),
      DT::dataTableOutput(ns("security_endpoints"))
    )
  )
}

securitymodule <- function(
  input,
  output,
  session,
  sel_fhir_version,
  sel_vendor,
  sel_auth_type_code
) {

  ns <- session$ns

  output$auth_type_count_table <- renderTable(
    isolate(app_data$auth_type_counts()),
    align = "llrr"
  )
  output$endpoint_summary_table <- renderTable(
    isolate(app_data$endpoint_security_counts())
  )

  securityPageSizeNum <- reactiveVal(NULL)

  selected_endpoints <- reactive({
    if (is.null(securityPageSizeNum())) {
      securityPageSizeNum(10)
    } 
    res <- isolate(app_data$security_endpoints_tbl())
    req(sel_fhir_version(), sel_vendor(), sel_auth_type_code())
    res <- res %>% filter(fhir_version %in% sel_fhir_version())
    if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
      res <- res %>% filter(vendor_name == sel_vendor())
    }
    res <- res %>%
    filter(code == sel_auth_type_code()) %>%
    select(url, organization_names, vendor_name, capability_fhir_version, tls_version, code)
    res
  })

  output$security_endpoints <-  DT::renderDataTable({
    datatable(selected_endpoints(),
              colnames = c("URL", "Organization", "Developer", "FHIR Version", "TLS Version", "Authorization"),
              selection = "none",
              rownames = FALSE,
              options = list(scrollX = TRUE, stateSave = TRUE, pageLength = isolate(securityPageSizeNum()))
    )
  })

  observeEvent(input$security_endpoints_state$length, {
    page <- input$security_endpoints_state$length
    securityPageSizeNum(page)
  })

}
