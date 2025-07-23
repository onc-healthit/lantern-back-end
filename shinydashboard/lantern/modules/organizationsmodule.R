library(DT)
library(purrr)
library(reactable)
library(leaflet)
library(glue)

organizationsmodule_UI <- function(id) {
  ns <- NS(id)

  tagList(
    fluidRow(
      h2("Endpoint List Organizations")
    ),
    fluidRow(
      column(width = 12, style = "padding-bottom:20px",
             downloadButton(ns("download_data"), "Download Organization Data (CSV)", icon = tags$i(class = "fa fa-download", "aria-hidden" = "true", role = "presentation", "aria-label" = "download icon")),
             downloadButton(ns("download_descriptions"), "Download Field Descriptions (CSV)", icon = tags$i(class = "fa fa-download", "aria-hidden" = "true", role = "presentation", "aria-label" = "download icon"))
      ),
    ),
    fluidRow(
      column(6, 
        textInput(ns("org_search_query"), "Search Organizations")
      )
    ),
    fluidRow(
      p("This table shows the organization name listed for each endpoint in the endpoint list it appears in."),
      reactable::reactableOutput(ns("endpoint_list_orgs_table")),
      htmlOutput(ns("note_text"))
    ),
    fluidRow(
      column(3, 
        div(style = "display: flex; justify-content: flex-start;", 
            uiOutput(ns("org_prev_button_ui"))
        )
      ),
      column(6,
        div(style = "display: flex; justify-content: center; align-items: center; gap: 10px; margin-top: 8px;",
            numericInput(ns("org_page_selector"), label = NULL, value = 1, min = 1, step = 1, width = "80px"),
            textOutput(ns("org_page_info"), inline = TRUE)
        )
      ),
      column(3, 
        div(style = "display: flex; justify-content: flex-end;",
            uiOutput(ns("org_next_button_ui"))
        )
      )
    )
  )
}

