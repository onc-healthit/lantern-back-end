# Define base user interface
ui <- dashboardPage(
  dashboardHeader(
    title = "Lantern Dashboard",
    titleWidth = 200,
    tags$li(
      class = "dropdown",
        column(
          width = 12,
          align = "right",
          span(textOutput("version"),
               style = "color: white; font-size: 16px; line-height: 45px")
        )
      )
  ),
  # Sidebar with menu items for each module
  dashboardSidebar(
    sidebarMenu(id = "side_menu",
      menuItem("Dashboard", tabName = "dashboard_tab", icon = icon("dashboard")),
      menuItem("Endpoints", tabName = "endpoints_tab", icon = icon("table")),
      menuItem("Downloads", tabName = "downloads_tab", icon = icon("download")),
      menuItem("Capability", icon = icon("list-alt"), tabName = "capability_tab"),
      menuItem("Implementation Guides", tabName = "implementation_tab", icon = icon("list-alt")),
      menuItem("CapabilityStatement / Conformance Fields", icon = icon("list-alt"), tabName = "fields_tab"),
      menuItem("Values", icon = icon("table"), tabName = "values_tab"),
      menuItem("Performance", icon = icon("bar-chart-o"), tabName = "performance_tab"),
      menuItem("CapabilityStatement / Conformance Size", icon = icon("hdd-o"), tabName = "capabilitystatementsize_tab"),
      menuItem("Validations", icon = icon("clipboard-check"), tabName = "validations_tab", badgeLabel = "new"),
      menuItem("Security", icon = icon("id-card-o"), tabName = "security_tab"),
      menuItem("SMART Response", icon = icon("list"), tabName = "smartresponse_tab"),
      menuItem("Location", tabName = "location_tab", icon = icon("map")),
      menuItem("About Lantern", tabName = "about_tab", icon = icon("info-circle")),
      style = "white-space: normal"
    )
  ),

  # Set up contents for each menu item (tab) in the sidebar
  dashboardBody(
    tags$head(tags$style(HTML('
      .skin-blue .main-header .navbar {
        background-color: #1B5A7F;
      }
      .skin-blue .main-header .logo {
        background-color: #1B5A7F;
      }
      .small-box {
        color: black!important;
      }
      .badge {
        color: black!important;
      }
      .NA {
        color: #696464!important;
      }
      .lantern-url {
        color: #0044FF!important;
        text-decoration: underline;
      }
      .small-box p {
        font-size: 20px;
      }
      .sidebar-menu {
         border-bottom: 1px solid white;
      }
    '))),
    tags$head(tags$link(rel = "shortcut icon", href = "images/favicon.ico")),
    development_banner(devbanner),
    h1(textOutput("page_title")),
    uiOutput("show_filters"),
    uiOutput("show_date_filters"),
    uiOutput("show_value_filters"),
    uiOutput("show_resource_checkboxes"),
    uiOutput("show_operation_checkboxes"),
    tabItems(
      tabItem("dashboard_tab",
              dashboard_UI("dashboard_page")
      ),
      tabItem("endpoints_tab",
              endpointsmodule_UI("endpoints_page")
      ),
      tabItem("downloads_tab",
              downloadsmodule_UI("downloads_page")
      ),
      tabItem("performance_tab",
              performance_UI("performance_page")
      ),
      tabItem("capabilitystatementsize_tab",
              capabilitystatementsize_UI("capabilitystatementsize_page")
      ),
      tabItem("capability_tab",
              capabilitymodule_UI("capability_page")
      ),
      tabItem("implementation_tab",
              implementationmodule_UI("implementation_page")
      ),
      tabItem("fields_tab",
              fieldsmodule_UI("fields_page")
      ),
      tabItem("values_tab",
              valuesmodule_UI("values_page")
      ),
      tabItem("validations_tab",
              validationsmodule_UI("validations_page")
      ),
      tabItem("security_tab",
              securitymodule_UI("security_page")
      ),
      tabItem("smartresponse_tab",
              smartresponsemodule_UI("smartresponse_page")
      ),
      tabItem("location_tab",
              locationmodule_UI("location_page")
      ),
      tabItem("about_tab",
              img(src = "images/lantern-logo@1x.png", width = "300px"),
              br(),
              div(
                class = "footer",
                includeHTML("about-lantern.html"),
                p("For information about the data sources, algorithms, and query intervals used by Lantern, please see the",
                a("documentation available here.", href = "Lantern_Data_Sources_And_Algorithms.pdf", target = "_blank"))
              )
        )
    ),
    div(class = "footer",
      includeHTML("disclaimer.html")
    )
  )
)
