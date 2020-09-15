# Location Module
library(leaflet)

locationmodule_UI <- function(id) {
  ns <- NS(id)
  tagList(
    h3("Map of Zip Codes with identified organization"),
    leafletOutput(ns("location_map"), width = "100%", height = "600px"),
    p("Lantern uses organization information from the NPPES provider NPI registry. Points above are mapped
      to the zip code associated with the primary address of identified organizations. It does not necessarily
      represent a phyical location where services are provided or a geolocation of any individual endpoint."),
    p("Green points represent indexed endpoints which have been mapped to an organization with a match score greater than 0.97. These locations are
      the zip code associated with the primary location of the organization mapped to the endpoint. For more information about match scores and how they are calculated see the \"About Lantern page\".")
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
    # Same as above but with the vendor dropdown
    if (sel_vendor() != ui_special_values$ALL_VENDORS) {
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

}
