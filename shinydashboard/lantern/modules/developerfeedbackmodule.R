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
    fluidRow(
      h2("Organization Data Quality")
    ),
    fluidRow(
      column(width = 12,
        p("This dashboard provides data quality metrics for organization data extracted from FHIR bundles. ",
          "Use this information to improve the quality of organization data in your endpoint implementations.")
      )
    ),
    # Summary cards row
    fluidRow(
      column(width = 4,
        div(class = "info-box bg-blue",
          div(class = "info-box-icon",
            tags$i(class = "fa fa-building")
          ),
          div(class = "info-box-content",
            span(class = "info-box-text", "Total Organizations"),
            span(class = "info-box-number", textOutput(ns("total_orgs"), inline = TRUE))
          )
        )
      ),
      column(width = 4,
        div(class = "info-box bg-green",
          div(class = "info-box-icon",
            tags$i(class = "fa fa-check-circle")
          ),
          div(class = "info-box-content",
            span(class = "info-box-text", "Conforming Organization Data"),
            span(class = "info-box-number", textOutput(ns("high_quality_count"), inline = TRUE))
          )
        )
      ),
      column(width = 4,
        div(class = "info-box bg-red",
          div(class = "info-box-icon",
            tags$i(class = "fa fa-exclamation-triangle")
          ),
          div(class = "info-box-content",
            span(class = "info-box-text", "Non-conforming Organization Data"),
            span(class = "info-box-number", textOutput(ns("low_quality_count"), inline = TRUE))
          )
        )
      )
    ),
    # Main content row
    fluidRow(
      # Left column - Charts
      column(width = 8,
        tabsetPanel(
          tabPanel("Overview Charts",
            fluidRow(
              column(width = 12,
                h3("Data Quality Overview"),
                plotOutput(ns("quality_overview_chart"), height = "400px")
              )
            ),
            fluidRow(
              column(width = 6,
                h4("Identifier Type Validation"),
                plotOutput(ns("identifier_chart"), height = "300px")
              ),
              column(width = 6,
                h4("Organization Name Quality"),
                plotOutput(ns("name_chart"), height = "300px")
              )
            ),
            fluidRow(
              column(width = 12,
                h4("Address Completeness"),
                plotOutput(ns("address_chart"), height = "300px")
              )
            )
          ),
          tabPanel("Identifier Type Analysis",
            fluidRow(
              column(width = 6,
                h3("Identifier Type Distribution"),
                plotOutput(ns("identifier_type_distribution_chart"), height = "400px")
              ),
              column(width = 6,
                h3("Conformance by Type"),
                plotOutput(ns("conformance_by_type_chart"), height = "400px")
              )
            ),
            fluidRow(
              column(width = 12,
                h4("Organization Identifier Status Breakdown"),
                plotOutput(ns("organization_identifier_status_chart"), height = "300px")
              )
            ),
            fluidRow(
              column(width = 12,
                h4("Identifier Type Details"),
                reactable::reactableOutput(ns("identifier_type_table"))
              )
            )
          ),
          tabPanel("Detailed Issues",
            fluidRow(
              column(width = 12,
                h3("Data Quality Issues by Category"),
                reactable::reactableOutput(ns("issues_detail_table"))
              )
            )
          )
        )
      ),
      # Right column - Filters and Summary
      column(width = 4,
        wellPanel(
          h4("Filters"),
          selectInput(ns("vendor_filter"), 
                     "Certified API Developer:", 
                     choices = NULL,
                     selected = "All Developers"),
          hr(),
          h4("Quality Metrics"),
          div(id = "quality-metrics",
            h5("Identifier Type Validation"),
            div(class = "progress-group",
              span(class = "progress-text", "Valid Identifiers"),
              span(class = "float-right", textOutput(ns("identifier_percentage"), inline = TRUE)),
              div(class = "progress progress-sm",
                div(class = "progress-bar bg-green", 
                    style = paste0("width: ", textOutput(ns("identifier_progress_width"), inline = TRUE)))
              )
            ),
            h5("Organization Names"),
            div(class = "progress-group",
              span(class = "progress-text", "Quality Names"),
              span(class = "float-right", textOutput(ns("name_percentage"), inline = TRUE)),
              div(class = "progress progress-sm",
                div(class = "progress-bar bg-blue", 
                    style = paste0("width: ", textOutput(ns("name_progress_width"), inline = TRUE)))
              )
            ),
            h5("Addresses"),
            div(class = "progress-group",
              span(class = "progress-text", "Complete Addresses"),
              span(class = "float-right", textOutput(ns("address_percentage"), inline = TRUE)),
              div(class = "progress progress-sm",
                div(class = "progress-bar bg-orange", 
                    style = paste0("width: ", textOutput(ns("address_progress_width"), inline = TRUE)))
              )
            )
          )
        ),
        wellPanel(
          h4("Identifier Breakdown"),
          div(id = "identifier-breakdown",
            div(class = "info-line",
              span("Organizations with valid identifiers: "),
              textOutput(ns("valid_identifier_count_display"), inline = TRUE)
            ),
            div(class = "info-line",
              span("Organizations with no identifier data: "),
              textOutput(ns("no_identifier_count_display"), inline = TRUE)
            ),
            div(class = "info-line",
              span("Organizations with only invalid identifiers: "),
              textOutput(ns("invalid_only_count_display"), inline = TRUE)
            )
          )
        ),
        wellPanel(
          h4("Recommendations"),
          uiOutput(ns("recommendations"))
        )
      )
    ),
    # Download section
    fluidRow(
      column(width = 12, style = "padding-top: 20px;",
        downloadButton(ns("download_feedback_report"), "Download Quality Report (CSV)", 
                      icon = tags$i(class = "fa fa-download"))
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
  
  # Chart outputs using pre-computed data - FIXED
  output$quality_overview_chart <- renderPlot({
    req(quality_summary())  # Ensure data is available
    
    summary <- quality_summary()
    
    # Create chart data with proper validation
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
    
    # Check if we have data
    if (sum(chart_data$Valid) == 0 && sum(chart_data$Invalid) == 0) {
      return(
        ggplot() + 
          geom_text(aes(x = 0.5, y = 0.5, label = "No data available"), size = 6) +
          xlim(0, 1) + ylim(0, 1) + theme_void() +
          labs(title = "Data Quality Overview")
      )
    }
    
    # Pivot data for visualization
    chart_data_long <- chart_data %>%
      pivot_longer(cols = c(Valid, Invalid), names_to = "Status", values_to = "Count")
    
    ggplot(chart_data_long, aes(x = Category, y = Count, fill = Status)) +
      geom_col(position = "dodge", width = 0.7) +
      geom_text(aes(label = format(Count, big.mark = ",")), 
                position = position_dodge(width = 0.7), vjust = -0.5) +
      scale_fill_manual(values = c("Valid" = "#28a745", "Invalid" = "#dc3545")) +
      labs(title = "Data Quality Overview",
           x = "Quality Category",
           y = "Number of Organizations") +
      theme_minimal() +
      theme(axis.text.x = element_text(angle = 45, hjust = 1),
            legend.position = "bottom")
  }, height = 400)
  
  # Individual charts
  output$identifier_chart <- renderPlot({
    req(quality_summary())
    
    summary <- quality_summary()
    
    # Ensure we have numeric data
    valid_count <- as.numeric(summary$valid_identifier_count)
    total_count <- as.numeric(summary$total_orgs)
    invalid_count <- total_count - valid_count
    
    if (total_count == 0) {
      return(
        ggplot() + 
          geom_text(aes(x = 0.5, y = 0.5, label = "No data available"), size = 6) +
          xlim(0, 1) + ylim(0, 1) + theme_void()
      )
    }
    
    pie_data <- data.frame(
      Status = c("Valid", "Invalid"),
      Count = c(valid_count, invalid_count),
      stringsAsFactors = FALSE
    )
    
    ggplot(pie_data, aes(x = "", y = Count, fill = Status)) +
      geom_col() +
      coord_polar("y", start = 0) +
      scale_fill_manual(values = c("Valid" = "#28a745", "Invalid" = "#dc3545")) +
      labs(title = paste0("Valid: ", summary$identifier_percentage, "%")) +
      theme_void() +
      theme(legend.position = "bottom")
  }, height = 300)
  
  output$name_chart <- renderPlot({
    req(quality_summary())
    
    summary <- quality_summary()
    
    # Ensure we have numeric data
    valid_count <- as.numeric(summary$valid_name_count)
    total_count <- as.numeric(summary$total_orgs)
    invalid_count <- total_count - valid_count
    
    if (total_count == 0) {
      return(
        ggplot() + 
          geom_text(aes(x = 0.5, y = 0.5, label = "No data available"), size = 6) +
          xlim(0, 1) + ylim(0, 1) + theme_void()
      )
    }
    
    pie_data <- data.frame(
      Status = c("Quality", "Needs Improvement"),
      Count = c(valid_count, invalid_count),
      stringsAsFactors = FALSE
    )
    
    ggplot(pie_data, aes(x = "", y = Count, fill = Status)) +
      geom_col() +
      coord_polar("y", start = 0) +
      scale_fill_manual(values = c("Quality" = "#007bff", "Needs Improvement" = "#ffc107")) +
      labs(title = paste0("Quality: ", summary$name_percentage, "%")) +
      theme_void() +
      theme(legend.position = "bottom")
  }, height = 300)
  
  output$address_chart <- renderPlot({
    req(quality_summary())
    
    summary <- quality_summary()
    
    # Ensure we have numeric data
    valid_count <- as.numeric(summary$valid_address_count)
    total_count <- as.numeric(summary$total_orgs)
    invalid_count <- total_count - valid_count
    
    if (total_count == 0) {
      return(
        ggplot() + 
          geom_text(aes(x = 0.5, y = 0.5, label = "No data available"), size = 6) +
          xlim(0, 1) + ylim(0, 1) + theme_void()
      )
    }
    
    pie_data <- data.frame(
      Status = c("Complete", "Incomplete"),
      Count = c(valid_count, invalid_count),
      stringsAsFactors = FALSE
    )
    
    ggplot(pie_data, aes(x = "", y = Count, fill = Status)) +
      geom_col() +
      coord_polar("y", start = 0) +
      scale_fill_manual(values = c("Complete" = "#fd7e14", "Incomplete" = "#6c757d")) +
      labs(title = paste0("Complete: ", summary$address_percentage, "%")) +
      theme_void() +
      theme(legend.position = "bottom")
  }, height = 300)
  
  # Organization identifier status breakdown chart - FIXED
  output$organization_identifier_status_chart <- renderPlot({
    req(identifier_type_summary())
    
    id_summary <- identifier_type_summary()
    
    status_data <- data.frame(
      Status = c("Organizations with Valid Identifiers", 
                 "Organizations with No Identifiers", 
                 "Organizations with Only Invalid Identifiers"),
      Count = c(
        as.numeric(id_summary$orgs_with_valid),
        as.numeric(id_summary$orgs_with_no_identifiers),
        as.numeric(id_summary$orgs_with_invalid_only)
      ),
      stringsAsFactors = FALSE
    )
    
    # Calculate percentages
    total_orgs <- sum(status_data$Count)
    if (total_orgs > 0) {
      status_data$Percentage <- round(status_data$Count / total_orgs * 100, 1)
    } else {
      status_data$Percentage <- 0
      return(
        ggplot() + 
          geom_text(aes(x = 0.5, y = 0.5, label = "No data available"), size = 6) +
          xlim(0, 1) + ylim(0, 1) + theme_void() +
          labs(title = "Organization Breakdown by Identifier Status")
      )
    }
    
    # Define colors
    colors <- c("Organizations with Valid Identifiers" = "#28a745",
                "Organizations with No Identifiers" = "#6c757d", 
                "Organizations with Only Invalid Identifiers" = "#dc3545")
    
    ggplot(status_data, aes(x = reorder(Status, Count), y = Count, fill = Status)) +
      geom_col(width = 0.7) +
      geom_text(aes(label = paste0(format(Count, big.mark = ","), "\n(", Percentage, "%)")), 
                hjust = -0.1, size = 3.5) +
      scale_fill_manual(values = colors) +
      coord_flip() +
      labs(title = "Organization Breakdown by Identifier Status",
           x = "",
           y = "Number of Organizations") +
      theme_minimal() +
      theme(legend.position = "none",
            axis.text.y = element_text(size = 10)) +
      scale_x_discrete(labels = function(x) str_wrap(x, width = 25))
  }, height = 300)
  
  # Identifier type distribution chart
  output$identifier_type_distribution_chart <- renderPlot({
    req(identifier_type_summary())
    
    tryCatch({
      id_summary <- identifier_type_summary()
      
      chart_data <- data.frame(
        Type = c("NPI", "CLIA", "NAIC", "Other", "No Identifier Data"),
        Count = c(
          as.numeric(id_summary$npi_count), 
          as.numeric(id_summary$clia_count), 
          as.numeric(id_summary$naic_count), 
          as.numeric(id_summary$other_count), 
          as.numeric(id_summary$no_identifier_count)
        ),
        stringsAsFactors = FALSE
      )
      
      # Filter out zero counts
      chart_data <- chart_data[chart_data$Count > 0, ]
      
      if (nrow(chart_data) == 0) {
        return(
          ggplot() + 
            geom_text(aes(x = 0.5, y = 0.5, label = "No identifier data found"), 
                     size = 6) +
            theme_void() + xlim(0, 1) + ylim(0, 1) +
            labs(title = "Distribution of Identifier Types")
        )
      }
      
      # Define colors for different types
      type_colors <- c("NPI" = "#28a745", "CLIA" = "#007bff", "NAIC" = "#fd7e14", 
                      "Other" = "#dc3545", "No Identifier Data" = "#6c757d")
      
      ggplot(chart_data, aes(x = reorder(Type, Count), y = Count, fill = Type)) +
        geom_col(width = 0.7) +
        geom_text(aes(label = format(Count, big.mark = ",")), hjust = -0.1) +
        scale_fill_manual(values = type_colors) +
        coord_flip() +
        labs(title = "Distribution of Identifier Types",
             x = "Identifier Type",
             y = "Count") +
        theme_minimal() +
        theme(legend.position = "none")
      
    }, error = function(e) {
      ggplot() + 
        geom_text(aes(x = 0.5, y = 0.5, label = paste("Error:", e$message)), 
                 size = 4) +
        theme_void() + xlim(0, 1) + ylim(0, 1) +
        labs(title = "Chart Error")
    })
  }, height = 400)
  
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
      filter(Valid + Invalid > 0)  # Only show types that have data
    
    if (nrow(conformance_data) == 0) {
      return(
        ggplot() + 
          geom_text(aes(x = 0.5, y = 0.5, label = "No conformance data available"), 
                   size = 6) +
          theme_void() + xlim(0, 1) + ylim(0, 1) +
          labs(title = "Conformance by Identifier Type")
      )
    }
    
    conformance_long <- conformance_data %>%
      pivot_longer(cols = c(Valid, Invalid), names_to = "Status", values_to = "Count")
    
    ggplot(conformance_long, aes(x = Type, y = Count, fill = Status)) +
      geom_col(position = "stack") +
      geom_text(aes(label = Count), position = position_stack(vjust = 0.5)) +
      scale_fill_manual(values = c("Valid" = "#28a745", "Invalid" = "#dc3545")) +
      labs(title = "US-Core Conformance by Type",
           x = "Identifier Type",
           y = "Count") +
      theme_minimal()
  }, height = 400)
  
  # Identifier type detail table with pre-computed data
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
        Identifier_Type = colDef(name = "Type", width = 120),
        Total_Count = colDef(name = "Total", format = colFormat(separators = TRUE), width = 80),
        Valid_Count = colDef(name = "Valid", format = colFormat(separators = TRUE), width = 80),
        Invalid_Count = colDef(name = "Invalid", format = colFormat(separators = TRUE), width = 80),
        Conformance_Rate = colDef(
          name = "Conformance Rate", 
          width = 120,
          cell = function(value) {
            if (value == "N/A") {
              div(style = "color: #6c757d;", value)
            } else {
              rate <- as.numeric(str_extract(value, "\\d+"))
              if (!is.na(rate)) {
                if (rate >= 90) {
                  div(style = "color: #28a745; font-weight: bold;", value)
                } else if (rate >= 70) {
                  div(style = "color: #ffc107; font-weight: bold;", value)  
                } else {
                  div(style = "color: #dc3545; font-weight: bold;", value)
                }
              } else {
                div(style = "color: #6c757d;", value)
              }
            }
          }
        ),
        Percentage_of_Orgs = colDef(name = "% of Orgs", width = 100),
        US_Core_Rules = colDef(name = "US-Core Rules", width = 150),
        Validation_Requirements = colDef(name = "Format Requirements", minWidth = 200)
      ),
      striped = TRUE,
      highlight = TRUE
    )
  })
  
  # Issues detail table using pre-computed summaries
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
        Issue_Category = colDef(name = "Issue Category", width = 150),
        Total_Count = colDef(name = "Total", format = colFormat(separators = TRUE)),
        Valid_Count = colDef(name = "Valid", format = colFormat(separators = TRUE)),
        Invalid_Count = colDef(name = "Invalid", format = colFormat(separators = TRUE)),
        Success_Rate = colDef(name = "Success Rate", width = 100),
        Common_Issues = colDef(name = "Common Issues", minWidth = 400),
        US_Core_Reference = colDef(
          name = "US-Core Reference",
          width = 150,
          cell = function(value) {
            tags$a(href = value, target = "_blank", "View Specification")
          }
        )
      ),
      striped = TRUE,
      highlight = TRUE
    )
  })
  
  # Enhanced recommendations using pre-computed data
  output$recommendations <- renderUI({
    req(quality_summary(), identifier_type_summary())
    
    summary <- quality_summary()
    id_summary <- identifier_type_summary()
    recommendations <- list()
    
    # No identifier data alert (highest priority)
    if (id_summary$no_identifier_count > 0) {
      no_id_percentage <- round(id_summary$no_identifier_count / summary$total_orgs * 100, 1)
      recommendations <- append(recommendations,
        tags$div(class = "alert alert-danger", style = "margin-bottom: 10px;",
          tags$strong("Missing Identifier Data: "),
          paste0(format(id_summary$no_identifier_count, big.mark = ","), 
                 " organizations (", no_id_percentage, "%) have no identifier data. "),
          tags$br(),
          tags$small("Organizations must include at least one identifier (NPI, CLIA, or NAIC) to meet US-Core requirements.")
        )
      )
    }
    
    # Invalid only identifiers alert
    if (id_summary$orgs_with_invalid_only > 0) {
      invalid_only_percentage <- round(id_summary$orgs_with_invalid_only / summary$total_orgs * 100, 1)
      recommendations <- append(recommendations,
        tags$div(class = "alert alert-danger", style = "margin-bottom: 10px;",
          tags$strong("Organizations with Only Invalid Identifiers: "),
          paste0(format(id_summary$orgs_with_invalid_only, big.mark = ","), 
                 " organizations (", invalid_only_percentage, "%) have identifiers but none are US-Core compliant. "),
          tags$br(),
          tags$small("Review identifier formats and ensure compliance with US-Core validation rules.")
        )
      )
    }
    
    # Identifier conformance recommendations
    if (summary$identifier_percentage < 80) {
      recommendations <- append(recommendations, 
        tags$div(class = "alert alert-warning", style = "margin-bottom: 10px;",
          tags$strong("US-Core Identifier Conformance Issues: "),
          paste0("Only ", summary$identifier_percentage, "% of organizations have conformant identifiers. "),
          tags$br(),
          tags$small("Ensure NPI identifiers are 10 digits with valid check digits, CLIA identifiers follow 2D7 format, and NAIC identifiers are 5 digits.")
        )
      )
    }
    
    if (id_summary$other_count > 0) {
      recommendations <- append(recommendations, 
        tags$div(class = "alert alert-warning", style = "margin-bottom: 10px;",
          tags$strong("Non-Standard Identifiers: "),
          paste0("Found ", format(id_summary$other_count, big.mark = ","), 
                 " non-standard identifier types. "),
          tags$br(),
          tags$small("Use US-Core compliant types: NPI (healthcare providers), CLIA (laboratories), NAIC (insurance).")
        )
      )
    }
    
    # Specific validation error recommendations
    if (id_summary$npi_invalid > 0) {
      recommendations <- append(recommendations,
        tags$div(class = "alert alert-warning", style = "margin-bottom: 10px;",
          tags$strong("Invalid NPI Identifiers: "),
          paste0(format(id_summary$npi_invalid, big.mark = ","), " NPIs failed validation (us-core-16/17). "),
          tags$br(),
          tags$small("Verify NPIs are exactly 10 digits and have valid Luhn check digits.")
        )
      )
    }
    
    if (id_summary$clia_invalid > 0) {
      recommendations <- append(recommendations,
        tags$div(class = "alert alert-warning", style = "margin-bottom: 10px;",
          tags$strong("Invalid CLIA Identifiers: "),
          paste0(format(id_summary$clia_invalid, big.mark = ","), " CLIAs failed validation (us-core-18). "),
          tags$br(),
          tags$small("CLIA format must be: 2 digits + 'D' + 7 digits (e.g., '12D3456789').")
        )
      )
    }
    
    if (id_summary$naic_invalid > 0) {
      recommendations <- append(recommendations,
        tags$div(class = "alert alert-warning", style = "margin-bottom: 10px;",
          tags$strong("Invalid NAIC Identifiers: "),
          paste0(format(id_summary$naic_invalid, big.mark = ","), " NAICs failed validation (us-core-19). "),
          tags$br(),
          tags$small("NAIC identifiers must be exactly 5 digits.")
        )
      )
    }
    
    if (summary$name_percentage < 80) {
      recommendations <- append(recommendations,
        tags$div(class = "alert alert-info", style = "margin-bottom: 10px;",
          tags$strong("Name Quality: "),
          "Use complete, meaningful organization names instead of placeholders."
        )
      )
    }
    
    if (summary$address_percentage < 80) {
      recommendations <- append(recommendations,
        tags$div(class = "alert alert-secondary", style = "margin-bottom: 10px;",
          tags$strong("Address Issues: "),
          "Include complete addresses with street, city, state, and ZIP code."
        )
      )
    }
    
    if (length(recommendations) == 0) {
      recommendations <- list(
        tags$div(class = "alert alert-success", style = "margin-bottom: 10px;",
          tags$strong("Excellent US-Core compliance! "),
          "Your organization data meets quality and conformance standards."
        )
      )
    }
    
    do.call(tagList, recommendations)
  })
  
  # Download handler using detailed organization data
  output$download_feedback_report <- downloadHandler(
    filename = function() {
      paste0("organization_data_quality_report_", Sys.Date(), ".csv")
    },
    content = function(file) {
      data <- filtered_org_data()
      
      if (nrow(data) > 0) {
        # Create a detailed report with the pre-computed validation results
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
        # Write empty file with headers if no data
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