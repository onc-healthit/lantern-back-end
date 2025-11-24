# Security Module - Enhanced with React components
# Properly displays static R tables AND React components

securitymodule_UI <- function(id) {
  ns <- NS(id)

  tagList(
    # Description
    p("This is the list of security authorization types reported by the CapabilityStatement / Conformance Resources from the endpoints."),

    # STATIC R TABLES - Keep these as original R tableOutput
    fluidRow(
      column(width = 6,
             h4("Endpoint Security Summary"),
             tableOutput(ns("endpoint_summary_table"))
      ),
      column(width = 6,
             h4("Authorization Type Counts"),
             tableOutput(ns("auth_type_count_table"))
      )
    ),
    
    # Add spacing
    tags$hr(style = "margin: 30px 0;"),

    h2("Endpoints by Authorization Type"),

    # Auth type filter dropdown (required by server.R)
    uiOutput("show_security_filter"),

    # React auth type badges
    div(id = ns("auth_type_badges_container"), style = "margin-bottom: 24px;"),

    # React search bar
    fluidRow(
      column(12,
        div(id = ns("security_search_container"),
          # Hidden input for Shiny integration
          div(style = "display: none;",
            textInput(ns("security_search_query"), "Search: ", value = "")
          )
        )
      )
    ),

    # Info text
    tags$p("The URL for each endpoint in the table below can be clicked on to see additional information for that individual endpoint.", role = "comment"),

    # Data table
    reactable::reactableOutput(ns("security_endpoints")),

    # Pagination
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
    ),

    # Initialize React components (auth badges and search only)
    tags$script(HTML(sprintf("
      (function() {
        var initAttempts = 0;
        var namespace = '%s';
        console.log('[Security Module] Using namespace:', namespace);

        var checkReactInterval = setInterval(function() {
          if (typeof window.SecurityReactComponents !== 'undefined' &&
              typeof React !== 'undefined' &&
              typeof ReactDOM !== 'undefined') {
            clearInterval(checkReactInterval);
            console.log('[Security Module] React libraries loaded, initializing components');

            initializeSecurityReactComponents(namespace);
          } else if (++initAttempts > 50) {
            clearInterval(checkReactInterval);
            console.error('[Security Module] React components failed to load after 5 seconds');
            console.error('[Security Module] SecurityReactComponents:', typeof window.SecurityReactComponents);
            console.error('[Security Module] React:', typeof React);
            console.error('[Security Module] ReactDOM:', typeof ReactDOM);
          }
        }, 100);
      })();
    ", ns(""))))
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
  current_request_id <- reactiveVal(0)

  # ============================================
  # STATIC R TABLES - Keep original implementation
  # ============================================
  
  output$auth_type_count_table <- renderTable({
    get_auth_type_count(db_connection)
  }, align = "llrr")
  
  output$endpoint_summary_table <- renderTable({
    get_endpoint_security_counts(db_connection)
  })

  # ============================================
  # PAGINATION LOGIC
  # ============================================

  # Sync page state with page selector input
  observe({
    new_page <- security_page_state()
    current_selector <- input$security_page_selector

    # Only update if different and valid
    if (is.null(current_selector) || is.na(current_selector) ||
        !is.numeric(current_selector) || current_selector != new_page) {
      isolate({
        updateNumericInput(session, "security_page_selector",
                          max = security_total_pages(),
                          value = new_page)
      })
    }
  })

  # Handle page selector input
  observeEvent(input$security_page_selector, {
    current_input <- input$security_page_selector

    if (!is.null(current_input) && !is.na(current_input) &&
        is.numeric(current_input) && current_input > 0) {
      new_page <- max(1, min(current_input, security_total_pages()))

      if (new_page != security_page_state()) {
        security_page_state(new_page)
      }

      if (new_page != current_input) {
        updateNumericInput(session, "security_page_selector", value = new_page)
      }
    } else {
      invalidateLater(100)
      updateNumericInput(session, "security_page_selector", value = security_page_state())
    }
  }, ignoreInit = TRUE)

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

  # ============================================
  # REACT COMPONENTS - Send data to React
  # ============================================

  # Get summary data for React badges (auth type badges only)
  summary_data <- reactive({
    auth_counts <- get_auth_type_count(db_connection)
    list(auth_counts = auth_counts)
  })

  # Send auth badge data to React
  observe({
    data <- summary_data()

    # Send auth badge data to React
    if(nrow(data$auth_counts) > 0 && "Code" %in% names(data$auth_counts)) {
      auth_aggregated <- data$auth_counts %>%
        group_by(Code) %>%
        summarise(total_endpoints = sum(as.numeric(Endpoints), na.rm = TRUE), .groups = 'drop') %>%
        arrange(desc(total_endpoints))

      auth_badge_data <- lapply(1:nrow(auth_aggregated), function(i) {
        row <- auth_aggregated[i, ]
        list(
          type = as.character(row$Code),
          count = as.numeric(row$total_endpoints),
          isActive = (as.character(row$Code) == sel_auth_type_code())
        )
      })

      session$sendCustomMessage(
        type = paste0(ns(""), "update_auth_badges"),
        message = auth_badge_data
      )
    }
  })

  # ============================================
  # DATA QUERIES
  # ============================================

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
    count_query <- paste0("SELECT COUNT(*) as count ", security_base_sql())
    count <- tbl(db_connection, sql(count_query)) %>% collect() %>% pull(count)
    max(1, ceiling(count / security_page_size))
  })

  # Main data query with race condition protection
  selected_endpoints <- reactive({
    req(sel_fhir_version(), sel_vendor(), sel_auth_type_code())

    request_id <- isolate(current_request_id()) + 1
    current_request_id(request_id)

    limit <- security_page_size
    offset <- (security_page_state() - 1) * security_page_size

    query <- paste0("SELECT * ", security_base_sql(),
                    " ORDER BY url LIMIT ", limit, " OFFSET ", offset)

    result <- tbl(db_connection, sql(query)) %>% collect()

    # Only return if this is still the latest request
    if (request_id == isolate(current_request_id())) {
      return(result)
    } else {
      return(data.frame())
    }
  })

  # ============================================
  # MAIN DATA TABLE OUTPUT
  # ============================================

  output$security_endpoints <- reactable::renderReactable({
    reactable(selected_endpoints(),
      columns = list(
        url = colDef(name = "URL", html = TRUE, minWidth = 250),
        condensed_organization_names = colDef(name = "Organization", html = TRUE, minWidth = 200),
        vendor_name = colDef(name = "Developer", minWidth = 150),
        capability_fhir_version = colDef(name = "FHIR Version", minWidth = 120),
        tls_version = colDef(name = "TLS Version", minWidth = 120),
        code = colDef(name = "Authorization", minWidth = 150)
      ),
      sortable = TRUE,
      showSortIcon = TRUE,
      highlight = TRUE,
      striped = TRUE,
      theme = reactableTheme(
        borderColor = "#e0e0e0",
        stripedColor = "#f9f9f9",
        highlightColor = "#f0f7ff",
        cellPadding = "12px 8px",
        style = list(
          fontFamily = "-apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Oxygen', 'Ubuntu', 'Cantarell', 'Fira Sans', 'Droid Sans', 'Helvetica Neue', sans-serif"
        ),
        headerStyle = list(
          background = "#f6f7f8",
          color = "#333",
          fontWeight = 600,
          fontSize = "14px",
          borderBottom = "2px solid #1B5A7F"
        ),
        rowStyle = list(cursor = "pointer")
      )
    )
  })

}