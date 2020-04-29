#
# This is a Shiny web application. You can run the application by clicking
# the 'Run App' button above.
#
# Find out more about building applications with Shiny here:
#
#    http://shiny.rstudio.com/
#

library(shiny)
library(shinydashboard)
library(DBI)
library(readr)
library(dplyr)
library(RPostgres)
library(ggplot2)
library(plotly)
library(scales)
library(forcats)
library(dygraphs)
library(config)
library(here)

http_response_code_tbl <- read_csv("./http_codes.csv") %>% mutate(code_chr=as.character(code))

db_config <- config::get("lantern")

con <- dbConnect(RPostgres::Postgres(),
                 dbname = db_config$database, 
                 host = db_config$server, # i.e. 'ec2-54-83-201-96.compute-1.amazonaws.com'
                 port = db_config$port, 
                 user = db_config$uid,
                 password = db_config$pwd
)

fhir_endpoints  <- tbl(con, "fhir_endpoints")
metrics_values  <- tbl(con, "metrics_values")
metrics_labels  <- tbl(con, "metrics_labels")
end_org         <- tbl(con, "endpoint_organization")
hit_prod        <- tbl(con, "healthit_products")
endpoint_export <- tbl(con, "endpoint_export")

fhir_endpoints_tbl <- as_tibble(fhir_endpoints)

http_response_ids <- metrics_labels %>%
    filter(metric_name == "AllEndpoints_http_request_responses") %>%
    select(id)

https_response_values <- metrics_values %>%
    right_join(http_response_ids, by = c("labels_id" = "id"))

curr_http_response_tally <- fhir_endpoints_tbl %>%
    select(http_response) %>%
    group_by(http_response) %>%
    tally()


totals <- list()
totals$all_endpoints <- nrow(fhir_endpoints_tbl)
totals$indexed_endpoints <- nrow(fhir_endpoints_tbl %>% filter(http_response != 0))

# Get the list of most recent HTTP responses when requesting the capability statement from the 
# fhir_endpoints
response_tally <- list()
response_tally$http_200 <- curr_http_response_tally %>% filter(http_response==200) %>% pull(n)
response_tally$http_404 <- curr_http_response_tally %>% filter(http_response==404) %>% pull(n)
response_tally$http_503 <- curr_http_response_tally %>% filter(http_response==503) %>% pull(n)

http_pct <- as_tibble(https_response_values %>% 
                          mutate(code=as.character(value)) %>%
                          group_by(labels_id,code,value) %>% 
                          summarise(Percentage=n()) %>% 
                          group_by(labels_id) %>% 
                          mutate(Percentage=Percentage/sum(Percentage,na.rm = TRUE)*100)
)

http_pctf <- http_pct %>% filter(value != 200) %>% mutate(name=as.factor(labels_id), code=as.factor(code)) 

http_summary <- http_pct %>%
    left_join(http_response_code_tbl, by=c("code" = "code_chr")) %>%
    select(code,label) %>%
    group_by("HTTP Response" = code,"Status"=label) %>%
    summarise(Count=n()) 

fhir_version_tbl <- as_tibble(tbl(con,sql("select id,url,vendor,capability_statement->>'fhirVersion' as FHIR from fhir_endpoints where capability_statement->>'fhirVersion' IS NOT NULL")))
fhir_version_list <- as.list(fhir_version_tbl %>% distinct("FHIR Version"=fhir))
fhir_version_vendor_count <- fhir_version_tbl %>%
    mutate(vendor = na_if(vendor,"")) %>%
    tidyr::replace_na(list(vendor="Unknown")) %>%
    group_by(vendor,fhir) %>%
    tally() %>%
    select(Vendor=vendor,"FHIR Version"=fhir,"Count"=n)


vendor_list <- as.list(as_tibble(fhir_endpoints %>% distinct(vendor)) %>% mutate(vendor = na_if(vendor,"")) %>% tidyr::replace_na(list(vendor="Unknown")) %>% pull(vendor))

