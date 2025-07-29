library(DT)
library(purrr)

downloadsmodule_UI <- function(id) {

  ns <- NS(id)

  tagList(
    fluidRow(
      column(width = 12, style = "padding-bottom:20px",
             p("The files below include the current endpoint data found on the endpoints tab in the CSV format,
              and the endpoint tab table field descriptions in the CSV format.")
      )
    ),
    fluidRow(
      column(width = 12,
              h2("CSV Download"),
              downloadButton(ns("download_data"), "Download Endpoint Data (CSV)", icon = tags$i(class = "fa fa-download", "aria-hidden" = "true", role = "presentation", "aria-label" = "download icon")),
              downloadButton(ns("download_descriptions"), "Download Field Descriptions (CSV)", icon = tags$i(class = "fa fa-download", "aria-hidden" = "true", role = "presentation", "aria-label" = "download icon"))
      ),
      column(width = 12,
            p("To see export files for previous months created by Lantern, visit the repository ",
            a("available here.", href = "https://github.com/onc-healthit/onc-open-data/tree/main/lantern-daily-data", target = "_blank"))
      )
    ),
    fluidRow(
      column(width = 12,
             h2("REST API"),
             style = "padding-bottom:10px;padding-top:10px",
             p(HTML("This REST API [GET]<b> https://lantern.healthit.gov/api/daily/download </b> enables programmatic access
              to download the daily Lantern data (available for download as a CSV above). The API will initiate
              the download of the data in CSV format automatically. This can be used to program the
              download for any purpose."))
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

  # Downloadable csv of selected dataset
  output$download_data <- downloadHandler(
    filename = function() {
      "fhir_endpoints.csv"
    },
    content = function(file) {
      write.csv(csv_format(), file, row.names = FALSE)
    }
  )

  # Create the format for the csv
  csv_format <- reactive({
    res <- get_fhir_endpoints_tbl() %>%
      select(-status, -availability, -fhir_version) %>%
      rowwise() %>%
      mutate(endpoint_names = ifelse(length(strsplit(endpoint_names, ";")[[1]]) > 100, paste0("Subset of Organizations, see Lantern Website for full list:", paste0(head(strsplit(endpoint_names, ";")[[1]], 100), collapse = ";")), endpoint_names),
             info_created = format(info_created, "%m/%d/%y %H:%M"),
             info_updated = format(info_updated, "%m/%d/%y %H:%M")) %>%
      rename(api_information_source_name = endpoint_names, certified_api_developer_name = vendor_name) %>%
      rename(created_at = info_created, updated = info_updated) %>%
      rename(http_response_time_second = response_time_seconds)
  })

  # Download csv of the field descriptions in the dataset csv
  output$download_descriptions <- downloadHandler(
    filename = function() {
      "fhir_endpoints_fields.csv"
    },
    content = function(file) {
      file.copy("fhir_endpoints_fields.csv", file)
    }
  )

  output$note_text <- renderUI({
    note_info <- "The endpoints queried by Lantern are limited to Fast Healthcare Interoperability
      Resources (FHIR) endpoints published publicly by Certified API Developers in conformance
      with the ONC Cures Act Final Rule. This data, therefore, may not represent all FHIR endpoints
      in existence. Insights gathered from this data should be framed accordingly."
    res <- paste("<div style='font-size: 18px;'><b>Note:</b>", note_info, "</div>")
    HTML(res)
  })

}