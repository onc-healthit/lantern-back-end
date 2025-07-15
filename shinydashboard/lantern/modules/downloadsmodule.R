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
              downloadButton(ns("organizations_download_data"), "Download Organization Data (CSV)", icon = tags$i(class = "fa fa-download", "aria-hidden" = "true", role = "presentation", "aria-label" = "download icon")),
              downloadButton(ns("organizations_download_descriptions"), "Download Organization Field Descriptions (CSV)", icon = tags$i(class = "fa fa-download", "aria-hidden" = "true", role = "presentation", "aria-label" = "download icon"))
      )
    ),
    fluidRow(
      column(width = 12,
            h2("REST API"),
            style = "padding-bottom:10px;padding-top:10px",
            p(HTML("These REST APIs enable programmatic access to download daily Lantern data in CSV format:
                  <br><br>
                  [GET]<b> https://lantern.healthit.gov/api/daily/download </b> – Downloads daily FHIR endpoint data.
                  <br>
                  [GET]<b> https://lantern.healthit.gov/api/organizations/v1 </b> – Downloads daily organization data associated with endpoints.
                  <br><br>
                  The organizations API supports the following optional query parameters for filtering results:
                  <br>
                  <code>developer</code> – Filter by certified API developer name. <br>
                  <code>fhir_version</code> – Comma-separated list of FHIR versions to include. <br>
                  <code>identifier</code> – Exact match on organization identifier (e.g., NPI). <br>
                  <code>hti1</code> – Use <code>hti1=present</code> to return only organizations with HTI-1 relevant data.
                  <br><br>
                  All filters can be used independently or in combination.
                  <br>
                  For example, to download data only for <i>Epic Systems Corporation</i> and FHIR version <i>No Cap Stat</i> or <i>4.0.1</i>, use:
                  <br>
                  <code>?developer=Epic%20Systems%20Corporation&fhir_version=No%20Cap%20Stat,4.0.1</code>
                  <br><br>
                  To return organizations with identifier <i>1750581864</i> that have HTI-1 data:
                  <br>
                  <code>?identifier=1750581864&hti1=present</code>
                  <br><br>
                  Developer names and other parameter values must match exactly as stored in the system.
                  If the value contains spaces, commas, or other special characters, it must be 
                  <a href='https://en.wikipedia.org/wiki/Percent-encoding' target='_blank'>URL encoded</a>.
                  Most browsers handle this automatically, but other tools may require manual encoding.
                  <br><br>
                  The APIs will initiate download of the data in CSV format automatically. 
                  These can be used to programmatically retrieve data for analysis or integration."))
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
      mutate(endpoint_names = ifelse(length(strsplit(endpoint_names, ";")[[1]]) > 100, paste0("Subset of Organizations, see Lantern Website for full list:", paste0(head(strsplit(endpoint_names, ";")[[1]], 100), collapse = ";")), endpoint_names)) %>%
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

  # Create the format for the csv
  organization_csv_format <- reactive({
    res <- get_endpoint_list_matches(db_connection)

  res <- res %>%
    mutate(organization_id = as.integer(organization_id)) %>%
    
    # Left join with deduplicated or collapsed identifiers
    left_join(
      get_org_identifiers_information(db_connection) %>%
        mutate(org_id = as.integer(org_id)) %>%
        group_by(org_id) %>%
        summarise(identifier = paste(unique(identifier), collapse = "\n")),
      by = c("organization_id" = "org_id")
    ) %>%
    
    # Left join with deduplicated or collapsed addresses
    left_join(
      get_org_addresses_information(db_connection) %>%
        mutate(org_id = as.integer(org_id)) %>%
        group_by(org_id) %>%
        summarise(address = paste(unique(address), collapse = "\n")),
      by = c("organization_id" = "org_id")
    ) %>%
    
    left_join(get_org_url_information(db_connection),
          by = c("organization_id" = "org_id")) %>%

    mutate(org_url = if_else(str_starts(org_url, "urn:uuid:"), "", org_url)) %>%

    select(-organization_id)

    res <- res %>%
      group_by(organization_name) %>%
      summarise(
        identifier = paste(unique(identifier), collapse = "\n"),
        address = paste(unique(address), collapse = "\n"),
        org_url = paste(unique(org_url), collapse = "\n"),
        fhir_version = paste(unique(fhir_version), collapse = "\n"),
        vendor_name = paste(unique(vendor_name), collapse = "\n"),
        .groups = "drop"
      ) %>%
      filter(organization_name != "Unknown") %>%
      mutate(address = toupper(address)) %>%
      arrange(organization_name)

    res
  })

  # Downloadable csv of selected dataset
  output$organizations_download_data <- downloadHandler(
    filename = function() {
      "fhir_endpoint_organizations.csv"
    },
    content = function(file) {
      write.csv(organization_csv_format(), file, row.names = FALSE)
    }
  )

  # Download csv of the field descriptions in the dataset csv
  output$organizations_download_descriptions <- downloadHandler(
    filename = function() {
      "fhir_endpoint_organizations_fields.csv"
    },
    content = function(file) {
      file.copy("fhir_endpoint_organizations_fields.csv", file)
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