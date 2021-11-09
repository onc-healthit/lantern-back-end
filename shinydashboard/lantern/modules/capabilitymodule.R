# Capability Module
library(reactable)

capabilitymodule_UI <- function(id) {

  ns <- NS(id)

  tagList(
    fluidRow(
      column(width = 5,
             reactable::reactableOutput(ns("resource_op_table"))),
      column(width = 7,
             h4("Resource Count"),
             uiOutput(ns("resource_full_plot"))
      )
    ),
    h2("FHIR Implementation Guides"),
    p("This is the list of FHIR implementation guides reported by the capability statements from the endpoints."),
    fluidRow(
      column(width = 12,
             h4("Implementation Guide Count"),
             uiOutput(ns("implementation_plot"))
      )
    ),
  )
}

capabilitymodule <- function(  #nolint
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
    if (length(sel_operations()) >= 1) {
      # get the selected operation
      first_elem <- sel_operations()[1]
      res <- isolate(get_fhir_resource_by_op(db_connection, first_elem))
      # get the data for each selected operation and then bind them together
      # in one data frame
      loopList <- isolate(as.list(sel_operations()))
      count <- 0
      for (op in loopList) {
        if (count != 0) {
          item <- isolate(get_fhir_resource_by_op(db_connection, op))
          res <- rbind(res, item)
        }
        count <- count + 1
      }
    } else {
       # If no operation is selected, then just get the resource list since it's
      # too complicated to get it with the operation_resource field
      res <- isolate(app_data$endpoint_resource_types())
    }

    req(sel_fhir_version(), sel_vendor(), sel_resources())
    # Filter data by selected FHIR version
    if (sel_fhir_version() != ui_special_values$ALL_FHIR_VERSIONS) {
      res <- res %>% filter(fhir_version == sel_fhir_version())
    }
    # Then filter data by selected vendor
    if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
      res <- res %>% filter(vendor_name == sel_vendor())
    }
    if (length(sel_operations()) >= 1) {
      # e.g. type is a string, it equals ["Allergy", "Binary", etc.]
      # The type array is formatted as a string, so remove the []
      # then split the string by `, `
      # then remove the " " around each element in the array
      res <- res %>%
        mutate(type = str_sub(type, 2, -2)) %>%
        separate_rows(type, sep = ", ") %>%
        mutate(type = str_sub(type, 2, -2))
      # Then filter by the current resources selected
      if (!(ui_special_values$ALL_RESOURCES %in% sel_resources())) {
        res <- res %>% filter(type %in% sel_resources())
      }

      # Filter by the current operations selected, then group by and count the resource
      # per endpoint. If the count of the resource is equal to the number of selected
      # operations, then the resource exists for all operations and we keep that resource
      # Then group by and count all resources left
      res <- res %>%
        group_by(endpoint_id, fhir_version, type) %>%
        count() %>%
        filter(n == length(sel_operations())) %>%
        ungroup() %>%
        select(-n) %>%
        group_by(type, fhir_version) %>%
        count()
    } else {
      # Then filter by the current resources selected
      if (!(ui_special_values$ALL_RESOURCES %in% sel_resources())) {
        res <- res %>% filter(type %in% sel_resources())
      }
        res <- res %>%
        group_by(type, fhir_version) %>%
        count()
    }
    res
  })

  number_resources <- reactive({
    res <- isolate(app_data$endpoint_resource_types()) %>% distinct(type) %>% count()
    res
  })

  select_table_format <- reactive({
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
                )
              ),
              groupBy = "Resource",
              sortable = TRUE,
              searchable = TRUE,
              striped = TRUE,
              showSortIcon = TRUE,
              defaultPageSize = number_resources()$n -1,
              showPageSizeOptions = TRUE,
              pageSizeOptions = c(25, 50, 100, number_resources()$n -1)

     )
  })

  select_operations_count <- reactive({
    select_operations() %>%
    rename("Endpoints" = n, "Resource" = type)
  })

  implementation_count <- reactive({
    get_implementation_guide_count(selected_implementation_guide())
  })

  vendor <- reactive({
    sel_vendor()
  })

  selected_implementation_guide <- reactive({
    res <- isolate(app_data$implementation_guide())
    req(sel_fhir_version(), sel_vendor())
    if (sel_fhir_version() != ui_special_values$ALL_FHIR_VERSIONS) {
      res <- res %>% filter(fhir_version == sel_fhir_version())
    }
    if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
      res <- res %>% filter(vendor_name == sel_vendor())
    }
    res
  })

  # Default plot heights are not good for large number of bars, so base on
  # number of rows in the result
  plot_height <- reactive({
    max(nrow(select_operations()) * 25, 400)
  })

  plot_height_implementation <- reactive({
    max(nrow(implementation_count()) * 25, 400)
  })

  output$resource_plot <- renderUI({
    tagList(
      plotOutput(ns("resource_bar_plot"), height = plot_height())
    )
  })

  output$implementation_plot <- renderUI({
    if (nrow(implementation_count()) != 0) {
      tagList(
        plotOutput(ns("implementation_guide_plot"), height = plot_height_implementation())
      )
    }
    else {
      tagList(
        plotOutput(ns("implementation_guide_empty_plot"), height = plot_height_implementation())
      )
    }
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
    ggplot(select_operations_count(), aes(x = fct_rev(as.factor(Resource)), y = Endpoints, fill = get_fill(fhir_version))) +
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
      list(sel_fhir_version(), sel_vendor(), sel_resources(), sel_operations(), app_data$last_updated())
    })

  output$implementation_guide_plot <- renderCachedPlot({
    ggplot(implementation_count(), aes(x = fct_rev(as.factor(Implementation)), y = Endpoints, fill = fhir_version)) +
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
      list(sel_fhir_version(), sel_vendor(), app_data$last_updated())
    })

  output$implementation_guide_empty_plot <- renderPlot({
    ggplot(implementation_count()) +
    geom_col(width = 0.8) +
    labs(x = "Implementation Guides", y = "Number of Endpoints") +
    theme(axis.text.x = element_blank(),
    axis.text.y = element_blank(), axis.ticks = element_blank()) +
    annotate("text", label = "There are no Implementation guides supported by the endpoints\nthat pass the selected filtering criteia", x = 1, y = 2, size = 4.5, colour = "red", hjust = 0.5)
  })
}
