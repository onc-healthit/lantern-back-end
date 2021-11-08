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
        h4("Capability Statement / Conformance Size Statistics"),
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
    res <- isolate(app_data$capstat_sizes_tbl())
    req(sel_fhir_version(), sel_vendor())
    # If the selected dropdown value for the fhir verison is not the default "All FHIR Versions", filter
    # the capability statement fields by which fhir verison they're associated with
    if (sel_fhir_version() != ui_special_values$ALL_FHIR_VERSIONS) {
      res <- res %>% filter(fhir_version == sel_fhir_version())
    }
    # Same as above but with the vendor dropdown
    if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
      res <- res %>% filter(vendor_name == sel_vendor())
    }
    res
  })

  selected_fhir_endpoints_stats <- reactive({
    res <- summarise(selected_fhir_endpoints(), count = length(size), max = ifelse(all(is.na(size)), NA, max(size, na.rm = T)), min = ifelse(all(is.na(size)), NA, min(size, na.rm = T)), mean = mean(size), sd = sd(size))
    res
  })

  output$cap_stat_size_plot <- renderCachedPlot({
   ggplot(selected_fhir_endpoints(), aes(fhir_version, size)) +
   geom_boxplot(aes(fill = factor(vendor_name))) +
   scale_y_continuous(labels = scales::comma) +
   labs(fill = "Developer",
          x = "FHIR Version",
          y = "Size (Bytes)",
          title = "Capability Statement / Conformance Sizes by Developer and FHIR Version")
  },
    res = 72,
    cache = "app",
    cacheKeyExpr = {
      list(sel_fhir_version(), sel_vendor(), app_data$last_updated())
    })

  output$notes_text <- renderUI({
    note_info <- "<br>(1) The endpoints queried by Lantern are limited to Fast Healthcare Interoperability
               Resources (FHIR) endpoints published publicly by Certified API Developers in conformance with
               the ONC Cures Act Final Rule, or discovered through the National Plan and Provider Enumeration
               System (NPPES).<br>
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
