# Values Module

valuesmodule_UI <- function(id) {
  ns <- NS(id)
  tagList(
    h1("Values of FHIR CapabilityStatement / Conformance Fields"),
    p("This is the set of values from the endpoints for a given field included in the FHIR CapabilityStatement / Conformance Resources."),
    fluidRow(
      column(width = 7,
             h2("Field Values"),
             textInput(ns("values_search_query"), "Search: ", value = ""),
             reactable::reactableOutput(ns("capstat_values_table")),
             fluidRow(
              column(3, 
                div(style = "display: flex; justify-content: flex-start;", 
                    uiOutput(ns("values_prev_button_ui"))
                )
              ),
              column(6,
                div(style = "display: flex; justify-content: center; align-items: center; gap: 10px; margin-top: 8px;",
                    numericInput(ns("values_page_selector"), label = NULL, value = 1, min = 1, max = 1, step = 1, width = "80px"),
                    textOutput(ns("values_page_info"), inline = TRUE)
                )
              ),
              column(3, 
                div(style = "display: flex; justify-content: flex-end;",
                    uiOutput(ns("values_next_button_ui"))
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
  values_page_size <- 10
  values_page_state <- reactiveVal(1)

  # Reset to first page on any filter/search change
  observeEvent(list(sel_fhir_version(), sel_vendor(), sel_capstat_values(), input$values_search_query), {
    values_page_state(1)
    updateNumericInput(session, "values_page_selector", value = 1)
  })

  # Page navigation buttons
  output$values_prev_button_ui <- renderUI({
  if (values_page_state() > 1) {
    actionButton(ns("values_prev_page"), "Previous", icon = icon("arrow-left"))
  } else {
    NULL
  }
  })

  output$values_next_button_ui <- renderUI({
    if (values_page_state() < values_total_pages()) {
      actionButton(ns("values_next_page"), "Next", icon = icon("arrow-right"))
    } else {
      NULL
    }
  })

  # Sync page selector
  observe({
    updateNumericInput(session, "values_page_selector", 
                      max = values_total_pages(),
                      value = values_page_state())
  })

  # Manual page input
  observeEvent(input$values_page_selector, {
    if (!is.null(input$values_page_selector) && !is.na(input$values_page_selector)) {
      new_page <- max(1, min(input$values_page_selector, values_total_pages()))
      values_page_state(new_page)
      if (new_page != input$values_page_selector) {
        updateNumericInput(session, "values_page_selector", value = new_page)
      }
    }
  })


  observeEvent(input$values_next_page, {
    if (values_page_state() < values_total_pages()) values_page_state(values_page_state() + 1)
  })

  observeEvent(input$values_prev_page, {
    if (values_page_state() > 1) values_page_state(values_page_state() - 1)
  })

  output$values_page_info <- renderText({
    paste("of", values_total_pages())
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

values_base_sql <- reactive({
  req(sel_fhir_version(), sel_vendor(), sel_capstat_values())

  versions <- paste0("'", sel_fhir_version(), "'", collapse = ", ")
  field <- sel_capstat_values()
  vendor <- sel_vendor()

  # Valid FHIR versions for the selected field
  valid_versions <- tbl(db_connection, 
                        sql(paste0("SELECT unnest(fhir_versions) AS version 
                                    FROM get_value_versions_mv 
                                    WHERE field = '", field, "'"))) %>%
                    collect() %>%
                    pull(version)
  valid_versions_sql <- paste0("('", paste(valid_versions, collapse = "', '"), "')")

  vendor_filter <- if (vendor != ui_special_values$ALL_DEVELOPERS) {
    paste0("AND \"Developer\" = '", vendor, "'")
  } else {
    ""
  }

  search_filter <- ""
  if (!is.null(input$values_search_query) && input$values_search_query != "") {
    q <- gsub("'", "''", input$values_search_query)
    search_filter <- paste0("AND (\"Developer\" ILIKE '%", q, "%' OR 
                                  \"FHIR Version\" ILIKE '%", q, "%' OR 
                                  field_value ILIKE '%", q, "%')")
  }

  paste0("FROM selected_fhir_endpoints_values_mv 
         WHERE field = '", field, "'
           AND \"FHIR Version\" IN ", valid_versions_sql, " 
           AND \"FHIR Version\" IN (", versions, ") ",
           vendor_filter, " ",
           search_filter)
})

values_total_pages <- reactive({
  count_query <- paste0("SELECT COUNT(*) as count ", values_base_sql())
  count <- tbl(db_connection, sql(count_query)) %>% collect() %>% pull(count)
  max(1, ceiling(count / values_page_size))
})


all_capstat_values <- reactive({
  query <- paste0(
    "SELECT \"Developer\", \"FHIR Version\", field_value, \"Endpoints\" ",
    values_base_sql()
  )
  tbl(db_connection, sql(query)) %>% collect()
})

paged_capstat_values <- reactive({
  limit <- values_page_size
  offset <- (values_page_state() - 1) * values_page_size

  query <- paste0(
    "SELECT \"Developer\", \"FHIR Version\", field_value, \"Endpoints\" ",
    values_base_sql(),
    " ORDER BY \"Endpoints\" DESC 
      LIMIT ", limit, " OFFSET ", offset
  )

  tbl(db_connection, sql(query)) %>% collect()
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
                defaultPageSize = values_page_size
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
