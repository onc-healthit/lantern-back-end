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
    mutate(url = paste0("<a onclick=\"Shiny.setInputValue(\'endpoint_popup\',&quot;", url, "&&", "None", "&quot,{priority: \'event\'});\">", url, "</a>")) %>%
    mutate_all(as.character)

    res
  })

  output$profile_table <- renderUI({
    if (length(sel_profile()) > 0) {
      if (sel_profile() != ui_special_values$ALL_PROFILES) {
        tagList(
          reactable::reactableOutput(ns("filter_profile_table"))
        )
      }
     else {
        tagList(
            reactable::reactableOutput(ns("no_filter_profile_table"))
        )
     }
    }
  })

  output$filter_profile_table <- reactable::renderReactable({
     reactable(
              selected_fhir_endpoint_profiles(),
              defaultColDef = colDef(
                align = "center"
              ),
              columns = list(
                  url = colDef(name = "Endpoint", minWidth = 300, sortable = TRUE, align = "left", html = TRUE),
                  profileurl = colDef(name = "Profile URL", minWidth = 300, align = "left", sortable = FALSE),
                  profilename = colDef(name = "Profile Name", minWidth = 200, sortable = FALSE),
                  resource = colDef(name = "Resource", minWidth = 200, sortable = FALSE),
                  fhir_version = colDef(name = "FHIR Version", sortable = FALSE),
                  vendor_name = colDef(name = "Certified API Developer Name", minWidth = 110, sortable = FALSE)
              ),
              striped = TRUE,
              searchable = TRUE,
              showSortIcon = TRUE,
              highlight = TRUE,
              defaultPageSize = 10

     )
  })

  output$no_filter_profile_table <- reactable::renderReactable({
     reactable(
              selected_fhir_endpoint_profiles(),
              defaultColDef = colDef(
                align = "center"
              ),
              columns = list(
                  url = colDef(name = "Endpoint", minWidth = 300, sortable = TRUE, align = "left", html = TRUE),
                  profileurl = colDef(name = "Profile URL", minWidth = 300, align = "left", sortable = FALSE, aggregate = "count",
                  format = list(aggregated = colFormat(prefix = "Count: "))),
                  profilename = colDef(name = "Profile Name", minWidth = 200, sortable = FALSE),
                  resource = colDef(name = "Resource", sortable = FALSE),
                  fhir_version = colDef(name = "FHIR Version", sortable = FALSE, aggregate = "unique"),
                  vendor_name = colDef(name = "Certified API Developer Name", minWidth = 110, sortable = FALSE)
              ),
              groupBy = "url",
              striped = TRUE,
              searchable = TRUE,
              showSortIcon = TRUE,
              highlight = TRUE,
              defaultPageSize = 10

     )
  })
}
