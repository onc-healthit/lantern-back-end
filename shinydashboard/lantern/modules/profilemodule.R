library(DT)
library(purrr)
library(reactable)
library(glue)

profilemodule_UI <- function(id) {
  ns <- NS(id)
  tagList(
    fluidRow(
      column(width = 6, textInput(ns("profile_search_query"), "Search:", value = ""))
    ),
    reactable::reactableOutput(ns("profiles_table")),
    fluidRow(
      column(3, 
        div(style = "display: flex; justify-content: flex-start;", 
            uiOutput(ns("profile_prev_button_ui"))
        )
      ),
      column(6,
        div(style = "display: flex; justify-content: center; align-items: center; gap: 10px; margin-top: 8px;",
            numericInput(ns("profile_page_selector"), label = NULL, value = 1, min = 1, max = 1, step = 1, width = "80px"),
            textOutput(ns("profile_page_info"), inline = TRUE)
        )
      ),
      column(3, 
        div(style = "display: flex; justify-content: flex-end;",
            uiOutput(ns("profile_next_button_ui"))
        )
      )
    )
  )
}

profilemodule <- function(
  input,
  output,
  session,
  sel_fhir_version,
  sel_vendor,
  sel_resource,
  sel_profile
) {
  ns <- session$ns

  profile_page_state <- reactiveVal(1)
  profile_page_size <- 10

  # FAST COUNT: Get total count without loading all data
  profile_total_count <- reactive({
    req(sel_fhir_version(), sel_vendor())
    
    # Count query - much faster than loading all data
    count_query <- "SELECT COUNT(*) as total FROM mv_profiles_paginated WHERE 1=1"
    params <- list()
    
    # Add same filters as main query
    count_query <- paste0(count_query, " AND fhir_version IN ({fhir_versions*})")
    params$fhir_versions <- sel_fhir_version()
    
    if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
      count_query <- paste0(count_query, " AND vendor_name = {vendor}")
      params$vendor <- sel_vendor()
    }
    
    if (length(sel_resource()) > 0 && sel_resource() != ui_special_values$ALL_RESOURCES) {
      count_query <- paste0(count_query, " AND resource = {resource}")
      params$resource <- sel_resource()
    }
    
    if (length(sel_profile()) > 0 && sel_profile() != ui_special_values$ALL_PROFILES) {
      count_query <- paste0(count_query, " AND profileurl = {profile}")
      params$profile <- sel_profile()
    }

    if (trimws(input$profile_search_query) != "") {
      keyword <- tolower(trimws(input$profile_search_query))
      count_query <- paste0(count_query, " AND (LOWER(url) LIKE {search} OR LOWER(profileurl) LIKE {search} OR LOWER(profilename) LIKE {search}")
      count_query <- paste0(count_query, " OR LOWER(resource) LIKE {search} OR LOWER(vendor_name) LIKE {search})")
      params$search <- paste0("%", keyword, "%")
    }
    
    query <- do.call(glue_sql, c(list(count_query, .con = db_connection), params))
    result <- tbl(db_connection, sql(query)) %>% collect()
    as.numeric(result$total[1])
  })

  # Calculate total pages from count
  profile_total_pages <- reactive({
    total_count <- profile_total_count()
    if (total_count == 0) {
      return(1)
    }
    max(1, ceiling(total_count / profile_page_size))
  })

  # Update page selector max when total pages change
  observe({
    updateNumericInput(session, "profile_page_selector", 
                      max = profile_total_pages(),
                      value = profile_page_state())
  })

  # Handle page selector input
  observeEvent(input$profile_page_selector, {
    if (!is.null(input$profile_page_selector) && !is.na(input$profile_page_selector)) {
      new_page <- max(1, min(input$profile_page_selector, profile_total_pages()))
      profile_page_state(new_page)
      
      if (new_page != input$profile_page_selector) {
        updateNumericInput(session, "profile_page_selector", value = new_page)
      }
    }
  })

  # Handle next page button
  observeEvent(input$profile_next_page, {
    current_time <- as.numeric(Sys.time()) * 1000
    if (!is.null(session$userData$last_next_time) && 
        (current_time - session$userData$last_next_time) < 300) {
      return()  # Ignore only rapid consecutive clicks
    }
    session$userData$last_next_time <- current_time
    if (profile_page_state() < profile_total_pages()) {
      new_page <- profile_page_state() + 1
      profile_page_state(new_page)
      updateNumericInput(session, "profile_page_selector", value = new_page)
    }
  })

  # Handle previous page button
  observeEvent(input$profile_prev_page, {
    current_time <- as.numeric(Sys.time()) * 1000
    if (!is.null(session$userData$last_prev_time) && 
        (current_time - session$userData$last_prev_time) < 300) {
      return()  # Ignore only rapid consecutive clicks
    }
    session$userData$last_prev_time <- current_time
    if (profile_page_state() > 1) {
      new_page <- profile_page_state() - 1
      profile_page_state(new_page)
      updateNumericInput(session, "profile_page_selector", value = new_page)
    }
  })

  # Reset to first page on any filter/search change 
  observeEvent(list(sel_fhir_version(), sel_vendor(), sel_resource(), sel_profile(), input$profile_search_query), {
    profile_page_state(1)
    updateNumericInput(session, "profile_page_selector", value = 1)
  })

  # Boundary condition handling using count 
  output$profile_prev_button_ui <- renderUI({
    if (profile_page_state() > 1) {
      actionButton(ns("profile_prev_page"), "Previous", icon = icon("arrow-left"))
    } else {
      NULL
    }
  })

  output$profile_next_button_ui <- renderUI({
    if (profile_page_state() < profile_total_pages()) {
      actionButton(ns("profile_next_page"), "Next", icon = icon("arrow-right"))
    } else {
      NULL
    }
  })

  output$profile_page_info <- renderText({
    paste("of", profile_total_pages())
  })

  # FAST PAGINATION: Only load the 10 rows needed for current page
  selected_fhir_endpoint_profiles <- reactive({
    req(sel_fhir_version(), sel_vendor())
    
    profile_offset <- (profile_page_state() - 1) * profile_page_size
    
    # Query only the data needed for current page
    base_query <- "SELECT url, profileurl, profilename, resource, fhir_version, vendor_name 
                   FROM mv_profiles_paginated 
                   WHERE 1=1"
    
    params <- list()
    
    # Add FHIR version filter
    base_query <- paste0(base_query, " AND fhir_version IN ({fhir_versions*})")
    params$fhir_versions <- sel_fhir_version()
    
    # Add vendor filter
    if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
      base_query <- paste0(base_query, " AND vendor_name = {vendor}")
      params$vendor <- sel_vendor()
    }
    
    # Add resource filter
    if (length(sel_resource()) > 0 && sel_resource() != ui_special_values$ALL_RESOURCES) {
      base_query <- paste0(base_query, " AND resource = {resource}")
      params$resource <- sel_resource()
    }
    
    # Add profile filter
    if (length(sel_profile()) > 0 && sel_profile() != ui_special_values$ALL_PROFILES) {
      base_query <- paste0(base_query, " AND profileurl = {profile}")
      params$profile <- sel_profile()
    }

    # Add search functionality
    if (trimws(input$profile_search_query) != "") {
      keyword <- tolower(trimws(input$profile_search_query))
      base_query <- paste0(base_query, " AND (LOWER(url) LIKE {search} OR LOWER(profileurl) LIKE {search} OR LOWER(profilename) LIKE {search}")
      base_query <- paste0(base_query, " OR LOWER(resource) LIKE {search} OR LOWER(vendor_name) LIKE {search})")
      params$search <- paste0("%", keyword, "%")
    }
    
    # Add pagination - only get 10 rows
    base_query <- paste0(base_query, " ORDER BY page_id LIMIT {limit} OFFSET {offset}")
    params$limit <- profile_page_size
    params$offset <- profile_offset
    
    # Execute the query
    query <- do.call(glue_sql, c(list(base_query, .con = db_connection), params))
    res <- tbl(db_connection, sql(query)) %>% collect()
    
    # Add clickable URLs only to the rows we're displaying
    if (nrow(res) > 0) {
      res <- res %>%
        mutate(url = paste0("<a class=\"lantern-url\" tabindex=\"0\" aria-label=\"Press enter to open pop up modal containing additional information for this endpoint.\" onkeydown = \"javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)\" onclick=\"Shiny.setInputValue(\'endpoint_popup\',&quot;", url, "&&", "None", "&quot,{priority: \'event\'});\">", url, "</a>"))
    }
    
    return(res)
  })

  output$profiles_table <- reactable::renderReactable({
    df <- selected_fhir_endpoint_profiles()

    if (nrow(df) == 0) {
      return(reactable(
        data.frame(Message = "No data matching the selected filters"),
        pagination = FALSE,
        searchable = FALSE
      ))
    }

    reactable(
      df,
      defaultColDef = colDef(align = "center"),
      columns = list(
        url = colDef(name = "Endpoint", minWidth = 300, align = "left", html = TRUE, sortable = TRUE),
        profileurl = colDef(name = "Profile URL", minWidth = 250, sortable = TRUE),
        profilename = colDef(name = "Profile Name", minWidth = 200, sortable = TRUE),
        resource = colDef(name = "Resource", sortable = TRUE),
        fhir_version = colDef(name = "FHIR Version", sortable = TRUE),
        vendor_name = colDef(name = "Certified API Developer Name", minWidth = 110, sortable = TRUE)
      ),
      searchable = FALSE,
      showSortIcon = TRUE,
      highlight = TRUE,
      defaultPageSize = profile_page_size
    )
  })
}