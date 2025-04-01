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

selected_fhir_endpoint_profiles <- reactive({
    res <- isolate(app_data$supported_profiles())
    req(sel_fhir_version(), sel_vendor())

    res <- res %>% filter(fhir_version %in% sel_fhir_version())

    if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
      res <- res %>% filter(vendor_name == sel_vendor())
    }

    if (length(sel_resource()) > 0) {
        if (sel_resource() != ui_special_values$ALL_RESOURCES) {
        res <- res %>% filter(resource == sel_resource())
        }
    }

    if (length(sel_profile()) > 0) {
        if (sel_profile() != ui_special_values$ALL_PROFILES) {
        res <- res %>% filter(profileurl == sel_profile())
        }
    }

    res <- res %>%
    distinct(url, profileurl, profilename, resource, fhir_version, vendor_name) %>%
    select(url, profileurl, profilename, resource, fhir_version, vendor_name) %>%
    group_by(url) %>%
    mutate(url = paste0("<a class=\"lantern-url\" tabindex=\"0\" aria-label=\"Press enter to open pop up modal containing additional information for this endpoint.\" onkeydown = \"javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)\" onclick=\"Shiny.setInputValue(\'endpoint_popup\',&quot;", url, "&&", "None", "&quot,{priority: \'event\'});\">", url, "</a>")) %>%
    mutate_all(as.character)

    res
  })

  output$profile_table <- renderUI({
    if (length(sel_profile()) > 0) {
        tagList(
            DT::dataTableOutput("filter_profile_table")
        )
    }
  })

}
