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
    textOutput(ns("last_updated")),
    br(),
    fluidRow(
      infoBoxOutput(ns("total_endpoints_box"), width = 4),
      infoBoxOutput(ns("indexed_endpoints_box"), width = 4),
      infoBoxOutput(ns("nonindexed_endpoints_box"), width = 4)
    ),
    h2("Current endpoint responses:"),
    fluidRow(
      valueBoxOutput(ns("http_200_box")),
      valueBoxOutput(ns("http_404_box")),
      valueBoxOutput(ns("http_503_box"))
    ),
    actionButton(ns("show_info"), "Info", icon = tags$i(class = "fa fa-question-circle", "aria-hidden" = "true", role = "presentation", "aria-label" = "question-circle icon")),
    h3("Endpoint Counts by Developer and FHIR Version"),
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
        left_join(app$http_response_code_tbl(), by = c("code" = "code_chr")) %>%
        select(id, code, label) %>%
        group_by(code, label) %>%
        summarise(count = n())
    } else {
      res <- res %>%
        left_join(app$http_response_code_tbl(), by = c("code" = "code_chr")) %>%
        select(id, code, label) %>%
        group_by(code, label) %>%
        summarise(count = n())
    }

    res
  })

  # create a summary table to show the response codes received along with
  # the description for each code

  output$last_updated <- renderText(paste("Last Updated:", get_endpoint_last_updated(db_tables)))

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

  output$nonindexed_endpoints_box <- renderInfoBox({
    infoBox(
      "Non-Indexed Endpoints", isolate(app_data$fhir_endpoint_totals()$nonindexed_endpoints), icon = tags$i(class = "fa fa-comment-slash", "aria-hidden" = "true", role = "presentation", "aria-label" = "comment-slash icon"),
      color = "maroon"
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
      rename("HTTP Response" = code, Status = label, Count = count)
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
    ggplot(selected_http_summary() %>% mutate(Response = paste(code, "-", label)), aes(x = code, fill = as.factor(Response), y = count)) +
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
      title = "Information About Lantern",
      "Lantern takes a strict approach to showing FHIR Version and Developer information. If a given FHIR
      endpoint returns an error or cannot be reached during the current query period, Lantern will report FHIR Version as 'No Cap Stat' and
      Developer information as 'Unknown'.
      Other endpoints may fail to properly indicate FHIR Version or Developer information in their CapabilityStatement / Conformance Resource.",
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
