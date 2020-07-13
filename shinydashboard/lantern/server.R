# Define server function
function(input, output, session) {

  callModule(
    dashboard,
    "dashboard_page")

  callModule(
    endpointsmodule,
    "endpoints_page",
    reactive(input$fhir_version),
    reactive(input$vendor))

  callModule(
    availability,
    "availability_page")

  callModule(
    capabilitymodule,
    "capability_page",
    reactive(input$fhir_version),
    reactive(input$vendor))

   page_name_list <- list("dashboard_tab" = "Current Endpoint Metrics",
                          "endpoints_tab" = "List of Endpoints",
                          "capability_tab" = "Capability Page",
                          "availability_tab" = "Endpoint Server Availability",
                          "location_tab" = "Location Map Page",
                          "about_tab" = "About Lantern"
                        )
  show_filter <- reactive(input$side_menu %in% c("endpoints_tab", "capability_tab"))

  page_name <- reactive({
    page_name_list[[input$side_menu]]
  })

  output$page_title <- renderText(page_name())

  output$show_filters <- renderUI({
    if (show_filter()) {
      fluidRow(
        column(width = 4,
          selectInput(
            inputId = "fhir_version",
            label = "FHIR Version:",
            choices = app$fhir_version_list,
            selected = ui_special_values$ALL_FHIR_VERSIONS,
            size = 1,
            selectize = FALSE)
        ),
        column(width = 4,
          selectInput(
            inputId = "vendor",
            label = "Vendor:",
            choices = app$vendor_list,
            selected = ui_special_values$ALL_VENDORS,
            size = 1,
            selectize = FALSE)
        )
      )
    }
  })

}
