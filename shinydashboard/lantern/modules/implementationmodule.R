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

# Summarize count of implementation guides by implementation_guide, fhir_version
get_implementation_guide_count <- function(sel_fhir_version, sel_vendor) {
  # Build filtering conditions for the SQL query
  fhir_versions <- paste0("'", paste(sel_fhir_version, collapse = "','"), "'")
  vendor_filter <- if(!is.null(sel_vendor) && sel_vendor != ui_special_values$ALL_DEVELOPERS) {
    paste0("AND vendor_name = '", sel_vendor, "'")
  } else {
    ""
  }

  # Direct query to the materialized view
  query <- paste0("
      SELECT implementation_guide as \"Implementation\", fhir_version, COUNT(*) as \"Endpoints\" 
      FROM 
        (SELECT * 
        FROM mv_implementation_guide
        WHERE fhir_version IN (", fhir_versions, ")
        ", vendor_filter, ") T
      GROUP BY implementation_guide, fhir_version
    ")

  # Execute the query
  res <- dbGetQuery(db_connection, query) %>% collect()
  res <- res %>% mutate(Endpoints = as.integer(Endpoints)) %>% as_tibble()
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
    req(sel_fhir_version(), sel_vendor())
    get_implementation_guide_count(sel_fhir_version(), sel_vendor())
  })

  vendor <- reactive({
    sel_vendor()
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
      list(sel_fhir_version(), sel_vendor(), now("UTC"))
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
