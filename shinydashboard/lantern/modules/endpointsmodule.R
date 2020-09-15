library(DT)
library(purrr)

endpointsmodule_UI <- function(id) {

  ns <- NS(id)

  tagList(
    fluidRow(
      column(width = 12, style = "padding-bottom:20px",
             h3(style = "margin-top:0", textOutput(ns("endpoint_count"))),
             downloadButton(ns("download_data"), "Download")
      ),
    ),
    DT::dataTableOutput(ns("endpoints_table")),
    htmlOutput(ns("note_text"))
  )
}

endpointsmodule <- function(
  input,
  output,
  session,
  sel_fhir_version,
  sel_vendor
) {
  ns <- session$ns

  output$endpoint_count <- renderText({
    paste("Matching Endpoints:", nrow(selected_fhir_endpoints()))
  })

  selected_fhir_endpoints <- reactive({
    res <- get_fhir_endpoints_tbl(db_tables) %>% select(-http_response, -label)
    req(sel_fhir_version(), sel_vendor())
    if (sel_fhir_version() != ui_special_values$ALL_FHIR_VERSIONS) {
      res <- res %>% filter(fhir_version == sel_fhir_version())
    }
    if (sel_vendor() != ui_special_values$ALL_VENDORS) {
      res <- res %>% filter(vendor_name == sel_vendor())
    }
    res
  })

  output$endpoints_table <- DT::renderDataTable({
    datatable(selected_fhir_endpoints() %>% select(-supported_resources),
              colnames = c("URL", "Organization", "Updated", "Vendor", "FHIR Version", "TLS Version", "MIME Types", "Status"),
              rownames = FALSE,
              options = list(scrollX = TRUE)
    )
    })
  # Downloadable csv of selected dataset ----
  output$download_data <- downloadHandler(
    filename = function() {
      "fhir_endpoints.csv"
    },
    content = function(file) {
      write.csv(selected_fhir_endpoints(), file, row.names = FALSE)
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
