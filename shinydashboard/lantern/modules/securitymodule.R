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
    h2("Endpoints by Authorization Type"),
    div(
      uiOutput("show_security_filter"),
      tags$p("The URL for each endpoint in the table below can be clicked on to see additional information for that individual endpoint.", role = "comment"),
      reactable::reactableOutput(ns("security_endpoints"))
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

  # url requested version is default set to None since this table filters on requested_version = 'None'
  selected_endpoints <- reactive({
  # Set default page size if needed
  if (is.null(securityPageSizeNum())) {
    securityPageSizeNum(10)
  }
  # Ensure required reactive values are available
  req(sel_fhir_version(), sel_vendor(), sel_auth_type_code())
  # Query the materialized view directly
  res <- tbl(db_connection, sql("SELECT url_modal as url, 
                                 condensed_organization_names, 
                                 vendor_name, 
                                 capability_fhir_version, 
                                 fhir_version, 
                                 tls_version, 
                                 code 
                          FROM selected_security_endpoints_mv")) %>%
    collect()
  # Apply filters based on user selections
  res <- res %>% filter(fhir_version %in% sel_fhir_version())
  if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
    res <- res %>% filter(vendor_name == sel_vendor())
  }
  res <- res %>%
    filter(code == sel_auth_type_code()) %>%
    distinct(url, condensed_organization_names, vendor_name, capability_fhir_version, tls_version, code) %>%
    select(url, condensed_organization_names, vendor_name, capability_fhir_version, tls_version, code)
  return(res)
})

  output$security_endpoints <-  reactable::renderReactable({
    reactable(selected_endpoints(),
                columns = list(
                  url = colDef(name = "URL", html = TRUE),
                  condensed_organization_names = colDef(name = "Organization", html = TRUE),
                  vendor_name = colDef(name = "Developer"),
                  capability_fhir_version = colDef(name = "FHIR Version"),
                  tls_version = colDef(name = "TLS Version"),
                  code = colDef(name = "Authorization")
                ),
                sortable = TRUE,
                searchable = TRUE,
                showSortIcon = TRUE,
                defaultPageSize = isolate(securityPageSizeNum())
    )
  })

  observeEvent(input$security_endpoints_state$length, {
    page <- input$security_endpoints_state$length
    securityPageSizeNum(page)
  })

}
