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
    rowwise() %>%
    mutate(condensed_organization_names = ifelse(length(organization_names) > 0, ifelse(length(strsplit(organization_names, ";")[[1]]) > 5, paste0(paste0(head(strsplit(organization_names, ";")[[1]], 5), collapse = ";"), "; ", paste0("<a class=\"lantern-url\" tabindex=\"0\" aria-label=\"Press enter to open a pop up modal containing the endpoint's entire list of API information source names.\" onkeydown = \"javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)\" onclick=\"Shiny.setInputValue(\'show_details\',&quot;", organization_names, "&quot,{priority: \'event\'});\"> Click For More... </a>")), organization_names), organization_names))

    res <- res %>%
    distinct(url, condensed_organization_names, vendor_name, capability_fhir_version, tls_version, code) %>%
    mutate(url = paste0("<a class=\"lantern-url\" tabindex=\"0\" aria-label=\"Press enter to open a pop up modal containing additional information for this endpoint.\" onkeydown = \"javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)\" onclick=\"Shiny.setInputValue(\'endpoint_popup\',&quot;", url, "&&", "None", "&quot,{priority: \'event\'});\">", url, "</a>")) %>%
    select(url, condensed_organization_names, vendor_name, capability_fhir_version, tls_version, code)
    res
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
