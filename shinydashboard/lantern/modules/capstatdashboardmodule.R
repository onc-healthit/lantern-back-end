# capstatdashboardmodule.R
# CapabilityStatement Dashboard Module — reactive version (with Grid/Table toggle)
# ------------------------------------------------------------------------

library(shiny)
library(dplyr)
library(reactable)
library(ggplot2)
library(scales)
library(stringr)

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
      .kpi-number {
        color: #007bff;
        font-size: 28px;
        font-weight: 600;
        margin: 0;
        line-height: 1.2;
        text-shadow: 0 0.5px 0 rgba(0,0,0,0.05);
        letter-spacing: 0.3px;
      }
      .filter-card {
        background: white;
        border-radius: 12px;
        padding: 18px 22px;
        margin-bottom: 25px;
        box-shadow: 0 1px 3px rgba(0,0,0,0.05);
      }
      .filter-card h3 {
        color: #007bff;
        font-weight: 600;
        font-size: 18px;
        margin-top: 0;
        margin-bottom: 12px;
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
      .view-endpoints-btn {
        background-color: #007bff;
        color: white;
        border: none;
        border-radius: 8px;
        font-size: 16px;
        font-weight: 500;
        padding: 12px 24px;
        margin-top: 10px;
        margin-bottom: 30px;
        transition: background-color 0.2s ease;
      }
      .view-endpoints-btn:hover {
        background-color: #0069d9;
      }

      /* Updated Grid Card Styles with Hover Outline and Pagination */
      .endpoint-grid {
        display: grid;
        grid-template-columns: repeat(3, 1fr);
        gap: 20px;
      }
      .endpoint-card {
        background: white;
        border-radius: 12px;
        border: 1.5px solid transparent;
        box-shadow: 0 1px 3px rgba(0,0,0,0.08);
        padding: 16px 20px;
        transition: transform 0.15s ease, box-shadow 0.15s ease, border-color 0.15s ease;
      }
      .endpoint-card:hover {
        transform: translateY(-3px);
        border-color: #007bff;
        box-shadow: 0 4px 10px rgba(0,0,0,0.15);
      }
      .endpoint-card h4 {
        margin-top: 0;
        margin-bottom: 6px;
        color: #1B5A7F;
        font-weight: 700;
        font-size: 16px;
      }
      .endpoint-card .subtitle-link a {
        color: #007bff !important;
        text-decoration: underline;
        font-size: 13px;
      }
      .endpoint-card .kv {
        margin: 8px 0;
        font-size: 13px;
        color: #333;
      }
      .endpoint-card .kv b {
        color: #6c757d;
        font-weight: 600;
        margin-right: 6px;
      }
      .endpoint-card .metrics {
        margin-top: 10px;
        font-size: 13px;
        color: #444;
      }
      .endpoint-card .metrics .metric {
        margin-right: 10px;
        display: inline-block;
      }
      .endpoint-card hr {
        border: none;
        border-top: 1px solid #eee;
        margin: 10px 0 12px 0;
      }
      .grid-pagination {
        text-align: center;
        margin-top: 20px;
      }
      .grid-pagination button {
        background-color: #007bff;
        color: white;
        border: none;
        border-radius: 6px;
        padding: 8px 14px;
        margin: 0 5px;
        font-size: 13px;
        transition: background-color 0.2s ease;
      }
      .grid-pagination button:hover {
        background-color: #0069d9;
      }
      .grid-pagination span {
        margin: 0 10px;
        font-weight: 500;
      }
    ")),

    div(class = "capstat-dashboard",

        # --- KPIs ---
        div(class = "stats-bar",
            div(class = "stat-card",
                h4("Total Capability Statements"),
                div(class = "kpi-number", textOutput(ns("total_statements"), container = span))
            ),
            div(class = "stat-card",
                h4("Unique Vendors"),
                div(class = "kpi-number", textOutput(ns("unique_vendors"), container = span))
            ),
            div(class = "stat-card",
                h4("Average Fields per Statement"),
                div(class = "kpi-number", textOutput(ns("avg_fields"), container = span))
            ),
            div(class = "stat-card",
                h4("Distinct FHIR Versions"),
                div(class = "kpi-number", textOutput(ns("distinct_fhir_versions"), container = span))
            )
        ),

        # --- Filters ---
        div(class = "filter-card",
            h3("Filters"),
            fluidRow(
              column(width = 4, selectInput(ns("fhir_select"), "FHIR Version",
                                            choices = c("All", "3.0.1", "4.0.1", "4.1.0", "No Cap Stat"),
                                            selected = "All")),
              column(width = 4, selectInput(ns("vendor_select"), "Vendor",
                                            choices = c("All Vendors", "Epic", "Cerner", "Athenahealth", "NextGen", "Allscripts"),
                                            selected = "All Vendors"))
            ),
            div(style = "text-align:right; margin-top:10px;",
                actionButton(ns("view_endpoints_btn"), "View All Capability Statements →",
                             class = "view-endpoints-btn"))
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

        # --- Vendor Summary Table ---
        div(class = "reactable-table",
            h3("Vendors by Capability Statement Count"),
            reactableOutput(ns("developer_table"))
        ),

        # --- Endpoint Grid/Table Section ---
        div(class = "reactable-table",
            h3("CapabilityStatement Details by Endpoint"),
            div(style = "text-align:right; margin-bottom:10px;",
                actionButton(ns("toggle_view"), "Switch to Grid View", class = "view-endpoints-btn")
            ),
            uiOutput(ns("grid_view_container")),
            reactableOutput(ns("grid_table"))
        )
    )
  )
}


