library(shinyWidgets)
library(reactable)
library(shinyBS)
library(listviewer)
library(leaflet)
library(dygraphs)

# Define server function
function(input, output, session) { #nolint

selected_fhir_endpoint_profiles <- reactive({
    res <- get_supported_profiles(db_connection)
    req(input$fhir_version, input$vendor)

    res <- res %>% filter(fhir_version %in% input$fhir_version)

    if (input$vendor != ui_special_values$ALL_DEVELOPERS) {
      res <- res %>% filter(vendor_name == input$vendor)
    }

     if (length(input$profile_resource) > 0) {
        if (input$profile_resource != ui_special_values$ALL_RESOURCES) {
          res <- res %>% filter(resource == input$profile_resource)
        }
    }

    if (length(input$profiles) > 0) {
        if (input$profiles != ui_special_values$ALL_PROFILES) {
        res <- res %>% filter(profileurl == input$profiles)
        }
    }

    res <- res %>%
    distinct(url, profileurl, profilename, resource, fhir_version, vendor_name) %>%
    select(url, profileurl, profilename, resource, fhir_version, vendor_name) %>%
    group_by(url) %>%
    mutate(url = paste0("<a class=\"lantern-url\" tabindex=\"0\" aria-label=\"Press enter to open pop up modal containing additional information for this endpoint.\" onkeydown = \"javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)\" onclick=\"Shiny.setInputValue(\'endpoint_popup\',&quot;", url, "&&", "None", "&quot,{priority: \'event\'});\">", url, "</a>")) %>%
    mutate_at(vars(-group_cols()), as.character)

    return(res)
  })

  # Trigger this observer every time the session changes, which is on first load of page, and switch tab to tab stored in url
  observeEvent(session, {
    message(sprintf("I am in observe session  *********************************** %s", database_fetch()))
    query <- parseQueryString(session$clientData$url_search)
    if (!is.null(query[["tab"]]) && (toString(query[["tab"]]) %in% c("dashboard_tab", "endpoints_tab", "resource_tab", "organizations_tab", "implementation_tab", "fields_tab", "profile_tab", "values_tab", "capabilitystatementsize_tab", "validations_tab", "security_tab", "smartresponse_tab", "about_tab", "contacts_tab"))) {
      current_tab <- toString(query[["tab"]])
      updateTabItems(session, "side_menu", selected = current_tab)
    } else {
      updateQueryString(paste0("?tab=", input$side_menu), mode = "push")
    }
  }, priority = 100)

  observeEvent(database_fetch, {
    message(sprintf("I am in observe event *********************************** %s", database_fetch()))
    if (database_fetch() == 1) {
      message("I am inside observe event ***********************************")
      show_modal_spinner(
        spin = "double-bounce",
        color = "#112446",
        text = "Please Wait, Lantern is fetching the most up-to-date data")
      app_fetcher()
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
        reactive(input$availability),
        reactive(input$is_chpl))

      callModule(
        downloadsmodule,
        "downloads_page")

      callModule(
        organizationsmodule,
        "organizations_page",
        reactive(input$fhir_version),
        reactive(input$vendor),
        reactive(input$match_confidence))

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
        resourcemodule,
        "resource_page",
        reactive(input$fhir_version),
        reactive(input$vendor),
        reactive(input$resources),
        reactive(input$operations))

      callModule(
        implementationmodule,
        "implementation_page",
        reactive(input$fhir_version),
        reactive(input$vendor))

      callModule(
        fieldsmodule,
        "fields_page",
        reactive(input$fhir_version),
        reactive(input$vendor))

      callModule(
        profilemodule,
        "profile_page",
        reactive(input$fhir_version),
        reactive(input$vendor),
        reactive(input$profile_resource),
        reactive(input$profiles))

      callModule(
        valuesmodule,
        "values_page",
        reactive(input$fhir_version),
        reactive(input$vendor),
        reactive(input$field))

      callModule(
        contactsmodule,
        "contacts_page",
        reactive(input$fhir_version),
        reactive(input$vendor),
        reactive(input$has_contact)
      )

      callModule(
        validationsmodule,
        "validations_page",
        reactive(input$fhir_version),
        reactive(input$vendor),
        reactive(input$validation_group))
    }
  })

  show_http_vendor_filter <- reactive(input$side_menu %in% c("dashboard_tab"))

  page_name_list <- list(
     "dashboard_tab" = "Current Endpoint Metrics",
     "endpoints_tab" = "List of Endpoints",
     "downloads_tab" = "Downloads Page",
     "organizations_tab" = "Organizations Page",
     "resource_tab" = "Resource Page",
     "implementation_tab" = "Implementation Page",
     "fields_tab" = "Fields Page",
     "profile_tab" = "Profile Page",
     "values_tab" = "Values Page",
     "contacts_tab" = "Contact Information Page",
     "about_tab" = "About Lantern",
     "security_tab" = "Security Authorization Types",
     "smartresponse_tab" = "SMART Core Capabilities Well Known Endpoint Response",
     "capabilitystatementsize_tab" = "CapabilityStatement / Conformance Size",
     "validations_tab" = "Validations Page"
  )

  output$resource_tab_popup <- renderUI({
    if (show_resource_tab_popup()) {
      div(class = "pull-right", actionButton("resource_popup", "How to use this page", icon = tags$i(class = "fa fa-question-circle", "aria-hidden" = "true", role = "presentation", "aria-label" = "question icon")))
    }
  })

  observeEvent(input$resource_popup, {
    showModal(modalDialog(
      title = "How to use this page...",
      p("By default, the list of resources below contains the supported resources across all endpoints and FHIR versions. Clicking a resource in the left box selects it and moves it to the right box. Remove a resource from the list by clicking the resource in the right box.", style = "font-size:16px; margin-left:5px;"),
      p("You may also change the FHIR Version or Developer filtering criteria to filter the applicable supported resources from the default list.
        Any resources at that point will be removed from the list of resources if no endpoints that pass the selected filtering criteria support the given resource.
        If you make other changes to the FHIR Version or Developer filtering criteria, resources that are filtered out of the list will re-appear on the left side of the list, regardless if they were selected previously.", style = "font-size:16px; margin-left:5px;"),
      p("You will have to re-select these resources, either by clicking the resource on the left box, or clicking the 'Select All Resources' button.", style = "font-size:16px; margin-left:5px;"),
      p("Note: This is the list of FHIR resource types reported by the CapabilityStatement / Conformance Resources from the endpoints. This reflects the most recent successful response only. Endpoints which are down, unreachable during the last query or have not returned a valid CapabilityStatement / Conformance Resource, are not included in this list.", style = "font-size:13px; margin-left:5px;")
  ))})


  show_filter <- reactive(
    input$side_menu %in% c("endpoints_tab", "organizations_tab", "resource_tab", "implementation_tab", "fields_tab", "security_tab", "smartresponse_tab", "values_tab", "capabilitystatementsize_tab", "validations_tab", "profile_tab", "contacts_tab")
  )

  fhir_version_no_capstat <- reactive(
    input$side_menu %in% c("endpoints_tab", "smartresponse_tab", "validations_tab")
  )

  show_availability_filter <- reactive(
    input$side_menu %in% c("endpoints_tab")
  )

  show_validations_filter <- reactive(
    input$side_menu %in% c("validations_tab")
  )

  show_has_contact_filter <- reactive(input$side_menu %in% c("contacts_tab"))

  show_resource_checkbox <- reactive(input$side_menu %in% c("resource_tab"))

  show_profiles_filters <- reactive(input$side_menu %in% c("profile_tab"))

  show_operation_checkbox <- reactive(input$side_menu %in% c("resource_tab"))

  show_resource_tab_popup <- reactive(input$side_menu %in% c("resource_tab"))

  show_value_filter <- reactive(input$side_menu %in% c("values_tab"))

  show_security_filter <- reactive(input$side_menu %in% c("security_tab"))

  show_confidence_filter <- reactive(FALSE)

  page_name <- reactive({
    page_name_list[[input$side_menu]]
  })

  output$htmlFooter <- renderUI({
    if (input$side_menu %in% c("about_tab")) {
      tags$footer(class = "footer",
        includeHTML("aboutInfo.html")
      )
    } else {
      tags$footer(class = "footer",
        includeHTML("disclaimer.html")
      )
    }
  })

  output$page_title <- renderText(page_name())
  output$version <- renderText(version_title)

  observeEvent(input$fhirversion_selectall, {
    if (input$fhirversion_selectall == 0) {
      return(NULL)
    } else {
      updatePickerInput(session, inputId = "fhir_version", label = "FHIR Version:", choices = isolate(app$fhir_version_list_no_capstat()), selected = isolate(app$distinct_fhir_version_list_no_capstat()))
    }
  })

  observeEvent(input$fhirversion_removeall, {
    if (input$fhirversion_removeall == 0) {
      return(NULL)
    } else {
      updatePickerInput(session, inputId = "fhir_version", label = "FHIR Version:", choices = isolate(app$fhir_version_list_no_capstat()))
    }
  })

  output$show_filters <- renderUI({
    if (show_filter()) {
      if (fhir_version_no_capstat()) {
        fhirDropdown <- pickerInput(inputId = "fhir_version", label = "FHIR Version:", multiple = TRUE, choices = isolate(app$fhir_version_list_no_capstat()), selected = isolate(app$distinct_fhir_version_list_no_capstat()), options = list(`multiple-separator` = " | ", size = 5))
        fhirDropdown_noLabel <- pickerInput(inputId = "fhir_version", multiple = TRUE, choices = isolate(app$fhir_version_list_no_capstat()), selected = isolate(app$distinct_fhir_version_list_no_capstat()), options = list(`multiple-separator` = " | ", size = 5))
      } else {
        fhirDropdown <- pickerInput(inputId = "fhir_version", label = "FHIR Version:", multiple = TRUE, choices = isolate(app$fhir_version_list()), selected = isolate(app$distinct_fhir_version_list()), options = list(`multiple-separator` = " | ", size = 5))
        fhirDropdown_noLabel <- pickerInput(inputId = "fhir_version", multiple = TRUE, choices = isolate(app$fhir_version_list_no_capstat()), selected = isolate(app$distinct_fhir_version_list_no_capstat()), options = list(`multiple-separator` = " | ", size = 5))
      }
      developerDropdown <- selectInput(inputId = "vendor", label = "Developer:", choices = app$vendor_list(), selected = ui_special_values$ALL_DEVELOPERS, size = 1, selectize = FALSE)
      availabilityDropdown <- selectInput(inputId = "availability", label = "Availability Percentage:", choices = list("0-100", "0", "50-100", "75-100", "95-100", "99-100", "100"), selected = "0-100", size = 1, selectize = FALSE)
      validationsDropdown <- selectInput(inputId = "validation_group", label = "Validation Group", choices = c("All Groups", validation_group_names), selected = "All Groups", size = 1, selectize = FALSE)
      confidenceDropdown <- selectInput(inputId = "match_confidence", label = "Match Confidence:", choices = c("97-100", "98-100", "99-100", "100"), selected = "97-100", size = 1, selectize = FALSE)
      contactDropdown <- selectInput(inputId = "has_contact", label = "Has Contact Data:", choices = c("True", "False", "Any"), selected = "Any", size = 1, selectize = FALSE)
      chplDropdown <- selectInput(inputId = "is_chpl", label = "From CHPL:", choices = c("True", "False", "All"), selected = "All", size = 1, selectize = FALSE)
      if (show_availability_filter()) {
        fluidRow(
          column(width = 3,
          tags$div(
            p("FHIR Version: ", style = "font-weight: 700; font-size: 14px;"),
            actionButton("fhirversion_selectall", "Select All FHIR Versions", width = "145px", style = "font-size: 11px; margin-bottom: 3px; margin-left: auto; background-color: white;"),
            actionButton("fhirversion_removeall", "Remove All FHIR Versions", width = "145px", style = "font-size: 11px; margin-bottom: 3px; margin-left: auto; background-color: white;")
          ),
          fhirDropdown_noLabel),
          column(width = 3, developerDropdown),
          column(width = 3, availabilityDropdown),
          column(width = 3, chplDropdown)
        )
      } else if (show_validations_filter()) {
        fluidRow(
          column(width = 4,
          tags$div(
            p("FHIR Version: ", style = "font-weight: 700; font-size: 14px;"),
            actionButton("fhirversion_selectall", "Select All FHIR Versions", width = "145px", style = "font-size: 11px; margin-bottom: 3px; margin-left: auto; background-color: white;"),
            actionButton("fhirversion_removeall", "Remove All FHIR Versions", width = "145px", style = "font-size: 11px; margin-bottom: 3px; margin-left: auto; background-color: white;")
          ),
          fhirDropdown_noLabel),
          column(width = 4, developerDropdown),
          column(width = 4, validationsDropdown)
        )
      } else if (show_confidence_filter()) {
        fluidRow(
          column(width = 4,
          tags$div(
            p("FHIR Version: ", style = "font-weight: 700; font-size: 14px;"),
            actionButton("fhirversion_selectall", "Select All FHIR Versions", width = "145px", style = "font-size: 11px; margin-bottom: 3px; margin-left: auto; background-color: white;"),
            actionButton("fhirversion_removeall", "Remove All FHIR Versions", width = "145px", style = "font-size: 11px; margin-bottom: 3px; margin-left: auto; background-color: white;")
          ),
          fhirDropdown_noLabel),
          column(width = 4, developerDropdown),
          column(width = 4, confidenceDropdown)
        )
      } else if (show_has_contact_filter()) {
        fluidRow(
          column(width = 4,
          tags$div(
            p("FHIR Version: ", style = "font-weight: 700; font-size: 14px;"),
            actionButton("fhirversion_selectall", "Select All FHIR Versions", width = "145px", style = "font-size: 11px; margin-bottom: 3px; margin-left: auto; background-color: white;"),
            actionButton("fhirversion_removeall", "Remove All FHIR Versions", width = "145px", style = "font-size: 11px; margin-bottom: 3px; margin-left: auto; background-color: white;")
          ),
          fhirDropdown_noLabel),
          column(width = 4, developerDropdown),
          column(width = 4, contactDropdown)
        )
      } else {
        fluidRow(
          column(width = 4,
          tags$div(
            p("FHIR Version: ", style = "font-weight: 700; font-size: 14px;"),
            actionButton("fhirversion_selectall", "Select All FHIR Versions", width = "145px", style = "font-size: 11px; margin-bottom: 3px; margin-left: auto; background-color: white;"),
            actionButton("fhirversion_removeall", "Remove All FHIR Versions", width = "145px", style = "font-size: 11px; margin-bottom: 3px; margin-left: auto; background-color: white;")
          ),
          fhirDropdown_noLabel),
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
            choices = app$vendor_list(),
            selected = ui_special_values$ALL_DEVELOPERS,
            selectize = FALSE
          )
        )
      )
    }
  })

  output$show_has_contact_filters <- renderUI({
    if (show_has_contact_filter()) {
      fluidRow(
        column(width = 4,
          selectInput(
            inputId = "has_contact",
            label = "Has Contact Data",
            choices = list("True", "False", "Any"),
            selected = "Any"
          )
        )
      )
    }
  })

  output$show_date_filters <- renderUI({
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
  })

    output$show_http_date_filters <- renderUI({
    fluidRow(
      column(width = 4,
        selectInput(
          inputId = "http_date",
          label = "Date range",
          choices = list("Past 7 days", "Past 14 days", "Past 30 days", "All time"),
          selected = "All time",
          size = 1,
          selectize = FALSE)
      )
    )
  })

  output$show_value_filters <- renderUI({
    if (show_value_filter()) {
      fluidRow(
        column(width = 4,
          selectInput(
            inputId = "field",
            label = "Field",
            choices = list("url", "fhirVersion", "name", "title", "date", "publisher", "description", "purpose", "copyright", "software.name", "software.version", "software.releaseDate", "implementation.description", "implementation.url", "implementation.custodian"),
            selected = "url",
            size = 1,
            selectize = FALSE)
        )
      )
    }
  })

  output$show_security_filter <- renderUI({
    if (show_security_filter()) {
      # Get the list of security codes directly from the existing materialized view
      security_codes <- tbl(db_connection, "mv_get_security_endpoints") %>%
      distinct(code) %>%
      collect() %>%
      pull(code) 
      
      fluidRow(
        column(width = 4,
          selectInput(
            inputId = "auth_type_code",
            label = "Supported Authorization Type:",
            choices = security_codes,
            selected = "SMART-on-FHIR",
            size = 1,
            selectize = FALSE)
        )
      )
    }
  })

  profile_options <- reactive({
    query <- tbl(db_connection, "endpoint_supported_profiles_mv") %>%
      filter(fhir_version %in% !!input$fhir_version)

    if (input$vendor != ui_special_values$ALL_DEVELOPERS) {
      query <- query %>% filter(vendor_name == !!input$vendor)
    }

  res <-  query %>%
      select(profileurl) %>%
      distinct() %>%
      arrange(profileurl) %>%
      collect()

    # split(.$profileurl) %>%
    # purrr::map(~ .$profileurl)

  res <- split(res$profileurl, res$profileurl)

  profile_list <- list(
    "All Profiles" = ui_special_values$ALL_PROFILES
  )

  return(c(profile_list, res))
  })

  resource_options <- reactive({
    res <- get_supported_profiles(db_connection)
    req(input$fhir_version, input$vendor)

    res <- res %>%
    filter(fhir_version %in% input$fhir_version) %>%
    filter(resource != "")

    if (input$vendor != ui_special_values$ALL_DEVELOPERS) {
      res <- res %>% filter(vendor_name == input$vendor)
    }

    resource_list <- list(
        "All Resources" = ui_special_values$ALL_RESOURCES
    )

    res <- res %>%
    distinct(resource) %>%
    arrange(resource) %>%
    split(.$resource) %>%
    purrr::map(~ .$resource)
    return(c(resource_list, res))
  })


  checkbox_resources <- reactive({
    req(input$fhir_version, input$vendor)
    
    res <- tbl(db_connection, "mv_endpoint_resource_types")
    
    res <- res %>% 
      filter(fhir_version %in% !!input$fhir_version)
    
    if (input$vendor != ui_special_values$ALL_DEVELOPERS) {
      res <- res %>% filter(vendor_name == !!input$vendor)
    }
    
    res <- res %>%
      distinct(type) %>%
      arrange(type) %>%
      collect() %>%
      split(.$type) %>%
      purrr::map(~ .$type)
    
    return(res)
  })

  checkbox_resources_no_filter <- reactive({
    res <- tbl(db_connection, "mv_endpoint_resource_types") %>%
      distinct(type) %>%
      arrange(type) %>%
      collect() %>%
      split(.$type) %>%
      purrr::map(~ .$type)
    
    return(res)
  })

  #                                          #
  # Display Resource and Operations Checkbox #
  #                                          #

  output$show_resource_operation_checkboxes <- renderUI({
    if (show_resource_checkbox() && show_operation_checkbox()) {
      fluidPage(
        fluidRow(
          h2("FHIR Resource Types"),
          tags$a("Skip Past Resources", href = "#selectall", class = "show-on-focus-resources", "aria-label" = "Click the enter key to skip past the resource checkbox options and jump directly to select all and deselect all resource buttons"),
          column(width = 4,
            multiInput(
              inputId = "resources",
              width = "500px",
              label = "Click a resource on the left to add, and on the right to remove:",
              choices = checkbox_resources_no_filter(),
              selected = checkbox_resources_no_filter(),
              options = list(
                non_selected_header = "Choose resources:",
                selected_header = "Selected resources:"
              )
            ),
            actionButton("selectall", "Select All Resources", style = "margin-top: -15px; margin-bottom: 20px;"),
            actionButton("removeall", "Remove All Resources", style = "margin-top: -15px; margin-bottom: 20px;")
          ),
          column(width = 8,
            selectizeInput("operations", "Click in the box below to add or remove operations:",
            choices = c("read", "vread", "update", "patch", "delete", "history-instance", "history-type", "create", "search-type", "not specified"),
            selected = c("read"), multiple = TRUE, options = list("plugins" = list("remove_button"), "create" = TRUE, "persist" = FALSE), width = "100%"),
            actionButton("removeallops", "Clear All Operations", style = "margin-top: -15px;"),
            p("Note: When selecting multiple operations, only the resources that implement all selected operations will be displayed in the table and graph below.
            Choosing the 'not specified' option will display resources where no operation was defined in the CapabilityStatement / Conformance Resource.", style = "font-size:15px; margin-left:5px; margin-top:5px;")
          )
        )
      )
    }
  })

  #                     #
  # Resource Checkbox #
  #                     #

  current_selection <- reactiveVal(NULL)

  observeEvent(input$resources, {
    current_selection(input$resources)
  })

  observeEvent(input$selectall, {
    if (input$selectall == 0) {
      return(NULL)
    } else {
      updateMultiInput(session, "resources", label = "Click a resource on the left to add, and on the right to remove:", choices = checkbox_resources(), selected = checkbox_resources())
    }
  })

  observeEvent(input$removeall, {
    if (input$removeall == 0) {
      return(NULL)
    } else {
      current_selection(NULL)
      updateMultiInput(session, "resources", label = "Click a resource on the left to add, and on the right to remove:", choices = checkbox_resources())
    }
  })

  observeEvent(input$fhir_version, {
    if (!show_resource_checkbox() || is.null(current_selection())) {
      return(NULL)
    } else {
      updateMultiInput(session, "resources", label = "Click a resource on the left to add, and on the right to remove:", choices = checkbox_resources(), selected = current_selection())
    }
  })

  observeEvent(input$vendor, {
    if (!show_resource_checkbox() || is.null(current_selection())) {
      return(NULL)
    } else {
      updateMultiInput(session, "resources", label = "Click a resource on the left to add, and on the right to remove:", choices = checkbox_resources(), selected = current_selection())
    }
  })

  #                     #
  # Operations Checkbox #
  #                     #

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
            label = "Click in the box below to add or remove operations:",
            choices = c("read", "vread", "update", "patch", "delete", "history-instance", "history-type", "create", "search-type", "not specified"),
            selected = c("read"))
    }
  })

  # Resets the display if the user clicks the "Remove All Operations" button
  observeEvent(input$removeallops, {
    if (input$removeallops == 0) {
      return(NULL)
    } else {
      updateSelectizeInput(session, "operations",
              label = "Click in the box below to add or remove operations:",
              choices = c("read", "vread", "update", "patch", "delete", "history-instance", "history-type", "create", "search-type", "not specified"),
              options = list("plugins" = list("remove_button"), "create" = TRUE, "persist" = FALSE))
    }
  })

  #                                          #
  #   Display Resource and Profile Filters   #
  #                                          #

  output$resource_filter_tab <- renderUI({
    fluidPage(
      fluidRow(
        column(width = 12,
          selectInput(
            inputId = "profile_resource",
            label = "Resources:",
            choices = resource_options(),
            selected = ui_special_values$ALL_RESOURCES,
            selectize = FALSE,
            size = 1,
            width = paste0(max(nchar(profile_options())) * 8, "px")
          )
        )
      ),
      p("Note: DSTU2 endpoints will not be visible if resource filter selected.")
    )
  })

  output$profile_filter_tab <- renderUI({
    fluidPage(
      fluidRow(
        column(width = 12,
          selectInput(
            inputId = "profiles",
            label = "Profiles:",
            choices = profile_options(),
            selected = ui_special_values$ALL_PROFILES,
            selectize = FALSE,
            size = 1,
            width = paste0(max(nchar(profile_options())) * 8, "px")
          )
        )
      )
    )
  })

  output$show_resource_profiles_dropdown <- renderUI({
    if (show_profiles_filters()) {
      tagList(
        fluidRow(
          column(width = 12,
            tabsetPanel(id = "profile_resource_tab", type = "tabs",
              tabPanel("Profile Filtering", uiOutput("profile_filter_tab")),
              tabPanel("Resource Filtering", uiOutput("resource_filter_tab")))
          )
        )
      )
    }
  })

  # Resets the filters when switching between filtering tabs
  observeEvent(input$profile_resource_tab, {
      updateSelectInput(session, "profiles",
        label = "Profiles:",
        choices = profile_options(),
        selected = ui_special_values$ALL_PROFILES)

      updateSelectInput(session, "profile_resource",
        label = "Resources:",
        choices = resource_options(),
        selected = ui_special_values$ALL_RESOURCES)
  })

  observeEvent(input$show_details, {
    showModal(modalDialog(
      title = "All API Information Source Names",
      p(HTML(str_replace_all(get_endpoint_organization_list(input$show_details), ";", "<br>"))),
      easyClose = TRUE
  ))
  })

  observeEvent(input$show_contact_modal, {
  # Get contact data directly from the materialized view
  contact_data <- tbl(db_connection, "mv_contacts_info") %>% collect()
  
  showModal(modalDialog(
    title = "All Contacts",
    p(input$show_contact_modal),
    p(ifelse(is.na(
      contact_data %>%
        filter(url == input$show_contact_modal) %>%
        distinct(endpoint_names) %>%
        select(endpoint_names))
        ||
        contact_data %>%
        filter(url == input$show_contact_modal) %>%
        distinct(endpoint_names) %>%
        select(endpoint_names) == "",
      "-",
      contact_data %>%
      filter(url == input$show_contact_modal) %>%
      mutate(endpoint_names = strsplit(endpoint_names, ";")[[1]][1]) %>%
      distinct(endpoint_names) %>%
      select(endpoint_names)
    ),
    reactable::renderReactable({
      reactable(
        contact_data %>%
        mutate(contact_name = ifelse(is.na(contact_name), "N/A", contact_name)) %>%
        filter(url == input$show_contact_modal) %>%
        arrange(contact_preference) %>%
        mutate(contact_name = ifelse(is.na(contact_name), "-", contact_name)) %>%
        select(contact_name, contact_type, contact_value) %>%
        mutate(contact_value = ifelse(contact_value == "", "-", contact_value)),
            defaultColDef = colDef(
              align = "center"
            ),
            columns = list(
                contact_name = colDef(name = "Contact Name"),
                contact_type = colDef(name = "Contact Type"),
                contact_value = colDef(name = "Contact Info")
            ),
            groupBy = "contact_name"
      )
    }),
    easyClose = TRUE
  )))
  })

