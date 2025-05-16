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
      menuItem("Contact Information", tabName = "contacts_tab", icon = tags$i(class = "fa fa-list-alt", "aria-hidden" = "true", role = "presentation", "aria-label" = "list-alt icon")),
      menuItem("Downloads", tabName = "downloads_tab", icon = tags$i(class = "fa fa-download", "aria-hidden" = "true", role = "presentation", "aria-label" = "download icon")),
      menuItem("About Lantern", tabName = "about_tab", icon = tags$i(class = "fa fa-info-circle", "aria-hidden" = "true", role = "presentation", "aria-label" = "info-circle icon")),
      style = "white-space: normal"
    )
  ),

  # Set up contents for each menu item (tab) in the sidebar
  dashboardBody(
    tags$head(tags$script(HTML("
    (function (w, d, s, l, i) {
    w[l] = w[l] || []; w[l].push({
      'gtm.start':
        new Date().getTime(), event: 'gtm.js'
      }); var f = d.getElementsByTagName(s)[0],
      j = d.createElement(s), dl = l != 'dataLayer' ? '&l=' + l : ''; j.async = true; j.src =
        'https://www.googletagmanager.com/gtm.js?id=' + i + dl; f.parentNode.insertBefore(j, f);
    })(window, document, 'script', 'dataLayer', 'GTM-KC3FP96');
    "))),
    tags$noscript(tags$iframe(src = "https://www.googletagmanager.com/ns.html?id=GTM-KC3FP96", height = "0", width = "0", style = "display:none;visibility:hidden")),
    tags$script(HTML("
      // Add hidden skip to content button at the top of the Lantern website
      $(document).ready(function() {
        $(\"header\").find(\"nav\").prepend(\"<a href='#content' aria-label='Click the enter key to skip to the main content of this page, skipping over the header elements and navigation tabs.' class='show-on-focus'>Skip to Content</a>\");
      })
     ")
    ),
    tags$head(tags$style(HTML("
      /* Hides the Skip To Content button when it is not focused on */
      .show-on-focus {     
        position: absolute;
        top: -10em;
        background: #fff;
        color: #112e51;
        display: block;
        font-weight: 600;
        
      }

      /* Makes the Skip To Content button appear when it is focused on */
      .show-on-focus:focus {  
        top: 5px;   
        position: absolute;
        background: #fff;
        color: #112e51;
        display: block;
        font-weight: 600;
        font-size: 20px;
      }

      /* Hides the Skip Past Resources button when it is not focused on */
      .show-on-focus-resources {     
        position: absolute;
        top: -200em;
        background: #fff;
        color: #112e51;
        display: block;
        font-weight: 600;  
        width: 180px;  
      }

      /* Makes the Skip To Resources button appear when it is focused on */
      .show-on-focus-resources:focus {  
        position: static;
        background: #fff;
        color: #112e51;
        display: block;
        font-weight: 600;
        font-size: 20px;
        width: 180px;
      }

      /* Changes the background color of the main Lantern content background*/
      .content-wrapper, .right-side {
        background-color: #F6F7F8;
      }

      /* Changes the color of the Lantern header, navbar, and borders */
      .skin-blue .main-header .navbar {
        background-color: #1B5A7F;
      }

      /* Changes the color of the Lantern main header logo */
      .skin-blue .main-header .logo {
        background-color: #1B5A7F;
      }

      /* Changes the text color of the Current Endpoint Responses info boxes on the dashboard page */
      .small-box {
        color: black!important;
      }

      /* Changes the font size of the text within the Current Endpoint Responses info boxes on the dashboard page */
      .small-box p {
        font-size: 20px;
      }

      /* Changes the text color of the New badges that are added to new sidebar tabs */
      .badge {
        color: black!important;
      }

      /* Changes the text color of NA text elements */
      .NA {
        color: #696464!important;
      }

      /* Styling for all the urls on the Lantern page with class lantern-url */
      .lantern-url {
        color: #0044FF!important;
        text-decoration: underline;
        cursor: pointer;
      }

      /* Adds a border to the bottom of the Lantern sidebar */
      .sidebar-menu {
         border-bottom: 1px solid white;
      }

      /* Changes the width of the popup modals */
      .modal-lg {
        width: 75%!important;
      }

      /* Changes the text color of the fields list on the Capability Statement Fields tab */
      #fields_page-capstat_fields_text{
        color: black!important;
      }

      /* Changes the text color of the navbar tabs on the Organization page */
      #organization_tabset li a {
        color: #024A96;
      }

      /* Changes the text color of the navbar tabs on the Resource page */
      #resource_tabset li a {
        color: #024A96;
      }

      /* Changes the text color of the navbar tabs on the Profile page */
      #profile_resource_tab li a {
        color: #024A96;
      }

      /* Changes the text color of the navbar tabs on the Endpoint modal popup */
      #endpoint_modal_tabset li a {
        color: #024A96;
      }

      /* Change the text color of the active navbar tab on the Organization page, Resource page, Profile page, and Endpoint Modal */
      .nav-tabs>li.active>a {
        color: #555!important;
      }
   
      /* Change color of the selected links in the resources checkbox */
      .multi-wrapper .selected-wrapper .item.selected  {
        color: #024A96;
      }

      /* Change color of the non selected links in the resources checkbox */
      .multi-wrapper .non-selected-wrapper .item.selected {
        color: #4F4F4F;
        opacity: 1!important;
      }
      
      /* Remove underline from a tag elements when they are visted */
      a:visited {
        text-decoration: none;
      }

      /* Bold a tag elements when they are hovered on */
      a:hover {
        font-weight: bold;
      }

      /* Change coloring for FHIR version dropdown */
      button.dropdown-toggle {
        background-color: white!important;
        color: black;
      }

      /* Add a border when a button is hovered over */
      button:hover {
        border: 3px solid!important;
      }     

      /* Add a border and when a select tag (dropdowns) is hovered over and make sure background stays white */
      select:hover {
        border: 3px solid!important;
        background-color: white!important;
        cursor: pointer;
      }

      /* Bold an a tag element when it is active */
      a:active {
        font-weight: bold;
      }

      /* Add a border to button when it is active */
      button:active {
        border: 3px solid!important;
      } 

      /* When an a tag is focused on, add a border, change background color to yellow, and change text font to black */
      a:focus-visible  {
        border: 4px solid!important;
        background-color: yellow!important;
        color: black!important;
      }

      /* When a button is focused on, add a border, change background color to yellow, and change text font to black */
      button:focus-visible  {
        border: 4px solid!important;
        background-color: yellow!important;
        color: black!important;
      }

      /* When an select tag (dropdowns) is focused on, add a border, change background color to yellow, and change text font to black */
      select:focus-visible  {
        border: 4px solid!important;
        background-color: yellow;
        color: black!important; 
      }

      /* Add a border when an input tag (textbox) is hovered over */
      input:hover {
        border: 3px solid!important;
      }

      /* When an input tag (textbox) is focused on, add a border, change background color to yellow, and change text font to black */
      input:focus-visible {
        border: 4px solid!important;
        background-color: yellow!important;
        color: black!important; 
      }

      /* When a sortable reactable table header is focused on, add a border, change background color to yellow, and change text font to black */
      .rt-th:focus-visible {
        border: 4px solid!important;
        background-color: yellow!important;
        color: black!important; 
      }

      /* Make text bold when a sortable reactable table header is hovered over */
      .rt-sort-header:hover {
        font-weight: bold;
      }

      /* When an element in a table is focused on, add a border, change background color to yellow, and change text font to black */
      .rt-td:focus-visible {
        border: 4px solid!important;
        background-color: yellow!important;
        color: black!important; 
      }

      /* When the organizations location map is focused on, add a border around it */
      #organizations_page-location_map:focus-visible {
        border: 4px solid!important;
      }

      /* When an element in a dataTable sortable header is focused on, add a border, change background color to yellow, and change text font to black */
      table.dataTable thead .sorting:focus-visible  {
        border: 4px solid!important;
        background-color: yellow!important;
        color: black!important;
      }

      /* Make text bold when a dataTable sortable header is hovered over */
      table.dataTable thead .sorting:hover {
        border: 2px solid!important;
      }

      /* Change the styling for download buttons */
      a.btn {
        background-color: #1B5A7F!important;
        border: 1px solid black!important;
        color: white;
      }

      /* When a download button is focused on, add a border, change background color to yellow, and change text font to black */
      a.btn:focus-visible  {
        border: 4px solid black!important;
        background-color: yellow!important;
        color: black!important;
      }

      /* When a download button is hovered over, add a border, change text color to white, and bold the text */
      a.btn:hover {
        border: 2px solid black!important;
        font-weight: bold!important;
        color: white!important;
      }

      /* Change the styling for the buttons on the Lantern page */
      .action-button {
        background-color: #1B5A7F!important;
        border: 1px solid black!important;
        color: white!important;
      }

      /* When a button is focused on, add a border, change background color to yellow, and change text font to black */
      .action-button:focus-visible  {
        border: 4px solid black!important;
        background-color: yellow!important;
        color: black!important;
      }

      /* When a button is hovered over, add a border, change text color to white, and bold the text */
      .action-button:hover {
        border: 2px solid black!important;
        font-weight: bold!important;
        color: white!important;
      }

    "))),
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
    uiOutput("htmlFooter"),
    tags$script(HTML("

      // Set the role of Lantern's main content containing element to 'main'
      let elems = document.getElementsByClassName('content-wrapper');
      elems[0].setAttribute('role', 'main');
      elems[0].id = 'content'


      // Create a new li element with all the same attributes as existing side menu element
      var e = document.getElementById('side_menu');
      var d = document.createElement('li');
      d.classList.add('sidebarMenuSelectedTabItem', 'shiny-bound-input');
      d.dataset.value = e.dataset.value;

      // Replace old div side menu element with new li element created above
      e.parentNode.replaceChild(d, e);
      e.remove();
      d.id = 'side_menu';

      /*
        // Create an observer that watches if an element's tabindex attribute has been set or altered
        // If the altered attribute is the tabindex attribute, and its value is not -5, remove the tabindex attribute
        // Tabindex is removed from elements for two different reasons: 
          // 1. R Shiny automatically adds a tabindex of some elements, like the sidebar tab items, to -1 when they are clicked off of, meaning you can no longer tab to those items. 
                Removing the tabindex solves this problem. Some elements are focusable by default, and thus we can remove the tab index attribute entirely.
          // 2. R Shiny automatically adds some tabindex attributes to elements we do not want to be focusable, so the attribute is removed on these elements.
      */
      let tabIndexObserver = new MutationObserver(function(mutations) {
        for (let mutation of mutations) {
          if (mutation.type === \"attributes\") {
            if (mutation.target.hasAttribute(\"tabindex\") && mutation.target.getAttribute(\"tabindex\") !== \"-5\") {
              mutation.target.removeAttribute(\"tabindex\");
            }
          }
        }
      });

      // Set the tabindex observer on each of Lantern's tab pages which will remove the tab index attribute whenever it is set or altered 
      let tabPanes = document.getElementsByClassName(\"tab-pane\");
      for (let tab of tabPanes) {
        tabIndexObserver.observe(tab, {
          attributes: true,
          attributeFilter: [\"tabindex\"]
        });
      }
      
      let sideMenu = document.getElementsByClassName(\"sidebar-menu\")
      let sideMenuList = sideMenu[0].getElementsByTagName(\"li\")
      
      /*
        // Set the tabindex observer on each of Lantern's sidebar tabs which will remove the tab index attribute
        // whenever it is set due to clicking on another tab
      */ 
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

      /*
        // Create an observer that watches for attribute changes on the sidebar to catch when it is opened or closed
        // If the sidebar is closed, set the tab index to -5 so that it is not tabbable
        // If the sidebar is open, remove the tab index attribute
      */
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

      /* 
        // Set the tabindex observer on the sidebar element which will remove the tabindex attribute
        // from the sidebar elements when the sidebar is opened, and it will add a tabindex of -5 when the sidebar 
        // is collapsed
      */
      let sideBarCollapsed = document.getElementById(\"sidebarCollapsed\")
      sideBarCollapsedObserver.observe(sideBarCollapsed, {
        attributes: true,
        attributeFilter: [\"data-collapsed\"]
      });

      // Create an observer that watches for when any new nodes are created or any existing elements are altered
      let newNodesObserver = new MutationObserver(function(mutations) {   
        for (let mutation of mutations) {
          if (mutation.addedNodes.length > 0) {
            for (let newNode of mutation.addedNodes) {
              
              /*
                // If the mutated node is the response time date filter within the endpoint modal popup and a new node was added to it which was a row element, this means the row containing the date filter dropdown was added
                // Adds a new aria label to the date filter dropdown that explains you can use the arrow keys to navigate the filter dropdown
              */
              if (mutation.target.id === \"show_date_filters\" && newNode.classList && newNode.classList.contains(\"row\")) {
                let selectInputNodes = newNode.querySelectorAll(\"select.shiny-bound-input\")
                for (let selectInputNode of selectInputNodes) {
                  selectInputNode.setAttribute('aria-label', 'Use the arrow keys to navigate the filter menu.')
                }
              }

              /*
                // If the mutated node is the http response date filter within the endpoint modal popup and a new node was added to it which was a row element, this means the row containing the http date filter dropdown was added
                // Adds a new aria label to the date filter dropdown that explains you can use the arrow keys to navigate the filter dropdown
              */
              if (mutation.target.id === \"show_http_date_filters\" && newNode.classList && newNode.classList.contains(\"row\")) {
                let selectInputNodes = newNode.querySelectorAll(\"select.shiny-bound-input\")
                for (let selectInputNode of selectInputNodes) {
                  selectInputNode.setAttribute('aria-label', 'Use the arrow keys to navigate the filter menu.')
                }
              }
              
              // Checks if the new node that was added was the shiny modal popup 
              if (newNode.id === \"shiny-modal-wrapper\") {
                
                // Set the tabindex observer on each of the modal popup's tab pages which will remove the tab index attribute whenever it is set or altered 
                let modalTabPanes = newNode.getElementsByClassName(\"tab-pane\");
                for (let tab of modalTabPanes) {
                  tabIndexObserver.observe(tab, {
                    attributes: true,
                    attributeFilter: [\"tabindex\"]
                  });
                }
                
                /* 
                  // Get all the navigation bar tabs on the endpoint modal popup and alter them to have the same 
                  // structure, classes, and attributes as the navigation bars on the rest of the Lantern pages
                */
                let navBarTabs = document.getElementsByClassName(\"nav nav-tabs\");
                for (let navTab of navBarTabs) {
                  
                  // Add the tablist role to the main navbar containing element
                  navTab.setAttribute(\"role\", \"tablist\")
                  
                  let navTabID = navTab.getAttribute(\"data-tabsetid\")
                  
                  // Add the shiny-tab-input and shiny-bound-input classes to the main navbar containing element
                  navTab.classList.add(\"shiny-tab-input\", \"shiny-bound-input\")
                  
                  let liElements = navTab.getElementsByTagName(\"li\")
                  for (let liElem of liElements) {
                    
                    // Add the presentation role to each of the navigation bar tabs
                    liElem.setAttribute(\"role\", \"presentation\")
                    
                    // Add the tab index observer to each of the navigation bar tab links to remove the tabindex attribute whenever it is set or altered
                    let aElements = liElem.getElementsByTagName(\"a\")
                    if (aElements.length > 0) {
                      tabIndexObserver.observe(aElements[0], {
                        attributes: true,
                        attributeFilter: [\"tabindex\"]
                      });
                      
                      /*
                        // Set the role of each tab element link to tab
                        // Set aria-selected to true
                        // Set the aria-control attribute equal to the id of the main navbar containing element
                        // Set aria-expanded to true
                        // Set the aria label of the link to explain you can click on the current tab 
                      */
                      for (let aElem of aElements) {
                        aElem.setAttribute(\"role\", \"tab\")
                        aElem.setAttribute(\"aria-selected\", \"true\")
                        aElem.setAttribute(\"aria-controls\", \"tab-\" + navTabID + \"-1\")
                        aElem.setAttribute(\"aria-expanded\", \"true\")
                        aElem.setAttribute(\"aria-label\", \"Press enter to select the \" + aElem.textContent +\" tab and show this tab's content below\")
                      }
                    }
                  }
                }
                
                // Set the aria label of each of the accordion dropdown panels in the endpoint modal popup to explain you can click on them to open and close
                let bsCollapses = newNode.getElementsByClassName(\"panel-group\")
                for (bsCollapse of bsCollapses) {
                  let tabInfos = bsCollapse.getElementsByClassName(\"panel-info\")
                  for (tabInfo of tabInfos) {
                    tabInfo.setAttribute('aria-label', 'You are currently on a collapsed panel. To open and view the additional information inside, press enter. To close once open, press enter again.')
                  }
                }

              }

              // Remove the tabindex attribute from the list of fields on the Fields Page
              if (newNode.className === \"field-list\") {
                let fieldsListTextSection = document.getElementById(\"fields_page-capstat_fields_text\");
                let fieldList = fieldsListTextSection.getElementsByClassName(\"field-list\")[0];
                let ulFieldList = fieldList.getElementsByTagName(\"ul\")[0];
                ulFieldList.removeAttribute(\"tabindex\");
              }

              // Remove the tabindex attribute from the list of extensions on the Fields Page
              if (newNode.className === \"extension-list\") {
                let fieldsListTextSection = document.getElementById(\"fields_page-capstat_extension_text\");
                let fieldList = fieldsListTextSection.getElementsByClassName(\"extension-list\")[0];
                let ulFieldList = fieldList.getElementsByTagName(\"ul\")[0];
                ulFieldList.removeAttribute(\"tabindex\");
              }   
            }

            /*
              // Set the aria labels for all the select HTML elements with class shiny-bound-input in the endpoint modal popup 
              // If the added node contains the class 'container-fluid', that means the added node is the endpoint modal popup
              // This sets the aria label of all of the dropdowns in the endpoint modal popup to explain you can use the arrow keys to navigate the filter dropdown
            */
            if (mutation.addedNodes[0].classList && mutation.addedNodes[0].classList.contains(\"container-fluid\")) {
              let containerNode = mutation.addedNodes[0]
              let selectDropdowns = containerNode.querySelectorAll(\"select.shiny-bound-input\")
              for (selectDropdown of selectDropdowns) {
                selectDropdown.setAttribute('aria-label', 'Use the arrow keys to navigate the filter menu.')
              }
            } 
          }

          // Sets the aria label of the validation table to explain how to use and navigate the validation table to filter the validation failure table
          if (mutation.target.id === \"validations_page-validation_details_table\") {
            let reactTable = mutation.target.getElementsByClassName(\"ReactTable\")
            let rtTable = reactTable[0].getElementsByClassName(\"rt-table\")
            let rtTHead = rtTable[0].getElementsByClassName(\"rt-thead\")
            let rtTr = rtTHead[0].getElementsByClassName(\"rt-tr\")
            
            let rtTh = rtTr[0].getElementsByClassName(\"rt-align-left -cursor-pointer rt-th\")
            let rtSortHeader = rtTh[0].getElementsByClassName(\"rt-sort-header\")
            let rtThContent = rtSortHeader[0].getElementsByClassName(\"rt-th-content\")

            rtThContent[0].setAttribute(\"aria-label\", \"You are currently on a table whose entries serve as a filter for the validation failure table. To enter the table, press the tab key. Then, use the up and down arrow keys to move through the filter options. The filter option you are currently focused on will be automatically selected to filter the validation failure table. To exit the filter table, press the tab key again\")
          }

          // Sets the aria label of the FHIR operations input filter box on the Resource tab to explain how to use and navigate the filter box
          if (mutation.target.classList && mutation.target.classList.contains(\"selectize-dropdown-content\")) {
            let optionElems = mutation.target.getElementsByClassName(\"option\")
            optionElems[0].setAttribute(\"aria-label\", \"You are currently in a FHIR operations input filter box. Type to search for an operation, or use the arrow keys to navigate through the operations and select using enter. Remove selected operations by pressing the backspace key. Exit the filter input box by pressing the tab key\")
          }

          // Checks if the mutated node was the page title to catch when the Lantern tab changes
          if (mutation.target.id === \"page_title\") {
            
            let pageTitleText = mutation.target.textContent
            /*
              // Checks if the current page is the Organization, Resource, or Profile page
              // Alter navigation tabs on these pages to have the same structure, classes, and attributes as the navigation bars on the rest of the Lantern pages
            */
            if (pageTitleText === \"Organizations Page\" || pageTitleText === \"Resource Page\" || pageTitleText === \"Profile Page\") {
              
              let navBarTabs = document.getElementsByClassName(\"nav nav-tabs\");

              
              let tabbable = document.getElementsByClassName(\"tabbable\")
              let tabContents = tabbable[0].getElementsByClassName(\"tab-content\")
              

              for (let navTab of navBarTabs) {
                
                // Add the tablist role to the main navbar containing element
                navTab.setAttribute(\"role\", \"tablist\")
                
                let navTabID = navTab.getAttribute(\"data-tabsetid\")
                
                // Add the shiny-tab-input and shiny-bound-input classes to the main navbar containing element
                navTab.classList.add(\"shiny-tab-input\", \"shiny-bound-input\")

                let liElements = navTab.getElementsByTagName(\"li\")
                for (liElem of liElements) {
                  
                  // Add the presentation role to each of the navigation bar tabs
                  liElem.setAttribute(\"role\", \"presentation\")
                  
                  // Add the tab index observer to each of the navigation bar tab links to remove the tabindex attribute whenever it is set or altered
                  let aElements = liElem.getElementsByTagName(\"a\")                 
                  if (aElements.length > 0) {
                    tabIndexObserver.observe(aElements[0], {
                      attributes: true,
                      attributeFilter: [\"tabindex\"]
                    });
                      
                    /*
                      // Set the role of each tab element link to tab
                      // Set aria-selected to true
                      // Set the aria-control attribute equal to the id of the main navbar containing element
                      // Set aria-expanded to true
                      // Set the aria label of the link to explain you can click on the current tab 
                    */
                    for (let aElem of aElements) {
                      aElem.setAttribute(\"role\", \"tab\")
                      aElem.setAttribute(\"aria-selected\", \"true\")
                      aElem.setAttribute(\"aria-controls\", \"tab-\" + navTabID + \"-1\")
                      aElem.setAttribute(\"aria-expanded\", \"true\")
                      aElem.setAttribute(\"aria-label\", \"Press enter to select the \" + aElem.textContent +\" tab and show this tab's content below\")
                    }
                  }
                }
              }

              /*
                // Set the tabindex observer on each of the tab pages within the navigation tabs on the Organization, Resource, or Profile page
                // Which will remove the tab index attribute whenever it is set or altered 
              */
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

            if (pageTitleText === \"List of Endpoints\") {
              let downloads_tab_link = document.getElementById(\"downloads_page_link\")
              downloads_tab_link.addEventListener('click', function (e) {
                let downloads_tab = document.querySelector(\"a[href = '#shiny-tab-downloads_tab']\"); 
                downloads_tab.click();
              });

              downloads_tab_link.addEventListener('keyup', function (e) {
                if (event.keyCode === 13) {
                  let downloads_tab = document.querySelector(\"a[href = '#shiny-tab-downloads_tab']\"); 
                  downloads_tab.click();
                }
              });
            }
            
            /*
              // Set the aria labels for all the select HTML elements with class shiny-bound-input on the Lantern website
              // This sets the aria label of all of the dropdowns on Lantern to explain you can use the arrow keys to navigate the filter dropdown
            */
            let selectInputButtons = document.querySelectorAll(\"select.shiny-bound-input\")
            for (let selectInput of selectInputButtons) {
              selectInput.setAttribute('aria-label', 'Use the arrow keys to navigate the filter menu.')
            }
            
          }
          
          // Add an aria-label attribute to the FHIR Version dropdown that explains how to navigate and use the filter
          if (mutation.target.id === \"show_filters\") {
            let dropDownButtons = document.getElementsByClassName(\"dropdown-toggle\")
            for (let dropDownButton of dropDownButtons) {
              dropDownButton.setAttribute('aria-label', 'Dropdown filter menu button. Press the down arrow key to open the filter menu, use the tab or arrow keys to navigate through options, press enter to select a filter option, and use the escape key to close the filter menu.')
            }  
          }
        }
      })

      // Set the newNodesObserver on the Lantern document so that it watches for any added elements or any changes to the elements within the overall document
      newNodesObserver.observe(document.body, {
          childList: true, 
          subtree: true, 
          attributes: false, 
          characterData: false
      })
    "))
  )
)
