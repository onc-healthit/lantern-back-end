development_banner <- function(devbanner) {
  if (all(devbanner != "")) {
    fluidRow(column(12, tags$p(devbanner, role = "banner"), style = "background-color: yellow; line-height: 50px; margin-top:-1em; font-size: 20px"))
  }
}
