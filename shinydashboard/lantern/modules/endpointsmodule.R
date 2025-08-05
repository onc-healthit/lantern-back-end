library(DT)
library(purrr)
library(reactable)
library(glue)

endpointsmodule_UI <- function(id) {
  
  ns <- NS(id)
  
  tagList(
    fluidRow(
      column(width = 12, style = "padding-bottom:20px",
             h2(style = "margin-top:0", textOutput(ns("endpoint_count"))),
             downloadButton(ns("download_data"), "Download Endpoint Data (CSV)", icon = tags$i(class = "fa fa-download", "aria-hidden" = "true", role = "presentation", "aria-label" = "download icon")),
             downloadButton(ns("download_descriptions"), "Download Field Descriptions (CSV)", icon = tags$i(class = "fa fa-download", "aria-hidden" = "true", role = "presentation", "aria-label" = "download icon")),
             htmlOutput(ns("anchorlink"))
      )
    ),
    tags$p("The URL for each endpoint in the table below can be clicked on to see additional information for that individual endpoint.", role = "comment"),
    fluidRow(
      column(width = 6, textInput(ns("search_query"), "Search:", value = "")
      )
    ),
    reactable::reactableOutput(ns("endpoints_table")),
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
    ),
    tags$p("* An asterisk after a 'true' value in the 'Capability Statement Returned' field indicates that the returned Capability Statement for the endpoint is not of kind 'instance', which is the kind Lantern expects.", role = "comment"),
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

  # Calculate total pages based on ACTUAL TABLE ROWS (after distinct operation)
  total_pages <- reactive({
    # Count the actual distinct rows that will be displayed in the table
    table_data <- selected_fhir_endpoints_without_limit() %>% 
      select(urlModal, condensed_endpoint_names, endpoint_names, vendor_name, capability_fhir_version, format, cap_stat_exists, status, availability) %>% 
      distinct(urlModal, condensed_endpoint_names, endpoint_names, vendor_name, capability_fhir_version, format, cap_stat_exists, status, availability)
    
    total_records <- nrow(table_data)
    max(1, ceiling(total_records / page_size))
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

  # Handle next page button
  observeEvent(input$next_page, {
    current_time <- as.numeric(Sys.time()) * 1000
    if (!is.null(session$userData$last_next_time) && 
        (current_time - session$userData$last_next_time) < 300) {
      return()  # Ignore only rapid consecutive clicks
    }
    session$userData$last_next_time <- current_time
    if (page_state() < total_pages()) {
      new_page <- page_state() + 1
      page_state(new_page)
      updateNumericInput(session, "page_selector", value = new_page)
    }
  })

  # Handle previous page button
  observeEvent(input$prev_page, {
    current_time <- as.numeric(Sys.time()) * 1000
    if (!is.null(session$userData$last_prev_time) && 
        (current_time - session$userData$last_prev_time) < 300) {
      return()  # Ignore only rapid consecutive clicks
    }
    session$userData$last_prev_time <- current_time
    if (page_state() > 1) {
      new_page <- page_state() - 1
      page_state(new_page)
      updateNumericInput(session, "page_selector", value = new_page)
    }
  })

  # Reset to first page on any filter/search change 
  observeEvent(list(sel_fhir_version(), sel_vendor(), sel_availability(), sel_is_chpl(), input$search_query), {
    page_state(1)
    updateNumericInput(session, "page_selector", value = 1)
  })

  output$prev_button_ui <- renderUI({
    if (page_state() > 1) {
      actionButton(ns("prev_page"), "Previous", icon = icon("arrow-left"))
    } else {
      NULL  # Hide the button
    }
  })

  output$next_button_ui <- renderUI({
    if (page_state() < total_pages()) {
      actionButton(ns("next_page"), "Next", icon = icon("arrow-right"))
    } else {
      NULL  # Hide the button
    }
  })

  output$page_info <- renderText({
    paste("of", total_pages())
  })

  output$anchorlink <- renderUI({
    HTML("<p>You may also download endpoint data by visiting the <a tabindex=\"0\" id=\"downloads_page_link\" class=\"lantern-url\">Downloads Page</a>.</p>")
  })

  # MATCHING ENDPOINTS: Count unique (url, fhir_version) combinations - the actual endpoints
  output$endpoint_count <- renderText({
    unique_endpoints <- nrow(selected_fhir_endpoints_without_limit() %>% distinct(url, fhir_version))
    paste("Matching Endpoints:", unique_endpoints)
  })

  # Main data query with LIMIT OFFSET pagination
  selected_fhir_endpoints <- reactive({
    req(sel_fhir_version(), sel_vendor(), sel_availability(), sel_is_chpl())
    
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
    res <- tbl(db_connection, sql(query)) %>% collect()
    res
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
    query_str <- paste0(query_str, " ORDER BY vendor_name")

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
     reactable(
              selected_fhir_endpoints() %>% select(urlModal, condensed_endpoint_names, endpoint_names, vendor_name, capability_fhir_version, format, cap_stat_exists, status, availability) %>% distinct(urlModal, condensed_endpoint_names, endpoint_names, vendor_name, capability_fhir_version, format, cap_stat_exists, status, availability) %>% group_by(urlModal) %>% mutate_at(vars(-group_cols()), as.character),
              defaultColDef = colDef(
                align = "center"
              ),
              columns = list(
                  urlModal = colDef(name = "URL", minWidth = 300,
                            style = JS("function(rowInfo, colInfo, state) {
                                    var prevRow = state.pageRows[rowInfo.viewIndex - 1]
                                    if (prevRow && rowInfo.row['urlModal'] === prevRow['urlModal']) {
                                      return { visibility: 'hidden' }
                                    }
                                  }"
                            ),
                            sortable = TRUE,
                            align = "left",
                            html = TRUE),
                  endpoint_names = colDef(show = FALSE, sortable = TRUE),
                  condensed_endpoint_names = colDef(name = "API Information Source Name", minWidth = 200, sortable = TRUE, html = TRUE),
                  vendor_name = colDef(name = "Certified API Developer Name", minWidth = 110, sortable = TRUE),
                  capability_fhir_version = colDef(name = "FHIR Version", sortable = TRUE),
                  format = colDef(name = "Supported Formats", sortable = TRUE),
                  cap_stat_exists = colDef(name = "Capability Statement Returned", sortable = TRUE),
                  status = colDef(name = "HTTP Response", sortable = TRUE),
                  availability = colDef(name = "Availability", sortable = TRUE)
              ),
              searchable = FALSE,
              showSortIcon = TRUE,
              highlight = TRUE,
              defaultPageSize = 10
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
    res <- paste("<div style='font-size: 18px;'><b>Note:</b>", note_info, "</div>")
    HTML(res)
  })

}
