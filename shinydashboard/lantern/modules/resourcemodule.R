# Capability Module
library(reactable)

resourcemodule_UI <- function(id) {

  ns <- NS(id)

  tagList(
    fluidRow(
      h3("Resource Count", style = "margin-left:5px"),
      column(width = 12, style = "margin-right: 5px; margin-left: 5px;",
        tabsetPanel(id = "resource_tabset", type = "tabs",
              tabPanel("Bar Graph", uiOutput(ns("resource_full_plot"))),
              tabPanel("Table", 
                tagList(
                  fluidRow(
                    column(width = 6, textInput(ns("search_query"), "Search:", value = "")
                    )
                  ),
                  reactable::reactableOutput(ns("resource_op_table")),
                  fluidRow(
                    column(3, 
                      div(style = "display: flex; justify-content: flex-start;", uiOutput(ns("prev_button_ui"))
                      )
                    ),
                    column(6,
                      div(style = "display: flex; justify-content: center; align-items: center; gap: 10px; margin-top: 8px;",
                          numericInput(ns("page_selector"), label = NULL, value = 1, min = 1, max = 1, step = 1, width = "80px"),
                          textOutput(ns("page_info"), inline = TRUE)
                      )
                    ),
                    column(3, 
                      div(style = "display: flex; justify-content: flex-end;", uiOutput(ns("next_button_ui"))
                      )
                    )
                  )
                )
              )
        )
      )
    )
  )
}

get_fhir_resource_types <- function(db_connection) {
  tbl(db_connection, "mv_endpoint_resource_types") %>%
    collect()
}

