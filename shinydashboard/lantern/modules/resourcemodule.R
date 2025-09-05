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
                  textInput(ns("res_search_query"), "Search:", value = ""),
                  reactable::reactableOutput(ns("resource_op_table")),
                  fluidRow(
                    column(3, 
                      div(style = "display: flex; justify-content: ;", uiOutput(ns("resource_prev_button_ui"))
                      )
                    ),
                    column(6, 
                      div(style = "display: flex; justify-content: center; align-items: center; gap: 10px; margin-top: 8px;",
                            numericInput(ns("res_page_selector"), label = NULL, value = 1, min = 1, max = 1, step = 1, width = "80px"),
                            textOutput(ns("res_page_info"), inline = TRUE)
                        )
                    ),
                    column(3, 
                      div(style = "display: flex; justify-content: flex-end;", uiOutput(ns("resource_next_button_ui"))
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

  res_page_state <- reactiveVal(1)
  res_page_size <- 50

  # Add request tracking to prevent race conditions
  current_request_id <- reactiveVal(0)

  # Compute total pages based on filtered data
  res_total_pages <- reactive({
    req(sel_fhir_version(), sel_vendor(), sel_resources())

    count_query <- get_fhir_resource_by_op(db_connection, as.list(sel_operations()), as.list(sel_fhir_version()), as.list(sel_resources()), as.list(sel_vendor()), page_size = -1, offset = -1, search_query = input$res_search_query)
    total <- nrow(count_query)
    max(1, ceiling(total / res_page_size))
  })

  # Break the feedback loop with isolate()
  observe({
    new_page <- res_page_state()
    current_selector <- input$res_page_selector
    
    # Only update if different (prevents infinite loop)
    # Add safety check for current_selector to prevent crashes
    if (is.null(current_selector) || 
        is.na(current_selector) || 
        !is.numeric(current_selector) ||
        current_selector != new_page) {
      
      isolate({  # This is the key fix to break feedback loops!
        updateNumericInput(session, "res_page_selector", 
                          max = res_total_pages(),
                          value = new_page)
      })
    }
  })

  # Handle page selector input
  observeEvent(input$res_page_selector, {
    # Get current input value
    current_input <- input$res_page_selector
    
    # Check if input is valid (not NULL, not NA, and is a number)
    if (!is.null(current_input) && 
        !is.na(current_input) && 
        is.numeric(current_input) &&
        current_input > 0) {
      
      new_page <- max(1, min(current_input, res_total_pages()))
      
      # Only update page state if it's actually different
      if (new_page != res_page_state()) {
        res_page_state(new_page)
      }

      # Correct the input field if the user entered an invalid page number
      if (new_page != current_input) {
        updateNumericInput(session, "res_page_selector", value = new_page)
      }
    } else {
      # If input is invalid (empty, NA, or <= 0), reset to current page
      # Use a small delay to prevent immediate feedback loop
      invalidateLater(100)
      updateNumericInput(session, "res_page_selector", value = res_page_state())
    }
  }, ignoreInit = TRUE)  # Prevent firing on initialization

  # Handle next page button 
  observeEvent(input$res_next_page, {
    if (res_page_state() < res_total_pages()) {
      new_page <- res_page_state() + 1
      res_page_state(new_page)
    }
  })

  # Handle previous page button
  observeEvent(input$res_prev_page, {
    if (res_page_state() > 1) {
      new_page <- res_page_state() - 1
      res_page_state(new_page)
    }
  })

  # Reset to first page on any filter/search change 
  observeEvent(list(sel_fhir_version(), sel_vendor(), sel_resources(), sel_operations(), input$res_search_query), {
    res_page_state(1)
  })

  output$resource_prev_button_ui <- renderUI({
    if (res_page_state() > 1) {
      actionButton(ns("res_prev_page"), "Previous", icon = icon("arrow-left"))
    } else {
      NULL  # Hide the button
    }
  })

  output$resource_next_button_ui <- renderUI({
    if (res_page_state() < res_total_pages()) {
      actionButton(ns("res_next_page"), "Next", icon = icon("arrow-right"))
    } else {
      NULL
    }
  })

  output$res_page_info <- renderText({
    paste("of", res_total_pages())
  })

  # Original select_operations function unchanged (for plots)
  select_operations <- reactive({
    req(sel_fhir_version(), sel_vendor(), sel_resources())
    get_fhir_resource_by_op(db_connection, as.list(sel_operations()), as.list(sel_fhir_version()), as.list(sel_resources()), as.list(sel_vendor()))
  })

  # Paginated select_operations function for the table - WITH RACE CONDITION PROTECTION
  paginated_select_operations <- reactive({
    req(sel_fhir_version(), sel_vendor(), sel_resources())
    
    # Generate unique request ID 
    request_id <- isolate(current_request_id()) + 1
    current_request_id(request_id)
    
    result <- get_fhir_resource_by_op(db_connection, as.list(sel_operations()), as.list(sel_fhir_version()), as.list(sel_resources()), as.list(sel_vendor()), res_page_size, (res_page_state() - 1) * res_page_size, input$res_search_query)
    
    # Only return results if this is still the latest request
    # Use isolate() to check without creating reactive dependency
    if (request_id == isolate(current_request_id())) {
      # This is the latest request, process normally
      return(result)
    } else {
      # This request was superseded, return empty to avoid flicker
      return(data.frame())
    }
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
    op_table <- paginated_select_operations()  # Use paginated data
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
              striped = TRUE,
              showSortIcon = TRUE,
              defaultExpanded = FALSE,
              pagination = FALSE

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
