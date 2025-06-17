# SMART-on-FHIR Well-known URI responses
library(glue)

smartresponsemodule_UI <- function(id) {

  ns <- NS(id)

  tagList(
    p("This is the SMART-on-FHIR Core Capabilities response page. FHIR endpoints
      requiring authorization shall provide a JSON document at the endpoint URL with ",
      code("/.well-known/smart-configuration"), "appended to the end of the base URL."
    ),
    fluidRow(
      column(width = 6,
             tableOutput(ns("well_known_summary_table")),
             tableOutput(ns("smart_vendor_table"))
      ),
      column(width = 6,
             tableOutput(ns("smart_capability_count_table")))
    ),
    h2("Endpoints by Well Known URI support"),
    p("This is the list of endpoints which have returned a valid SMART Core Capabilities JSON document at the", code("/.well-known/smart-configuration"), " URI."),
    tags$p("The URL for each endpoint in the table below can be clicked on to see additional information for that individual endpoint.", role = "comment"),
    reactable::reactableOutput(ns("well_known_endpoints")),
    fluidRow(
      column(6, 
        div(style = "display: flex; justify-content: flex-start;", 
            uiOutput(ns("smartres_prev_button_ui"))
        )
      ),
      column(6, 
        div(style = "display: flex; justify-content: flex-end;",
            uiOutput(ns("smartres_next_button_ui"))
        )
      )
    )
  )
}

get_selected_smart_count_total <- function(db_connection, fhir_version, vendor) {
  query <- tbl(db_connection, "mv_endpoint_export_tbl")
  
  # Apply filters in SQL
  if (!is.null(fhir_version) && length(fhir_version) > 0) {
    query <- query %>% filter(fhir_version %in% !!fhir_version)
  }
  
  if (!is.null(vendor) && vendor != ui_special_values$ALL_DEVELOPERS) {
    query <- query %>% filter(vendor_name == !!vendor)
  }
  
  # Perform distinct and count in SQL before collecting
  result <- query %>%
    distinct(url) %>%
    summarise(n = n()) %>%
    collect() %>%
    pull(n)
  
  return(result)
}

get_selected_smart_count_200 <- function(db_connection, fhir_version, vendor) {
  query <- tbl(db_connection, "mv_http_pct")
  
  # Apply filtering on fhir_version
  if (!is.null(fhir_version) && length(fhir_version) > 0) {
    query <- query %>% filter(fhir_version %in% !!fhir_version)
  }
  
  # Apply filtering on vendor
  if (!is.null(vendor) && vendor != ui_special_values$ALL_DEVELOPERS) {
    query <- query %>% filter(vendor_name == !!vendor)
  }
  
  # Filter for HTTP 200 responses
  query <- query %>% filter(http_response == 200)
  
  # Count the filtered rows directly in SQL
  result <- query %>%
    summarise(n = n()) %>%
    collect() %>%
    pull(n)
  
  # Return 0 if no rows were found
  if (length(result) == 0) {
    return(0)
  } else {
    return(result)
  }
}

get_selected_well_known_endpoints_count <- function(db_connection, fhir_version, vendor) {
  query <- tbl(db_connection, "mv_endpoint_export_tbl")
  
  # Apply filtering on fhir_version
  if (!is.null(fhir_version) && length(fhir_version) > 0) {
    query <- query %>% filter(fhir_version %in% !!fhir_version)
  }
  
  # Apply filtering on vendor
  if (!is.null(vendor) && vendor != ui_special_values$ALL_DEVELOPERS) {
    query <- query %>% filter(vendor_name == !!vendor)
  }
  
  # Filter for smart endpoints with HTTP response 200
  query <- query %>% filter(smart_http_response == 200)
  
  # Count distinct URLs in SQL before collecting the result
  result <- query %>%
    distinct(url) %>%
    summarise(n = n()) %>%
    collect() %>%
    pull(n)
  
  if (length(result) == 0) {
    return(0)
  } else {
    return(result)
  }
}

