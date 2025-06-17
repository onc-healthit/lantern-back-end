library(DT)
library(purrr)
library(reactable)

profilemodule_UI <- function(id) {
  ns <- NS(id)
  tagList(
    fluidRow(
      column(width = 6, textInput(ns("search_query"), "Search:", value = "")
      )
    ),
    reactable::reactableOutput(ns("profiles_table")),
    fluidRow(
      column(3, 
        div(style = "display: flex; justify-content: flex-start;", uiOutput(ns("prev_button_ui"))
        )
      ),
      column(6,
        div(style = "display: flex; justify-content: center; align-items: center; gap: 10px; margin-top: 8px;",
            numericInput(ns("page_selector"), label = NULL, value = 1, min = 1, max = 1, step = 1, width = "80px"),
            textOutput(ns("page_info"), inline = TRUE)
        )
      ),
      column(3, 
        div(style = "display: flex; justify-content: flex-end;", uiOutput(ns("next_button_ui"))
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

  page_state <- reactiveVal(1)
  page_size <- 10

  # Calculate total pages based on filtered data
  total_pages <- reactive({
    total_records <- nrow(selected_fhir_endpoint_profiles() %>% distinct(url, profileurl, fhir_version))
    max(1, ceiling(total_records / page_size))
  })

  # Update page selector max when total pages change
  observe({
    updateNumericInput(session, ns("page_selector"), 
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
        updateNumericInput(session, ns("page_selector"), value = new_page)
      }
    }
  })

  # Handle next page button
  observeEvent(input$next_page, {
    if (page_state() < total_pages()) {
      new_page <- page_state() + 1
      page_state(new_page)
      updateNumericInput(session, ns("page_selector"), value = new_page)
    }
  })

  # Handle previous page button
  observeEvent(input$prev_page, {
    if (page_state() > 1) {
      new_page <- page_state() - 1
      page_state(new_page)
      updateNumericInput(session, ns("page_selector"), value = new_page)
    }
  })

  # Reset to first page on any filter/search change
  observeEvent(list(sel_fhir_version(), sel_vendor(), sel_resource(), sel_profile(), input$search_query), {
    page_state(1)
    updateNumericInput(session, ns("page_selector"), value = 1)
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

  # Main data query without pagination
  selected_fhir_endpoint_profiles <- reactive({
    req(sel_fhir_version(), sel_vendor())
    
    query_str <- "SELECT DISTINCT url, profileurl, profilename, resource, fhir_version, vendor_name FROM endpoint_supported_profiles_mv WHERE fhir_version IN ({vals*})"
    params <- list(vals = sel_fhir_version())

    if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
      query_str <- paste0(query_str, " AND vendor_name = {vendor}")
      params$vendor <- sel_vendor()
    }

    if (length(sel_resource()) > 0) {
      if (sel_resource() != ui_special_values$ALL_RESOURCES) {
        query_str <- paste0(query_str, " AND resource = {resource}")
        params$resource <- sel_resource()
      }
    }

    if (length(sel_profile()) > 0) {
      if (sel_profile() != ui_special_values$ALL_PROFILES) {
        query_str <- paste0(query_str, " AND profileurl = {profile}")
        params$profile <- sel_profile()
      }
    }

    # Apply external search filter at database level
    if (trimws(input$search_query) != "") {
      keyword <- tolower(trimws(input$search_query))
      query_str <- paste0(query_str, " AND (LOWER(url) LIKE {search} OR LOWER(profileurl) LIKE {search} OR LOWER(profilename) LIKE {search}")
      query_str <- paste0(query_str, " OR LOWER(resource) LIKE {search} OR LOWER(vendor_name) LIKE {search} OR LOWER(fhir_version) LIKE {search})")
      params$search <- paste0("%", keyword, "%")
    }

    query <- do.call(glue_sql, c(list(query_str, .con = db_connection), params))
    res <- tbl(db_connection, sql(query)) %>% collect()
    
    res <- res %>%
      group_by(url) %>%
      mutate(url = paste0("<a class=\"lantern-url\" tabindex=\"0\" aria-label=\"Press enter to open pop up modal containing additional information for this endpoint.\" onkeydown = \"javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)\" onclick=\"Shiny.setInputValue(\'endpoint_popup\',&quot;", url, "&&", "None", "&quot,{priority: \'event\'});\">", url, "</a>")) %>%
      mutate_at(vars(-group_cols()), as.character)

    return(res)
  })

  # Paginate using R slicing
  paged_profiles <- reactive({
    all_data <- selected_fhir_endpoint_profiles()
    start <- (page_state() - 1) * page_size + 1
    end <- min(nrow(all_data), page_state() * page_size)
    if (nrow(all_data) == 0 || start > nrow(all_data)) return(all_data[0, ])
    all_data[start:end, ]
  })

  output$profiles_table <- reactable::renderReactable({
    df <- paged_profiles()

    if (nrow(df) == 0) {
      return(reactable(
        data.frame(Message = "No data matching the selected filters"),
        pagination = FALSE,
        searchable = FALSE
      ))
    }

    reactable(
      df,
      defaultColDef = colDef(
        align = "center"
      ),
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
      defaultPageSize = 10
    )
  })
}