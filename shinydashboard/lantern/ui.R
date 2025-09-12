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
      menuItem("CapabilityStatement / Conformance Field Values", icon = tags$i(class = "fa fa-table", "aria-hidden" = "true", role = "presentation", "aria-label" = "table icon"), tabName = "values_tab"),
      menuItem("CapabilityStatement / Conformance Profiles", icon = tags$i(class = "fa fa-list-alt", "aria-hidden" = "true", role = "presentation", "aria-label" = "list-alt icon"), tabName = "profile_tab"),
      menuItem("CapabilityStatement / Conformance Size", icon = tags$i(class = "fa fa-hdd-o", "aria-hidden" = "true", role = "presentation", "aria-label" = "hdd-o icon"), tabName = "capabilitystatementsize_tab"),
      menuItem("Validations", icon = tags$i(class = "fa fa-clipboard-check", "aria-hidden" = "true", role = "presentation", "aria-label" = "clipboard-check icon"), tabName = "validations_tab"),
      menuItem("Security", icon = tags$i(class = "fa fa-id-card-o", "aria-hidden" = "true", role = "presentation", "aria-label" = "id-card-o icon"), tabName = "security_tab"),
      menuItem("SMART Response", icon = tags$i(class = "fa fa-list", "aria-hidden" = "true", role = "presentation", "aria-label" = "list icon"), tabName = "smartresponse_tab"),
      menuItem("Contact Information", tabName = "contacts_tab", icon = tags$i(class = "fa fa-list-alt", "aria-hidden" = "true", role = "presentation", "aria-label" = "list-alt icon")),
      menuItem("Downloads", tabName = "downloads_tab", icon = tags$i(class = "fa fa-download", "aria-hidden" = "true", role = "presentation", "aria-label" = "download icon")),
      menuItem("About Lantern", tabName = "about_tab", icon = tags$i(class = "fa fa-info-circle", "aria-hidden" = "true", role = "presentation", "aria-label" = "info-circle icon")),
      menuItem("Release Notes", tabName = "release_notes", icon = tags$i(class = "fa fa-info-circle", "aria-hidden" = "true", role = "presentation", "aria-label" = "info-circle icon")),
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
      
      # Critical CSS inline for fast initial render
      tags$style(HTML("
        /* Critical styles needed for initial render */
        .content-wrapper, .right-side { background-color: #F6F7F8; }
        .skin-blue .main-header .navbar { background-color: #1B5A7F; }
        .skin-blue .main-header .logo { background-color: #1B5A7F; }
        .show-on-focus { position: absolute; top: -10em; background: #fff; color: #112e51; display: block; font-weight: 600; }
        .show-on-focus:focus { top: 5px; position: absolute; background: #fff; color: #112e51; display: block; font-weight: 600; font-size: 20px; }
        .small-box { color: black !important; position: relative; display: block; margin-bottom: 20px; box-shadow: 0 1px 1px rgba(0,0,0,0.1); }
        .small-box > .inner { padding: 10px; }
        .small-box h3 { font-size: 38px; font-weight: bold; margin: 0 0 10px 0; white-space: nowrap; padding: 0; }
        .small-box p { font-size: 20px; }
        
        /* Font display swap */
        @font-face {
          font-family: 'FontAwesome';
          font-display: swap;
        }
        
        /* Utility styles */
        .badge { color: black !important; }
        .lantern-url { color: #0044FF !important; text-decoration: underline; }
        
        /* Fix for duplicate h2 issue */
        #prerender-duplicate-fix {
          display: none !important;
          visibility: hidden !important;
          height: 0 !important;
          opacity: 0 !important;
          pointer-events: none !important;
          position: absolute !important;
          left: -9999px !important;
        }
      ")),
      
      # Non-blocking CSS loading
      tags$link(rel = "preload", href = "css/lantern-styles.min.css", as = "style"),
      tags$link(rel = "stylesheet", href = "css/lantern-styles.min.css", media = "print", onload = "this.media='all'"),
      tags$noscript(tags$link(rel = "stylesheet", href = "css/lantern-styles.min.css")),
      
      # Script to fix duplicate h2 issue
      tags$script(HTML("
        document.addEventListener('DOMContentLoaded', function() {
          // Give time for the real dashboard to load
          setTimeout(function() {
            var h2Elements = document.querySelectorAll('h2');
            
            // If we have multiple h2s with the same text, hide all but the last one
            if (h2Elements.length > 0) {
              var h2Text = {};
              
              // Find duplicate h2 elements
              for (var i = 0; i < h2Elements.length; i++) {
                var text = h2Elements[i].textContent.trim();
                if (text in h2Text) {
                  h2Text[text].push(h2Elements[i]);
                } else {
                  h2Text[text] = [h2Elements[i]];
                }
              }
              
              // For each set of duplicate h2s, hide all but the last one
              for (var text in h2Text) {
                if (h2Text[text].length > 1) {
                  // Keep only the last instance visible (which is likely the 'real' one)
                  for (var i = 0; i < h2Text[text].length - 1; i++) {
                    h2Text[text][i].id = 'prerender-duplicate-fix';
                  }
                }
              }
            }
          }, 300); // Wait 300ms for dashboard to load
        });
      ")),
      
      # Deferred script loading
      tags$script(HTML("
        // Load non-critical scripts after page loads
        window.addEventListener('load', function() {
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
      # Dashboard tab - removed the pre-rendered h2
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
        ),
      tabItem("release_notes",
              p(HTML('
                Lantern displayed Organization\'s HTI-1 data as a modal, we now replaced modal with inline columns. You\'ll now see these fields as separate columns:<br/>
                <ul>
                  <li>Organization Identifier Type</li>
                  <li>Organization Identifier</li>
                  <li>Organization Name</li>
                  <li>Organization Address</li>
                </ul>

                Lantern now shows only organizations that are marked as "active" in their respective FHIR bundles.<br/><br/>

                <b>Download (page-level):</b> Added a Download Organizations action that returns data based on the filters applied on the page.<br/>
                <ul>
                  <li>No filters → downloads all rows.</li>
                  <li>With filters → downloads filtered rows.</li>
                </ul>

                <b>Performance:</b> General speed improvements on the Organizations page<br/><br/>

                <b>New Organizations Download API:</b><br/>
                <a href="https://lantern.healthit.gov/api/organizations/v1" target="_blank">
                  https://lantern.healthit.gov/api/organizations/v1
                </a><br/><br/>

                <b>Access:</b> Call directly from a browser or tools like Postman.<br/>
                <b>Filtering:</b> Use URL-encoded query parameters (you can combine them):<br/>
                <ul>
                  <li><code>developer</code> — filter by certified API developer name</li>
                  <li><code>fhir_version</code> — comma-separated FHIR versions (e.g., 4.0.1)</li>
                  <li><code>identifier</code> — exact organization identifier (e.g., NPI, Other)</li>
                  <li><code>hti1</code> — use <code>hti1=present</code> to return only orgs with HTI-1 data</li>
                </ul>

                <b>Examples:</b><br/>
                By Developer:<br/>
                .../api/organizations/v1?developer=Cerner%20Corporation<br/><br/>
                By NPI:<br/>
                .../api/organizations/v1?identifier=1922195171<br/><br/>
                By FHIR Version:<br/>
                .../api/organizations/v1?fhir_version=4.0.1<br/><br/>
                Only with HTI-1 Data:<br/>
                .../api/organizations/v1?hti1=present<br/><br/>

                <b>Organization Data visibility & ingestion notes</b><br/>
                Lantern now ingests all organizations found in FHIR bundles, even when HTI-1 fields are missing. 
                This can help developers spot gaps in their data and fix them.<br/><br/>

                <b>Bug Fixes:</b>
                <ul>
                  <li>1UP was not showing as a developer though they have data. Fixed to display 1UP.</li>
                  <li>Lantern organizations were grouped by Organization name on the UI, potentially grouping unrelated organizations. This issue is resolved and we no longer group by name.</li>
                  <li>Organization page changes to display HTI-1 data as separate columns slowed down the page. Changes were made to improve the performance of this page.</li>
                  <li>Organization page results were showing same data on pages 1 and 3 due to skipping organizations with bad names like a hyphen for a name. Resolved this issue.</li>
                  <li>If an organization information is changed or removed, changes to backend data processing to keep the database clean. This is only a backend change, no impact to the data on the front-end.</li>
                </ul>
              '))
      )
    ),
    uiOutput("htmlFooter"),
    
    # Load accessibility script with defer and async
    tags$script(src = "js/accessibility.js", defer = TRUE, async = TRUE)
  )
)