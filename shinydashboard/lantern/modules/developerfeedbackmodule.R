library(DT)
library(purrr)
library(reactable)
library(ggplot2)
library(dplyr)
library(stringr)
library(shiny)
library(htmltools)
library(tidyr)

developerfeedbackmodule_UI <- function(id) {
  ns <- NS(id)
  
  tagList(
    # Custom CSS for modern styling
    tags$head(
      tags$style(HTML("
        /* Modern card styling */
        .modern-card {
          background: white;
          border-radius: 8px;
          box-shadow: 0 2px 8px rgba(0,0,0,0.1);
          padding: 20px;
          margin-bottom: 20px;
          transition: box-shadow 0.3s ease;
        }
        
        .modern-card:hover {
          box-shadow: 0 4px 12px rgba(0,0,0,0.15);
        }
        
        /* Enhanced info boxes */
        .info-box {
          border-radius: 8px;
          box-shadow: 0 2px 8px rgba(0,0,0,0.08);
          transition: all 0.3s ease;
          border: none;
        }
        
        .info-box:hover {
          transform: translateY(-2px);
          box-shadow: 0 4px 12px rgba(0,0,0,0.15);
        }
        
        .info-box-icon {
          border-radius: 8px 0 0 8px;
        }
        
        /* Modern headers */
        .page-header {
          color: #1B5A7F;
          font-weight: 600;
          margin-bottom: 15px;
          padding-bottom: 10px;
          border-bottom: 3px solid #1B5A7F;
        }
        
        .section-header {
          color: #2c3e50;
          font-weight: 600;
          margin-top: 25px;
          margin-bottom: 15px;
          font-size: 1.3em;
        }
        
        .subsection-header {
          color: #34495e;
          font-weight: 500;
          margin-top: 15px;
          margin-bottom: 10px;
          font-size: 1.1em;
        }
        
        /* Modern wellPanel styling */
        .well {
          background: white;
          border: 1px solid #e0e0e0;
          border-radius: 8px;
          box-shadow: 0 2px 6px rgba(0,0,0,0.06);
          padding: 20px;
        }
        
        /* Modern progress bars */
        .progress {
          height: 8px;
          border-radius: 4px;
          background-color: #ecf0f1;
          box-shadow: inset 0 1px 2px rgba(0,0,0,0.1);
        }
        
        .progress-bar {
          border-radius: 4px;
          transition: width 0.6s ease;
        }
        
        .progress-group {
          margin-bottom: 20px;
        }
        
        .progress-text {
          font-weight: 500;
          color: #2c3e50;
        }
        
        /* Enhanced filter section */
        .filter-section {
          background: linear-gradient(135deg, #f8f9fa 0%, #ffffff 100%);
          border-radius: 8px;
          padding: 15px;
          margin-bottom: 15px;
        }
        
        /* Modern select inputs */
        .selectize-input {
          border-radius: 6px;
          border: 1.5px solid #d0d0d0;
          transition: all 0.3s ease;
        }
        
        .selectize-input:hover {
          border-color: #1B5A7F;
        }
        
        .selectize-input:focus {
          border-color: #1B5A7F;
          box-shadow: 0 0 0 3px rgba(27, 90, 127, 0.1);
        }
        
        /* Info line styling */
        .info-line {
          padding: 8px 0;
          border-bottom: 1px solid #f0f0f0;
          display: flex;
          justify-content: space-between;
          align-items: center;
        }
        
        .info-line:last-child {
          border-bottom: none;
        }
        
        .info-line span:first-child {
          color: #5a6c7d;
          font-weight: 500;
        }
        
        .info-line span:last-child {
          color: #2c3e50;
          font-weight: 600;
        }
        
        /* Alert styling */
        .alert {
          border-radius: 8px;
          border-left: 4px solid;
          box-shadow: 0 2px 6px rgba(0,0,0,0.08);
          padding: 12px 15px;
        }
        
        .alert-danger {
          background-color: #fff5f5;
          border-left-color: #dc3545;
          color: #721c24;
        }
        
        .alert-warning {
          background-color: #fffbf0;
          border-left-color: #ffc107;
          color: #856404;
        }
        
        .alert-info {
          background-color: #f0f8ff;
          border-left-color: #007bff;
          color: #004085;
        }
        
        .alert-success {
          background-color: #f0fff4;
          border-left-color: #28a745;
          color: #155724;
        }
        
        .alert-secondary {
          background-color: #f8f9fa;
          border-left-color: #6c757d;
          color: #383d41;
        }
        
        /* Download button styling */
        .btn-download {
          background: linear-gradient(135deg, #1B5A7F 0%, #2874a6 100%);
          color: white;
          border: none;
          border-radius: 8px;
          padding: 12px 24px;
          font-weight: 500;
          transition: all 0.3s ease;
          box-shadow: 0 2px 6px rgba(27, 90, 127, 0.3);
        }
        
        .btn-download:hover {
          background: linear-gradient(135deg, #2874a6 0%, #1B5A7F 100%);
          transform: translateY(-2px);
          box-shadow: 0 4px 10px rgba(27, 90, 127, 0.4);
        }
        
        /* Chart container styling */
        .chart-container {
          background: white;
          border-radius: 8px;
          padding: 15px;
          box-shadow: 0 2px 6px rgba(0,0,0,0.06);
          margin-bottom: 20px;
        }
        
        /* Reactable modern styling */
        .reactable {
          border-radius: 8px;
          overflow: hidden;
          box-shadow: 0 2px 8px rgba(0,0,0,0.08);
        }
        
        /* Metric cards styling */
        .metric-card {
          background: white;
          border-radius: 8px;
          padding: 15px;
          box-shadow: 0 2px 6px rgba(0,0,0,0.06);
          margin-bottom: 15px;
        }
        
        .metric-title {
          font-size: 0.9em;
          color: #7f8c8d;
          font-weight: 500;
          margin-bottom: 8px;
        }
        
        .metric-value {
          font-size: 1.5em;
          font-weight: 600;
          color: #2c3e50;
        }
        
        /* Maintain existing Lantern styles for accessibility */
        a:focus-visible, button:focus-visible, select:focus-visible, input:focus-visible {
          border: 4px solid #000 !important;
          background-color: yellow !important;
          color: black !important;
          outline: none;
        }
      "))
    ),
    
    fluidRow(
      column(width = 12,
        h2(class = "page-header", "Organization Data Quality Dashboard")
      )
    ),
    fluidRow(
      column(width = 12,
        div(style = "background: linear-gradient(135deg, #f8f9fa 0%, #ffffff 100%); 
                     padding: 15px; border-radius: 8px; margin-bottom: 20px; 
                     border-left: 4px solid #1B5A7F;",
          p(style = "margin: 0; color: #5a6c7d; line-height: 1.6;",
            tags$strong("About this dashboard:"), 
            " This dashboard provides comprehensive data quality metrics for organization data extracted from FHIR bundles. ",
            "Use this information to improve the quality of organization data in your endpoint implementations."
          )
        )
      )
    ),
    
    # Enhanced summary cards row
    fluidRow(
      column(width = 4,
        div(class = "info-box bg-blue",
          div(class = "info-box-icon",
            tags$i(class = "fa fa-building", style = "font-size: 40px;")
          ),
          div(class = "info-box-content",
            span(class = "info-box-text", style = "font-weight: 500;", "Total Organizations"),
            span(class = "info-box-number", style = "font-size: 32px; font-weight: 600;", 
                 textOutput(ns("total_orgs"), inline = TRUE))
          )
        )
      ),
      column(width = 4,
        div(class = "info-box bg-green",
          div(class = "info-box-icon",
            tags$i(class = "fa fa-check-circle", style = "font-size: 40px;")
          ),
          div(class = "info-box-content",
            span(class = "info-box-text", style = "font-weight: 500;", "Conforming Organizations"),
            span(class = "info-box-number", style = "font-size: 32px; font-weight: 600;", 
                 textOutput(ns("high_quality_count"), inline = TRUE))
          )
        )
      ),
      column(width = 4,
        div(class = "info-box bg-red",
          div(class = "info-box-icon",
            tags$i(class = "fa fa-exclamation-triangle", style = "font-size: 40px;")
          ),
          div(class = "info-box-content",
            span(class = "info-box-text", style = "font-weight: 500;", "Non-conforming Organizations"),
            span(class = "info-box-number", style = "font-size: 32px; font-weight: 600;", 
                 textOutput(ns("low_quality_count"), inline = TRUE))
          )
        )
      )
    ),
    
    # Main content row
    fluidRow(
      # Left column - Charts and Tables
      column(width = 8,
        # Data Quality Overview
        div(class = "modern-card",
          h3(class = "section-header", 
             tags$i(class = "fa fa-chart-bar", style = "margin-right: 8px;"), 
             "Data Quality Overview"),
          div(class = "chart-container",
            plotOutput(ns("quality_overview_chart"), height = "400px")
          )
        ),
        
        # Identifier Type Analysis
        div(class = "modern-card",
          h3(class = "section-header",
             tags$i(class = "fa fa-id-card", style = "margin-right: 8px;"),
             "Identifier Analysis"),
          fluidRow(
            column(width = 6,
              div(class = "chart-container",
                h4(class = "subsection-header", "Type Distribution"),
                plotOutput(ns("identifier_type_distribution_chart"), height = "350px")
              )
            ),
            column(width = 6,
              div(class = "chart-container",
                h4(class = "subsection-header", "Conformance by Type"),
                plotOutput(ns("conformance_by_type_chart"), height = "350px")
              )
            )
          ),
          div(class = "chart-container",
            h4(class = "subsection-header", "Organization Status Breakdown"),
            plotOutput(ns("organization_identifier_status_chart"), height = "300px")
          ),
          div(style = "margin-top: 20px;",
            h4(class = "subsection-header", "Detailed Identifier Metrics"),
            reactable::reactableOutput(ns("identifier_type_table"))
          )
        ),
        
        # Detailed Issues
        div(class = "modern-card", style = "margin-top: 20px;",
          h3(class = "section-header",
             tags$i(class = "fa fa-exclamation-circle", style = "margin-right: 8px;"),
             "Data Quality Issues by Category"),
          reactable::reactableOutput(ns("issues_detail_table"))
        ),

        # Data Issues in Lantern
        div(class = "modern-card", style = "margin-top: 20px;",
          h3(class = "section-header",
             tags$i(class = "fa fa-database", style = "margin-right: 8px;"),
             "Data Issues in Lantern"),
          div(style = "margin-bottom: 15px;",
            p(style = "color: #5a6c7d; line-height: 1.6;",
              "This section tracks developers with data collection issues. ",
              tags$strong("Note: "), "Counts show the current state of endpoint data (endpoint_names field). ",
              "Developers may still appear in Lantern filters if organization records exist in the database ",
              "from previous successful extractions or as 'Unknown' organization placeholders. ",
              "Check the 'Organizations' column to see if database records exist."
            )
          ),
          fluidRow(
            column(width = 3,
              div(class = "metric-card",
                div(class = "metric-title",
                  tags$i(class = "fa fa-exclamation-triangle", style = "margin-right: 5px;"),
                  "Developers w/ No Org Data"
                ),
                div(class = "metric-value", style = "color: #dc3545;",
                  textOutput(ns("developers_no_org_data_count"), inline = TRUE)
                )
              )
            ),
            column(width = 3,
              div(class = "metric-card",
                div(class = "metric-title",
                  tags$i(class = "fa fa-inbox", style = "margin-right: 5px;"),
                  "Endpoints w/ No Org Data"
                ),
                div(class = "metric-value", style = "color: #dc3545;",
                  textOutput(ns("endpoints_no_org_data_count"), inline = TRUE)
                ),
                div(style = "margin-top: 8px; font-size: 0.85em; color: #7f8c8d;",
                  "Endpoints with no organization data"
                )
              )
            ),
            column(width = 3,
              div(class = "metric-card",
                div(class = "metric-title",
                  tags$i(class = "fa fa-share-alt", style = "margin-right: 5px;"),
                  "Shared Service Base URLs"
                ),
                div(class = "metric-value", style = "color: #ffc107;",
                  textOutput(ns("developers_sharing_list_sources_count"), inline = TRUE)
                ),
                div(style = "margin-top: 8px; font-size: 0.85em; color: #7f8c8d;",
                  "Developers Sharing the Same Service Base URL"
                )
              )
            ),
            column(width = 3,
              div(class = "metric-card",
                div(class = "metric-title",
                  tags$i(class = "fa fa-unlink", style = "margin-right: 5px;"),
                  "Inaccessible Sources"
                ),
                div(class = "metric-value", style = "color: #dc3545;",
                  textOutput(ns("inaccessible_list_sources_count"), inline = TRUE)
                ),
                div(style = "margin-top: 8px; font-size: 0.85em; color: #7f8c8d;",
                  "Unreachable list sources"
                )
              )
            )
          ),
          fluidRow(style = "margin-top: 15px;",
            column(width = 3,
              div(class = "metric-card",
                div(class = "metric-title",
                  tags$i(class = "fa fa-folder-open", style = "margin-right: 5px;"),
                  "Empty FHIR Bundles"
                ),
                div(class = "metric-value", style = "color: #dc3545;",
                  textOutput(ns("developers_empty_bundles_count"), inline = TRUE)
                ),
                div(style = "margin-top: 8px; font-size: 0.85em; color: #7f8c8d;",
                  "CHPL developers with no endpoints discovered"
                )
              )
            )
          ),
          div(style = "margin-top: 20px;",
            h4(class = "subsection-header",
               tags$i(class = "fa fa-table", style = "margin-right: 5px;"),
               "All Developers with Data Issues"),
            p(style = "color: #5a6c7d; font-size: 0.9em; margin-bottom: 10px;",
              "Complete list of all developers showing endpoints, organizations extracted, and data completeness."
            ),
            reactable::reactableOutput(ns("developer_data_issues_table"))
          )
        )
      ),
      
      # Right column - Filters and Summary
      column(width = 4,
        # Filters
        div(class = "modern-card filter-section",
          h4(style = "color: #1B5A7F; margin-top: 0;",
             tags$i(class = "fa fa-filter", style = "margin-right: 8px;"),
             "Filters"),
          selectInput(
            inputId = ns("vendor_filter"),
            label = "Certified API Developer:",
            choices = NULL,
            selected = "All Developers"
          )
        ),
        
        # Quality Metrics
        div(class = "modern-card",
          h4(style = "color: #1B5A7F; margin-top: 0;",
             tags$i(class = "fa fa-tachometer-alt", style = "margin-right: 8px;"),
             "Quality Metrics"),
          div(style = "margin-top: 15px;",
            div(class = "metric-card",
              div(class = "metric-title", "Identifier Type Validation"),
              div(class = "progress-group",
                div(style = "display: flex; justify-content: space-between; margin-bottom: 8px;",
                  span(class = "progress-text", "Valid Identifiers"),
                  span(style = "font-weight: 600; color: #28a745;", 
                       textOutput(ns("identifier_percentage"), inline = TRUE))
                ),
                div(class = "progress",
                  div(class = "progress-bar bg-success", 
                      style = paste0("width: ", textOutput(ns("identifier_progress_width"), inline = TRUE)))
                )
              )
            ),
            div(class = "metric-card",
              div(class = "metric-title", "Organization Names"),
              div(class = "progress-group",
                div(style = "display: flex; justify-content: space-between; margin-bottom: 8px;",
                  span(class = "progress-text", "Quality Names"),
                  span(style = "font-weight: 600; color: #007bff;", 
                       textOutput(ns("name_percentage"), inline = TRUE))
                ),
                div(class = "progress",
                  div(class = "progress-bar bg-primary", 
                      style = paste0("width: ", textOutput(ns("name_progress_width"), inline = TRUE)))
                )
              )
            ),
            div(class = "metric-card",
              div(class = "metric-title", "Addresses"),
              div(class = "progress-group",
                div(style = "display: flex; justify-content: space-between; margin-bottom: 8px;",
                  span(class = "progress-text", "Complete Addresses"),
                  span(style = "font-weight: 600; color: #fd7e14;", 
                       textOutput(ns("address_percentage"), inline = TRUE))
                ),
                div(class = "progress",
                  div(class = "progress-bar bg-warning", 
                      style = paste0("width: ", textOutput(ns("address_progress_width"), inline = TRUE)))
                )
              )
            )
          )
        ),
        
        # Identifier Breakdown
        div(class = "modern-card",
          h4(style = "color: #1B5A7F; margin-top: 0;",
             tags$i(class = "fa fa-list-alt", style = "margin-right: 8px;"),
             "Identifier Breakdown"),
          div(id = "identifier-breakdown", style = "margin-top: 15px;",
            div(class = "info-line",
              span("Valid identifiers:"),
              span(textOutput(ns("valid_identifier_count_display"), inline = TRUE))
            ),
            div(class = "info-line",
              span("No identifier data:"),
              span(textOutput(ns("no_identifier_count_display"), inline = TRUE))
            ),
            div(class = "info-line",
              span("Only invalid identifiers:"),
              span(textOutput(ns("invalid_only_count_display"), inline = TRUE))
            )
          )
        ),
        
        # Recommendations
        div(class = "modern-card",
          h4(style = "color: #1B5A7F; margin-top: 0;",
             tags$i(class = "fa fa-lightbulb", style = "margin-right: 8px;"),
             "Recommendations"),
          uiOutput(ns("recommendations"))
        )
      )
    ),
    
    # Download section
    fluidRow(
      column(width = 12, style = "padding-top: 20px; text-align: center;",
        downloadButton(
          outputId = ns("download_feedback_report"),
          label = "Download Quality Report (CSV)",
          class = "btn-download",
          icon = icon("download")
        )
      )
    )
  )
}

developerfeedbackmodule <- function(
  input,
  output,
  session
) {
  ns <- session$ns
  
  # Initialize vendor choices
  observe({
    vendor_choices <- c("All Developers", app$vendor_list())
    updateSelectInput(session, "vendor_filter", choices = vendor_choices, selected = "All Developers")
  })
  
  # Get filtered organization data from materialized views
  filtered_quality_summary <- reactive({
    current_vendor <- input$vendor_filter
    if (is.null(current_vendor)) current_vendor <- "All Developers"
    
    # Query the summary materialized view
    query_str <- "SELECT * FROM mv_organization_quality_summary WHERE vendor_name = {vendor}"
    
    data_query <- glue::glue_sql(query_str, vendor = current_vendor, .con = db_connection)
    
    result <- tbl(db_connection, sql(data_query)) %>% collect()
    
    # Debug output
    if (nrow(result) == 0) {
      cat("No data found for vendor:", current_vendor, "\n")
      # Return default values
      return(data.frame(
        vendor_name = current_vendor,
        total_organizations = 0,
        organizations_with_valid_identifiers = 0,
        organizations_with_no_identifiers = 0,
        organizations_with_invalid_only = 0,
        organizations_all_valid = 0,
        organizations_mixed_valid = 0,
        organizations_with_valid_names = 0,
        organizations_with_valid_addresses = 0,
        high_quality_organizations = 0,
        low_quality_organizations = 0,
        fully_conformant = 0,
        partially_conformant = 0,
        minimally_conformant = 0,
        non_conformant = 0,
        avg_conformance_rate = 0,
        avg_quality_score = 0,
        identifier_percentage = 0,
        name_percentage = 0,
        address_percentage = 0,
        stringsAsFactors = FALSE
      ))
    }
    
    # Ensure numeric columns are properly typed
    result <- result %>%
      mutate(
        total_organizations = as.numeric(total_organizations),
        organizations_with_valid_identifiers = as.numeric(organizations_with_valid_identifiers),
        organizations_with_no_identifiers = as.numeric(organizations_with_no_identifiers),
        organizations_with_invalid_only = as.numeric(organizations_with_invalid_only),
        organizations_with_valid_names = as.numeric(organizations_with_valid_names),
        organizations_with_valid_addresses = as.numeric(organizations_with_valid_addresses),
        high_quality_organizations = as.numeric(high_quality_organizations),
        low_quality_organizations = as.numeric(low_quality_organizations),
        identifier_percentage = as.numeric(identifier_percentage),
        name_percentage = as.numeric(name_percentage),
        address_percentage = as.numeric(address_percentage)
      )
    
    return(result)
  })
  
  # Get identifier breakdown summary
  filtered_identifier_summary <- reactive({
    current_vendor <- input$vendor_filter
    if (is.null(current_vendor)) current_vendor <- "All Developers"
    
    query_str <- "SELECT * FROM mv_organization_identifier_summary WHERE vendor_name = {vendor}"
    
    data_query <- glue::glue_sql(query_str, vendor = current_vendor, .con = db_connection)
    
    result <- tbl(db_connection, sql(data_query)) %>% collect()
    
    if (nrow(result) == 0) {
      # Return default values
      return(data.frame(
        vendor_name = current_vendor,
        total_npi = 0, total_clia = 0, total_naic = 0, total_other = 0, total_no_identifiers = 0,
        total_npi_valid = 0, total_clia_valid = 0, total_naic_valid = 0,
        total_npi_invalid = 0, total_clia_invalid = 0, total_naic_invalid = 0,
        total_other_invalid = 0, total_all_identifiers = 0, total_all_conformant = 0,
        npi_percentage = 0, clia_percentage = 0, naic_percentage = 0, other_percentage = 0, conformance_rate = 0,
        stringsAsFactors = FALSE
      ))
    }
    
    # Ensure numeric columns are properly typed
    result <- result %>%
      mutate(
        total_npi = as.numeric(total_npi),
        total_clia = as.numeric(total_clia),
        total_naic = as.numeric(total_naic),
        total_other = as.numeric(total_other),
        total_no_identifiers = as.numeric(total_no_identifiers),
        total_npi_valid = as.numeric(total_npi_valid),
        total_clia_valid = as.numeric(total_clia_valid),
        total_naic_valid = as.numeric(total_naic_valid),
        total_npi_invalid = as.numeric(total_npi_invalid),
        total_clia_invalid = as.numeric(total_clia_invalid),
        total_naic_invalid = as.numeric(total_naic_invalid),
        total_other_invalid = as.numeric(total_other_invalid),
        total_all_identifiers = as.numeric(total_all_identifiers),
        total_all_conformant = as.numeric(total_all_conformant),
        npi_percentage = as.numeric(npi_percentage),
        clia_percentage = as.numeric(clia_percentage),
        naic_percentage = as.numeric(naic_percentage),
        other_percentage = as.numeric(other_percentage),
        conformance_rate = as.numeric(conformance_rate)
      )
    
    return(result)
  })
  
  # Get individual organization data for detailed views and downloads
  filtered_org_data <- reactive({
    current_vendor <- input$vendor_filter
    if (is.null(current_vendor)) current_vendor <- "All Developers"
    
    # Query the detailed organization quality data
    if (current_vendor == "All Developers") {
      query_str <- "SELECT * FROM mv_organization_quality"
      data_query <- glue::glue_sql(query_str, .con = db_connection)
    } else {
      query_str <- "SELECT * FROM mv_organization_quality WHERE vendor_names_array && ARRAY[{vendor}]"
      data_query <- glue::glue_sql(query_str, vendor = current_vendor, .con = db_connection)
    }
    
    result <- tbl(db_connection, sql(data_query)) %>% collect()
    
    return(result)
  })
  
  # Summary statistics using materialized view data 
  quality_summary <- reactive({
    summary_data <- filtered_quality_summary()
    
    if (nrow(summary_data) == 0) {
      return(list(
        total_orgs = 0,
        valid_identifier_count = 0,
        valid_name_count = 0,
        valid_address_count = 0,
        high_quality_count = 0,
        low_quality_count = 0,
        identifier_percentage = 0,
        name_percentage = 0,
        address_percentage = 0,
        no_identifiers = 0,
        invalid_only = 0,
        all_valid = 0
      ))
    }
    
    # Extract the first (and only) row
    row <- summary_data[1, ]
    
    # Convert to list with proper numeric values
    list(
      total_orgs = as.numeric(row$total_organizations),
      valid_identifier_count = as.numeric(row$organizations_with_valid_identifiers),
      valid_name_count = as.numeric(row$organizations_with_valid_names),
      valid_address_count = as.numeric(row$organizations_with_valid_addresses),
      high_quality_count = as.numeric(row$high_quality_organizations),
      low_quality_count = as.numeric(row$low_quality_organizations),
      identifier_percentage = as.numeric(row$identifier_percentage),
      name_percentage = as.numeric(row$name_percentage),
      address_percentage = as.numeric(row$address_percentage),
      no_identifiers = as.numeric(row$organizations_with_no_identifiers),
      invalid_only = as.numeric(row$organizations_with_invalid_only),
      all_valid = as.numeric(row$organizations_all_valid)
    )
  })
  
  # Identifier summary using materialized view data
  identifier_type_summary <- reactive({
    id_data <- filtered_identifier_summary()
    summary_data <- filtered_quality_summary()

    if (nrow(id_data) == 0 || nrow(summary_data) == 0) {
      return(list(
        npi_count = 0, clia_count = 0, naic_count = 0, other_count = 0, no_identifier_count = 0,
        npi_valid = 0, clia_valid = 0, naic_valid = 0,
        npi_invalid = 0, clia_invalid = 0, naic_invalid = 0, other_invalid = 0,
        total_identifiers = 0, total_conformant = 0,
        npi_percentage = 0, clia_percentage = 0, naic_percentage = 0, other_percentage = 0,
        no_identifier_percentage = 0, conformance_rate = 0,
        orgs_with_no_identifiers = 0, orgs_with_invalid_only = 0, orgs_with_valid = 0,
        total_organizations = 0
      ))
    }

    id_row <- id_data[1, ]
    summary_row <- summary_data[1, ]

    # Convert to list with proper numeric values
    list(
      npi_count = as.numeric(id_row$total_npi),
      clia_count = as.numeric(id_row$total_clia),
      naic_count = as.numeric(id_row$total_naic),
      other_count = as.numeric(id_row$total_other),
      no_identifier_count = as.numeric(id_row$total_no_identifiers),
      npi_valid = as.numeric(id_row$total_npi_valid),
      clia_valid = as.numeric(id_row$total_clia_valid),
      naic_valid = as.numeric(id_row$total_naic_valid),
      npi_invalid = as.numeric(id_row$total_npi_invalid),
      clia_invalid = as.numeric(id_row$total_clia_invalid),
      naic_invalid = as.numeric(id_row$total_naic_invalid),
      other_invalid = as.numeric(id_row$total_other_invalid),
      total_identifiers = as.numeric(id_row$total_all_identifiers),
      total_conformant = as.numeric(id_row$total_all_conformant),
      npi_percentage = as.numeric(id_row$npi_percentage),
      clia_percentage = as.numeric(id_row$clia_percentage),
      naic_percentage = as.numeric(id_row$naic_percentage),
      other_percentage = as.numeric(id_row$other_percentage),
      no_identifier_percentage = if(as.numeric(summary_row$total_organizations) > 0)
        round(as.numeric(id_row$total_no_identifiers) / as.numeric(summary_row$total_organizations) * 100, 1) else 0,
      conformance_rate = as.numeric(id_row$conformance_rate),
      orgs_with_no_identifiers = as.numeric(summary_row$organizations_with_no_identifiers),
      orgs_with_invalid_only = as.numeric(summary_row$organizations_with_invalid_only),
      orgs_with_valid = as.numeric(summary_row$organizations_with_valid_identifiers),
      total_organizations = as.numeric(summary_row$total_organizations)
    )
  })

  # Data issues summary - system-wide statistics
  data_issues_summary <- reactive({
    # Query the data issues summary materialized view
    query_str <- "SELECT * FROM mv_data_issues_summary LIMIT 1"

    result <- tbl(db_connection, sql(query_str)) %>% collect()

    if (nrow(result) == 0) {
      return(list(
        developers_with_no_org_data_count = 0,
        endpoints_with_no_org_data_count = 0,
        shared_list_sources_count = 0,
        developers_sharing_list_sources_count = 0,
        inaccessible_list_sources_count = 0,
        endpoints_with_inaccessible_list_sources_count = 0,
        developers_with_empty_bundles_count = 0
      ))
    }

    # Extract the first (and only) row
    row <- result[1, ]

    # Convert to list with proper numeric values
    list(
      developers_with_no_org_data_count = as.numeric(row$developers_with_no_org_data_count),
      endpoints_with_no_org_data_count = as.numeric(row$endpoints_with_no_org_data_count),
      shared_list_sources_count = as.numeric(row$shared_list_sources_count),
      developers_sharing_list_sources_count = as.numeric(row$developers_sharing_list_sources_count),
      inaccessible_list_sources_count = as.numeric(row$inaccessible_list_sources_count),
      endpoints_with_inaccessible_list_sources_count = as.numeric(row$endpoints_with_inaccessible_list_sources_count),
      developers_with_empty_bundles_count = as.numeric(row$developers_with_empty_bundles_count)
    )
  })

  # Developer data issues - comprehensive view
  developer_data_issues <- reactive({
    # Query the comprehensive developer data issues view
    query_str <- "SELECT * FROM mv_developer_data_issues ORDER BY
                  no_org_data_endpoints DESC,
                  vendor_name"

    result <- tbl(db_connection, sql(query_str)) %>% collect()

    return(result)
  })
  
  # Render summary outputs
  output$total_orgs <- renderText({
    format(quality_summary()$total_orgs, big.mark = ",")
  })
  
  output$high_quality_count <- renderText({
    format(quality_summary()$high_quality_count, big.mark = ",")
  })
  
  output$low_quality_count <- renderText({
    format(quality_summary()$low_quality_count, big.mark = ",")
  })
  
  output$identifier_percentage <- renderText({
    paste0(quality_summary()$identifier_percentage, "%")
  })
  
  output$name_percentage <- renderText({
    paste0(quality_summary()$name_percentage, "%")
  })
  
  output$address_percentage <- renderText({
    paste0(quality_summary()$address_percentage, "%")
  })
  
  output$identifier_progress_width <- renderText({
    paste0(quality_summary()$identifier_percentage, "%")
  })
  
  output$name_progress_width <- renderText({
    paste0(quality_summary()$name_percentage, "%")
  })
  
  output$address_progress_width <- renderText({
    paste0(quality_summary()$address_percentage, "%")
  })
  
  # Identifier breakdown displays
  output$valid_identifier_count_display <- renderText({
    summary <- quality_summary()
    paste0(format(summary$valid_identifier_count, big.mark = ","), " (", summary$identifier_percentage, "%)")
  })
  
  output$no_identifier_count_display <- renderText({
    summary <- quality_summary()
    id_summary <- identifier_type_summary()
    paste0(format(id_summary$orgs_with_no_identifiers, big.mark = ","), " (", id_summary$no_identifier_percentage, "%)")
  })
  
  output$invalid_only_count_display <- renderText({
    summary <- quality_summary()
    id_summary <- identifier_type_summary()
    invalid_only_count <- id_summary$orgs_with_invalid_only
    invalid_only_percentage <- if(summary$total_orgs > 0) round(invalid_only_count / summary$total_orgs * 100, 1) else 0
    paste0(format(invalid_only_count, big.mark = ","), " (", invalid_only_percentage, "%)")
  })
  
  # Chart outputs using pre-computed data with modern theme
  output$quality_overview_chart <- renderPlot({
    req(quality_summary())
    
    summary <- quality_summary()
    
    chart_data <- data.frame(
      Category = c("Identifier Type Validation", "Organization Name", "Address Completeness"),
      Valid = c(
        as.numeric(summary$valid_identifier_count),
        as.numeric(summary$valid_name_count),
        as.numeric(summary$valid_address_count)
      ),
      Invalid = c(
        as.numeric(summary$total_orgs) - as.numeric(summary$valid_identifier_count),
        as.numeric(summary$total_orgs) - as.numeric(summary$valid_name_count),
        as.numeric(summary$total_orgs) - as.numeric(summary$valid_address_count)
      ),
      stringsAsFactors = FALSE
    )
    
    if (sum(chart_data$Valid) == 0 && sum(chart_data$Invalid) == 0) {
      return(
        ggplot() + 
          geom_text(aes(x = 0.5, y = 0.5, label = "No data available"), size = 6, color = "#7f8c8d") +
          xlim(0, 1) + ylim(0, 1) + theme_void()
      )
    }
    
    chart_data_long <- chart_data %>%
      pivot_longer(cols = c(Valid, Invalid), names_to = "Status", values_to = "Count")
    
    ggplot(chart_data_long, aes(x = Category, y = Count, fill = Status)) +
      geom_col(position = "dodge", width = 0.7) +
      geom_text(aes(label = format(Count, big.mark = ",")), 
                position = position_dodge(width = 0.7), vjust = -0.5, 
                fontface = "bold", size = 4) +
      scale_fill_manual(values = c("Valid" = "#28a745", "Invalid" = "#dc3545")) +
      labs(x = NULL, y = "Number of Organizations") +
      theme_minimal() +
      theme(
        axis.text.x = element_text(angle = 30, hjust = 1, size = 11, face = "bold"),
        axis.text.y = element_text(size = 10),
        axis.title.y = element_text(size = 12, face = "bold", margin = margin(r = 10)),
        legend.position = "bottom",
        legend.title = element_blank(),
        legend.text = element_text(size = 11, face = "bold"),
        panel.grid.major.x = element_blank(),
        panel.grid.minor = element_blank(),
        plot.margin = margin(10, 10, 10, 10)
      )
  }, height = 400)
  
  # Organization identifier status breakdown chart
  output$organization_identifier_status_chart <- renderPlot({
    req(identifier_type_summary())
    
    id_summary <- identifier_type_summary()
    
    status_data <- data.frame(
      Status = c("Valid Identifiers", 
                 "No Identifiers", 
                 "Only Invalid Identifiers"),
      Count = c(
        as.numeric(id_summary$orgs_with_valid),
        as.numeric(id_summary$orgs_with_no_identifiers),
        as.numeric(id_summary$orgs_with_invalid_only)
      ),
      stringsAsFactors = FALSE
    )
    
    total_orgs <- sum(status_data$Count)
    if (total_orgs > 0) {
      status_data$Percentage <- round(status_data$Count / total_orgs * 100, 1)
    } else {
      status_data$Percentage <- 0
      return(
        ggplot() + 
          geom_text(aes(x = 0.5, y = 0.5, label = "No data available"), size = 6, color = "#7f8c8d") +
          xlim(0, 1) + ylim(0, 1) + theme_void()
      )
    }
    
    colors <- c("Valid Identifiers" = "#28a745",
                "No Identifiers" = "#6c757d", 
                "Only Invalid Identifiers" = "#dc3545")
    
    ggplot(status_data, aes(x = reorder(Status, Count), y = Count, fill = Status)) +
      geom_col(width = 0.6) +
      geom_text(aes(label = paste0(format(Count, big.mark = ","), "\n(", Percentage, "%)")), 
                hjust = -0.1, size = 3.5, fontface = "bold") +
      scale_fill_manual(values = colors) +
      coord_flip() +
      labs(x = NULL, y = "Number of Organizations") +
      theme_minimal() +
      theme(
        legend.position = "none",
        axis.text.y = element_text(size = 10, face = "bold"),
        axis.text.x = element_text(size = 10),
        axis.title.x = element_text(size = 11, face = "bold", margin = margin(t = 10)),
        panel.grid.major.y = element_blank(),
        panel.grid.minor = element_blank()
      ) +
      scale_y_continuous(expand = expansion(mult = c(0, 0.2)))
  }, height = 300)
  
  # Identifier type distribution chart
  output$identifier_type_distribution_chart <- renderPlot({
    req(identifier_type_summary())
    
    id_summary <- identifier_type_summary()
    
    chart_data <- data.frame(
      Type = c("NPI", "CLIA", "NAIC", "Other", "No Data"),
      Count = c(
        as.numeric(id_summary$npi_count), 
        as.numeric(id_summary$clia_count), 
        as.numeric(id_summary$naic_count), 
        as.numeric(id_summary$other_count), 
        as.numeric(id_summary$no_identifier_count)
      ),
      stringsAsFactors = FALSE
    )
    
    chart_data <- chart_data[chart_data$Count > 0, ]
    
    if (nrow(chart_data) == 0) {
      return(
        ggplot() + 
          geom_text(aes(x = 0.5, y = 0.5, label = "No identifier data found"), 
                   size = 6, color = "#7f8c8d") +
          theme_void() + xlim(0, 1) + ylim(0, 1)
      )
    }
    
    type_colors <- c("NPI" = "#28a745", "CLIA" = "#007bff", "NAIC" = "#fd7e14", 
                    "Other" = "#dc3545", "No Data" = "#6c757d")
    
    ggplot(chart_data, aes(x = reorder(Type, Count), y = Count, fill = Type)) +
      geom_col(width = 0.6) +
      geom_text(aes(label = format(Count, big.mark = ",")), 
                hjust = -0.1, fontface = "bold", size = 3.5) +
      scale_fill_manual(values = type_colors) +
      coord_flip() +
      labs(x = NULL, y = "Count") +
      theme_minimal() +
      theme(
        legend.position = "none",
        axis.text.y = element_text(size = 10, face = "bold"),
        axis.text.x = element_text(size = 10),
        axis.title.x = element_text(size = 11, face = "bold", margin = margin(t = 10)),
        panel.grid.major.y = element_blank(),
        panel.grid.minor = element_blank()
      ) +
      scale_y_continuous(expand = expansion(mult = c(0, 0.15)))
  }, height = 350)
  
  # Conformance by type chart
  output$conformance_by_type_chart <- renderPlot({
    req(identifier_type_summary())
    
    id_summary <- identifier_type_summary()
    
    conformance_data <- data.frame(
      Type = c("NPI", "CLIA", "NAIC"),
      Valid = c(
        as.numeric(id_summary$npi_valid), 
        as.numeric(id_summary$clia_valid), 
        as.numeric(id_summary$naic_valid)
      ),
      Invalid = c(
        as.numeric(id_summary$npi_invalid), 
        as.numeric(id_summary$clia_invalid), 
        as.numeric(id_summary$naic_invalid)
      ),
      stringsAsFactors = FALSE
    ) %>%
      filter(Valid + Invalid > 0)
    
    if (nrow(conformance_data) == 0) {
      return(
        ggplot() + 
          geom_text(aes(x = 0.5, y = 0.5, label = "No conformance data available"), 
                   size = 6, color = "#7f8c8d") +
          theme_void() + xlim(0, 1) + ylim(0, 1)
      )
    }
    
    conformance_long <- conformance_data %>%
      pivot_longer(cols = c(Valid, Invalid), names_to = "Status", values_to = "Count")
    
    ggplot(conformance_long, aes(x = Type, y = Count, fill = Status)) +
      geom_col(position = "stack") +
      geom_text(aes(label = Count), position = position_stack(vjust = 0.5), 
                fontface = "bold", color = "white", size = 4) +
      scale_fill_manual(values = c("Valid" = "#28a745", "Invalid" = "#dc3545")) +
      labs(x = "Identifier Type", y = "Count") +
      theme_minimal() +
      theme(
        axis.text.x = element_text(size = 11, face = "bold"),
        axis.text.y = element_text(size = 10),
        axis.title = element_text(size = 11, face = "bold"),
        legend.position = "bottom",
        legend.title = element_blank(),
        legend.text = element_text(size = 10, face = "bold"),
        panel.grid.major.x = element_blank(),
        panel.grid.minor = element_blank()
      )
  }, height = 350)
  
  # Identifier type detail table
  output$identifier_type_table <- reactable::renderReactable({
    req(identifier_type_summary())
    
    id_summary <- identifier_type_summary()
    
    type_data <- data.frame(
      Identifier_Type = c("NPI", "CLIA", "NAIC", "Other", "No Identifier Data"),
      Total_Count = c(
        as.numeric(id_summary$npi_count), 
        as.numeric(id_summary$clia_count), 
        as.numeric(id_summary$naic_count), 
        as.numeric(id_summary$other_count),
        as.numeric(id_summary$no_identifier_count)
      ),
      Valid_Count = c(
        as.numeric(id_summary$npi_valid), 
        as.numeric(id_summary$clia_valid),
        as.numeric(id_summary$naic_valid), 
        0, 0
      ),
      Invalid_Count = c(
        as.numeric(id_summary$npi_invalid), 
        as.numeric(id_summary$clia_invalid),
        as.numeric(id_summary$naic_invalid), 
        as.numeric(id_summary$other_invalid), 
        0
      ),
      Conformance_Rate = c(
        if(id_summary$npi_count > 0) paste0(round(id_summary$npi_valid / id_summary$npi_count * 100, 1), "%") else "N/A",
        if(id_summary$clia_count > 0) paste0(round(id_summary$clia_valid / id_summary$clia_count * 100, 1), "%") else "N/A",
        if(id_summary$naic_count > 0) paste0(round(id_summary$naic_valid / id_summary$naic_count * 100, 1), "%") else "N/A",
        "0%",
        "N/A"
      ),
      Percentage_of_Orgs = c(
        paste0(id_summary$npi_percentage, "%"),
        paste0(id_summary$clia_percentage, "%"),
        paste0(id_summary$naic_percentage, "%"),
        paste0(id_summary$other_percentage, "%"),
        paste0(id_summary$no_identifier_percentage, "%")
      ),
      US_Core_Rules = c("us-core-16, us-core-17", "us-core-18", "us-core-19", "Non-Compliant", "Missing Data"),
      Validation_Requirements = c(
        "10 digits + Luhn check digit",
        "2 digits + 'D' + 7 digits", 
        "5 digits",
        "Non-standard format",
        "No identifier provided"
      ),
      stringsAsFactors = FALSE
    )
    
    reactable(
      type_data,
      columns = list(
        Identifier_Type = colDef(name = "Type", width = 120, 
                                 style = list(fontWeight = 600)),
        Total_Count = colDef(name = "Total", format = colFormat(separators = TRUE), width = 80),
        Valid_Count = colDef(name = "Valid", format = colFormat(separators = TRUE), width = 80,
                            style = function(value) {
                              if (value > 0) list(color = "#28a745", fontWeight = 600)
                            }),
        Invalid_Count = colDef(name = "Invalid", format = colFormat(separators = TRUE), width = 80,
                              style = function(value) {
                                if (value > 0) list(color = "#dc3545", fontWeight = 600)
                              }),
        Conformance_Rate = colDef(
          name = "Conformance Rate", 
          width = 130,
          cell = function(value) {
            if (value == "N/A") {
              div(style = "color: #6c757d; font-weight: 500;", value)
            } else {
              rate <- as.numeric(str_extract(value, "\\d+"))
              if (!is.na(rate)) {
                if (rate >= 90) {
                  div(style = "color: #28a745; font-weight: 700; font-size: 14px;", value)
                } else if (rate >= 70) {
                  div(style = "color: #ffc107; font-weight: 700; font-size: 14px;", value)  
                } else {
                  div(style = "color: #dc3545; font-weight: 700; font-size: 14px;", value)
                }
              } else {
                div(style = "color: #6c757d; font-weight: 500;", value)
              }
            }
          }
        ),
        Percentage_of_Orgs = colDef(name = "% of Orgs", width = 100),
        US_Core_Rules = colDef(name = "US-Core Rules", width = 150,
                              style = list(fontSize = "13px", color = "#5a6c7d")),
        Validation_Requirements = colDef(name = "Format Requirements", minWidth = 200,
                                        style = list(fontSize = "13px", color = "#5a6c7d"))
      ),
      striped = TRUE,
      highlight = TRUE,
      bordered = TRUE,
      theme = reactableTheme(
        borderColor = "#e0e0e0",
        stripedColor = "#f8f9fa",
        highlightColor = "#f0f8ff",
        headerStyle = list(
          background = "#1B5A7F",
          color = "white",
          fontWeight = 600,
          fontSize = "14px"
        )
      )
    )
  })
  
  # Issues detail table
  output$issues_detail_table <- reactable::renderReactable({
    req(quality_summary(), identifier_type_summary())
    
    summary <- quality_summary()
    id_summary <- identifier_type_summary()
    
    issues_data <- data.frame(
      Issue_Category = c("Identifier Type Validation", "Organization Names", "Address Completeness"),
      Total_Count = rep(as.numeric(summary$total_orgs), 3),
      Valid_Count = c(
        as.numeric(summary$valid_identifier_count), 
        as.numeric(summary$valid_name_count), 
        as.numeric(summary$valid_address_count)
      ),
      Invalid_Count = c(
        as.numeric(summary$total_orgs) - as.numeric(summary$valid_identifier_count),
        as.numeric(summary$total_orgs) - as.numeric(summary$valid_name_count),
        as.numeric(summary$total_orgs) - as.numeric(summary$valid_address_count)
      ),
      Success_Rate = c(
        paste0(summary$identifier_percentage, "%"),
        paste0(summary$name_percentage, "%"),
        paste0(summary$address_percentage, "%")
      ),
      Common_Issues = c(
        paste0("HTI-1 Final Rule violations: No identifier data (", format(id_summary$no_identifier_count, big.mark = ","), "), ",
               "invalid NPI check digits (", format(id_summary$npi_invalid, big.mark = ","), "), ",
               "incorrect CLIA format (", format(id_summary$clia_invalid, big.mark = ","), "), ",
               "wrong NAIC length (", format(id_summary$naic_invalid, big.mark = ","), "), ",
               "non-standard types (", format(id_summary$other_count, big.mark = ","), ")"),
        "Placeholder names (-, ., N/A), names too short (<3 chars), excessive special characters",
        "Missing street/city/state/ZIP, placeholder addresses (123 Main St), incomplete components"
      ),
      US_Core_Reference = c(
        "https://build.fhir.org/ig/HL7/US-Core/StructureDefinition-us-core-organization.html",
        "https://build.fhir.org/ig/HL7/US-Core/StructureDefinition-us-core-organization.html",
        "https://build.fhir.org/ig/HL7/US-Core/StructureDefinition-us-core-organization.html"
      ),
      stringsAsFactors = FALSE
    )
    
    reactable(
      issues_data,
      columns = list(
        Issue_Category = colDef(name = "Issue Category", width = 180,
                               style = list(fontWeight = 600, color = "#2c3e50")),
        Total_Count = colDef(name = "Total", format = colFormat(separators = TRUE), width = 90),
        Valid_Count = colDef(name = "Valid", format = colFormat(separators = TRUE), width = 90,
                            style = function(value) {
                              list(color = "#28a745", fontWeight = 600)
                            }),
        Invalid_Count = colDef(name = "Invalid", format = colFormat(separators = TRUE), width = 90,
                              style = function(value) {
                                list(color = "#dc3545", fontWeight = 600)
                              }),
        Success_Rate = colDef(name = "Success Rate", width = 110,
                             style = list(fontWeight = 600, fontSize = "14px")),
        Common_Issues = colDef(name = "Common Issues", minWidth = 350,
                              style = list(fontSize = "13px", color = "#5a6c7d", lineHeight = "1.5")),
        US_Core_Reference = colDef(
          name = "US-Core Reference",
          width = 150,
          cell = function(value) {
            tags$a(href = value, target = "_blank", 
                  style = "color: #1B5A7F; font-weight: 500; text-decoration: none;",
                  "View Specification")
          }
        )
      ),
      striped = TRUE,
      highlight = TRUE,
      bordered = TRUE,
      theme = reactableTheme(
        borderColor = "#e0e0e0",
        stripedColor = "#f8f9fa",
        highlightColor = "#f0f8ff",
        headerStyle = list(
          background = "#1B5A7F",
          color = "white",
          fontWeight = 600,
          fontSize = "14px"
        )
      )
    )
  })
  
  # Data Issues outputs
  output$developers_no_org_data_count <- renderText({
    format(data_issues_summary()$developers_with_no_org_data_count, big.mark = ",")
  })

  output$endpoints_no_org_data_count <- renderText({
    format(data_issues_summary()$endpoints_with_no_org_data_count, big.mark = ",")
  })

  output$developers_sharing_list_sources_count <- renderText({
    format(data_issues_summary()$developers_sharing_list_sources_count, big.mark = ",")
  })

  output$inaccessible_list_sources_count <- renderText({
    format(data_issues_summary()$inaccessible_list_sources_count, big.mark = ",")
  })

  output$developers_empty_bundles_count <- renderText({
    format(data_issues_summary()$developers_with_empty_bundles_count, big.mark = ",")
  })

  # Comprehensive developer data issues table
  output$developer_data_issues_table <- reactable::renderReactable({
    req(developer_data_issues())

    dev_data <- developer_data_issues()

    if (nrow(dev_data) == 0) {
      # Return empty state
      dev_data <- data.frame(
        vendor_name = "No data issues found",
        total_endpoints = 0,
        endpoints_with_org_data = 0,
        no_org_data_endpoints = 0,
        accessible_endpoints = 0,
        inaccessible_endpoints = 0,
        organization_count = 0,
        data_completeness_percentage = 100,
        has_empty_bundle = FALSE,
        stringsAsFactors = FALSE
      )
    }

    reactable(
      dev_data,
      filterable = TRUE,
      searchable = TRUE,
      defaultPageSize = 20,
      columns = list(
        vendor_name = colDef(
          name = "Developer Name",
          minWidth = 200,
          style = list(fontWeight = 600, color = "#2c3e50")
        ),
        total_endpoints = colDef(
          name = "Total Endpoints",
          width = 120,
          format = colFormat(separators = TRUE),
          align = "center"
        ),
        endpoints_with_org_data = colDef(
          name = "With Org Data",
          width = 120,
          format = colFormat(separators = TRUE),
          align = "center",
          style = function(value) {
            if (value > 0) list(color = "#28a745", fontWeight = 600)
            else list(color = "#dc3545", fontWeight = 600)
          }
        ),
        no_org_data_endpoints = colDef(
          name = "No Org Data",
          width = 120,
          format = colFormat(separators = TRUE),
          align = "center",
          style = function(value) {
            if (value > 0) list(color = "#dc3545", fontWeight = 700)
            else list(color = "#6c757d")
          }
        ),
        accessible_endpoints = colDef(
          name = "Accessible",
          width = 100,
          format = colFormat(separators = TRUE),
          align = "center",
          style = function(value) {
            if (value > 0) list(color = "#28a745", fontWeight = 600)
            else list(color = "#6c757d")
          }
        ),
        inaccessible_endpoints = colDef(
          name = "Inaccessible",
          width = 110,
          format = colFormat(separators = TRUE),
          align = "center",
          style = function(value) {
            if (value > 0) list(color = "#dc3545", fontWeight = 600)
            else list(color = "#6c757d")
          }
        ),
        organization_count = colDef(
          name = "Organizations",
          width = 120,
          format = colFormat(separators = TRUE),
          align = "center",
          style = function(value) {
            if (value == 0) list(color = "#dc3545", fontWeight = 600)
            else list(color = "#28a745", fontWeight = 600)
          }
        ),
        data_completeness_percentage = colDef(
          name = "Completeness %",
          width = 130,
          format = colFormat(digits = 1, suffix = "%"),
          align = "center",
          style = function(value) {
            if (value == 0) list(color = "#dc3545", fontWeight = 700, backgroundColor = "#fff5f5")
            else if (value < 50) list(color = "#ffc107", fontWeight = 600, backgroundColor = "#fffbf0")
            else if (value < 100) list(color = "#17a2b8", fontWeight = 600)
            else list(color = "#28a745", fontWeight = 600)
          }
        ),
        has_empty_bundle = colDef(
          name = "Empty Bundle",
          width = 120,
          align = "center",
          cell = function(value) {
            if (value == TRUE) {
              tags$span(
                style = "color: #dc3545; font-weight: 700;",
                tags$i(class = "fa fa-check-circle", style = "margin-right: 5px;"),
                "Yes"
              )
            } else {
              tags$span(
                style = "color: #6c757d;",
                tags$i(class = "fa fa-times-circle", style = "margin-right: 5px;"),
                "No"
              )
            }
          }
        )
      ),
      striped = TRUE,
      highlight = TRUE,
      bordered = TRUE,
      defaultSorted = list(no_org_data_endpoints = "desc"),
      theme = reactableTheme(
        borderColor = "#e0e0e0",
        stripedColor = "#f8f9fa",
        highlightColor = "#f0f8ff",
        headerStyle = list(
          background = "#1B5A7F",
          color = "white",
          fontWeight = 600,
          fontSize = "13px"
        )
      )
    )
  })

  # Enhanced recommendations
  output$recommendations <- renderUI({
    req(quality_summary(), identifier_type_summary())

    summary <- quality_summary()
    id_summary <- identifier_type_summary()
    recommendations <- list()
    
    # No identifier data alert
    if (id_summary$no_identifier_count > 0) {
      no_id_percentage <- round(id_summary$no_identifier_count / summary$total_orgs * 100, 1)
      recommendations <- append(recommendations,
        tags$div(class = "alert alert-danger",
          tags$strong(tags$i(class = "fa fa-times-circle", style = "margin-right: 5px;"), 
                     "Missing Identifier Data: "),
          paste0(format(id_summary$no_identifier_count, big.mark = ","), 
                 " organizations (", no_id_percentage, "%) have no identifier data."),
          tags$br(),
          tags$small("Organizations must include at least one identifier (NPI, CLIA, or NAIC) to meet US-Core requirements.")
        )
      )
    }
    
    # Invalid only identifiers alert
    if (id_summary$orgs_with_invalid_only > 0) {
      invalid_only_percentage <- round(id_summary$orgs_with_invalid_only / summary$total_orgs * 100, 1)
      recommendations <- append(recommendations,
        tags$div(class = "alert alert-danger",
          tags$strong(tags$i(class = "fa fa-exclamation-triangle", style = "margin-right: 5px;"),
                     "Organizations with Only Invalid Identifiers: "),
          paste0(format(id_summary$orgs_with_invalid_only, big.mark = ","), 
                 " organizations (", invalid_only_percentage, "%) have identifiers but none are US-Core compliant."),
          tags$br(),
          tags$small("Review identifier formats and ensure compliance with US-Core validation rules.")
        )
      )
    }
    
    # Identifier conformance recommendations
    if (summary$identifier_percentage < 80) {
      recommendations <- append(recommendations, 
        tags$div(class = "alert alert-warning",
          tags$strong(tags$i(class = "fa fa-clipboard-check", style = "margin-right: 5px;"),
                     "US-Core Identifier Conformance Issues: "),
          paste0("Only ", summary$identifier_percentage, "% of organizations have conformant identifiers."),
          tags$br(),
          tags$small("Ensure NPI identifiers are 10 digits with valid check digits, CLIA identifiers follow 2D7 format, and NAIC identifiers are 5 digits.")
        )
      )
    }
    
    if (id_summary$other_count > 0) {
      recommendations <- append(recommendations, 
        tags$div(class = "alert alert-warning",
          tags$strong(tags$i(class = "fa fa-question-circle", style = "margin-right: 5px;"),
                     "Non-Standard Identifiers: "),
          paste0("Found ", format(id_summary$other_count, big.mark = ","), 
                 " non-standard identifier types."),
          tags$br(),
          tags$small("Use US-Core compliant types: NPI (healthcare providers), CLIA (laboratories), NAIC (insurance).")
        )
      )
    }
    
    # Specific validation error recommendations
    if (id_summary$npi_invalid > 0) {
      recommendations <- append(recommendations,
        tags$div(class = "alert alert-warning",
          tags$strong(tags$i(class = "fa fa-id-badge", style = "margin-right: 5px;"),
                     "Invalid NPI Identifiers: "),
          paste0(format(id_summary$npi_invalid, big.mark = ","), " NPIs failed validation (us-core-16/17)."),
          tags$br(),
          tags$small("Verify NPIs are exactly 10 digits and have valid Luhn check digits.")
        )
      )
    }
    
    if (id_summary$clia_invalid > 0) {
      recommendations <- append(recommendations,
        tags$div(class = "alert alert-warning",
          tags$strong(tags$i(class = "fa fa-flask", style = "margin-right: 5px;"),
                     "Invalid CLIA Identifiers: "),
          paste0(format(id_summary$clia_invalid, big.mark = ","), " CLIAs failed validation (us-core-18)."),
          tags$br(),
          tags$small("CLIA format must be: 2 digits + 'D' + 7 digits (e.g., '12D3456789').")
        )
      )
    }
    
    if (id_summary$naic_invalid > 0) {
      recommendations <- append(recommendations,
        tags$div(class = "alert alert-warning",
          tags$strong(tags$i(class = "fa fa-shield-alt", style = "margin-right: 5px;"),
                     "Invalid NAIC Identifiers: "),
          paste0(format(id_summary$naic_invalid, big.mark = ","), " NAICs failed validation (us-core-19)."),
          tags$br(),
          tags$small("NAIC identifiers must be exactly 5 digits.")
        )
      )
    }
    
    if (summary$name_percentage < 80) {
      recommendations <- append(recommendations,
        tags$div(class = "alert alert-info",
          tags$strong(tags$i(class = "fa fa-building", style = "margin-right: 5px;"),
                     "Name Quality: "),
          "Use complete, meaningful organization names instead of placeholders."
        )
      )
    }
    
    if (summary$address_percentage < 80) {
      recommendations <- append(recommendations,
        tags$div(class = "alert alert-secondary",
          tags$strong(tags$i(class = "fa fa-map-marker-alt", style = "margin-right: 5px;"),
                     "Address Issues: "),
          "Include complete addresses with street, city, state, and ZIP code."
        )
      )
    }
    
    if (length(recommendations) == 0) {
      recommendations <- list(
        tags$div(class = "alert alert-success",
          tags$strong(tags$i(class = "fa fa-check-circle", style = "margin-right: 8px;"),
                     "Excellent US-Core compliance!"),
          " Your organization data meets quality and conformance standards."
        )
      )
    }
    
    do.call(tagList, recommendations)
  })
  
  # Download handler
  output$download_feedback_report <- downloadHandler(
    filename = function() {
      paste0("organization_data_quality_report_", Sys.Date(), ".csv")
    },
    content = function(file) {
      data <- filtered_org_data()
      
      if (nrow(data) > 0) {
        report_data <- data %>%
          mutate(
            identifier_issues = ifelse(!has_valid_identifiers, "Missing or incomplete identifier data", "Valid"),
            name_issues = ifelse(!has_valid_name, "Placeholder name or too short", "Valid"),
            address_issues = ifelse(!has_valid_address, "Incomplete address information", "Valid"),
            quality_score = paste0(overall_quality_score, "/3"),
            conformance_summary = paste0(conformant_identifier_count, "/", total_identifier_count, " (", identifier_conformance_rate, "%)"),
            us_core_compliant = case_when(
              identifier_conformance_rate == 100 ~ "Fully Compliant",
              identifier_conformance_rate > 0 ~ "Partially Compliant", 
              TRUE ~ "Non-Compliant"
            ),
            clean_identifier_types = str_replace_all(identifier_types_html, "<br/>", "; "),
            clean_identifier_values = str_replace_all(identifier_values_html, "<br/>", "; "),
            identifier_status_description = case_when(
              identifier_status == "no_identifiers" ~ "No identifier data provided",
              identifier_status == "invalid_only" ~ "Has identifiers but all are invalid",
              identifier_status == "all_valid" ~ "All identifiers are valid",
              identifier_status == "mixed_valid_invalid" ~ "Mix of valid and invalid identifiers",
              TRUE ~ "Unknown status"
            )
          ) %>%
          select(
            organization_name,
            has_valid_identifiers, 
            has_valid_name, 
            has_valid_address, 
            overall_quality_score,
            conformant_identifier_count, 
            total_identifier_count, 
            identifier_conformance_rate,
            identifier_conformance_category,
            identifier_status,
            identifier_issues,
            name_issues,
            address_issues,
            quality_score,
            conformance_summary,
            us_core_compliant,
            clean_identifier_types,
            clean_identifier_values,
            identifier_status_description
          )
        
        write.csv(report_data, file, row.names = FALSE)
      } else {
        empty_data <- data.frame(
          organization_name = character(0),
          has_valid_identifiers = logical(0),
          message = "No data available for selected vendor"
        )
        write.csv(empty_data, file, row.names = FALSE)
      }
    }
  )
}