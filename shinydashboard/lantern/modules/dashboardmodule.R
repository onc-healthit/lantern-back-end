library(shiny)
library(shinydashboard)
library(readr)
library(scales)

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
    fluidRow(
      infoBoxOutput(ns("updated_time_box"), width = 4),
      infoBoxOutput(ns("total_endpoints_box"), width = 4),
      infoBoxOutput(ns("indexed_endpoints_box"), width = 4)
    ),
    h2("Current endpoint responses:"),
    fluidRow(
      valueBoxOutput(ns("http_200_box")),
      valueBoxOutput(ns("http_404_box")),
      valueBoxOutput(ns("http_503_box"))
    ),
    actionButton(ns("show_info"), "Info", icon = tags$i(class = "fa fa-question-circle", "aria-hidden" = "true", role = "presentation", "aria-label" = "question-circle icon")),
    h2("Endpoint Counts by Developer and FHIR Version"),
    fluidRow(
      custom_column_small(
            reactable::reactableOutput((ns("fhir_vendor_table")))
      ),
      custom_column_large(
              uiOutput(ns("vendors_plot")),
             htmlOutput(ns("note_text"))
      )
    ),
    h3("All Endpoint Responses"),
    uiOutput("show_http_vendor_filters"),
    fluidRow(
      custom_column_small(
             tableOutput(ns("http_code_table")),
             p("All HTTP response codes ever received and count of endpoints which returned that code at some point in history"),
      ),
      custom_column_large(
           plotOutput(ns("response_code_plot"))
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
                  percentage = colDef(name = "Developer Percentage", format = colFormat(suffix = "%"))
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

  output$updated_time_box <- renderInfoBox({
    infoBox(
      "Endpoints Last Queried:", get_endpoint_last_updated(db_tables), icon = tags$i(class = "fa fa-clock", "aria-hidden" = "true", role = "presentation", "aria-label" = "clock icon"),
      color = "purple"
    )
  })

  output$total_endpoints_box <- renderInfoBox({
    infoBox(
      "Total Endpoints", get_endpoint_totals_list(db_tables)$all_endpoints, icon = tags$i(class = "glyphicon glyphicon-fire", "aria-hidden" = "true", role = "presentation", "aria-label" = "fire icon"),
      color = "blue"
    )
  })

  output$indexed_endpoints_box <- renderInfoBox({
    infoBox(
      "Indexed Endpoints*",
      get_endpoint_totals_list(db_tables)$indexed_endpoints,
      icon =  tags$i(class = "glyphicon glyphicon-flash", "aria-hidden" = "true", role = "presentation", "aria-label" = "flash icon"),
      color = "teal"
    )
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

  output$http_code_table <- renderTable(
    selected_http_summary() %>%
      rename("HTTP Response" = http_code, Status = code_label, Count = count_endpoints)
  )

  plot_height_vendors <- reactive({
    max(fhirVendorTableSize() * 75, 400)
  })

  output$vendors_plot <- renderUI({
    plotOutput(ns("vendor_share_plot"), height = plot_height_vendors())
  })

  output$vendor_share_plot <- renderCachedPlot({
  vendor_plot_data <- prepare_vendor_data(db_tables) %>%
    filter(n > 0)  # Filter out zero counts
  
  # Create ordered factor levels based on sort_order
  # This ensures the plot shows vendors in the correct order
  vendor_levels <- vendor_plot_data %>%
    arrange(sort_order) %>%
    distinct(vendor_name) %>%
    pull(vendor_name)
  
  vendor_plot_data$vendor_name <- factor(
    vendor_plot_data$vendor_name,
    levels = vendor_levels
  )
  
  ggplot(vendor_plot_data, aes(y = n, x = fct_rev(vendor_name), fill = fhir_version)) +
    # Reduce bar width from 0.8 to 0.4 (half the current width)
    geom_col(width = 0.4) +
    geom_text(aes(label = n), position = position_stack(vjust = 0.5)) +
    theme(legend.position = "top") +
    theme(text = element_text(size = 15)) +
    labs(fill = "FHIR Version",
         x = "",
         y = "Number of Endpoints",
         title = "Endpoints by Developer and FHIR Version") +
    scale_y_continuous(sec.axis = sec_axis(~., name = "Number of Endpoints")) +
    coord_flip()
}, sizePolicy = sizeGrowthRatio(width = 400,
                                height = 400,
                                growthRate = 1.2),
  res = 72, cache = "app", cacheKeyExpr = {
    now("UTC")
  }
)
  
  output$response_code_plot <- renderCachedPlot({
    ggplot(selected_http_summary() %>% mutate(http_code = as.factor(http_code), Response = paste(http_code, "-", code_label)), aes(x = http_code, fill = as.factor(Response), y = count_endpoints)) +
    geom_bar(stat = "identity", show.legend = FALSE) +
      geom_text(aes(label = stat(y), group = http_code),
                stat = "summary", fun = sum, vjust = -1
      ) +
      theme(text = element_text(size = 15)) +
      labs(fill = "Code",
         title = "HTTP Response Codes Received from Endpoints During Most Recent Query",
         x = "HTTP Response Received",
         y = "Count of endpoints")
  }, sizePolicy = sizeGrowthRatio(width = 400,
                                  height = 400,
                                  growthRate = 1.2),
  res = 72, cache = "app", cacheKeyExpr = {
    list(now("UTC"), sel_vendor())
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