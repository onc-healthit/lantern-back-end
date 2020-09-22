# Values Module

valuesmodule_UI <- function(id) {
  ns <- NS(id)
  tagList(
    h1("Values of FHIR Capability Statement Fields"),
    p("This is the set of values from the endpoints for a given field included in the FHIR capability statements."),
    fluidRow(
      column(width = 5,
             h4("Field Values"),
             tableOutput(ns("capstat_values_table")),
            ),
      column(width = 7,
             h4("Given Values for Chosen Field"),
            #  tableOutput(ns("values_chart"))
            uiOutput(ns("values_chart"))
      )
    ),
  )
}

valuesmodule <- function(
  input,
  output,
  session,
  sel_fhir_version,
  sel_vendor,
  sel_capstat_values
) {

  ns <- session$ns

  selected_fhir_endpoints <- reactive({
    res <- app_data$capstat_values
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
    # Repeat with filtering fields to see values
    # if (all(sel_capstat_values() != "All fields")) {
      res <- res %>% group_by_at(vars("vendor_name", "fhir_version", sel_capstat_values())) %>%
        count() %>%
        rename(Count = n, Vendor = vendor_name, "FHIR Version" = fhir_version)
    # }
    res
  })

  capstat_values_list <- reactive({
    get_capstat_values_list(selected_fhir_endpoints())
  })

  # Table of the required fields
  output$capstat_values_table <- renderTable(
    capstat_values_list()
  )

  chart_group <- reactive({
    capstat_values_list() %>%
    ungroup() %>%
    # group_by_at(vars(sel_capstat_values())) %>%
    # count() %>%
    select(c(Count, sel_capstat_values())) %>%
    rename(value = Count, group = sel_capstat_values())
  })

  # output$values_chart <- renderTable(
  #   chart_group()
  # )

  # chart_values <- reactive({
  #   capstat_values_list() %>%
  #   select(Count)
  # })

  # df <- data.frame(
  #   group = chart_group(),
  #   value = chart_values()
  #   )
  # head(df)

  # bp <- ggplot(chart_group(), aes(x="", y=value, fill=group)) +
  #     geom_bar(width = 1, stat = "identity")

  output$values_chart <- renderUI({
    tagList(
      plotOutput(ns("values_chart_plot"), height = 800)
    )
  })

  # output$values_chart <-  renderCachedPlot({bp + coord_polar("y", start=0)},
   output$values_chart_plot <-  renderCachedPlot({
      ggplot(chart_group(), aes(x="", y=value, fill=group)) +
      geom_col(width = 0.8) +
      geom_bar(stat = "identity") +
      coord_polar("y", start=0)},
    sizePolicy = sizeGrowthRatio( width = 300,
                                  height = 400,
                                  growthRate = 1.2),
    res = 72,
    cache = "app",
    cacheKeyExpr = {
      list(sel_fhir_version(), sel_vendor(), sel_capstat_values())
    }
  )

}
