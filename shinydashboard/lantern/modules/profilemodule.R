library(DT)
library(purrr)
library(reactable)
library(glue)
library(dplyr)
library(ggplot2)

profilemodule_UI <- function(id) {
  ns <- NS(id)
  tagList(
    # Simple CSS styling
    tags$head(
      tags$style(HTML("
        .profile-metric-box {
          background: white;
          border: 2px solid #e0e0e0;
          border-radius: 8px;
          padding: 20px;
          text-align: center;
          margin-bottom: 20px;
        }
        .profile-metric-value {
          font-size: 2em;
          font-weight: bold;
          color: #2c3e50;
          margin: 10px 0;
        }
        .profile-metric-label {
          font-size: 0.9em;
          color: #7f8c8d;
          text-transform: uppercase;
          letter-spacing: 1px;
        }
        .profile-section-header {
          background: #3498db;
          color: white;
          padding: 15px 20px;
          border-radius: 5px;
          margin: 30px 0 20px 0;
        }
        .profile-section-header h3 {
          margin: 0;
          font-size: 1.3em;
        }
        .profile-badge {
          display: inline-block;
          padding: 4px 10px;
          border-radius: 12px;
          font-size: 0.8em;
          font-weight: 600;
        }
        .badge-us-core {
          background-color: #d4edff;
          color: #0066cc;
        }
        .badge-other {
          background-color: #f0f0f0;
          color: #666666;
        }
        .info-text {
          background: #ecf0f1;
          padding: 15px;
          border-left: 4px solid #3498db;
          margin: 15px 0;
          border-radius: 4px;
        }
      "))
    ),
    
    # Page Title
    h2("FHIR Profile Support Analysis"),
    div(class = "info-text",
      p(strong("What are Profiles?"), " Profiles define the specific data structures that endpoints support. They determine what patient information can be accessed through each API."),
      p(strong("US Core Profiles:"), " These are mandated by ONC regulations for certified EHR systems. Higher US Core support means better regulatory compliance.")
    ),
    
    # Key Metrics Row
    fluidRow(
      column(3,
        div(class = "profile-metric-box",
          div(class = "profile-metric-label", "Total Endpoints"),
          div(class = "profile-metric-value", textOutput(ns("total_endpoints"), inline = TRUE))
        )
      ),
      column(3,
        div(class = "profile-metric-box",
          div(class = "profile-metric-label", "Unique Profiles"),
          div(class = "profile-metric-value", textOutput(ns("unique_profiles"), inline = TRUE))
        )
      ),
      column(3,
        div(class = "profile-metric-box",
          div(class = "profile-metric-label", "US Core Profiles"),
          div(class = "profile-metric-value", textOutput(ns("us_core_count"), inline = TRUE))
        )
      ),
      column(3,
        div(class = "profile-metric-box",
          div(class = "profile-metric-label", "Avg Profiles/Endpoint"),
          div(class = "profile-metric-value", textOutput(ns("avg_profiles"), inline = TRUE))
        )
      )
    ),
    
    # US Core vs Other
    div(class = "profile-section-header",
      h3("Profile Type Distribution")
    ),
    fluidRow(
      column(6,
        plotOutput(ns("profile_type_chart"), height = "250px")
      ),
      column(6,
        plotOutput(ns("resource_chart"), height = "250px")
      )
    ),
    
    # US Core Profile Support
    div(class = "profile-section-header",
      h3("US Core Profile Support")
    ),
    p("US Core profiles are required by the ONC Cures Act for certified EHR systems. This table shows which US Core profiles are supported by the filtered endpoints."),
    reactable::reactableOutput(ns("us_core_table")),
    
    # Detailed Profile Explorer
    div(class = "profile-section-header",
      h3("All Supported Profiles")
    ),
    fluidRow(
      column(4,
        selectInput(ns("profile_type_filter"), "Filter by Profile Type:",
          choices = c("All Profiles", "US Core Only", "Other Profiles"),
          selected = "All Profiles"
        )
      ),
      column(4,
        selectInput(ns("resource_dropdown"), "Filter by Resource:",
          choices = NULL,
          selected = "All Resources"
        )
      ),
      column(4,
        textInput(ns("profile_search_query"), "Search:", value = "")
      )
    ),
    reactable::reactableOutput(ns("profiles_table")),
    fluidRow(
      column(3, 
        div(style = "display: flex; justify-content: flex-start;", 
            uiOutput(ns("profile_prev_button_ui"))
        )
      ),
      column(6,
        div(style = "display: flex; justify-content: center; align-items: center; gap: 10px; margin-top: 8px;",
            numericInput(ns("profile_page_selector"), label = NULL, value = 1, min = 1, max = 1, step = 1, width = "80px"),
            textOutput(ns("profile_page_info"), inline = TRUE)
        )
      ),
      column(3, 
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
  current_request_id <- reactiveVal(0)
  
  # Simplified categorization: US Core vs Other
  categorize_profile <- function(profile_url) {
    ifelse(grepl("/us/core/StructureDefinition/", profile_url, fixed = TRUE), 
           "US Core", 
           "Other")
  }
  
  # Get all profile data for summaries (not paginated)
  all_profiles_data <- reactive({
    req(sel_fhir_version(), sel_vendor())
    
    base_query <- "SELECT url, profileurl, profilename, resource, fhir_version, vendor_name 
                   FROM mv_profiles_paginated 
                   WHERE 1=1"
    params <- list()
    
    base_query <- paste0(base_query, " AND fhir_version IN ({fhir_versions*})")
    params$fhir_versions <- sel_fhir_version()
    
    if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
      base_query <- paste0(base_query, " AND vendor_name = {vendor}")
      params$vendor <- sel_vendor()
    }
    
    if (length(sel_resource()) > 0 && sel_resource() != ui_special_values$ALL_RESOURCES) {
      base_query <- paste0(base_query, " AND resource = {resource}")
      params$resource <- sel_resource()
    }
    
    if (length(sel_profile()) > 0 && sel_profile() != ui_special_values$ALL_PROFILES) {
      base_query <- paste0(base_query, " AND profileurl = {profile}")
      params$profile <- sel_profile()
    }
    
    query <- do.call(glue_sql, c(list(base_query, .con = db_connection), params))
    result <- tbl(db_connection, sql(query)) %>% collect()
    
    if (nrow(result) > 0) {
      result <- result %>%
        mutate(profile_type = categorize_profile(profileurl))
    }
    
    result
  })
  
  # Update resource dropdown dynamically
  observe({
    data <- all_profiles_data()
    if (nrow(data) > 0) {
      resources <- c("All Resources", sort(unique(data$resource)))
      updateSelectInput(session, "resource_dropdown", choices = resources)
    }
  })
  
  # Metric: Total Endpoints
  output$total_endpoints <- renderText({
    data <- all_profiles_data()
    if (nrow(data) == 0) return("0")
    format(length(unique(data$url)), big.mark = ",")
  })
  
  # Metric: Unique Profiles
  output$unique_profiles <- renderText({
    data <- all_profiles_data()
    if (nrow(data) == 0) return("0")
    format(length(unique(data$profileurl)), big.mark = ",")
  })
  
  # Metric: US Core Count
  output$us_core_count <- renderText({
    data <- all_profiles_data()
    if (nrow(data) == 0) return("0")
    us_count <- data %>% 
      filter(profile_type == "US Core") %>%
      summarise(n = n_distinct(profileurl)) %>%
      pull(n)
    format(us_count, big.mark = ",")
  })
  
  # Metric: Average Profiles per Endpoint
  output$avg_profiles <- renderText({
    data <- all_profiles_data()
    if (nrow(data) == 0) return("0")
    avg <- data %>%
      group_by(url) %>%
      summarise(count = n()) %>%
      pull(count) %>%
      mean()
    format(round(avg, 1), nsmall = 1)
  })
  
  # Profile Type Chart (US Core vs Other)
  output$profile_type_chart <- renderPlot({
    data <- all_profiles_data()
    
    if (nrow(data) == 0) {
      plot.new()
      text(0.5, 0.5, "No data available", cex = 1.5, col = "gray")
      return()
    }
    
    type_counts <- data %>%
      group_by(profile_type) %>%
      summarise(count = n_distinct(profileurl)) %>%
      arrange(desc(count))
    
    colors <- c("US Core" = "#0066cc", "Other" = "#999999")
    
    ggplot(type_counts, aes(x = reorder(profile_type, count), y = count, fill = profile_type)) +
      geom_bar(stat = "identity", width = 0.6) +
      geom_text(aes(label = count), hjust = -0.3, size = 6) +
      scale_fill_manual(values = colors) +
      coord_flip() +
      labs(x = "", y = "Number of Unique Profiles", title = "US Core vs Other Profiles") +
      theme_minimal(base_size = 14) +
      theme(legend.position = "none",
            plot.title = element_text(face = "bold", size = 12)) +
      ylim(0, max(type_counts$count) * 1.2)
  })
  
  # Top Resources Chart
  output$resource_chart <- renderPlot({
    data <- all_profiles_data()
    
    if (nrow(data) == 0) {
      plot.new()
      text(0.5, 0.5, "No data available", cex = 1.5, col = "gray")
      return()
    }
    
    total_endpoints <- length(unique(data$url))
    
    resource_support <- data %>%
      group_by(resource) %>%
      summarise(
        endpoints = n_distinct(url),
        percentage = (n_distinct(url) / total_endpoints) * 100
      ) %>%
      arrange(desc(percentage)) %>%
      head(10)
    
    ggplot(resource_support, aes(x = reorder(resource, percentage), y = percentage)) +
      geom_bar(stat = "identity", fill = "#3498db", width = 0.6) +
      geom_text(aes(label = paste0(round(percentage, 0), "%")), hjust = -0.2, size = 3.5) +
      coord_flip() +
      labs(x = "", y = "% of Endpoints", title = "Top 10 Resources by Support") +
      theme_minimal(base_size = 14) +
      theme(plot.title = element_text(face = "bold", size = 12)) +
      ylim(0, max(resource_support$percentage) * 1.15)
  })
  
  # US Core Profile Support Table - DATA-DRIVEN
  output$us_core_table <- renderReactable({
    data <- all_profiles_data()
    
    if (nrow(data) == 0) {
      return(reactable(
        data.frame(Message = "No data available"),
        columns = list(Message = colDef(name = "")),
        pagination = FALSE
      ))
    }
    
    # Get actual US Core profiles from the data
    us_core_profiles <- data %>%
      filter(profile_type == "US Core") %>%
      mutate(profile_short = gsub(".*/", "", profileurl)) %>%
      group_by(profile_short, profileurl) %>%
      summarise(
        endpoints = n_distinct(url),
        .groups = "drop"
      ) %>%
      arrange(desc(endpoints), profile_short) %>%
      mutate(status = "✓ Supported")
    
    if (nrow(us_core_profiles) == 0) {
      return(reactable(
        data.frame(Message = "No US Core profiles found in current selection"),
        columns = list(Message = colDef(name = "")),
        pagination = FALSE
      ))
    }
    
    reactable(
      us_core_profiles,
      columns = list(
        profile_short = colDef(
          name = "US Core Profile", 
          minWidth = 250
        ),
        profileurl = colDef(show = FALSE),
        endpoints = colDef(
          name = "Supporting Endpoints",
          width = 150,
          align = "center",
          style = list(color = "#27ae60", fontWeight = "bold")
        ),
        status = colDef(
          name = "Status", 
          width = 120, 
          align = "center"
        )
      ),
      defaultPageSize = 20,
      striped = TRUE,
      searchable = TRUE,
      compact = TRUE
    )
  })
  
  # Paginated count for detailed table
  profile_total_count <- reactive({
    req(sel_fhir_version(), sel_vendor())
    
    count_query <- "SELECT COUNT(*) as total FROM mv_profiles_paginated WHERE 1=1"
    params <- list()
    
    count_query <- paste0(count_query, " AND fhir_version IN ({fhir_versions*})")
    params$fhir_versions <- sel_fhir_version()
    
    if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
      count_query <- paste0(count_query, " AND vendor_name = {vendor}")
      params$vendor <- sel_vendor()
    }
    
    if (length(sel_resource()) > 0 && sel_resource() != ui_special_values$ALL_RESOURCES) {
      count_query <- paste0(count_query, " AND resource = {resource}")
      params$resource <- sel_resource()
    }
    
    if (length(sel_profile()) > 0 && sel_profile() != ui_special_values$ALL_PROFILES) {
      count_query <- paste0(count_query, " AND profileurl = {profile}")
      params$profile <- sel_profile()
    }
    
    # Additional filters from new UI
    if (!is.null(input$resource_dropdown) && input$resource_dropdown != "All Resources") {
      count_query <- paste0(count_query, " AND resource = {resource_filter}")
      params$resource_filter <- input$resource_dropdown
    }
    
    if (!is.null(input$profile_type_filter) && input$profile_type_filter != "All Profiles") {
      if (input$profile_type_filter == "US Core Only") {
        count_query <- paste0(count_query, " AND profileurl LIKE '%/us/core/StructureDefinition/%'")
      } else if (input$profile_type_filter == "Other Profiles") {
        count_query <- paste0(count_query, " AND profileurl NOT LIKE '%/us/core/StructureDefinition/%'")
      }
    }
    
    if (trimws(input$profile_search_query) != "") {
      keyword <- tolower(trimws(input$profile_search_query))
      count_query <- paste0(count_query, 
        " AND (LOWER(url) LIKE {search} OR LOWER(profileurl) LIKE {search} OR LOWER(profilename) LIKE {search}",
        " OR LOWER(resource) LIKE {search} OR LOWER(vendor_name) LIKE {search})")
      params$search <- paste0("%", keyword, "%")
    }
    
    query <- do.call(glue_sql, c(list(count_query, .con = db_connection), params))
    result <- tbl(db_connection, sql(query)) %>% collect()
    as.numeric(result$total[1])
  })
  
  profile_total_pages <- reactive({
    total_count <- profile_total_count()
    if (total_count == 0) return(1)
    max(1, ceiling(total_count / profile_page_size))
  })
  
  observe({
    new_page <- profile_page_state()
    current_selector <- input$profile_page_selector
    
    if (is.null(current_selector) || 
        is.na(current_selector) || 
        !is.numeric(current_selector) ||
        current_selector != new_page) {
      
      isolate({
        updateNumericInput(session, "profile_page_selector", 
                          max = profile_total_pages(),
                          value = new_page)
      })
    }
  })
  
  observeEvent(input$profile_page_selector, {
    current_input <- input$profile_page_selector
    
    if (!is.null(current_input) && 
        !is.na(current_input) && 
        is.numeric(current_input) &&
        current_input > 0) {
      
      new_page <- max(1, min(current_input, profile_total_pages()))
      
      if (new_page != profile_page_state()) {
        profile_page_state(new_page)
      }
      
      if (new_page != current_input) {
        updateNumericInput(session, "profile_page_selector", value = new_page)
      }
    } else {
      invalidateLater(100)
      updateNumericInput(session, "profile_page_selector", value = profile_page_state())
    }
  }, ignoreInit = TRUE)
  
  observeEvent(input$profile_next_page, {
    if (profile_page_state() < profile_total_pages()) {
      profile_page_state(profile_page_state() + 1)
    }
  })
  
  observeEvent(input$profile_prev_page, {
    if (profile_page_state() > 1) {
      profile_page_state(profile_page_state() - 1)
    }
  })
  
  observeEvent(list(sel_fhir_version(), sel_vendor(), sel_resource(), sel_profile(), 
                    input$profile_search_query, input$profile_type_filter, input$resource_dropdown), {
    profile_page_state(1)
  })
  
  output$profile_prev_button_ui <- renderUI({
    if (profile_page_state() > 1) {
      actionButton(ns("profile_prev_page"), "Previous", icon = icon("arrow-left"))
    }
  })
  
  output$profile_next_button_ui <- renderUI({
    if (profile_page_state() < profile_total_pages()) {
      actionButton(ns("profile_next_page"), "Next", icon = icon("arrow-right"))
    }
  })
  
  output$profile_page_info <- renderText({
    paste("of", profile_total_pages())
  })
  
  # Paginated data for detailed table
  selected_fhir_endpoint_profiles <- reactive({
    req(sel_fhir_version(), sel_vendor())
    
    request_id <- isolate(current_request_id()) + 1
    current_request_id(request_id)
    
    profile_offset <- (profile_page_state() - 1) * profile_page_size
    
    base_query <- "SELECT url, profileurl, profilename, resource, fhir_version, vendor_name 
                   FROM mv_profiles_paginated 
                   WHERE 1=1"
    params <- list()
    
    base_query <- paste0(base_query, " AND fhir_version IN ({fhir_versions*})")
    params$fhir_versions <- sel_fhir_version()
    
    if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
      base_query <- paste0(base_query, " AND vendor_name = {vendor}")
      params$vendor <- sel_vendor()
    }
    
    if (length(sel_resource()) > 0 && sel_resource() != ui_special_values$ALL_RESOURCES) {
      base_query <- paste0(base_query, " AND resource = {resource}")
      params$resource <- sel_resource()
    }
    
    if (length(sel_profile()) > 0 && sel_profile() != ui_special_values$ALL_PROFILES) {
      base_query <- paste0(base_query, " AND profileurl = {profile}")
      params$profile <- sel_profile()
    }
    
    # Additional filters
    if (!is.null(input$resource_dropdown) && input$resource_dropdown != "All Resources") {
      base_query <- paste0(base_query, " AND resource = {resource_filter}")
      params$resource_filter <- input$resource_dropdown
    }
    
    if (!is.null(input$profile_type_filter) && input$profile_type_filter != "All Profiles") {
      if (input$profile_type_filter == "US Core Only") {
        base_query <- paste0(base_query, " AND profileurl LIKE '%/us/core/StructureDefinition/%'")
      } else if (input$profile_type_filter == "Other Profiles") {
        base_query <- paste0(base_query, " AND profileurl NOT LIKE '%/us/core/StructureDefinition/%'")
      }
    }
    
    if (trimws(input$profile_search_query) != "") {
      keyword <- tolower(trimws(input$profile_search_query))
      base_query <- paste0(base_query, 
        " AND (LOWER(url) LIKE {search} OR LOWER(profileurl) LIKE {search} OR LOWER(profilename) LIKE {search}",
        " OR LOWER(resource) LIKE {search} OR LOWER(vendor_name) LIKE {search})")
      params$search <- paste0("%", keyword, "%")
    }
    
    base_query <- paste0(base_query, " ORDER BY page_id LIMIT {limit} OFFSET {offset}")
    params$limit <- profile_page_size
    params$offset <- profile_offset
    
    query <- do.call(glue_sql, c(list(base_query, .con = db_connection), params))
    result <- tbl(db_connection, sql(query)) %>% collect()
    
    if (request_id == isolate(current_request_id())) {
      if (nrow(result) > 0) {
        result <- result %>%
          mutate(
            profile_type = categorize_profile(profileurl),
            url = paste0("<a class=\"lantern-url\" tabindex=\"0\" aria-label=\"Press enter to open pop up modal containing additional information for this endpoint.\" onkeydown = \"javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)\" onclick=\"Shiny.setInputValue(\'endpoint_popup\',&quot;", url, "&&", "None", "&quot,{priority: \'event\'});\">", url, "</a>")
          )
      }
      return(result)
    } else {
      return(data.frame())
    }
  })
  
  output$profiles_table <- renderReactable({
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
      defaultColDef = colDef(align = "center"),
      columns = list(
        url = colDef(
          name = "Endpoint", 
          minWidth = 280, 
          align = "left", 
          html = TRUE, 
          sortable = TRUE
        ),
        profileurl = colDef(
          name = "Profile URL", 
          minWidth = 280, 
          sortable = TRUE
        ),
        profilename = colDef(
          name = "Profile Name", 
          minWidth = 180, 
          sortable = TRUE
        ),
        profile_type = colDef(
          name = "Type",
          width = 110,
          align = "center",
          cell = function(value) {
            badge_class <- ifelse(value == "US Core", "badge-us-core", "badge-other")
            tags$span(class = paste("profile-badge", badge_class), value)
          },
          html = TRUE
        ),
        resource = colDef(name = "Resource", width = 120, sortable = TRUE),
        fhir_version = colDef(name = "FHIR Version", width = 110, sortable = TRUE),
        vendor_name = colDef(name = "Developer", minWidth = 150, sortable = TRUE)
      ),
      searchable = FALSE,
      showSortIcon = TRUE,
      highlight = TRUE,
      striped = TRUE,
      defaultPageSize = profile_page_size
    )
  })
}