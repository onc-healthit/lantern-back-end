library(DT)
library(purrr)
library(reactable)
library(htmlwidgets)
library(glue)

# Get contact information function directly included in the module file
get_contact_information <- function(db_connection) {
  # Simply get all data from the materialized view
  tbl(db_connection, "mv_contacts_info") %>% collect()
}

contactsmodule_UI <- function(id) {
  ns <- NS(id)

  tagList(
    fluidRow(
      column(width = 6, textInput(ns("contacts_search_query"), "Search:", value = "")
      )
    ),
    reactable::reactableOutput(ns("contacts_table")),
    fluidRow(
      column(3, 
        div(style = "display: flex; justify-content: flex-start;", 
            uiOutput(ns("contacts_prev_button_ui"))
        )
      ),
      column(6,
        div(style = "display: flex; justify-content: center; align-items: center; gap: 10px; margin-top: 8px;",
            numericInput(ns("contacts_page_selector"), label = NULL, value = 1, min = 1, max = 1, step = 1, width = "80px"),
            textOutput(ns("contacts_page_info"), inline = TRUE)
        )
      ),
      column(3, 
        div(style = "display: flex; justify-content: flex-end;",
            uiOutput(ns("contacts_next_button_ui"))
        )
      )
    )
  )
}

contactsmodule <- function(
    input,
    output,
    session,
    sel_fhir_version,
    sel_vendor,
    sel_has_contact
) {
    ns <- session$ns

    contacts_page_state <- reactiveVal(1)
    contacts_page_size <- 10

    # Add request tracking to prevent race conditions
    current_request_id <- reactiveVal(0)

    # Calculate total pages based on filtered data
    contacts_total_pages <- reactive({
      total_records <- nrow(selected_contacts_without_limit() %>% distinct(url, fhir_version))
      max(1, ceiling(total_records / contacts_page_size))
    })

    # Break the feedback loop with isolate()
    observe({
      new_page <- contacts_page_state()
      current_selector <- input$contacts_page_selector
      
      # Only update if different (prevents infinite loop)
      # Add safety check for current_selector to prevent crashes
      if (is.null(current_selector) || 
          is.na(current_selector) || 
          !is.numeric(current_selector) ||
          current_selector != new_page) {
        
        isolate({  # This is the key fix to break feedback loops
          updateNumericInput(session, "contacts_page_selector", 
                            max = contacts_total_pages(),
                            value = new_page)
        })
      }
    })

    # Handle page selector input
    observeEvent(input$contacts_page_selector, {
      # Get current input value
      current_input <- input$contacts_page_selector
      
      # Check if input is valid (not NULL, not NA, and is a number)
      if (!is.null(current_input) && 
          !is.na(current_input) && 
          is.numeric(current_input) &&
          current_input > 0) {
        
        new_page <- max(1, min(current_input, contacts_total_pages()))
        
        # Only update page state if it's actually different
        if (new_page != contacts_page_state()) {
          contacts_page_state(new_page)
        }

        # Correct the input field if the user entered an invalid page number
        if (new_page != current_input) {
          updateNumericInput(session, "contacts_page_selector", value = new_page)
        }
      } else {
        # If input is invalid (empty, NA, or <= 0), reset to current page
        # Use a small delay to prevent immediate feedback loop
        invalidateLater(100)
        updateNumericInput(session, "contacts_page_selector", value = contacts_page_state())
      }
    }, ignoreInit = TRUE)  # Prevent firing on initialization

    # Handle next page button
    observeEvent(input$contacts_next_page, {
      if (contacts_page_state() < contacts_total_pages()) {
        new_page <- contacts_page_state() + 1
        contacts_page_state(new_page)
      }
    })

    # Handle previous page button
    observeEvent(input$contacts_prev_page, {
      if (contacts_page_state() > 1) {
        new_page <- contacts_page_state() - 1
        contacts_page_state(new_page)
      }
    })

    # Reset to first page on any filter/search change
    observeEvent(list(sel_fhir_version(), sel_vendor(), sel_has_contact(), input$contacts_search_query), {
      contacts_page_state(1)
    })

    output$contacts_prev_button_ui <- renderUI({
      if (contacts_page_state() > 1) {
        actionButton(ns("contacts_prev_page"), "Previous", icon = icon("arrow-left"))
      } else {
        NULL  # Hide the button
      }
    })

    output$contacts_next_button_ui <- renderUI({
      if (contacts_page_state() < contacts_total_pages()) {
        actionButton(ns("contacts_next_page"), "Next", icon = icon("arrow-right"))
      } else {
        NULL  # Hide the button
      }
    })

    output$contacts_page_info <- renderText({
      paste("of", contacts_total_pages())
    })

    # Main data query for pagination and filtering - WITH RACE CONDITION PROTECTION
    selected_contacts <- reactive({
        req(sel_fhir_version(), sel_vendor(), sel_has_contact())
        
        # Generate unique request ID 
        request_id <- isolate(current_request_id()) + 1
        current_request_id(request_id)
        
        contacts_offset <- (contacts_page_state() - 1) * contacts_page_size
        
        # ESSENTIAL CHANGE: Get unique URLs first, then paginate
        query_str <- "
        SELECT DISTINCT ON (url) *
        FROM mv_contacts_info 
        WHERE fhir_version IN ({vals*})"
        
        params <- list(vals = sel_fhir_version())

        if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
            query_str <- paste0(query_str, " AND vendor_name = {vendor}")
            params$vendor <- sel_vendor()
        }

        # Apply has_contact filter at DATABASE level
        if (sel_has_contact() != "Any") {
            if (sel_has_contact() == "True") {
                query_str <- paste0(query_str, " AND has_contact = TRUE")
            } else {
                query_str <- paste0(query_str, " AND has_contact = FALSE")
            }
        }

        # Apply external search filter at database level
        if (trimws(input$contacts_search_query) != "") {
          keyword <- tolower(trimws(input$contacts_search_query))
          query_str <- paste0(query_str, " AND (LOWER(url) LIKE {search} OR LOWER(endpoint_names) LIKE {search} OR LOWER(vendor_name) LIKE {search}")
          query_str <- paste0(query_str, " OR LOWER(contact_name) LIKE {search} OR LOWER(contact_type) LIKE {search} OR LOWER(contact_value) LIKE {search})")
          params$search <- paste0("%", keyword, "%")
        }

        # Add ordering and pagination
        query_str <- paste0(query_str, " 
        ORDER BY url, contact_preference
        LIMIT {limit} OFFSET {offset}")
        
        params$limit <- contacts_page_size
        params$offset <- contacts_offset

        query <- do.call(glue_sql, c(list(query_str, .con = db_connection), params))
        result <- tbl(db_connection, sql(query)) %>% collect()

        # Only return results if this is still the latest request
        # Use isolate() to check without creating reactive dependency
        if (request_id == isolate(current_request_id())) {
          # This is the latest request, process normally
          res <- result %>%
              mutate(linkurl = paste0("<a class=\"lantern-url\" tabindex=\"0\" aria-label=\"Press enter to open pop up modal containing additional information for this endpoint.\" onkeydown = \"javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)\" onclick=\"Shiny.setInputValue(\'endpoint_popup\',&quot;", url, "&&", requested_fhir_version, "&quot,{priority: \'event\'});\">", url, "</a>")) %>%
              rowwise() %>%
              mutate(contact_name = ifelse(is.na(contact_name), ifelse(is.na(contact_value), "-", "N/A"), toString(contact_name))) %>%
              mutate(contact_type = ifelse(is.na(contact_type), "-", toString(contact_type))) %>%
              mutate(contact_value = ifelse(is.na(contact_value), "-", toString(contact_value))) %>%
              mutate(condensed_endpoint_names = ifelse(length(endpoint_names) > 0, ifelse(length(strsplit(endpoint_names, ";")[[1]]) > 5, paste0(paste0(head(strsplit(endpoint_names, ";")[[1]], 5), collapse = ";"), "; ", paste0("<a class=\"lantern-url\" tabindex=\"0\" aria-label=\"Press enter to open a pop up modal containing the endpoint's entire list of API information source names.\" onkeydown = \"javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)\" onclick=\"Shiny.setInputValue(\'show_details\',&quot;", url, "&quot,{priority: \'event\'});\"> Click For More... </a>")), endpoint_names), endpoint_names)) %>%
              mutate(show_all = ifelse(has_contact, paste0("<a class=\"lantern-url\" tabindex=\"0\" aria-label=\"Press enter to show all contact information.\" onkeydown = \"javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)\" onclick=\"Shiny.setInputValue(\'show_contact_modal\',&quot;", url, "&quot,{priority: \'event\'});\"> Show All Contacts </a>"), "-"))

          return(res)
        } else {
          # This request was superseded, return empty to avoid flicker
          return(data.frame())
        }
    })

    # Query without limit for total count calculation
    selected_contacts_without_limit <- reactive({
        req(sel_fhir_version(), sel_vendor(), sel_has_contact())
        
        # Same query as main but without LIMIT OFFSET
        query_str <- "
        SELECT DISTINCT ON (url) *
        FROM mv_contacts_info 
        WHERE fhir_version IN ({vals*})"
        
        params <- list(vals = sel_fhir_version())

        if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
            query_str <- paste0(query_str, " AND vendor_name = {vendor}")
            params$vendor <- sel_vendor()
        }

        # Apply has_contact filter at DATABASE level
        if (sel_has_contact() != "Any") {
            if (sel_has_contact() == "True") {
                query_str <- paste0(query_str, " AND has_contact = TRUE")
            } else {
                query_str <- paste0(query_str, " AND has_contact = FALSE")
            }
        }

        # Apply external search filter at database level
        if (trimws(input$contacts_search_query) != "") {
          keyword <- tolower(trimws(input$contacts_search_query))
          query_str <- paste0(query_str, " AND (LOWER(url) LIKE {search} OR LOWER(endpoint_names) LIKE {search} OR LOWER(vendor_name) LIKE {search}")
          query_str <- paste0(query_str, " OR LOWER(contact_name) LIKE {search} OR LOWER(contact_type) LIKE {search} OR LOWER(contact_value) LIKE {search})")
          params$search <- paste0("%", keyword, "%")
        }

        # Add ordering but no pagination
        query_str <- paste0(query_str, " ORDER BY url, contact_preference")

        query <- do.call(glue_sql, c(list(query_str, .con = db_connection), params))
        res <- tbl(db_connection, sql(query)) %>% collect()

        res
    })

    output$contacts_table <- reactable::renderReactable({
     reactable(
              selected_contacts() %>%
              select(linkurl, fhir_version, condensed_endpoint_names, vendor_name, has_contact, contact_name, contact_type, contact_value, contact_preference, show_all) %>%
              arrange(linkurl),
              defaultColDef = colDef(
                align = "center"
              ),
              columns = list(
                  linkurl = colDef(name = "URL", minWidth = 300, html = TRUE, sortable = TRUE),
                  fhir_version = colDef(name = "FHIR Version", sortable = TRUE, aggregate = "unique"),
                  condensed_endpoint_names = colDef(name = "API Information Source Name", aggregate = "unique", minWidth = 200, sortable = TRUE, html = TRUE),
                  vendor_name = colDef(name = "Certified API Developer Name", aggregate = "unique", minWidth = 110, sortable = TRUE),
                  has_contact = colDef(name = "Has Contact Information", aggregate = "unique", sortable = TRUE),
                  contact_name = colDef(name = "Preferred Contact Name", sortable = TRUE),
                  contact_type = colDef(name = "Preferred Contact Type", sortable = TRUE),
                  contact_value = colDef(name = "Preferred Contact Info", sortable = TRUE),
                  contact_preference = colDef(show = FALSE, sortable = TRUE),
                  show_all = colDef(name = "All Contacts", html = TRUE, sortable = TRUE)
              ),
              striped = TRUE,
              searchable = FALSE,
              showSortIcon = TRUE,
              highlight = TRUE,
              defaultPageSize = 10
     )
    })
}