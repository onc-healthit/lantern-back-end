# Capability Module
library(treemapify)

capabilitymodule_UI <- function(id) {
  
  ns <- NS(id)
  
  tagList(
    h1("FHIR Resource Types"),
    p("This is the list of FHIR resource types reported by the capability statements from the endpoints."),
    fluidRow(
      column(width = 4, 
             tableOutput(ns("resource_type_table"))),
      column(width = 8,
             p("Resource Count"),
             plotOutput(ns("resource_bar_plot"), height = "3200px")
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
  endpoint_resource_types <- get_fhir_resources_tbl(db_tables)
  endpoint_resource_count <- endpoint_resource_types %>% group_by(type,fhir_version) %>% count() %>% rename(Resource=type,Endpoints=n)
  
  output$resource_type_table   <- renderTable(erc())
  erc <- reactive(selected_fhir_endpoints() %>% group_by(type,fhir_version) %>% count() %>% rename(Resource=type,Endpoints=n))
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
  
    output$resource_bar_plot <- renderPlot({
    ggplot(endpoint_resource_count,aes(x = fct_rev(as.factor(Resource)), y = Endpoints, fill = fhir_version)) +
    geom_col() +
    theme(text = element_text(size = 14)) +
      labs(x="FHIR Resource") +
    coord_flip()
  },height=3200)
  }
