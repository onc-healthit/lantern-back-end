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
              uiOutput(ns("prev_button_ui"))
          )
        ),
        column(6,
          div(style = "display: flex; justify-content: center; align-items: center; gap: 10px; margin-top: 8px;",
              numericInput(ns("page_selector"), label = NULL, value = 1, min = 1, max = 1, step = 1, width = "80px"),
              textOutput(ns("page_info"), inline = TRUE)
          )
        ),
        column(3, 
          div(style = "display: flex; justify-content: flex-end;",
              uiOutput(ns("next_button_ui"))
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

  page_size <- 10
  page_state <- reactiveVal(1)

total_pages <- reactive({
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
  max(1, ceiling(count / page_size))
})

  # Update page selector max when total pages change
  observe({
    updateNumericInput(session, "page_selector", 
                      max = total_pages(),
                      value = page_state())
  })

  # Handle page selector input
  observeEvent(input$page_selector, {
    if (!is.null(input$page_selector) && !is.na(input$page_selector)) {
      new_page <- max(1, min(input$page_selector, total_pages()))
      page_state(new_page)
      
      # Update the input if user entered invalid value
      if (new_page != input$page_selector) {
        updateNumericInput(session, "page_selector", value = new_page)
      }
    }
  })

  observeEvent(input$next_page, {
    if (page_state() < total_pages()) page_state(page_state() + 1)
  })

  observeEvent(input$prev_page, {
    if (page_state() > 1) page_state(page_state() - 1)
  })

  output$prev_button_ui <- renderUI({
    if (page_state() > 1) actionButton(ns("prev_page"), "Previous") else NULL
  })

  output$next_button_ui <- renderUI({
    if (page_state() < total_pages()) actionButton(ns("next_page"), "Next") else NULL
  })

  output$page_info <- renderText({
    paste("of", total_pages())
  })

  output$current_page_info <- renderText({
    paste("Page", page_state(), "of", total_pages())
  })

  # Reset page when filters change
  observeEvent(list(sel_fhir_version(), sel_vendor(), sel_auth_type_code()), {
    page_state(1)
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
  
  limit <- page_size
  offset <- (page_state() - 1) * page_size
  
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
                defaultPageSize = page_size
    )
  })

}
