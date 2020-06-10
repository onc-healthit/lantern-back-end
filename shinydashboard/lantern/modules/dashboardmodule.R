library(shiny)
library(shinydashboard)
library(readr)
library(scales)


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
    h3("Endpoint Counts by Vendor and FHIR Version"),
    fluidRow(
      column(width=4,
             tableOutput(ns("fhir_vendor_table"))
      ),
      column(width=8,
             plotOutput(ns("vendor_share_plot"))
      )
    ),
    h3("All Endpoint Responses"),
    fluidRow(
      column(width=4,
             tableOutput(ns("http_code_table")),
             p("All HTTP response codes ever received and count of endpoints which returned that code at some point in history"),
      ),
      column(width=8,
           plotOutput(ns("response_code_plot"))
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
  
  fhir_endpoint_totals <- get_endpoint_totals_list(db_tables)
  response_tally       <- get_response_tally_list(db_tables)
  http_pct             <- get_http_response_summary_tbl(db_tables)

  # create a summary table to show the response codes received along with 
  # the description for each code
  http_summary <- http_pct %>%
    left_join(http_response_code_tbl, by=c("code" = "code_chr")) %>%
    select(id,code,label) %>%
    group_by(code,label) %>%
    summarise(count=n()) 
  
  vendor_count_tbl <- get_fhir_version_vendor_count(endpoint_export_tbl)
  
  vc_totals <- vendor_count_tbl %>%
    filter(!(vendor_name == "Unknown")) %>%
    group_by(vendor_name) %>%
    summarise(total=sum(n))
  
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

  output$http_code_table   <- renderTable(http_summary %>% rename("HTTP Response" = code,Status=label,Count=count))

  output$fhir_vendor_table <- renderTable(vendor_count_tbl %>% select(Vendor=vendor_name,'FHIR Version'=fhir_version,Count=n))

  output$vendor_share_plot <- renderPlot({
   ggplot(vendor_count_tbl, aes(y = n, x = short_name, fill = fhir_version)) + 
      geom_bar(stat = "identity") +
      geom_text( aes(label = stat(y), group=short_name),
        stat = 'summary', fun = sum, vjust = -1
      ) +
      theme(text = element_text(size = 15)) +
      labs(fill = "FHIR Version",
           x = NULL,
           y = "Number of Endpoints",
           title = "Endpoints by Vendor and FHIR Version") +
      scale_fill_manual(values=c("#66C2A5","#8DA0CB","#EFA182","#E78AC3","#A6D854"))
  })

  output$response_code_plot <- renderPlot({
    ggplot(http_summary %>% mutate(Response=paste(code,"-",label)), aes(x=code,fill=as.factor(Response),y=count)) + 
    geom_bar(stat="identity") +
      theme(text = element_text(size = 15)) +
      labs(fill="Code",
         title="All HTTP Response Codes Ever Received from Endpoints",
         x="HTTP Response Received",
         y = "Count of endpoints")
  })  
}