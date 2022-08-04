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
             tableOutput(ns("fhir_vendor_table"))
      ),
      custom_column_large(
             plotOutput(ns("vendor_share_plot")),
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
    )
  )
}

dashboard <- function(
    input,
    output,
    session,
    sel_vendor
) {
  ns <- session$ns

  selected_http_summary <- reactive({
    res <- isolate(app_data$http_pct())
    req(sel_vendor())
    if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
      res <- res %>%
        filter(vendor_name == sel_vendor()) %>%
        left_join(app$http_response_code_tbl, by = c("code" = "code_chr")) %>%
        select(id, code, label, fhir_version) %>%
        group_by(code, label, fhir_version) %>%
        summarise(count = n())
    } else {
      res <- res %>%
        left_join(app$http_response_code_tbl, by = c("code" = "code_chr")) %>%
        select(id, code, label, fhir_version) %>%
        group_by(code, label, fhir_version) %>%
        summarise(count = n())
        
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
      "Total Endpoints", isolate(app_data$fhir_endpoint_totals()$all_endpoints), icon = tags$i(class = "glyphicon glyphicon-fire", "aria-hidden" = "true", role = "presentation", "aria-label" = "fire icon"),
      color = "blue"
    )
  })

  output$indexed_endpoints_box <- renderInfoBox({
    infoBox(
      "Indexed Endpoints",
      isolate(app_data$fhir_endpoint_totals()$indexed_endpoints),
      icon =  tags$i(class = "glyphicon glyphicon-flash", "aria-hidden" = "true", role = "presentation", "aria-label" = "flash icon"),
      color = "teal"
    )
  })

  output$http_200_box <- renderValueBox({
    valueBox(
      isolate(app_data$response_tally()$http_200), "200 (Success)", icon = tags$i(class = "glyphicon glyphicon-thumbs-up", "aria-hidden" = "true", role = "presentation", "aria-label" = "thumbs-up icon"),
      color = "green"
    )
  })

  output$http_404_box <- renderValueBox({
    valueBox(
      isolate(app_data$response_tally()$http_404), "404 (Not found)", icon = tags$i(class = "glyphicon glyphicon-thumbs-down", "aria-hidden" = "true", role = "presentation", "aria-label" = "thumbs-down icon"),
      color = "yellow"
    )
  })

  output$http_503_box <- renderValueBox({
    valueBox(
      isolate(app_data$response_tally()$http_503), "503 (Unavailable)", icon = tags$i(class = "glyphicon glyphicon-ban-circle", "aria-hidden" = "true", role = "presentation", "aria-label" = "ban-circle icon"),
      color = "orange"
    )
  })

  output$http_code_table   <- renderTable(
    selected_http_summary() %>%
      rename("HTTP Response" = code, Status = label, "FHIR Version" = fhir_version, Count = count) 
  )

  output$fhir_vendor_table <- renderTable(
    isolate(app_data$vendor_count_tbl()) %>%
      select(Vendor = vendor_name, "FHIR Version" = fhir_version, Count = n)
  )

  output$vendor_share_plot <- renderCachedPlot({
   ggplot(isolate(app_data$vendor_count_tbl()), aes(y = n, x = short_name, fill = fhir_version)) +
      geom_bar(stat = "identity") +
      geom_text(aes(label = stat(y), group = short_name),
        stat = "summary", fun = sum, vjust = -1
      ) +
      theme(text = element_text(size = 15)) +
      labs(fill = "FHIR Version",
           x = NULL,
           y = "Number of Endpoints",
           title = "Endpoints by Developer and FHIR Version")
  }, sizePolicy = sizeGrowthRatio(width = 400,
                                  height = 333,
                                  growthRate = 1.2),
    res = 72, cache = "app", cacheKeyExpr = {
      app_data$last_updated()
    }
  )
  output$response_code_plot <- renderCachedPlot({
    ggplot(selected_http_summary() %>% mutate(Response = paste(code, "-", label)), aes(x = code, fill = fhir_version, y = count)) +
    geom_bar(stat = "identity") +
      geom_text(aes(label = stat(y), group = code),
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
    list(app_data$last_updated(), sel_vendor())
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
               the ONC Cures Act Final Rule, or discovered through the National Plan and Provider Enumeration
               System (NPPES). This data, therefore, may not represent all FHIR endpoints in existence.
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
