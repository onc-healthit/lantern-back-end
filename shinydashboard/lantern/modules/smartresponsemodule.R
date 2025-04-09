# SMART-on-FHIR Well-known URI responses

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
    reactable::reactableOutput(ns("well_known_endpoints"))
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

get_selected_endpoints <- function(db_connection, fhir_version, vendor) {
  query <- tbl(db_connection, "mv_selected_endpoints")
  
  # Apply filtering on fhir_version
  if (!is.null(fhir_version) && length(fhir_version) > 0) {
    query <- query %>% filter(capability_fhir_version %in% !!fhir_version)
  }
  
  # Apply filtering on vendor
  if (!is.null(vendor) && vendor != ui_special_values$ALL_DEVELOPERS) {
    query <- query %>% filter(vendor_name == !!vendor)
  }

  # Remove unique_id column
  query <- query %>% select(-mv_id) 
  
  # Collect the filtered data
  result <- query %>% collect()
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

  smartPageSizeNum <- reactiveVal(NULL)

  # url requested version is default set to None since this table filters on requested_version = 'None'
  selected_endpoints <- reactive({
  if (is.null(isolate(smartPageSizeNum()))) {
    smartPageSizeNum(10)
  }
  
  current_fhir <- sel_fhir_version()
  current_vendor <- sel_vendor()
  req(current_fhir, current_vendor)
  
  res <- get_selected_endpoints(
    db_connection,
    fhir_version = current_fhir,
    vendor = current_vendor
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
                searchable = TRUE,
                showSortIcon = TRUE,
                defaultPageSize = isolate(smartPageSizeNum())
    )
  })

  observeEvent(input$well_known_endpoints_state$length, {
    if (is.null(isolate(smartPageSizeNum()))) {
      smartPageSizeNum(10)
    }
    page <- input$well_known_endpoints_state$length
    smartPageSizeNum(page)
  })

}
