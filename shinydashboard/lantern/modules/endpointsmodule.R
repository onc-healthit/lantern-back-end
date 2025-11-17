library(DT)
library(purrr)
library(reactable)
library(glue)
library(htmltools)

endpointsmodule_UI <- function(id) {
  
  ns <- NS(id)
  
  tagList(
    # Custom CSS for modern styling
    tags$head(
      tags$style(HTML(paste0("
        /* Modern table styling */
        .modern-endpoints-table {
          font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
        }
        
        .modern-endpoints-table .rt-table {
          border: 1px solid #e1e4e8;
          border-radius: 8px;
          overflow: hidden;
          box-shadow: 0 1px 3px rgba(0,0,0,0.08);
        }
        
        .modern-endpoints-table .rt-thead {
          background: linear-gradient(to bottom, #f8f9fa 0%, #f1f3f5 100%);
          border-bottom: 2px solid #dee2e6;
        }
        
        .modern-endpoints-table .rt-th {
          color: #495057;
          font-weight: 600;
          font-size: 13px;
          text-transform: uppercase;
          letter-spacing: 0.5px;
          padding: 12px 8px;
          border-right: 1px solid #e9ecef;
        }
        
        .modern-endpoints-table .rt-td {
          padding: 12px 8px;
          border-right: 1px solid #f1f3f5;
          font-size: 14px;
          color: #212529;
        }
        
        .modern-endpoints-table .rt-tr:hover {
          background-color: #f8f9fa;
          transition: background-color 0.2s ease;
        }
        
        .modern-endpoints-table .rt-tr-striped {
          background-color: #fafbfc;
        }
        
        /* Status badges */
        .status-badge {
          display: inline-block;
          padding: 4px 12px;
          border-radius: 12px;
          font-size: 12px;
          font-weight: 500;
          text-align: center;
        }
        
        .status-success {
          background-color: #d4edda;
          color: #155724;
          border: 1px solid #c3e6cb;
        }
        
        .status-warning {
          background-color: #fff3cd;
          color: #856404;
          border: 1px solid #ffeaa7;
        }
        
        .status-error {
          background-color: #f8d7da;
          color: #721c24;
          border: 1px solid #f5c6cb;
        }
        
        .status-info {
          background-color: #d1ecf1;
          color: #0c5460;
          border: 1px solid #bee5eb;
        }
        
        /* Availability progress bar */
        .availability-container {
          display: flex;
          align-items: center;
          gap: 8px;
        }
        
        .availability-bar {
          flex: 1;
          height: 8px;
          background-color: #e9ecef;
          border-radius: 4px;
          overflow: hidden;
        }
        
        .availability-fill {
          height: 100%;
          border-radius: 4px;
          transition: width 0.3s ease;
        }
        
        .availability-text {
          font-weight: 600;
          font-size: 13px;
          min-width: 40px;
          text-align: right;
        }
        
        /* Version badge */
        .version-badge {
          background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
          color: white;
          padding: 4px 10px;
          border-radius: 10px;
          font-size: 12px;
          font-weight: 500;
          display: inline-block;
        }
        
        /* Modern search box */
        #", ns("search_query"), " {
          border: 2px solid #e1e4e8;
          border-radius: 8px;
          padding: 10px 16px;
          font-size: 14px;
          transition: border-color 0.2s ease, box-shadow 0.2s ease;
        }
        
        #", ns("search_query"), ":focus {
          border-color: #667eea;
          box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
          outline: none;
        }
        
        /* Modern buttons */
        .modern-nav-button {
          background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
          border: none;
          color: white;
          padding: 10px 20px;
          border-radius: 8px;
          font-weight: 500;
          cursor: pointer;
          transition: transform 0.2s ease, box-shadow 0.2s ease;
        }
        
        .modern-nav-button:hover {
          transform: translateY(-2px);
          box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
        }
        
        .modern-nav-button:active {
          transform: translateY(0);
        }
        
        /* Page selector */
        #", ns("page_selector"), " {
          border: 2px solid #e1e4e8;
          border-radius: 6px;
          text-align: center;
          font-weight: 600;
        }
        
        /* Download buttons */
        .btn-primary {
          background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
          border: none;
          border-radius: 8px;
          padding: 10px 20px;
          font-weight: 500;
          transition: transform 0.2s ease, box-shadow 0.2s ease;
        }
        
        .btn-primary:hover {
          transform: translateY(-2px);
          box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
        }
        
        /* Info text styling */
        .endpoint-note {
          background: #f8f9fa;
          border-left: 4px solid #667eea;
          padding: 12px 16px;
          border-radius: 4px;
          margin: 16px 0;
        }
      ")))
    ),
    fluidRow(
      column(width = 12, style = "padding-bottom:20px",
             h2(style = "margin-top:0; color: #212529; font-weight: 600;", textOutput(ns("endpoint_count"))),
             # Add note for the endpoint table count and Matching Unique Endpoints count discrepancy
             div(class = "endpoint-note",
                 tags$strong(style = "font-style: italic; color: #495057;",
                    "Note: The table below may show multiple rows per endpoint depending on the number of FHIR versions supported by the endpoint.")
             ),
             downloadButton(ns("download_data"), "Download Endpoint Data (CSV)", icon = tags$i(class = "fa fa-download", "aria-hidden" = "true", role = "presentation", "aria-label" = "download icon")),
             downloadButton(ns("download_descriptions"), "Download Field Descriptions (CSV)", icon = tags$i(class = "fa fa-download", "aria-hidden" = "true", role = "presentation", "aria-label" = "download icon")),
             htmlOutput(ns("anchorlink"))
      )
    ),
    tags$p("The URL for each endpoint in the table below can be clicked on to see additional information for that individual endpoint.", role = "comment", style = "color: #6c757d;"),
    fluidRow(
      column(width = 6, 
             tags$label("Search:", style = "font-weight: 600; color: #495057; margin-bottom: 8px;"),
             textInput(ns("search_query"), label = NULL, value = "", placeholder = "Search endpoints...")
      )
    ),
    div(class = "modern-endpoints-table",
        reactable::reactableOutput(ns("endpoints_table"))
    ),
    fluidRow(
      column(3, 
        div(style = "display: flex; justify-content: flex-start; margin-top: 16px;", 
            uiOutput(ns("prev_button_ui"))
        )
      ),
      column(6,
        div(style = "display: flex; justify-content: center; align-items: center; gap: 10px; margin-top: 16px;",
            numericInput(ns("page_selector"), label = NULL, value = 1, min = 1, max = 1, step = 1, width = "80px"),
            textOutput(ns("page_info"), inline = TRUE)
        )
      ),
      column(3, 
        div(style = "display: flex; justify-content: flex-end; margin-top: 16px;",
            uiOutput(ns("next_button_ui"))
        )
      )
    ),
    tags$p("* An asterisk after a 'true' value in the 'Capability Statement Returned' field indicates that the returned Capability Statement for the endpoint is not of kind 'instance', which is the kind Lantern expects.", 
           role = "comment", 
           style = "color: #6c757d; font-style: italic; margin-top: 16px;"),
    htmlOutput(ns("note_text"))
  )
}

endpointsmodule <- function(
  input,
  output,
  session,
  sel_fhir_version,
  sel_vendor,
  sel_availability,
  sel_is_chpl
) {
  ns <- session$ns

  page_state <- reactiveVal(1)
  page_size <- 10

  # Add request tracking to prevent race conditions
  current_request_id <- reactiveVal(0)

  # Calculate total pages based on ACTUAL TABLE ROWS (after distinct operation)
  total_pages <- reactive({
    # Count the actual distinct rows that will be displayed in the table
    table_data <- selected_fhir_endpoints_without_limit() %>% 
      select(urlModal, condensed_endpoint_names, endpoint_names, vendor_name, capability_fhir_version, format, cap_stat_exists, status, availability) %>% 
      distinct(urlModal, condensed_endpoint_names, endpoint_names, vendor_name, capability_fhir_version, format, cap_stat_exists, status, availability)
    
    total_records <- nrow(table_data)
    max(1, ceiling(total_records / page_size))
  })

  # Break the feedback loop with isolate()
  observe({
    new_page <- page_state()
    current_selector <- input$page_selector
    
    # Only update if different (prevents infinite loop)
    # Add safety check for current_selector to prevent crashes
    if (is.null(current_selector) || 
        is.na(current_selector) || 
        !is.numeric(current_selector) ||
        current_selector != new_page) {
      
      isolate({  # This is the key fix to break feedback loops
        updateNumericInput(session, "page_selector", 
                          max = total_pages(),
                          value = new_page)
      })
    }
  })

  # Handle page selector input
  observeEvent(input$page_selector, {
    # Get current input value
    current_input <- input$page_selector
    
    # Check if input is valid (not NULL, not NA, and is a number)
    if (!is.null(current_input) && 
        !is.na(current_input) && 
        is.numeric(current_input) &&
        current_input > 0) {
      
      new_page <- max(1, min(current_input, total_pages()))
      
      # Only update page state if it's actually different
      if (new_page != page_state()) {
        page_state(new_page)
      }

      # Correct the input field if the user entered an invalid page number
      if (new_page != current_input) {
        updateNumericInput(session, "page_selector", value = new_page)
      }
    } else {
      # If input is invalid (empty, NA, or <= 0), reset to current page
      # Use a small delay to prevent immediate feedback loop
      invalidateLater(100)
      updateNumericInput(session, "page_selector", value = page_state())
    }
  }, ignoreInit = TRUE)  # Prevent firing on initialization

  # Handle next page button 
  observeEvent(input$next_page, {
    if (page_state() < total_pages()) {
      new_page <- page_state() + 1
      page_state(new_page)
    }
  })

  # Handle previous page button 
  observeEvent(input$prev_page, {
    if (page_state() > 1) {
      new_page <- page_state() - 1
      page_state(new_page)
    }
  })

  # Reset to first page on any filter/search change 
  observeEvent(list(sel_fhir_version(), sel_vendor(), sel_availability(), sel_is_chpl(), input$search_query), {
    page_state(1)
  })

  output$prev_button_ui <- renderUI({
    if (page_state() > 1) {
      actionButton(
        ns("prev_page"), 
        label = tagList(
          tags$i(class = "fa fa-arrow-left", style = "margin-right: 8px;"),
          "Previous"
        ),
        class = "modern-nav-button"
      )
    } else {
      NULL  # Hide the button
    }
  })

  output$next_button_ui <- renderUI({
    if (page_state() < total_pages()) {
      actionButton(
        ns("next_page"),
        label = tagList(
          "Next",
          tags$i(class = "fa fa-arrow-right", style = "margin-left: 8px;")
        ),
        class = "modern-nav-button"
      )
    } else {
      NULL  # Hide the button
    }
  })

  output$page_info <- renderText({
    paste("of", total_pages())
  })

  output$anchorlink <- renderUI({
    HTML("<p>You may also download endpoint data by visiting the <a tabindex=\"0\" id=\"downloads_page_link\" class=\"lantern-url\" style=\"color: #667eea; font-weight: 500;\">Downloads Page</a>.</p>")
  })

  # MATCHING ENDPOINTS: Count unique (url, fhir_version) combinations - the actual endpoints
  output$endpoint_count <- renderText({
    unique_endpoints <- nrow(selected_fhir_endpoints_without_limit() %>% distinct(url, fhir_version))
    paste("Matching Endpoints:", unique_endpoints)
  })

  # Main data query with LIMIT OFFSET pagination - WITH RACE CONDITION PROTECTION
  selected_fhir_endpoints <- reactive({
    req(sel_fhir_version(), sel_vendor(), sel_availability(), sel_is_chpl())
    
    # Generate unique request ID
    request_id <- isolate(current_request_id()) + 1
    current_request_id(request_id)
    
    offset <- (page_state() - 1) * page_size

    query_str <- "SELECT * FROM selected_fhir_endpoints_mv WHERE fhir_version IN ({vals*})"
    params <- list(vals = sel_fhir_version())

    if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
      query_str <- paste0(query_str, " AND vendor_name = {vendor}")
      params$vendor <- sel_vendor()
    }

    if (sel_is_chpl() != "All") {
      query_str <- paste0(query_str, " AND is_chpl = {chpl}")
      params$chpl <- toupper(sel_is_chpl())
    }

    if (sel_availability() != "0-100") {
      if (sel_availability() == "0" || sel_availability() == "100") {
        query_str <- paste0(query_str, " AND availability = {availability}")
        params$availability <- as.numeric(sel_availability())
      } else {
        availability_range <- strsplit(sel_availability(), "-")[[1]]
        query_str <- paste0(query_str, " AND availability BETWEEN {low} AND {high}")
        params$low <- as.numeric(availability_range[1])
        params$high <- as.numeric(availability_range[2])
      }
    }

    # Apply external search filter
    if (trimws(input$search_query) != "") {
      keyword <- tolower(trimws(input$search_query))
      query_str <- paste0(query_str, " AND (LOWER(url) LIKE {search} OR LOWER(condensed_endpoint_names) LIKE {search} OR LOWER(vendor_name) LIKE {search}")
      query_str <- paste0(query_str, " OR LOWER(capability_fhir_version) LIKE {search} OR LOWER(format) LIKE {search} OR LOWER(cap_stat_exists) LIKE {search}")
      query_str <- paste0(query_str, " OR LOWER(status) LIKE {search} OR LOWER(availability::TEXT) LIKE {search})")
      params$search <- paste0("%", keyword, "%")
    }

    # Add LIMIT OFFSET for pagination
    query_str <- paste0(query_str, " LIMIT {limit} OFFSET {offset}")
    params$limit <- page_size
    params$offset <- offset

    query <- do.call(glue_sql, c(list(query_str, .con = db_connection), params))
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

  # Query without limit for total count and download
  selected_fhir_endpoints_without_limit <- reactive({
    req(sel_fhir_version(), sel_vendor(), sel_availability(), sel_is_chpl())
    
    query_str <- "SELECT * FROM selected_fhir_endpoints_mv WHERE fhir_version IN ({vals*})"
    params <- list(vals = sel_fhir_version())

    if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
      query_str <- paste0(query_str, " AND vendor_name = {vendor}")
      params$vendor <- sel_vendor()
    }

    if (sel_is_chpl() != "All") {
      query_str <- paste0(query_str, " AND is_chpl = {chpl}")
      params$chpl <- toupper(sel_is_chpl())
    }

    if (sel_availability() != "0-100") {
      if (sel_availability() == "0" || sel_availability() == "100") {
        query_str <- paste0(query_str, " AND availability = {availability}")
        params$availability <- as.numeric(sel_availability())
      } else {
        availability_range <- strsplit(sel_availability(), "-")[[1]]
        query_str <- paste0(query_str, " AND availability BETWEEN {low} AND {high}")
        params$low <- as.numeric(availability_range[1])
        params$high <- as.numeric(availability_range[2])
      }
    }

    # Apply external search filter
    if (trimws(input$search_query) != "") {
      keyword <- tolower(trimws(input$search_query))
      query_str <- paste0(query_str, " AND (LOWER(url) LIKE {search} OR LOWER(condensed_endpoint_names) LIKE {search} OR LOWER(vendor_name) LIKE {search}")
      query_str <- paste0(query_str, " OR LOWER(capability_fhir_version) LIKE {search} OR LOWER(format) LIKE {search} OR LOWER(cap_stat_exists) LIKE {search}")
      query_str <- paste0(query_str, " OR LOWER(status) LIKE {search} OR LOWER(availability::TEXT) LIKE {search})")
      params$search <- paste0("%", keyword, "%")
    }

    # Add ordering by vendor name
    query_str <- paste0(query_str, " ORDER BY vendor_name, list_source, url, requested_fhir_version")

    query <- do.call(glue_sql, c(list(query_str, .con = db_connection), params))
    res <- tbl(db_connection, sql(query)) %>% collect()
    res
  })

  # Downloadable csv of selected dataset
  output$download_data <- downloadHandler(
    filename = function() {
      "fhir_endpoints.csv"
    },
    content = function(file) {
      write.csv(csv_format(), file, row.names = FALSE)
    }
  )

  # Download csv of the field descriptions in the dataset csv
  output$download_descriptions <- downloadHandler(
    filename = function() {
      "fhir_endpoints_fields.csv"
    },
    content = function(file) {
      file.copy("fhir_endpoints_fields.csv", file)
    }
  )

  output$endpoints_table <- reactable::renderReactable({
    data <- selected_fhir_endpoints() %>% 
      select(urlModal, condensed_endpoint_names, endpoint_names, vendor_name, capability_fhir_version, format, cap_stat_exists, status, availability) %>% 
      distinct(urlModal, condensed_endpoint_names, endpoint_names, vendor_name, capability_fhir_version, format, cap_stat_exists, status, availability) %>% 
      group_by(urlModal) %>% 
      mutate_at(vars(-group_cols()), as.character)
    
    reactable(
      data,
      defaultColDef = colDef(
        align = "center",
        headerStyle = list(
          background = "#f8f9fa",
          color = "#495057",
          fontWeight = "600",
          fontSize = "13px",
          textTransform = "uppercase",
          letterSpacing = "0.5px"
        )
      ),
      columns = list(
        urlModal = colDef(
          name = "URL", 
          minWidth = 300,
          style = JS("function(rowInfo, colInfo, state) {
            var prevRow = state.pageRows[rowInfo.viewIndex - 1]
            if (prevRow && rowInfo.row['urlModal'] === prevRow['urlModal']) {
              return { visibility: 'hidden' }
            }
          }"),
          sortable = TRUE,
          align = "left",
          html = TRUE
        ),
        endpoint_names = colDef(show = FALSE, sortable = TRUE),
        condensed_endpoint_names = colDef(
          name = "API Information Source Name", 
          minWidth = 200, 
          sortable = TRUE, 
          html = TRUE,
          style = list(fontWeight = "500")
        ),
        vendor_name = colDef(
          name = "Certified API Developer Name", 
          minWidth = 150, 
          sortable = TRUE,
          style = list(color = "#495057")
        ),
        capability_fhir_version = colDef(
          name = "FHIR Version", 
          sortable = TRUE,
          cell = function(value) {
            if (!is.na(value) && value != "") {
              div(class = "version-badge", value)
            } else {
              value
            }
          }
        ),
        format = colDef(
          name = "Supported Formats", 
          sortable = TRUE,
          style = list(fontSize = "13px")
        ),
        cap_stat_exists = colDef(
          name = "Capability Statement", 
          sortable = TRUE,
          cell = function(value) {
            if (grepl("true", tolower(value))) {
              div(class = "status-badge status-success", "Available")
            } else if (grepl("false", tolower(value))) {
              div(class = "status-badge status-error", "Not Available")
            } else {
              div(class = "status-badge status-info", value)
            }
          }
        ),
        status = colDef(
          name = "HTTP Response", 
          sortable = TRUE,
          cell = function(value) {
            badge_class <- if (grepl("200", value)) {
              "status-success"
            } else if (grepl("404|503", value)) {
              "status-error"
            } else if (grepl("^[45]", value)) {
              "status-warning"
            } else {
              "status-info"
            }
            div(class = paste("status-badge", badge_class), value)
          }
        ),
        availability = colDef(
          name = "Availability", 
          sortable = TRUE,
          cell = function(value) {
            # Convert to numeric and handle any edge cases
            availability_num <- suppressWarnings(as.numeric(value))
            if (is.na(availability_num)) availability_num <- 0
            
            # Color based on availability percentage
            fill_color <- if (availability_num >= 90) {
              "#28a745"  # Green
            } else if (availability_num >= 70) {
              "#ffc107"  # Yellow
            } else if (availability_num >= 50) {
              "#fd7e14"  # Orange
            } else {
              "#dc3545"  # Red
            }
            
            div(
              class = "availability-container",
              div(
                class = "availability-bar",
                div(
                  class = "availability-fill",
                  style = list(
                    width = paste0(availability_num, "%"),
                    background = fill_color
                  )
                )
              ),
              div(class = "availability-text", style = list(color = fill_color), paste0(availability_num, "%"))
            )
          }
        )
      ),
      searchable = FALSE,
      showSortIcon = TRUE,
      highlight = TRUE,
      striped = TRUE,
      bordered = FALSE,
      defaultPageSize = 10,
      showPageInfo = FALSE,
      paginationType = "simple"
    )
  })

  # Create the format for the csv
  csv_format <- reactive({
    res <- selected_fhir_endpoints_without_limit() %>%
      select(-id, -status, -availability, -fhir_version, -urlModal, -condensed_endpoint_names) %>%
      rowwise() %>%
      mutate(endpoint_names = ifelse(length(strsplit(endpoint_names, ";")[[1]]) > 100, paste0("Subset of Organizations, see Lantern Website for full list:", paste0(head(strsplit(endpoint_names, ";")[[1]], 100), collapse = ";")), endpoint_names),
             info_created = format(info_created, "%m/%d/%y %H:%M"),
             info_updated = format(info_updated, "%m/%d/%y %H:%M")) %>%
      ungroup() %>%
      rename(api_information_source_name = endpoint_names, certified_api_developer_name = vendor_name) %>%
      rename(created_at = info_created, updated = info_updated) %>%
      rename(http_response_time_second = response_time_seconds)
  })

  output$note_text <- renderUI({
    note_info <- "The endpoints queried by Lantern are limited to Fast Healthcare Interoperability
      Resources (FHIR) endpoints published publicly by Certified API Developers in conformance
      with the ONC Cures Act Final Rule. This data, therefore, may not represent all FHIR endpoints
      in existence. Insights gathered from this data should be framed accordingly."
    res <- paste("<div style='font-size: 16px; background: #f8f9fa; border-left: 4px solid #667eea; padding: 12px 16px; border-radius: 4px; margin: 16px 0;'><b>Note:</b>", note_info, "</div>")
    HTML(res)
  })

}