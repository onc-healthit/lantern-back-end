library(DT)
library(purrr)
library(reactable)

organizationsmodule_UI <- function(id) {

  ns <- NS(id)

  tagList(
    fluidRow(
      h2("Endpoint Organizations"),
      column(width = 12, style = "padding-bottom:20px",
             h3(style = "margin-top:0", textOutput(ns("endpoint_count")))
      ),
    ),
    tabsetPanel(id = "organization_tabset", type = "tabs",
              tabPanel("NPI Organizations", h3("NPI Organization Matches"), reactable::reactableOutput(ns("npi_orgs_table"))),
              tabPanel("Endpoint List Organizations", h3("Endpoint List Organization Matches"),reactable::reactableOutput(ns("endpoint_list_orgs_table")))
    ),
    htmlOutput(ns("note_text"))
  )
}

organizationsmodule <- function(
  input,
  output,
  session,
  sel_fhir_version,
  sel_vendor,
  sel_availability
) {
  ns <- session$ns

  selected_npi_orgs <- reactive({
    res <- get_npi_organization_matches()
    req(sel_fhir_version(), sel_vendor())

    res <- res %>% filter(fhir_version %in% sel_fhir_version())

    if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
      res <- res %>% filter(vendor_name == sel_vendor())
    }

    res
  })

 selected_endpoint_list_orgs <- reactive({
    res <- get_endpoint_list_matches()
    req(sel_fhir_version(), sel_vendor())

    res <- res %>% filter(fhir_version %in% sel_fhir_version())

    if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
      res <- res %>% filter(vendor_name == sel_vendor())
    }

    res
  })

  output$npi_orgs_table <- reactable::renderReactable({
     reactable(
              selected_npi_orgs() %>% select(organization_name, url, npi_id, zipcode,  organization_secondary_name, fhir_version, vendor_name, match_score) %>% distinct(organization_name, url, npi_id, zipcode,  organization_secondary_name, fhir_version, vendor_name, match_score) %>% group_by(organization_name),
              defaultColDef = colDef(
                align = "center"
              ),
              columns = list(
                  organization_name = colDef(name = "Organization Name", sortable = TRUE, align = "left"),
                  url = colDef(name = "URL", minWidth = 300, sortable = FALSE),
                  npi_id = colDef(name = "NPI ID", sortable = FALSE),
                  zipcode = colDef(name = "Zipcode", sortable = FALSE),
                  organization_secondary_name = colDef(name = "Organization Secondary Name", sortable = FALSE),
                  fhir_version = colDef(name = "FHIR Version", sortable = FALSE),
                  vendor_name = colDef(name = "Certified API Developer Name", minWidth = 110, sortable = FALSE),
                  match_score = colDef(name = "Confidence", sortable = FALSE, aggregate = "count", format = list(aggregated = colFormat(prefix = "Total: ")))
              ),
              groupBy = c('organization_name', 'url'),
              striped = TRUE,
              searchable = TRUE,
              showSortIcon = TRUE,
              highlight = TRUE,
              defaultPageSize = 10
     )
  })

    output$endpoint_list_orgs_table <- reactable::renderReactable({
     reactable(
              selected_endpoint_list_orgs() %>% select(organization_name, url, fhir_version, vendor_name) %>% distinct(organization_name, url, fhir_version, vendor_name) %>% group_by(organization_name),
              defaultColDef = colDef(
                align = "center"
              ),
              columns = list(
                  organization_name = colDef(name = "Organization Name", sortable = TRUE, align = "left"),
                  url = colDef(name = "URL", minWidth = 300, sortable = FALSE),
                  fhir_version = colDef(name = "FHIR Version", sortable = FALSE),
                  vendor_name = colDef(name = "Certified API Developer Name", minWidth = 110, sortable = FALSE, aggregate = "count", format = list(aggregated = colFormat(prefix = "Total: ")))
              ),
              groupBy = c('organization_name'),
              striped = TRUE,
              searchable = TRUE,
              showSortIcon = TRUE,
              highlight = TRUE,
              defaultPageSize = 10
     )
  })

  output$note_text <- renderUI({
    note_info <- "The endpoints queried by Lantern are limited to Fast Healthcare Interoperability
      Resources (FHIR) endpoints published publicly by Certified API Developers in conformance
      with the ONC Cures Act Final Rule, or discovered through the National Plan and Provider
      Enumeration System (NPPES). This data, therefore, may not represent all FHIR endpoints
      in existence. Insights gathered from this data should be framed accordingly."
    res <- paste("<div style='font-size: 18px;'><b>Note:</b>", note_info, "</div>")
    HTML(res)
  })

}
