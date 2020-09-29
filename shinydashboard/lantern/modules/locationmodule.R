# Location Module
library(leaflet)

locationmodule_UI <- function(id) {
  ns <- NS(id)
  tagList(
    h3("Map of Endpoints Linked to an Organization"),
    p("This map visualizes the locations of the API Information Sources which Lantern has associated with a FHIR endpoint by matching an API
    Information Source (organization name), as reported by a Certified API Developer, with an organization name in the National Payer and
    Provider Enumeration System (NPPES). Caution should be taken when gathering insights from this map as linking an API Information Source
    to an organization name in NPPES based on reported organization name may not be done with 100% confidence. See note below the map for more information."),
    p("The points on the map, below, represent the zip code associated with the primary address of matched organizations. The location reported by
     NPPES may not be the physical location of the API Information Source serviced by a given endpoint, may not represent a physical location where
     services are provided, or may not be the geolocation of any individual endpoint. This is especially true for API Information Sources which may
     have more than one physical location, which may vary by facility type and geographic location."),
    leafletOutput(ns("location_map"), width = "100%", height = "600px"),
    htmlOutput(ns("note_text"))
  )
}

locationmodule <- function(
  input,
  output,
  session,
  sel_fhir_version,
  sel_vendor
) {

  ns <- session$ns

  selected_fhir_endpoints <- reactive({
    res <- app_data$endpoint_locations
    req(sel_fhir_version(), sel_vendor())
    # If the selected dropdown value for the fhir verison is not the default "All FHIR Versions", filter
    # the capability statement fields by which fhir verison they're associated with
    if (sel_fhir_version() != ui_special_values$ALL_FHIR_VERSIONS) {
      res <- res %>% filter(fhir_version == sel_fhir_version())
    }
    # Same as above but with the developer dropdown
    if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
      res <- res %>% filter(vendor_name == sel_vendor())
    }
    res
  })


  output$location_map  <- renderLeaflet({
    map <- leaflet() %>%
      addProviderTiles(providers$CartoDB.Positron) %>%
      addCircles(data = selected_fhir_endpoints(), lat = ~ lat, lng = ~ lng, popup = ~endpoint_name,  weight = 10, color = "#33bb33", fillOpacity = 0.8, fillColor = "#00ff00") %>%
      setView(-98.9, 37.7, zoom = 4)
    map
  })

  output$note_text <- renderUI({
    note_info <- "These points only represent indexed endpoints which have been mapped to an
    organization with a match score greater than 0.97. The match scores are derived from the
    algorithms used by the Lantern application and are subject to change."
    res <- paste("<div style='font-size: 18px;'><b>Note:</b>", note_info, "</div>")
    HTML(res)
  })

}
