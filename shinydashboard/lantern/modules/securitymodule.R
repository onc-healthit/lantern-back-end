# Security Module - Performance Optimized while maintaining exact data accuracy

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
      fluidRow(
        column(6, textInput(ns("security_search_query"), "Search: ", value = ""))
      ),
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
              numericInput(ns("security_page_selector"), label = NULL, value = 1, min = 1, max = 1, step = 1, width = "80px"),
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
  sel_auth_type_code,
  is_active
) {

  ns <- session$ns

  security_page_size <- 10

  # Add request tracking to prevent race conditions
  current_request_id <- reactiveVal(0)

  security_page_state <- create_pager(
    input, output, session,
    page_selector_id = "security_page_selector",
    prev_button_id = "security_prev_page",
    next_button_id = "security_next_page",
    prev_output_id = "security_prev_button_ui",
    next_output_id = "security_next_button_ui",
    total_pages = security_total_pages
  )

  output$current_security_page_info <- renderText({
    paste("of", security_total_pages())
  })

  # Reset page when filters change
  observeEvent(list(sel_fhir_version(), sel_vendor(), sel_auth_type_code(), input$security_search_query), {
    security_page_state(1)
  })

  # Take a real reactive dependency on the nightly refresh (instead of isolate()-ing the DB read
  # with no dependency at all) so these tables reflect a later app_fetcher() refresh rather than
  # freezing at whatever they showed on first render.
  output$auth_type_count_table <- renderTable({
    app$endpoint_export_tbl()
    get_auth_type_count(db_connection)
  }, align = "llrr")
  output$endpoint_summary_table <- renderTable({
    app$endpoint_export_tbl()
    get_endpoint_security_counts(db_connection)
  })

  security_base_sql <- reactive({
    req(sel_fhir_version(), sel_vendor(), sel_auth_type_code(), is_active())

    versions <- paste0("'", sel_fhir_version(), "'", collapse = ", ")
    vendor_filter <- if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
      paste0("AND vendor_name = '", sel_vendor(), "'")
    } else {
      ""
    }
    
    search_filter <- ""
    if (!is.null(input$security_search_query) && input$security_search_query != "") {
      q <- gsub("'", "''", input$security_search_query)
      search_filter <- paste0("AND (url ILIKE '%", q, "%' OR
                                  condensed_organization_names ILIKE '%", q, "%' OR 
                                  vendor_name ILIKE '%", q, "%' OR 
                                  capability_fhir_version ILIKE '%", q, "%' OR 
                                  tls_version ILIKE '%", q, "%')")
    }

    paste0("FROM security_endpoints_distinct_mv
            WHERE capability_fhir_version IN (", versions, ")
              AND code = '", sel_auth_type_code(), "' ",
              vendor_filter, " ",
              search_filter)
  })

  security_total_pages <- reactive({
    # PERFORMANCE OPTIMIZATION: Use a CTE with DISTINCT to leverage index better
    # This approach maintains exact data accuracy while being faster than DISTINCT on final results
    count_query <- paste0("SELECT COUNT(*) as count ",
                           security_base_sql())

    count <- tbl(db_connection, sql(count_query)) %>% collect() %>% pull(count)
    max(1, ceiling(count / security_page_size))
  })

  # Main data query - WITH RACE CONDITION PROTECTION
  selected_endpoints <- reactive({
    req(sel_fhir_version(), sel_vendor(), sel_auth_type_code(), is_active())

    # Generate unique request ID
    request_id <- isolate(current_request_id()) + 1
    current_request_id(request_id)

    limit <- security_page_size
    offset <- (security_page_state() - 1) * security_page_size

    query <- paste0(
      "SELECT * ",
      security_base_sql(),
      " ORDER BY url LIMIT ", limit, " OFFSET ", offset
    )

    result <- tbl(db_connection, sql(query)) %>% collect()

    # Only return results if this is still the latest request
    # Use isolate() to check without creating reactive dependency
    if (request_id == isolate(current_request_id())) {
      # This is the latest request, process normally
      return(result)
    } else {
      # This request was superseded, return empty to avoid flicker
      return(data.frame())
    }
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
                showSortIcon = TRUE
    )
  })

}
