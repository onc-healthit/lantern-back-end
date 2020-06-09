# Define base user interface

ui <- dashboardPage(
  dashboardHeader(
    title = "Lantern Dashboard",
    titleWidth = 200
  ),
  # Sidebar with menu items for each module
  dashboardSidebar(
    sidebarMenu(
      menuItem("Dashboard", tabName = "dashboard_tab", icon = icon("dashboard"),selected=TRUE),
      menuItem("Endpoints", tabName = "endpoints_tab", icon = icon("table"), badgeLabel = "new",
               badgeColor = "green"),
      menuItem("Availability", icon = icon("th"), tabName = "availability_tab"
      ),
      menuItem("Performance", icon = icon("bar-chart-o"),
               menuSubItem("Mean Response Time", tabName = "performance_tab")
      ),
      menuItem("Location", tabName = "location_tab", icon=icon("map")),
      menuItem("About Lantern",tabName = "about_tab", icon=icon("info-circle")),
      hr()
    )
  ),
  
  # Set up contents for each menu item (tab) in the sidebar
  dashboardBody(
    tabItems(
      tabItem("dashboard_tab",
              h1("Current Endpoint Metrics"),
              dashboard_UI("dashboard_page")
      ),
      tabItem("endpoints_tab",
              h1("List of Endpoints"),
              endpointsmodule_UI("endpoints_page")
      ),
      tabItem("performance_tab",
              performance_UI("performance_page")
      ),
      tabItem("availability_tab",
              availability_UI("availability_page")
      ),
      tabItem("location_tab",
              h3("Map of Zip Codes with identified endpoint/organization"),
              img(src="images/endpoint_zcta_map.png",width="100%"),
              p("This is a placeholder map for showing endpoints associated with a location.
                      Will be updated with interactive map with pins for endpoints")
      ),
      tabItem("about_tab",
              h1("About Lantern"),
              img(src="images/lantern-logo@1x.png",width="300px"),
              p("This is a description of Lantern, the dashboard, the project, etc. "))
    )
  )
)
