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
    # JS handler to toggle active CSS class on clickable cards
    tags$script(HTML("
      Shiny.addCustomMessageHandler('toggleCardActive', function(message) {
        var cardId = message.cardId;
        var active = message.active;
        var el = document.getElementById(cardId);
        if (el) {
          if (active) {
            el.classList.add('card-active');
          } else {
            el.classList.remove('card-active');
          }
        }
      });
      Shiny.addCustomMessageHandler('toggleCardDisabled', function(message) {
        var el = document.getElementById(message.cardId);
        if (el) {
          if (message.disabled) {
            el.classList.add('card-disabled');
            el.classList.remove('card-active');
          } else {
            el.classList.remove('card-disabled');
          }
        }
      });
    ")),
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
          height: 100%;
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
          margin-bottom: 0;
          border: 2px solid transparent;
          height: 100%;
          min-height: 140px;
          box-sizing: border-box;
          display: flex;
          flex-direction: column;
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

        /* Clickable card styling */
        .card-clickable {
          border: 2px solid #d0d0d0 !important;
          cursor: pointer;
          transition: transform 0.2s, box-shadow 0.2s, border-color 0.2s, background-color 0.2s !important;
        }

        .card-clickable:hover {
          transform: translateY(-3px);
          box-shadow: 0 6px 16px rgba(0,0,0,0.15) !important;
          border-color: #1B5A7F !important;
        }

        /* Active (toggled ON) card state */
        .card-active {
          border-color: #1B5A7F !important;
          background-color: #f0f7ff !important;
          box-shadow: 0 4px 12px rgba(27, 90, 127, 0.3) !important;
        }

        .card-active .metric-title {
          color: #1B5A7F;
          font-weight: 600;
        }
        
        /* Disabled card state - used for CHPL-only cards when Others filter is selected */
        .card-disabled {
          opacity: 0.45;
          pointer-events: none;
          cursor: default;
          filter: grayscale(60%);
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
        h2(class = "page-header", "Service Base URL Data Quality")
      )
    ),

    tabsetPanel(
      id = "main_tabs",
      type = "tabs",

      # ── TAB 1: Service Base URL Quality (CHPL / Developer level) ─────────
      tabPanel(
        title = "Developer Data Quality",
        value = "tier1",

        fluidRow(style = "margin-top: 20px;",
          column(width = 12,
            div(style = "background: linear-gradient(135deg, #f8f9fa 0%, #ffffff 100%);
                         padding: 15px; border-radius: 8px; margin-bottom: 20px;
                         border-left: 4px solid #1B5A7F;",
              p(style = "margin: 0; color: #5a6c7d; line-height: 1.6;",
                tags$strong("About this tab:"),
                " This tab shows CHPL Certified API Developer-level data quality — whether developers are",
                " publicly posting service base URLs in FHIR bundle format and whether those bundles",
                " return endpoint data."
              )
            )
          )
        ),

        # Data Issues in Lantern section
        fluidRow(
          column(width = 12,
            div(class = "modern-card",
              h3(class = "section-header",
                 tags$i(class = "fa fa-database", style = "margin-right: 8px;"),
                 "Developer Data Discrepancies"),

              # Timestamp status bar
              div(style = "background: #e8f4f8; border-left: 4px solid #17a2b8; padding: 8px 14px;
                           border-radius: 4px; margin-bottom: 16px; font-size: 0.88em; color: #2c3e50;",
                tags$i(class = "fa fa-clock-o", style = "margin-right: 6px; color: #17a2b8;"),
                tags$strong("CHPL data last fetched: "),
                textOutput(ns("chpl_last_updated"), inline = TRUE)
              ),

              # 5-card flex row: static Coverage Overview + 4 clickable cards
              div(style = "display: flex; gap: 14px; flex-wrap: nowrap; align-items: stretch; margin-bottom: 15px;",
                # Static card — Coverage Overview
                div(style = "flex: 1; min-width: 0;",
                  uiOutput(ns("counts_static_card"))
                ),
                # Clickable card 1 — Empty/Invalid FHIR Bundle URL
                div(style = "flex: 1; min-width: 0;",
                  div(class = "metric-card card-clickable",
                      id = ns("empty_bundles_card"),
                      onclick = sprintf("Shiny.setInputValue('%s', Math.random());", ns("empty_bundles_card_click")),
                    div(class = "metric-title",
                      tags$i(class = "fa fa-folder-open", style = "margin-right: 5px;"),
                      "Developers with an Empty/Invalid FHIR Bundle URL"
                    ),
                    div(class = "metric-value", style = "color: #dc3545;",
                      textOutput(ns("developers_empty_bundles_count"), inline = TRUE),
                      uiOutput(ns("empty_bundles_denom"), inline = TRUE)
                    ),
                    div(style = "margin-top: 8px; font-size: 0.82em; color: #7f8c8d;",
                      "CHPL developers with empty, unreachable, or invalid FHIR bundles",
                      tags$br(),
                      tags$span(style = "color: #1B5A7F; font-size: 0.9em; font-style: italic;",
                        tags$i(class = "fa fa-filter", style = "margin-right: 3px;"),
                        "Click to filter table below. Click again to reset."
                      )
                    )
                  )
                ),
                # Clickable card 2 — Shared FHIR Bundle Hyperlinks
                div(style = "flex: 1; min-width: 0;",
                  div(class = "metric-card card-clickable",
                      id = ns("shared_sources_card"),
                      onclick = sprintf("Shiny.setInputValue('%s', Math.random());", ns("shared_sources_card_click")),
                    div(class = "metric-title",
                      tags$i(class = "fa fa-share-alt", style = "margin-right: 5px;"),
                      "Shared FHIR Bundle Hyperlinks"
                    ),
                    div(class = "metric-value", style = "color: #ffc107;",
                      textOutput(ns("developers_sharing_list_sources_count"), inline = TRUE),
                      uiOutput(ns("shared_sources_denom"), inline = TRUE)
                    ),
                    div(style = "margin-top: 8px; font-size: 0.82em; color: #7f8c8d;",
                      "Developers sharing the same FHIR bundle URL",
                      tags$br(),
                      tags$span(style = "color: #1B5A7F; font-size: 0.9em; font-style: italic;",
                        tags$i(class = "fa fa-filter", style = "margin-right: 3px;"),
                        "Click to filter table below. Click again to reset."
                      )
                    )
                  )
                ),
                # Clickable card 3 — Developers Sharing FHIR Endpoints
                div(style = "flex: 1; min-width: 0;",
                  div(class = "metric-card card-clickable",
                      id = ns("shared_endpoints_card"),
                      onclick = sprintf("Shiny.setInputValue('%s', Math.random());", ns("shared_endpoints_card_click")),
                    div(class = "metric-title",
                      tags$i(class = "fa fa-code-fork", style = "margin-right: 5px;"),
                      "Developers Sharing FHIR Endpoints"
                    ),
                    div(class = "metric-value", style = "color: #ffc107;",
                      textOutput(ns("developers_sharing_fhir_endpoints_count"), inline = TRUE),
                      uiOutput(ns("shared_endpoints_denom"), inline = TRUE)
                    ),
                    div(style = "margin-top: 8px; font-size: 0.82em; color: #7f8c8d;",
                      "Developers whose FHIR endpoint sets are identical to another developer's",
                      tags$br(),
                      tags$span(style = "color: #1B5A7F; font-size: 0.9em; font-style: italic;",
                        tags$i(class = "fa fa-filter", style = "margin-right: 3px;"),
                        "Click to filter table below. Click again to reset."
                      )
                    )
                  )
                ),
                # Clickable card 4 — Developers w/ No Org Data
                div(style = "flex: 1; min-width: 0;",
                  div(class = "metric-card card-clickable",
                      id = ns("no_org_data_card"),
                      onclick = sprintf("Shiny.setInputValue('%s', Math.random());", ns("no_org_data_card_click")),
                    div(class = "metric-title",
                      tags$i(class = "fa fa-exclamation-triangle", style = "margin-right: 5px;"),
                      "Developers w/ No Org Data"
                    ),
                    div(class = "metric-value", style = "color: #dc3545;",
                      textOutput(ns("developers_no_org_data_count"), inline = TRUE),
                      uiOutput(ns("no_org_data_denom"), inline = TRUE)
                    ),
                    div(style = "margin-top: 8px; font-size: 0.82em; color: #7f8c8d;",
                      "Developers missing org data for any endpoint",
                      tags$br(),
                      tags$span(style = "color: #1B5A7F; font-size: 0.9em; font-style: italic;",
                        tags$i(class = "fa fa-filter", style = "margin-right: 3px;"),
                        "Click to filter table below. Click again to reset."
                      )
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
                fluidRow(
                  column(width = 3,
                    selectInput(
                      inputId = ns("source_filter"),
                      label = "Source:",
                      choices = c("CHPL Developers", "Others"),
                      selected = "CHPL Developers"
                    )
                  ),
                  column(width = 2, style = "padding-top: 25px;",
                    actionButton(
                      inputId = ns("reset_filter_btn"),
                      label = "Show All",
                      icon = icon("times-circle"),
                      class = "btn btn-default btn-sm"
                    )
                  ),
                  column(width = 3, style = "padding-top: 25px;",
                    downloadButton(
                      outputId = ns("download_highlighted_report"),
                      label = "Download Developers with Issues (CSV)",
                      class = "btn btn-warning btn-sm",
                      icon = icon("download")
                    )
                  ),
                  column(width = 3, style = "padding-top: 25px;",
                    downloadButton(
                      outputId = ns("download_tier1_report"),
                      label = "Download All (CSV)",
                      class = "btn btn-info btn-sm",
                      icon = icon("download")
                    )
                  )
                ),
                reactable::reactableOutput(ns("developer_data_issues_table"))
              )
            )
          )
        )
      ),

      # ── TAB 2: Organization Data Quality ─────────────────────────────────
      tabPanel(
        title = "Organization Data Quality",
        value = "tier2",

        fluidRow(style = "margin-top: 20px;",
          column(width = 12,
            div(style = "background: linear-gradient(135deg, #f8f9fa 0%, #ffffff 100%);
                         padding: 15px; border-radius: 8px; margin-bottom: 20px;
                         border-left: 4px solid #1B5A7F;",
              p(style = "margin: 0; color: #5a6c7d; line-height: 1.6;",
                tags$strong("About this tab:"),
                " This tab provides comprehensive data quality metrics for organization data extracted",
                " from FHIR bundles. Use this information to improve the quality of organization data",
                " in your endpoint implementations."
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

            # Detailed Issues
            div(class = "modern-card", style = "margin-top: 20px;",
              h3(class = "section-header",
                 tags$i(class = "fa fa-exclamation-circle", style = "margin-right: 8px;"),
                 "Data Quality Issues by Category"),
              reactable::reactableOutput(ns("issues_detail_table"))
            ),

            # Identifier Type Analysis
            div(class = "modern-card",
              h3(class = "section-header",
                 tags$i(class = "fa fa-id-card", style = "margin-right: 8px;"),
                 "Organization Identifier Analysis"),
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
                h4(class = "subsection-header", "Unique Identifier Values by Type"),
                p(style = "font-size: 13px; color: #5a6c7d; margin-bottom: 8px;",
                  "Counts reflect globally distinct identifier values across all organizations. ",
                  "The same NPI or CLIA appearing in multiple organizations is counted once."),
                reactable::reactableOutput(ns("identifier_type_table"))
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

            # Recommendations
            div(class = "modern-card",
              h4(style = "color: #1B5A7F; margin-top: 0;",
                 tags$i(class = "fa fa-lightbulb", style = "margin-right: 8px;"),
                 "Recommendations"),
              uiOutput(ns("recommendations"))
            )
          )
        ),

        # Tier 2 download
        fluidRow(
          column(width = 12, style = "padding-top: 20px; text-align: center;",
            downloadButton(
              outputId = ns("download_feedback_report"),
              label = "Download Organization Quality Report (CSV)",
              class = "btn-download",
              icon = icon("download")
            )
          )
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

  # Reactive value to track active card filter: NULL = no filter, "shares_list_source" or "has_empty_bundle"
  table_filter <- reactiveVal(NULL)

  # Initialize vendor choices as soon as the vendor list is available.
  # Do NOT gate on input$main_tabs == "tier2": with choices = NULL, Shiny sets
  # input$vendor_filter = "" (not NULL), so the query runs with vendor_name = ''
  # and returns 0 rows before the tab is ever visited.
  observe({
    req(app$vendor_list())
    # app$vendor_list() already contains "All Developers" as its first entry
    vendor_choices <- app$vendor_list()
    updateSelectInput(session, "vendor_filter", choices = vendor_choices, selected = "All Developers")
  })

  # Handle click on Shared FHIR Bundle Hyperlinks card — toggle filter
  observeEvent(input$shared_sources_card_click, {
    if (isTRUE(input$source_filter == "Others")) return()
    if (identical(table_filter(), "shares_list_source")) {
      table_filter(NULL)  # toggle off
      session$sendCustomMessage("toggleCardActive", list(cardId = ns("shared_sources_card"), active = FALSE))
    } else {
      table_filter("shares_list_source")
      session$sendCustomMessage("toggleCardActive", list(cardId = ns("shared_sources_card"), active = TRUE))
      # Deactivate the other clickable cards
      session$sendCustomMessage("toggleCardActive", list(cardId = ns("empty_bundles_card"), active = FALSE))
      session$sendCustomMessage("toggleCardActive", list(cardId = ns("shared_endpoints_card"), active = FALSE))
      session$sendCustomMessage("toggleCardActive", list(cardId = ns("no_org_data_card"), active = FALSE))
    }
  })

  # Handle click on Empty FHIR Bundles card — toggle filter
  observeEvent(input$empty_bundles_card_click, {
    if (isTRUE(input$source_filter == "Others")) return()
    if (identical(table_filter(), "has_empty_bundle")) {
      table_filter(NULL)  # toggle off
      session$sendCustomMessage("toggleCardActive", list(cardId = ns("empty_bundles_card"), active = FALSE))
    } else {
      table_filter("has_empty_bundle")
      session$sendCustomMessage("toggleCardActive", list(cardId = ns("empty_bundles_card"), active = TRUE))
      # Deactivate the other clickable cards
      session$sendCustomMessage("toggleCardActive", list(cardId = ns("shared_sources_card"), active = FALSE))
      session$sendCustomMessage("toggleCardActive", list(cardId = ns("shared_endpoints_card"), active = FALSE))
      session$sendCustomMessage("toggleCardActive", list(cardId = ns("no_org_data_card"), active = FALSE))
    }
  })

  # Handle click on Developers Sharing FHIR Endpoints card — toggle filter
  observeEvent(input$shared_endpoints_card_click, {
    if (isTRUE(input$source_filter == "Others")) return()
    if (identical(table_filter(), "shares_fhir_endpoints")) {
      table_filter(NULL)  # toggle off
      session$sendCustomMessage("toggleCardActive", list(cardId = ns("shared_endpoints_card"), active = FALSE))
    } else {
      table_filter("shares_fhir_endpoints")
      session$sendCustomMessage("toggleCardActive", list(cardId = ns("shared_endpoints_card"), active = TRUE))
      # Deactivate the other clickable cards
      session$sendCustomMessage("toggleCardActive", list(cardId = ns("shared_sources_card"), active = FALSE))
      session$sendCustomMessage("toggleCardActive", list(cardId = ns("empty_bundles_card"), active = FALSE))
      session$sendCustomMessage("toggleCardActive", list(cardId = ns("no_org_data_card"), active = FALSE))
    }
  })

  # Handle click on Developers w/ No Org Data card — toggle filter
  observeEvent(input$no_org_data_card_click, {
    if (identical(table_filter(), "no_org_data")) {
      table_filter(NULL)  # toggle off
      session$sendCustomMessage("toggleCardActive", list(cardId = ns("no_org_data_card"), active = FALSE))
    } else {
      table_filter("no_org_data")
      session$sendCustomMessage("toggleCardActive", list(cardId = ns("no_org_data_card"), active = TRUE))
      # Deactivate the other clickable cards
      session$sendCustomMessage("toggleCardActive", list(cardId = ns("shared_sources_card"), active = FALSE))
      session$sendCustomMessage("toggleCardActive", list(cardId = ns("empty_bundles_card"), active = FALSE))
      session$sendCustomMessage("toggleCardActive", list(cardId = ns("shared_endpoints_card"), active = FALSE))
    }
  })

  # Reset filter button — clears all card filters and resets source filter to "CHPL Developers"
  observeEvent(input$reset_filter_btn, {
    table_filter(NULL)
    updateSelectInput(session, "source_filter", selected = "CHPL Developers")
    session$sendCustomMessage("toggleCardActive", list(cardId = ns("shared_sources_card"), active = FALSE))
    session$sendCustomMessage("toggleCardActive", list(cardId = ns("empty_bundles_card"), active = FALSE))
    session$sendCustomMessage("toggleCardActive", list(cardId = ns("shared_endpoints_card"), active = FALSE))
    session$sendCustomMessage("toggleCardActive", list(cardId = ns("no_org_data_card"), active = FALSE))
  })

  # When "Others" is selected, disable the 3 CHPL-only cards (FHIR bundle data is CHPL-only).
  # When "CHPL Developers" is selected, re-enable them.
  observeEvent(input$source_filter, {
    is_others <- isTRUE(input$source_filter == "Others")
    chpl_only_ids <- c(ns("empty_bundles_card"), ns("shared_sources_card"), ns("shared_endpoints_card"))
    for (card_id in chpl_only_ids) {
      session$sendCustomMessage("toggleCardDisabled", list(cardId = card_id, disabled = is_others))
    }
    session$sendCustomMessage("toggleCardDisabled", list(cardId = ns("no_org_data_card"), disabled = FALSE))
    # If a CHPL-only card filter was active, clear it
    if (is_others && !is.null(table_filter()) &&
        table_filter() %in% c("has_empty_bundle", "shares_list_source", "shares_fhir_endpoints")) {
      table_filter(NULL)
      for (card_id in chpl_only_ids) {
        session$sendCustomMessage("toggleCardActive", list(cardId = card_id, active = FALSE))
      }
    }
  }, ignoreInit = FALSE)
  
  # Get filtered organization data from materialized views
  filtered_quality_summary <- reactive({
    current_vendor <- input$vendor_filter
    if (is.null(current_vendor) || current_vendor == "") current_vendor <- "All Developers"
    
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
    if (is.null(current_vendor) || current_vendor == "") current_vendor <- "All Developers"
    
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
    if (is.null(current_vendor) || current_vendor == "") current_vendor <- "All Developers"
    
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
        npi_count = 0, clia_count = 0, naic_count = 0, ccn_count = 0, other_count = 0, no_identifier_count = 0,
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
      ccn_count = 0,  # CCN not yet in Lantern DB; placeholder for display
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
  # Timestamp of when CHPL data was last fetched (MAX updated_at from shared_list_sources)
  chpl_last_updated <- reactive({
    result <- tbl(db_connection, sql(
      "SELECT MAX(updated_at) AS last_updated FROM shared_list_sources"
    )) %>% collect()
    if (nrow(result) == 0 || is.na(result$last_updated[1])) return("Unknown")
    format(as.POSIXct(result$last_updated[1]), "%B %d, %Y at %I:%M %p %Z")
  })

  output$chpl_last_updated <- renderText({ chpl_last_updated() })

  # CHPL vs Lantern coverage counts for the static card
  chpl_lantern_counts <- reactive({
    tbl(db_connection, sql(
      "SELECT
        (SELECT COUNT(DISTINCT developer_name) FROM shared_list_sources) AS chpl_dev_count,
        (SELECT COUNT(DISTINCT v.name)
           FROM list_source_info lsi
           JOIN fhir_endpoints fe ON lsi.list_source = fe.list_source
           JOIN fhir_endpoints_info fei ON fe.url = fei.url AND fei.requested_fhir_version = 'None'
           JOIN vendors v ON fei.vendor_id = v.id
           WHERE lsi.is_chpl = 'CHPL' AND v.name IS NOT NULL) AS lantern_chpl_dev_count,
        (SELECT COUNT(DISTINCT list_source) FROM shared_list_sources) AS chpl_bundle_count,
        (SELECT COUNT(DISTINCT list_source) FROM list_source_info
           WHERE is_chpl = 'CHPL') AS lantern_chpl_bundle_count"
    )) %>% collect()
  })

  output$counts_static_card <- renderUI({
    counts <- chpl_lantern_counts()
    if (nrow(counts) == 0) return(NULL)
    div(class = "metric-card",
      style = "background: #f0f7ff; border: 2px solid #1B5A7F; height: 100%; box-sizing: border-box;",
      div(class = "metric-title",
        tags$i(class = "fa fa-info-circle", style = "margin-right: 5px; color: #1B5A7F;"),
        tags$strong(style = "color: #1B5A7F;", "Coverage Overview")
      ),
      tags$table(style = "width: 100%; font-size: 0.88em; margin-top: 8px; border-collapse: collapse;",
        tags$tr(
          tags$td(style = "color: #5a6c7d; padding: 4px 6px 4px 0;", "CHPL Dev Count:"),
          tags$td(style = "font-weight: 700; color: #1B5A7F; text-align: right; padding: 4px 0;",
            format(as.integer(counts$chpl_dev_count[1]), big.mark = ","))
        ),
        tags$tr(
          tags$td(style = "color: #5a6c7d; padding: 4px 6px 4px 0;", "Lantern Dev Count from CHPL:"),
          tags$td(style = "font-weight: 700; color: #2c3e50; text-align: right; padding: 4px 0;",
            format(as.integer(counts$lantern_chpl_dev_count[1]), big.mark = ","))
        ),
        tags$tr(
          tags$td(style = "color: #5a6c7d; padding: 4px 6px 4px 0;", "CHPL FHIR Bundle Hyperlinks:"),
          tags$td(style = "font-weight: 700; color: #1B5A7F; text-align: right; padding: 4px 0;",
            format(as.integer(counts$chpl_bundle_count[1]), big.mark = ","))
        ),
        tags$tr(
          tags$td(style = "color: #5a6c7d; padding: 4px 6px 4px 0;", "Lantern FHIR Bundle Hyperlinks from CHPL:"),
          tags$td(style = "font-weight: 700; color: #2c3e50; text-align: right; padding: 4px 0;",
            format(as.integer(counts$lantern_chpl_bundle_count[1]), big.mark = ","))
        )
      )
    )
  })

  # Developer data issues - comprehensive view (kept for org quality tab vendor filter)
  developer_data_issues <- reactive({
    query_str <- "SELECT * FROM mv_developer_data_issues ORDER BY
                  no_org_data_endpoints DESC,
                  vendor_name"
    result <- tbl(db_connection, sql(query_str)) %>% collect()
    return(result)
  })

  # Bundle-level data issues — one row per (developer_name, list_source) pair
  # All counts (endpoints, orgs) are per bundle URL, not per developer
  developer_bundle_issues <- reactive({
    query_str <- "SELECT * FROM mv_developer_bundle_issues ORDER BY developer_name, list_source"
    tbl(db_connection, sql(query_str)) %>% collect()
  })

  # Combined data: CHPL rows from mv_developer_bundle_issues (per bundle URL) +
  # non-CHPL rows from mv_developer_data_issues (per vendor, is_chpl_developer = FALSE)
  # This preserves the "Others" filter by keeping non-CHPL developers visible.
  all_data_issues <- reactive({
    chpl_rows <- developer_bundle_issues()

    non_chpl_rows <- developer_data_issues() %>%
      filter(is_chpl_developer == FALSE) %>%
      transmute(
        developer_name      = vendor_name,
        list_source         = NA_character_,
        total_endpoints     = total_endpoints,
        endpoints_with_org_data = endpoints_with_org_data,
        no_org_data_endpoints   = no_org_data_endpoints,
        organization_count  = organization_count,
        has_empty_bundle    = has_empty_bundle,
        shares_list_source  = shares_list_source,
        shares_fhir_endpoints = shares_fhir_endpoints,
        is_chpl_developer   = FALSE
      )

    bind_rows(chpl_rows, non_chpl_rows)
  })

  # Filtered card counts — sourced from combined CHPL+non-CHPL data, deduplicated to developer level
  # Cards always show developer counts (not bundle URL counts)
  filtered_data_issues_counts <- reactive({
    dev_data <- all_data_issues()
    source_filter_val <- input$source_filter

    if (!is.null(source_filter_val) && source_filter_val == "CHPL Developers") {
      dev_data <- dev_data[dev_data$is_chpl_developer == TRUE, ]
    } else if (!is.null(source_filter_val) && source_filter_val == "Others") {
      dev_data <- dev_data[dev_data$is_chpl_developer == FALSE, ]
    }

    # Deduplicate to developer level for card counts
    unique_devs <- dev_data[!duplicated(dev_data$developer_name), ]

    list(
      developers_with_no_org_data_count       = sum(unique_devs$no_org_data_endpoints > 0, na.rm = TRUE),
      endpoints_with_no_org_data_count        = sum(dev_data$no_org_data_endpoints, na.rm = TRUE),
      developers_sharing_list_sources_count   = sum(unique_devs$shares_list_source == TRUE, na.rm = TRUE),
      inaccessible_list_sources_count         = 0L,
      developers_with_empty_bundles_count     = sum(unique_devs$has_empty_bundle == TRUE, na.rm = TRUE),
      developers_sharing_fhir_endpoints_count = sum(unique_devs$shares_fhir_endpoints == TRUE, na.rm = TRUE),
      total_developers_count                  = nrow(unique_devs)
    )
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
      Category = c("Identifier", "Organization Name", "Address"),
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
      Type = c("NPI", "CLIA", "NAIC", "CCN", "Other", "No Data"),
      Count = c(
        as.numeric(id_summary$npi_count),
        as.numeric(id_summary$clia_count),
        as.numeric(id_summary$naic_count),
        0,  # CCN placeholder — not yet in Lantern DB
        as.numeric(id_summary$other_count),
        as.numeric(id_summary$no_identifier_count)
      ),
      stringsAsFactors = FALSE
    )

    # Only filter out zero-count rows for types OTHER than CCN (CCN always shown as placeholder)
    chart_data <- chart_data[chart_data$Count > 0 | chart_data$Type == "CCN", ]

    if (nrow(chart_data[chart_data$Type != "CCN", ]) == 0 && chart_data[chart_data$Type == "CCN", "Count"] == 0) {
      # All types (including non-CCN) are zero — show empty state
      non_ccn <- chart_data[chart_data$Type != "CCN", ]
      if (nrow(non_ccn) == 0) {
        return(
          ggplot() +
            geom_text(aes(x = 0.5, y = 0.5, label = "No identifier data found"),
                     size = 6, color = "#7f8c8d") +
            theme_void() + xlim(0, 1) + ylim(0, 1)
        )
      }
    }

    type_colors <- c("NPI" = "#28a745", "CLIA" = "#007bff", "NAIC" = "#fd7e14",
                    "CCN" = "#17a2b8", "Other" = "#9b59b6", "No Data" = "#6c757d")
    
    ggplot(chart_data, aes(x = reorder(Type, Count), y = Count, fill = Type)) +
      geom_col(width = 0.6) +
      geom_text(aes(label = format(Count, big.mark = ",")),
                hjust = -0.1, fontface = "bold", size = 3.5) +
      scale_fill_manual(values = type_colors) +
      coord_flip() +
      labs(x = NULL, y = "Number of Distinct Identifier Values",
           title = "Distinct Identifier Values by Type") +
      theme_minimal() +
      theme(
        legend.position = "none",
        plot.title = element_text(size = 12, face = "bold", margin = margin(b = 8)),
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
      labs(x = "Identifier Type", y = "Number of Distinct Identifier Values",
           title = "Valid vs. Invalid Identifier Values (NPI, CLIA, NAIC)") +
      theme_minimal() +
      theme(
        plot.title = element_text(size = 12, face = "bold", margin = margin(b = 8)),
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
      Identifier_Type = c("NPI", "CLIA", "NAIC", "CCN", "Other"),
      Total_Count = c(
        as.numeric(id_summary$npi_count),
        as.numeric(id_summary$clia_count),
        as.numeric(id_summary$naic_count),
        0,  # CCN placeholder — not yet in Lantern DB
        as.numeric(id_summary$other_count)
      ),
      Valid_Count = c(
        as.numeric(id_summary$npi_valid),
        as.numeric(id_summary$clia_valid),
        as.numeric(id_summary$naic_valid),
        0,
        as.numeric(id_summary$other_count)  # all "other" types are now valid per 89 FR 1288
      ),
      Invalid_Count = c(
        as.numeric(id_summary$npi_invalid),
        as.numeric(id_summary$clia_invalid),
        as.numeric(id_summary$naic_invalid),
        0,
        0  # "other" types no longer counted as invalid
      ),
      Conformance_Rate = c(
        if(id_summary$npi_count > 0) paste0(round(id_summary$npi_valid / id_summary$npi_count * 100, 1), "%") else "N/A",
        if(id_summary$clia_count > 0) paste0(round(id_summary$clia_valid / id_summary$clia_count * 100, 1), "%") else "N/A",
        if(id_summary$naic_count > 0) paste0(round(id_summary$naic_valid / id_summary$naic_count * 100, 1), "%") else "N/A",
        "N/A",
        if(id_summary$other_count > 0) "100%" else "N/A"
      ),
      Percentage_of_Orgs = c(
        paste0(id_summary$npi_percentage, "%"),
        paste0(id_summary$clia_percentage, "%"),
        paste0(id_summary$naic_percentage, "%"),
        "0%",
        paste0(id_summary$other_percentage, "%")
      ),
      US_Core_Rules = c("us-core-16, us-core-17", "us-core-18", "us-core-19", "us-core-20", "89 FR 1288 (Other)"),
      Validation_Requirements = c(
        "10 digits + Luhn check digit",
        "2 digits + 'D' + 7 digits",
        "5 digits",
        "6 digits (Medicare Certification Number)",
        "Any non-empty value accepted (other health system IDs)"
      ),
      stringsAsFactors = FALSE
    )
    
    reactable(
      type_data,
      columns = list(
        Identifier_Type = colDef(name = "Identifier Type", width = 140,
                                 style = list(fontWeight = 600)),
        Total_Count = colDef(name = "Unique Values", format = colFormat(separators = TRUE), width = 110),
        Valid_Count = colDef(name = "Valid Values", format = colFormat(separators = TRUE), width = 100,
                            style = function(value) {
                              if (value > 0) list(color = "#28a745", fontWeight = 600)
                            }),
        Invalid_Count = colDef(name = "Invalid Values", format = colFormat(separators = TRUE), width = 110,
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
        Percentage_of_Orgs = colDef(name = "% of All Identifiers", width = 140),
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
      Issue_Category = c("Identifier", "Organization Names", "Address"),
      Common_Issues = c(
        paste0("Missing identifier data (", format(id_summary$no_identifier_count, big.mark = ","), "), ",
               "invalid NPI check digits (", format(id_summary$npi_invalid, big.mark = ","), "), ",
               "incorrect CLIA format (", format(id_summary$clia_invalid, big.mark = ","), "), ",
               "wrong NAIC length (", format(id_summary$naic_invalid, big.mark = ","), "). ",
               "Note: other health system IDs (", format(id_summary$other_count, big.mark = ","), ") are accepted per 89 FR 1288."),
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
  
  # Data Issues outputs — sourced from filtered_data_issues_counts() so cards
  # reflect the currently selected Source filter
  output$developers_no_org_data_count <- renderText({
    format(filtered_data_issues_counts()$developers_with_no_org_data_count, big.mark = ",")
  })

  output$endpoints_no_org_data_count <- renderText({
    format(filtered_data_issues_counts()$endpoints_with_no_org_data_count, big.mark = ",")
  })

  output$developers_sharing_list_sources_count <- renderText({
    format(filtered_data_issues_counts()$developers_sharing_list_sources_count, big.mark = ",")
  })

  output$inaccessible_list_sources_count <- renderText({
    format(filtered_data_issues_counts()$inaccessible_list_sources_count, big.mark = ",")
  })

  output$developers_empty_bundles_count <- renderText({
    format(filtered_data_issues_counts()$developers_with_empty_bundles_count, big.mark = ",")
  })

  output$developers_sharing_fhir_endpoints_count <- renderText({
    format(filtered_data_issues_counts()$developers_sharing_fhir_endpoints_count, big.mark = ",")
  })

  # Denominator displays for each card (unique output IDs to avoid duplicate binding)
  denom_span <- function(n) {
    span(style = "font-size: 0.7em; color: #7f8c8d; font-weight: 400;",
         paste0(" / ", format(n, big.mark = ",")))
  }
  output$empty_bundles_denom <- renderUI({
    denom_span(filtered_data_issues_counts()$total_developers_count)
  })
  output$shared_sources_denom <- renderUI({
    denom_span(filtered_data_issues_counts()$total_developers_count)
  })
  output$shared_endpoints_denom <- renderUI({
    denom_span(filtered_data_issues_counts()$total_developers_count)
  })
  output$no_org_data_denom <- renderUI({
    denom_span(filtered_data_issues_counts()$total_developers_count)
  })

  # Comprehensive developer data issues table
  output$developer_data_issues_table <- reactable::renderReactable({
    req(all_data_issues())

    dev_data <- all_data_issues()
    active_filter <- table_filter()
    source_filter_val <- input$source_filter

    # Apply card filter (show only bundle rows matching the clicked card)
    if (!is.null(active_filter)) {
      if (active_filter == "no_org_data") {
        dev_data <- dev_data[dev_data$no_org_data_endpoints > 0, ]
      } else {
        dev_data <- dev_data[dev_data[[active_filter]] == TRUE, ]
      }
    }

    # Apply source filter
    if (!is.null(source_filter_val) && source_filter_val == "CHPL Developers") {
      dev_data <- dev_data[dev_data$is_chpl_developer == TRUE, ]
    } else if (!is.null(source_filter_val) && source_filter_val == "Others") {
      dev_data <- dev_data[dev_data$is_chpl_developer == FALSE, ]
    }

    # Compute compliant column: non-compliant if has_empty_bundle, shares_list_source, or shares_fhir_endpoints
    dev_data$compliant <- !(dev_data$has_empty_bundle | dev_data$shares_list_source | dev_data$shares_fhir_endpoints)

    if (nrow(dev_data) == 0) {
      dev_data <- data.frame(
        developer_name = "No data found",
        list_source = "",
        total_endpoints = 0,
        endpoints_with_org_data = 0,
        no_org_data_endpoints = 0,
        organization_count = 0,
        compliant = TRUE,
        has_empty_bundle = FALSE,
        shares_list_source = FALSE,
        shares_fhir_endpoints = FALSE,
        is_chpl_developer = FALSE,
        stringsAsFactors = FALSE
      )
    }

    reactable(
      dev_data,
      filterable = TRUE,
      searchable = TRUE,
      defaultPageSize = 20,
      defaultSorted = list(developer_name = "asc"),
      columns = list(
        developer_name = colDef(
          name = "Developer Name",
          minWidth = 180,
          style = list(fontWeight = 600, color = "#2c3e50")
        ),
        list_source = colDef(
          name = "FHIR Bundle URL",
          minWidth = 220,
          cell = function(value) {
            if (is.na(value) || value == "") return(value)
            tags$a(
              href = value, target = "_blank",
              style = "color: #1B5A7F; text-decoration: none; word-break: break-all;",
              value
            )
          }
        ),
        total_endpoints = colDef(
          name = "Total Endpoints",
          width = 130,
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
          width = 110,
          format = colFormat(separators = TRUE),
          align = "center",
          style = function(value) {
            if (value > 0) list(color = "#dc3545", fontWeight = 700)
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
        compliant = colDef(
          name = "Compliant",
          width = 110,
          align = "center",
          cell = function(value) {
            if (isTRUE(value)) {
              tags$span(
                style = "color: #28a745; font-weight: 700;",
                tags$i(class = "fa fa-check-circle", style = "margin-right: 5px;"),
                "Yes"
              )
            } else {
              tags$span(
                style = "color: #dc3545; font-weight: 700;",
                tags$i(class = "fa fa-times-circle", style = "margin-right: 5px;"),
                "No"
              )
            }
          }
        ),
        has_empty_bundle = colDef(
          name = "Empty Bundle",
          width = 120,
          align = "center",
          cell = function(value) {
            if (isTRUE(value)) {
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
        ),
        shares_list_source = colDef(
          name = "Shares FHIR Bundle URL",
          width = 160,
          align = "center",
          cell = function(value) {
            if (isTRUE(value)) {
              tags$span(
                style = "color: #ffc107; font-weight: 700;",
                tags$i(class = "fa fa-share-alt", style = "margin-right: 5px;"),
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
        ),
        shares_fhir_endpoints = colDef(
          name = "Shares FHIR Endpoints",
          width = 170,
          align = "center",
          cell = function(value) {
            if (isTRUE(value)) {
              tags$span(
                style = "color: #ffc107; font-weight: 700;",
                tags$i(class = "fa fa-code-fork", style = "margin-right: 5px;"),
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
        ),
        is_chpl_developer = colDef(show = FALSE)
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
      recommendations <- c(recommendations, list(
        tags$div(class = "alert alert-danger",
          tags$strong(tags$i(class = "fa fa-times-circle", style = "margin-right: 5px;"),
                     "Missing Identifier Data: "),
          paste0(format(id_summary$no_identifier_count, big.mark = ","),
                 " organizations (", no_id_percentage, "%) have no identifier data."),
          tags$br(),
          tags$small("Organizations must include at least one identifier to meet US-Core requirements.",
                     " Per 89 FR 1288, NPI, CLIA, CCN, or other health system IDs are all acceptable.")
        )
      ))
    }

    # Invalid only identifiers alert (only NPI/CLIA/NAIC format failures; other types are now valid)
    if (id_summary$orgs_with_invalid_only > 0) {
      invalid_only_percentage <- round(id_summary$orgs_with_invalid_only / summary$total_orgs * 100, 1)
      recommendations <- c(recommendations, list(
        tags$div(class = "alert alert-danger",
          tags$strong(tags$i(class = "fa fa-exclamation-triangle", style = "margin-right: 5px;"),
                     "Organizations with Only Invalid Identifiers: "),
          paste0(format(id_summary$orgs_with_invalid_only, big.mark = ","),
                 " organizations (", invalid_only_percentage, "%) have NPI/CLIA/NAIC identifiers that fail format validation."),
          tags$br(),
          tags$small("Review NPI (10-digit + Luhn), CLIA (2D7 format), and NAIC (5-digit) identifier formats.")
        )
      ))
    }

    # Identifier conformance recommendations
    if (summary$identifier_percentage < 80) {
      recommendations <- c(recommendations, list(
        tags$div(class = "alert alert-warning",
          tags$strong(tags$i(class = "fa fa-clipboard-check", style = "margin-right: 5px;"),
                     "US-Core Identifier Conformance Issues: "),
          paste0("Only ", summary$identifier_percentage, "% of organizations have conformant identifiers."),
          tags$br(),
          tags$small("Ensure NPI identifiers are 10 digits with valid check digits, CLIA identifiers follow 2D7 format, and NAIC identifiers are 5 digits.")
        )
      ))
    }

    # Specific validation error recommendations
    if (id_summary$npi_invalid > 0) {
      recommendations <- c(recommendations, list(
        tags$div(class = "alert alert-warning",
          tags$strong(tags$i(class = "fa fa-id-badge", style = "margin-right: 5px;"),
                     "Invalid NPI Identifiers: "),
          paste0(format(id_summary$npi_invalid, big.mark = ","), " NPIs failed validation (us-core-16/17)."),
          tags$br(),
          tags$small("Verify NPIs are exactly 10 digits and have valid Luhn check digits.")
        )
      ))
    }

    if (id_summary$clia_invalid > 0) {
      recommendations <- c(recommendations, list(
        tags$div(class = "alert alert-warning",
          tags$strong(tags$i(class = "fa fa-flask", style = "margin-right: 5px;"),
                     "Invalid CLIA Identifiers: "),
          paste0(format(id_summary$clia_invalid, big.mark = ","), " CLIAs failed validation (us-core-18)."),
          tags$br(),
          tags$small("CLIA format must be: 2 digits + 'D' + 7 digits (e.g., '12D3456789').")
        )
      ))
    }

    if (id_summary$naic_invalid > 0) {
      recommendations <- c(recommendations, list(
        tags$div(class = "alert alert-warning",
          tags$strong(tags$i(class = "fa fa-shield-alt", style = "margin-right: 5px;"),
                     "Invalid NAIC Identifiers: "),
          paste0(format(id_summary$naic_invalid, big.mark = ","), " NAICs failed validation (us-core-19)."),
          tags$br(),
          tags$small("NAIC identifiers must be exactly 5 digits.")
        )
      ))
    }

    if (summary$name_percentage < 80) {
      recommendations <- c(recommendations, list(
        tags$div(class = "alert alert-info",
          tags$strong(tags$i(class = "fa fa-building", style = "margin-right: 5px;"),
                     "Name Quality: "),
          "Use complete, meaningful organization names instead of placeholders."
        )
      ))
    }

    if (summary$address_percentage < 80) {
      recommendations <- c(recommendations, list(
        tags$div(class = "alert alert-secondary",
          tags$strong(tags$i(class = "fa fa-map-marker-alt", style = "margin-right: 5px;"),
                     "Address Issues: "),
          "Include complete addresses with street, city, state, and ZIP code."
        )
      ))
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
  
  # Helper: apply source filter to developer data
  apply_source_filter <- function(data) {
    source_filter_val <- input$source_filter
    if (!is.null(source_filter_val) && source_filter_val == "CHPL Developers") {
      data <- data[data$is_chpl_developer == TRUE, ]
    } else if (!is.null(source_filter_val) && source_filter_val == "Others") {
      data <- data[data$is_chpl_developer == FALSE, ]
    }
    data
  }

  # Helper: select and rename columns for CSV export
  format_for_csv <- function(data) {
    data %>%
      select(
        developer_name,
        list_source,
        total_endpoints,
        endpoints_with_org_data,
        no_org_data_endpoints,
        organization_count,
        compliant,
        has_empty_bundle,
        shares_list_source,
        shares_fhir_endpoints
      )
  }

  # Helper: add compliant column to raw developer data before CSV export
  add_compliant_col <- function(data) {
    data$compliant <- !(data$has_empty_bundle | data$shares_list_source | data$shares_fhir_endpoints)
    data
  }

  # Tier 1 download handler: highlighted developers (empty bundles OR sharing FHIR bundle URL)
  output$download_highlighted_report <- downloadHandler(
    filename = function() {
      paste0("highlighted_developers_", Sys.Date(), ".csv")
    },
    content = function(file) {
      data <- add_compliant_col(apply_source_filter(all_data_issues()))
      data <- data[data$has_empty_bundle == TRUE | data$shares_list_source == TRUE, ]
      if (nrow(data) > 0) {
        write.csv(format_for_csv(data), file, row.names = FALSE)
      } else {
        write.csv(data.frame(message = "No highlighted developers found"), file, row.names = FALSE)
      }
    }
  )

  # Tier 1 download handler: all developers
  output$download_tier1_report <- downloadHandler(
    filename = function() {
      paste0("chpl_developer_service_base_url_report_", Sys.Date(), ".csv")
    },
    content = function(file) {
      data <- add_compliant_col(apply_source_filter(all_data_issues()))
      if (nrow(data) > 0) {
        write.csv(format_for_csv(data), file, row.names = FALSE)
      } else {
        write.csv(data.frame(
          developer_name = character(0),
          list_source = character(0),
          total_endpoints = integer(0),
          endpoints_with_org_data = integer(0),
          no_org_data_endpoints = integer(0),
          organization_count = integer(0),
          compliant = logical(0),
          has_empty_bundle = logical(0),
          shares_list_source = logical(0),
          shares_fhir_endpoints = logical(0)
        ), file, row.names = FALSE)
      }
    }
  )

  # Tier 2 (Organization) download handler
  output$download_feedback_report <- downloadHandler(
    filename = function() {
      paste0("service_base_url_data_quality_report_", Sys.Date(), ".csv")
    },
    content = function(file) {
      data <- filtered_org_data()

      if (nrow(data) > 0) {
        report_data <- data %>%
          mutate(
            developer_names = sapply(vendor_names_array, function(x) {
              if (is.null(x) || length(x) == 0) return("Unknown")
              paste(x, collapse = "; ")
            }),
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
            developer_names,
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