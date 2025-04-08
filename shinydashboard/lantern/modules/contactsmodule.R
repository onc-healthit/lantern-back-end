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
      reactable::reactableOutput(ns("contacts_table"))
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

    selected_contacts <- reactive({
        # Get data directly from the materialized view
        res <- get_contact_information(db_connection)
        req(sel_fhir_version(), sel_vendor(), sel_has_contact())

        res <- res %>% filter(fhir_version %in% sel_fhir_version())

        if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
            res <- res %>% filter(vendor_name == sel_vendor())
        }

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
                  linkurl = colDef(name = "URL", minWidth = 300, html = TRUE),
                  fhir_version = colDef(name = "FHIR Version", sortable = FALSE, aggregate = "unique"),
                  condensed_endpoint_names = colDef(name = "API Information Source Name", aggregate = "unique", minWidth = 200, sortable = FALSE, html = TRUE),
                  vendor_name = colDef(name = "Certified API Developer Name", aggregate = "unique", minWidth = 110, sortable = FALSE),
                  has_contact = colDef(name = "Has Contact Information", aggregate = "unique"),
                  contact_name = colDef(name = "Preferred Contact Name"),
                  contact_type = colDef(name = "Preferred Contact Type"),
                  contact_value = colDef(name = "Preferred Contact Info"),
                  contact_preference = colDef(show = FALSE),
                  show_all = colDef(name = "All Contacts", html = TRUE)
              ),
              striped = TRUE,
              searchable = TRUE,
              showSortIcon = TRUE,
              highlight = TRUE,
              defaultPageSize = 10
     )
    })
}