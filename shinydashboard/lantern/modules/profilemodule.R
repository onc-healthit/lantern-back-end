library(DT)
library(purrr)
library(reactable)

profilemodule_UI <- function(id) {

  ns <- NS(id)

  tagList(
    uiOutput(ns("profile_table"))
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

  output$profile_table <- renderUI({
    if (length(sel_profile()) > 0) {
        tagList(
            DT::dataTableOutput("filter_profile_table")
        )
    }
  })
}