library(shiny)
library(shinydashboard)
library(readr)
library(scales)
library(dplyr)
library(ggplot2)
library(plotly)
library(echarts4r)

custom_column_small <- function(...) {
    tags$div(
      class = "col-md-4",
      ...
    )
}

get_endpoint_totals_list <- function(db_tables) {
  totals_data <- db_tables$mv_endpoint_totals %>%
    as.data.frame() %>%
    slice(1)
  
  fhir_endpoint_totals <- list(
    "all_endpoints"     = totals_data$all_endpoints,
    "indexed_endpoints" = totals_data$indexed_endpoints,
    "nonindexed_endpoints" = totals_data$nonindexed_endpoints
  )
  
  return(fhir_endpoint_totals)
}

get_response_tally_list <- function(db_tables) {
  response_tally <- db_tables$mv_response_tally %>%
                    as.data.frame() %>%
                    slice(1)
  
  return(response_tally)
}

custom_column_large <- function(...) {
    tags$div(
      class = "col-md-8",
      ...
    )
}

dashboard_UI <- function(id) {

  ns <- NS(id)

  tagList(
  tags$style(HTML('\n+      /* Hover lift for status cards */\n+      .status-card {\n+        transition: transform 0.18s ease, box-shadow 0.18s ease;\n+        will-change: transform;\n+        cursor: default;\n+      }\n+      .status-card:hover {\n+        transform: translateY(-8px);\n+        box-shadow: 0 12px 30px rgba(0,0,0,0.35);\n+        z-index: 5;\n+      }\n+      .status-card .status-icon {\n+        transition: transform 0.18s ease;\n+      }\n+      .status-card:hover .status-icon {\n+        transform: translateY(-2px) scale(1.03);\n+      }\n+      /* Plot card styling - mirrors status card lift and shadow */\n+      .plot-card {\n+        background: #ffffff;\n+        padding: 12px;\n+        border-radius: 8px;\n+        box-shadow: 0 2px 6px rgba(0,0,0,0.08);\n+        transition: transform 0.18s ease, box-shadow 0.18s ease;\n+        overflow: hidden; /* ensure rounded corners clip plot content */\n+      }\n+      .plot-card:hover {\n+        transform: translateY(-8px);\n+        box-shadow: 0 12px 30px rgba(0,0,0,0.18);\n+        z-index: 4;\n+      }\n+      .plot-card svg, .plot-card canvas {\n+        display: block;\n+        border-radius: 6px;\n+      }\n+    ')),
    fluidRow(
    column(width = 4,
       div(class = "status-box-wrapper",
         div(class = "status-card success",
           div(class = "status-header",
             div(class = "status-icon", HTML('<i class="fa fa-clock" aria-hidden="true" role="presentation" aria-label="clock icon"></i>')),
             div(class = "status-title", "Endpoints Last Queried")
           ),
           div(class = "status-value", textOutput(ns("updated_time_box"))),
           div(class = "status-subtitle", span(class = "pulse"), "Operational")
         )
       )
    ),
    column(width = 4,
       div(class = "status-box-wrapper",
         div(class = "status-card warning",
           div(class = "status-header",
             div(class = "status-icon", HTML('<i class="glyphicon glyphicon-fire" aria-hidden="true" role="presentation" aria-label="fire icon"></i>')),
             div(class = "status-title", "Total Endpoints")
           ),
           div(class = "status-value", textOutput(ns("total_endpoints_box"))),
           div(class = "status-subtitle", "Slow responses")
         )
       )
    ),
    column(width = 4,
       div(class = "status-box-wrapper",
         div(class = "status-card error",
           div(class = "status-header",
             div(class = "status-icon", HTML('<i class="glyphicon glyphicon-flash" aria-hidden="true" role="presentation" aria-label="flash icon"></i>')),
             div(class = "status-title", "Indexed Endpoints")
           ),
           div(class = "status-value", textOutput(ns("indexed_endpoints_box"))),
           div(class = "status-subtitle", "Non-responsive")
         )
       )
    )
    ),

    # spacer between top row and second row of status cards
    tags$div(style = "height: 18px;"),

    fluidRow(
    column(width = 3,
       div(class = "status-box-wrapper",
         div(class = "status-card success",
           div(class = "status-header",
             div(class = "status-icon", "✓"),
             div(class = "status-title", "Healthy Endpoints")
           ),
           div(class = "status-value", textOutput(ns("total_endpoints_box_plain"))),
           div(class = "status-subtitle", span(class = "pulse"), "Operational")
         )
       )
    ),
    column(width = 3,
       div(class = "status-box-wrapper",
         div(class = "status-card warning",
           div(class = "status-header",
             div(class = "status-icon", "⚠"),
             div(class = "status-title", "Degraded Performance")
           ),
           div(class = "status-value", textOutput(ns("degraded_count"))),
           div(class = "status-subtitle", "Slow responses")
         )
       )
    ),
    column(width = 3,
       div(class = "status-box-wrapper",
         div(class = "status-card error",
           div(class = "status-header",
             div(class = "status-icon", "✕"),
             div(class = "status-title", "Failed Endpoints")
           ),
           div(class = "status-value", textOutput(ns("failed_count"))),
           div(class = "status-subtitle", "Non-responsive")
         )
       )
    ),
    column(width = 3,
       div(class = "status-box-wrapper",
         div(class = "status-card info",
           div(class = "status-header",
             div(class = "status-icon", "⚡"),
             div(class = "status-title", "Avg Response Time")
           ),
           div(class = "status-value", textOutput(ns("avg_response_time"))),
           div(class = "status-subtitle", "Avg latency")
         )
       )
    )
    ),

    tags$hr(),
    fluidRow(
      column(width = 6,
        # response plot moved into the left column (was fhir_vendor_table)
        h4(tags$b("HTTP Response Distribution")),
        uiOutput("show_http_vendor_filters"),
        div(class = "plot-card",
          echarts4r::echarts4rOutput(ns("response_code_plot"), height = "520px")
        )
      ),
      column(width = 6,
        # vendors_plot reduced to half page (col-md-6)
        div(class = "plot-card",
          uiOutput(ns("vendors_plot"))
        )
      )
    ),
    fluidRow(
    column(width = 12,
      # fhir_vendor_table moved here (previously the full-width response plot)
      div(class = "modern-endpoints-table",
        reactable::reactableOutput(ns("fhir_vendor_table"))
      ),
        htmlOutput(ns("note_text"))
      )
    ),
    tags$p("*An endpoint is considered to be an \"Indexed Endpoint\" when it has been queried by the Lantern system at least once. If an endpoint has never been queried by the Lantern system yet, it will not be counted towards the total number of \"Indexed Endpoints\".", style = "font-style: italic;")
  )
}

