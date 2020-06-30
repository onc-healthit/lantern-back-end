# Capability Module
library(treemapify)

capabilitymodule_UI <- function(id) {
  
  ns <- NS(id)
  
  tagList(
    h1("FHIR Resource Types"),
    p("This is the list of FHIR resource types reported by the capability statements from the endpoints."),
    fluidRow(
      column(width=5,
             tableOutput(ns("resource_type_table"))),
      column(width=7,
             h4("Resource Count"),
             plotOutput(ns("resource_bar_plot"))
      )
    )
  )
}

capabilitymodule <- function(
  input, 
  output, 
  session,
  sel_fhir_version,
  sel_vendor
){

  ns <- session$ns
  endpoint_resource_types <- get_fhir_resource_types(db_connection)
 
  selected_fhir_endpoints <- reactive({
    res <- endpoint_resource_types
    req(sel_fhir_version(), sel_vendor())
    if (sel_fhir_version() != ui_special_values$ALL_FHIR_VERSIONS) {
      res <- res %>% filter(fhir_version == sel_fhir_version())
    }
    if (sel_vendor() != ui_special_values$ALL_VENDORS) {
      res <- res %>% filter(vendor_name == sel_vendor())
    }
    res
  })
  
  endpoint_resource_count <- reactive({get_fhir_resource_count(selected_fhir_endpoints())})
 
  output$resource_type_table <- renderTable(endpoint_resource_count() %>% rename("FHIR Version"=fhir_version))
  
  output$resource_bar_plot <- renderPlot({
    df <- endpoint_resource_count()
    ggplot(df,aes(x = fct_rev(as.factor(Resource)), y = Endpoints, fill = fhir_version)) +
      geom_col(width = 0.8) +
      theme(legend.position="top") +
      theme(text = element_text(size = 14)) +
      labs(x="",fill="FHIR Version") +
      coord_flip()
  },height = function() {
    max(nrow(endpoint_resource_count()) * 24,100)
  })
  
}