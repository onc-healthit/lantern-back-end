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

  values_total_pages <- reactive({
    total <- capstat_total_count()
    max(1, ceiling(total / values_page_size))
  })

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

  get_value_table_header <- reactive({
    req(sel_capstat_values(), sel_fhir_version())

    valid_versions <- valid_field_versions()
      header <- ""
      if (length(sel_fhir_version()) == 1) {
        header <- sel_capstat_values()
    } else {
        version_labels <- sort(unique(
          case_when(
            valid_versions %in% dstu2 ~ "DSTU2",
            valid_versions %in% stu3  ~ "STU3",
            valid_versions %in% r4    ~ "R4",
            TRUE                      ~ "DSTU2"
          )
        ))

      header <- paste0(sel_capstat_values(), " (", paste(version_labels, collapse = ", "), ")")
    }

    header
})
  valid_field_versions <- reactive({
    req(sel_capstat_values())

    res <- tbl(db_connection, 
        sql(paste0("SELECT unnest(fhir_versions) AS version 
                    FROM get_value_versions_mv 
                    WHERE field = '", sel_capstat_values(), "'"))) %>%
      collect() %>%
      pull(version)

    res
  })

  get_base_values_sql <- reactive({
    req(sel_fhir_version(), sel_vendor(), sel_capstat_values())

    fhir_versions <- sel_fhir_version()
    vendor <- sel_vendor()
    field_name <- sel_capstat_values()

    fhir_versions_sql <- paste0("('", paste(fhir_versions, collapse = "', '"), "')")
    valid_versions <- valid_field_versions()

    valid_versions_sql <- paste0("('", paste(valid_versions, collapse = "', '"), "')")

    search_filter <- ""
    if (!is.null(input$values_search_query) && input$values_search_query != "") {
      q <- gsub("'", "''", input$values_search_query)
      search_filter <- paste0("AND (\"Developer\" ILIKE '%", q, "%' OR 
                                    \"FHIR Version\" ILIKE '%", q, "%' OR 
                                    field_value ILIKE '%", q, "%')")
    }

    sql_base <- paste0("
      FROM selected_fhir_endpoints_values_mv
      WHERE field = '", field_name, "'
        AND \"FHIR Version\" IN ", fhir_versions_sql, "
        AND \"FHIR Version\" IN ", valid_versions_sql, 
        search_filter)

    if (vendor != ui_special_values$ALL_DEVELOPERS) {
      sql_base <- paste0(sql_base, " AND \"Developer\" = '", vendor, "'")
    }

    return(sql_base)
  })

  paged_capstat_values <- reactive({
    limit <- values_page_size
    offset <- (values_page_state() - 1) * values_page_size

    query_str <- paste0(
      "SELECT \"Developer\", \"FHIR Version\", field_value, \"Endpoints\" ",
      get_base_values_sql(),
      " ORDER BY \"Endpoints\" DESC LIMIT ", limit, " OFFSET ", offset
    )

    res <- tbl(db_connection, sql(query_str))

    res
  })

  capstat_total_count <- reactive({
    count_query <- paste0("SELECT COUNT(*) as count ", get_base_values_sql())
    res <- tbl(db_connection, sql(count_query)) %>% collect() %>% pull(count)
    res
  })

  output$capstat_values_table <- reactable::renderReactable({
    reactable(paged_capstat_values() %>% collect(),
                columns = list(
                  field_value = colDef(name = get_value_table_header())
                ),
                sortable = TRUE,
                searchable = FALSE,
                striped = TRUE,
                showSortIcon = TRUE,
                defaultPageSize = values_page_size
    )
  })

  # Gets the total number of endpoints that are using the currently selected field
  capstat_value_usage_summary <- reactive({
    req(sel_fhir_version(), sel_vendor(), sel_capstat_values())

    fhir_versions <- sel_fhir_version()
    vendor <- sel_vendor()
    field_name <- sel_capstat_values()

    fhir_versions_sql <- paste0("('", paste(fhir_versions, collapse = "', '"), "')")

    summary_query <- paste0("
      SELECT is_used AS used, SUM(count) AS count
      FROM capstat_usage_summary_mv
      WHERE field = '", field_name, "'
        AND \"FHIR Version\" IN ", fhir_versions_sql)

    if (vendor != ui_special_values$ALL_DEVELOPERS) {
      summary_query <- paste0(summary_query, " AND \"Developer\" = '", vendor, "'")
    }

    summary_query <- paste0(summary_query, " GROUP BY is_used")

    res <- tbl(db_connection, sql(summary_query)) %>% collect()

    res
  })

  # Data format for the Pie Chart
  percent_used_chart <- reactive({
    data <- capstat_value_usage_summary()

    # Default to 0 if row is missing
    yes <- if ("yes" %in% data$used) as.numeric(data$count[data$used == "yes"]) else 0
    no  <- if ("no" %in% data$used)  as.numeric(data$count[data$used == "no"])  else 0

    data.frame(group = c("Yes", "No"), value = c(yes, no))
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
