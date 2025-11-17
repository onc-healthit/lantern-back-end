library(DT)
library(purrr)
library(reactable)
library(htmltools)

validationsmodule_UI <- function(id) {
  ns <- NS(id)

  tagList(
    # Custom CSS for modern validation dashboard
    tags$head(
      tags$style(HTML(paste0("
        /* Modern validation dashboard styling */
        .validation-dashboard {
          font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
        }
        
        /* KPI Cards */
        .kpi-card {
          background: white;
          border-radius: 12px;
          padding: 20px;
          box-shadow: 0 2px 8px rgba(0,0,0,0.1);
          transition: transform 0.2s ease, box-shadow 0.2s ease;
          margin-bottom: 20px;
        }
        
        .kpi-card:hover {
          transform: translateY(-4px);
          box-shadow: 0 4px 16px rgba(0,0,0,0.15);
        }
        
        .kpi-value {
          font-size: 36px;
          font-weight: 700;
          margin: 8px 0;
        }
        
        .kpi-label {
          font-size: 14px;
          color: #6c757d;
          text-transform: uppercase;
          letter-spacing: 0.5px;
          font-weight: 600;
        }
        
        .kpi-trend {
          font-size: 13px;
          margin-top: 8px;
        }
        
        .kpi-success {
          color: #28a745;
        }
        
        .kpi-warning {
          color: #ffc107;
        }
        
        .kpi-danger {
          color: #dc3545;
        }
        
        /* Section headers */
        .section-header {
          display: flex;
          align-items: center;
          margin: 30px 0 20px 0;
          padding-bottom: 12px;
          border-bottom: 2px solid #e9ecef;
        }
        
        .section-header h3 {
          margin: 0;
          font-size: 20px;
          font-weight: 600;
          color: #212529;
        }
        
        .section-icon {
          width: 32px;
          height: 32px;
          background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
          border-radius: 8px;
          display: flex;
          align-items: center;
          justify-content: center;
          color: white;
          margin-right: 12px;
        }
        
        /* Chart container */
        .chart-container {
          background: white;
          border-radius: 12px;
          padding: 24px;
          box-shadow: 0 2px 8px rgba(0,0,0,0.08);
          margin-bottom: 24px;
        }
        
        /* Validation rules list */
        .validation-rules-container {
          background: white;
          border-radius: 12px;
          padding: 20px;
          box-shadow: 0 2px 8px rgba(0,0,0,0.08);
          max-height: 600px;
          overflow-y: auto;
        }
        
        /* Modern table styling */
        .modern-validation-table {
          font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
        }
        
        .modern-validation-table .rt-table {
          border: 1px solid #e1e4e8;
          border-radius: 8px;
          overflow: hidden;
          box-shadow: 0 1px 3px rgba(0,0,0,0.08);
        }
        
        .modern-validation-table .rt-thead {
          background: linear-gradient(to bottom, #f8f9fa 0%, #f1f3f5 100%);
          border-bottom: 2px solid #dee2e6;
        }
        
        .modern-validation-table .rt-th {
          color: #495057;
          font-weight: 600;
          font-size: 13px;
          text-transform: uppercase;
          letter-spacing: 0.5px;
          padding: 12px 8px;
        }
        
        .modern-validation-table .rt-td {
          padding: 12px 8px;
          font-size: 14px;
          color: #212529;
        }
        
        .modern-validation-table .rt-tr:hover {
          background-color: #f8f9fa;
          transition: background-color 0.2s ease;
        }
        
        /* Info boxes */
        .info-box {
          background: #e7f3ff;
          border-left: 4px solid #2196F3;
          padding: 16px;
          border-radius: 6px;
          margin: 16px 0;
        }
        
        .info-box.warning {
          background: #fff3cd;
          border-left-color: #ffc107;
        }
        
        .info-box-title {
          font-weight: 600;
          color: #212529;
          margin-bottom: 8px;
        }
        
        .info-box-content {
          color: #495057;
          font-size: 14px;
          line-height: 1.6;
        }
        
        /* Modern buttons */
        .modern-nav-button {
          background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
          border: none;
          color: white;
          padding: 10px 20px;
          border-radius: 8px;
          font-weight: 500;
          cursor: pointer;
          transition: transform 0.2s ease, box-shadow 0.2s ease;
        }
        
        .modern-nav-button:hover {
          transform: translateY(-2px);
          box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
        }
        
        /* Page selector */
        #", ns("validation_page_selector"), " {
          border: 2px solid #e1e4e8;
          border-radius: 6px;
          text-align: center;
          font-weight: 600;
        }
        
        /* Tab-like sections */
        .insight-card {
          background: white;
          border-radius: 12px;
          padding: 20px;
          box-shadow: 0 2px 8px rgba(0,0,0,0.08);
          margin-bottom: 20px;
        }
      ")))
    ),
    
    div(class = "validation-dashboard",
      # Header section
      fluidRow(
        column(12,
          h2(style = "margin-top:0; color: #212529; font-weight: 700; font-size: 28px;", 
             "FHIR Compliance & Validation Dashboard"),
          div(class = "info-box",
            div(class = "info-box-title", 
                tags$i(class = "fa fa-info-circle", style = "margin-right: 8px;"),
                "About Validation Rules"),
            div(class = "info-box-content",
              "Lantern evaluates FHIR endpoints against validation rules to ensure compliance and quality. ",
              a("View complete validation rule documentation", 
                href = "Lantern Validation Rules and Descriptions.pdf", 
                target = "_blank", 
                class = "lantern-url",
                style = "color: #667eea; font-weight: 500;")
            )
          )
        )
      ),
      
      # KPI Cards Row
      fluidRow(
        column(3,
          div(class = "kpi-card",
            div(class = "kpi-label", "Total Endpoints Tested"),
            div(class = "kpi-value kpi-success", textOutput(ns("kpi_total_endpoints"))),
            div(class = "kpi-trend", "Across all selected filters")
          )
        ),
        column(3,
          div(class = "kpi-card",
            div(class = "kpi-label", "Compliance Rate"),
            div(class = "kpi-value", textOutput(ns("kpi_compliance_rate"))),
            div(class = "kpi-trend", textOutput(ns("kpi_compliance_text")))
          )
        ),
        column(3,
          div(class = "kpi-card",
            div(class = "kpi-label", "Passing Validations"),
            div(class = "kpi-value kpi-success", textOutput(ns("kpi_passing"))),
            div(class = "kpi-trend", "Tests passed")
          )
        ),
        column(3,
          div(class = "kpi-card",
            div(class = "kpi-label", "Failing Validations"),
            div(class = "kpi-value kpi-danger", textOutput(ns("kpi_failing"))),
            div(class = "kpi-trend", "Require attention")
          )
        )
      ),
      
      # Overview Chart Section
      fluidRow(
        column(12,
          div(class = "section-header",
            div(class = "section-icon", 
                tags$i(class = "fa fa-bar-chart")),
            h3("Validation Results Overview")
          ),
          div(class = "chart-container",
            htmlOutput(ns("anchorlink")),
            uiOutput(ns("validation_results_plot"))
          )
        )
      ),
      
      # Important Notes Section
      fluidRow(
        column(12,
          div(class = "info-box warning",
            div(class = "info-box-title",
              tags$i(class = "fa fa-exclamation-triangle", style = "margin-right: 8px;"),
              "Important Notes"
            ),
            div(class = "info-box-content",
              tags$p(style = "margin-bottom: 8px;", 
                "• The ONC Final Rule requires endpoints to support FHIR version 4.0.1. Other versions are included for reference."),
              tags$p(style = "margin: 0;",
                "• Note regarding ", tags$strong("messagingEndptRule"), ": There is a known issue with the Capability Statement invariant ",
                a("(cpb-3)", href = "http://hl7.org/fhir/capabilitystatement.html#invs", 
                  target = "_blank", class = "lantern-url", style = "color: #667eea;"),
                ". The FHIRPath expression is inconsistent with the stated expectation.")
            )
          )
        )
      ),
      
      # Detailed Analysis Section
      fluidRow(
        # Left column - Validation Rules List
        column(4,
          div(class = "section-header",
            div(class = "section-icon", 
                tags$i(class = "fa fa-list-ul")),
            h3("Validation Rules")
          ),
          div(class = "insight-card",
            p(style = "color: #6c757d; font-size: 14px; margin-bottom: 16px;",
              tags$i(class = "fa fa-hand-pointer-o", style = "margin-right: 6px;"),
              "Select a rule below to view detailed failure information →"
            ),
            div(class = "validation-rules-container",
              reactable::reactableOutput(ns("validation_details_table"))
            )
          )
        ),
        
        # Right column - Failure Details
        column(8,
          div(class = "section-header",
            div(class = "section-icon", 
                tags$i(class = "fa fa-exclamation-circle")),
            h3("Validation Failure Details")
          ),
          htmlOutput(ns("anchorpoint")),
          div(class = "insight-card",
            htmlOutput(ns("failure_table_subtitle")),
            tags$p(style = "color: #6c757d; font-size: 13px; margin-bottom: 16px;",
              tags$i(class = "fa fa-link", style = "margin-right: 6px;"),
              "Click any endpoint URL to view detailed information in a modal"),
            div(class = "modern-validation-table",
              reactable::reactableOutput(ns("validation_failure_table"))
            ),
            # Pagination controls
            fluidRow(
              column(3, 
                div(style = "display: flex; justify-content: flex-start; margin-top: 16px;", 
                    uiOutput(ns("validation_prev_page_ui"))
                )
              ),
              column(6,
                div(style = "display: flex; justify-content: center; align-items: center; gap: 10px; margin-top: 16px;",
                    numericInput(ns("validation_page_selector"), label = NULL, value = 1, min = 1, step = 1, width = "80px"),
                    textOutput(ns("validations_current_page_info"), inline = TRUE)
                )
              ),
              column(3, 
                div(style = "display: flex; justify-content: flex-end; margin-top: 16px;",
                    uiOutput(ns("validation_next_page_ui"))
                )
              )
            ),
            tags$p(style = "color: #6c757d; font-size: 13px; margin-top: 16px; font-style: italic;",
              tags$i(class = "fa fa-check-circle", style = "color: green; margin-right: 4px;"),
              "Green check = Capability Statement returned successfully | ",
              tags$i(class = "fa fa-times-circle", style = "color: red; margin-right: 4px;"),
              "Red X = No Capability Statement returned"
            )
          )
        )
      )
    )
  )
}

validationsmodule <- function(
  input,
  output,
  session,
  sel_fhir_version,
  sel_vendor,
  sel_validation_group
) {
  ns <- session$ns
  validations_page_size <- 10
  validation_page_state <- reactiveVal(1)
  current_request_id <- reactiveVal(0)
  
  # CRITICAL FIX: Add a reactive value to store the selected validation rule index
  selected_validation_index <- reactiveVal(1)

  # KPI Calculations
  output$kpi_total_endpoints <- renderText({
    req(sel_fhir_version(), sel_vendor(), sel_validation_group())
    
    fhir_versions <- paste0("'", paste(sel_fhir_version(), collapse = "','"), "'")
    vendor_filter <- if(sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
      paste0("AND vendor_name = '", sel_vendor(), "'")
    } else ""
    
    validation_group_filter <- if(sel_validation_group() != "All Groups") {
      references <- paste0("'", paste(validation_group_list[[sel_validation_group()]], collapse = "','"), "'")
      paste0("AND reference IN (", references, ")")
    } else ""
    
    query <- paste0(
      "SELECT COUNT(DISTINCT url) as count FROM mv_validation_results_plot ",
      "WHERE fhir_version IN (", fhir_versions, ") ",
      vendor_filter, " ", validation_group_filter
    )
    
    count <- dbGetQuery(db_connection, query)$count
    format(count, big.mark = ",")
  })
  
  output$kpi_compliance_rate <- renderText({
    results <- select_validation_results()
    if (nrow(results) == 0) return("N/A")
    
    total_tests <- sum(results$count)
    passing_tests <- sum(results$count[results$valid == "Success"])
    rate <- round((passing_tests / total_tests) * 100, 1)
    
    paste0(rate, "%")
  })
  
  output$kpi_compliance_text <- renderText({
    results <- select_validation_results()
    if (nrow(results) == 0) return("")
    
    total_tests <- sum(results$count)
    passing_tests <- sum(results$count[results$valid == "Success"])
    rate <- round((passing_tests / total_tests) * 100, 1)
    
    if (rate >= 90) {
      "Excellent compliance"
    } else if (rate >= 75) {
      "Good compliance"
    } else if (rate >= 50) {
      "Needs improvement"
    } else {
      "Critical attention needed"
    }
  })
  
  output$kpi_passing <- renderText({
    results <- select_validation_results()
    if (nrow(results) == 0) return("0")
    
    passing <- sum(results$count[results$valid == "Success"])
    format(passing, big.mark = ",")
  })
  
  output$kpi_failing <- renderText({
    results <- select_validation_results()
    if (nrow(results) == 0) return("0")
    
    failing <- sum(results$count[results$valid == "Failure"])
    format(failing, big.mark = ",")
  })

  # FIXED: Get total pages calculation using stored index
  validation_total_pages <- reactive({
    req(sel_fhir_version(), sel_vendor(), sel_validation_group())

    # Use the stored selected index instead of getReactableState
    selected_index <- selected_validation_index()
    selected_rule <- if (!is.null(selected_index) && selected_index > 0) {
      rules <- validation_rules()
      if (nrow(rules) >= selected_index) {
        rules$rule_name[selected_index]
      } else {
        "NO_RULES"
      }
    } else {
      "NO_RULES"
    }

    fhir_versions <- paste0("'", paste(sel_fhir_version(), collapse = "','"), "'")
    vendor_filter <- if(sel_vendor() != ui_special_values$ALL_DEVELOPERS) paste0("AND vendor_name = '", sel_vendor(), "'") else ""
    validation_group_filter <- if(sel_validation_group() != "All Groups") {
      references <- paste0("'", paste(validation_group_list[[sel_validation_group()]], collapse = "','"), "'")
      paste0("AND reference IN (", references, ")")
    } else ""

    query <- paste0(
      "SELECT COUNT(*) as count FROM mv_validation_failures ",
      "WHERE rule_name = '", selected_rule, "' ",
      "AND fhir_version IN (", fhir_versions, ") ",
      vendor_filter, " ",
      validation_group_filter
    )

    count <- dbGetQuery(db_connection, query)$count
    max(1, ceiling(count / validations_page_size))
  })

  # Break feedback loop
  observe({
    new_page <- validation_page_state()
    current_selector <- input$validation_page_selector
    
    if (is.null(current_selector) || is.na(current_selector) || 
        !is.numeric(current_selector) || current_selector != new_page) {
      isolate({
        updateNumericInput(session, "validation_page_selector", 
                          max = validation_total_pages(),
                          value = new_page)
      })
    }
  })

  # Handle page selector
  observeEvent(input$validation_page_selector, {
    current_input <- input$validation_page_selector
    
    if (!is.null(current_input) && !is.na(current_input) && 
        is.numeric(current_input) && current_input > 0) {
      new_page <- max(1, min(current_input, validation_total_pages()))
      
      if (new_page != validation_page_state()) {
        validation_page_state(new_page)
      }
      if (new_page != current_input) {
        updateNumericInput(session, "validation_page_selector", value = new_page)
      }
    } else {
      invalidateLater(100)
      updateNumericInput(session, "validation_page_selector", value = validation_page_state())
    }
  }, ignoreInit = TRUE)

  # Navigation buttons
  output$validation_prev_page_ui <- renderUI({
    if (validation_page_state() > 1) {
      actionButton(
        ns("validation_prev_page"),
        label = tagList(
          tags$i(class = "fa fa-arrow-left", style = "margin-right: 8px;"),
          "Previous"
        ),
        class = "modern-nav-button"
      )
    } else NULL
  })

  output$validation_next_page_ui <- renderUI({
    if (validation_page_state() < validation_total_pages()) {
      actionButton(
        ns("validation_next_page"),
        label = tagList(
          "Next",
          tags$i(class = "fa fa-arrow-right", style = "margin-left: 8px;")
        ),
        class = "modern-nav-button"
      )
    } else NULL
  })

  observeEvent(input$validation_next_page, {
    if (validation_page_state() < validation_total_pages()) {
      validation_page_state(validation_page_state() + 1)
    }
  })

  observeEvent(input$validation_prev_page, {
    if (validation_page_state() > 1) {
      validation_page_state(validation_page_state() - 1)
    }
  })
  
  output$validations_current_page_info <- renderText({
    paste("of", validation_total_pages())
  })

  output$anchorpoint <- renderUI({
    HTML("<span id='anchorid'></span>")
  })

  output$anchorlink <- renderUI({
    HTML("<p style='color: #6c757d; font-size: 14px;'><i class='fa fa-arrow-down' style='margin-right: 6px;'></i>See detailed validation failures <a class='lantern-url' style='color: #667eea; font-weight: 500;' href='#anchorid'>below</a></p>")
  })

  # FIXED: Reset page on filter change, but reset selection index only on filter change (not on selection change)
  observeEvent(list(sel_fhir_version(), sel_vendor(), sel_validation_group()), {
    validation_page_state(1)
    selected_validation_index(1)  # Reset to first rule when filters change
  })
  
  # FIXED: Update selected index when user clicks on a validation rule
  observeEvent(getReactableState("validation_details_table")$selected, {
    new_selection <- getReactableState("validation_details_table")$selected
    if (!is.null(new_selection) && length(new_selection) > 0) {
      selected_validation_index(new_selection)
      validation_page_state(1)  # Reset to first page when changing rules
    }
  }, ignoreNULL = TRUE, ignoreInit = TRUE)

  # Validation rules
  validation_rules <- reactive({
    req(sel_fhir_version(), sel_vendor(), sel_validation_group())
    
    fhir_versions <- paste0("'", paste(sel_fhir_version(), collapse = "','"), "'")
    vendor_filter <- if(sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
      paste0("AND vendor_name = '", sel_vendor(), "'")
    } else ""
    
    validation_group_filter <- if(sel_validation_group() != "All Groups") {
      references <- paste0("'", paste(validation_group_list[[sel_validation_group()]], collapse = "','"), "'")
      paste0("AND reference IN (", references, ")")
    } else ""
    
    query <- paste0("
      SELECT DISTINCT rule_name
      FROM mv_validation_results_plot
      WHERE fhir_version IN (", fhir_versions, ")
      ", vendor_filter, "
      ", validation_group_filter, "
      ORDER BY rule_name
    ")
    
    dbGetQuery(db_connection, query)
  })

  # Validation details
  validation_details <- reactive({
    res <- validation_rules()
    
    fhir_version_filter <- FALSE
    req(sel_fhir_version())
    
    if (length(sel_fhir_version()) != 1 || sel_fhir_version() == "Unknown") {
      query <- paste0("
        SELECT rule_name, fhir_version_names
        FROM mv_validation_details
        WHERE rule_name IN ('", paste(res$rule_name, collapse = "','"), "')
      ")
      
      versions <- dbGetQuery(db_connection, query)
      res <- res %>%
        left_join(versions, by = "rule_name") %>%
        mutate(versions_line = paste("Versions:", fhir_version_names))
      
      fhir_version_filter <- TRUE
    }
    
    res <- res %>%
      mutate(comment_line = paste("Comment:", validation_rules_descriptions[rule_name])) %>%
      mutate(rule_name_line = paste("Name:", rule_name)) %>%
      mutate(num = paste(row_number(), "."))
    
    if (fhir_version_filter) {
      res <- res %>%
        distinct(num, rule_name_line, comment_line, versions_line) %>%
        mutate(entry = paste(num, rule_name_line, versions_line, comment_line, sep = "<br>")) %>%
        select(entry)
    } else {
      res <- res %>%
        distinct(num, rule_name_line, comment_line) %>%
        mutate(entry = paste(num, rule_name_line, comment_line, sep = "<br>")) %>%
        select(entry)
    }
    
    res
  })

  # Selected validations
  selected_validations <- reactive({
    query <- paste0("SELECT * FROM mv_validation_results_plot")
    res <- dbGetQuery(db_connection, query)
    
    req(sel_fhir_version(), sel_vendor(), sel_validation_group())
    res <- res %>% filter(fhir_version %in% sel_fhir_version())
    
    if (sel_validation_group() != "All Groups") {
      res <- res %>% filter(reference %in% validation_group_list[[sel_validation_group()]])
    }
    if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
      res <- res %>% filter(vendor_name == sel_vendor())
    }

    res <- res %>%
      mutate(linkURL = paste0("<a class=\"lantern-url\" tabindex=\"0\" aria-label=\"Press enter to open pop up modal containing additional information for this endpoint.\" onkeydown = \"javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)\" onclick=\"Shiny.setInputValue(\'endpoint_popup\',&quot;", url, "&&", "None", "&quot,{priority: \'event\'});\">", url, "</a>"))
  })

  # Select validation results
  select_validation_results <- reactive({
    req(sel_fhir_version(), sel_vendor(), sel_validation_group())
    
    fhir_versions <- paste0("'", paste(sel_fhir_version(), collapse = "','"), "'")
    vendor_filter <- if(sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
      paste0("AND vendor_name = '", sel_vendor(), "'")
    } else ""
    
    validation_group_filter <- if(sel_validation_group() != "All Groups") {
      references <- paste0("'", paste(validation_group_list[[sel_validation_group()]], collapse = "','"), "'")
      paste0("AND reference IN (", references, ")")
    } else ""
    
    query <- paste0("
      SELECT rule_name, valid, COUNT(*) as count
      FROM mv_validation_results_plot
      WHERE fhir_version IN (", fhir_versions, ")
      ", vendor_filter, "
      ", validation_group_filter, "
      GROUP BY rule_name, valid
      ORDER BY rule_name
    ")
    
    res <- dbGetQuery(db_connection, query) %>%
      mutate(valid = if_else(valid == TRUE, "Success", "Failure")) %>%
      mutate(count = as.double(count))
    
    return(res)
  })

  # FIXED: Paged failed validation results with race condition protection using stored index
  paged_failed_validation_results <- reactive({
    req(sel_fhir_version(), sel_vendor(), sel_validation_group())
    
    # Generate unique request ID
    request_id <- isolate(current_request_id()) + 1
    current_request_id(request_id)
    
    # Use the stored selected index
    selected_index <- selected_validation_index()
    selected_rule <- if (!is.null(selected_index) && selected_index > 0) {
      rules <- validation_rules()
      if (nrow(rules) >= selected_index) {
        rules$rule_name[selected_index]
      } else {
        "NO_RULES"
      }
    } else {
      "NO_RULES"
    }
    
    # Build filters
    fhir_versions <- paste0("'", paste(sel_fhir_version(), collapse = "','"), "'")
    vendor_filter <- if(sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
      paste0("AND vendor_name = '", sel_vendor(), "'")
    } else ""
    
    validation_group_filter <- if(sel_validation_group() != "All Groups") {
      references <- paste0("'", paste(validation_group_list[[sel_validation_group()]], collapse = "','"), "'")
      paste0("AND reference IN (", references, ")")
    } else ""

    limit <- validations_page_size
    offset <- (validation_page_state() - 1) * validations_page_size
    
    # Query failed validations
    query <- paste0("
      SELECT fhir_version, url, expected, actual, vendor_name
      FROM mv_validation_failures
      WHERE rule_name = '", selected_rule, "'
      AND fhir_version IN (", fhir_versions, ") ",
      vendor_filter, " ", 
      validation_group_filter, " ",
      "ORDER BY url LIMIT ", limit, " OFFSET ", offset
    )
    
    result <- dbGetQuery(db_connection, query)
    
    # Only return if latest request
    if (request_id == isolate(current_request_id())) {
      res <- result %>%
        mutate(url = paste0("<a class=\"lantern-url\" tabindex=\"0\" aria-label=\"Press enter to open pop up modal containing additional information for this endpoint.\" onkeydown = \"javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)\" onclick=\"Shiny.setInputValue(\'endpoint_popup\',&quot;", url, "&&", "None", "&quot,{priority: \'event\'});\">", url, "</a>"))
      return(res)
    } else {
      return(data.frame())
    }
  })

  # FIXED: Render validation details table with modern styling using stored index
  output$validation_details_table <- reactable::renderReactable({
    data <- validation_details() %>% select(entry)
    
    reactable(
      data,
      columns = list(
        entry = colDef(
          name = "Validation Rules", 
          html = TRUE,
          style = list(cursor = "pointer")
        )
      ),
      selection = "single",
      onClick = "select",
      defaultSelected = selected_validation_index(),  # Use the stored index
      pagination = FALSE,
      height = 500,
      striped = FALSE,
      borderless = TRUE,
      rowStyle = function(index) {
        # Use the stored selected index
        is_selected <- !is.null(selected_validation_index()) && 
                       length(selected_validation_index()) > 0 && 
                       index == selected_validation_index()
        
        list(
          padding = "12px",
          borderLeft = if (is_selected) "4px solid #667eea" else "4px solid transparent",
          background = if (is_selected) "#f0f4ff" else "#f8f9fa",
          borderRadius = "6px",
          marginBottom = "8px",
          transition = "all 0.2s ease"
        )
      }
    )
  })

  # Calculate plot height
  validation_plot_height <- reactive({
    max(nrow(select_validation_results()) * 25, 400)
  })

  # Render validation results plot
  output$validation_results_plot <- renderUI({
    if (nrow(select_validation_results()) != 0) {
      tagList(
        plotOutput(ns("validation_bar_plot"), height = validation_plot_height())
      )
    } else {
      tagList(
        plotOutput(ns("validation_bar_empty_plot"), height = validation_plot_height())
      )
    }
  })

  # Render validation bar plot
  output$validation_bar_plot <- renderCachedPlot({
    ggplot(select_validation_results(), aes(x = fct_rev(as.factor(rule_name)), y = count, fill = valid)) +
      geom_col(width = 0.7) +
      geom_text(aes(label = stat(y)), position = position_stack(vjust = 0.5), 
                color = "white", fontface = "bold", size = 4) +
      ggtitle("Validation Test Results by Rule") +
      theme_minimal() +
      theme(
        plot.title = element_text(hjust = 0.5, size = 18, face = "bold", color = "#212529"),
        legend.position = "bottom",
        legend.title = element_blank(),
        text = element_text(size = 13, color = "#495057"),
        axis.text = element_text(color = "#495057"),
        panel.grid.major.y = element_line(color = "#e9ecef"),
        panel.grid.minor = element_blank(),
        panel.grid.major.x = element_blank()
      ) +
      labs(x = "", y = "Number of Endpoints", fill = "Result") +
      scale_y_continuous(sec.axis = sec_axis(~.)) +
      scale_fill_manual(
        values = c("Failure" = "#dc3545", "Success" = "#28a745"), 
        limits = c("Failure", "Success")
      ) +
      guides(fill = guide_legend(reverse = TRUE)) +
      coord_flip()
  },
    sizePolicy = sizeGrowthRatio(width = 400, height = 400, growthRate = 1.2),
    res = 72,
    cache = "app",
    cacheKeyExpr = {
      list(sel_fhir_version(), sel_vendor(), sel_validation_group(), now("UTC"))
    })

  # Render empty plot
  output$validation_bar_empty_plot <- renderPlot({
    ggplot(select_validation_results()) +
      geom_col(width = 0.8) +
      labs(x = "", y = "") +
      theme_minimal() +
      theme(
        axis.text.x = element_blank(),
        axis.text.y = element_blank(), 
        axis.ticks = element_blank(),
        panel.grid = element_blank()
      ) +
      annotate("text", 
               label = "No validation results available\nfor the selected filters", 
               x = 1, y = 2, size = 5, colour = "#dc3545", hjust = 0.5, fontface = "bold")
  })

  # Helper function for capability statement icon
  cap_stat_icon <- function(fhir_version) {
    if (fhir_version == "No Cap Stat") {
      tags$i(class = "fa fa-times-circle", style = "color: #dc3545; font-size: 16px;", 
             `aria-hidden` = "true")
    } else {
      tags$i(class = "fa fa-check-circle", style = "color: #28a745; font-size: 16px;", 
             `aria-hidden` = "true")
    }
  }

  # FIXED: Render failure table subtitle using stored index
  output$failure_table_subtitle <- renderUI({
    selected_index <- selected_validation_index()  # Use stored index
    if (!is.null(selected_index) && selected_index > 0) {
      rules <- validation_rules()
      if (nrow(rules) >= selected_index) {
        rule_name <- rules$rule_name[selected_index]
        div(
          style = "margin-bottom: 16px; padding: 12px; background: #f8f9fa; border-radius: 6px; border-left: 4px solid #667eea;",
          tags$strong(style = "color: #212529; font-size: 15px;", 
                     tags$i(class = "fa fa-filter", style = "margin-right: 8px;"),
                     "Showing failures for rule: "),
          tags$span(style = "color: #667eea; font-weight: 600;", rule_name)
        )
      } else {
        div(
          style = "margin-bottom: 16px; padding: 12px; background: #fff3cd; border-radius: 6px; border-left: 4px solid #ffc107;",
          tags$i(class = "fa fa-info-circle", style = "margin-right: 8px; color: #856404;"),
          tags$span(style = "color: #856404;", "Please select a validation rule from the left to view failure details")
        )
      }
    } else {
      div(
        style = "margin-bottom: 16px; padding: 12px; background: #fff3cd; border-radius: 6px; border-left: 4px solid #ffc107;",
        tags$i(class = "fa fa-info-circle", style = "margin-right: 8px; color: #856404;"),
        tags$span(style = "color: #856404;", "Please select a validation rule from the left to view failure details")
      )
    }
  })

  # Render validation failure table with modern styling
  output$validation_failure_table <- reactable::renderReactable({
    paged_data <- paged_failed_validation_results()
    
    reactable(
      paged_data,
      defaultColDef = colDef(
        headerStyle = list(
          background = "#f8f9fa",
          color = "#495057",
          fontWeight = "600",
          fontSize = "13px",
          textTransform = "uppercase",
          letterSpacing = "0.5px"
        ),
        style = function(value, index) {
          if (nrow(paged_data) > 0 && paged_data$fhir_version[index] == "No Cap Stat") {
            list(background = "#fff3cd")
          } else {
            list()
          }
        }
      ),
      columns = list(
        fhir_version = colDef(
          name = "FHIR Version",
          minWidth = 150,
          cell = function(value, index) {
            icon <- cap_stat_icon(paged_data$fhir_version[index])
            tagList(
              div(style = list(display = "inline-block", width = "30px", marginRight = "8px"), icon),
              tags$span(style = "font-weight: 500;", value)
            )
          }
        ),
        url = colDef(
          name = "Endpoint URL", 
          html = TRUE, 
          minWidth = 300,
          style = list(fontSize = "13px")
        ),
        expected = colDef(
          name = "Expected Value",
          minWidth = 150,
          cell = function(value) {
            div(
              style = "background: #d4edda; color: #155724; padding: 4px 8px; border-radius: 4px; font-size: 12px; display: inline-block;",
              tags$i(class = "fa fa-check", style = "margin-right: 4px;"),
              value
            )
          }
        ),
        actual = colDef(
          name = "Actual Value",
          minWidth = 150,
          cell = function(value) {
            div(
              style = "background: #f8d7da; color: #721c24; padding: 4px 8px; border-radius: 4px; font-size: 12px; display: inline-block;",
              tags$i(class = "fa fa-times", style = "margin-right: 4px;"),
              value
            )
          }
        ),
        vendor_name = colDef(
          name = "Developer",
          minWidth = 150,
          style = list(fontWeight = "500", fontSize = "13px")
        )
      ),
      striped = TRUE,
      highlight = TRUE,
      bordered = FALSE
    )
  })
}