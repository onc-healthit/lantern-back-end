library(shiny)
library(shinydashboard)
library(readr)
library(scales)

capabilitystatementsize_UI <- function(id) {

  ns <- NS(id)
  tagList(
    fluidRow(
      column(width = 8,
        plotOutput(ns("cap_stat_size_plot")),
        htmlOutput(ns("notes_text"))
      ),
        column(width = 4,
        h2("CapabilityStatement / Conformance Resource Size"),
        tableOutput(ns("stats_table"))
      )
    )
  )
}

capabilitystatementsizemodule <- function(
    input,
    output,
    session,
    sel_fhir_version,
    sel_vendor
) {
  ns <- session$ns

  selected_fhir_endpoints <- reactive({
    # Get current filter values
    current_fhir <- sel_fhir_version()
    current_vendor <- sel_vendor()

    req(current_fhir, current_vendor)

    # Get filtered data from the materialized view function
    res <- get_cap_stat_sizes(
      db_connection,
      fhir_version = current_fhir,
      vendor = current_vendor
    )

    res
  })

  get_cap_stat_sizes <- function(db_connection, fhir_version = NULL, vendor = NULL) {
  # Start with base query
  query <- tbl(db_connection, "mv_capstat_sizes_tbl")

  # Apply filters in SQL before collecting data
  if (!is.null(fhir_version) && length(fhir_version) > 0) {
    query <- query %>% filter(fhir_version %in% !!fhir_version)
  }

  if (!is.null(vendor) && vendor != ui_special_values$ALL_DEVELOPERS) {
    query <- query %>% filter(vendor_name == !!vendor)
  }

  # Collect the data after applying filters in SQL
  result <- query %>%
    collect()
  
  return(result)
  }

  selected_fhir_endpoints_stats <- reactive({
    res <- summarise(selected_fhir_endpoints(), count = length(size), max = ifelse(all(is.na(size)), NA, max(size, na.rm = TRUE)), min = ifelse(all(is.na(size)), NA, min(size, na.rm = TRUE)), mean = mean(size), sd = sd(size))
    res
  })

  output$cap_stat_size_plot <- renderCachedPlot({
   ggplot(selected_fhir_endpoints(), aes(fhir_version, size)) +
   geom_boxplot(aes(fill = factor(vendor_name))) +
   scale_y_continuous(labels = scales::comma) +
   labs(fill = "Developer",
          x = "FHIR Version",
          y = "Size (Bytes)",
          title = "CapabilityStatement / Conformance Sizes by Developer and FHIR Version")
  },
    res = 72,
    cache = "app",
    cacheKeyExpr = {
      list(sel_fhir_version(), sel_vendor(), now("UTC"))
    })

  output$notes_text <- renderUI({
    note_info <- "<br>(1) The endpoints queried by Lantern are limited to Fast Healthcare Interoperability
               Resources (FHIR) endpoints published publicly by Certified API Developers in conformance with
               the ONC Cures Act Final Rule.<br>
               (2) This figure represents the sizes of the CapabilityStatement documents as they are stored in the Lantern,
               the sizes of the CapabilityStatements may vary slightly when downloaded directly from their sources."
    res <- paste("<div style='font-size: 16px;'><b>Notes:</b>", note_info, "</div>")
    HTML(res)
  })

  output$stats_table <- renderTable(
    selected_fhir_endpoints_stats() %>%
    rename("Max" = max, "Min" = min, "Count" = count, "Mean" = mean, "Standard Deviation" = sd)
  )

}
