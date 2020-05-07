#
# Lantern metrics dashboard
# This Shiny application will display metrics on FHIR endpoints as 
# monitored by the Lantern application. 
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
library(lubridate)
library(xts)

# Load table of http response codes and descriptions
root <- ifelse(Sys.getenv("HOME")=='/home/shiny',".","lantern")
http_response_code_tbl <- read_csv(here(root,"http_codes.csv")) %>% mutate(code_chr=as.character(code))

# Read database connection information from .Renviron file
# local development: readRenviron("../.env"); db_config$host = "localhost"
# db_config <- config::get("lantern") 

db_config <- list("dbname" = Sys.getenv("LANTERN_DBNAME"),
                  "host" = Sys.getenv("LANTERN_DBHOST"),
                  "port" = Sys.getenv("LANTERN_DBPORT"),
                  "user" = Sys.getenv("LANTERN_DBUSER"),
                  "password" = Sys.getenv("LANTERN_DBPASSWORD")
)

# Connect to the Lantern database
con <- dbConnect(RPostgres::Postgres(),
                 dbname = db_config$dbname, 
                 host = db_config$host, # i.e. 'ec2-54-83-201-96.compute-1.amazonaws.com'
                 port = db_config$port, 
                 user = db_config$user,
                 password = db_config$password
)

# Make connections to the various lantern tables
fhir_endpoints  <- tbl(con, "fhir_endpoints")
metrics_values  <- tbl(con, "metrics_values")
metrics_labels  <- tbl(con, "metrics_labels")
end_org         <- tbl(con, "endpoint_organization")
hit_prod        <- tbl(con, "healthit_products")
endpoint_export <- tbl(con, "endpoint_export")


# Get the table of fhir endpoints. There may be endpoints we have not reached
# so get counts of indexed and non-indexed endpoints
fhir_endpoints_tbl <- as_tibble(fhir_endpoints)
fhir_endpoint_totals <- list(
    "all_endpoints"     = nrow(fhir_endpoints_tbl),
    "indexed_endpoints" = nrow(fhir_endpoints_tbl %>% filter(http_response != 0)),
    "nonindexed_endpoints" = nrow(fhir_endpoints_tbl %>% filter(http_response == 0))
)

# get the endpoint tally by http_response received 
curr_http_response_tally <- fhir_endpoints_tbl %>%
    select(http_response) %>%
    group_by(http_response) %>%
    tally()

# Get the list of most recent HTTP responses when requesting the capability statement from the 
# fhir_endpoints 
response_tally <- list(
    "http_200" = nrow(curr_http_response_tally %>% filter(http_response==200)),
    "http_404" = nrow(curr_http_response_tally %>% filter(http_response==404)),
    "http_503" = nrow(curr_http_response_tally %>% filter(http_response==503))
)

# we want the current set of http response codes from the endpoint monitoring
# first get the entries from the metrics_labels table for http_request_responses
http_response_ids <- metrics_labels %>%
    filter(metric_name == "AllEndpoints_http_request_responses") %>%
    select(id)

# next, right_join against the value for each endpoint
http_response_values <- metrics_values %>%
    right_join(http_response_ids, by = c("labels_id" = "id"))

# Compute the percentage of each response code for all responses received
http_pct <- as_tibble(http_response_values %>% 
                          mutate(code=as.character(value)) %>%
                          group_by(labels_id,code,value) %>% 
                          summarise(Percentage=n()) %>% 
                          group_by(labels_id) %>% 
                          mutate(Percentage=Percentage/sum(Percentage,na.rm = TRUE)*100)
)
# we want to graph all non-200 results by response code, but they need to be factors
# so they can be shown as separate categories on the graph, rather than as a scalar value
http_pctf <- http_pct %>% filter(value != 200) %>% mutate(name=as.factor(labels_id), code=as.factor(code)) 

# create a summary table to show the response codes received along with 
# the description for each code
http_summary <- http_pct %>%
    left_join(http_response_code_tbl, by=c("code" = "code_chr")) %>%
    select(code,label) %>%
    group_by("HTTP Response" = code,"Status"=label) %>%
    summarise(Count=n()) 

# Get the FHIR version for each endpoint
fhir_version_tbl <- as_tibble(tbl(con,sql("select id,url,vendor,capability_statement->>'fhirVersion' as FHIR from fhir_endpoints where capability_statement->>'fhirVersion' IS NOT NULL")))

# Get the count of endpoints by vendor, and use "Unknown" for any entries
# where the vendor field is empty
fhir_version_vendor_count <- fhir_version_tbl %>%
    mutate(vendor = na_if(vendor,"")) %>%
    tidyr::replace_na(list(vendor="Unknown")) %>%
    group_by(vendor,fhir) %>%
    tally() %>%
    select(Vendor=vendor,"FHIR Version"=fhir,"Count"=n)

