# Values Module

valuesmodule_UI <- function(id) {
  ns <- NS(id)
  tagList(
    h1("Values of FHIR CapabilityStatement / Conformance Fields"),
    p("This is the set of values from the endpoints for a given field included in the FHIR CapabilityStatement / Conformance Resources."),
    fluidRow(
      column(width = 7,
             h2("Field Values"),
             reactable::reactableOutput(ns("capstat_values_table")),
             fluidRow(
              column(3, 
                div(style = "display: flex; justify-content: flex-start;", 
                    uiOutput(ns("prev_button_ui"))
                )
              ),
              column(6,
                div(style = "display: flex; justify-content: center; align-items: center; gap: 10px; margin-top: 8px;",
                    numericInput(ns("page_selector"), label = NULL, value = 1, min = 1, max = 1, step = 1, width = "80px"),
                    textOutput(ns("page_info"), inline = TRUE)
                )
              ),
              column(3, 
                div(style = "display: flex; justify-content: flex-end;",
                    uiOutput(ns("next_button_ui"))
                )
              )
            ),
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
  page_size <- 10
  page_state <- reactiveVal(1)

  total_pages <- reactive({
  total <- capstat_total_count()
  max(1, ceiling(total / page_size))
})

  # Update page selector max when total pages change
  observe({
    updateNumericInput(session, "page_selector", 
                      max = total_pages(),
                      value = page_state())
  })

  # Handle page selector input
  observeEvent(input$page_selector, {
    if (!is.null(input$page_selector) && !is.na(input$page_selector)) {
      new_page <- max(1, min(input$page_selector, total_pages()))
      page_state(new_page)
      
      # Update the input if user entered invalid value
      if (new_page != input$page_selector) {
        updateNumericInput(session, "page_selector", value = new_page)
      }
    }
  })

  # Reset to first page on any filter/search change
  observeEvent(list(sel_fhir_version(), sel_vendor(), sel_capstat_values()), {
    page_state(1)
  })

  # Page navigation buttons
  output$prev_button_ui <- renderUI({
  if (page_state() > 1) {
    actionButton(ns("prev_page"), "Previous", icon = icon("arrow-left"))
  } else {
    NULL
  }
  })

  output$next_button_ui <- renderUI({
    if (page_state() < total_pages()) {
      actionButton(ns("next_page"), "Next", icon = icon("arrow-right"))
    } else {
      NULL
    }
  })

  output$page_info <- renderText({
    paste("of", total_pages())
  })

  observeEvent(input$next_page, {
    if (page_state() < total_pages()) page_state(page_state() + 1)
  })

  observeEvent(input$prev_page, {
    if (page_state() > 1) page_state(page_state() - 1)
  })

  output$current_page_info <- renderText({
    paste("Page", page_state(), "of", total_pages())
  })

  get_value_versions <- reactive({
  req(sel_capstat_values())
  # Query the materialized view for the selected field
  result <- tbl(db_connection, 
                sql(paste0("SELECT unnest(fhir_versions) AS fhir_version 
                           FROM get_value_versions_mv 
                           WHERE field = '", sel_capstat_values(), "'"))) %>%
    collect()
  # Extract the versions as a character vector
  if (nrow(result) > 0) {
    versions <- result$fhir_version
  } else {
    versions <- character(0)  # Empty character vector if no results
  }
  return(versions)
})

  get_value_table_header <- reactive({
    res <- isolate(get_capstat_fields(db_connection))
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

get_capstat_values_list <- function(capstat_values_tbl) {
  res <- capstat_values_tbl
}

all_capstat_values <- reactive({
    req(sel_fhir_version(), sel_vendor(), sel_capstat_values())
    # Get the selected values
    fhir_versions <- sel_fhir_version()
    vendor <- sel_vendor()
    field_name <- sel_capstat_values()
    # Create SQL for FHIR versions filtering
    fhir_versions_sql <- paste0("('", paste(fhir_versions, collapse = "', '"), "')")
    # Get the valid versions for the selected field
    valid_versions <- tbl(db_connection, 
                        sql(paste0("SELECT unnest(fhir_versions) AS version 
                                  FROM get_value_versions_mv 
                                  WHERE field = '", field_name, "'"))) %>%
                      collect() %>%
                      pull(version)
    # Create a comma-separated string of valid versions for SQL
    valid_versions_sql <- paste0("('", paste(valid_versions, collapse = "', '"), "')")
    # Build the base query with field and FHIR version filtering
    query_str <- paste0("
        SELECT \"Developer\", \"FHIR Version\", field_value, \"Endpoints\"
        FROM selected_fhir_endpoints_values_mv
        WHERE field = '", field_name, "'
        AND \"FHIR Version\" IN ", fhir_versions_sql, "
        AND \"FHIR Version\" IN ", valid_versions_sql)
    # Add vendor filtering if needed
    if (vendor != ui_special_values$ALL_DEVELOPERS) {
        query_str <- paste0(query_str, " AND \"Developer\" = '", vendor, "'")
    }
    res <- tbl(db_connection, sql(query_str)) %>% collect()
    return(res)
})

paged_capstat_values <- reactive({
  req(sel_fhir_version(), sel_vendor(), sel_capstat_values())

  fhir_versions <- sel_fhir_version()
  vendor <- sel_vendor()
  field_name <- sel_capstat_values()

  fhir_versions_sql <- paste0("('", paste(fhir_versions, collapse = "', '"), "')")
  valid_versions <- tbl(db_connection, 
                        sql(paste0("SELECT unnest(fhir_versions) AS version 
                                    FROM get_value_versions_mv 
                                    WHERE field = '", field_name, "'"))) %>%
                    collect() %>%
                    pull(version)
  valid_versions_sql <- paste0("('", paste(valid_versions, collapse = "', '"), "')")

  limit <- page_size
  offset <- (page_state() - 1) * page_size

  query_str <- paste0("
    SELECT \"Developer\", \"FHIR Version\", field_value, \"Endpoints\"
    FROM selected_fhir_endpoints_values_mv
    WHERE field = '", field_name, "'
      AND \"FHIR Version\" IN ", fhir_versions_sql, "
      AND \"FHIR Version\" IN ", valid_versions_sql)

  if (vendor != ui_special_values$ALL_DEVELOPERS) {
    query_str <- paste0(query_str, " AND \"Developer\" = '", vendor, "'")
  }

  query_str <- paste0(query_str, " ORDER BY \"Endpoints\" DESC LIMIT ", limit, " OFFSET ", offset)

  tbl(db_connection, sql(query_str)) %>% collect()
})

capstat_total_count <- reactive({
  req(sel_fhir_version(), sel_vendor(), sel_capstat_values())

  fhir_versions <- sel_fhir_version()
  vendor <- sel_vendor()
  field_name <- sel_capstat_values()

  fhir_versions_sql <- paste0("('", paste(fhir_versions, collapse = "', '"), "')")
  valid_versions <- tbl(db_connection, 
                        sql(paste0("SELECT unnest(fhir_versions) AS version 
                                    FROM get_value_versions_mv 
                                    WHERE field = '", field_name, "'"))) %>%
                    collect() %>%
                    pull(version)

  valid_versions_sql <- paste0("('", paste(valid_versions, collapse = "', '"), "')")

  count_query <- paste0("
    SELECT COUNT(*) as count
    FROM selected_fhir_endpoints_values_mv
    WHERE field = '", field_name, "'
      AND \"FHIR Version\" IN ", fhir_versions_sql, "
      AND \"FHIR Version\" IN ", valid_versions_sql)

  if (vendor != ui_special_values$ALL_DEVELOPERS) {
    count_query <- paste0(count_query, " AND \"Developer\" = '", vendor, "'")
  }

  tbl(db_connection, sql(count_query)) %>% collect() %>% pull(count)
})

  output$capstat_values_table <- reactable::renderReactable({
    reactable(paged_capstat_values() %>% select(Developer, "FHIR Version", field_value, Endpoints),
                columns = list(
                  field_value = colDef(name = get_value_table_header())
                ),
                sortable = TRUE,
                searchable = TRUE,
                striped = TRUE,
                showSortIcon = TRUE,
                defaultPageSize = page_size
    )
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
    all_capstat_values() %>%
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