observeEvent(input$show_organization_modal, {
  showModal(modalDialog(
    title = "Organization Details",

    p(HTML(paste("<b>Organization Active Status:</b><br/>",
      paste(
        {
          active_vals <- get_org_active_information(db_connection) %>%
            filter(org_id == input$show_organization_modal) %>%
            pull(active)
          if (length(active_vals) == 0 || all(is.na(active_vals))) {
            "N/A"
          } else {
            active_vals
          }
        },
        collapse = "<br/>"
      )
    ))),

    p(HTML(paste("<b>Organization Identifiers:</b><br/>",
      paste(
        {
          identifier_vals <- get_org_identifiers_information(db_connection) %>%
          filter(org_id == input$show_organization_modal) %>%
          pull(identifier)
          if (length(identifier_vals) == 0 || all(is.na(identifier_vals))) {
            "N/A"
          } else {
            identifier_vals
          }
        },
    collapse = "<br/>"
      )
    ))),

    p(HTML(paste("<b>Organization Addresses:</b><br/>",
      paste(
        {
          address_vals <- get_org_addresses_information(db_connection) %>%
          filter(org_id == input$show_organization_modal) %>%
          pull(address)
          if (length(address_vals) == 0 || all(is.na(address_vals))) {
            "N/A"
          } else {
            address_vals
          }
        },
    collapse = "<br/>"
      )
    ))),

    easyClose = TRUE
  ))
})