organizationsmodule <- function(
  input,
  output,
  session,
  sel_fhir_version,
  sel_vendor,
  sel_confidence
) {
  ns <- session$ns

  org_page_state <- reactiveVal(1)
  org_page_size <- 10

  # Add request tracking to prevent race conditions
  current_request_id <- reactiveVal(0)

  # Helper function to determine if all FHIR versions are selected
  is_all_fhir_versions_selected <- reactive({
    current_selection <- sel_fhir_version()
    all_available <- app$distinct_fhir_version_list_no_capstat()
    
    # If either is NULL, we can't compare
    if (is.null(current_selection) || is.null(all_available)) {
      return(FALSE)
    }
    
    # Convert to vectors for comparison
    current_vec <- unlist(current_selection)
    all_vec <- unlist(all_available)
    
    # Check if current selection equals all available versions
    return(length(current_vec) == length(all_vec) && setequal(current_vec, all_vec))
  })

  # Calculate total pages using original array-based filtering
  org_total_pages <- reactive({
    fhir_versions <- sel_fhir_version()
    vendor <- sel_vendor()

    req(sel_fhir_version(), sel_vendor())

    # Build parameterized query for count using the materialized view
    count_query_str <- "SELECT COUNT(*) as count FROM mv_organizations_aggregated WHERE TRUE"
    count_params <- list()

    # Add FHIR version filter using array overlap 
    if (!is_all_fhir_versions_selected()) {
      count_query_str <- paste0(count_query_str, " AND fhir_versions_array && ARRAY[{fhir_versions*}]")
      count_params$fhir_versions <- fhir_versions
    }

    # Add vendor filter using array overlap 
    if (vendor != ui_special_values$ALL_DEVELOPERS) {
      count_query_str <- paste0(count_query_str, " AND vendor_names_array && ARRAY[{vendor}]")
      count_params$vendor <- vendor
    }

    # Add search filter if present
    search_term <- input$org_search_query
    if (!is.null(search_term) && search_term != "") {
      count_query_str <- paste0(count_query_str, " AND (
        organization_name ILIKE {search_pattern} OR 
        identifiers_html ILIKE {search_pattern} OR 
        addresses_html ILIKE {search_pattern} OR 
        endpoint_urls_html ILIKE {search_pattern} OR 
        fhir_versions_html ILIKE {search_pattern} OR 
        vendor_names_html ILIKE {search_pattern})")
      count_params$search_pattern <- paste0("%", search_term, "%")
    }

    # Execute count query
    if (length(count_params) > 0) {
      count_query <- do.call(glue_sql, c(list(count_query_str, .con = db_connection), count_params))
    } else {
      count_query <- glue_sql(count_query_str, .con = db_connection)
    }
    
    count <- tbl(db_connection, sql(count_query)) %>% collect() %>% pull(count)
    max(1, ceiling(count / org_page_size))
  })

  # Handle next page button
  observeEvent(input$org_next_page, {
    if (org_page_state() < org_total_pages()) {
      new_page <- org_page_state() + 1
      org_page_state(new_page)
    }
  })

  # Handle previous page button
  observeEvent(input$org_prev_page, {
    if (org_page_state() > 1) {
      new_page <- org_page_state() - 1
      org_page_state(new_page)
    }
  })

  # Reset to first page on any filter/search change 
  observeEvent(list(sel_fhir_version(), sel_vendor(), sel_confidence(), input$org_search_query), {
    org_page_state(1)
    updateNumericInput(session, "org_page_selector", value = 1)
  })

  # Break the feedback loop with isolate()
  observe({
    new_page <- org_page_state()
    current_selector <- input$org_page_selector
    
    # Only update if different (prevents infinite loop)
    # Add safety check for current_selector to prevent crashes
    if (is.null(current_selector) || 
        is.na(current_selector) || 
        !is.numeric(current_selector) ||
        current_selector != new_page) {
      
      isolate({  # This is the key fix to break feedback loops!
        updateNumericInput(session, "org_page_selector",
                          max = org_total_pages(),
                          value = new_page)
      })
    }
  })

  # Manual page input
  observeEvent(input$org_page_selector, {
    # Get current input value
    current_input <- input$org_page_selector
    
    # Check if input is valid (not NULL, not NA, and is a number)
    if (!is.null(current_input) && 
        !is.na(current_input) && 
        is.numeric(current_input) &&
        current_input > 0) {
      
      new_page <- max(1, min(current_input, org_total_pages()))
      
      # Only update page state if it's actually different
      if (new_page != org_page_state()) {
        org_page_state(new_page)
      }

      # Correct the input field if the user entered an invalid page number
      if (new_page != current_input) {
        updateNumericInput(session, "org_page_selector", value = new_page)
      }
    } else {
      # If input is invalid (empty, NA, or <= 0), reset to current page
      # Use a small delay to prevent immediate feedback loop
      invalidateLater(100)
      updateNumericInput(session, "org_page_selector", value = org_page_state())
    }
  }, ignoreInit = TRUE)  # Prevent observer from firing on initialization or first load 

  # Data fetching with race condition protection
  paged_endpoint_list_orgs <- reactive({
    current_fhir <- sel_fhir_version()
    current_vendor <- sel_vendor()

    req(current_fhir, current_vendor)

    # Generate unique request ID
    request_id <- isolate(current_request_id()) + 1
    current_request_id(request_id)

    limit <- org_page_size
    
    is_initial_load <- (
        is_all_fhir_versions_selected() &&
        sel_vendor() == ui_special_values$ALL_DEVELOPERS &&
        (is.null(input$org_search_query) || input$org_search_query == "")
    )
 
    offset <- if (is_initial_load && org_page_state() == 1) {
      20  # Skip first 20 rows on very first load
    } else {
      (org_page_state() - 1) * org_page_size
    }

    # Build query that constructs filtered HTML based on selected filters
    query_str <- "
      WITH base_data AS (
        SELECT 
          organization_name,
          identifiers_html as identifier,
          addresses_html as address,
          org_urls_html as org_url,
          endpoint_urls_html as url,
          fhir_versions_array,
          vendor_names_array
        FROM mv_organizations_aggregated 
        WHERE TRUE"
    
    params <- list()

    # Add FHIR version filter using array overlap
    if (!is_all_fhir_versions_selected()) {
      query_str <- paste0(query_str, " AND fhir_versions_array && ARRAY[{fhir_versions*}]")
      params$fhir_versions <- current_fhir
    }

    # Add vendor filter using array overlap
    if (current_vendor != ui_special_values$ALL_DEVELOPERS) {
      query_str <- paste0(query_str, " AND vendor_names_array && ARRAY[{vendor}]")
      params$vendor <- current_vendor
    }

    # Add search filter if present
    search_term <- input$org_search_query
    if (!is.null(search_term) && search_term != "") {
      query_str <- paste0(query_str, " AND (
        organization_name ILIKE {search_pattern} OR 
        identifiers_html ILIKE {search_pattern} OR 
        addresses_html ILIKE {search_pattern} OR 
        endpoint_urls_html ILIKE {search_pattern} OR 
        fhir_versions_html ILIKE {search_pattern} OR 
        vendor_names_html ILIKE {search_pattern})")
      params$search_pattern <- paste0("%", search_term, "%")
    }

    # Close the base_data CTE and add the filtered aggregation
    query_str <- paste0(query_str, "
      )
      SELECT 
        organization_name,
        identifier,
        address,
        org_url,
        url,
        -- Only show FHIR versions that match the current filter
        string_agg(
          DISTINCT fhir_version, 
          '<br/>'
        ) as fhir_version,
        -- Only show vendor names that match the current filter  
        string_agg(
          DISTINCT vendor_name,
          '<br/>'
        ) as vendor_name
      FROM base_data bd
      CROSS JOIN LATERAL unnest(bd.fhir_versions_array) AS fhir_version
      CROSS JOIN LATERAL unnest(bd.vendor_names_array) AS vendor_name
      WHERE 1=1")

    # Apply the same filters to the individual FHIR versions and vendors
    if (!is_all_fhir_versions_selected()) {
      query_str <- paste0(query_str, " AND fhir_version = ANY(ARRAY[{fhir_versions_display*}])")
      params$fhir_versions_display <- current_fhir
    }

    if (current_vendor != ui_special_values$ALL_DEVELOPERS) {
      query_str <- paste0(query_str, " AND vendor_name = {vendor_display}")
      params$vendor_display <- current_vendor
    }

    # Add GROUP BY, ordering and pagination
    query_str <- paste0(query_str, " 
      GROUP BY organization_name, identifier, address, org_url, url
      ORDER BY organization_name 
      LIMIT {limit} OFFSET {offset}")
    params$limit <- limit
    params$offset <- offset

    # Execute the optimized query
    if (length(params) > 0) {
      data_query <- do.call(glue_sql, c(list(query_str, .con = db_connection), params))
    } else {
      data_query <- glue_sql(query_str, .con = db_connection)
    }
    
    # Execute query
    result <- tbl(db_connection, sql(data_query)) %>% collect()
    
    # Only return results if this is still the latest request
    # Use isolate() to check without creating reactive dependency
    if (request_id == isolate(current_request_id())) {
      # This is the latest request, process normally
      result <- result %>%
        mutate(
          # Convert NA to empty string 
          org_url = case_when(
            is.na(org_url) | org_url == "NA" ~ "",
            TRUE ~ org_url
          ),
          identifier = case_when(
            is.na(identifier) ~ "",
            TRUE ~ identifier
          ),
          address = case_when(
            is.na(address) ~ "",
            TRUE ~ address
          )
        )
      return(result)
    } else {
      # This request was superseded, return empty to avoid flicker
      return(data.frame())
    }
  })

  # CSV format using filtered approach
  csv_format <- reactive({
    current_fhir <- sel_fhir_version()
    current_vendor <- sel_vendor()

    req(current_fhir, current_vendor)

    # Build query for CSV export using the same filtering logic
    query_str <- "
      WITH base_data AS (
        SELECT 
          organization_name,
          identifiers_csv as identifier,
          addresses_csv as address,
          org_urls_csv as org_url,
          endpoint_urls_csv as url,
          fhir_versions_array,
          vendor_names_array,
          -- Include HTML fields for search functionality
          identifiers_html,
          addresses_html,
          endpoint_urls_html,
          fhir_versions_html,
          vendor_names_html
        FROM mv_organizations_aggregated 
        WHERE TRUE"
    
    params <- list()

    # Add FHIR version filter using array overlap
    if (!is_all_fhir_versions_selected()) {
      query_str <- paste0(query_str, " AND fhir_versions_array && ARRAY[{fhir_versions*}]")
      params$fhir_versions <- current_fhir
    }

    # Add vendor filter using array overlap
    if (current_vendor != ui_special_values$ALL_DEVELOPERS) {
      query_str <- paste0(query_str, " AND vendor_names_array && ARRAY[{vendor}]")
      params$vendor <- current_vendor
    }

    # Add search filter if present (same logic as pagination and count)
    search_term <- input$org_search_query
    if (!is.null(search_term) && search_term != "") {
      query_str <- paste0(query_str, " AND (
        organization_name ILIKE {search_pattern} OR 
        identifiers_html ILIKE {search_pattern} OR 
        addresses_html ILIKE {search_pattern} OR 
        endpoint_urls_html ILIKE {search_pattern} OR 
        fhir_versions_html ILIKE {search_pattern} OR 
        vendor_names_html ILIKE {search_pattern})")
      params$search_pattern <- paste0("%", search_term, "%")
    }

    # Close the base_data CTE and add the filtered aggregation
    query_str <- paste0(query_str, "
      )
      SELECT 
        organization_name,
        identifier,
        address,
        org_url,
        url,
        -- Only show FHIR versions that match the current filter (CSV format)
        string_agg(
          DISTINCT fhir_version, 
          E'\\n'
        ) as fhir_version,
        -- Only show vendor names that match the current filter (CSV format)
        string_agg(
          DISTINCT vendor_name,
          E'\\n'
        ) as vendor_name
      FROM base_data bd
      CROSS JOIN LATERAL unnest(bd.fhir_versions_array) AS fhir_version
      CROSS JOIN LATERAL unnest(bd.vendor_names_array) AS vendor_name
      WHERE 1=1")

    # Apply the same filters to the individual FHIR versions and vendors
    if (!is_all_fhir_versions_selected()) {
      query_str <- paste0(query_str, " AND fhir_version = ANY(ARRAY[{fhir_versions_display*}])")
      params$fhir_versions_display <- current_fhir
    }

    if (current_vendor != ui_special_values$ALL_DEVELOPERS) {
      query_str <- paste0(query_str, " AND vendor_name = {vendor_display}")
      params$vendor_display <- current_vendor
    }

    # Add GROUP BY and ordering
    query_str <- paste0(query_str, " 
      GROUP BY organization_name, identifier, address, org_url, url
      ORDER BY organization_name")

    # Execute query
    if (length(params) > 0) {
      data_query <- do.call(glue_sql, c(list(query_str, .con = db_connection), params))
    } else {
      data_query <- glue_sql(query_str, .con = db_connection)
    }
    
    res <- tbl(db_connection, sql(data_query)) %>% collect()
    
    # Handle empty fields gracefully for CSV - leave them empty
    res <- res %>%
      mutate(
        # Convert NA to empty string 
        org_url = case_when(
          is.na(org_url) | org_url == "NA" ~ "",
          TRUE ~ org_url
        ),
        identifier = case_when(
          is.na(identifier) ~ "",
          TRUE ~ identifier
        ),
        address = case_when(
          is.na(address) ~ "",
          TRUE ~ address
        )
      )

    return(res)
  })

  # Reactable output
  output$endpoint_list_orgs_table <- reactable::renderReactable({
     display_data <- paged_endpoint_list_orgs()

     if (nrow(display_data) == 0) {
       return(
         reactable(
           data.frame(Message = "No data matching the selected filters"),
           pagination = FALSE,
           searchable = FALSE
         )
       )
     }

     reactable(
       display_data,
       defaultColDef = colDef(
         align = "center"
       ),
       columns = list(
         organization_name = colDef(name = "Organization Name", sortable = TRUE, align = "left",
                                    grouped = JS("function(cellInfo) {return cellInfo.value}")),
         identifier = colDef(name = "Organization Identifiers", minWidth = 300, sortable = FALSE, html = TRUE),
         address = colDef(name = "Organization Addresses", minWidth = 300, sortable = FALSE, html = TRUE),
         url = colDef(name = "FHIR Endpoint URL", minWidth = 300, sortable = FALSE, html = TRUE),
         org_url = colDef(name = "Organization URL", minWidth = 300, sortable = FALSE, html = TRUE),
         fhir_version = colDef(name = "FHIR Version", sortable = FALSE, html = TRUE),
         vendor_name = colDef(name = "Certified API Developer Name", minWidth = 110, sortable = FALSE, html = TRUE)
       ),
       striped = TRUE,
       searchable = FALSE,
       showSortIcon = TRUE,
       highlight = TRUE,
       pagination = FALSE,
       defaultExpanded = TRUE
     )
   })

  # Button UI outputs
  output$org_prev_button_ui <- renderUI({
    if (org_page_state() > 1) {
      actionButton(ns("org_prev_page"), "Previous", icon = icon("arrow-left"))
    } else {
      NULL  # Hide the button
    }
  })

  output$org_next_button_ui <- renderUI({
    if (org_page_state() < org_total_pages()) {
      actionButton(ns("org_next_page"), "Next", icon = icon("arrow-right"))
    } else {
      NULL  # Hide the button
    }
  })

  output$org_page_info <- renderText({
    paste("of", org_total_pages())
  })

  # Downloadable csv of selected dataset
  output$download_data <- downloadHandler(
    filename = function() {
      "fhir_endpoint_organizations.csv"
    },
    content = function(file) {
      write.csv(csv_format(), file, row.names = FALSE)
    }
  )

  # Download csv of the field descriptions in the dataset csv
  output$download_descriptions <- downloadHandler(
    filename = function() {
      "fhir_endpoint_organizations_fields.csv"
    },
    content = function(file) {
      file.copy("fhir_endpoint_organizations_fields.csv", file)
    }
  )

  output$note_text <- renderUI({
    note_info <- "The endpoints queried by Lantern are limited to Fast Healthcare Interoperability
      Resources (FHIR) endpoints published publicly by Certified API Developers in conformance
      with the ONC Cures Act Final Rule. This data, therefore, may not represent all FHIR endpoints
      in existence. Insights gathered from this data should be framed accordingly."
    res <- paste("<div style='font-size: 18px;'><b>Note:</b>", note_info, "</div>")
    HTML(res)
  })
}