get_selected_well_known_count_doc <- function(db_connection, fhir_version, vendor) {
  query <- tbl(db_connection, "mv_well_known_endpoints")
  
  # Apply filtering on fhir_version
  if (!is.null(fhir_version) && length(fhir_version) > 0) {
    query <- query %>% filter(fhir_version %in% !!fhir_version)
  }
  
  # Apply filtering on vendor
  if (!is.null(vendor) && vendor != ui_special_values$ALL_DEVELOPERS) {
    query <- query %>% filter(vendor_name == !!vendor)
  }
  
  # Count the rows in SQL before collecting the result
  result <- query %>%
    summarise(n = n()) %>%
    collect() %>%
    pull(n)
  
  if (length(result) == 0) {
    return(0)
  } else {
    return(result)
  }
}

get_selected_well_known_count_no_doc <- function(db_connection, fhir_version, vendor) {
  query <- tbl(db_connection, "mv_well_known_no_doc")
  
  # Apply filtering on fhir_version
  if (!is.null(fhir_version) && length(fhir_version) > 0) {
    query <- query %>% filter(fhir_version %in% !!fhir_version)
  }
  
  # Apply filtering on vendor
  if (!is.null(vendor) && vendor != ui_special_values$ALL_DEVELOPERS) {
    query <- query %>% filter(vendor_name == !!vendor)
  }
  
  # Count the rows in SQL before collecting the result
  result <- query %>%
    summarise(n = n()) %>%
    collect() %>%
    pull(n)
  
  if (length(result) == 0) {
    return(0)
  } else {
    return(result)
  }
}

# Summarize the count of capabilities reported in SMART Core Capabilities JSON doc
get_smart_response_capability_count <- function(db_connection, fhir_version, vendor) {
  query <- tbl(db_connection, "mv_smart_response_capabilities")
  
  # Apply filtering on fhir_version
  if (!is.null(fhir_version) && length(fhir_version) > 0) {
    query <- query %>% filter(fhir_version %in% !!fhir_version)
  }
  
  # Apply filtering on vendor
  if (!is.null(vendor) && vendor != ui_special_values$ALL_DEVELOPERS) {
    query <- query %>% filter(vendor_name == !!vendor)
  }
  
  # Group by fhir_version and capability, count the rows, and rename columns in SQL
  result <- query %>%
    group_by(fhir_version, capability) %>%
    summarise(n = n(), .groups = "drop") %>%
    rename("FHIR Version" = fhir_version, Capability = capability, Endpoints = n) %>%
    collect()
  
  result
}

get_smart_vendor_table <- function(db_connection, fhir_version, vendor) {
  query <- tbl(db_connection, "mv_smart_response_capabilities")
  
  # Apply filtering on fhir_version
  if (!is.null(fhir_version) && length(fhir_version) > 0) {
    query <- query %>% filter(fhir_version %in% !!fhir_version)
  }
  
  # Apply filtering on vendor
  if (!is.null(vendor) && vendor != ui_special_values$ALL_DEVELOPERS) {
    query <- query %>% filter(vendor_name == !!vendor)
  }
  
  # Group by FHIR version and vendor, and count distinct IDs in SQL
  result <- query %>%
    group_by(fhir_version, vendor_name) %>%
    summarise(Endpoints = n_distinct(id), .groups = "drop") %>%
    rename("FHIR Version" = fhir_version, "Developer" = vendor_name) %>%
    collect()
  
  result
}

# Query fhir endpoints and return list of endpoints that have
# returned a valid JSON document at /.well-known/smart-configuration
# This implies a smart_http_response of 200.
#
get_well_known_endpoints_tbl <- function(db_connection) {
  tbl(db_connection, "mv_well_known_endpoints") %>% collect()
}

