# Capability Module

capabilitymodule_UI <- function(id) {

  ns <- NS(id)

  tagList(
    h1("FHIR Resource Types"),
    p("This is the list of FHIR resource types reported by the capability statements from the endpoints. This reflects the most recent successful response only. Endpoints which are down, unreachable during the last query or have not returned a valid capability statement, are not included in this list."),
    fluidRow(
      column(width = 5,
             tableOutput(ns("resource_type_table"))),
      column(width = 7,
             h4("Resource Count"),
             uiOutput(ns("resource_plot"))
      )
    ),
    h1("FHIR Implementation Guides"),
    p("This is the list of FHIR implementation guides reported by the capability statements from the endpoints."),
    fluidRow(
      column(width = 12,
             h4("Implementation Guide Count"),
             uiOutput(ns("implementation_plot"))
      )
    ),
  )
}

capabilitymodule <- function(
  input,
  output,
  session,
  sel_fhir_version,
  sel_vendor,
  sel_resources
) {

  ns <- session$ns

  selected_fhir_endpoints <- reactive({
    res <- app_data$endpoint_resource_types
    req(sel_fhir_version(), sel_vendor(), sel_resources())
    if (sel_fhir_version() != ui_special_values$ALL_FHIR_VERSIONS) {
      res <- res %>% filter(fhir_version == sel_fhir_version())
    }
    if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
      res <- res %>% filter(vendor_name == sel_vendor())
    }

    if (!(ui_special_values$ALL_RESOURCES %in% sel_resources())) {
      list <- get_resource_list(res)
      req(sel_resources() %in% list)
      res <- res %>% filter(type %in% sel_resources())
    }

    res
  })

  endpoint_resource_count <- reactive({
    get_fhir_resource_count(selected_fhir_endpoints())
  })

  implementation_count <- reactive({
    get_implementation_guide_count(selected_implementation_guide())
  })

  output$resource_type_table <- renderTable(
    endpoint_resource_count() %>%
    rename("FHIR Version" = fhir_version)
  )

  vendor <- reactive({
    sel_vendor()
  })

  selected_implementation_guide <- reactive({
    res <- app_data$implementation_guide
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
    max(nrow(endpoint_resource_count()) * 25, 400)
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

  output$resource_bar_plot <- renderCachedPlot({
    ggplot(endpoint_resource_count(), aes(x = fct_rev(as.factor(Resource)), y = Endpoints, fill = fhir_version)) +
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
      list(sel_fhir_version(), sel_vendor(), sel_resources(), app_data$last_updated)
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
      list(sel_fhir_version(), sel_vendor(), app_data$last_updated)
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
