# Values Module

valuesmodule_UI <- function(id) {
  ns <- NS(id)
  tagList(
    h1("Values of FHIR CapabilityStatement / Conformance Fields"),
    p("This is the set of values from the endpoints for a given field included in the FHIR CapabilityStatement / Conformance Resources."),
    fluidRow(
      column(width = 7,
             h2("Field Values"),
             reactable::reactableOutput(ns("capstat_values_table"))
            ),
      column(width = 5,
             h2("Endpoints that Include a Value for the Given Field"),
             uiOutput(ns("values_chart")),
      )
    ),
  )
}

valuesmodule <- function(
  input,
  output,
  session,
  sel_fhir_version,
  sel_vendor,
  sel_capstat_values
) {

  ns <- session$ns

  get_value_versions <- reactive({
    res <- isolate(app_data$capstat_fields())
    req(sel_capstat_values())
    res <- res %>%
    group_by(field) %>%
    arrange(fhir_version, .by_group = TRUE) %>%
    subset(field == sel_capstat_values())
    versions <- c(unique(res$fhir_version))
    versions
  })

  get_value_table_header <- reactive({
    res <- isolate(app_data$capstat_fields())
    req(sel_capstat_values(), sel_fhir_version())
    header <- ""
    if (length(sel_fhir_version()) == 1) {
      header <- sel_capstat_values()
    } else {
      res <- res %>%
      group_by(field) %>%
      arrange(fhir_version, .by_group = TRUE) %>%
      subset(field == sel_capstat_values()) %>%
      mutate(fhir_version_name = case_when(
      fhir_version %in% dstu2 ~ "DSTU2",
      fhir_version %in% stu3 ~ "STU3",
      fhir_version %in% r4 ~ "R4",
      TRUE ~ "DSTU2"
      )) %>%
      summarise(fhir_version_names = paste(unique(fhir_version_name), collapse = ", "))
      versions <- res %>% pull(2)
      header <- paste(sel_capstat_values(), " (", versions, ")", sep = "")
    }
    header
  })

  selected_fhir_endpoints <- reactive({
    res <- isolate(app_data$capstat_values())
    req(sel_fhir_version(), sel_vendor())
    # If the selected dropdown value for the fhir version is not the default "All FHIR Versions", filter
    # the capability statement fields by which fhir version they're associated with
    res <- res %>% filter(filter_fhir_version %in% sel_fhir_version())
    # Same as above but with the vendor dropdown
    if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
      res <- res %>% filter(vendor_name == sel_vendor())
    }
    # Filter by the versions that the given field exists in
    value_versions_list <- get_value_versions()
    res <- res %>% filter(filter_fhir_version %in% value_versions_list)
    # Repeat with filtering fields to see values
    res <- res %>%
      rename(fhirVersion = fhir_version, software.name = software_name, software.version = software_version, software.releaseDate = software_release_date, implementation.description = implementation_description, implementation.url = implementation_url, implementation.custodian = implementation_custodian) %>%
      group_by_at(vars("vendor_name", "filter_fhir_version", sel_capstat_values())) %>%
      count() %>%
      rename(Endpoints = n, Developer = vendor_name, "FHIR Version" = filter_fhir_version) %>%
      rename(field_value = sel_capstat_values()) %>%
      # If the field is empty then put an "[Empty]" string
      tidyr::replace_na(list(field_value = "[Empty]"))
    res
  })

  capstatPageSizeNum <- reactiveVal(NULL)

  capstat_values_list <- reactive({
    if (is.null(capstatPageSizeNum())) {
      capstatPageSizeNum(10)
    }
    get_capstat_values_list(selected_fhir_endpoints())
  })

  output$capstat_values_table <- reactable::renderReactable({
    reactable(capstat_values_list() %>% select(Developer, "FHIR Version", field_value, Endpoints),
                columns = list(
                  field_value = colDef(name = get_value_table_header())
                ),
                sortable = TRUE,
                searchable = TRUE,
                striped = TRUE,
                showSortIcon = TRUE,
                defaultPageSize = isolate(capstatPageSizeNum())
    )
  })

  observeEvent(input$capstat_values_table_state$length, {
    page <- input$capstat_values_table_state$length
    capstatPageSizeNum(page)
  })

  # Group by who has added a value vs who hasn't
  #
  # EXAMPLE:
  # capstat_values_list                   returned value
  # field_value      Endpoints            field_value   Endpoints   used
  # 1.0.1            3                    1.0.1         3           yes
  # 3.4.1            6                    3.4.1         6           yes
  # [Empty]          4                    [Empty]       4           no
  is_field_being_used <- reactive({
    capstat_values_list() %>%
    # necessary to ungroup because you can't select a subset of fields in a dataset
    # that is grouped
    ungroup() %>%
    select(c(Endpoints, field_value)) %>%
    # create a new column called "used"
    # if the field is not being used, set it to "no", otherwise set it to "yes"
    mutate(used = ifelse(field_value == "[Empty]", "no", "yes"))
  })

  # Gets the total number of endpoints that are using the currently selected field
  being_used <- reactive({
    # Filter by the endpoints that have a value in the currently selected field,
    # then pull the Endpoints column which has the count of endpoints
    #
    # EXAMPLE:
    # is_field_being_used                     res
    # field_value   Endpoints   used          Endpoints
    # 1.0.1         3           yes           3
    # 3.4.1         6           yes           6
    # [Empty]       4           no
    res <- is_field_being_used() %>%
      filter(used == "yes") %>%
      pull(Endpoints)

    # Get the total of all of the values in the Endpoints column if the column
    # is not empty. If the column is empty then the total is 0.
    total_endpts <- 0
    if (!is.null(res)) {
      total_endpts <- sum(res)
    }
    total_endpts
  })

  # Gets the total number of endpoints that are not using the currently selected field
  not_being_used <- reactive({
    # Filter by the endpoints that don't have a value in the currently selected field,
    # then pull the Endpoints column which has the count of endpoints
    #
    # EXAMPLE:
    # is_field_being_used                     res
    # field_value   Endpoints   used          Endpoints
    # 1.0.1         3           yes           4
    # 3.4.1         6           yes
    # [Empty]       4           no
    res <- is_field_being_used() %>%
      filter(used == "no") %>%
      pull(Endpoints)

    # Get the total of all of the values in the Endpoints column if the column
    # is not empty. If the column is empty then the total is 0.
    total_endpts <- 0
    if (!is.null(res)) {
      total_endpts <- sum(res)
    }
    total_endpts
  })

  # Data format for the Pie Chart
  percent_used_chart <- reactive({
    data.frame(
      group = c("Yes", "No"),
      value = c(being_used(), not_being_used())
    )
  })

  output$values_chart <- renderUI({
    if (nrow(subset(percent_used_chart(), value != 0))) {
      tagList(
        plotOutput(ns("values_chart_plot"), height = 600)
      )
    } else {
      tagList(
        plotOutput(ns("values_chart_empty_plot"), height = 600)
      )
    }
  })

  # Pie chart of the percent of the endpoints that use the given field
  output$values_chart_plot <-  renderCachedPlot({
      ggplot(percent_used_chart(), aes(x = "", y = value, fill = group)) +
      geom_col(width = 0.8) +
      geom_bar(stat = "identity") +
      # Turns the plot into a Pie Chart
      coord_polar("y", start = 0) +
      # Change Legend label
      labs(fill = "Includes a Value \nfor the Given Field") +
      # Only display labels that are non-zero, position the label in the middle of the pie chart area, and increase the label size
      geom_text(data = subset(percent_used_chart(), value != 0), aes(label = value), position = position_stack(vjust = 0.5), size = 10) +
      # Increase label size
      theme(legend.text = element_text(size = 20),
            legend.title = element_text(size = 20),
            # remove axes labels
            axis.text = element_blank(),
            # remove line around pie chart
            panel.grid = element_blank(),
            # remove x & y axis labels
            axis.title.y = element_blank(),
            axis.title.x = element_blank())
    },
    sizePolicy = sizeGrowthRatio(width = 300,
                                 height = 400,
                                 growthRate = 1.2),
    res = 72,
    cache = "app",
    cacheKeyExpr = {
      list(sel_fhir_version(), sel_vendor(), sel_capstat_values())
    }
  )

  # Pie chart of the percent of the endpoints that use the given field without the labels to support null data
  output$values_chart_empty_plot <-  renderPlot({
      ggplot(percent_used_chart(), aes(x = "", y = value, fill = group)) +
      geom_col(width = 0.8) +
      geom_bar(stat = "identity") +
      # Turns the plot into a Pie Chart
      coord_polar("y", start = 0) +
      # Change Legend label
      labs(fill = "Includes a Value \nfor the Given Field") +
      # Increase label size
      theme(legend.text = element_text(size = 20),
            legend.title = element_text(size = 20),
            # remove axes labels
            axis.text = element_blank(),
            # remove line around pie chart
            panel.grid = element_blank(),
            # remove x & y axis labels
            axis.title.y = element_blank(),
            axis.title.x = element_blank())
    })
}
