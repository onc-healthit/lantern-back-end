library(DT)
library(purrr)
library(reactable)
library(leaflet)

organizationsmodule_UI <- function(id) {

  ns <- NS(id)

  tagList(
    fluidRow(
      h2("Endpoint Organizations")
    ),
    tabsetPanel(id = "organization_tabset", type = "tabs",
              tabPanel("NPI Organizations",
                        h3("NPI Organization Matches"),
                        p("Endpoints can be linked to organizations in two ways, either by the National Provider Identifier (NPI)
                            as found in the National Payer and Provider Enumeration System (NPPES), which is preferred,
                            or by the organization name as reported by a Certified API Developer. Links made between organizations and endpoints using an
                            NPI ID are given a match confidence value of 100%, which is higher than any possible confidence
                            value for matches made using the organization name. In instances where a unique identifier to match an organization to an endpoint is not provided,
                            Lantern uses the organization name which each endpoint list provides, and the primary and
                            secondary organization names provided by the NPPES NPI data set to match npi organizations to endpoints
                            based on their names and assign a match confidence score. If a zipcode is included in the endpoint's endpoint list, it will be used in the matching to
                            try to increase the confidence of matches that have a confidence of 85% or higher. This table shows matches with a match confidence of 97% and up."),
                        htmlOutput(ns("map_anchor_link")),
                        reactable::reactableOutput(ns("npi_orgs_table")),
                        tagList(
                          h3("Map of Endpoints Linked to an Organization"),
                          p("This map visualizes the locations of endpoints which Lantern has associated with an organization
                          name in NPPES. An endpoint will have an entry on the map for each version of FHIR which it supports. Caution should be
                          taken when gathering insights from this map as linking an API Information Source to an organization name in NPPES based on reported organization
                          name may not be done with 100% confidence. See note below the map for more information."),
                          p("The points on the map, below, represent the zip code associated with the primary address of matched organizations. The location reported by
                          NPPES may not be the physical location of the API Information Source serviced by a given endpoint, may not represent a physical location where
                          services are provided, or may not be the geolocation of any individual endpoint. This is especially true for API Information Sources which may
                          have more than one physical location, which may vary by facility type and geographic location."),
                          htmlOutput(ns("map_anchor_point")),
                          leafletOutput(ns("location_map"), width = "100%", height = "600px"),
                          htmlOutput(ns("note_text_nppes_organizations"))
                        )),
              tabPanel("Endpoint List Organizations",
                        h3("Endpoint List Organization Matches"),
                        p("This table shows the organization name listed for each endpoint in the endpoint list it appears in."),
                        reactable::reactableOutput(ns("endpoint_list_orgs_table")),
                        htmlOutput(ns("note_text")))
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

  selected_npi_orgs <- reactive({
    res <- get_npi_organization_matches(db_tables)
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
    rowwise() %>%
    mutate(url = paste0("<a class=\"lantern-url\" tabindex=\"0\" aria-label=\"Press enter to open a pop up modal containing the endpoint's entire list of API information source names.\" onkeydown = \"javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)\" onclick=\"Shiny.setInputValue(\'endpoint_popup\',&quot;", url, "&&", requested_fhir_version, "&quot,{priority: \'event\'});\">", url, "</a>"))

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
    mutate(url = paste0("<a class=\"lantern-url\" tabindex=\"0\" aria-label=\"Press enter to open a pop up modal containing additional information for this endpoint.\" onkeydown = \"javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)\" onclick=\"Shiny.setInputValue(\'endpoint_popup\',&quot;", url, "&&", requested_fhir_version, "&quot,{priority: \'event\'});\">", url, "</a>"))

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

  selected_fhir_endpoints <- reactive({
    res <- isolate(app_data$endpoint_locations())
    req(sel_fhir_version(), sel_vendor())
    # If the selected dropdown value for the fhir version is not the default "All FHIR Versions", filter
    # the capability statement fields by which fhir version they're associated with
    res <- res %>% filter(fhir_version %in% sel_fhir_version())
    # Same as above but with the developer dropdown
    if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
      res <- res %>% filter(vendor_name == sel_vendor())
    }
    res
  })


  output$location_map  <- renderLeaflet({
    map <- leaflet() %>%
      addProviderTiles(providers$CartoDB.Positron) %>%
      addCircles(data = selected_fhir_endpoints(), lat = ~ lat, lng = ~ lng, popup = paste(isolate(selected_fhir_endpoints())$organization_name, "<br>", isolate(selected_fhir_endpoints())$url),  weight = 10, color = "#33bb33", fillOpacity = 0.8, fillColor = "#00ff00") %>%
      setView(-98.9, 37.7, zoom = 4)
    map
  })

  output$note_text_nppes_organizations <- renderUI({

    note_info <- "<br>(1) These points only represent indexed endpoints which have been mapped to an
    organization with a match score greater than 0.97. The match scores are derived from the
    algorithms used by the Lantern application and are subject to change.<br>
    (2) The endpoints queried by Lantern are limited to Fast Healthcare Interoperability
    Resources (FHIR) endpoints published publicly by Certified API Developers in conformance
    with the ONC Cures Act Final Rule, or discovered through the National Plan and Provider
    Enumeration System (NPPES). This data, therefore, may not represent all FHIR endpoints
    in existence. Insights gathered from this data should be framed accordingly.
    "

    res <- paste("<div style='font-size: 18px;'><b>Notes:</b>", note_info, "</div>")
    HTML(res)
  })

  output$map_anchor_point <- renderUI({
    HTML("<span id='mapanchorid'></span>")
  })

  output$map_anchor_link <- renderUI({
    HTML("<br><p>See the locations of endpoints on the map <a class=\"lantern-url\" href='#mapanchorid'>below</a></p>")
  })

}