# Modified function to support LIMIT OFFSET pagination
get_selected_endpoints <- function(db_connection, fhir_version, vendor, limit = NULL, offset = NULL) {
  # Build SQL query with glue_sql 
  query_str <- "SELECT url, condensed_organization_names, vendor_name, capability_fhir_version FROM mv_selected_endpoints WHERE 1=1"
  params <- list()
  
  # Apply filtering on fhir_version
  if (!is.null(fhir_version) && length(fhir_version) > 0) {
    query_str <- paste0(query_str, " AND capability_fhir_version IN ({vals*})")
    params$vals <- fhir_version
  }
  
  # Apply filtering on vendor
  if (!is.null(vendor) && vendor != ui_special_values$ALL_DEVELOPERS) {
    query_str <- paste0(query_str, " AND vendor_name = {vendor}")
    params$vendor <- vendor
  }
  
  # Add LIMIT OFFSET if provided
  if (!is.null(limit) && !is.null(offset)) {
    query_str <- paste0(query_str, " LIMIT {limit} OFFSET {offset}")
    params$limit <- limit
    params$offset <- offset
  }
  
  # Execute query
  query <- do.call(glue_sql, c(list(query_str, .con = db_connection), params))
  result <- tbl(db_connection, sql(query)) %>% collect()
  
  result
}

