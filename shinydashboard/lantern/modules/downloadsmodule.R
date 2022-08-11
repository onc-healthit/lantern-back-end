library(DT)
library(purrr)

downloadsmodule_UI <- function(id) {

  ns <- NS(id)

  tagList(
    fluidRow(
      column(width = 12, style = "padding-bottom:20px",
             p("These files include the endpoint data over time in JSON format.")
      )
    ),
    fluidRow(
      column(width = 12,
             h2("JSON Downloads"),
             downloadButton(ns("download_data_json"), "Download Endpoint Data"),
             downloadButton(ns("download_descriptions_markdown"), "Download Field Descriptions"),
      ),
      column(width = 12,
            p("Formerly, the json export file included all data, but now only includes the past 30 days. To see export files for previous months created by Lantern, visit the repository ",
            a("available here.", href = "https://github.com/onc-healthit/lantern-back-end", target = "_blank"))
      )
    ),
    fluidRow(
      column(width = 12, style = "padding-top:50px",
             htmlOutput(ns("note_text"))
      )
    )
  )
}

downloadsmodule <- function(
  input,
  output,
  session
) {
  ns <- session$ns

  output$download_data_json <- downloadHandler(
    filename = function() {
      "fhir_endpoints.json"
    },
    content = function(file) {
      file.copy("/srv/shiny-server/exportfolder/fhir_endpoints_fields.json", file)
    }
  )

  output$download_descriptions_markdown <- downloadHandler(
    filename = function() {
      "fhir_endpoints_fields_json.md"
    },
    content = function(file) {
      file.copy("fhir_endpoints_fields_json.md", file)
    }
  )

  output$note_text <- renderUI({
    note_info <- "The endpoints queried by Lantern are limited to Fast Healthcare Interoperability
      Resources (FHIR) endpoints published publicly by Certified API Developers in conformance
      with the ONC Cures Act Final Rule, or discovered through the National Plan and Provider
      Enumeration System (NPPES). This data, therefore, may not represent all FHIR endpoints
      in existence. Insights gathered from this data should be framed accordingly."
    res <- paste("<div style='font-size: 18px;'><b>Note:</b>", note_info, "</div>")
    HTML(res)
  })

}
