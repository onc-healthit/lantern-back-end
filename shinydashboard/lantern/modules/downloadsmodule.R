library(DT)
library(purrr)

downloadsmodule_UI <- function(id) {

  ns <- NS(id)

  tagList(
    fluidRow(
      column(width = 12, style = "padding-bottom:20px",
             p("These files include the endpoint data in CSV and JSON format. Note: The CSV file only includes the most recent query of the data, while the JSON file includes historic data.")
      )
    ),
    fluidRow(
      column(width = 12,
             h3("CSV Downloads"),
             downloadButton(ns("download_data"), "Download Endpoint Data"),
             downloadButton(ns("download_descriptions"), "Download Field Descriptions"),
      )
    ),
    fluidRow(
      column(width = 12,
             h3("JSON Downloads"),
             downloadButton(ns("download_data_json"), "Download Endpoint Data"),
             downloadButton(ns("download_descriptions_markdown"), "Download Field Descriptions"),
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

  # Create the format for the csv
  csv_format <- reactive({
    res <- get_fhir_endpoints_tbl() %>%
      select(-supported_resources, -updated, -label, -status) %>%
      rename(api_information_source_name = endpoint_names, certified_api_developer_name = vendor_name) %>%
      rename(created_at = info_created, updated = info_updated) %>%
      rename(http_response_time_second = response_time_seconds)
  })

  # Downloadable csv of selected dataset
  output$download_data <- downloadHandler(
    filename = function() {
      "fhir_endpoints.csv"
    },
    content = function(file) {
      write.csv(csv_format(), file, row.names = FALSE)
    }
  )

  # Download csv of the field descriptions in the dataset csv
  output$download_descriptions <- downloadHandler(
    filename = function() {
      "fhir_endpoints_fields.csv"
    },
    content = function(file) {
      file.copy("fhir_endpoints_fields.csv", file)
    }
  )

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
