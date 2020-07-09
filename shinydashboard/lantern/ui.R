# Define base user interface
ui <- dashboardPage(
  dashboardHeader(
    string <- readChar('/VERSION', file.info('/VERSION')$size),
    titleStr <- paste("Lantern Dashboard ", string),
    title = titleStr,
    titleWidth = 200
  ),
  # Sidebar with menu items for each module
  dashboardSidebar(
    sidebarMenu(id="side_menu",
      menuItem("Dashboard", tabName = "dashboard_tab", icon = icon("dashboard"), selected = TRUE),
      menuItem("Endpoints", tabName = "endpoints_tab", icon = icon("table")),
      menuItem("Availability", icon = icon("th"), tabName = "availability_tab"),
      menuItem("Capability", icon=icon("list-alt"), tabName = "capability_tab", badgeLabel = "new", badgeColor = "green"),
      menuItem("Location", tabName = "location_tab", icon = icon("map")),
      menuItem("About Lantern", tabName = "about_tab", icon = icon("info-circle")),
      hr()
    )
  ),

  # Set up contents for each menu item (tab) in the sidebar
  dashboardBody(
    tags$head(tags$link(rel="shortcut icon", href="images/favicon.ico")),
    h1(textOutput("page_title")),
    uiOutput("show_filters"),
    tabItems(
      tabItem("dashboard_tab",
              dashboard_UI("dashboard_page")
      ),
      tabItem("endpoints_tab",
              endpointsmodule_UI("endpoints_page")
      ),
      tabItem("availability_tab",
              availability_UI("availability_page")
      ),
      tabItem("capability_tab",
              capabilitymodule_UI("capability_page")
      ),
      tabItem("location_tab",
              h3("Map of Zip Codes with identified endpoint/organization"),
              img(src = "images/endpoint_zcta_map.png", width = "100%"),
              p("This is a placeholder map for showing endpoints associated with a location.
                      Will be updated with interactive map with pins for endpoints")
      ),
      tabItem("about_tab",
              img(src = "images/lantern-logo@1x.png", width = "300px"),
              br(),
              includeHTML("about-lantern.html"),
              p("For information about the data sources, algorithms, and query intervals used by Lantern, please see the", a("documentation available here.", href= "Lantern_Data_Sources_And_Algorithms.pdf", target="_blank")))
    ),
    div(class = "footer",
    includeHTML("disclaimer.html")
    )
  )
)