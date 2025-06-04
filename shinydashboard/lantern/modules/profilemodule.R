library(DT)
library(purrr)
library(reactable)

profilemodule_UI <- function(id) {
  ns <- NS(id)
  tagList(
    tags$style(HTML("
      div.dataTables_filter {
        display: none;
      }
    ")),
    DTOutput("filter_profile_table")
  )
}

profilemodule <- function(
  input,
  output,
  session,
  sel_fhir_version,
  sel_vendor,
  sel_resource,
  sel_profile
) {
  ns <- session$ns
  
}