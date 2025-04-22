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
              tabPanel("Table", reactable::reactableOutput(ns("resource_op_table")))
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

  select_operations <- reactive({
    req(sel_fhir_version(), sel_vendor(), sel_resources())
    get_fhir_resource_by_op(db_connection, as.list(sel_operations()), as.list(sel_fhir_version()), as.list(sel_resources()), as.list(sel_vendor()))
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
    op_table <- select_operations()
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
              searchable = TRUE,
              striped = TRUE,
              showSortIcon = TRUE,
              defaultPageSize = isolate(pageSizeNum()),
              showPageSizeOptions = TRUE,
              pageSizeOptions = c(25, 50, 100, number_resources()$n - 1)

     )
  })

  select_operations_count <- reactive({
    select_operations() %>%
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
