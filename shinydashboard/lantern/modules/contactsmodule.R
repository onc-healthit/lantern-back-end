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
    sel_has_contact,
    is_active
) {
    ns <- session$ns

    contacts_page_size <- 10

    # Add request tracking to prevent race conditions
    current_request_id <- reactiveVal(0)

    # Shared WHERE-clause builder reused by the paginated fetch and the total_pages count query --
    # previously duplicated verbatim between this and a since-removed selected_contacts_without_limit()
    # that existed only to be nrow()'d for the page count.
    contacts_filter_query <- reactive({
        req(sel_fhir_version(), sel_vendor(), sel_has_contact(), is_active())

        query_str <- "
        SELECT DISTINCT ON (url, vendor_name) *
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

        list(query_str = query_str, params = params)
    })

    # Calculate total pages via a real COUNT(*) instead of pulling the full filtered dataset into
    # R just to nrow() it.
    contacts_total_pages <- reactive({
      filt <- contacts_filter_query()
      count_query_str <- paste0("SELECT COUNT(*) as count FROM (", filt$query_str, ") base")
      count_query <- do.call(glue_sql, c(list(count_query_str, .con = db_connection), filt$params))
      total_records <- tbl(db_connection, sql(count_query)) %>% collect() %>% pull(count)
      max(1, ceiling(total_records / contacts_page_size))
    })

    contacts_page_state <- create_pager(
      input, output, session,
      page_selector_id = "contacts_page_selector",
      prev_button_id = "contacts_prev_page",
      next_button_id = "contacts_next_page",
      prev_output_id = "contacts_prev_button_ui",
      next_output_id = "contacts_next_button_ui",
      total_pages = contacts_total_pages
    )

    # Reset to first page on any filter/search change
    observeEvent(list(sel_fhir_version(), sel_vendor(), sel_has_contact(), input$contacts_search_query), {
      contacts_page_state(1)
    })

    output$contacts_page_info <- renderText({
      paste("of", contacts_total_pages())
    })

    # Main data query for pagination and filtering - WITH RACE CONDITION PROTECTION
    selected_contacts <- reactive({
        filt <- contacts_filter_query()

        # Generate unique request ID
        request_id <- isolate(current_request_id()) + 1
        current_request_id(request_id)

        contacts_offset <- (contacts_page_state() - 1) * contacts_page_size

        query_str <- paste0(filt$query_str, "
        ORDER BY url, vendor_name, contact_preference DESC
        LIMIT {limit} OFFSET {offset}")
        params <- c(filt$params, list(limit = contacts_page_size, offset = contacts_offset))

        query <- do.call(glue_sql, c(list(query_str, .con = db_connection), params))
        result <- tbl(db_connection, sql(query)) %>% collect()

        # Only return results if this is still the latest request
        # Use isolate() to check without creating reactive dependency
        if (request_id == isolate(current_request_id())) {
          # This is the latest request, process normally
          res <- result %>%
              mutate(linkurl = paste0("<a class=\"lantern-url\" tabindex=\"0\" aria-label=\"Press enter to open pop up modal containing additional information for this endpoint.\" onkeydown = \"javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)\" onclick=\"Shiny.setInputValue(\'endpoint_popup\',&quot;", url, "&&", requested_fhir_version, "&&", vendor_name, "&quot,{priority: \'event\'});\">", url, "</a>")) %>%
              rowwise() %>%
              mutate(contact_name = ifelse(is.na(contact_name), ifelse(is.na(contact_value), "-", "N/A"), toString(contact_name))) %>%
              mutate(contact_type = ifelse(is.na(contact_type), "-", toString(contact_type))) %>%
              mutate(contact_value = ifelse(is.na(contact_value), "-", toString(contact_value))) %>%
              mutate(condensed_endpoint_names = ifelse(length(endpoint_names) > 0, ifelse(length(strsplit(endpoint_names, ";")[[1]]) > 5, paste0(paste0(head(strsplit(endpoint_names, ";")[[1]], 5), collapse = ";"), "; ", paste0("<a class=\"lantern-url\" tabindex=\"0\" aria-label=\"Press enter to open a pop up modal containing the endpoint's entire list of API information source names.\" onkeydown = \"javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)\" onclick=\"Shiny.setInputValue(\'show_details\',&quot;", url, "&&", vendor_name, "&quot,{priority: \'event\'});\"> Click For More... </a>")), endpoint_names), endpoint_names)) %>%
              mutate(show_all = ifelse(has_contact, paste0("<a class=\"lantern-url\" tabindex=\"0\" aria-label=\"Press enter to show all contact information.\" onkeydown = \"javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)\" onclick=\"Shiny.setInputValue(\'show_contact_modal\',&quot;", url, "&quot,{priority: \'event\'});\"> Show All Contacts </a>"), "-"))

          return(res)
        } else {
          # This request was superseded, return empty to avoid flicker
          return(data.frame())
        }
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