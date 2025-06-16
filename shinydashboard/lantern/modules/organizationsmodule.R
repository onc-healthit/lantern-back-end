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
      column(width = 6, textInput(ns("org_search_query"), "Search:", value = ""))
    ),
    fluidRow(
      p("This table shows the organization name listed for each endpoint in the endpoint list it appears in."),
      reactable::reactableOutput(ns("endpoint_list_orgs_table")),
      htmlOutput(ns("note_text"))
    ),
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

  page_state <- reactiveVal(1)
  page_size <- 10

  # Calculate total pages based on filtered data
  total_pages <- reactive({
    total_records <- nrow(selected_endpoint_list_orgs() %>% distinct(organization_name, organization_id, url, fhir_version, vendor_name))
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
    message("NEXT PAGE BUTTON CLICKED")
    if (page_state() < total_pages()) {
      new_page <- page_state() + 1
      page_state(new_page)
      updateNumericInput(session, "page_selector", value = new_page)
    }
  })

  # Handle previous page button
  observeEvent(input$prev_page, {
    message("PREV PAGE BUTTON CLICKED")
    if (page_state() > 1) {
      new_page <- page_state() - 1
      page_state(new_page)
      updateNumericInput(session, "page_selector", value = new_page)
    }
  })

  # Reset to first page on any filter/search change 
  observeEvent(list(sel_fhir_version(), sel_vendor(), sel_confidence(), input$org_search_query), {
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

 selected_endpoint_list_orgs <- reactive({
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

      # Format URL for HTML display with modal popup
      res <- res %>%
        mutate(url = paste0("<a class=\"lantern-url\" tabindex=\"0\" aria-label=\"Press enter to open a pop up modal containing additional information for this endpoint.\" onkeydown = \"javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)\" onclick=\"Shiny.setInputValue(\'endpoint_popup\',&quot;", url, "&quot,{priority: \'event\'});\">", url, "</a>"))
      
      res <- res %>%
        mutate(organization_id = paste0("<a class=\"lantern-url\" tabindex=\"0\" aria-label=\"Press enter to open a pop up modal containing additional information for this organization.\" onkeydown = \"javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)\" onclick=\"Shiny.setInputValue(\'show_organization_modal\',&quot;", organization_id, "&quot,{priority: \'event\'});\"> HTI-1 Data </a>"))
      
      res
    })


  output$endpoint_list_orgs_table <- reactable::renderReactable({
     # Get all data
     display_data <- selected_endpoint_list_orgs()

     if (nrow(display_data) == 0) {
       return(
         reactable(
           data.frame(Message = "No data matching the selected filters"),
           pagination = FALSE,
           searchable = FALSE
         )
       )
     }

     display_data <- display_data %>% 
      select(organization_name, organization_id, url, fhir_version, vendor_name) %>% 
      distinct(organization_name, organization_id, url, fhir_version, vendor_name)
      
     if (trimws(input$org_search_query) != ""){
      display_data <- display_data %>%
       filter(if_any(everything(), ~ str_detect(tolower(as.character(.x)), tolower(trimws(input$org_search_query)))))
     }

    display_data <- display_data %>% arrange(organization_name) %>% group_by(organization_name)

    # Paginate using R slicing
    paged_data <- reactive({
      all_data <- display_data
      start <- (page_state() - 1) * page_size + 1
      end <- min(nrow(all_data), page_state() * page_size)
      if (nrow(all_data) == 0 || start > nrow(all_data)) return(all_data[0, ])
      all_data[start:end, ]
    })

     reactable(
       paged_data(),
       defaultColDef = colDef(
         align = "center"
       ),
       columns = list(
         organization_name = colDef(name = "Organization Name", sortable = TRUE, align = "left",
                                    grouped = JS("function(cellInfo) {return cellInfo.value}")),
         organization_id = colDef(name = "Organization Details", sortable = FALSE, html = TRUE),
         url = colDef(name = "URL", minWidth = 300, sortable = FALSE, html = TRUE),
         fhir_version = colDef(name = "FHIR Version", sortable = FALSE),
         vendor_name = colDef(name = "Certified API Developer Name", minWidth = 110, sortable = FALSE)
       ),
       groupBy = c("organization_name"),
       striped = TRUE,
       searchable = FALSE,
       showSortIcon = TRUE,
       highlight = TRUE,
       pagination = FALSE,
       defaultExpanded = TRUE
     )
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