# Current Endpoint that is selected to view in Modal
current_endpoint <- reactive({
  req(input$endpoint_popup)

  # Check if both URL and version are provided (from Endpoints tab)
  if (grepl("&&", input$endpoint_popup)) {
    splitString <- strsplit(input$endpoint_popup, "&&")[[1]]
    endpointURL <- splitString[1]
    endpoint_requested_fhir_version <- splitString[2]
  } else {
    # Only URL is provided (from Organizations tab)
    endpointURL <- input$endpoint_popup

    # Query DB for the most recent requested_fhir_version
    res <- tbl(db_connection, "selected_fhir_endpoints_mv") %>%
      filter(url == !!endpointURL) %>%
      arrange(desc(info_updated)) %>%
      collect()

    if (nrow(res) == 0) {
      warning(paste("No matching rows found for URL:", endpointURL))
      endpoint_requested_fhir_version <- NA 
    } else {
      endpoint_requested_fhir_version <- res$requested_fhir_version[1]
    }
  }
  current_endpoint_list <- list(url = endpointURL, requested_fhir_version = endpoint_requested_fhir_version)
  current_endpoint_list
})


### CHPL Products Modal Page ###
endpoint_products <- reactive({
  endpoint <- current_endpoint()
  res <- get_endpoint_products(db_connection, endpoint$url, endpoint$requested_fhir_version)
  res
})