foo <- tbl(con,sql("SELECT floor(extract(epoch from metrics_values.time)/240)*240 AS time, AVG(metrics_values.value) 
FROM metrics_labels, metrics_values
WHERE metrics_labels.metric_name = 'AllEndpoints_http_response_time' 
AND metrics_labels.id = metrics_values.labels_id
AND metrics_values.time BETWEEN '2020-04-07T22:40:39.024Z' AND '2020-04-14T22:40:39.024Z'
GROUP BY time
ORDER BY time "))
ff <- as_tibble(foo)

# Define the UI components
ui <- dashboardPage(
    dashboardHeader(
        title = "Lantern Dashboard",
        titleWidth = 200
    ),
    # Sidebar with a slider input for number of bins 
    dashboardSidebar(
        sidebarMenu(
            menuItem("Dashboard", tabName = "dashboard", icon = icon("dashboard"),selected=TRUE),
            menuItem("Availability", icon = icon("th"), tabName = "availability", badgeLabel = "new",
                     badgeColor = "green"
            ),
            menuItem("Performance", icon = icon("bar-chart-o"),
                     menuSubItem("Mean Response Time", tabName = "subitem1"),
                     menuSubItem("Performance sub-item", tabName = "subitem2")
            ),
            menuItem("Location", tabName = "location", icon=icon("map")),
            selectInput(
                inputId = "fhir_version",
                label = "FHIR Version:",
                choices = fhir_version_list,
                selected = 99,
                size = length(fhir_version_list),
                selectize = FALSE
            ),
            selectInput(
                inputId = "vendor",
                label = "Vendor:",
                choices = vendor_list,
                selected = 99,
                size = length(vendor_list),
                selectize = FALSE
            )
            
        )
    ),
    
    # Show a plot of the generated distribution
    dashboardBody(
        tabItems(
            tabItem("dashboard",
                    fluidRow(
                        infoBoxOutput("total_endpoints_box",width=6),
                        infoBoxOutput("indexed_endpoints_box",width=6)
                    ),
                    p("Current endpoint responses:"),
                    fluidRow(
                        valueBoxOutput("http_200_box"),
                        valueBoxOutput("http_404_box"),
                        valueBoxOutput("http_503_box")
                    ),
                    fluidRow(
                        column(width=6,
                               h3("Endpoint Counts by Vendor and FHIR Version"),
                               tableOutput("fhir_vendor_table")
                        ),
                        column(width=6,
                               h3("All Endpoint Responses"),
                               tableOutput("http_code_table"),
                               p("All HTTP response codes ever received and count of endpoints which returned that code at some point in history"),
                        )
                    )
            ),
            tabItem("subitem1",
                    dygraphOutput("mean_response_time_plot")
            ),
            tabItem("availability",
                    plotlyOutput("non_200")
                    
            ),
            tabItem("location",
                    h3("Map of Zip Codes with identified endpoint/organization"),
                    img(src="images/endpoint_zcta_map.png",width="100%"))
        )
    )
)

# Define server logic required to draw a histogram
server <- function(input, output) {
    
    output$http_code_table <- renderTable(http_summary)
    output$fhir_vendor_table <-renderTable(fhir_version_vendor_count)
    
    output$non_200 <- renderPlotly({
        ggplotly(ggplot(http_pctf,aes(x=name,y=Percentage,fill=code)) +
                     geom_bar(stat="identity") + ggtitle("Endpoints returning non-HTTP 200 responses"))
    })
    
    output$total_endpoints_box <- renderInfoBox({
        infoBox(
            "Total Endpoints", totals$all_endpoints, icon = icon("fire", lib = "glyphicon"),
            color = "blue"
        )
    })
    output$mean_response_time_plot <- renderDygraph({
        dygraph(ff,main="Mean Response Time")
    })
    output$indexed_endpoints_box <- renderInfoBox({
        infoBox(
            "Indexed Endpoints", totals$indexed_endpoints, icon = icon("flash", lib = "glyphicon"),
            color = "teal"
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
}

# Run the application 
shinyApp(ui = ui, server = server)
