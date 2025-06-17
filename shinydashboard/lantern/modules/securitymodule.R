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
      reactable::reactableOutput(ns("security_endpoints")),
      fluidRow(
        column(3, 
          div(style = "display: flex; justify-content: flex-start;", 
              uiOutput(ns("security_prev_button_ui"))
          )
        ),
        column(6,
          div(style = "display: flex; justify-content: center; align-items: center; gap: 10px; margin-top: 8px;",
              textOutput(ns("current_security_page_info"), inline = TRUE)
          )
        ),
        column(3, 
          div(style = "display: flex; justify-content: flex-end;",
              uiOutput(ns("security_next_button_ui"))
          )
        )
      )
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

  security_page_size <- 10
  security_page_state <- reactiveVal(1)

security_total_pages <- reactive({
  req(sel_fhir_version(), sel_vendor(), sel_auth_type_code())

  versions <- paste0("'", sel_fhir_version(), "'", collapse = ", ")
  vendor_filter <- if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
    paste0("AND vendor_name = '", sel_vendor(), "'")
  } else {
    ""
  }

  query <- paste0(
    "SELECT COUNT(*) as count 
     FROM selected_security_endpoints_mv 
     WHERE fhir_version IN (", versions, ") 
       AND code = '", sel_auth_type_code(), "' ",
    vendor_filter
  )

  count <- tbl(db_connection, sql(query)) %>% collect() %>% pull(count)
  max(1, ceiling(count / security_page_size))
})

  observeEvent(input$security_next_page, {
    if (security_page_state() < security_total_pages()) security_page_state(security_page_state() + 1)
  })

  observeEvent(input$security_prev_page, {
    if (security_page_state() > 1) security_page_state(security_page_state() - 1)
  })

  output$security_prev_button_ui <- renderUI({
    if (security_page_state() > 1) actionButton(ns("security_prev_page"), "Previous") else NULL
  })

  output$security_next_button_ui <- renderUI({
    if (security_page_state() < security_total_pages()) actionButton(ns("security_next_page"), "Next") else NULL
  })

  output$current_security_page_info <- renderText({
    paste("Page", security_page_state(), "of", security_total_pages())
  })

  # Reset page when filters change
  observeEvent(list(sel_fhir_version(), sel_vendor(), sel_auth_type_code()), {
    security_page_state(1)
  })


  output$auth_type_count_table <- renderTable(
    isolate(get_auth_type_count(db_connection)),
    align = "llrr"
  )
  output$endpoint_summary_table <- renderTable(
    isolate(get_endpoint_security_counts(db_connection))
  )

  # url requested version is default set to None since this table filters on requested_version = 'None'
  selected_endpoints <- reactive({
  # Ensure required reactive values are available
  req(sel_fhir_version(), sel_vendor(), sel_auth_type_code())

  versions <- paste0("'", sel_fhir_version(), "'", collapse = ", ")
  vendor_filter <- if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
    paste0("AND vendor_name = '", sel_vendor(), "'")
  } else {
    ""
  }
  
  limit <- security_page_size
  offset <- (security_page_state() - 1) * security_page_size
  
  # Query the materialized view directly
  query <- paste0(
    "SELECT url_modal as url, 
            condensed_organization_names, 
            vendor_name, 
            capability_fhir_version, 
            tls_version, 
            code 
     FROM selected_security_endpoints_mv 
     WHERE fhir_version IN (", versions, ") 
       AND code = '", sel_auth_type_code(), "' ",
       vendor_filter, 
     " ORDER BY url_modal 
       LIMIT ", limit, " OFFSET ", offset
  )

  tbl(db_connection, sql(query)) %>% collect()

  # TODO do we need distinct? this was there previously
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
                defaultPageSize = security_page_size
    )
  })

}
