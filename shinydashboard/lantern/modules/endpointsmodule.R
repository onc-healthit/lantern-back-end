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
    DT::dataTableOutput(ns("endpoints_table"))
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
    req(sel_fhir_version(),sel_vendor())
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
}
