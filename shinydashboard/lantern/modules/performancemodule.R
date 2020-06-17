library(dygraphs)
library(xts)

performance_UI <- function(id) {

  ns <- NS(id)

  tagList(
    dygraphOutput(ns("mean_response_time_plot")),
    p("Click and drag on plot to zoom in, double-click to zoom out. Will add more time-series charting features here...")
  )
}

performance <- function(
    input,
    output,
    session
) {
  ns <- session$ns

  response_time_xts <- get_avg_response_time()

  output$mean_response_time_plot <- renderDygraph({
    dygraph(response_time_xts,
            main = "Endpoint Mean Response Time",
            ylab = "seconds",
            xlab = "Date")
  })
}
