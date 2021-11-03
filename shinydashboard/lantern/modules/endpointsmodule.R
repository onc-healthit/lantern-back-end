library(DT)
library(purrr)
library(gt)

endpointsmodule_UI <- function(id) {

  ns <- NS(id)

  tagList(
    fluidRow(
      column(width = 12, style = "padding-bottom:20px",
             h3(style = "margin-top:0", textOutput(ns("endpoint_count"))),
             downloadButton(ns("download_data"), "Download Endpoint Data (CSV)"),
             downloadButton(ns("download_descriptions"), "Download Field Descriptions (CSV)")
      ),
    ),
    gt::gt_output(ns("endpoints_table")),
    htmlOutput(ns("note_text"))
  )
}

endpointsmodule <- function(
  input,
  output,
  session,
  sel_fhir_version,
  sel_vendor,
  sel_availability
) {
  ns <- session$ns

  output$endpoint_count <- renderText({
    paste("Matching Endpoints:", nrow(selected_fhir_endpoints()))
  })

  selected_fhir_endpoints <- reactive({
    res <- get_fhir_endpoints_tbl()
    req(sel_fhir_version(), sel_vendor())
    if (sel_fhir_version() != ui_special_values$ALL_FHIR_VERSIONS) {
      res <- res %>% filter(fhir_version == sel_fhir_version())
    }
    if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
      res <- res %>% filter(vendor_name == sel_vendor())
    }
    if (sel_availability() != "0-100") {
      if (sel_availability() == "0" || sel_availability() == "100") {
        availability_filter_num <- as.numeric(sel_availability()) / 100
        availability_filter <- as.character(availability_filter_num)
        res <- res %>% filter(availability == availability_filter)
      }
      else {
        availability_upper_num <- as.numeric(strsplit(sel_availability(), "-")[[1]][2]) / 100
        availability_lower_num <- as.numeric(strsplit(sel_availability(), "-")[[1]][1]) / 100
        availability_lower <- as.character(availability_lower_num)
        availability_upper <- as.character(availability_upper_num)

        res <- res %>% filter(availability >= availability_lower, availability <= availability_upper)
      }
    }
    res <- res %>% mutate(availability = availability * 100)
    res
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

#  output$endpoints_table <- reactable::renderReactable({
#     reactable(
#              selected_fhir_endpoints() %>% select(url, endpoint_names, updated, vendor_name, fhir_version, tls_version, mime_types, status, availability) %>% mutate_all(as.character),
#              groupBy ="url",
#              columns = list(
#                  url = colDef(name = "URL", minWidth = 300),
#                  endpoint_names = colDef(name = "API Information Source Name"),
#                  updated = colDef(name = "Updated"),
#                  vendor_name = colDef(name = "Certified API Developer Name"),
#                  fhir_version = colDef(name = "FHIR Version"),
#                  tls_version = colDef(name = "TLS Version"),
#                  mime_types = colDef(name = "MIME Types", minWidth = 150),
#                  status = colDef(name = "HTTP Response"),
#                  availability = colDef(name = "Availability")
#              ),
#              sortable = TRUE,
#              searchable = TRUE,
#              striped = TRUE,
#              showSortIcon = TRUE,
#              defaultPageSize = 5
#
#     )
#  })

  output$endpoints_table <-
    gt::render_gt(
      expr = selected_fhir_endpoints() %>% select(url, endpoint_names, updated, vendor_name, fhir_version, tls_version, mime_types, status, availability) %>% 
      gt(id = "endpoints_table", groupname_col = "url", rowname_col = "fhir_version") %>% 
      tab_stubhead(label = "FHIR Version") %>% 
      cols_label(
        endpoint_names = md("API Information Source Name"),
        updated = md("Updated"),
        vendor_name = md("Certified API Developer Name"),
        tls_version = md("TLS Version"),
        mime_types = md("MIME Types"),
        status = md("HTTP Response"),
        availability = md("Availability")
      ) %>% 
      tab_options(row_group.background.color = "#c6dbef", column_labels.font.weight = "bold", row_group.font.weight = "lighter", container.overflow.y = TRUE),
      height = "600px",
      width = "100%"
    )




  # Create the format for the csv
  csv_format <- reactive({
    res <- selected_fhir_endpoints() %>%
      select(-updated, -label, -status, -availability) %>%
      rename(api_information_source_name = endpoint_names, certified_api_developer_name = vendor_name) %>%
      rename(created_at = info_created, updated = info_updated) %>%
      rename(http_response_time_second = response_time_seconds)
  })

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