capstatdashboardmodule <- function(input, output, session, fhir_version, vendor, field) {
  ns <- session$ns
  message("capstatdashboardmodule: server started")

  # === Static Mock Example Data (aggregate-level) ===
  all_data <- tibble(
    vendor = rep(c("Epic", "Cerner", "Athenahealth", "NextGen", "Allscripts", "eClinicalWorks"), each = 3),
    fhir_version = rep(c("3.0.1", "4.0.1", "4.1.0"), times = 6),
    fields_per_statement = c(55, 48, 52, 43, 47, 50, 58, 54, 49, 46, 44, 53, 56, 57, 45, 52, 50, 48),
    statements = c(2100, 1900, 1500, 2300, 1800, 1600, 1700, 1400, 900, 1250, 1100, 1000, 950, 870, 890, 720, 810, 650)
  )

  # === Reactive Filtering ===
  filtered_data <- reactive({
    req(input$fhir_select, input$vendor_select)
    df <- all_data
    if (input$fhir_select != "All") {
      df <- df %>% filter(fhir_version == input$fhir_select)
    }
    if (input$vendor_select != "All Vendors") {
      df <- df %>% filter(vendor == input$vendor_select)
    }
    df
  })

  # === KPIs ===
  output$total_statements <- renderText({
    df <- filtered_data()
    if (nrow(df) == 0) return("0")
    format(sum(df$statements, na.rm = TRUE), big.mark = ",")
  })

  output$unique_vendors <- renderText({
    df <- filtered_data()
    if (nrow(df) == 0) return("0")
    n_distinct(df$vendor)
  })

  output$avg_fields <- renderText({
    df <- filtered_data()
    if (nrow(df) == 0) return("0.0")
    round(mean(df$fields_per_statement, na.rm = TRUE), 1)
  })

  output$distinct_fhir_versions <- renderText({
    df <- filtered_data()
    if (nrow(df) == 0) return("0")
    n_distinct(df$fhir_version)
  })

  # === Top Fields Plot (reactive to filters, single color) ===
  output$top_fields_plot <- renderPlot({
    df <- filtered_data()
    fields <- c("status","kind","format","fhirVersion","implementationGuide",
                "publisher","date","jurisdiction","version","description","contact","software.name")
    set.seed(42 + nrow(df))
    fields_data <- tibble(
      field = fields,
      statements = round(runif(length(fields), 0.6, 1.0) * sum(df$statements, na.rm = TRUE) / length(fields))
    )

    if (nrow(fields_data) == 0 || sum(fields_data$statements, na.rm = TRUE) == 0) {
      ggplot() + theme_void() +
        labs(title = "Most Common CapabilityStatement Fields") +
        annotate("text", x = 0, y = 0, label = "No data for current filters")
    } else {
      ggplot(fields_data, aes(x = reorder(field, statements), y = statements)) +
        geom_col(fill = "#007bff") +
        coord_flip() +
        labs(x = "", y = "Statements", title = "Most Common CapabilityStatement Fields") +
        theme_minimal(base_size = 13)
    }
  })

  # === FHIR Version Pie (reactive) ===
  output$fhir_version_plot <- renderPlot({
    df <- filtered_data() %>%
      group_by(fhir_version) %>%
      summarise(count = sum(statements), .groups = "drop")

    if (nrow(df) == 0) {
      ggplot() + theme_void() +
        labs(title = "FHIR Version Distribution") +
        annotate("text", x = 0, y = 0, label = "No data for current filters")
    } else {
      ggplot(df, aes(x = "", y = count, fill = fhir_version)) +
        geom_bar(stat = "identity", width = 1, color = "white") +
        coord_polar("y") +
        theme_void() +
        labs(title = "FHIR Version Distribution", fill = "Version")
    }
  })

  # === Vendor Summary Table ===
  output$developer_table <- renderReactable({
    df <- filtered_data() %>%
      group_by(vendor) %>%
      summarise(`Capability Statements` = sum(statements), .groups = "drop") %>%
      arrange(desc(`Capability Statements`))

    reactable(
      df,
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

  # === Endpoint Grid/Table Data (mock, derived from filters) ===
  # Create deterministic, filter-aware endpoint rows from aggregate slice
  vendor_grid_data <- reactive({
    df <- filtered_data()

    if (nrow(df) == 0) {
      return(tibble(
        endpoint_name = character(0),
        capability_url = character(0),
        developer = character(0),
        source = character(0),
        fhir_version = character(0),
        status = character(0),
        instance = character(0),
        resources = integer(0),
        search_params = integer(0),
        uptime = integer(0),
        operations = integer(0)
      ))
    }

    # For each (vendor, version) row, create a couple of endpoints proportionally
    rows <- purrr::pmap_dfr(df, function(vendor, fhir_version, fields_per_statement, statements) {
      n <- max(1, round(statements / 1200))   # rough scaling for demo
      tibble(
        endpoint_name = paste(vendor, "Endpoint", seq_len(n)),
        capability_url = paste0("https://api.", str_to_lower(gsub('[^A-Za-z]', '', vendor)), ".com/", fhir_version, "/metadata"),
        developer = vendor,
        source = sample(c("ONC", "CHPL", "Manual"), n, TRUE),
        fhir_version = fhir_version,
        status = sample(c("Active", "Inactive"), n, TRUE, prob = c(0.85, 0.15)),
        instance = sample(c("Yes", "No"), n, TRUE, prob = c(0.7, 0.3)),
        resources = round(runif(n, 30, 95)),
        search_params = round(runif(n, 50, 160)),
        uptime = round(runif(n, 90, 100)),
        operations = round(runif(n, 2, 12))
      )
    })

    # Ensure at least some rows
    if (nrow(rows) == 0) {
      rows <- tibble(
        endpoint_name = "Sample Endpoint",
        capability_url = "https://api.sample.com/fhir/metadata",
        developer = "Sample Vendor",
        source = "Manual",
        fhir_version = "4.0.1",
        status = "Active",
        instance = "Yes",
        resources = 56,
        search_params = 120,
        uptime = 98,
        operations = 8
      )
    }
    rows
  })

  # === Toggle View State ===
  view_mode <- reactiveVal("table")
  grid_page <- reactiveVal(1)
  page_size <- 9  # show 3x3 cards per page

  observeEvent(input$toggle_view, {
    new_mode <- ifelse(view_mode() == "table", "grid", "table")
    view_mode(new_mode)
    updateActionButton(session, "toggle_view",
                      label = ifelse(new_mode == "table", "Switch to Grid View", "Switch to Table View"))
    grid_page(1)  # reset pagination when toggled
  })

  # === Pagination Handlers ===
  observeEvent(input$grid_next, {
    grid_page(grid_page() + 1)
  })
  observeEvent(input$grid_prev, {
    if (grid_page() > 1) grid_page(grid_page() - 1)
  })

  # === Table View ===
  output$grid_table <- renderReactable({
    req(view_mode() == "table")
    d <- vendor_grid_data()
    reactable(
      d,
      columns = list(
        endpoint_name = colDef(name = "Endpoint"),
        capability_url = colDef(name = "Capability URL", cell = function(value) {
          sprintf("<a href='%s' target='_blank'>%s</a>", value, value)
        }, html = TRUE),
        developer = colDef(name = "Developer"),
        source = colDef(name = "Source"),
        fhir_version = colDef(name = "FHIR Version"),
        status = colDef(name = "Status"),
        instance = colDef(name = "Instance"),
        resources = colDef(name = "Resources", align = "right"),
        search_params = colDef(name = "Search Params", align = "right"),
        uptime = colDef(name = "Uptime (%)", align = "right"),
        operations = colDef(name = "Operations", align = "right")
      ),
      sortable = TRUE,
      searchable = TRUE,
      striped = TRUE,
      defaultPageSize = 8,
      highlight = TRUE,
      bordered = TRUE
    )
  })

  # === Grid View (Paginated) ===
  output$grid_view_container <- renderUI({
    req(view_mode() == "grid")
    d <- vendor_grid_data()

    # Calculate pages
    total_pages <- ceiling(nrow(d) / page_size)
    current_page <- grid_page()
    start_row <- ((current_page - 1) * page_size) + 1
    end_row <- min(start_row + page_size - 1, nrow(d))
    d_page <- d[start_row:end_row, , drop = FALSE]

    if (nrow(d_page) == 0) {
      return(div(em("No endpoints for current filters.")))
    }

    tagList(
      div(
        class = "endpoint-grid",
        lapply(seq_len(nrow(d_page)), function(i) {
          row <- d_page[i, ]
          div(class = "endpoint-card",
              h4(row$endpoint_name),
              div(class = "subtitle-link",
                  tags$a(href = row$capability_url, target = "_blank", "View CapabilityStatement")
              ),
              tags$hr(),
              div(class = "kv", tags$b("Developer:"), row$developer),
              div(class = "kv", tags$b("Source:"), row$source),
              div(class = "kv", tags$b("FHIR Version:"), row$fhir_version),
              div(class = "kv", tags$b("Status:"), row$status),
              div(class = "kv", tags$b("Instance:"), row$instance),
              div(class = "metrics",
                  span(class = "metric", tags$b("Resources:"), row$resources), " ",
                  span(class = "metric", tags$b("Search Params:"), row$search_params), " ",
                  span(class = "metric", tags$b("Uptime:"), paste0(row$uptime, "%")), " ",
                  span(class = "metric", tags$b("Ops:"), row$operations)
              )
          )
        })
      ),
      div(class = "grid-pagination",
          actionButton(ns("grid_prev"), "← Prev"),
          span(paste("Page", current_page, "of", total_pages)),
          actionButton(ns("grid_next"), "Next →")
      )
    )
  })

    # --- Placeholder navigation action ---
  observeEvent(input$view_endpoints_btn, {
    message("Navigate to Endpoints Grid (mock action for now)")
  })
}