output$endpoint_products_table <- DT::renderDataTable({
  datatable(endpoint_products(),
            colnames = c("Name", "Version", "CHPL ID", "API URL", "Certification Status", "Certification Edition", "Certification Date", "Last Modified in CHPL"),
            rownames = FALSE,
            selection = "none",
            options = list(scrollX = TRUE))
})

endpoint_products_page <- function() {
  page <- fluidPage(
    h1("Endpoint CHPL Products"),
    DT::dataTableOutput("endpoint_products_table"),
    p("Note: The software products shown in the table above are matched with the best guess possible given the information Lantern has available, and therefore may not be completely accurate.")
  )
  page
}


### IGs and Profiles Modal Page ###

endpoint_implementation_guides <- reactive({
  endpoint <- current_endpoint()

  implementation_guides <- get_endpoint_implementation_guide(db_connection, endpoint$url, endpoint$requested_fhir_version)
  implementation_guides
})

endpoint_profiles <- reactive({
  endpoint <- current_endpoint()

  profiles <- get_endpoint_supported_profiles(db_connection, endpoint$url, endpoint$requested_fhir_version)
  profiles

})

output$endpoint_IG_table <- DT::renderDataTable({
  datatable(endpoint_implementation_guides() %>% select(implementation_guide),
            colnames = c("Implementation_Guides"),
            rownames = FALSE,
            selection = "none",
            options = list(scrollX = TRUE))
})

