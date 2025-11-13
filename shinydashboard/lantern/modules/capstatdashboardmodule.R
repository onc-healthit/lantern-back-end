# capstatdashboardmodule.R
# CapabilityStatement Dashboard Module — DB-backed KPIs + Mock Endpoint Grid/Table
# ------------------------------------------------------------------------

library(shiny)
library(dplyr)
library(dbplyr)
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
        text-shadow: 0 0.5px 0 rgba(0, 0, 0, 0.05);
        letter-spacing: 0.3px;
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
      }
      .grid-pagination button:hover {
        background-color: #0069d9;
      }
      .grid-pagination span {
        margin: 0 10px;
        font-weight: 500;
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
      }
    ")),

    div(class = "capstat-dashboard",

        # --- KPIs ---
        div(class = "stats-bar",
            div(class = "stat-card",
                h4("Total Capability Statements"),
                div(class = "kpi-number", textOutput(ns("kpi_total_capstats"), container = span))
            ),
            div(class = "stat-card",
                h4("Unique Vendors"),
                div(class = "kpi-number", textOutput(ns("kpi_unique_vendors"), container = span))
            ),
            div(class = "stat-card",
                h4("Average Fields per Endpoint"),
                div(class = "kpi-number", textOutput(ns("kpi_avg_fields"), container = span))
            ),
            div(class = "stat-card",
                h4("Distinct FHIR Versions"),
                div(class = "kpi-number", textOutput(ns("kpi_distinct_versions"), container = span))
            )
        ),

        # --- Charts ---
        div(class = "chart-section",
            div(class = "chart-card",
                h3("Top CapabilityStatement Fields"),
                plotOutput(ns("top_fields_plot"), height = "300px")
            ),
            div(class = "chart-card",
                h3("FHIR Version Distribution"),
                plotOutput(ns("fhir_version_plot"), height = "300px")
            )
        ),

        # --- Vendor Summary ---
        div(class = "reactable-table",
            h3("Vendors by Endpoint Count (with CapabilityStatement)"),
            reactableOutput(ns("vendor_summary_tbl"))
        ),

        # --- Endpoint Grid/Table Section (Mock Data) ---
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

  # -------------------------
  # GLOBAL FILTERS (passed from dashboard)
  # -------------------------
  eff_vendor <- reactive({
    v <- vendor()
    if (is.null(v) || identical(v, "All Developers") || identical(v, "")) NULL else v
  })
  eff_versions <- reactive({
    fv <- fhir_version()
    if (is.null(fv) || length(fv) == 0) NULL else fv
  })

  # -------------------------
  # KPI DATA
  # -------------------------
  # ---- Scope driven by global filters ----
  endpoints_scope <- reactive({
    q <- tbl(db_connection, "selected_fhir_endpoints_mv") %>%
      filter(cap_stat_exists == "true")   # character "true" in your MV

    # apply global vendor filter only when a real vendor is chosen
    v <- vendor()
    if (!is.null(v) && length(v) == 1 && !identical(v, "All Developers") && nzchar(v)) {
      q <- q %>% filter(vendor_name == !!v)
    }

    # apply global fhir_version filter only when a non-empty vector is provided
    fv <- fhir_version()
    if (!is.null(fv) && length(fv) > 0) {
      q <- q %>% filter(fhir_version %in% !!fv)
    }
    q
  })

  # small helper that turns empty/NULL into 0
  pull_scalar_or_zero <- function(x) {
    val <- tryCatch({ x %>% collect() %>% pull(1) }, error = function(e) numeric(0))
    if (is.null(val) || length(val) == 0 || is.na(val)) 0 else val
  }

  output$kpi_total_capstats <- renderText({
    n <- pull_scalar_or_zero(endpoints_scope() %>% summarise(n = n_distinct(id)))
    format(n, big.mark = ",")
  })

  output$kpi_unique_vendors <- renderText({
    n <- pull_scalar_or_zero(endpoints_scope() %>% summarise(n = n_distinct(vendor_name)))
    format(n, big.mark = ",")
  })

  output$kpi_distinct_versions <- renderText({
    n <- pull_scalar_or_zero(endpoints_scope() %>% summarise(n = n_distinct(fhir_version)))
    format(n, big.mark = ",")
  })

  output$kpi_avg_fields <- renderText({
    req(db_connection)

    v <- vendor()
    fv <- fhir_version()

    vendor_filter <- if (!is.null(v) && !identical(v, "All Developers") && nzchar(v)) {
      glue::glue_sql("AND vendor_name = {v}", .con = db_connection)
    } else {
      DBI::SQL("")
    }

    fhir_filter <- if (!is.null(fv) && length(fv) > 0) {
      glue::glue_sql("AND fhir_version IN ({vals*})", vals = fv, .con = db_connection)
    } else {
      DBI::SQL("")
    }

    sql_query <- glue::glue_sql("
      SELECT ROUND(AVG(field_count), 1) AS avg_fields_per_endpoint
      FROM (
        SELECT endpoint_id, COUNT(DISTINCT field) AS field_count
        FROM mv_capstat_fields
        WHERE exist = 'true' 
          AND extension = 'false'
          {vendor_filter}
          {fhir_filter}
        GROUP BY endpoint_id
      ) sub;
    ", .con = db_connection)

    df <- DBI::dbGetQuery(db_connection, sql_query)

    if (nrow(df) == 0 || is.na(df$avg_fields_per_endpoint)) {
      return("0.0")
    } else {
      df$avg_fields_per_endpoint
    }
  })

  # -------------------------
  # CHARTS
  # -------------------------
  output$top_fields_plot <- renderPlot({
    req(db_connection)

    # Build filter clauses dynamically
    vendor_filter <- if (!is.null(vendor()) && vendor() != "All Developers") {
      glue::glue_sql("AND vendor_name = {vendor()}", .con = db_connection)
    } else {
      DBI::SQL("")
    }

    version_filter <- if (!is.null(fhir_version()) && length(fhir_version()) > 0) {
      glue::glue_sql("AND fhir_version IN ({fhir_version()*})", .con = db_connection)
    } else {
      DBI::SQL("")
    }

    # Compose SQL query
    sql_query <- glue::glue_sql("
      SELECT field, COUNT(DISTINCT endpoint_id) AS endpoint_count
      FROM mv_capstat_fields
      WHERE exist = 'true' AND extension = 'false'
      {vendor_filter}
      {version_filter}
      GROUP BY field
      ORDER BY endpoint_count DESC
      LIMIT 15;
    ", .con = db_connection)

    # Execute query
    fields_agg <- DBI::dbGetQuery(db_connection, sql_query)

    # Plot results
    if (nrow(fields_agg) == 0) {
      ggplot() +
        theme_void() +
        labs(title = 'Top CapabilityStatement Fields') +
        annotate('text', x = 0, y = 0, label = 'No data for current filters')
    } else {
      ggplot(fields_agg, aes(x = reorder(field, endpoint_count), y = endpoint_count)) +
        geom_col(fill = "#007bff") +
        coord_flip() +
        theme_minimal(base_size = 13) +
        labs(x = "", y = "Endpoints", title = "Top CapabilityStatement Fields")
    }
  })

  output$fhir_version_plot <- renderPlot({
    req(db_connection)

    # Dynamic SQL filters
    vendor_filter <- if (!is.null(vendor()) && vendor() != "All Developers") {
      glue::glue_sql("AND vendor_name = {vendor()}", .con = db_connection)
    } else {
      DBI::SQL("")
    }

    version_filter <- if (!is.null(fhir_version()) && length(fhir_version()) > 0) {
      glue::glue_sql("AND fhir_version IN ({fhir_version()*})", .con = db_connection)
    } else {
      DBI::SQL("")
    }

    # Compose query
    sql_query <- glue::glue_sql("
      SELECT fhir_version, COUNT(DISTINCT endpoint_id) AS endpoint_count
      FROM mv_capstat_fields
      WHERE exist = 'true' AND extension = 'false'
      {vendor_filter}
      {version_filter}
      GROUP BY fhir_version
      ORDER BY endpoint_count DESC;
    ", .con = db_connection)

    # Execute query
    ver_agg <- DBI::dbGetQuery(db_connection, sql_query)

    # Plot result
    if (nrow(ver_agg) == 0) {
      ggplot() +
        theme_void() +
        labs(title = "FHIR Version Distribution") +
        annotate("text", x = 0, y = 0, label = "No data for current filters")
    } else {
      ggplot(ver_agg, aes(x = "", y = endpoint_count, fill = fhir_version)) +
        geom_bar(stat = "identity", width = 1, color = "white") +
        coord_polar("y") +
        theme_void() +
        labs(title = "FHIR Version Distribution", fill = "Version")
    }
  })

  output$vendor_summary_tbl <- renderReactable({
    req(db_connection)

    # Build reactive filters
    vendor_filter <- if (!is.null(vendor()) && vendor() != "All Developers") {
      glue::glue_sql("AND vendor_name = {vendor()}", .con = db_connection)
    } else {
      DBI::SQL("")
    }

    version_filter <- if (!is.null(fhir_version()) && length(fhir_version()) > 0) {
      glue::glue_sql("AND fhir_version IN ({fhir_version()*})", .con = db_connection)
    } else {
      DBI::SQL("")
    }

    # Compose SQL query
    sql_query <- glue::glue_sql("
      SELECT vendor_name,
            COUNT(DISTINCT endpoint_id) AS endpoints_with_capstat
      FROM mv_capstat_fields
      WHERE exist = 'true'
        AND extension = 'false'
        {vendor_filter}
        {version_filter}
      GROUP BY vendor_name
      ORDER BY endpoints_with_capstat DESC;
    ", .con = db_connection)

    # Execute
    vdf <- DBI::dbGetQuery(db_connection, sql_query)

    # Render table
    reactable(
      vdf,
      columns = list(
        vendor_name = colDef(name = "Vendor"),
        endpoints_with_capstat = colDef(name = "Endpoints with CapStat", align = "right")
      ),
      sortable = TRUE,
      striped = TRUE,
      defaultPageSize = 8,
      highlight = TRUE,
      bordered = TRUE
    )
  })

  # --- Reactive SQL data fetch using mv_endpoint_capstat_summary ---
  endpoint_data <- reactive({
    req(db_connection)

    # Dynamic SQL filters
    vendor_filter <- if (!is.null(vendor()) && vendor() != "All Developers") {
      glue::glue_sql("AND e.vendor_name = {vendor()}", .con = db_connection)
    } else {
      DBI::SQL("")
    }

    version_filter <- if (!is.null(fhir_version()) && length(fhir_version()) > 0) {
      glue::glue_sql("AND e.fhir_version IN ({fhir_version()*})", .con = db_connection)
    } else {
      DBI::SQL("")
    }

    # Compose SQL to join selected_fhir_endpoints_mv with mv_endpoint_capstat_summary
    sql_query <- glue::glue_sql("
      SELECT 
          e.id AS endpoint_id,
          e.endpoint_names,
          e.url,
          e.vendor_name,
          e.list_source,
          e.fhir_version,
          e.status,
          COALESCE(m.resources, 0) AS resources,
          COALESCE(m.search_params, 0) AS search_params,
          COALESCE(m.interactions, 0) AS interactions
      FROM selected_fhir_endpoints_mv e
      LEFT JOIN mv_endpoint_capstat_summary m ON e.url = m.url
      WHERE e.cap_stat_exists = 'true'
        {vendor_filter}
        {version_filter}
      ORDER BY e.vendor_name, e.endpoint_names;
    ", .con = db_connection)

    df <- DBI::dbGetQuery(db_connection, sql_query)
    if (nrow(df) == 0) return(tibble())
    df
  })

  # -------------------------
  # ENDPOINT TABLE + GRID
  # -------------------------

  # # -------------------------
  # # MOCK GRID/TABLE SECTION
  # # -------------------------
  # mock_data <- tibble(
  #   endpoint_name = paste("Endpoint", seq_len(9)),
  #   capability_url = paste0("https://api.example.com/", seq_len(9), "/metadata"),
  #   developer = sample(c("Epic", "Cerner", "Athenahealth", "eCW", "NextGen"), 9, TRUE),
  #   source = sample(c("ONC", "CHPL", "Manual"), 9, TRUE),
  #   fhir_version = sample(c("3.0.1", "4.0.1", "4.1.0"), 9, TRUE),
  #   status = sample(c("Active", "Inactive"), 9, TRUE, prob = c(0.85, 0.15)),
  #   resources = sample(40:95, 9),
  #   search_params = sample(60:140, 9),
  #   operations = sample(2:12, 9)
  # )

  view_mode <- reactiveVal("table")
  grid_page <- reactiveVal(1)
  page_size <- 9

  observeEvent(input$toggle_view, {
    new_mode <- ifelse(view_mode() == "table", "grid", "table")
    view_mode(new_mode)
    updateActionButton(session, "toggle_view",
                       label = ifelse(new_mode == "table", "Switch to Grid View", "Switch to Table View"))
    grid_page(1)
  })
  observeEvent(input$grid_next, grid_page(grid_page() + 1))
  observeEvent(input$grid_prev, if (grid_page() > 1) grid_page(grid_page() - 1))

  output$grid_table <- renderReactable({
    req(view_mode() == "table")
    df <- endpoint_data()

    if (nrow(df) == 0) {
      return(reactable(tibble(Message = "No endpoints found for current filters.")))
    }

    reactable(
      df,
      columns = list(
        endpoint_names = colDef(name = "Endpoint"),
        url = colDef(
          name = "Capability URL",
          cell = function(x) {
            full <- paste0(x, "/metadata")
            sprintf("<a href='%s' target='_blank'>%s</a>", full, full)
          },
          html = TRUE
        ),
        vendor_name = colDef(name = "Vendor"),
        # list_source = colDef(name = "Source"),
        fhir_version = colDef(name = "FHIR Version"),
        status = colDef(name = "Status"),
        resources = colDef(name = "Resources", align = "right"),
        search_params = colDef(name = "Search Parameters", align = "right"),
        interactions = colDef(name = "Operations", align = "right")
      ),
      sortable = TRUE,
      searchable = TRUE,
      striped = TRUE,
      highlight = TRUE,
      bordered = TRUE,
      defaultPageSize = 8
    )
  })

  output$grid_view_container <- renderUI({
    req(view_mode() == "grid")
    d <- endpoint_data()

    total_pages <- ceiling(nrow(d) / page_size)
    current_page <- grid_page()
    start_row <- ((current_page - 1) * page_size) + 1
    end_row <- min(start_row + page_size - 1, nrow(d))
    d_page <- d[start_row:end_row, ]

    if (nrow(d_page) == 0) {
      return(div(em("No endpoints for current filters.")))
    }

    tagList(
      div(class = "endpoint-grid",
          lapply(seq_len(nrow(d_page)), function(i) {
            row <- d_page[i, ]
            div(class = "endpoint-card",
                h4(row$endpoint_names),
                div(class = "kv", tags$b("Vendor:"), row$vendor_name),
                div(
                  class = "kv",
                  tags$b("Capability URL:"),
                  tags$a(href = paste0(row$url, "/metadata"), target = "_blank",
                        paste0(row$url, "/metadata"))
                ),
                div(class = "kv", tags$b("FHIR Version:"), row$fhir_version),
                div(class = "kv", tags$b("Status:"), row$status),
                tags$hr(),
                div(class = "kv", tags$b("Resources:"), row$resources),
                div(class = "kv", tags$b("Search Params:"), row$search_params),
                div(class = "kv", tags$b("Operations:"), row$interactions)
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
}