# Fields Module

fieldsmodule_UI <- function(id) {
  
  ns <- NS(id)
  
  tagList(
    h1("FHIR Capability Statement Fields"),
    p("This is the list of fields included in the FHIR capability statements from the endpoints."),
    fluidRow(
      column(width=5,
             h4("Required Fields"),
             tableOutput(ns("capstat_fields_table_required")),
             h4("Optional Fields"),
             tableOutput(ns("capstat_fields_table_optional"))),
      column(width=7,
             h4("Supported Capability Statement Fields"),
             uiOutput(ns("fields_plot"))
      )
    )
  )
}

fieldsmodule <- function(
  input, 
  output, 
  session,
  sel_fhir_version,
  sel_vendor
){

  ns <- session$ns
  capstat_fields <- get_capstat_fields(db_connection)

  selected_fhir_endpoints <- reactive({
    res <- app_data$capstat_fields
    req(sel_fhir_version(), sel_vendor())
    # If the selected dropdown value for the fhir verison is not the default "All FHIR Versions", filter
    # the capability statement fields by which fhir verison they're associated with
    if (sel_fhir_version() != ui_special_values$ALL_FHIR_VERSIONS) {
      res <- res %>% filter(fhir_version == sel_fhir_version())
    }
    # Same as above but with the vendor dropdown
    if (sel_vendor() != ui_special_values$ALL_VENDORS) {
      res <- res %>% filter(vendor_name == sel_vendor())
    }
    res
  })
  
  capstat_field_count <- reactive({
    get_capstat_fields_count(selected_fhir_endpoints())
  })

  # Required Capability Statement fields that we are tracking
  required_fields <- c("status", "kind", "fhirVersion", "format", "patchFormat")

  # Table of the required fields
  output$capstat_fields_table_required <- renderTable(
    capstat_field_count() %>%
    filter(Fields %in% required_fields) %>%
    rename("FHIR Version" = fhir_version)
  )

  # Table of the optional fields
  output$capstat_fields_table_optional <- renderTable(
    capstat_field_count() %>%
    filter(!(Fields %in% required_fields)) %>%
    rename("FHIR Version"=fhir_version)
  )

  vendor <- reactive({
    sel_vendor()
  })

  plot_height <- reactive({
    max(nrow(capstat_field_count()) * 25, 400)
  })

  output$fields_plot <- renderUI({
    tagList(
      plotOutput(ns("fields_bar_plot"), height = plot_height())
    )
  })
  
  output$fields_bar_plot <- renderCachedPlot({
    ggplot(capstat_field_count(), aes(x = fct_rev(as.factor(Fields)), y = Endpoints, fill = fhir_version)) +
      geom_col(width = 0.8) +
      theme(legend.position = "top") +
      theme(text = element_text(size = 14)) +
      labs(x="", y = "Number of Endpoints", fill = "FHIR Version", title = vendor()) +
      coord_flip()
  },
    sizePolicy = sizeGrowthRatio(width = 400,
                                  height = 400,
                                  growthRate = 1.2),
    res = 72,
    cache = "app",
    cacheKeyExpr = {
      list(sel_fhir_version(), sel_vendor(), app_data$last_updated)
    }
  )
  
}