output$endpoint_profile_table <- DT::renderDataTable({
  datatable(endpoint_profiles() %>% select(profileurl, profilename, resource),
            colnames = c("Profile URL", "Profile Name", "Resource"),
            rownames = FALSE,
            selection = "none",
            options = list(scrollX = TRUE))
})

implementation_guide_profiles_page <- function() {
  page <- fluidPage(
    h3("Implementation Guides"),
    DT::dataTableOutput("endpoint_IG_table"),

    h3("Endpoint Profiles"),
    DT::dataTableOutput("endpoint_profile_table"))
  page
}

### Capabilities Modal Page ###

# Required Capability Statement fields that we are tracking
required_fields <- c("status", "kind", "fhirVersion", "format", "date")

endpoint_fields <- reactive({
  endpoint <- current_endpoint()

  res <- get_endpoint_capstat_fields(db_connection, endpoint$url, endpoint$requested_fhir_version, "false")
  res
})

endpoint_extensions <- reactive({
  endpoint <- current_endpoint()

  res <- get_endpoint_capstat_fields(db_connection, endpoint$url, endpoint$requested_fhir_version, "true")
  res
})

endpoint_resources <- reactive({
  endpoint <- current_endpoint()

  res <- get_endpoint_resources(db_connection, endpoint$url, endpoint$requested_fhir_version)
  res

})

