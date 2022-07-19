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
    ),
    tags$li(a(href = "https://github.com/onc-healthit/lantern-back-end",
                                      img(src = "images/GitHub-Mark-Light-32px.png", height = "60%", width = "60%"),
                                      title = "Github Link"),
                                    class = "dropdown")
  ),
  # Sidebar with menu items for each module
  dashboardSidebar(
    sidebarMenu(id = "side_menu",
      menuItem("Dashboard", tabName = "dashboard_tab", icon = tags$i(class = "fa fa-dashboard", "aria-hidden" = "true", role = "presentation", "aria-label" = "dashboard icon")),
      menuItem("Endpoints", tabName = "endpoints_tab", icon = tags$i(class = "fa fa-table", "aria-hidden" = "true", role = "presentation", "aria-label" = "table icon")),
      menuItem("Organizations", tabName = "organizations_tab", icon = tags$i(class = "fa fa-hospital", "aria-hidden" = "true", role = "presentation", "aria-label" = "hospital icon")),
      menuItem("Resources", icon = tags$i(class = "fa fa-list-alt", "aria-hidden" = "true", role = "presentation", "aria-label" = "list-alt icon"), tabName = "resource_tab"),
      menuItem("Implementation Guides", tabName = "implementation_tab", icon = tags$i(class = "fa fa-list-alt", "aria-hidden" = "true", role = "presentation", "aria-label" = "list-alt icon")),
      menuItem("CapabilityStatement / Conformance Fields", icon = tags$i(class = "fa fa-list-alt", "aria-hidden" = "true", role = "presentation", "aria-label" = "list-alt icon"), tabName = "fields_tab"),
      menuItem("CapabilityStatement / Conformance Profiles", icon = tags$i(class = "fa fa-list-alt", "aria-hidden" = "true", role = "presentation", "aria-label" = "list-alt icon"), tabName = "profile_tab"),
      menuItem("Values", icon = tags$i(class = "fa fa-table", "aria-hidden" = "true", role = "presentation", "aria-label" = "table icon"), tabName = "values_tab"),
      menuItem("CapabilityStatement / Conformance Size", icon = tags$i(class = "fa fa-hdd-o", "aria-hidden" = "true", role = "presentation", "aria-label" = "hdd-o icon"), tabName = "capabilitystatementsize_tab"),
      menuItem("Validations", icon = tags$i(class = "fa fa-clipboard-check", "aria-hidden" = "true", role = "presentation", "aria-label" = "clipboard-check icon"), tabName = "validations_tab"),
      menuItem("Security", icon = tags$i(class = "fa fa-id-card-o", "aria-hidden" = "true", role = "presentation", "aria-label" = "id-card-o icon"), tabName = "security_tab"),
      menuItem("SMART Response", icon = tags$i(class = "fa fa-list", "aria-hidden" = "true", role = "presentation", "aria-label" = "list icon"), tabName = "smartresponse_tab"),
      menuItem("Location", tabName = "location_tab", icon = tags$i(class = "fa fa-map", "aria-hidden" = "true", role = "presentation", "aria-label" = "map icon")),
      menuItem("Contact Information", tabName = "contacts_tab", icon = tags$i(class = "fa fa-list-alt", "aria-hidden" = "true", role = "presentation", "aria-label" = "list-alt icon")),
      menuItem("Downloads", tabName = "downloads_tab", icon = tags$i(class = "fa fa-download", "aria-hidden" = "true", role = "presentation", "aria-label" = "download icon")),
      menuItem("About Lantern", tabName = "about_tab", icon = tags$i(class = "fa fa-info-circle", "aria-hidden" = "true", role = "presentation", "aria-label" = "info-circle icon")),
      style = "white-space: normal"
    )
  ),

  # Set up contents for each menu item (tab) in the sidebar
  dashboardBody(
    tags$head(tags$style(HTML("
      .content-wrapper, .right-side {
        background-color: #F6F7F8;
      }
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
      #dashboard_page-show_info {
        color: black;
      }
      .modal-lg {
        width: 75%!important;
      }
      #fields_page-capstat_fields_text{
        color: black!important;
      }
      #organization_tabset li a {
        color: #024A96;
      }
      #resource_tabset li a {
        color: #024A96;
      }
      #profile_resource_tab li a {
        color: #024A96;
      }
      .nav-tabs>li.active>a, .nav-tabs>li.active>a:focus, .nav-tabs>li.active>a:hover {
        color: #555!important;
      }
      .multi-wrapper div a {
        color: #024A96;
      }
      .multi-wrapper .non-selected-wrapper .item.selected {
        color: #4F4F4F;
        opacity: 1!important;
      }
    "))),
    tags$script(
      "let elems = document.getElementsByClassName('content-wrapper');
        elems[0].setAttribute('role', 'main');

        var e = document.getElementById('side_menu');

        var d = document.createElement('li');
        d.classList.add('sidebarMenuSelectedTabItem', 'shiny-bound-input')
        d.dataset.value = e.dataset.value;

        e.parentNode.replaceChild(d, e);
        e.remove();
        d.id = 'side_menu'
        "
    ),
    tags$head(tags$link(rel = "shortcut icon", href = "images/favicon.ico")),
    development_banner(devbanner),
    uiOutput("resource_tab_popup"),
    h1(textOutput("page_title")),
    uiOutput("show_filters"),
    uiOutput("show_value_filters"),
    uiOutput("show_resource_operation_checkboxes"),
    uiOutput("show_resource_profiles_dropdown"),
    uiOutput("organizations_filter"),
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
      tabItem("organizations_tab",
              organizationsmodule_UI("organizations_page")
      ),
      tabItem("capabilitystatementsize_tab",
              capabilitystatementsize_UI("capabilitystatementsize_page")
      ),
      tabItem("resource_tab",
              resourcemodule_UI("resource_page")
      ),
      tabItem("implementation_tab",
              implementationmodule_UI("implementation_page")
      ),
      tabItem("fields_tab",
              fieldsmodule_UI("fields_page")
      ),
        tabItem("profile_tab",
              profilemodule_UI("profile_page")
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
      tabItem("contacts_tab",
              contactsmodule_UI("contacts_page")
      ),
      tabItem("about_tab",
              img(src = "images/lantern-logo@1x.png", width = "300px"),
              br(),
              div(
                class = "footer",
                includeHTML("about-lantern.html"),
                p("For information about the data sources, algorithms, and query intervals used by Lantern, please see the",
                a("documentation available here.", href = "Lantern_Data_Sources_And_Algorithms.pdf", target = "_blank")),
                h3("Source Code"),
                p("The code behind Lantern can be found on GitHub ",
                a("here.", href = "https://github.com/onc-healthit/lantern-back-end"))
              )
        )
    ),
    tags$footer(class = "footer",
      includeHTML("disclaimer.html")
    )
  )
)
