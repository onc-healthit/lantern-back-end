# Fields Module
library(treemapify)

fieldsmodule_UI <- function(id) {
  
  ns <- NS(id)
  
  tagList(
    h1("FHIR Capability Statement Fields"),
    p("This is the list of FHIR capability statement fields include in the capability statements from the endpoints."),
    fluidRow(
      column(width=5,
             tableOutput(ns("capstat_fields_table"))),
      column(width=7,
             h4("Capability Statement Fields Count"),
             plotOutput(ns("fields_bar_plot"))
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

  #
  selected_fhir_endpoints <- reactive({
    res <- capstat_fields
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
  
  capstat_field_count <- reactive({get_capstat_fields_count(selected_fhir_endpoints())})
 
  output$capstat_fields_table <- renderTable(capstat_field_count() %>%
    rename("FHIR Version"=fhir_version))
  
  output$fields_bar_plot <- renderPlot({
    df <- capstat_field_count()
    ggplot(df,aes(x = fct_rev(as.factor(Fields)), y = Endpoints, fill = fhir_version)) +
      geom_col(width = 0.8) +
      theme(legend.position="top") +
      theme(text = element_text(size = 14)) +
      labs(x="",fill="FHIR Version") +
      coord_flip()
  },height = function() {
    max(nrow(capstat_field_count()) * 24,100)
  })
  
}