endpoint_smart_capabilities <- reactive({
  endpoint <- current_endpoint()

  res <- get_endpoint_smart_response_capabilities(db_connection, endpoint$url, endpoint$requested_fhir_version)
  res

})

output$endpoint_fields_table_required <- DT::renderDataTable({
  datatable(endpoint_fields() %>% filter(field %in% required_fields) %>% select(field, exist),
            colnames = c("Field Name", "Exists"),
            rownames = FALSE,
            selection = "none",
            options = list(scrollX = TRUE))
})

output$endpoint_fields_table_optional <- DT::renderDataTable({
  datatable(endpoint_fields() %>% select(field, exist),
            colnames = c("Field Name", "Exists"),
            rownames = FALSE,
            selection = "none",
            options = list(scrollX = TRUE))
})

output$endpoint_extensions_table <- DT::renderDataTable({
  datatable(endpoint_extensions() %>% select(field, exist),
            colnames = c("Extension Name", "Exists"),
            rownames = FALSE,
            selection = "none",
            options = list(scrollX = TRUE))
})


output$endpoint_resource_op_table <- reactable::renderReactable({
  reactable(
          endpoint_resources(),
          columns = list(
            Operation = colDef(
              aggregate = "count",
              format = list(aggregated = colFormat(prefix = "Total: "))
            ),
            Resource = colDef(
              minWidth = 150
            )
          ),
          groupBy = "Operation",
          sortable = TRUE,
          searchable = TRUE,
          striped = TRUE,
          showSortIcon = TRUE,
          defaultPageSize = 10,
          showPageSizeOptions = TRUE,
          pageSizeOptions = c(10, 25, 50, 100)

  )
})

output$smart_capabilities_table <- DT::renderDataTable({
  datatable(endpoint_smart_capabilities(),
            colnames = c("SMART Capabilities"),
            rownames = FALSE,
            selection = "none",
            options = list(scrollX = TRUE))
})

get_capability_statement_json <- reactive({
  endpoint <- current_endpoint()

  res <- get_capability_and_smart_response(db_connection, endpoint$url, endpoint$requested_fhir_version)

  capability_statement_json <- res$capability_statement

  if (length(res$capability_statement) <= 0) {
    capability_statement_json <- "{\"Not Available\": \"No Capability Statement Returned\"}"
  }

  capability_statement_json
})


get_smart_response_json <- reactive({
  endpoint <- current_endpoint()

  res <- get_capability_and_smart_response(db_connection, endpoint$url, endpoint$requested_fhir_version)

  smart_response_json <- res$smart_response

  if (length(res$smart_response) <= 0) {
    smart_response_json <- "{\"Not Available\": \"No SMART Response Returned\"}"
  }

  smart_response_json
})

