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

# selected_fhir_endpoint_profiles <- reactive({
#     res <- get_supported_profiles(db_connection)
#     req(input$fhir_version, input$vendor)

#     res <- res %>% filter(fhir_version %in% input$fhir_version)

#     if (input$vendor != ui_special_values$ALL_DEVELOPERS) {
#       res <- res %>% filter(vendor_name == input$vendor)
#     }

#      if (length(input$profile_resource) > 0) {
#         if (input$profile_resource != ui_special_values$ALL_RESOURCES) {
#           res <- res %>% filter(resource == input$profile_resource)
#         }
#     }

#     if (length(input$profiles) > 0) {
#         if (input$profiles != ui_special_values$ALL_PROFILES) {
#         res <- res %>% filter(profileurl == input$profiles)
#         }
#     }

#     res <- res %>%
#     distinct(url, profileurl, profilename, resource, fhir_version, vendor_name) %>%
#     select(url, profileurl, profilename, resource, fhir_version, vendor_name) %>%
#     group_by(url) %>%
#     mutate(url = paste0("<a class=\"lantern-url\" tabindex=\"0\" aria-label=\"Press enter to open pop up modal containing additional information for this endpoint.\" onkeydown = \"javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)\" onclick=\"Shiny.setInputValue(\'endpoint_popup\',&quot;", url, "&&", "None", "&quot,{priority: \'event\'});\">", url, "</a>")) %>%
#     mutate_at(vars(-group_cols()), as.character)

#     return(res)
#   })

  output$profile_table <- renderUI({
    if (length(sel_profile()) > 0) {
        tagList(
            DT::dataTableOutput("filter_profile_table")
        )
    }
  })
}