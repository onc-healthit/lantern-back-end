# capstatdashboardmodule.R
# CapabilityStatement Dashboard Module — with reactive mock data
# ------------------------------------------------------------------------

library(shiny)
library(dplyr)
library(reactable)
library(ggplot2)
library(scales)

capstatdashboardmodule_UI <- function(id) {
  ns <- NS(id)
  tagList(
    tags$style(HTML("
      .capstat-dashboard {
        background-color: #f6f7fb;
        padding: 25px 30px;
        font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial;
      }
      .stats-bar {
        display: grid;
        grid-template-columns: repeat(4, 1fr);
        gap: 20px;
        margin-bottom: 25px;
      }
      .stat-card {
        background: white;
        border-radius: 12px;
        box-shadow: 0 1px 3px rgba(0,0,0,0.08);
        padding: 18px;
        text-align: center;
      }
      .stat-card h4 {
        color: #6c757d;
        font-size: 14px;
        margin-bottom: 6px;
        font-weight: 500;
      }
      .stat-card h2 {
        color: #007bff;
        font-size: 28px;
        font-weight: 600;
        margin: 0;
      }
      .filter-card {
        background: white;
        border-radius: 12px;
        padding: 18px 22px;
        margin-bottom: 25px;
        box-shadow: 0 1px 3px rgba(0,0,0,0.05);
      }
      .chart-section {
        display: grid;
        grid-template-columns: repeat(2, 1fr);
        gap: 25px;
        margin-bottom: 25px;
      }
      .chart-card {
        background: white;
        border-radius: 12px;
        padding: 18px;
        box-shadow: 0 1px 3px rgba(0,0,0,0.05);
      }
      .reactable-table {
        background: white;
        border-radius: 12px;
        padding: 15px;
        box-shadow: 0 1px 3px rgba(0,0,0,0.05);
      }
    ")),

    div(class = "capstat-dashboard",

        # --- Reactive Stats Bar ---
        div(class = "stats-bar",
            div(class = "stat-card", h4("Total Capability Statements"), h2(textOutput(ns("total_statements")))),
            div(class = "stat-card", h4("Unique Vendors"), h2(textOutput(ns("unique_vendors")))),
            div(class = "stat-card", h4("Average Fields per Statement"), h2(textOutput(ns("avg_fields")))),
            div(class = "stat-card", h4("Distinct FHIR Versions"), h2(textOutput(ns("distinct_fhir_versions"))))
        ),

        # --- Filter Card ---
        div(class = "filter-card",
            fluidRow(
              column(width = 4, selectInput(ns("fhir_select"), "FHIR Version", 
                                            choices = c("All", "3.0.1", "4.0.1", "4.1.0", "No Cap Stat"), 
                                            selected = "All")),
              column(width = 4, selectInput(ns("vendor_select"), "Vendor", 
                                            choices = c("All Vendors", "Epic", "Cerner", "Athenahealth", "NextGen", "Allscripts"),
                                            selected = "All Vendors"))
            )
        ),

        # --- Charts ---
        div(class = "chart-section",
            div(class = "chart-card",
                h3("Top CapabilityStatement Fields by Occurrence"),
                plotOutput(ns("top_fields_plot"), height = "300px")
            ),
            div(class = "chart-card",
                h3("Distribution of FHIR Versions"),
                plotOutput(ns("fhir_version_plot"), height = "300px")
            )
        ),

        # --- Vendor Table ---
        div(class = "reactable-table",
            h3("Vendors by Capability Statement Count"),
            reactableOutput(ns("developer_table"))
        )
    )
  )
}


capstatdashboardmodule <- function(input, output, session, fhir_version, vendor, field) {
  ns <- session$ns

  # === Mock Example Data (as if pulled from Lantern DB) ===
  all_data <- reactive({
    tibble(
      vendor = rep(c("Epic", "Cerner", "Athenahealth", "NextGen", "Allscripts", "eClinicalWorks"), each = 3),
      fhir_version = rep(c("3.0.1", "4.0.1", "4.1.0"), times = 6),
      fields_per_statement = sample(35:60, 18, replace = TRUE),
      statements = sample(800:2500, 18, replace = TRUE)
    )
  })

  # === Reactive Filtering ===
  filtered_data <- reactive({
    df <- all_data()
    if (input$fhir_select != "All") {
      df <- df %>% filter(fhir_version == input$fhir_select)
    }
    if (input$vendor_select != "All Vendors") {
      df <- df %>% filter(vendor == input$vendor_select)
    }
    df
  })

  # === Reactive KPIs ===
  output$total_statements <- renderText({
    format(sum(filtered_data()$statements), big.mark = ",")
  })

  output$unique_vendors <- renderText({
    n_distinct(filtered_data()$vendor)
  })

  output$avg_fields <- renderText({
    round(mean(filtered_data()$fields_per_statement), 1)
  })

  output$distinct_fhir_versions <- renderText({
    n_distinct(filtered_data()$fhir_version)
  })

  # === Plots ===
  output$top_fields_plot <- renderPlot({
    fields_data <- tibble(
      field = c("status", "kind", "format", "fhirVersion", "implementationGuide"),
      endpoints = c(12000, 11700, 9500, 8800, 7000)
    )
    ggplot(fields_data, aes(x = reorder(field, endpoints), y = endpoints)) +
      geom_col(fill = "#007bff") +
      coord_flip() +
      labs(x = "", y = "Statements", title = "Most Common CapabilityStatement Fields") +
      theme_minimal(base_size = 13)
  })

  output$fhir_version_plot <- renderPlot({
    fhir_data <- filtered_data() %>%
      group_by(fhir_version) %>%
      summarise(count = sum(statements)) %>%
      arrange(desc(count))

    ggplot(fhir_data, aes(x = "", y = count, fill = fhir_version)) +
      geom_bar(stat = "identity", width = 1, color = "white") +
      coord_polar("y") +
      theme_void() +
      labs(title = "FHIR Version Distribution", fill = "Version")
  })

  # === Vendor Table ===
  output$developer_table <- renderReactable({
    reactable(
      filtered_data() %>%
        group_by(vendor) %>%
        summarise(`Capability Statements` = sum(statements)) %>%
        arrange(desc(`Capability Statements`)),
      columns = list(
        vendor = colDef(name = "Vendor"),
        `Capability Statements` = colDef(align = "right")
      ),
      sortable = TRUE,
      striped = TRUE,
      defaultPageSize = 6,
      highlight = TRUE,
      bordered = TRUE
    )
  })
}