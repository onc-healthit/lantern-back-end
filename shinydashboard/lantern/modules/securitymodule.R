# Security Module - Performance Optimized while maintaining exact data accuracy

log_duration <- function(label, expr) {
  start_time <- Sys.time()
  result <- expr
  duration <- Sys.time() - start_time
  result
}

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
  sel_auth_type_code
) {

  ns <- session$ns

  security_page_size <- 10
  security_page_state <- reactiveVal(1)

  # Handle page selection
  observe({
    updateNumericInput(session, "security_page_selector", 
                      max = security_total_pages(),
                      value = security_page_state())
  })

  observeEvent(input$security_page_selector, {
    if (!is.null(input$security_page_selector) && !is.na(input$security_page_selector)) {
      new_page <- max(1, min(input$security_page_selector, security_total_pages()))
      security_page_state(new_page)
      if (new_page != input$security_page_selector) {
        updateNumericInput(session, "security_page_selector", value = new_page)
      }
    }
  })

  observeEvent(input$security_next_page, {
    # Double-click protection
    current_time <- as.numeric(Sys.time()) * 1000
    if (!is.null(session$userData$last_security_next_time) && 
        (current_time - session$userData$last_security_next_time) < 1000) {
      return()  # Ignore rapid consecutive clicks
    }
    session$userData$last_security_next_time <- current_time
    
    if (security_page_state() < security_total_pages()) {
      new_page <- security_page_state() + 1
      security_page_state(new_page)
      updateNumericInput(session, "security_page_selector", value = new_page)
    }
  })

  observeEvent(input$security_prev_page, {
    # Double-click protection
    current_time <- as.numeric(Sys.time()) * 1000
    if (!is.null(session$userData$last_security_prev_time) && 
        (current_time - session$userData$last_security_prev_time) < 1000) {
      return()  # Ignore rapid consecutive clicks
    }
    session$userData$last_security_prev_time <- current_time
    
    if (security_page_state() > 1) {
      new_page <- security_page_state() - 1
      security_page_state(new_page)
      updateNumericInput(session, "security_page_selector", value = new_page)
    }
  })

  output$security_prev_button_ui <- renderUI({
    if (security_page_state() > 1) actionButton(ns("security_prev_page"), "Previous") else NULL
  })

  output$security_next_button_ui <- renderUI({
    if (security_page_state() < security_total_pages()) actionButton(ns("security_next_page"), "Next") else NULL
  })

  output$current_security_page_info <- renderText({
    paste("of", security_total_pages())
  })

  # Reset page when filters change
  observeEvent(list(sel_fhir_version(), sel_vendor(), sel_auth_type_code(), input$security_search_query), {
    security_page_state(1)
    updateNumericInput(session, "security_page_selector", value = 1)
  })

  output$auth_type_count_table <- renderTable(
    isolate(get_auth_type_count(db_connection)),
    align = "llrr"
  )
  output$endpoint_summary_table <- renderTable(
    isolate(get_endpoint_security_counts(db_connection))
  )

  security_base_sql <- reactive({
    log_duration("security_base_sql", {
    req(sel_fhir_version(), sel_vendor(), sel_auth_type_code())

    versions <- paste0("'", sel_fhir_version(), "'", collapse = ", ")
    vendor_filter <- if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
      paste0("AND vendor_name = '", sel_vendor(), "'")
    } else {
      ""
    }
    
    search_filter <- ""
    if (!is.null(input$security_search_query) && input$security_search_query != "") {
      q <- gsub("'", "''", input$security_search_query)
      search_filter <- paste0("AND (url_modal ILIKE '%", q, "%' OR 
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
  })

  security_total_pages <- reactive({
    # PERFORMANCE OPTIMIZATION: Use a CTE with DISTINCT to leverage index better
    # This approach maintains exact data accuracy while being faster than DISTINCT on final results
    log_duration("security_total_pages", {
    count_query <- paste0("SELECT COUNT(*) as count ",
                           security_base_sql())
    
    count <- tbl(db_connection, sql(count_query)) %>% collect() %>% pull(count)
    max(1, ceiling(count / security_page_size))
    })
  })

  selected_endpoints <- reactive({
    log_duration("selected_endpoints", {
    limit <- security_page_size
    offset <- (security_page_state() - 1) * security_page_size

    query <- paste0(
      "SELECT * ",
      security_base_sql(),
      " ORDER BY url LIMIT ", limit, " OFFSET ", offset
    )

    tbl(db_connection, sql(query)) %>% collect()
    })
  })

  output$security_endpoints <-  reactable::renderReactable({
    log_duration("renderReactable_security_endpoints", {
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
  })

}
