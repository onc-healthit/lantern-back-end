library(plotly)

availability_UI <- function(id) {

  ns <- NS(id)

  tagList(
    plotlyOutput(ns("non_200")),
    htmlOutput(ns("count_200_sub")),
    plotlyOutput(ns("plot_200_sub"))
  )
}

availability <- function(
    input, 
    output, 
    session
){
  ns <- session$ns

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

}
