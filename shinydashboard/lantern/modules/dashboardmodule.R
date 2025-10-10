library(shiny)
library(shinydashboard)
library(readr)
library(scales)
library(dplyr)
library(ggplot2)
library(plotly)
library(highcharter)

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
        h3("All Endpoint Responses"),
        uiOutput("show_http_vendor_filters"),
        div(class = "plot-card",
          highcharter::highchartOutput(ns("response_code_plot"))
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
        reactable::reactableOutput(ns("fhir_vendor_table")),
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
    vendor_table_data <- prepare_vendor_data(db_tables)
    if (is.null(fhirVendorTableSize())) {
      fhirVendorTableSize(ceiling(nrow(vendor_table_data) / 2))
    }
    
    # Create a filtered version of the data for the table display
    display_data <- vendor_table_data %>%
      select(vendor_name, fhir_version, n, percentage)
      
    reactable(display_data,
                columns = list(
                  vendor_name = colDef(name = "Vendor"),
                  fhir_version = colDef(name = "FHIR Version"),
                  n = colDef(name = "Count"),
                  percentage = colDef(name = "FHIR Version %", format = colFormat(suffix = "%"))
                ),
                sortable = TRUE,
                searchable = TRUE,
                showSortIcon = TRUE,
                defaultPageSize = (ceiling(nrow(vendor_table_data) / 2))
    )
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
    # Estimate: use a stored metric if available, else dash
    avg <- tryCatch({ round(mean(na.omit(get_response_tally_list(db_tables)$response_time_mean)), 0) }, error = function(e) { NA })
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
    highcharter::highchartOutput(ns("vendor_share_plot"), height = plot_height_vendors())
  })
  output$vendor_share_plot <- highcharter::renderHighchart({
    vendor_plot_data <- prepare_vendor_data(db_tables) %>%
      filter(n > 0)  # Filter out zero counts

    # Determine top 10 vendors by total endpoints
    top_vendors <- vendor_plot_data %>%
      group_by(vendor_name) %>%
      summarise(total_endpoints = sum(n, na.rm = TRUE)) %>%
      arrange(desc(total_endpoints)) %>%
      slice_head(n = 10) %>%
      pull(vendor_name)

    vendor_plot_data <- vendor_plot_data %>%
      filter(vendor_name %in% top_vendors)

    # Order vendors by total endpoints (descending)
    vendor_levels <- vendor_plot_data %>%
      group_by(vendor_name) %>%
      summarise(total = sum(n, na.rm = TRUE)) %>%
      arrange(desc(total)) %>%
      pull(vendor_name)

    # Build series per FHIR version (stacked columns)
    versions <- vendor_plot_data %>% pull(fhir_version) %>% unique()
    series_list <- lapply(versions, function(v) {
      s <- vendor_plot_data %>%
        filter(fhir_version == v) %>%
        group_by(vendor_name) %>%
        summarise(n = sum(n, na.rm = TRUE)) %>%
        ungroup()
      # align to vendor_levels and fill missing with 0
      s_aligned <- data.frame(vendor_name = vendor_levels, stringsAsFactors = FALSE) %>%
        left_join(s, by = "vendor_name") %>%
        mutate(n = ifelse(is.na(n), 0, n))
      list(name = as.character(v), data = as.list(s_aligned$n))
    })

    # Create 3D stacked column chart with 10k cap, simple data labels and tooltip
    highchart() %>%
      hc_chart(type = "column", options3d = list(enabled = TRUE, alpha = 15, beta = 15, depth = 60)) %>%
      hc_xAxis(categories = vendor_levels, title = list(text = "Developer")) %>%
      hc_yAxis(max = 10000, title = list(text = "Number of Endpoints")) %>%
      hc_plotOptions(column = list(stacking = "normal", depth = 40, dataLabels = list(enabled = TRUE))) %>%
      hc_add_series_list(series_list) %>%
      hc_tooltip(pointFormat = "{series.name}: <b>{point.y}</b><br/>", shared = FALSE) %>%
      hc_legend(enabled = TRUE) %>%
      hc_title(text = "Top 10 Developers by Endpoint Count")
  })
  
  output$response_code_plot <- renderHighchart({
    pie_data <- selected_http_summary() %>%
      mutate(Response = paste(http_code, "-", code_label)) %>%
      group_by(Response) %>%
      summarise(count = sum(count_endpoints, na.rm = TRUE)) %>%
      ungroup()

    highchart() %>%
      hc_chart(type = "pie", options3d = list(enabled = TRUE, alpha = 45, beta = 0)) %>%
      hc_plotOptions(pie = list(allowPointSelect = TRUE, cursor = "pointer", depth = 35)) %>%
      hc_add_series(
        type = "pie",
        data = list_parse2(pie_data %>% select(Response, count))
      )
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