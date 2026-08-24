# Capability Module
library(reactable)

resourcemodule_UI <- function(id) {

  ns <- NS(id)

  tagList(
    div(class = "lantern-resource-panel",
      h3("Resource Count", style = "margin: 0 0 8px 0;"),
      tabsetPanel(id = "resource_tabset", type = "tabs",
            tabPanel("Bar Graph", uiOutput(ns("resource_full_plot"))),
            tabPanel("Table",
              tagList(
                fluidRow(class = "lantern-resource-pager-row",
                  column(width = 4, textInput(ns("res_search_query"), "Search:", value = "")),
                  column(width = 4, offset = 4,
                    div(style = "display: flex; justify-content: flex-end; align-items: center; gap: 8px;",
                          uiOutput(ns("resource_prev_button_ui")),
                          numericInput(ns("res_page_selector"), label = NULL, value = 1, min = 1, max = 1, step = 1, width = "70px"),
                          textOutput(ns("res_page_info"), inline = TRUE),
                          uiOutput(ns("resource_next_button_ui"))
                      )
                  )
                ),
                reactable::reactableOutput(ns("resource_op_table"))
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
  sel_operations,
  is_active
) {

  ns <- session$ns

  res_page_size <- 50

  # Add request tracking to prevent race conditions
  current_request_id <- reactiveVal(0)

  # Compute total pages via a real COUNT of the grouped rows instead of materializing the full
  # grouped result (page_size = -1, offset = -1) and nrow()-ing it.
  res_total_pages <- reactive({
    req(sel_fhir_version(), sel_vendor(), sel_resources(), is_active())

    total <- get_fhir_resource_by_op_count(db_connection, as.list(sel_operations()), as.list(sel_fhir_version()), as.list(sel_resources()), as.list(sel_vendor()), search_query = input$res_search_query)
    max(1, ceiling(total / res_page_size))
  })

  res_page_state <- create_pager(
    input, output, session,
    page_selector_id = "res_page_selector",
    prev_button_id = "res_prev_page",
    next_button_id = "res_next_page",
    prev_output_id = "resource_prev_button_ui",
    next_output_id = "resource_next_button_ui",
    total_pages = res_total_pages
  )

  # Reset to first page on any filter/search change
  observeEvent(list(sel_fhir_version(), sel_vendor(), sel_resources(), sel_operations(), input$res_search_query), {
    res_page_state(1)
  })

  output$res_page_info <- renderText({
    paste("of", res_total_pages())
  })

  # Original select_operations function unchanged (for plots)
  select_operations <- reactive({
    req(sel_fhir_version(), sel_vendor(), sel_resources(), is_active())
    get_fhir_resource_by_op(db_connection, as.list(sel_operations()), as.list(sel_fhir_version()), as.list(sel_resources()), as.list(sel_vendor()))
  })

  # Paginated select_operations function for the table - WITH RACE CONDITION PROTECTION
  paginated_select_operations <- reactive({
    req(sel_fhir_version(), sel_vendor(), sel_resources(), is_active())
    
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
      geom_text(aes(label = after_stat(y)), position = position_stack(vjust = 0.5)) +
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
      list(sel_fhir_version(), sel_vendor(), sel_resources(), sel_operations(), get_endpoint_last_updated(db_tables))
    })
}