smartresponsemodule <- function(
  input,
  output,
  session,
  sel_fhir_version,
  sel_vendor
) {

  ns <- session$ns

  smartres_page_state <- reactiveVal(1)
  smartres_page_size <- 10

  # Handle next page button
  observeEvent(input$smartres_next_page, {
    new_page <- smartres_page_state() + 1
    smartres_page_state(new_page)
  })

  # Handle previous page button
  observeEvent(input$smartres_prev_page, {
    if (smartres_page_state() > 1) {
      new_page <- smartres_page_state() - 1
      smartres_page_state(new_page)
    }
  })

  # Reset to first page on any filter change
  observeEvent(list(sel_fhir_version(), sel_vendor()), {
    smartres_page_state(1)
  })

  output$smartres_prev_button_ui <- renderUI({
    if (smartres_page_state() > 1) {
      actionButton(ns("smartres_prev_page"), "Previous", icon = icon("arrow-left"))
    } else {
      NULL  # Hide the button
    }
  })

  output$smartres_next_button_ui <- renderUI({
    # Always show next button - let the database handle empty results
    actionButton(ns("smartres_next_page"), "Next", icon = icon("arrow-right"))
  })

  # IN PROGRESS - need to get the correct query with smart_http_response and smart_response columns
  # can we show a summary table of how many endpoints supporting /.well-known/smart-configuration ?

  get_filtered_data <- function(table_val) {
  res <- table_val
  req(sel_fhir_version(), sel_vendor())
  res <- res %>% filter(fhir_version %in% sel_fhir_version())
  if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
    res <- res %>% filter(vendor_name == sel_vendor())
  }
  res
}

  selected_smart_capabilities <- reactive({
  # Get current filter values
  current_fhir <- sel_fhir_version()
  current_vendor <- sel_vendor()
  
  req(current_fhir, current_vendor)
  
  # Retrieve aggregated capabilities directly from the materialized view via SQL
  res <- get_smart_response_capability_count(
    db_connection,
    fhir_version = current_fhir,
    vendor = current_vendor
  )
  
  res <- res %>% mutate(Endpoints = as.integer(Endpoints))
  res
  })

  selected_smart_vendors <- reactive({
  # Get current filter values
  current_fhir <- sel_fhir_version()
  current_vendor <- sel_vendor()
  
  req(current_fhir, current_vendor)
  
  # Retrieve the aggregated vendor table from SQL
  res <- get_smart_vendor_table(
    db_connection,
    fhir_version = current_fhir,
    vendor = current_vendor
  )
  res <- res %>% mutate(Endpoints = as.integer(Endpoints))
  res
  })

  selected_smart_count_total <- reactive({
    # Get current filter values
    current_fhir <- sel_fhir_version()
    current_vendor <- sel_vendor()
    
    req(current_fhir, current_vendor)
    
    # Get the count directly from the materialized view with SQL filtering
    count <- get_selected_smart_count_total(
      db_connection,
      fhir_version = current_fhir,
      vendor = current_vendor
    )
    
    count
  })

  selected_smart_count_200 <- reactive({
  # Get current filter values
  current_fhir <- sel_fhir_version()
  current_vendor <- sel_vendor()
  
  req(current_fhir, current_vendor)
  
  # Retrieve the count directly from the materialized view with all SQL filtering
  count <- get_selected_smart_count_200(
    db_connection,
    fhir_version = current_fhir,
    vendor = current_vendor
  )
  
  count
  })

  selected_well_known_endpoints_count <- reactive({
  # Get current filter values
  current_fhir <- sel_fhir_version()
  current_vendor <- sel_vendor()
  
  req(current_fhir, current_vendor)
  
  # Retrieve the count directly from the materialized view with all SQL filtering applied
  count <- get_selected_well_known_endpoints_count(
    db_connection,
    fhir_version = current_fhir,
    vendor = current_vendor
  )
  
  count
  })

  selected_well_known_count_doc <- reactive({
  # Get current filter values
  current_fhir <- sel_fhir_version()
  current_vendor <- sel_vendor()
  
  req(current_fhir, current_vendor)
  
  # Retrieve the count directly from the materialized view with SQL filtering
  count <- get_selected_well_known_count_doc(
    db_connection,
    fhir_version = current_fhir,
    vendor = current_vendor
  )
  
  count
  })

  selected_well_known_count_no_doc <- reactive({
  # Get current filter values
  current_fhir <- sel_fhir_version()
  current_vendor <- sel_vendor()
  
  req(current_fhir, current_vendor)
  
  # Retrieve the count directly from the materialized view with SQL filtering
  count <- get_selected_well_known_count_no_doc(
    db_connection,
    fhir_version = current_fhir,
    vendor = current_vendor
  )
  
  count
  })

  selected_well_known_endpoint_counts <- reactive({
    res <- tribble(
      ~Status, ~Endpoints,
      "Total Indexed Endpoints", as.integer(selected_smart_count_total()),
      "Endpoints with successful response (HTTP 200)", as.integer(selected_smart_count_200()),
      "Well Known URI Endpoints with successful response (HTTP 200)", as.integer(selected_well_known_endpoints_count()),
      "Well Known URI Endpoints with valid response JSON document", as.integer(selected_well_known_count_doc()),
      "Well Known URI Endpoints without valid response JSON document", as.integer(selected_well_known_count_no_doc())
    )
  })

  output$smart_capability_count_table <- renderTable(
    selected_smart_capabilities()
  )

  output$smart_vendor_table <- renderTable(
    selected_smart_vendors()
  )

  output$well_known_summary_table <- renderTable(
    selected_well_known_endpoint_counts()
  )

  # Modified selected_endpoints with LIMIT OFFSET pagination
  selected_endpoints <- reactive({
    current_fhir <- sel_fhir_version()
    current_vendor <- sel_vendor()
    req(current_fhir, current_vendor)
    
    smartres_offset <- (smartres_page_state() - 1) * smartres_page_size
    
    res <- get_selected_endpoints(
      db_connection,
      fhir_version = current_fhir,
      vendor = current_vendor,
      limit = smartres_page_size,
      offset = smartres_offset
    )

    # Format the URL for HTML display with a modal popup.
    res <- res %>%
      mutate(url = paste0("<a class=\"lantern-url\" tabindex=\"0\" aria-label=\"Press enter to open a pop up modal containing additional information for this endpoint.\" onkeydown = \"javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)\" onclick=\"Shiny.setInputValue(\'endpoint_popup\',&quot;", url, "&&", "None", "&quot,{priority: \'event\'});\">", url, "</a>"))
    
    res
  })

  output$well_known_endpoints <-  reactable::renderReactable({
    reactable(selected_endpoints(),
                columns = list(
                  url = colDef(name = "URL", html = TRUE),
                  condensed_organization_names = colDef(name = "Organization", html = TRUE),
                  vendor_name = colDef(name = "Developer"),
                  capability_fhir_version = colDef(name = "FHIR Version")
                ),
                sortable = TRUE,
                searchable = FALSE,  # Disabled search for performance
                showSortIcon = TRUE,
                defaultPageSize = 10,
                showPageSizeOptions = FALSE,  # Disabled page size options
                pageSizeOptions = NULL  # Removed page size options
    )
  })
}