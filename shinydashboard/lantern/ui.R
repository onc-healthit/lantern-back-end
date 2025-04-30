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
        tags$picture(
          tags$source(srcset = "images/GitHub-Mark-Light-32px.webp", type = "image/webp"),
          tags$img(src = "images/GitHub-Mark-Light-32px.png", width = "19", height = "19", alt = "Github logo")
        ),
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
    tags$head(
      # Meta tags for better SEO and rendering
      tags$meta(charset = "UTF-8"),
      tags$meta(name = "viewport", content = "width=device-width, initial-scale=1"),
      tags$meta(name = "description", content = "Lantern Dashboard - FHIR API Monitoring Tool. Track health IT interoperability capabilities across endpoints."),
      
      # Preconnect to external domains
      tags$link(rel = "preconnect", href = "https://www.googletagmanager.com"),
      tags$link(rel = "preconnect", href = "https://stats.g.doubleclick.net"),
      
      # Favicon with dimensions
      tags$link(rel = "shortcut icon", href = "images/favicon.webp", sizes = "32x32"),
      
      # Critical CSS specifically for the dashboard h2 causing LCP issues
      tags$style(HTML("
        /* Dashboard h2 optimization styles to reduce LCP time */
        h2, #dashboard_page h2 {
          font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif !important;
          font-weight: 500 !important;
          line-height: 1.1 !important;
          color: inherit !important;
          margin-top: 20px !important;
          margin-bottom: 10px !important;
          font-size: 30px !important;
          display: block !important;
          contain: content !important;
        }
        
        /* Optimize dashboard elements that appear early */
        .content-wrapper, .right-side { background-color: #F6F7F8; }
        .skin-blue .main-header .navbar { background-color: #1B5A7F; }
        .skin-blue .main-header .logo { background-color: #1B5A7F; }
        
        /* Dashboard-specific critical styles */
        .small-box { color: black !important; position: relative; display: block; margin-bottom: 20px; box-shadow: 0 1px 1px rgba(0,0,0,0.1); }
        .small-box > .inner { padding: 10px; }
        .small-box h3 { font-size: 38px; font-weight: bold; margin: 0 0 10px 0; white-space: nowrap; padding: 0; }
        .small-box p { font-size: 20px; }
        
        /* Font optimization */
        @font-face {
          font-family: 'FontAwesome';
          font-display: swap;
        }
        
        /* Utility styles */
        .badge { color: black !important; }
        .lantern-url { color: #0044FF !important; text-decoration: underline; }
        .show-on-focus { position: absolute; top: -10em; background: #fff; color: #112e51; display: block; font-weight: 600; }
        .show-on-focus:focus { top: 5px; position: absolute; background: #fff; color: #112e51; display: block; font-weight: 600; font-size: 20px; }
      ")),
      
      # Non-blocking CSS loading
      tags$link(rel = "preload", href = "css/lantern-styles.min.css", as = "style"),
      tags$link(rel = "stylesheet", href = "css/lantern-styles.min.css", media = "print", onload = "this.media='all'"),
      tags$noscript(tags$link(rel = "stylesheet", href = "css/lantern-styles.min.css")),
      
      # LCP optimization script
      tags$script(HTML("
        // Immediately optimize h2 elements for LCP
        document.addEventListener('DOMContentLoaded', function() {
          // Try to find all h2 elements
          var h2Elements = document.querySelectorAll('h2');
          
          // Apply optimizations to all h2 elements
          h2Elements.forEach(function(h2) {
            // Make sure these elements are visible immediately
            h2.style.visibility = 'visible';
            h2.style.display = 'block';
            
            // Apply advanced rendering optimizations for modern browsers
            if ('contentVisibility' in h2.style) {
              h2.style.contentVisibility = 'auto';
              h2.style.contain = 'content';
              
              // Measure and set intrinsic size to avoid layout shifts
              var rect = h2.getBoundingClientRect();
              h2.style.containIntrinsicSize = 'auto ' + rect.height + 'px';
            }
          });
          
          // Monitor LCP to confirm which element is causing it
          if ('PerformanceObserver' in window) {
            try {
              var lcpObserver = new PerformanceObserver(function(list) {
                var entries = list.getEntries();
                var lastEntry = entries[entries.length - 1];
                
                // If we find the LCP element, prioritize it even more
                if (lastEntry && lastEntry.element) {
                  lastEntry.element.style.visibility = 'visible';
                  lastEntry.element.style.display = 'block';
                  
                  if ('contentVisibility' in lastEntry.element.style) {
                    lastEntry.element.style.contentVisibility = 'auto';
                    lastEntry.element.style.contain = 'content';
                  }
                }
              });
              
              lcpObserver.observe({type: 'largest-contentful-paint', buffered: true});
            } catch (e) {
              // Silent fail - just a performance optimization
            }
          }
        });
      ")),
      
      # Deferred script loading
      tags$script(HTML("
        // Load non-critical scripts after page loads
        window.addEventListener('load', function() {
          // Delay non-essential scripts
          setTimeout(function() {
            // Load GTM asynchronously
            var script = document.createElement('script');
            script.src = 'https://www.googletagmanager.com/gtm.js?id=GTM-KC3FP96';
            script.async = true;
            document.head.appendChild(script);
          }, 1000);
        });
      "))
    ),
    
    # Google Tag Manager noscript fallback
    tags$noscript(tags$iframe(src = "https://www.googletagmanager.com/ns.html?id=GTM-KC3FP96", height = "0", width = "0", style = "display:none;visibility:hidden")),
    
    development_banner(devbanner),
    uiOutput("resource_tab_popup"),
    h1(textOutput("page_title")),
    uiOutput("show_filters"),
    uiOutput("show_value_filters"),
    uiOutput("show_resource_operation_checkboxes"),
    uiOutput("show_resource_profiles_dropdown"),
    uiOutput("organizations_filter"),
    tabItems(
      # Dashboard tab with optimization for LCP
      tabItem("dashboard_tab",
              # Pre-render the h2 that's causing LCP issues
              tags$div(
                id = "dashboard-prerender",
                # This is a static duplicate of the h2 from dashboard_UI
                tags$h2("Current endpoint responses:", id = "prerendered-h2", style = "visibility: visible !important; display: block !important;"),
                # Then include the real dashboard UI
                dashboard_UI("dashboard_page")
              )
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
              tags$picture(
                tags$source(srcset = "images/lantern-logo@1x.webp", type = "image/webp"),
                tags$img(src = "images/lantern-logo@1x.png", width = "300", height = "100", alt = "Lantern Logo", loading = "lazy")
              ),
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
    
    # Load accessibility script with defer and async
    tags$script(src = "js/accessibility.min.js", defer = TRUE, async = TRUE),
    
    # Script to hide duplicate h2 once real content loads
    tags$script(HTML("
      // Once the dashboard is fully loaded, remove the pre-rendered h2
      document.addEventListener('DOMContentLoaded', function() {
        setTimeout(function() {
          var prerenderedH2 = document.getElementById('prerendered-h2');
          if (prerenderedH2) {
            // First check if the real h2 from dashboard_UI is loaded
            var dashboardH2s = document.querySelectorAll('#dashboard_page h2');
            if (dashboardH2s.length > 0) {
              // If the real h2 exists, hide the pre-rendered one
              prerenderedH2.style.display = 'none';
            }
          }
        }, 500); // Check after half a second
      });
    "))
  )
)