resourcemodule <- function(  #nolint
  input,
  output,
  session,
  sel_fhir_version,
  sel_vendor,
  sel_resources,
  sel_operations
) {

  ns <- session$ns

  page_state <- reactiveVal(1)
  page_size <- 10

  # Calculate total pages based on filtered data
  total_pages <- reactive({
    total_records <- nrow(select_operations_filtered() %>% distinct(type, fhir_version))
    max(1, ceiling(total_records / page_size))
  })

  # Update page selector max when total pages change
  observe({
    updateNumericInput(session, ns("page_selector"), 
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
        updateNumericInput(session, ns("page_selector"), value = new_page)
      }
    }
  })

  # Handle next page button
  observeEvent(input$next_page, {
    if (page_state() < total_pages()) {
      new_page <- page_state() + 1
      page_state(new_page)
      updateNumericInput(session, ns("page_selector"), value = new_page)
    }
  })

  # Handle previous page button
  observeEvent(input$prev_page, {
    if (page_state() > 1) {
      new_page <- page_state() - 1
      page_state(new_page)
      updateNumericInput(session, ns("page_selector"), value = new_page)
    }
  })

  # Reset to first page on any filter/search change
  observeEvent(list(sel_fhir_version(), sel_vendor(), sel_resources(), sel_operations(), input$search_query), {
    page_state(1)
    updateNumericInput(session, ns("page_selector"), value = 1)
  })

  output$prev_button_ui <- renderUI({
    if (page_state() > 1) {
      actionButton(ns("prev_page"), "Previous", icon = icon("arrow-left"))
    } else {
      NULL  # Hide the button
    }
  })

  output$next_button_ui <- renderUI({
    if (page_state() < total_pages()) {
      actionButton(ns("next_page"), "Next", icon = icon("arrow-right"))
    } else {
      NULL  # Hide the button
    }
  })

  output$page_info <- renderText({
    paste("of", total_pages())
  })

  # Original select_operations function unchanged (for plots)
  select_operations <- reactive({
    req(sel_fhir_version(), sel_vendor(), sel_resources())
    get_fhir_resource_by_op(db_connection, as.list(sel_operations()), as.list(sel_fhir_version()), as.list(sel_resources()), as.list(sel_vendor()))
  })

  # New function for search in table
  select_operations_filtered <- reactive({
    req(sel_fhir_version(), sel_vendor(), sel_resources())
    
    base_data <- get_fhir_resource_by_op(db_connection, as.list(sel_operations()), as.list(sel_fhir_version()), as.list(sel_resources()), as.list(sel_vendor()))
    
    # Apply search filter if present
    if (trimws(input$search_query) != "") {
      keyword <- tolower(trimws(input$search_query))
      base_data <- base_data %>% filter(
        grepl(keyword, tolower(type), fixed = TRUE) |
        grepl(keyword, tolower(fhir_version), fixed = TRUE) |
        grepl(keyword, tolower(as.character(n)), fixed = TRUE)
      )
    }
    
    base_data
  })

  # Paginate using R slicing
  paged_operations <- reactive({
    all_data <- select_operations_filtered()
    start <- (page_state() - 1) * page_size + 1
    end <- min(nrow(all_data), page_state() * page_size)
    if (nrow(all_data) == 0 || start > nrow(all_data)) return(all_data[0, ])
    all_data[start:end, ]
  })

  number_resources <- reactive({
    # Query the MV directly for counting distinct resource types
    res <- tbl(db_connection, "mv_endpoint_resource_types") %>% 
      distinct(type) %>% 
      count() %>%
      collect()
    res
  })

  pageSizeNum <- reactiveVal(NULL)

  observe({
    page <- getReactableState("resource_op_table", "pageSize")
    pageSizeNum(page)
  })

  select_table_format <- reactive({
    if (is.null(pageSizeNum())) {
      pageSizeNum(50)
    }
    op_table <- paged_operations()  # Use paginated data
    if ("type" %in% colnames(op_table)) {
      op_table <- op_table %>% rename("Endpoints" = n, "Resource" = type, "FHIR Version" = fhir_version)
    }
    op_table
  })

   output$resource_op_table <- reactable::renderReactable({
     reactable(
              select_table_format(),
              columns = list(
                Endpoints = colDef(
                  aggregate = "sum",
                  format = list(aggregated = colFormat(prefix = "Total: "))
                ),
                Resource = colDef(
                  minWidth = 150
                ),
                "FHIR Version" = colDef(
                  align = "center"
                )
              ),
              groupBy = "Resource",
              sortable = TRUE,
              searchable = FALSE,  # Disabled built-in search
              striped = TRUE,
              showSortIcon = TRUE,
              defaultPageSize = 10,  # Fixed page size
              showPageSizeOptions = FALSE,  # Disabled page size options
              pageSizeOptions = NULL  # Removed page size options

     )
  })

  select_operations_count <- reactive({
    select_operations() %>%  # Use original data for plots
    rename("Endpoints" = n, "Resource" = type)  %>%
    mutate(Endpoints = as.numeric(Endpoints))
  })

  vendor <- reactive({
    sel_vendor()
  })

  # Default plot heights are not good for large number of bars, so base on
  # number of rows in the result
  plot_height <- reactive({
    max(nrow(select_operations()) * 25, 400)
  })

  output$resource_plot <- renderUI({
    tagList(
      plotOutput(ns("resource_bar_plot"), height = plot_height())
    )
  })

  output$resource_full_plot <- renderUI({
    if (nrow(select_operations_count()) != 0) {
      tagList(
        plotOutput(ns("resource_bar_plot"), height = plot_height())
      )
    }
  })

  get_fill <- function(fhir_version) {
    res <- fhir_version
    if (length(fhir_version) == 0) {
      res <- "No fill"
    }
    res
  }

  output$resource_bar_plot <- renderCachedPlot({
    ggplot(select_operations_count(), aes(x = fct_rev(as.factor(Resource)), y = Endpoints, fill = as.factor(fhir_version))) +
      geom_col(width = 0.8) +
      geom_text(aes(label = stat(y)), position = position_stack(vjust = 0.5)) +
      theme(legend.position = "top") +
      theme(text = element_text(size = 14)) +
      labs(x = "", y = "Number of Endpoints", fill = "FHIR Version", title = vendor()) +
      scale_y_continuous(sec.axis = sec_axis(~., name = "Number of Endpoints")) +
      coord_flip()
  },
    sizePolicy = sizeGrowthRatio(width = 400,
                                  height = 400,
                                  growthRate = 1.2),
    res = 72,
    cache = "app",
    cacheKeyExpr = {
      list(sel_fhir_version(), sel_vendor(), sel_resources(), sel_operations(), now("UTC"))
    })
}
