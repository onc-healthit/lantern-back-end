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
    reactable::reactableOutput(ns("contacts_table")),
    fluidRow(
      column(6, 
        div(style = "display: flex; justify-content: flex-start;", 
            uiOutput(ns("contacts_prev_button_ui"))
        )
      ),
      column(6, 
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

    # Handle next page button
    observeEvent(input$contacts_next_page, {
      new_page <- contacts_page_state() + 1
      contacts_page_state(new_page)
    })

    # Handle previous page button
    observeEvent(input$contacts_prev_page, {
      if (contacts_page_state() > 1) {
        new_page <- contacts_page_state() - 1
        contacts_page_state(new_page)
      }
    })

    # Reset to first page on any filter change
    observeEvent(list(sel_fhir_version(), sel_vendor(), sel_has_contact()), {
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
      # Always show next button - let the database handle empty results
      actionButton(ns("contacts_next_page"), "Next", icon = icon("arrow-right"))
    })

    # Main data query with LIMIT OFFSET pagination
    selected_contacts <- reactive({
        req(sel_fhir_version(), sel_vendor(), sel_has_contact())
        
        contacts_offset <- (contacts_page_state() - 1) * contacts_page_size
        
        query_str <- "SELECT * FROM mv_contacts_info WHERE fhir_version IN ({vals*})"
        params <- list(vals = sel_fhir_version())

        if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
            query_str <- paste0(query_str, " AND vendor_name = {vendor}")
            params$vendor <- sel_vendor()
        }

        # Add LIMIT OFFSET for pagination
        query_str <- paste0(query_str, " LIMIT {limit} OFFSET {offset}")
        params$limit <- contacts_page_size
        params$offset <- contacts_offset

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