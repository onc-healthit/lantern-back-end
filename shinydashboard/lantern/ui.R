tags$a("Skip to Content", href = "#content", class = "show-on-focus")
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
                                      img(src = "images/GitHub-Mark-Light-32px.png", height = "60%", width = "60%", alt = "Github logo"),
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
    tags$script(HTML("
      $(document).ready(function() {
        $(\"header\").find(\"nav\").prepend(\"<a href='#content' class='show-on-focus'>Skip to Content</a>\");
      })
     ")
    ),
    tags$head(tags$style(HTML("
      .show-on-focus {     
        position: absolute;
        top: -10em;
        background: #fff;
        color: #112e51;
        display: block;
        font-weight: 600;
        
      }
      .show-on-focus:focus {  
        top: 5px;   
        position: absolute;
        background: #fff;
        color: #112e51;
        display: block;
        font-weight: 600;
        font-size: 20px;
      }
      .show-on-focus-resources {     
        position: absolute;
        top: -10em;
        background: #fff;
        color: #112e51;
        display: block;
        font-weight: 600;  
        width: 180px;  
      }
      .show-on-focus-resources:focus {  
        position: static;
        background: #fff;
        color: #112e51;
        display: block;
        font-weight: 600;
        font-size: 20px;
        width: 180px;
      }
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
      a:link {
        text-decoration: none;
      }
      a:visited {
        text-decoration: none;
      }
      a:hover {
        font-weight: bold;
      }
      button.dropdown-toggle {
        background-color: white!important;
        color: black;
      }
      button:hover {
        border: 3px solid!important;
      }      
      select:hover {
        border: 3px solid!important;
        background-color: white!important;
      }
      a:active {
        font-weight: bold;
      }
      button:active {
        border: 3px solid!important;
      }      
      a:focus-visible  {
        border: 4px solid!important;
        background-color: yellow!important;
        color: black!important;
      }
      button:focus-visible  {
        border: 4px solid!important;
        background-color: yellow!important;
        color: black!important;
      }
      select:focus-visible  {
        border: 4px solid!important;
        background-color: yellow;
        color: black!important; 
      }
      .selectize-input:hover {
        border: 3px solid!important;
      }
      .selectize-input:focus-visible {
        border: 4px solid!important;
        background-color: yellow!important;
        color: black!important; 
      }
      input:hover {
        border: 3px solid!important;
      }
      input:focus-visible {
        border: 4px solid!important;
        background-color: yellow!important;
        color: black!important; 
      }
      .rt-th:focus-visible {
        border: 4px solid!important;
        background-color: yellow!important;
        color: black!important; 
      }
      .rt-sort-header:hover {
        font-weight: bold;
      }
      .rt-td:focus-visible {
        border: 4px solid!important;
        background-color: yellow!important;
        color: black!important; 
      }
      .rt-td:focus-visible {
        border: 4px solid!important;
        background-color: yellow!important;
        color: black!important; 
      }
      #location_page-location_map:focus-visible {
        border: 4px solid!important;
      }
      table.dataTable thead .sorting:focus-visible  {
        border: 4px solid!important;
        background-color: yellow!important;
        color: black!important;
      }
      table.dataTable thead .sorting:hover {
        border: 2px solid!important;
      }
      a.btn {
        background-color: #1B5A7F!important;
        border: 1px solid black!important;
        color: white;
      }
      a.btn:focus-visible  {
        border: 4px solid black!important;
        background-color: yellow!important;
        color: black!important;
      }
      a.btn:hover {
        border: 2px solid black!important;
        font-weight: bold!important;
        color: white!important;
      }
      .action-button {
        background-color: #1B5A7F!important;
        border: 1px solid black!important;
        color: white!important;
      }
      .action-button:focus-visible  {
        border: 4px solid black!important;
        background-color: yellow!important;
        color: black!important;
      }
      .action-button:hover {
        border: 2px solid black!important;
        font-weight: bold!important;
        color: white!important;
      }

    "))),
    tags$script(HTML(
      "let elems = document.getElementsByClassName('content-wrapper');
      elems[0].setAttribute('role', 'main');
      elems[0].id = 'content'

      var e = document.getElementById('side_menu');

      var d = document.createElement('li');
      d.classList.add('sidebarMenuSelectedTabItem', 'shiny-bound-input');
      d.dataset.value = e.dataset.value;

      e.parentNode.replaceChild(d, e);
      e.remove();
      d.id = 'side_menu';
      "
    )),
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
              img(src = "images/lantern-logo@1x.png", width = "300px", alt = "Lantern Logo"),
              br(),
              div(
                class = "footer",
                includeHTML("about-lantern.html"),
                p("For information about the data sources, algorithms, and query intervals used by Lantern, please see the",
                a("documentation available here.", href = "Lantern_Data_Sources_And_Algorithms.pdf", target = "_blank", class = "lantern-url")),
                h3("Source Code"),
                p("The code behind Lantern can be found on GitHub ",
                a("here.", href = "https://github.com/onc-healthit/lantern-back-end", class = "lantern-url"))
              )
        )
    ),
    tags$footer(class = "footer",
      includeHTML("disclaimer.html")
    ),
    tags$script(HTML("
      let tabIndexObserver = new MutationObserver(function(mutations) {
        for (let mutation of mutations) {
          if (mutation.type === \"attributes\") {
            if (mutation.target.hasAttribute(\"tabindex\") && mutation.target.getAttribute(\"tabindex\") !== \"-5\") {
              mutation.target.removeAttribute(\"tabindex\");
            }
          }
        }
      });

      let tabPanes = document.getElementsByClassName(\"tab-pane\");
      for (let tab of tabPanes) {
        tabIndexObserver.observe(tab, {
          attributes: true,
          attributeFilter: [\"tabindex\"]
        });
      }
      
      let sideMenu = document.getElementsByClassName(\"sidebar-menu\")
      let sideMenuList = sideMenu[0].getElementsByTagName(\"li\")
      for (let liElem of sideMenuList) {
        let sideMenuLinks = liElem.getElementsByTagName(\"a\")
        if (sideMenuLinks.length > 0) {
          for (let aElem of sideMenuLinks) {
            tabIndexObserver.observe(aElem, {
              attributes: true,
              attributeFilter: [\"tabindex\"]
            });
          }
        }
      }

      let sideBarCollapsedObserver = new MutationObserver(function(mutations) {
        for (let mutation of mutations) {
          if (mutation.type === \"attributes\") {
            let sidebarMenu = document.getElementsByClassName(\"sidebar-menu\")
            let sidebarMenuList = sidebarMenu[0].getElementsByTagName(\"li\")
            for (let liElem of sidebarMenuList) {
              let sidebarMenuLinks = liElem.getElementsByTagName(\"a\")
              if (sidebarMenuLinks.length > 0) {
                for (let aElem of sidebarMenuLinks) {
                  if (mutation.target.getAttribute(\"data-collapsed\") === \"true\") {
                    aElem.setAttribute(\"tabindex\", \"-5\")
                  } else {
                    aElem.removeAttribute(\"tabindex\");
                  }
                }
              }
            }
          }
        }
      });

      let sideBarCollapsed = document.getElementById(\"sidebarCollapsed\")
      sideBarCollapsedObserver.observe(sideBarCollapsed, {
        attributes: true,
        attributeFilter: [\"data-collapsed\"]
      });

      let newNodesObserver = new MutationObserver(function(mutations) {
        for (let mutation of mutations) {
          if (mutation.addedNodes.length > 0) {
            for (let newNode of mutation.addedNodes) {
              
              if (mutation.target.id === \"show_date_filters\" && newNode.classList && newNode.classList.contains(\"row\")) {
                let selectInputNodes = newNode.querySelectorAll(\"select.shiny-bound-input\")
                for (let selectInputNode of selectInputNodes) {
                  selectInputNode.setAttribute('aria-label', 'Use the arrow keys to naviate the filter menu.')
                }
              }
              
              if (newNode.id === \"shiny-modal-wrapper\") {
                
                let modalTabPanes = newNode.getElementsByClassName(\"tab-pane\");
                for (let tab of modalTabPanes) {
                  tabIndexObserver.observe(tab, {
                    attributes: true,
                    attributeFilter: [\"tabindex\"]
                  });
                }

                let navBarTabs = document.getElementsByClassName(\"nav nav-tabs\");
                for (let navTab of navBarTabs) {
                  let liElements = navTab.getElementsByTagName(\"li\")
                  for (liElem of liElements) {
                    let aElements = liElem.getElementsByTagName(\"a\")
                    if (aElements.length > 0) {
                      tabIndexObserver.observe(aElements[0], {
                        attributes: true,
                        attributeFilter: [\"tabindex\"]
                      });
                    }
                  }
                }
              }

              if (newNode.className === \"field-list\") {
                let fieldsListTextSection = document.getElementById(\"fields_page-capstat_fields_text\");
                let fieldList = fieldsListTextSection.getElementsByClassName(\"field-list\")[0];
                let ulFieldList = fieldList.getElementsByTagName(\"ul\")[0];
                ulFieldList.removeAttribute(\"tabindex\");
              }

              if (newNode.className === \"extension-list\") {
                let fieldsListTextSection = document.getElementById(\"fields_page-capstat_extension_text\");
                let fieldList = fieldsListTextSection.getElementsByClassName(\"extension-list\")[0];
                let ulFieldList = fieldList.getElementsByTagName(\"ul\")[0];
                ulFieldList.removeAttribute(\"tabindex\");
              }   
            }

            if (mutation.addedNodes[0].classList && mutation.addedNodes[0].classList.contains(\"container-fluid\")) {
              let containerNode = mutation.addedNodes[0]
              let selectDropdowns = containerNode.querySelectorAll(\"select.shiny-bound-input\")
              for (selectDropdown of selectDropdowns) {
                selectDropdown.setAttribute('aria-label', 'Use the arrow keys to naviate the filter menu.')
              }
            } 
          }

          if (mutation.target.id === \"page_title\") {
            let pageTitleText = mutation.target.textContent
            if (pageTitleText === \"Organizations Page\" || pageTitleText === \"Resource Page\" || pageTitleText === \"Profile Page\") {
              
              let navBarTabs = document.getElementsByClassName(\"nav nav-tabs\");
              
              let tabbable = document.getElementsByClassName(\"tabbable\")
              let tabContents = tabbable[0].getElementsByClassName(\"tab-content\")
              
              for (let navTab of navBarTabs) {
                let liElements = navTab.getElementsByTagName(\"li\")
                for (liElem of liElements) {
                  let aElements = liElem.getElementsByTagName(\"a\")
                  if (aElements.length > 0) {
                    tabIndexObserver.observe(aElements[0], {
                      attributes: true,
                      attributeFilter: [\"tabindex\"]
                    });
                  }
                }
              }

              for (let tabContent of tabContents) {
                let tabPanes = tabContent.getElementsByClassName(\"tab-pane\")
                for (let tabPane of tabPanes) {
                  tabIndexObserver.observe(tabPane, {
                      attributes: true,
                      attributeFilter: [\"tabindex\"]
                    });
                }
              }            
            }
            
            let selectInputButtons = document.querySelectorAll(\"select.shiny-bound-input\")
            for (let selectInput of selectInputButtons) {
              selectInput.setAttribute('aria-label', 'Use the arrow keys to naviate the filter menu.')
            }
          }
          
          if (mutation.target.id === \"show_filters\") {
            let dropDownButtons = document.getElementsByClassName(\"dropdown-toggle\")
            for (let dropDownButton of dropDownButtons) {
              dropDownButton.setAttribute('aria-label', 'Dropdown filter menu button. Press the down arrow key to open the filter menu, use the tab or arrow keys to navigate through options, press enter to select a filter option, and use the escape key to close the filter menu.')
            }  
          }
        }
      })

      newNodesObserver.observe(document.body, {
          childList: true, 
          subtree: true, 
          attributes: false, 
          characterData: false
      })

      let tabContent = document.getElementsByClassName(\"tab-content\")
      newNodesObserver.observe(tabContent[0], {
          childList: true, 
          subtree: true, 
          attributes: false, 
          characterData: false
      })
    "))
  )
)