endpoint_capabilities_page <- function() {
  page <- fluidPage(
    h1("Endpoint Capabilities"),
    bsCollapse(id = "capabilities_collapse", multiple = TRUE,
      bsCollapsePanel("Capability/Conformance Fields", fluidPage(
        h3("Required Fields"),
        DT::dataTableOutput("endpoint_fields_table_required"),
        h3("Optional Fields"),
        DT::dataTableOutput("endpoint_fields_table_optional"),
        h3("Extensions"),
        DT::dataTableOutput("endpoint_extensions_table"),
      ), style = "info"),
      bsCollapsePanel("Capability/Conformance Resources", reactable::reactableOutput("endpoint_resource_op_table"), style = "info"),
      bsCollapsePanel("SMART Response Fields", DT::dataTableOutput("smart_capabilities_table"), style = "info"),
      bsCollapsePanel("Capability Statement/Conformance Resource", renderJsonedit(jsonedit(get_capability_statement_json(),
              mode = "view", modes =  c("view", "code"),
              "onEditable" = htmlwidgets::JS("function() { return false;}"))
        ), style = "info"
      ),
      bsCollapsePanel("SMART Response", renderJsonedit(jsonedit(get_smart_response_json(),
              mode = "view", modes =  c("view", "code"),
              "onEditable" = htmlwidgets::JS("function() { return false;}"))
          ), style = "info"
      )
    )
  )
}


### Organizations Modal Page ###

 get_endpoint_list_orgs <- reactive({
    endpoint <- current_endpoint()

    # Get the actual cap_fhir_version using the url only
    cap_fhir_ver <- get_endpoint_list_matches(db_connection, fhir_version = NULL, vendor = NULL) %>%
      filter(url == endpoint$url) %>%
      pull(fhir_version) %>% 
      unique()
    
    # Now use cap_fhir_ver to filter
    res <- get_endpoint_list_matches(db_connection, fhir_version = NULL, vendor = NULL)
    res <- res %>%
      filter(url == endpoint$url) %>%
      filter(fhir_version == cap_fhir_ver) %>%
      mutate(organization_name = if_else(organization_name == "Unknown", "Not Available", organization_name))

    res
  })

  output$endpoint_list_org_table <- DT::renderDataTable({
    datatable(get_endpoint_list_orgs() %>% distinct(organization_name),
              colnames = c("Organization Name"),
              rownames = FALSE,
              selection = "none",
              options = list(scrollX = TRUE))
  })


organization_endpoint_page <- function() {
  page <- fluidPage(
    h1("Endpoint Organizations"),
    DT::dataTableOutput("endpoint_list_org_table")
    )
  page
}

### Endpoint Details Modal Page ###
get_range <- function(date) {
    if (all(date == "Past 7 days")) {
      range <- "604800"
    } else if (all(date == "Past 14 days")) {
      range <- "1209600"
    } else if (all(date == "Past 30 days")) {
      range <- "2592000"
    } else {
      range <- "maxdate.maximum"
    }
    range
}

response_time_xts <- reactive({
  endpoint <- current_endpoint()

  range <- get_range(input$date)
  res <- get_endpoint_response_time(db_connection, range, endpoint$url, endpoint$requested_fhir_version)
  # convert to xts format for use in dygraph
  xts(x = cbind(res$response),
      order.by = res$date 
  )
})

output$no_plot <- renderText({
  if (nrow(response_time_xts()) == 0) {
    "Sorry, there isn't enough data to show response times!"
  }
})

output$endpoint_response_time_plot <- renderDygraph({
  if (nrow(response_time_xts()) > 0) {
    dygraph(response_time_xts(),
          main = "Endpoint Response Time",
          ylab = "seconds",
          xlab = "Date") %>%
    dyAxis("y", valueRange = c(-1.30, NA)) %>%
    dySeries("V1", label = "ResponseTime") %>%
    dyLegend(width = 450)
  }
})

output$plot_note_text <- renderUI({
  note_info <- "There are many variables that influence response time, such
    as network congestion, geographic location, hosting configurations, etc.
    This graphic only intends to convey the health of the FHIR endpoint ecosystem
    as a whole, drastic changes to which may represent some widespread issue
    throughout the ecosystem."
  res <- paste("<div style='font-size: 18px;'><b>Note:</b>", note_info, "</div>")
  HTML(res)
})

endpoint_http_responses <- reactive({
  endpoint <- current_endpoint()
  range <- get_range(input$http_date)
  res <- get_endpoint_http_over_time(db_connection, range, endpoint$url, endpoint$requested_fhir_version) %>%
  left_join(app$http_response_code_tbl(), by = c("http_response" = "code")) %>%
  mutate(http_response = paste(http_response, "-", label)) %>%
  select(date, http_response)
  res
})

endpoint_http_codes_table <- reactive({
  endpoint <- current_endpoint()

  range <- get_range(input$http_date)
  res <- get_endpoint_http_over_time(db_connection, range, endpoint$url, endpoint$requested_fhir_version)

  http_code_table <- app$http_response_code_tbl() %>%
  inner_join(res, by = c("code" = "http_response")) %>%
  distinct(code, label) %>%
  mutate(row_num = row_number()) %>%
  select(code, row_num, label)
})

endpoint_http_responses_mapping <- reactive({
  endpoint <- current_endpoint()

  range <- get_range(input$http_date)
  res <- get_endpoint_http_over_time(db_connection, range, endpoint$url, endpoint$requested_fhir_version)

  http_code_table <- endpoint_http_codes_table()

  res <- res %>%
  left_join(http_code_table, by = c("http_response" = "code")) %>%
  tidyr::replace_na(list(row_num = 0)) %>%
  mutate(http_response = paste(http_response, "-", label)) %>%
  select(date, http_response, row_num)
  res

})

create_dygraph_json <- reactive({
  res <- endpoint_http_responses_mapping() %>%
  distinct(row_num, http_response) %>%
  rename(v = row_num, label = http_response)

  toJSON(res)

})

endpoint_http_responses_xts <- reactive({
  res <- endpoint_http_responses_mapping()
  xts(x = cbind(res$row_num), order.by = res$date)
})

output$http_no_plot <- renderText({
  if (nrow(endpoint_http_responses_xts()) == 0) {
    "Sorry, there isn't enough data to show http responses over time!"
  }
})

