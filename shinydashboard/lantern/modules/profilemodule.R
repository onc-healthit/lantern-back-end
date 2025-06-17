library(DT)
library(purrr)
library(reactable)
library(glue)

profilemodule_UI <- function(id) {
  ns <- NS(id)
  tagList(
    reactable::reactableOutput(ns("profiles_table")),
    fluidRow(
      column(6, 
        div(style = "display: flex; justify-content: flex-start;", 
            uiOutput(ns("profile_prev_button_ui"))
        )
      ),
      column(6, 
        div(style = "display: flex; justify-content: flex-end;",
            uiOutput(ns("profile_next_button_ui"))
        )
      )
    )
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

  profile_page_state <- reactiveVal(1)
  profile_page_size <- 10

  # Handle next page button
  observeEvent(input$profile_next_page, {
    new_page <- profile_page_state() + 1
    profile_page_state(new_page)
  })

  # Handle previous page button
  observeEvent(input$profile_prev_page, {
    if (profile_page_state() > 1) {
      new_page <- profile_page_state() - 1
      profile_page_state(new_page)
    }
  })

  # Reset to first page on any filter change
  observeEvent(list(sel_fhir_version(), sel_vendor(), sel_resource(), sel_profile()), {
    profile_page_state(1)
  })

  output$profile_prev_button_ui <- renderUI({
    if (profile_page_state() > 1) {
      actionButton(ns("profile_prev_page"), "Previous", icon = icon("arrow-left"))
    } else {
      NULL  # Hide the button
    }
  })

  output$profile_next_button_ui <- renderUI({
    # Always show next button - let the database handle empty results
    actionButton(ns("profile_next_page"), "Next", icon = icon("arrow-right"))
  })

  # Main data query with LIMIT OFFSET pagination
  selected_fhir_endpoint_profiles <- reactive({
    req(sel_fhir_version(), sel_vendor())
    
    profile_offset <- (profile_page_state() - 1) * profile_page_size
    
    query_str <- "SELECT DISTINCT url, profileurl, profilename, resource, fhir_version, vendor_name FROM endpoint_supported_profiles_mv WHERE fhir_version IN ({vals*})"
    params <- list(vals = sel_fhir_version())

    if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
      query_str <- paste0(query_str, " AND vendor_name = {vendor}")
      params$vendor <- sel_vendor()
    }

    if (length(sel_resource()) > 0) {
      if (sel_resource() != ui_special_values$ALL_RESOURCES) {
        query_str <- paste0(query_str, " AND resource = {resource}")
        params$resource <- sel_resource()
      }
    }

    if (length(sel_profile()) > 0) {
      if (sel_profile() != ui_special_values$ALL_PROFILES) {
        query_str <- paste0(query_str, " AND profileurl = {profile}")
        params$profile <- sel_profile()
      }
    }

    # Add LIMIT OFFSET for pagination
    query_str <- paste0(query_str, " LIMIT {limit} OFFSET {offset}")
    params$limit <- profile_page_size
    params$offset <- profile_offset

    query <- do.call(glue_sql, c(list(query_str, .con = db_connection), params))
    res <- tbl(db_connection, sql(query)) %>% collect()
    
    res <- res %>%
      group_by(url) %>%
      mutate(url = paste0("<a class=\"lantern-url\" tabindex=\"0\" aria-label=\"Press enter to open pop up modal containing additional information for this endpoint.\" onkeydown = \"javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)\" onclick=\"Shiny.setInputValue(\'endpoint_popup\',&quot;", url, "&&", "None", "&quot,{priority: \'event\'});\">", url, "</a>")) %>%
      mutate_at(vars(-group_cols()), as.character)

    return(res)
  })

  output$profiles_table <- reactable::renderReactable({
    df <- selected_fhir_endpoint_profiles()

    if (nrow(df) == 0) {
      return(reactable(
        data.frame(Message = "No data matching the selected filters"),
        pagination = FALSE,
        searchable = FALSE
      ))
    }

    reactable(
      df,
      defaultColDef = colDef(
        align = "center"
      ),
      columns = list(
        url = colDef(name = "Endpoint", minWidth = 300, align = "left", html = TRUE, sortable = TRUE),
        profileurl = colDef(name = "Profile URL", minWidth = 250, sortable = TRUE),
        profilename = colDef(name = "Profile Name", minWidth = 200, sortable = TRUE),
        resource = colDef(name = "Resource", sortable = TRUE),
        fhir_version = colDef(name = "FHIR Version", sortable = TRUE),
        vendor_name = colDef(name = "Certified API Developer Name", minWidth = 110, sortable = TRUE)
      ),
      searchable = FALSE,
      showSortIcon = TRUE,
      highlight = TRUE,
      defaultPageSize = 10
    )
  })
}