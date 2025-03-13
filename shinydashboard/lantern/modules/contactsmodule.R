library(DT)
library(purrr)
library(reactable)
library(htmlwidgets)

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
    
    # Load and process all data that matches the filters
    selected_contacts <- reactive({
        # Get current filter values
        current_fhir <- sel_fhir_version()
        current_vendor <- sel_vendor()
        current_has_contact <- sel_has_contact()
        
        req(current_fhir, current_vendor, current_has_contact)
        
        # Get all data with filters applied
        all_data <- app_data$contact_info_tbl()
        
        # Apply all filters immediately
        filtered_data <- all_data %>% 
            filter(fhir_version %in% current_fhir)
        
        if (current_vendor != ui_special_values$ALL_DEVELOPERS) {
            filtered_data <- filtered_data %>% filter(vendor_name == current_vendor)
        }
        
        if (current_has_contact != "Any") {
            has_contact_value <- current_has_contact == "True"
            filtered_data <- filtered_data %>% filter(has_contact == has_contact_value)
        }
        
        # Find best contact for each URL
        filtered_data <- filtered_data %>%
            filter(contact_rank == 1 | is.na(contact_rank)) %>%
            arrange(url, contact_rank)  # Ensure consistent ordering
        
        # Format the data for display
        filtered_data$linkurl <- paste0("<a class=\"lantern-url\" tabindex=\"0\" aria-label=\"Press enter to open a pop up modal containing additional information for this endpoint.\" onkeydown = \"javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)\" onclick=\"Shiny.setInputValue(\'endpoint_popup\',&quot;", filtered_data$url, "&&", filtered_data$requested_fhir_version, "&quot;,{priority: \'event\'});\">", filtered_data$url, "</a>")
        
        filtered_data$contact_name <- ifelse(is.na(filtered_data$contact_name), ifelse(is.na(filtered_data$contact_value), "-", "N/A"), filtered_data$contact_name)
        filtered_data$contact_type <- ifelse(is.na(filtered_data$contact_type), "-", filtered_data$contact_type)
        filtered_data$contact_value <- ifelse(is.na(filtered_data$contact_value), "-", filtered_data$contact_value)
        
        # Fix show_all links - ensure they use correct HTML formatting
        filtered_data$show_all <- ifelse(filtered_data$num_contacts > 1, 
                                      paste0("<a class=\"lantern-url\" tabindex=\"0\" aria-label=\"Press enter to show all contact information.\" onkeydown = \"javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)\" onclick=\"Shiny.setInputValue(\'show_contact_modal\',&quot;", filtered_data$url, "&quot;,{priority: \'event\'});\"> Show All Contacts </a>"), 
                                      "-")
        
        # Clean up endpoint names formatting using purrr instead of sapply
        filtered_data$condensed_endpoint_names <- map_chr(seq_len(nrow(filtered_data)), function(i) {
            names <- filtered_data$endpoint_names[i]
            if (is.null(names) || is.na(names) || names == "") {
                return("-")
            }
            
            # Remove curly braces, quotes, and backslashes
            names <- gsub("[{}\"\\\\]", "", names)
            
            # Check if we need to truncate
            parts <- strsplit(names, ";")[[1]]
            if (length(parts) > 5) {
                return(paste0(
                    paste0(head(parts, 5), collapse = ";"), 
                    "; ", 
                    paste0("<a class=\"lantern-url\" tabindex=\"0\" aria-label=\"Press enter to open a pop up modal containing the endpoint's entire list of API information source names.\" onkeydown = \"javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)\" onclick=\"Shiny.setInputValue(\'show_details\',&quot;", filtered_data$url[i], "&quot;,{priority: \'event\'});\"> Click For More... </a>")
                ))
            } else {
                return(names)
            }
        })
        
        filtered_data
    })

    output$contacts_table <- reactable::renderReactable({
        # Get data to display
        display_data <- selected_contacts()
        
        if (nrow(display_data) == 0) {
            return(
                reactable(
                    data.frame(Message = "No data matching the selected filters"),
                    pagination = FALSE,
                    searchable = FALSE
                )
            )
        }
        
        # Select only needed columns for display
        display_data <- display_data %>%
            select(linkurl, fhir_version, condensed_endpoint_names, vendor_name, 
                   has_contact, contact_name, contact_type, contact_value, 
                   contact_preference, show_all)
        
        # Render optimized table
        reactable(
            display_data,
            defaultColDef = colDef(align = "center"),
            columns = list(
                linkurl = colDef(name = "URL", minWidth = 300, html = TRUE),
                fhir_version = colDef(name = "FHIR Version", sortable = FALSE, aggregate = "unique"),
                condensed_endpoint_names = colDef(name = "API Information Source Name", 
                                               aggregate = "unique", minWidth = 200, 
                                               sortable = FALSE, html = TRUE),
                vendor_name = colDef(name = "Certified API Developer Name", 
                                   aggregate = "unique", minWidth = 110, 
                                   sortable = FALSE),
                has_contact = colDef(name = "Has Contact Information", aggregate = "unique"),
                contact_name = colDef(name = "Preferred Contact Name"),
                contact_type = colDef(name = "Preferred Contact Type"),
                contact_value = colDef(name = "Preferred Contact Info"),
                contact_preference = colDef(show = FALSE),
                show_all = colDef(name = "All Contacts", html = TRUE)  # Make sure html=TRUE is set
            ),
            striped = TRUE,
            searchable = TRUE,
            showSortIcon = TRUE,
            highlight = TRUE,
            defaultPageSize = 25,
            showPageSizeOptions = TRUE,
            pageSizeOptions = c(25, 50, 100, 250),
            minRows = 10,
            paginationType = "jump"
        )
    })
}