output$endpoint_http_response_plot <- renderDygraph({
  if (nrow(endpoint_http_responses_xts()) > 0) {
    dygraph(endpoint_http_responses_xts(),
          main = "Endpoint HTTP Responses",
          ylab = "HTTP Codes",
          xlab = "Date") %>%
    dyAxis("y", valueRange = c(-0.2, nrow(endpoint_http_codes_table()) + .5),
    axisLabelWidth = 70, ticker = htmlwidgets::JS(
      paste("function(min, max, pixels, opts, dygraph, vals) {
      return ", create_dygraph_json(), ";}")),
      valueFormatter = htmlwidgets::JS(
      paste("function(v){
        let jsonfile = `", create_dygraph_json(), "`;
        let jsonobj = JSON.parse(jsonfile);
        for (let obj of jsonobj) {
          if (obj.v === v) {
            return obj.label;
          }
        }
      }"))
    ) %>%
    dySeries("V1", label = "HTTPCode") %>%
    dyLegend(width = 450)
  }
})

output$endpoint_http_response_table <- reactable::renderReactable({
  reactable(
        endpoint_http_responses() %>% select(date, http_response) %>% mutate_all(as.character),
        defaultColDef = colDef(
          align = "center"
        ),
        columns = list(
            date = colDef(name = "Date", sortable = TRUE),
            http_response = colDef(name = "HTTP Response", sortable = FALSE)
        ),
        searchable = TRUE,
        showSortIcon = TRUE,
        highlight = TRUE,
        defaultPageSize = 10
  )
})

 detailPage <- function() {

  endpoint <- current_endpoint()

  detailsInfo <- get_details_page_info(endpoint$url, endpoint$requested_fhir_version, db_connection)
  metricsInfo <- get_details_page_metrics(endpoint$url, endpoint$requested_fhir_version)

  page <- fluidPage(
    h1("Endpoint Details"),
    tags$p(paste0("Updated at ", as.character(detailsInfo$info_updated), " | Created at ", as.character(detailsInfo$info_created)), style = "font-style: italic;"),
    br(),
    mainPanel(
      fluidRow(
        infoBox("FHIR Version", as.character(detailsInfo$fhir_version), icon = icon("code"), width = 6),
        infoBox("Supported Versions", tags$p(as.character(detailsInfo$supported_versions), style = "overflow-wrap: break-word;"), icon = icon("check"), width = 6, color = "red")
      ),
      fluidRow(
        infoBox("Vendor", as.character(detailsInfo$vendor_name), icon = icon("building"), width = 6, color = "green"),
        infoBox("List Source", tags$p(as.character(detailsInfo$list_source), style = "overflow-wrap: break-word;"), icon = icon("list"), width = 6, color = "teal")
      ),
      h3("Software"),
      fluidRow(
        infoBox("Software Name", as.character(detailsInfo$software_name), icon = icon("code-branch"), width = 6, color = "blue"),
          infoBox("Software Version", as.character(detailsInfo$software_version), icon = icon("code"), width = 6, color = "orange"),
          infoBox("Format", as.character(detailsInfo$format), icon = icon("file-code"), width = 6, color = "yellow"),
          infoBox("Security", as.character(detailsInfo$security), icon = icon("lock", lib = "glyphicon"), width = 6, color = "purple")
      ),
      fluidRow(
          tags$p(paste0("Last Software Version Update: ", as.character(detailsInfo$software_releasedate)), style = "font-style: italic;")
      ),
      br(),
      uiOutput("show_date_filters"),
      textOutput("no_plot"),
      dygraphOutput("endpoint_response_time_plot"),
      p("Click and drag on plot to zoom in, double-click to zoom out."),
      htmlOutput("plot_note_text"),
      br(),
      uiOutput("show_http_date_filters"),
      fluidRow(
            textOutput("http_no_plot"),
            dygraphOutput("endpoint_http_response_plot"),
            p("Click and drag on plot to zoom in, double-click to zoom out.")
          ),
      fluidRow(
        reactable::reactableOutput("endpoint_http_response_table")
      )
    ),
    sidebarPanel(
      h2("Metrics"),
      h4("Status:"),
      p(metricsInfo$status),
      h4("Last HTTP Response:"),
      p(metricsInfo$http_response),
      h4("Availability:"),
      p(metricsInfo$availability),
      h4("Capability Statement Returned:"),
      p(metricsInfo$cap_stat_exists),
      h4("Errors:"),
      p(metricsInfo$errors),
      h4("SMART HTTP Response:"),
      p(metricsInfo$smart_http_response)
    )
  )
}


  ### Endpoint Popup Modal ###
  observeEvent(input$endpoint_popup, {
    endpoint <- current_endpoint()
    showModal(modalDialog(
      title = "Endpoint Details",
      h1("Endpoint URL:"),
      h3(tags$a(as.character(endpoint$url)), style = "word-wrap: break-word;"),
      p("Note: The blue boxes found in many of the tabs below can be clicked on and expanded to display additional information."),
      tabsetPanel(id = "endpoint_modal_tabset", type = "tabs",
          tabPanel("Details", detailPage()),
          tabPanel("Organizations", organization_endpoint_page()),
          tabPanel("Capabilities", endpoint_capabilities_page()),
          tabPanel("Implementation Guides & Profiles", implementation_guide_profiles_page()),
          tabPanel("Products", endpoint_products_page())
      ),
      size = "l",
      easyClose = TRUE
    ))
  })

output$filter_profile_table <- DT::renderDataTable({
      DT::datatable(
        selected_fhir_endpoint_profiles(),
        escape = FALSE,
        colnames = c("Endpoint", "Profile URL", "Profile Name", "Resource", "FHIR Version", "Certified API Developer Name"),
        options = list(
        lengthMenu = c(5, 30, 50),
        pageLength = 5,
        scrollX = TRUE
        )
        )
})

}
