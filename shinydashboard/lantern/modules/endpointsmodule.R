library(DT)
library(purrr)

endpointsmodule_UI <- function(id) {

  fhir_version_list <- get_fhir_version_list(endpoint_export_tbl)
  vendor_list <- get_vendor_list(endpoint_export_tbl)

  ns <- NS(id)

  tagList(
    fluidRow(
      column(width = 4, style = "padding-bottom:20px",
             h3(style = "margin-top:0", textOutput(ns("endpoint_count"))),
             downloadButton(ns("download_data"), "Download")
      ),
      column(width = 4,
             selectInput(
               inputId = ns("fhir_version"),
               label = "FHIR Version:",
               choices = fhir_version_list,
               selected = ui_special_values$ALL_FHIR_VERSIONS,
               size = 1,
               selectize = FALSE)
      ),
      column(width = 4,
             selectInput(
               inputId = ns("vendor"),
               label = "Vendor:",
               choices = vendor_list,
               selected = ui_special_values$ALL_VENDORS,
               size = 1,
               selectize = FALSE)
      )
    ),
    DT::dataTableOutput(ns("endpoints_table"))
  )
}

endpointsmodule <- function(
  input,
  output,
  session
) {
  ns <- session$ns

  output$endpoint_count <- renderText({
    paste("Matching Endpoints:", nrow(selected_fhir_endpoints()))
  })

  selected_fhir_endpoints <- reactive({
    res <- get_fhir_endpoints_tbl(db_tables) %>% select(-http_response, -label)
    req(input$fhir_version, input$vendor)
    if (input$fhir_version != ui_special_values$ALL_FHIR_VERSIONS) {
      res <- res %>% filter(fhir_version == input$fhir_version)
    }
    if (input$vendor != ui_special_values$ALL_VENDORS) {
      res <- res %>% filter(vendor_name == input$vendor)
    }
    res
  })

  output$endpoints_table <- DT::renderDataTable({
    datatable(selected_fhir_endpoints(),
              colnames = c("URL", "Organization", "Updated", "Vendor", "FHIR Version", "TLS Version", "Status"),
              rownames = FALSE
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
