# Define server function
function(input, output, session) { #nolint

  # Trigger this observer every time the session changes, which is on first load of page, and switch tab to tab stored in url
  observeEvent(session, {
    query <- parseQueryString(session$clientData$url_search)
    if (!is.null(query[["tab"]]) && (toString(query[["tab"]]) %in% c("dashboard_tab", "endpoints_tab", "capability_tab", "fields_tab", "values_tab", "performance_tab", "security_tab", "smartresponse_tab", "location_tab", "about_tab"))) {
      current_tab <- toString(query[["tab"]])
      updateTabItems(session, "side_menu", selected = current_tab)
    } else {
      updateQueryString(paste0("?tab=", input$side_menu), mode = "push")
    }
  }, priority = 100)

  observeEvent(database_fetch, {
    if (database_fetch() == 1) {
      show_modal_spinner(
        spin = "double-bounce",
        color = "#112446",
        text = "Please Wait, Lantern is fetching the most up-to-date data")
      database_fetcher()
      database_fetch(0)
      remove_modal_spinner()
    }
  }, priority = 90)

  # Trigger this observer every time side_menu changes, and change the url to contain the new tab name
  observeEvent(input$side_menu, {
    updateQueryString(paste0("?tab=", input$side_menu), mode = "push")
  }, ignoreInit = TRUE)

  callModule(
        dashboard,
        "dashboard_page",
        reactive(input$httpvendor))

  observeEvent(database_fetch, {
    if (database_fetch() == 0) {
      callModule(
        endpointsmodule,
        "endpoints_page",
        reactive(input$fhir_version),
        reactive(input$vendor),
        reactive(input$availability))

      callModule(
        downloadsmodule,
        "downloads_page")

      callModule(
        locationmodule,
        "location_page",
        reactive(input$fhir_version),
        reactive(input$vendor))

      callModule(
        performancemodule,
        "performance_page",
        reactive(input$date))

      callModule(
        capabilitystatementsizemodule,
        "capabilitystatementsize_page",
        reactive(input$fhir_version),
        reactive(input$vendor))

      callModule(
        securitymodule,
        "security_page",
        reactive(input$fhir_version),
        reactive(input$vendor),
        reactive(input$auth_type_code))

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
        reactive(input$resources),
        reactive(input$operations))

      callModule(
        fieldsmodule,
        "fields_page",
        reactive(input$fhir_version),
        reactive(input$vendor))

      callModule(
        valuesmodule,
        "values_page",
        reactive(input$fhir_version),
        reactive(input$vendor),
        reactive(input$field))
    }
  })

  show_http_vendor_filter <- reactive(input$side_menu %in% c("dashboard_tab"))

  page_name_list <- list(
     "dashboard_tab" = "Current Endpoint Metrics",
     "endpoints_tab" = "List of Endpoints",
     "downloads_tab" = "Downloads Page",
     "capability_tab" = "Capability Page",
     "fields_tab" = "Fields Page",
     "values_tab" = "Values Page",
     "location_tab" = "Location Map",
     "about_tab" = "About Lantern",
     "security_tab" = "Security Authorization Types",
     "smartresponse_tab" = "SMART Core Capabilities Well Known Endpoint Response",
     "performance_tab" = "Response Time Performance",
     "capabilitystatementsize_tab" = "Capability Statement Size"
  )

  show_filter <- reactive(
    input$side_menu %in% c("endpoints_tab", "capability_tab", "fields_tab", "security_tab", "smartresponse_tab", "location_tab", "values_tab", "capabilitystatementsize_tab")
  )

  show_availability_filter <- reactive(
    input$side_menu %in% c("endpoints_tab")
  )

  show_date_filter <- reactive(input$side_menu %in% c("performance_tab"))

  show_resource_checkbox <- reactive(input$side_menu %in% c("capability_tab"))

  show_operation_checkbox <- reactive(input$side_menu %in% c("capability_tab"))

  show_value_filter <- reactive(input$side_menu %in% c("values_tab"))

  show_security_filter <- reactive(input$side_menu %in% c("security_tab"))

  page_name <- reactive({
    page_name_list[[input$side_menu]]
  })

  output$page_title <- renderText(page_name())
  output$version <- renderText(version_title)

  output$show_filters <- renderUI({
    if (show_filter()) {
      fhirDropdown <- selectInput(inputId = "fhir_version", label = "FHIR Version:", choices = isolate(app$fhir_version_list()), selected = ui_special_values$ALL_FHIR_VERSIONS, size = 1, selectize = FALSE)
      developerDropdown <- selectInput(inputId = "vendor", label = "Developer:", choices = app$vendor_list, selected = ui_special_values$ALL_DEVELOPERS, size = 1, selectize = FALSE)
      availabilityDropdown <- selectInput(inputId = "availability", label = "Availability Percentage:", choices = list("0-100", "0", "50-100", "75-100", "95-100", "99-100", "100"), selected = "0-100", size = 1, selectize = FALSE)
      if (show_availability_filter()) {
        fluidRow(
          column(width = 4, fhirDropdown),
          column(width = 4, developerDropdown),
          column(width = 4, availabilityDropdown)
        )
      } else {
        fluidRow(
          column(width = 4, fhirDropdown),
          column(width = 4, developerDropdown)
        )
      }
    }
  })

  output$show_http_vendor_filters <- renderUI({
    if (show_http_vendor_filter()) {
      fluidRow(
        column(width = 4,
          selectInput(
            inputId = "httpvendor",
            label = "Developer:",
            choices = app$vendor_list,
            selected = ui_special_values$ALL_DEVELOPERS
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

  output$show_value_filters <- renderUI({
    if (show_value_filter()) {
      fluidRow(
        column(width = 4,
          selectInput(
            inputId = "field",
            label = "Field",
            choices = list("url", "version", "name", "title", "date", "publisher", "description", "purpose", "copyright", "software_name", "software_version", "software_release_date", "implementation_description", "implementation_url", "implementation_custodian"),
            selected = "url",
            size = 1,
            selectize = FALSE)
        )
      )
    }
  })

  output$show_security_filter <- renderUI({
    if (show_security_filter()) {
      fluidRow(
        column(width = 4,
          selectInput(
            inputId = "auth_type_code",
            label = "Supported Authorization Type:",
            choices = isolate(app_data$security_code_list()),
            selected = "SMART-on-FHIR",
            size = 1,
            selectize = FALSE)
        )
      )
    }
  })

  checkbox_resources <- reactive({
    res <- isolate(app_data$endpoint_resource_types())
    req(input$fhir_version, input$vendor)
    if (input$fhir_version != ui_special_values$ALL_FHIR_VERSIONS) {
      res <- res %>% filter(fhir_version == input$fhir_version)
    }
    if (input$vendor != ui_special_values$ALL_DEVELOPERS) {
      res <- res %>% filter(vendor_name == input$vendor)
    }

    res <- res %>%
           distinct(type) %>%
           arrange(type) %>%
           split(.$type) %>%
           purrr::map(~ .$type)

    return(res)
  })

  output$show_resource_checkboxes <- renderUI({
    if (show_resource_checkbox()) {
      fluidPage(
        fluidRow(
          actionButton("selectall", "Select All Resources"),
          actionButton("removeall", "Clear All Resources"),
          selectizeInput("resources", "Choose or type in any resource from the list below:", choices = checkbox_resources(), selected = checkbox_resources(), multiple = TRUE, options = list("plugins" = list("remove_button"), "create" = TRUE, "persist" = FALSE), width = "100%"),
          p("Note: The resource list will only contain resources that are supported by endpoints that pass the selected filtering criteria.", style = "font-size:13px; margin-top:-15px")
        )
      )
    }
  })

  current_selection <- reactiveVal(NULL)

  observeEvent(input$resources, {
    current_selection(input$resources)
  })

  observe({
    req(input$side_menu)
    if (show_resource_checkbox()) {
      updateSelectInput(session, "resources", label = "Choose or type in any resource from the list below:", choices = checkbox_resources(), selected = checkbox_resources())
    }
  })

  observe({
    req(input$fhir_version, input$vendor)
    updateSelectInput(session, "resources", label = "Choose or type in any resource from the list below:", choices = checkbox_resources(), selected = current_selection())
  })

  observeEvent(input$selectall, {
    if (input$selectall == 0) {
      return(NULL)
    }
    else{
      updateSelectizeInput(session, "resources", label = "Choose or type in any resource from the list below:", choices = checkbox_resources(), selected = checkbox_resources(), options = list("plugins" = list("remove_button"), "create" = TRUE, "persist" = FALSE))
    }
  })

  observeEvent(input$removeall, {
    if (input$removeall == 0) {
      return(NULL)
    }
    else{
      updateSelectizeInput(session, "resources", label = "Choose or type in any resource from the list below:", choices = checkbox_resources(), options = list("plugins" = list("remove_button"), "create" = TRUE, "persist" = FALSE))
    }
  })

  #                     #
  # Operations Checkbox #
  #                     #

  # Operations checkbox display
  output$show_operation_checkboxes <- renderUI({
    if (show_operation_checkbox()) {
      fluidPage(
        fluidRow(
          actionButton("removeallops", "Clear All Operations"),
          selectizeInput("operations", "Choose or type in any resource from the list below:",
          choices = c("read", "vread", "update", "patch", "delete", "history-instance", "history-type", "create", "search-type", "not specified"),
          selected = c(), multiple = TRUE, options = list("plugins" = list("remove_button"), "create" = TRUE, "persist" = FALSE), width = "100%"),
          p("Note: Selecting multiple operations will only display the resources that implement all selected operations.", style = "font-size:13px; margin-top:-15px")
        )
      )
    }
  })

  current_op_selection <- reactiveVal(NULL)

  # Updates what the user has currently selected
  observeEvent(input$operations, {
    current_op_selection(input$operations)
  })

  # Resets the display if the user is navigating to this page
  observe({
    req(input$side_menu)
    if (show_operation_checkbox()) {
      updateSelectInput(session, "operations",
            label = "Choose or type in any operation from the list below:",
            choices = c("read", "vread", "update", "patch", "delete", "history-instance", "history-type", "create", "search-type", "not specified"),
            selected = c())
    }
  })

  # Resets the display if the user clicks the "Remove All Operations" button
  observeEvent(input$removeallops, {
    if (input$removeallops == 0) {
      return(NULL)
    }
    else{
      updateSelectizeInput(session, "operations",
              label = "Choose or type in any operation from the list below:",
              choices = c("read", "vread", "update", "patch", "delete", "history-instance", "history-type", "create", "search-type", "not specified"),
              options = list("plugins" = list("remove_button"), "create" = TRUE, "persist" = FALSE))
    }
  })
}
