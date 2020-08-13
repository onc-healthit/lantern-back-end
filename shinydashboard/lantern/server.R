# Define server function
function(input, output, session) {

  callModule(
    dashboard,
    "dashboard_page",
    reactive(input$httpvendor))

  callModule(
    endpointsmodule,
    "endpoints_page",
    reactive(input$fhir_version),
    reactive(input$vendor))

  callModule(
    availabilitymodule,
    "availability_page")

  callModule(
    performancemodule,
    "performance_page",
    reactive(input$date))

  callModule(
    securitymodule,
    "security_page",
    reactive(input$fhir_version),
    reactive(input$vendor))

  callModule(
    smartresponsemodule,
    "smartresponse_page",
    reactive(input$fhir_version),
    reactive(input$vendor))

  callModule(
    capabilitymodule,
    "capability_page",
    reactive(input$fhir_version),
    reactive(input$vendor),
    reactive(input$resources))

  callModule(
    fieldsmodule,
    "fields_page",
    reactive(input$fhir_version),
    reactive(input$vendor))

  show_http_vendor_filter <- reactive(input$side_menu %in% c("dashboard_tab"))

  show_datefilter <- reactive(input$side_menu %in% c("performance_tab"))

   page_name_list <- list(
     "dashboard_tab" = "Current Endpoint Metrics",
     "endpoints_tab" = "List of Endpoints",
     "capability_tab" = "Capability Page",
     "fields_tab" = "Fields Page",
     "availability_tab" = "Endpoint Server Availability",
     "location_tab" = "Location Map Page",
     "about_tab" = "About Lantern",
     "security_tab" = "Security Authorization Types",
     "smartresponse_tab" = "SMART Core Capabilities Well Known Endpoint Response",
     "performance_tab" = "Response Time Performance"
  )

  show_filter <- reactive(
    input$side_menu %in% c("endpoints_tab", "capability_tab", "fields_tab", "security_tab", "smartresponse_tab")
  )

  show_http_vendor_filter <- reactive(input$side_menu %in% c("dashboard_tab"))

  show_date_filter <- reactive(input$side_menu %in% c("performance_tab"))
  
  show_resource_checkbox <- reactive(input$side_menu %in% c("capability_tab"))

  page_name <- reactive({
    page_name_list[[input$side_menu]]
  })

  output$page_title <- renderText(page_name())
  output$version <- renderText(version_title)

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

  output$show_http_vendor_filters <- renderUI({
    if (show_http_vendor_filter()) {
      fluidRow(
        column(width = 4,
          selectInput(
            inputId = "httpvendor",
            label = "Vendor:",
            choices = app$vendor_list,
            selected = ui_special_values$ALL_VENDORS,
          )
        )
      )
    }
  })

  output$show_date_filters <- renderUI({
    if (show_date_filter()) {
      fluidRow(
        column(width = 4,
          selectInput(
            inputId = "date",
            label = "Date range",
            choices = list("Past 7 days", "Past 14 days", "Past 30 days", "All time"),
            selected = "All time",
            size = 1,
            selectize = FALSE)
        )
      )
    }
  })

  output$show_resource_checkboxes <- renderUI({
    if (show_resource_checkbox()) {
      fluidPage(
        checkboxGroupInput("resources", "Choose Resources:", choices = get_resource_list(app_data$endpoint_resource_types), selected = ui_special_values$ALL_RESOURCES, inline = TRUE)
      )
    }
  })
}
