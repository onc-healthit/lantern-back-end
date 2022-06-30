library(DT)
library(purrr)
library(reactable)

organizationsmodule_UI <- function(id) {

  ns <- NS(id)

  tagList(
    fluidRow(
      h2("Endpoint Organizations")
    ),
    tabsetPanel(id = "organization_tabset", type = "tabs",
              tabPanel("NPI Organizations",
                        h3("NPI Organization Matches"),
                        p("Endpoints can be linked to organizations in two ways, either by the NPI ID (preferred), or by the
                            organization name. Links made between organizations and endpoints using an
                            NPI ID are given a match confidence value of 100%, which is higher than any possible confidence
                            value for matches made using the organization name. In instances where a unique identifier to match an organization to an endpoint is not provided,
                            Lantern uses the organization name which each endpoint list provides, and the primary and
                            secondary organization names provided by the NPPES NPI data set to match npi organizations to endpoints
                            based on their names and assign a match confidence score. This table shows matches with a match confidence of 97% and up."),
                        reactable::reactableOutput(ns("npi_orgs_table"))),
              tabPanel("Endpoint List Organizations",
                        h3("Endpoint List Organization Matches"),
                        p("This table shows the organization name listed for each endpoint in the endpoint list it appears in."),
                        reactable::reactableOutput(ns("endpoint_list_orgs_table")))
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
  sel_confidence
) {
  ns <- session$ns

  selected_npi_orgs <- reactive({
    res <- get_npi_organization_matches()
    req(sel_fhir_version(), sel_vendor(), sel_confidence())

    res <- res %>% filter(fhir_version %in% sel_fhir_version())

    if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
      res <- res %>% filter(vendor_name == sel_vendor())
    }

    if (sel_confidence() != "97-100") {
      if (sel_confidence() == "100") {
        confidence_filter_num <- as.numeric(sel_confidence())
        res <- res %>% filter(match_score == confidence_filter_num)
      } else {
        confidence_upper_num <- as.numeric(strsplit(sel_confidence(), "-")[[1]][2])
        confidence_lower_num <- as.numeric(strsplit(sel_confidence(), "-")[[1]][1])

        res <- res %>% filter(match_score >= confidence_lower_num, match_score <= confidence_upper_num)
      }
    }

    res <- res %>%
    mutate(url = paste0("<a onclick=\"Shiny.setInputValue(\'endpoint_popup\',&quot;", url, "&&", requested_fhir_version, "&quot,{priority: \'event\'});\">", url, "</a>"))

    res
  })

 selected_endpoint_list_orgs <- reactive({
    res <- get_endpoint_list_matches()
    req(sel_fhir_version(), sel_vendor())

    res <- res %>% filter(fhir_version %in% sel_fhir_version())

    if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
      res <- res %>% filter(vendor_name == sel_vendor())
    }

    res <- res %>%
    mutate(url = paste0("<a onclick=\"Shiny.setInputValue(\'endpoint_popup\',&quot;", url, "&&", requested_fhir_version, "&quot,{priority: \'event\'});\">", url, "</a>"))

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
                  url = colDef(name = "URL", minWidth = 300, sortable = FALSE, html = TRUE),
                  npi_id = colDef(name = "NPI ID", sortable = FALSE),
                  zipcode = colDef(name = "Zipcode", sortable = FALSE),
                  organization_secondary_name = colDef(name = "Organization Secondary Name", sortable = FALSE),
                  fhir_version = colDef(name = "FHIR Version", sortable = FALSE),
                  vendor_name = colDef(name = "Certified API Developer Name", minWidth = 110, sortable = FALSE),
                  match_score = colDef(name = "Confidence", sortable = FALSE, aggregate = "count", format = list(aggregated = colFormat(prefix = "Total: ")))
              ),
              groupBy = c("organization_name", "url"),
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
                  url = colDef(name = "URL", minWidth = 300, sortable = FALSE, html = TRUE),
                  fhir_version = colDef(name = "FHIR Version", sortable = FALSE),
                  vendor_name = colDef(name = "Certified API Developer Name", minWidth = 110, sortable = FALSE, aggregate = "count", format = list(aggregated = colFormat(prefix = "Total: ")))
              ),
              groupBy = c("organization_name"),
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