# Get the list of distinct fhir versions for use in filtering
fhir_version_list <- as.list(fhir_version_tbl %>% distinct("FHIR Version"=fhir))

# Get the list of distinct vendors for use in filtering
vendor_list <- as.list(as_tibble(fhir_endpoints %>% distinct(vendor)) %>% mutate(vendor = na_if(vendor,"")) %>% tidyr::replace_na(list(vendor="Unknown")) %>% pull(vendor))

# get time series of response time metrics for all endpoints
# will update with dynamic time ranges, group by 4 minute intervals
all_endpoints_response_time <- as_tibble(tbl(con,sql("SELECT floor(extract(epoch from metrics_values.time)/240)*240 AS time, AVG(metrics_values.value) 
FROM metrics_labels, metrics_values
WHERE metrics_labels.metric_name = 'AllEndpoints_http_response_time' 
AND metrics_labels.id = metrics_values.labels_id
AND metrics_values.time BETWEEN '2020-01-01T00:00:00Z' AND '2020-07-01T00:00:00Z'
GROUP BY time
ORDER BY time ")) ) %>% mutate(date=as_datetime(time)) %>% select(date,avg)

# convert to xts format for use in dygraph
response_time_xts <- xts(x = all_endpoints_response_time$avg, order.by = all_endpoints_response_time$date)

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
                     menuSubItem("Mean Response Time", tabName = "subitem1")
            ),
            menuItem("Location", tabName = "location", icon=icon("map")),
            menuItem("About Lantern",tabName = "about", icon=icon("info-circle")),
            hr(),
            p("For future use..."),
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
                    h1("Current Endpoint Metrics"),
                    fluidRow(
                        infoBoxOutput("total_endpoints_box",width=4),
                        infoBoxOutput("indexed_endpoints_box",width=4),
                        infoBoxOutput("nonindexed_endpoints_box",width=4)
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
                    dygraphOutput("mean_response_time_plot"),
                    p("Click and drag on plot to zoom in, double-click to zoom out. Will add more time-series charting features here...")
            ),
            tabItem("availability",
                    plotlyOutput("non_200"),
                    htmlOutput("count_200_sub"),
                    plotlyOutput("plot_200_sub")
                    
            ),
            tabItem("location",
                    h3("Map of Zip Codes with identified endpoint/organization"),
                    img(src="images/endpoint_zcta_map.png",width="100%"),
                    p("This is a placeholder map for showing endpoints associated with a location.
                      Will be updated with interactive map with pins for endpoints")
            ),
            tabItem("about",
                    h1("About Lantern"),
                    img(src="images/lantern-logo@1x.png",width="300px"),
                    p("This is a description of Lantern, the dashboard, the project, etc. "))
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
    output$count_200_sub <- renderText({
        count_200_sub  <- nrow(http_pct %>% filter(value==200,Percentage < 99.8))
        paste("<br><p>There are",count_200_sub,"endpoints which have returned HTTP 200 (Success) responses less than <strong>99.8%</strong> of the time.</p>")
    })
    output$plot_200_sub <- renderPlotly({
        http_200 <- http_pct %>% filter(value == 200,Percentage < 99.8) %>% arrange(Percentage) %>% mutate(name=as.factor(labels_id))
        http_200f <- http_200 %>% mutate(name = forcats::fct_reorder(name,Percentage))
        g200 <- ggplot(http_200f,aes(x=name,y=Percentage))
        g200 + geom_bar(stat="identity",fill="#DD8888", width=0.9 ) + 
            coord_cartesian(ylim = c(0, 100)) +
            ggtitle("HTTP 200 Responses\nFor endpoints less than 99.8% success") +
            labs(y="Percentage of Responses",x = "Endpoint ID")
    })
    output$total_endpoints_box <- renderInfoBox({
        infoBox(
            "Total Endpoints", fhir_endpoint_totals$all_endpoints, icon = icon("fire", lib = "glyphicon"),
            color = "blue"
        )
    })
    output$mean_response_time_plot <- renderDygraph({
        dygraph(response_time_xts,main="Endpoint Mean Response Time",ylab="seconds",xlab="Date")
    })
    output$indexed_endpoints_box <- renderInfoBox({
        infoBox(
            "Indexed Endpoints", fhir_endpoint_totals$indexed_endpoints, icon = icon("flash", lib = "glyphicon"),
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
}

# Run the application 
shinyApp(ui = ui, server = server)
