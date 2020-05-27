library(shiny)
library(shinydashboard)
library(readr)

dashboard_UI <- function(id) {

  ns <- NS(id)

  tagList(
    fluidRow(
      infoBoxOutput(ns("total_endpoints_box"),width=4),
      infoBoxOutput(ns("indexed_endpoints_box"),width=4),
      infoBoxOutput(ns("nonindexed_endpoints_box"),width=4)
    ),
    p("Current endpoint responses:"),
    fluidRow(
      valueBoxOutput(ns("http_200_box")),
      valueBoxOutput(ns("http_404_box")),
      valueBoxOutput(ns("http_503_box"))
    ),
    fluidRow(
      column(width=6,
             h3("Endpoint Counts by Vendor and FHIR Version"),
             tableOutput(ns("fhir_vendor_table"))
      ),
      column(width=6,
             h3("All Endpoint Responses"),
             tableOutput(ns("http_code_table")),
             p("All HTTP response codes ever received and count of endpoints which returned that code at some point in history"),
      )
    )
  )
}

dashboard <- function(
    input, 
    output, 
    session
){
  ns <- session$ns

  # Will make endpoint totals reactive
  
  fhir_endpoint_totals <- get_endpoint_totals(fhir_endpoints,fhir_endpoints_info)
  response_tally       <- get_response_tally(fhir_endpoints_info)
  http_pct             <- get_http_response_summary(fhir_endpoints_info_history)
  
  # Get the count of endpoints by vendor
  fhir_version_vendor_count <- endpoint_export_tbl %>%
    group_by(vendor_name,fhir_version) %>%
    tally() %>%
    select(Vendor=vendor_name,"FHIR Version"=fhir_version,"Count"=n)
  
  # create a summary table to show the response codes received along with 
  # the description for each code
  http_summary <- http_pct %>%
    left_join(http_response_code_tbl, by=c("code" = "code_chr")) %>%
    select(id,code,label) %>%
    group_by("HTTP Response" = code,"Status"=label) %>%
    summarise(Count=n()) 
  
  output$total_endpoints_box <- renderInfoBox({
    infoBox(
      "Total Endpoints", fhir_endpoint_totals$all_endpoints, icon = icon("fire", lib = "glyphicon"),
      color = "blue"
    )
  })
  output$indexed_endpoints_box <- renderInfoBox({
    infoBox(
      "Indexed Endpoints", 
      fhir_endpoint_totals$indexed_endpoints,
      icon = icon("flash", lib = "glyphicon"),
      color = "teal"
    )
  })
  output$nonindexed_endpoints_box <- renderInfoBox({
    infoBox(
      "Non-Indexed Endpoints", fhir_endpoint_totals$nonindexed_endpoints, icon = icon("comment-slash", lib = "font-awesome"),
      color = "maroon"
    )
  })
  output$http_200_box <- renderValueBox({
    valueBox(
      response_tally$http_200, "200 (Success)", icon = icon("thumbs-up", lib = "glyphicon"),
      color = "green"
    )
  })
  output$http_404_box <- renderValueBox({
    valueBox(
      response_tally$http_404, "404 (Not found)", icon = icon("thumbs-down", lib = "glyphicon"),
      color = "yellow"
    )
  })
  output$http_503_box <- renderValueBox({
    valueBox(
      response_tally$http_503, "503 (Unavailable)", icon = icon("ban-circle", lib = "glyphicon"),
      color = "orange"
    )
  })

  output$http_code_table   <- renderTable(http_summary)
  output$fhir_vendor_table <- renderTable(get_fhir_version_vendor_count(endpoint_export_tbl))
  
}