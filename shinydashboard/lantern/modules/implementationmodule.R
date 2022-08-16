# Capability Module
implementationmodule_UI <- function(id) {

  ns <- NS(id)

  tagList(
    h2("FHIR Implementation Guides"),
    p("This is the list of FHIR implementation guides reported by the CapabilityStatement / Conformance Resources from the endpoints."),
    fluidRow(
      column(width = 12,
             h3("Implementation Guide Count"),
             uiOutput(ns("implementation_plot"))
      )
    ),
  )
}

implementationmodule <- function(  #nolint
  input,
  output,
  session,
  sel_fhir_version,
  sel_vendor
) {

  ns <- session$ns

  implementation_count <- reactive({
    get_implementation_guide_count(selected_implementation_guide())
  })

  vendor <- reactive({
    sel_vendor()
  })

  selected_implementation_guide <- reactive({
    res <- isolate(app_data$implementation_guide())
    req(sel_fhir_version(), sel_vendor())

    res <- res %>% filter(fhir_version %in% sel_fhir_version())

    if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
      res <- res %>% filter(vendor_name == sel_vendor())
    }
    res
  })

  # Default plot heights are not good for large number of bars, so base on
  # number of rows in the result of implementation count
  plot_height_implementation <- reactive({
    max(nrow(implementation_count()) * 25, 400)
  })

  output$implementation_plot <- renderUI({
    if (nrow(implementation_count()) != 0) {
      tagList(
        plotOutput(ns("implementation_guide_plot"), height = plot_height_implementation())
      )
    } else {
      tagList(
        plotOutput(ns("implementation_guide_empty_plot"), height = plot_height_implementation())
      )
    }
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
    annotate("text", label = "There are no Implementation guides supported by the endpoints\nthat pass the selected filtering criteria", x = 1, y = 2, size = 4.5, colour = "red", hjust = 0.5)
  })
}