dashboard <- function(
    input,
    output,
    session,
    sel_vendor
) {
  ns <- session$ns

  fhirVendorTableSize <- reactiveVal(NULL)

  # Fixed prepare_vendor_data to handle integer64 data type and sort based on top developers
  prepare_vendor_data <- function(db_tables) {
    # Directly use the materialized view and convert integer64 to regular integers
    fhir_data <- db_tables$mv_vendor_fhir_counts %>% 
      collect() %>%
      mutate(n = as.integer(n))  # Convert integer64 to regular integer
    
    # Replace NA values with "Unknown" in vendor_name and fhir_version
    fhir_data <- fhir_data %>%
      mutate(
        vendor_name = ifelse(is.na(vendor_name), "Unknown", vendor_name),
        fhir_version = ifelse(is.na(fhir_version), "Unknown", fhir_version)
      )
    
    # Calculate percentage for each vendor
    all_vendor_counts <- fhir_data %>%
      group_by(vendor_name) %>%
      summarise(developer_count = sum(n))
    
    # Join back to get percentages
    fhir_data <- fhir_data %>%
      left_join(all_vendor_counts, by = "vendor_name") %>%
      mutate(percentage = as.integer(round((n / developer_count) * 100, digits = 0))) %>%
      # Select only the columns needed
      select(vendor_name, fhir_version, n, percentage, sort_order) %>%
      # Arrange by sort_order for consistent display
      arrange(sort_order, fhir_version)
    
    return(fhir_data)
  }

  output$fhir_vendor_table <-  reactable::renderReactable({
    # Read the materialized view mv_developer_endpoint_summary directly and display columns as-is
    vendor_table_data <- data.frame()
    if (!is.null(db_tables) && !is.null(db_tables$mv_developer_endpoint_summary)) {
      vendor_table_data <- tryCatch(
        collect(db_tables$mv_developer_endpoint_summary),
        error = function(e) {
          tryCatch(as.data.frame(db_tables$mv_developer_endpoint_summary), error = function(e) data.frame())
        }
      )
    }

    display_data <- if (is.null(vendor_table_data) || length(vendor_table_data) == 0) data.frame() else as.data.frame(vendor_table_data)

    # track table size for layout logic
    if (is.null(fhirVendorTableSize())) {
      fhirVendorTableSize(ifelse(nrow(display_data) > 0, ceiling(nrow(display_data) / 2), 5))
    }

    # Render the table without mutating the data; only provide simple formatting where sensible
    if (nrow(display_data) == 0) {
      reactable::reactable(display_data,
                  sortable = TRUE,
                  searchable = TRUE,
                  showSortIcon = TRUE,
                  defaultPageSize = 5
      )
    } else {
      cols <- list()
      nm <- names(display_data)
      if ("developer_name" %in% nm) cols$developer_name <- reactable::colDef(name = "Developer")
      if ("total_endpoints_count" %in% nm) cols$total_endpoints_count <- reactable::colDef(name = "Total Endpoints", format = reactable::colFormat(separators = TRUE))
      if ("healthy_endpoints_count" %in% nm) cols$healthy_endpoints_count <- reactable::colDef(name = "Healthy Endpoints", format = reactable::colFormat(separators = TRUE))
      if ("status" %in% nm) cols$status <- reactable::colDef(
        name = "Status",
        sortable = TRUE,
        cell = function(value) {
          v <- as.character(value)
          if (grepl("true", tolower(v))) {
            div(class = "status-badge status-success", "Available")
          } else if (grepl("false", tolower(v))) {
            div(class = "status-badge status-error", "Not Available")
          } else if (grepl("^critical$", tolower(v))) {
            # map explicit 'Critical' status to a red error badge
            div(class = "status-badge status-error", v)
          } else if (grepl("^degraded$", tolower(v))) {
            # map explicit 'Degraded' status to a yellow warning badge
            div(class = "status-badge status-warning", v)
          } else {
            # fallback: show original value with an info badge
            div(class = "status-badge status-info", v)
          }
        }
      )
      if ("avg_response_time_seconds" %in% nm) cols$avg_response_time_seconds <- reactable::colDef(name = "Avg Response Time (ms)", format = reactable::colFormat(digits = 2))
      if ("uptime_pct" %in% nm) cols$uptime_pct <- reactable::colDef(
        name = "Uptime %",
        sortable = TRUE,
        cell = function(value) {
          uptime_num <- suppressWarnings(as.numeric(value))
          if (is.na(uptime_num)) uptime_num <- 0

          fill_color <- if (uptime_num >= 90) {
            "#28a745"  # Green
          } else if (uptime_num >= 70) {
            "#ffc107"  # Yellow
          } else if (uptime_num >= 50) {
            "#fd7e14"  # Orange
          } else {
            "#dc3545"  # Red
          }

          div(
            class = "availability-container",
            div(
              class = "availability-bar",
              div(
                class = "availability-fill",
                style = list(
                  width = paste0(uptime_num, "%"),
                  background = fill_color
                )
              )
            ),
            div(class = "availability-text", style = list(color = fill_color), paste0(round(uptime_num, 1), "%"))
          )
        }
      )

      reactable::reactable(display_data,
                  columns = cols,
                  sortable = TRUE,
                  searchable = TRUE,
                  showSortIcon = TRUE,
                  defaultPageSize = ifelse(nrow(display_data) > 0, ceiling(nrow(display_data) / 2), 5)
      )
    }
  })

  observeEvent(input$fhir_vendor_table_state$length, {
    page <- input$fhir_vendor_table_state$length
    fhirVendorTableSize(page)
  })

  selected_http_summary <- reactive({
    res <- isolate(get_http_response_tbl_all())
    req(sel_vendor())
    if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
      res <- isolate(get_http_response_tbl(sel_vendor()))
    } else {
      res <- isolate(get_http_response_tbl_all())
    }

    res
  })

  # create a summary table to show the response codes received along with
  # the description for each code

  output$updated_time_box <- renderText({
    val <- tryCatch({ get_endpoint_last_updated(db_tables) }, error = function(e) { NA })
    if (is.na(val) || is.null(val)) return("-")
    as.character(val)
  })

  output$total_endpoints_box <- renderText({
    val <- tryCatch({ get_endpoint_totals_list(db_tables)$all_endpoints }, error = function(e) { NA })
    if (is.na(val) || is.null(val)) return("-")
    format(val, big.mark = ",")
  })

  # Plain text outputs to feed the custom status cards
  output$total_endpoints_box_plain <- renderText({
    format(get_endpoint_totals_list(db_tables)$all_endpoints, big.mark = ",")
  })

  output$degraded_count <- renderText({
    # fallback: use http_404 + http_503 as a rough degraded/failed metric if no dedicated metric exists
    val <- tryCatch({ as.integer(get_response_tally_list(db_tables)$http_404 + get_response_tally_list(db_tables)$http_503) }, error = function(e) { NA })
    format(ifelse(is.na(val), "-", val), big.mark = ",")
  })

  output$failed_count <- renderText({
    val <- tryCatch({ as.integer(get_response_tally_list(db_tables)$http_503) }, error = function(e) { NA })
    format(ifelse(is.na(val), "-", val), big.mark = ",")
  })

  output$avg_response_time <- renderText({
    # Run the requested SQL against the DB and compute mean of the returned values
    avg <- tryCatch({
      # Use the project's db_connection tbl/sql pattern when available
  query <- "SELECT fem.response_time_seconds FROM fhir_endpoints_info fei JOIN fhir_endpoints_metadata fem ON fei.metadata_id = fem.id WHERE fei.requested_fhir_version = 'None' AND fem.response_time_seconds != -1"
      # Prefer using tbl(db_connection, sql(...)) to stay consistent with project's DB handling
      res <- tryCatch({
        tbl(db_connection, sql(query)) %>% collect()
      }, error = function(e) {
        # fallback: try DBI::dbGetQuery if tbl/sql fails
        tryCatch(DBI::dbGetQuery(db_connection, query), error = function(e) data.frame())
      })

      if (is.null(res) || nrow(res) == 0) return(NA_real_)

      # extract column (first column expected to be http_response)
  vals <- res[[1]]
  vals_num <- suppressWarnings(as.numeric(vals))
  if (all(is.na(vals_num))) return(NA_real_)
  # compute mean in seconds, convert to milliseconds for display
  mean_seconds <- mean(na.omit(vals_num))
  round(mean_seconds * 1000, 0)
    }, error = function(e) { NA })

    if (is.na(avg)) return("- ms")
    paste0(avg, " ms")
  })

  output$indexed_endpoints_box <- renderText({
    val <- tryCatch({ get_endpoint_totals_list(db_tables)$indexed_endpoints }, error = function(e) { NA })
    if (is.na(val) || is.null(val)) return("-")
    format(val, big.mark = ",")
  })

  output$http_200_box <- renderValueBox({
    valueBox(
      get_response_tally_list(db_tables) %>% pull(http_200), "200 (Success)", icon = tags$i(class = "glyphicon glyphicon-thumbs-up", "aria-hidden" = "true", role = "presentation", "aria-label" = "thumbs-up icon"),
      color = "green"
    )
  })

  output$http_404_box <- renderValueBox({
    valueBox(
      get_response_tally_list(db_tables) %>% pull(http_404), "404 (Not found)", icon = tags$i(class = "glyphicon glyphicon-thumbs-down", "aria-hidden" = "true", role = "presentation", "aria-label" = "thumbs-down icon"),
      color = "yellow"
    )
  })

  output$http_503_box <- renderValueBox({
    valueBox(
      get_response_tally_list(db_tables) %>% pull(http_503), "503 (Unavailable)", icon = tags$i(class = "glyphicon glyphicon-ban-circle", "aria-hidden" = "true", role = "presentation", "aria-label" = "ban-circle icon"),
      color = "orange"
    )
  })

  # http_code_table removed; response plot shows aggregated HTTP responses instead

  plot_height_vendors <- reactive({
    # Attempt to read the rendered plot width from clientData so we can ensure
    # the width is at least 1.5x the height. Fallback to previous sizing logic
    # when clientData is not yet available.
    base_height <- max(fhirVendorTableSize() * 75, 400)
    # clientData key for the output width uses the namespaced output id
    out_id <- ns("vendor_share_plot")
    width_key <- paste0("output_", out_id, "_width")
    w <- session$clientData[[width_key]]
    if (!is.null(w) && is.numeric(w) && w > 0) {
      cap_height <- floor(w / 1.5)
      # ensure a sensible minimum height
      height <- max(min(base_height, cap_height), 240)
      return(height)
    }
    base_height
  })

  output$vendors_plot <- renderUI({
  div(style = "margin-top:200px;",
    echarts4r::echarts4rOutput(ns("vendor_share_plot"), height = plot_height_vendors())
  )
  })
  output$vendor_share_plot <- echarts4r::renderEcharts4r({
    vendor_plot_data <- prepare_vendor_data(db_tables) %>%
      filter(n > 0)  # Filter out zero counts

    # Aggregate and take top 10 by number of endpoints
    df_plot <- vendor_plot_data %>%
      group_by(vendor_name) %>%
      summarise(total = sum(n, na.rm = TRUE)) %>%
      arrange(desc(total)) %>%
      slice_head(n = 10) %>%
      ungroup()

    if (nrow(df_plot) == 0) {
      # return an empty echarts object with a placeholder title
      e_charts(data.frame(x = character(0), y = numeric(0))) %>%
        e_title(text = "Top 10 Developers by Endpoint Count")
    } else {
      df_plot %>%
        e_charts(vendor_name) %>%
  e_bar(total, name = "Total Endpoints") %>%
  # balanced grid so visual center lines up with pie
  e_grid(left = "8%", right = "8%", containLabel = TRUE) %>%
        e_tooltip(trigger = "axis") %>%
        # reduce font size and wrap long developer names to multiple lines
  e_x_axis(axisLabel = list(interval = 0, rotate = -30, fontSize = 10, formatter = htmlwidgets::JS("function(value){ return value.length > 18 ? value.replace(/(.{18})/g,'$1\\n') : value }") )) %>%
  e_title(text = "Top 10 Developers by Endpoint Count") %>%
        e_legend(show = FALSE)
    }
  })
  
  output$response_code_plot <- echarts4r::renderEcharts4r({
    pie_data <- selected_http_summary() %>%
      mutate(Response = paste(http_code, "-", code_label)) %>%
      group_by(Response) %>%
      summarise(count = sum(count_endpoints, na.rm = TRUE)) %>%
      ungroup()

    if (is.null(pie_data) || nrow(pie_data) == 0) {
      e_charts(data.frame(name = character(0), value = numeric(0)))
    } else {
      pie_data %>%
        e_charts(Response) %>%
        e_pie(count, radius = c("40%", "60%")) %>%
        e_tooltip(trigger = "item", formatter = htmlwidgets::JS("function(params){ return params.name + ': <b>' + params.value + '</b>' }") ) %>%
        e_legend(show = TRUE)
    }
  })

  observeEvent(input$show_info, {
    showModal(modalDialog(
      title = "Information About Lantern FHIR Version and Developer Data",
       p(HTML("Lantern takes a strict approach to showing FHIR Version and Developer Information. Some terminology Lantern uses to describe FHIR Version and Developer Information are as follows: <br><br>
       
      <strong>Endpoints may return an error, may not be able to be reached during the current query period, or may not return a CapabilityStatement / Conformance Resource. Lantern reports FHIR Version and Developer Information for these situations as follows:</strong> <br><br>
       &ensp;- <b>Developer:</b> Lantern will report Developer information as \"Unknown\" in any of these situations, since Developer information is collected from the publisher field of an endpoint's CapabilityStatement / Conformance Resource. <br>
       &ensp;- <b>FHIR Version:</b> Lantern will report a FHIR Version as \"No Cap Stat\" in any of these situations, since FHIR Version information is collected from the fhirVersion field of an endpoint's CapabilityStatement / Conformance Resource.<br><br>
       
       <strong>Endpoints may fail to properly indicate FHIR Version or Developer information in their CapabilityStatement / Conformance Resource. Lantern handles these situations as follows:</strong> <br><br>
       &ensp;- <b>Developer:</b> If an endpoint fails to properly indicate Developer Information such that Lantern cannot make a match between the Developer information included in the publisher field of the CapabilityStatement / Conformance Resource and the list of Developers Lantern 
       has stored, Lantern will report the Developer information as \"Unknown\". <br>
       &ensp;- <b>FHIR Version:</b> If an endpoint fails to properly indicate FHIR Version Information such that Lantern cannot recognize the FHIR Version included in the fhirVersion field of the CapabilityStatement / Conformance Resource as one of the valid published FHIR Versions, Lantern will take the following steps: <br>
       &emsp;1. Lantern will check if the FHIR Version contains any dash (-) characters. If it does, Lantern will remove the dash and everything that comes after it, and then check if it is a valid FHIR Version. <br>
       &emsp;2. If the FHIR Version does not have any dashes, or if after removing the dash and everything after it from the reported FHIR Version it is still is invalid, Lantern will report the FHIR Version as \"Unknown\". <br>
       &emsp;- <i>Note: Lantern will still display the invalid FHIR Version exactly as indicated by the endpoint's capability statement on the endpoint tab table for that endpoint, and within the popup modal for that particular endpoint.</i>
       ")),
      easyClose = TRUE
    ))
  })

  output$note_text <- renderUI({
    note_info <- "(1) The endpoints queried by Lantern are limited to Fast Healthcare Interoperability
               Resources (FHIR) endpoints published publicly by Certified API Developers in conformance with
               the ONC Cures Act Final Rule. This data, therefore, may not represent all FHIR endpoints in existence.
               (2) The number of endpoints for each Certified API Developer and FHIR version is a sum of all
               API Information Sources and unique endpoints discovered for each unique Certified API Developer.
               The API Information Source name associated with each endpoint may be represented as different
               organization types, including as a single clinician, practice group, facility or health system.
               Due to this variation in how API Information Sources are represented, insights gathered from this
               data should be framed accordingly."
    res <- paste("<div style='font-size: 16px;'><b>Note:</b>", note_info, "</div>")
    HTML(res)
  })

}