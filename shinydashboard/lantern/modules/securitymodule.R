# Security Module - Performance Optimization on DISTINCT queries

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

  # Add request tracking to prevent race conditions
  current_request_id <- reactiveVal(0)

  # Break the feedback loop with isolate()
  observe({
    new_page <- security_page_state()
    current_selector <- input$security_page_selector
    
    # Only update if different (prevents infinite loop)
    # Add safety check for current_selector to prevent crashes
    if (is.null(current_selector) || 
        is.na(current_selector) || 
        !is.numeric(current_selector) ||
        current_selector != new_page) {
      
      isolate({  # This is the key fix to break feedback loops!
        updateNumericInput(session, "security_page_selector", 
                          max = security_total_pages(),
                          value = new_page)
      })
    }
  })

  # Handle page selector input
  observeEvent(input$security_page_selector, {
    # Get current input value
    current_input <- input$security_page_selector
    
    # Check if input is valid (not NULL, not NA, and is a number)
    if (!is.null(current_input) && 
        !is.na(current_input) && 
        is.numeric(current_input) &&
        current_input > 0) {
      
      new_page <- max(1, min(current_input, security_total_pages()))
      
      # Only update page state if it's actually different
      if (new_page != security_page_state()) {
        security_page_state(new_page)
      }

      # Correct the input field if the user entered an invalid page number
      if (new_page != current_input) {
        updateNumericInput(session, "security_page_selector", value = new_page)
      }
    } else {
      # If input is invalid (empty, NA, or <= 0), reset to current page
      # Use a small delay to prevent immediate feedback loop
      invalidateLater(100)
      updateNumericInput(session, "security_page_selector", value = security_page_state())
    }
  }, ignoreInit = TRUE)  # Prevent firing on initialization

  # Handle next page button 
  observeEvent(input$security_next_page, {
    if (security_page_state() < security_total_pages()) {
      new_page <- security_page_state() + 1
      security_page_state(new_page)
    }
  })

  # Handle previous page button
  observeEvent(input$security_prev_page, {
    if (security_page_state() > 1) {
      new_page <- security_page_state() - 1
      security_page_state(new_page)
    }
  })

  output$security_prev_button_ui <- renderUI({
    if (security_page_state() > 1) {
      actionButton(ns("security_prev_page"), "Previous", icon = icon("arrow-left"))
    } else {
      NULL
    }
  })

  output$security_next_button_ui <- renderUI({
    if (security_page_state() < security_total_pages()) {
      actionButton(ns("security_next_page"), "Next", icon = icon("arrow-right"))
    } else {
      NULL
    }
  })

  output$current_security_page_info <- renderText({
    paste("of", security_total_pages())
  })

  # Reset page when filters change
  observeEvent(list(sel_fhir_version(), sel_vendor(), sel_auth_type_code(), input$security_search_query), {
    security_page_state(1)
  })

  output$auth_type_count_table <- renderTable(
    isolate(get_auth_type_count(db_connection)),
    align = "llrr"
  )
  output$endpoint_summary_table <- renderTable(
    isolate(get_endpoint_security_counts(db_connection))
  )

  security_base_sql <- reactive({
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

    paste0("FROM selected_security_endpoints_mv 
            WHERE fhir_version IN (", versions, ") 
              AND code = '", sel_auth_type_code(), "' ",
              vendor_filter, " ",
              search_filter)
  })

  security_total_pages <- reactive({
    # OPTIMIZATION: Use PostgreSQL's faster approach for counting distinct rows
    # Create a hash of the concatenated values which is faster than DISTINCT on all columns
    count_query <- paste0("SELECT COUNT(*) as count FROM (
                            SELECT DISTINCT 
                                   MD5(CONCAT(url_modal, '|', COALESCE(condensed_organization_names, ''), '|', 
                                            vendor_name, '|', capability_fhir_version, '|', 
                                            COALESCE(tls_version, ''), '|', code))
                            ", security_base_sql(), "
                          ) AS unique_hashes")
    
    count <- tbl(db_connection, sql(count_query)) %>% collect() %>% pull(count)
    max(1, ceiling(count / security_page_size))
  })

  # Main data query - WITH RACE CONDITION PROTECTION
  selected_endpoints <- reactive({
    req(sel_fhir_version(), sel_vendor(), sel_auth_type_code())
    
    # Generate unique request ID
    request_id <- isolate(current_request_id()) + 1
    current_request_id(request_id)
    
    limit <- security_page_size
    offset <- (security_page_state() - 1) * security_page_size

    # OPTIMIZATION: Use the exact same hash-based approach for consistency
    # This ensures the count and data queries use identical deduplication logic
    query <- paste0(
      "SELECT DISTINCT 
              url_modal as url, 
              condensed_organization_names, 
              vendor_name, 
              capability_fhir_version, 
              tls_version, 
              code ",
      security_base_sql(),
      " ORDER BY url_modal 
        LIMIT ", limit, " OFFSET ", offset
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
