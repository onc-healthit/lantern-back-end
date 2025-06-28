library(DT)
library(purrr)
library(reactable)
library(leaflet)

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

  # Calculate total pages based on UNIQUE ORGANIZATION NAMES, not total rows
  org_total_pages <- reactive({
    fhir_versions <- sel_fhir_version()
    vendor <- sel_vendor()

    req(sel_fhir_version(), sel_vendor())

    # Use parameterized query for count as well
    count_query_str <- "
      SELECT COUNT(DISTINCT CASE 
        WHEN organization_name IS NULL OR organization_name = '' THEN 'Unknown'
        ELSE organization_name
      END) as count
      FROM mv_endpoint_list_organizations
      WHERE fhir_version IN ({fhir_versions*})"
    
    count_params <- list(fhir_versions = fhir_versions)

    # Add vendor filter
    if (vendor != ui_special_values$ALL_DEVELOPERS) {
      count_query_str <- paste0(count_query_str, " AND vendor_name = {vendor}")
      count_params$vendor <- vendor
    }

    # Add search filter if present
    search_term <- input$org_search_query
    if (!is.null(search_term) && search_term != "") {
      count_query_str <- paste0(count_query_str, " AND (
        organization_name ILIKE {search_pattern} OR 
        organization_id ILIKE {search_pattern} OR 
        fhir_version ILIKE {search_pattern} OR 
        vendor_name ILIKE {search_pattern})")
      count_params$search_pattern <- paste0("%", search_term, "%")
    }

    count_query <- do.call(glue_sql, c(list(count_query_str, .con = db_connection), count_params))
    count <- tbl(db_connection, sql(count_query)) %>% collect() %>% pull(count)
    max(1, ceiling(count / org_page_size))
  })

  # Handle next page button
  observeEvent(input$org_next_page, {
    message("NEXT PAGE BUTTON CLICKED")
    if (org_page_state() < org_total_pages()) {
      new_page <- org_page_state() + 1
      org_page_state(new_page)
    }
  })

  # Handle previous page button
  observeEvent(input$org_prev_page, {
    message("PREV PAGE BUTTON CLICKED")
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

  # Sync page selector
  observe({
    updateNumericInput(session, "org_page_selector",
                      max = org_total_pages(),
                      value = org_page_state())
  })

  # Manual page input
  observeEvent(input$org_page_selector, {
    if (!is.null(input$org_page_selector) && !is.na(input$org_page_selector)) {
      new_page <- max(1, min(input$org_page_selector, org_total_pages()))
      org_page_state(new_page)

      if (new_page != input$org_page_selector) {
        updateNumericInput(session, "org_page_selector", value = new_page)
      }
    }
})

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

  # Modified query to get organizations for pagination
  paged_endpoint_list_orgs <- reactive({
    current_fhir <- sel_fhir_version()
    current_vendor <- sel_vendor()

    req(current_fhir, current_vendor)

    limit <- org_page_size
    
    is_initial_load <- (
        all(sel_fhir_version() == ui_special_values$ALL_FHIR_VERSIONS) &&
        sel_vendor() == ui_special_values$ALL_DEVELOPERS &&
        (is.null(input$org_search_query) || input$org_search_query == "")
    )
 
    offset <- if (is_initial_load && org_page_state() == 1) {
      20  # Skip first 20 rows on very first load
    } else {
      (org_page_state() - 1) * org_page_size
    }

    # Build base query with parameterized approach
    query_str <- "
      SELECT DISTINCT 
        CASE 
          WHEN organization_name IS NULL OR organization_name = '' THEN 'Unknown'
          ELSE organization_name
        END AS organization_name
      FROM mv_endpoint_list_organizations
      WHERE fhir_version IN ({fhir_versions*})"
    
    params <- list(fhir_versions = current_fhir)

    # Add vendor filter using parameters
    if (current_vendor != ui_special_values$ALL_DEVELOPERS) {
      query_str <- paste0(query_str, " AND vendor_name = {vendor}")
      params$vendor <- current_vendor
    }

    # Add search filter if present
    search_term <- input$org_search_query
    if (!is.null(search_term) && search_term != "") {
      query_str <- paste0(query_str, " AND (
        organization_name ILIKE {search_pattern} OR 
        organization_id ILIKE {search_pattern} OR 
        fhir_version ILIKE {search_pattern} OR 
        vendor_name ILIKE {search_pattern})")
      params$search_pattern <- paste0("%", search_term, "%")
    }

    # Add ordering and pagination
    query_str <- paste0(query_str, " ORDER BY organization_name LIMIT {limit} OFFSET {offset}")
    params$limit <- limit
    params$offset <- offset

    # Execute first query to get organization names
    org_names_query <- do.call(glue_sql, c(list(query_str, .con = db_connection), params))
    org_names <- tbl(db_connection, sql(org_names_query)) %>% 
      collect() %>% 
      pull(organization_name)

    if (length(org_names) == 0) {
      return(data.frame())
    }

    # Second query to get all data for these organization names using parameters
    data_query_str <- "
      SELECT DISTINCT 
        CASE 
          WHEN organization_name IS NULL OR organization_name = '' THEN 'Unknown'
          ELSE organization_name
        END AS organization_name,
        organization_id,
        fhir_version,
        vendor_name
      FROM mv_endpoint_list_organizations
      WHERE fhir_version IN ({fhir_versions*})"
    
    data_params <- list(fhir_versions = current_fhir)

    # Add vendor filter
    if (current_vendor != ui_special_values$ALL_DEVELOPERS) {
      data_query_str <- paste0(data_query_str, " AND vendor_name = {vendor}")
      data_params$vendor <- current_vendor
    }

    # Add search filter if present
    if (!is.null(search_term) && search_term != "") {
      data_query_str <- paste0(data_query_str, " AND (
        organization_name ILIKE {search_pattern} OR 
        organization_id ILIKE {search_pattern} OR 
        fhir_version ILIKE {search_pattern} OR 
        vendor_name ILIKE {search_pattern})")
      data_params$search_pattern <- paste0("%", search_term, "%")
    }

    # Add organization names filter using parameters
    data_query_str <- paste0(data_query_str, " AND CASE 
      WHEN organization_name IS NULL OR organization_name = '' THEN 'Unknown'
      ELSE organization_name
    END IN ({org_names*}) ORDER BY organization_name")
    data_params$org_names <- org_names

    # Execute second query
    data_query <- do.call(glue_sql, c(list(data_query_str, .con = db_connection), data_params))
    res <- tbl(db_connection, sql(data_query)) %>% collect()

    res <- res %>%
      mutate(organization_id = as.integer(organization_id)) %>%
      
      # Left join with deduplicated or collapsed identifiers
      left_join(
        get_org_identifiers_information(db_connection) %>%
          mutate(org_id = as.integer(org_id)) %>%
          group_by(org_id) %>%
          summarise(identifier = paste(unique(identifier), collapse = "<br/>")),
        by = c("organization_id" = "org_id")
      ) %>%
      
      # Left join with deduplicated or collapsed addresses
      left_join(
        get_org_addresses_information(db_connection) %>%
          mutate(org_id = as.integer(org_id)) %>%
          group_by(org_id) %>%
          summarise(address = paste(unique(address), collapse = "<br/>")),
        by = c("organization_id" = "org_id")
      ) %>%
      
      left_join(get_org_url_information(db_connection),
          by = c("organization_id" = "org_id")) %>%
        
      select(-organization_id)

    res <- res %>%
      group_by(organization_name) %>%
      summarise(
        identifier = paste(unique(identifier), collapse = "<br/>"),
        address = paste(unique(address), collapse = "<br/>"),
        org_url = paste(unique(org_url), collapse = "<br/>"),
        fhir_version = paste(unique(fhir_version), collapse = "<br/>"),
        vendor_name = paste(unique(vendor_name), collapse = "<br/>"),
        .groups = "drop"
      ) %>%
      filter(organization_name != "Unknown") %>%
      mutate(address = toupper(address)) %>%
      arrange(organization_name)


    res
  })

  csv_format <- reactive({
      # Get current filter values
      current_fhir <- sel_fhir_version()
      current_vendor <- sel_vendor()

      req(current_fhir, current_vendor)

      # Get filtered data from the materialized view function
      res <- get_endpoint_list_matches(
        db_connection,
        fhir_version = current_fhir,
        vendor = current_vendor
      )

    res <- res %>%
      mutate(organization_id = as.integer(organization_id)) %>%
      
      # Left join with deduplicated or collapsed identifiers
      left_join(
        get_org_identifiers_information(db_connection) %>%
          mutate(org_id = as.integer(org_id)) %>%
          group_by(org_id) %>%
          summarise(identifier = paste(unique(identifier), collapse = "\n")),
        by = c("organization_id" = "org_id")
      ) %>%
      
      # Left join with deduplicated or collapsed addresses
      left_join(
        get_org_addresses_information(db_connection) %>%
          mutate(org_id = as.integer(org_id)) %>%
          group_by(org_id) %>%
          summarise(address = paste(unique(address), collapse = "\n")),
        by = c("organization_id" = "org_id")
      ) %>%
      
      left_join(get_org_url_information(db_connection),
          by = c("organization_id" = "org_id")) %>%
        
      select(-organization_id)

    res <- res %>%
      group_by(organization_name) %>%
      summarise(
        identifier = paste(unique(identifier), collapse = "\n"),
        address = paste(unique(address), collapse = "\n"),
        org_url = paste(unique(org_url), collapse = "\n"),
        fhir_version = paste(unique(fhir_version), collapse = "\n"),
        vendor_name = paste(unique(vendor_name), collapse = "\n"),
        .groups = "drop"
      ) %>%
      filter(organization_name != "Unknown") %>%
      mutate(address = toupper(address)) %>%
      arrange(organization_name)

    res
  })

  output$endpoint_list_orgs_table <- reactable::renderReactable({
     # Get all data
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
         org_url = colDef(name = "Organization URL", minWidth = 300, sortable = FALSE, html = TRUE),
         fhir_version = colDef(name = "FHIR Version", sortable = FALSE),
         vendor_name = colDef(name = "Certified API Developer Name", minWidth = 110, sortable = FALSE)
       ),
       striped = TRUE,
       searchable = FALSE,
       showSortIcon = TRUE,
       highlight = TRUE,
       pagination = FALSE,
       defaultExpanded = TRUE
     )
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
