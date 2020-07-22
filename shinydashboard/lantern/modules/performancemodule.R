library(dygraphs)
library(xts)

performance_UI <- function(id) {

  ns <- NS(id)

  tagList(
    dygraphOutput(ns("mean_response_time_plot")),
    p("Click and drag on plot to zoom in, double-click to zoom out.")
  )
}

performance <- function(
    input,
    output,
    session, 
    sel_date
) {
  ns <- session$ns

  response_time_xts <- reactive({
      if (all(sel_date() == "Past 7 days")){
        range <- "604800"
      }
      else if (all(sel_date() == "Past 14 days")){
        range <- "1209600"
      }
      else if (all(sel_date() == "Past 30 days")){
        range <- "2592000"
      }
      else{
        range <- "maxdate.maximum"
      }
      res <- get_avg_response_time(db_connection, range)
      res
    })
  
  if (nrow(app_data$avg_response_time)== 0){}

  else{
    output$mean_response_time_plot <- renderDygraph({
      dygraph(response_time_xts(),
              main = "Endpoint Mean Response Time",
              ylab = "seconds",
              xlab = "Date")
    })
  }
}
