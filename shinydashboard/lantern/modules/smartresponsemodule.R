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
