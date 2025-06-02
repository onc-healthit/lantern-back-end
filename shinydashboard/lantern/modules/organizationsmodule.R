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
      p("This table shows the organization name listed for each endpoint in the endpoint list it appears in."),
      reactable::reactableOutput(ns("endpoint_list_orgs_table")),
      htmlOutput(ns("note_text"))
    )
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
        mutate(url = paste0("<a class=\"lantern-url\" tabindex=\"0\" aria-label=\"Press enter to open a pop up modal containing additional information for this endpoint.\" onkeydown = \"javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)\" onclick=\"Shiny.setInputValue(\'endpoint_popup\',&quot;", url, "&&", fhir_version, "&quot,{priority: \'event\'});\">", url, "</a>"))
      
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

     reactable(
       display_data %>% 
         select(organization_name, organization_id, url, fhir_version, vendor_name) %>% 
         distinct(organization_name, organization_id, url, fhir_version, vendor_name) %>% 
         group_by(organization_name),
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
       searchable = TRUE,
       showSortIcon = TRUE,
       highlight = TRUE,
       defaultPageSize = 10,
       showPageSizeOptions = TRUE,
       pageSizeOptions = c(10, 25, 50, 100),
       minRows = 10,
       paginationType = "jump"
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