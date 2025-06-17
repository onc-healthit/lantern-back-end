library(DT)
library(purrr)
library(reactable)
library(htmlwidgets)

# Get contact information function directly included in the module file
get_contact_information <- function(db_connection) {
  # Simply get all data from the materialized view
  tbl(db_connection, "mv_contacts_info") %>% collect()
}

contactsmodule_UI <- function(id) {
  ns <- NS(id)

  tagList(
    fluidRow(
      column(width = 6, textInput(ns("search_query"), "Search:", value = "")
      )
    ),
    reactable::reactableOutput(ns("contacts_table")),
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

    page_state <- reactiveVal(1)
    page_size <- 10

    # Calculate total pages based on filtered data
    total_pages <- reactive({
      total_records <- nrow(selected_contacts() %>% distinct(url, fhir_version))
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
    observeEvent(list(sel_fhir_version(), sel_vendor(), sel_has_contact(), input$search_query), {
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

    # Main data query with database-level search
    selected_contacts <- reactive({
        req(sel_fhir_version(), sel_vendor(), sel_has_contact())
        
        query_str <- "SELECT * FROM mv_contacts_info WHERE fhir_version IN ({vals*})"
        params <- list(vals = sel_fhir_version())

        if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
            query_str <- paste0(query_str, " AND vendor_name = {vendor}")
            params$vendor <- sel_vendor()
        }

        # Apply external search filter at database level
        if (trimws(input$search_query) != "") {
          keyword <- tolower(trimws(input$search_query))
          query_str <- paste0(query_str, " AND (LOWER(url) LIKE {search} OR LOWER(endpoint_names) LIKE {search} OR LOWER(vendor_name) LIKE {search}")
          query_str <- paste0(query_str, " OR LOWER(contact_name) LIKE {search} OR LOWER(contact_type) LIKE {search} OR LOWER(contact_value) LIKE {search})")
          params$search <- paste0("%", keyword, "%")
        }

        query <- do.call(glue_sql, c(list(query_str, .con = db_connection), params))
        res <- tbl(db_connection, sql(query)) %>% collect()

        res <- res %>%
            arrange(contact_preference) %>%
            group_by(url) %>%
            mutate(num_contacts = n()) %>%
            distinct(url, .keep_all = TRUE) %>%
            mutate(linkurl = paste0("<a class=\"lantern-url\" tabindex=\"0\" aria-label=\"Press enter to open a pop up modal containing additional information for this endpoint.\" onkeydown = \"javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)\" onclick=\"Shiny.setInputValue(\'endpoint_popup\',&quot;", url, "&&", requested_fhir_version, "&quot,{priority: \'event\'});\">", url, "</a>"))

        res <- res %>%
            rowwise() %>%
                mutate(has_contact = (!is.na(contact_name) || !is.na(contact_type) || !is.na(contact_value))) %>%
                mutate(contact_name = ifelse(is.na(contact_name), ifelse(is.na(contact_value), "-", "N/A"), toString(contact_name))) %>%
                mutate(contact_type = ifelse(is.na(contact_type), "-", toString(contact_type))) %>%
                mutate(contact_value = ifelse(is.na(contact_value), "-", toString(contact_value))) %>%
                mutate(condensed_endpoint_names = ifelse(length(endpoint_names) > 0, ifelse(length(strsplit(endpoint_names, ";")[[1]]) > 5, paste0(paste0(head(strsplit(endpoint_names, ";")[[1]], 5), collapse = ";"), "; ", paste0("<a class=\"lantern-url\" tabindex=\"0\" aria-label=\"Press enter to open a pop up modal containing the endpoint's entire list of API information source names.\" onkeydown = \"javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)\" onclick=\"Shiny.setInputValue(\'show_details\',&quot;", url, "&quot,{priority: \'event\'});\"> Click For More... </a>")), endpoint_names), endpoint_names))

        res <- res %>%
            rowwise() %>%
            mutate(show_all = ifelse(has_contact, paste0("<a class=\"lantern-url\" tabindex=\"0\" aria-label=\"Press enter to show all contact information.\" onkeydown = \"javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)\" onclick=\"Shiny.setInputValue(\'show_contact_modal\',&quot;", url, "&quot,{priority: \'event\'});\"> Show All Contacts </a>"), "-"))

        if (sel_has_contact() != "Any") {
            res <- res %>% filter(ifelse(sel_has_contact() == "True", has_contact == TRUE, has_contact == FALSE))
        }

        res
    })

    # Paginate using R slicing
    paged_contacts <- reactive({
      all_data <- selected_contacts()
      start <- (page_state() - 1) * page_size + 1
      end <- min(nrow(all_data), page_state() * page_size)
      if (nrow(all_data) == 0 || start > nrow(all_data)) return(all_data[0, ])
      all_data[start:end, ]
    })

    output$contacts_table <- reactable::renderReactable({
     reactable(
              paged_contacts() %>%
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