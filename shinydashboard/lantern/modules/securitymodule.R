# Security Module

securitymodule_UI <- function(id) {

  ns <- NS(id)

  tagList(
    p("This is the list of security authorization types reported by the capability statements from the endpoints."),
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
      # uiOutput(ns("show_security_filter")),
      fluidRow(
        column(width = 4,
               selectInput(
                 inputId = ns("auth_type_code"),
                 label = "Supported Authorization Type:",
                 choices = isolate(app_data$security_code_list()),
                 selected = "SMART-on-FHIR",
                 size = 1,
                 selectize = FALSE)
        )
      ),
      DT::dataTableOutput(ns("security_endpoints"))
    )
  )
}

securitymodule <- function(
  input,
  output,
  session,
  sel_fhir_version,
  sel_vendor
) {

  ns <- session$ns

  output$auth_type_count_table <- renderTable(
    isolate(app_data$auth_type_counts()),
    align = "llrr"
  )
  output$endpoint_summary_table <- renderTable(
    isolate(app_data$endpoint_security_counts())
  )

  selected_endpoints <- reactive({
    res <- isolate(app_data$security_endpoints_tbl())
    req(sel_fhir_version(), sel_vendor(), input$auth_type_code)
    if (sel_fhir_version() != ui_special_values$ALL_FHIR_VERSIONS) {
      res <- res %>% filter(fhir_version == sel_fhir_version())
    }
    if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
      res <- res %>% filter(vendor_name == sel_vendor())
    }
    res <- res %>% filter(code == input$auth_type_code)
    res
  })

  output$security_endpoints <-  DT::renderDataTable({
    datatable(selected_endpoints(),
              colnames = c("URL", "Organization", "Developer", "FHIR Version", "TLS Version", "Authorization"),
              rownames = FALSE,
              options = list(scrollX = TRUE)
    )
  })
}
