# Location Module
library(leaflet)

locationmodule_UI <- function(id) {
  ns <- NS(id)
  tagList(
    h3("Map of Zip Codes with identified organization"),
    p(strong("Demonstration Only:"), "This map is for demonstration purposes and is still a work in progress."),
    leafletOutput(ns("location_map"), width = "100%", height = "600px"),
    p("Lantern uses organization information from the NPPES provider NPI registry. Points above are mapped
      to the zip code associated with the primary address of identified organizations. It does not necessarily
      represent a phyical location where services are provided or a geolocation of any individual endpoint."),
    p("Green points represent indexed endpoints which have been mapped to an organization. These locations are
      the zip code associated with the primary location of the organization mapped to the endpoint.")
  )
}

locationmodule <- function(
  input,
  output,
  session
) {

  ns <- session$ns

  output$location_map  <- renderLeaflet({
    map <- leaflet() %>%
      addProviderTiles(providers$CartoDB.Positron) %>%
      addCircles(data = app_data$org_locations, lat = ~ lat, lng = ~ lng, popup = ~name,  weight = 7, color = "#ff6633", opacity = 0.3, fillOpacity = 0.3, fillColor = "#ff6633") %>%
      addCircles(data = app_data$endpoint_locations, lat = ~ lat, lng = ~ lng, popup = ~endpoint_name,  weight = 10, color = "#33bb33", fillOpacity = 0.8, fillColor = "#00ff00") %>%
      setView(-98.9, 37.7, zoom = 4)
    map
  })

}
