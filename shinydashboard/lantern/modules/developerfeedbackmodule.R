library(DT)
library(purrr)
library(reactable)
library(ggplot2)
library(dplyr)
library(stringr)
library(shiny)
library(htmltools)

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
  
  # validate the NPI using Luhn's algorithm
  validate_npi_luhn <- function(npi) {
    if (nchar(npi) != 10 || !grepl("^[0-9]{10}$", npi)) {
      return(FALSE)
    }
    
    digits <- as.numeric(strsplit(npi, "")[[1]])
    checksum <- 0
    
    # Double digits in positions 1,3,5,7,9 (1-indexed → 0,2,4,6,8 in 0-indexed)
    for (i in c(1, 3, 5, 7, 9)) {
      doubled <- digits[i] * 2
      if (doubled > 9) {
        doubled <- doubled - 9
      }
      checksum <- checksum + doubled
    }
    
    # Add the other digits (positions 2,4,6,8,10)
    for (i in c(2, 4, 6, 8, 10)) {
      checksum <- checksum + digits[i]
    }
    
    checksum <- checksum + 24
    return((checksum %% 10) == 0)
  }
  
  # Validate individual identifier based on type and value
  validate_identifier_value <- function(identifier_type, identifier_value) {
    if (is.na(identifier_value) || identifier_value == "") {
      return(list(valid = FALSE, error = "Missing identifier value"))
    }
    
    identifier_type <- toupper(str_trim(identifier_type))
    identifier_value <- str_trim(identifier_value)
    
    if (identifier_type == "NPI") {
      # us-core-16: NPI must be 10 digits
      if (!grepl("^[0-9]{10}$", identifier_value)) {
        return(list(valid = FALSE, error = "NPI must be exactly 10 digits"))
      }
      
      # us-core-17: NPI check digit must be valid (Luhn algorithm)
      if (!validate_npi_luhn(identifier_value)) {
        return(list(valid = FALSE, error = "NPI check digit is invalid (Luhn algorithm failed)"))
      }
      
      return(list(valid = TRUE, error = NULL))
      
    } else if (identifier_type == "CLIA") {
      # us-core-18: CLIA number must be 10 digits with a letter "D" in third position
      if (!grepl("^[0-9]{2}D[0-9]{7}$", identifier_value)) {
        return(list(valid = FALSE, error = "CLIA must be 10 characters: 2 digits + 'D' + 7 digits"))
      }
      
      return(list(valid = TRUE, error = NULL))
      
    } else if (identifier_type == "NAIC") {
      # us-core-19: NAIC must be 5 digits
      if (!grepl("^[0-9]{5}$", identifier_value)) {
        return(list(valid = FALSE, error = "NAIC must be exactly 5 digits"))
      }
      
      return(list(valid = TRUE, error = NULL))
      
    } else {
      # Non-standard identifier type
      return(list(valid = FALSE, error = "Non-standard identifier type (should use NPI, CLIA, or NAIC)"))
    }
  }
  
  # Enhanced identifier validation function
  is_valid_identifier <- function(identifier_types, identifier_values) {
    # Check if both fields are completely empty/missing
    if ((is.na(identifier_types) || identifier_types == "") && 
        (is.na(identifier_values) || identifier_values == "")) {
      return(list(
        valid = FALSE, 
        errors = list("No identifier data provided"), 
        conformant_count = 0, 
        total_count = 0,
        status = "no_identifiers"
      ))
    }
    
    # Check if either field is missing when the other has content
    if ((is.na(identifier_types) || identifier_types == "") || 
        (is.na(identifier_values) || identifier_values == "")) {
      return(list(
        valid = FALSE, 
        errors = list("Incomplete identifier data - missing types or values"), 
        conformant_count = 0, 
        total_count = 0,
        status = "incomplete_data"
      ))
    }
    
    # Parse identifier types and values
    type_lines <- unlist(str_split(identifier_types, "<br/>|<br>|\\n"))
    value_lines <- unlist(str_split(identifier_values, "<br/>|<br>|\\n"))
    
    type_lines <- str_trim(type_lines)
    value_lines <- str_trim(value_lines)
    
    # Remove empty entries
    type_lines <- type_lines[type_lines != ""]
    value_lines <- value_lines[value_lines != ""]
    
    # Check if counts match
    if (length(type_lines) != length(value_lines)) {
      return(list(
        valid = FALSE, 
        errors = list("Mismatch between number of identifier types and values"), 
        conformant_count = 0, 
        total_count = length(type_lines),
        status = "mismatched_data"
      ))
    }
    
    if (length(type_lines) == 0) {
      return(list(
        valid = FALSE, 
        errors = list("No valid identifier data found"), 
        conformant_count = 0, 
        total_count = 0,
        status = "no_identifiers"
      ))
    }
    
    # Validate each identifier
    errors <- list()
    conformant_count <- 0
    total_count <- length(type_lines)
    
    for (i in seq_along(type_lines)) {
      validation_result <- validate_identifier_value(type_lines[i], value_lines[i])
      
      if (!validation_result$valid) {
        errors <- append(errors, paste0(type_lines[i], ": ", validation_result$error))
      } else {
        conformant_count <- conformant_count + 1
      }
    }
    
    # Organization is valid if at least one identifier is conformant
    overall_valid <- conformant_count > 0
    
    # Determine status
    status <- if (conformant_count == 0) {
      "invalid_only"
    } else if (conformant_count == total_count) {
      "all_valid"
    } else {
      "mixed_valid_invalid"
    }
    
    return(list(
      valid = overall_valid, 
      errors = if(length(errors) > 0) errors else NULL,
      conformant_count = conformant_count,
      total_count = total_count,
      conformance_rate = round(conformant_count / total_count * 100, 1),
      status = status
    ))
  }
  
  # Enhanced identifier counting function with validation details
  get_identifier_counts <- function(identifier_types, identifier_values = NULL) {
    counts <- list(
      NPI = 0, CLIA = 0, NAIC = 0, Other = 0, NoIdentifier = 0,
      NPI_valid = 0, CLIA_valid = 0, NAIC_valid = 0,
      NPI_invalid = 0, CLIA_invalid = 0, NAIC_invalid = 0, Other_invalid = 0,
      total_conformant = 0, total_identifiers = 0
    )
    
    if (is.na(identifier_types) || identifier_types == "") {
      counts$NoIdentifier <- 1
      return(counts)
    }
    
    # Parse types
    type_lines <- unlist(str_split(identifier_types, "<br/>|<br>|\\n"))
    type_lines <- str_trim(type_lines)
    type_lines <- type_lines[type_lines != ""]
    
    if (length(type_lines) == 0) {
      counts$NoIdentifier <- 1
      return(counts)
    }
    
    # Parse values if provided
    value_lines <- NULL
    if (!is.null(identifier_values) && !is.na(identifier_values) && identifier_values != "") {
      value_lines <- unlist(str_split(identifier_values, "<br/>|<br>|\\n"))
      value_lines <- str_trim(value_lines)
      value_lines <- value_lines[value_lines != ""]
    }
    
    # Count and validate each identifier
    for (i in seq_along(type_lines)) {
      type_upper <- toupper(str_trim(type_lines[i]))
      
      # Count by type
      if (type_upper == "NPI") {
        counts$NPI <- counts$NPI + 1
      } else if (type_upper == "CLIA") {
        counts$CLIA <- counts$CLIA + 1
      } else if (type_upper == "NAIC") {
        counts$NAIC <- counts$NAIC + 1
      } else {
        counts$Other <- counts$Other + 1
        counts$Other_invalid <- counts$Other_invalid + 1  # All "Other" types are invalid
      }
      
      counts$total_identifiers <- counts$total_identifiers + 1
      
      # Validate if we have both type and value
      if (!is.null(value_lines) && i <= length(value_lines)) {
        validation_result <- validate_identifier_value(type_lines[i], value_lines[i])
        
        if (validation_result$valid) {
          counts$total_conformant <- counts$total_conformant + 1
          
          if (type_upper == "NPI") {
            counts$NPI_valid <- counts$NPI_valid + 1
          } else if (type_upper == "CLIA") {
            counts$CLIA_valid <- counts$CLIA_valid + 1
          } else if (type_upper == "NAIC") {
            counts$NAIC_valid <- counts$NAIC_valid + 1
          }
        } else {
          if (type_upper == "NPI") {
            counts$NPI_invalid <- counts$NPI_invalid + 1
          } else if (type_upper == "CLIA") {
            counts$CLIA_invalid <- counts$CLIA_invalid + 1
          } else if (type_upper == "NAIC") {
            counts$NAIC_invalid <- counts$NAIC_invalid + 1
          }
        }
      }
    }
    
    return(counts)
  }
  
  # Address-like name detection function
  is_address_like <- function(name) {
    if (is.na(name) || name == "") return(FALSE)
    
    # Clean the name first
    clean_name <- str_remove_all(name, "<[^>]+>")
    clean_name <- str_trim(clean_name)
    
    # Common street suffixes (more comprehensive for healthcare)
    street_pattern <- "\\b(St|Street|Ave|Avenue|Blvd|Boulevard|Rd|Road|Dr|Drive|Ln|Lane|Ct|Court|Cir|Circle|Way|Pl|Place|Pkwy|Parkway|Ter|Terrace)\\b"
    unit_pattern <- "\\b(Suite|Ste|Apt|Apartment|Unit|Floor|Fl|Room|Rm|Building|Bldg|#)\\b"
    zip_pattern <- "\\b[0-9]{5}(-[0-9]{4})?\\b"
    state_pattern <- "\\b(AL|AK|AZ|AR|CA|CO|CT|DE|FL|GA|HI|ID|IL|IN|IA|KS|KY|LA|ME|MD|MA|MI|MN|MS|MO|MT|NE|NV|NH|NJ|NM|NY|NC|ND|OH|OK|OR|PA|RI|SC|SD|TN|TX|UT|VT|VA|WA|WV|WI|WY)\\b"
    
    score <- 0
    
    # Strong address indicators
    if (grepl("^[0-9]+", clean_name)) score <- score + 3  # Starts with number
    if (grepl(street_pattern, clean_name, ignore.case = TRUE)) score <- score + 3  # Street suffix
    if (grepl(unit_pattern, clean_name, ignore.case = TRUE)) score <- score + 2   # Unit/suite
    if (grepl(zip_pattern, clean_name)) score <- score + 3                        # ZIP code
    if (grepl(state_pattern, clean_name)) score <- score + 2                      # State abbreviation
    
    # Additional address-like patterns
    if (grepl("\\b(North|South|East|West|N|S|E|W)\\b", clean_name, ignore.case = TRUE)) score <- score + 1  # Directional
    if (str_count(clean_name, ",") >= 2) score <- score + 2  # Multiple commas (address format)
    
    # Healthcare/organization keywords that reduce address likelihood
    healthcare_org_pattern <- "\\b(Hospital|Clinic|Center|Centre|Health|Medical|System|Services|LLC|Corp|Corporation|Inc|Incorporated|Ltd|Limited|Associates|Group|Foundation|Institute|University|College|Pharmacy|Laboratory|Labs?)\\b"
    
    if (grepl(healthcare_org_pattern, clean_name, ignore.case = TRUE)) {
      score <- score - 3  # Strong negative signal
    }
    
    # Additional org-like terms specific to healthcare
    if (grepl("\\b(Family|Internal|Primary|Urgent|Emergency|Pediatric|Cardiology|Orthopedic|Dental|Vision|Eye|Care)\\b", clean_name, ignore.case = TRUE)) {
      score <- score - 2
    }
    
    return(score >= 4)  # Adjusted threshold
  }
  
  # Enhanced name validation function
  is_valid_name <- function(org_name) {
    if (is.na(org_name) || org_name == "") return(FALSE)
    
    clean_name <- str_remove_all(org_name, "<[^>]+>")
    clean_name <- str_trim(clean_name)
    
    if (nchar(clean_name) < 3) return(FALSE)
    
    placeholder_patterns <- c("-", "\\.", "N/A", "NA", "UNKNOWN", "TEST", "EXAMPLE", "TBD", "TODO")
    upper_name <- toupper(clean_name)
    
    for (pattern in placeholder_patterns) {
      if (grepl(paste0("^", pattern, "$"), upper_name)) return(FALSE)
    }
    
    # Reject if all digits or digits with only separators/symbols 
    if (grepl("^[0-9]+$", clean_name)) return(FALSE) # only digits 
    if (grepl("^[0-9()/.\\-]+$", clean_name)) return(FALSE) # digits + separators only (ZIP codes, codes, phone-like strings)
    if (grepl("^\\W+$", clean_name)) return(FALSE) # only non-word chars (".", "-", "...")
    
    # Reject if it matches a phone number pattern (###-###-#### or (###) ###-####) 
    if (grepl("^\\(?\\d{3}\\)?[- ]?\\d{3}[- ]?\\d{4}$", clean_name)) return(FALSE)
    
    # Reject if it looks like an address
    if (is_address_like(clean_name)) return(FALSE)
    
    special_chars <- str_count(clean_name, "[^a-zA-Z0-9 ]")
    if (special_chars / nchar(clean_name) > 0.3) return(FALSE)
    
    return(TRUE)
  }
  
  # Address validation function
  is_valid_address <- function(address) {
    if (is.na(address) || address == "") return(FALSE)
    
    clean_address <- str_remove_all(address, "<[^>]+>")
    clean_address <- str_trim(clean_address)
    
    if (nchar(clean_address) < 10) return(FALSE)
    
    placeholder_addresses <- c("123 MAIN ST", "123 TEST ST", "123 MAIN STREET", "123 TEST STREET")
    upper_address <- toupper(clean_address)
    
    for (placeholder in placeholder_addresses) {
      if (grepl(placeholder, upper_address)) return(FALSE)
    }
    
    has_street_number <- grepl("\\d+", clean_address)
    has_city_state <- str_count(clean_address, ",") >= 2
    has_zip <- grepl("\\d{5}", clean_address)
    
    return(has_street_number && has_city_state && has_zip)
  }
  
  # Initialize vendor choices
  observe({
    vendor_choices <- c("All Developers", app$vendor_list())
    updateSelectInput(session, "vendor_filter", choices = vendor_choices, selected = "All Developers")
  })
  
  # Get filtered organization data with enhanced validation
  filtered_org_data <- reactive({
    current_vendor <- input$vendor_filter
    if (is.null(current_vendor)) current_vendor <- "All Developers"
    
    # Build query to get organization data
    query_str <- "
      SELECT
        organization_name,
        identifier_types_html as identifier_types,
        identifier_values_html as identifier_values,
        addresses_html as address,
        vendor_names_array
      FROM mv_organizations_final
      WHERE TRUE"
    
    params <- list()
    
    if (current_vendor != "All Developers") {
      query_str <- paste0(query_str, " AND vendor_names_array && ARRAY[{vendor}]")
      params$vendor <- current_vendor
    }
    
    # Execute query
    if (length(params) > 0) {
      data_query <- do.call(glue::glue_sql, c(list(query_str, .con = db_connection), params))
    } else {
      data_query <- glue::glue_sql(query_str, .con = db_connection)
    }
    
    result <- tbl(db_connection, sql(data_query)) %>% collect()
    
    # Add validation columns using enhanced validation
    result <- result %>%
      mutate(
        # Enhanced identifier validation
        identifier_validation = map2(identifier_types, identifier_values, is_valid_identifier),
        valid_identifier = map_lgl(identifier_validation, ~ .$valid),
        identifier_conformant_count = map_dbl(identifier_validation, ~ .$conformant_count),
        identifier_total_count = map_dbl(identifier_validation, ~ .$total_count),
        identifier_conformance_rate = map_dbl(identifier_validation, ~ .$conformance_rate %||% 0),
        identifier_errors = map(identifier_validation, ~ .$errors),
        identifier_status = map_chr(identifier_validation, ~ .$status),
        
        # Enhanced identifier counts with validation details
        identifier_counts = map2(identifier_types, identifier_values, get_identifier_counts),
        
        # Keep existing validations
        valid_name = map_lgl(organization_name, is_valid_name),
        valid_address = map_lgl(address, is_valid_address),
        
        # Updated overall quality score (now includes conformance)
        overall_quality = valid_identifier + valid_name + valid_address,
        
        # Add conformance-specific metrics
        has_conformant_identifiers = identifier_conformant_count > 0,
        identifier_conformance_category = case_when(
          identifier_conformance_rate == 100 ~ "Fully Conformant",
          identifier_conformance_rate >= 50 ~ "Partially Conformant", 
          identifier_conformant_count > 0 ~ "Minimally Conformant",
          TRUE ~ "Non-Conformant"
        )
      )
    
    return(result)
  })
  
  # Enhanced identifier type summary with validation details including no identifier tracking
  identifier_type_summary <- reactive({
    data <- filtered_org_data()
    
    if (nrow(data) == 0) {
      return(list(
        npi_count = 0, clia_count = 0, naic_count = 0, other_count = 0, no_identifier_count = 0,
        npi_valid = 0, clia_valid = 0, naic_valid = 0,
        npi_invalid = 0, clia_invalid = 0, naic_invalid = 0, other_invalid = 0,
        total_identifiers = 0, total_conformant = 0,
        npi_percentage = 0, clia_percentage = 0, naic_percentage = 0, other_percentage = 0, no_identifier_percentage = 0,
        conformance_rate = 0,
        orgs_with_no_identifiers = 0, orgs_with_invalid_only = 0, orgs_with_valid = 0,
        total_organizations = 0
      ))
    }
    
    # Aggregate counts across all organizations
    total_npi <- 0
    total_clia <- 0  
    total_naic <- 0
    total_other <- 0
    total_no_identifier <- 0
    total_npi_valid <- 0
    total_clia_valid <- 0
    total_naic_valid <- 0
    total_npi_invalid <- 0
    total_clia_invalid <- 0
    total_naic_invalid <- 0
    total_other_invalid <- 0
    total_conformant <- 0
    total_identifiers <- 0
    
    # Count organizations by status
    orgs_with_no_identifiers <- sum(data$identifier_status == "no_identifiers", na.rm = TRUE)
    orgs_with_invalid_only <- sum(data$identifier_status == "invalid_only", na.rm = TRUE)
    orgs_with_valid <- sum(data$identifier_status %in% c("all_valid", "mixed_valid_invalid"), na.rm = TRUE)
    
    for (i in seq_along(data$identifier_counts)) {
      counts <- data$identifier_counts[[i]]
      if (is.list(counts)) {
        total_npi <- total_npi + (counts$NPI %||% 0)
        total_clia <- total_clia + (counts$CLIA %||% 0)
        total_naic <- total_naic + (counts$NAIC %||% 0)
        total_other <- total_other + (counts$Other %||% 0)
        total_no_identifier <- total_no_identifier + (counts$NoIdentifier %||% 0)
        
        total_npi_valid <- total_npi_valid + (counts$NPI_valid %||% 0)
        total_clia_valid <- total_clia_valid + (counts$CLIA_valid %||% 0)
        total_naic_valid <- total_naic_valid + (counts$NAIC_valid %||% 0)
        
        total_npi_invalid <- total_npi_invalid + (counts$NPI_invalid %||% 0)
        total_clia_invalid <- total_clia_invalid + (counts$CLIA_invalid %||% 0)
        total_naic_invalid <- total_naic_invalid + (counts$NAIC_invalid %||% 0)
        total_other_invalid <- total_other_invalid + (counts$Other_invalid %||% 0)
        
        total_conformant <- total_conformant + (counts$total_conformant %||% 0)
        total_identifiers <- total_identifiers + (counts$total_identifiers %||% 0)
      }
    }
    
    total_organizations <- nrow(data)
    
    list(
      npi_count = total_npi,
      clia_count = total_clia,
      naic_count = total_naic,
      other_count = total_other,
      no_identifier_count = total_no_identifier,
      npi_valid = total_npi_valid,
      clia_valid = total_clia_valid,
      naic_valid = total_naic_valid,
      npi_invalid = total_npi_invalid,
      clia_invalid = total_clia_invalid,
      naic_invalid = total_naic_invalid,
      other_invalid = total_other_invalid,
      total_identifiers = total_identifiers,
      total_conformant = total_conformant,
      npi_percentage = if (total_identifiers > 0) round(total_npi / total_identifiers * 100, 1) else 0,
      clia_percentage = if (total_identifiers > 0) round(total_clia / total_identifiers * 100, 1) else 0,
      naic_percentage = if (total_identifiers > 0) round(total_naic / total_identifiers * 100, 1) else 0,
      other_percentage = if (total_identifiers > 0) round(total_other / total_identifiers * 100, 1) else 0,
      no_identifier_percentage = if (total_organizations > 0) round(total_no_identifier / total_organizations * 100, 1) else 0,
      conformance_rate = if (total_identifiers > 0) round(total_conformant / total_identifiers * 100, 1) else 0,
      orgs_with_no_identifiers = orgs_with_no_identifiers,
      orgs_with_invalid_only = orgs_with_invalid_only,
      orgs_with_valid = orgs_with_valid,
      total_organizations = total_organizations
    )
  })
  
  # Updated summary statistics with conformance metrics
  quality_summary <- reactive({
    data <- filtered_org_data()
    
    list(
      total_orgs = nrow(data),
      valid_identifier_count = sum(data$valid_identifier, na.rm = TRUE),
      conformant_identifier_count = sum(data$has_conformant_identifiers, na.rm = TRUE),
      valid_name_count = sum(data$valid_name, na.rm = TRUE),
      valid_address_count = sum(data$valid_address, na.rm = TRUE),
      high_quality_count = sum(data$overall_quality >= 2, na.rm = TRUE),
      low_quality_count = sum(data$overall_quality <= 1, na.rm = TRUE),
      identifier_percentage = round(sum(data$valid_identifier, na.rm = TRUE) / nrow(data) * 100, 1),
      identifier_conformance_percentage = round(sum(data$has_conformant_identifiers, na.rm = TRUE) / nrow(data) * 100, 1),
      name_percentage = round(sum(data$valid_name, na.rm = TRUE) / nrow(data) * 100, 1),
      address_percentage = round(sum(data$valid_address, na.rm = TRUE) / nrow(data) * 100, 1),
      # Conformance category breakdown
      fully_conformant = sum(data$identifier_conformance_category == "Fully Conformant", na.rm = TRUE),
      partially_conformant = sum(data$identifier_conformance_category == "Partially Conformant", na.rm = TRUE),
      minimally_conformant = sum(data$identifier_conformance_category == "Minimally Conformant", na.rm = TRUE),
      non_conformant = sum(data$identifier_conformance_category == "Non-Conformant", na.rm = TRUE),
      # Status breakdown
      no_identifiers = sum(data$identifier_status == "no_identifiers", na.rm = TRUE),
      invalid_only = sum(data$identifier_status == "invalid_only", na.rm = TRUE),
      mixed_valid_invalid = sum(data$identifier_status == "mixed_valid_invalid", na.rm = TRUE),
      all_valid = sum(data$identifier_status == "all_valid", na.rm = TRUE)
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
  
  # New identifier breakdown displays
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
    invalid_only_percentage <- round(invalid_only_count / summary$total_orgs * 100, 1)
    paste0(format(invalid_only_count, big.mark = ","), " (", invalid_only_percentage, "%)")
  })
  
  # Overview chart
  output$quality_overview_chart <- renderPlot({
    summary <- quality_summary()
    
    chart_data <- data.frame(
      Category = c("Identifier Type Validation", "Organization Name", "Address Completeness"),
      Valid = c(summary$valid_identifier_count, summary$valid_name_count, summary$valid_address_count),
      Invalid = c(summary$total_orgs - summary$valid_identifier_count,
                  summary$total_orgs - summary$valid_name_count,
                  summary$total_orgs - summary$valid_address_count)
    )
    
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
  })
  
  # Individual charts
  output$identifier_chart <- renderPlot({
    summary <- quality_summary()
    
    pie_data <- data.frame(
      Status = c("Valid", "Invalid"),
      Count = c(summary$valid_identifier_count, summary$total_orgs - summary$valid_identifier_count),
      Percentage = c(summary$identifier_percentage, 100 - summary$identifier_percentage)
    )
    
    ggplot(pie_data, aes(x = "", y = Count, fill = Status)) +
      geom_col() +
      coord_polar("y", start = 0) +
      scale_fill_manual(values = c("Valid" = "#28a745", "Invalid" = "#dc3545")) +
      labs(title = paste0("Valid: ", summary$identifier_percentage, "%")) +
      theme_void() +
      theme(legend.position = "bottom")
  })
  
  # Name chart
  output$name_chart <- renderPlot({
    summary <- quality_summary()
    
    pie_data <- data.frame(
      Status = c("Quality", "Needs Improvement"),
      Count = c(summary$valid_name_count, summary$total_orgs - summary$valid_name_count),
      Percentage = c(summary$name_percentage, 100 - summary$name_percentage)
    )
    
    ggplot(pie_data, aes(x = "", y = Count, fill = Status)) +
      geom_col() +
      coord_polar("y", start = 0) +
      scale_fill_manual(values = c("Quality" = "#007bff", "Needs Improvement" = "#ffc107")) +
      labs(title = paste0("Quality: ", summary$name_percentage, "%")) +
      theme_void() +
      theme(legend.position = "bottom")
  })
  
  # Address chart
  output$address_chart <- renderPlot({
    summary <- quality_summary()
    
    pie_data <- data.frame(
      Status = c("Complete", "Incomplete"),
      Count = c(summary$valid_address_count, summary$total_orgs - summary$valid_address_count),
      Percentage = c(summary$address_percentage, 100 - summary$address_percentage)
    )
    
    ggplot(pie_data, aes(x = "", y = Count, fill = Status)) +
      geom_col() +
      coord_polar("y", start = 0) +
      scale_fill_manual(values = c("Complete" = "#fd7e14", "Incomplete" = "#6c757d")) +
      labs(title = paste0("Complete: ", summary$address_percentage, "%")) +
      theme_void() +
      theme(legend.position = "bottom")
  })
  
  # NEW: Organization identifier status breakdown chart
  output$organization_identifier_status_chart <- renderPlot({
    id_summary <- identifier_type_summary()
    
    status_data <- data.frame(
      Status = c("Organizations with Valid Identifiers", 
                 "Organizations with No Identifiers", 
                 "Organizations with Only Invalid Identifiers"),
      Count = c(id_summary$orgs_with_valid,
                id_summary$orgs_with_no_identifiers,
                id_summary$orgs_with_invalid_only),
      stringsAsFactors = FALSE
    )
    
    # Calculate percentages
    total_orgs <- sum(status_data$Count)
    status_data$Percentage <- round(status_data$Count / total_orgs * 100, 1)
    
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
  })
  
  # Identifier type distribution chart
  output$identifier_type_distribution_chart <- renderPlot({
    tryCatch({
      id_summary <- identifier_type_summary()
      
      chart_data <- data.frame(
        Type = c("NPI", "CLIA", "NAIC", "Other", "No Identifier Data"),
        Count = c(id_summary$npi_count, id_summary$clia_count, 
                  id_summary$naic_count, id_summary$other_count, 
                  id_summary$no_identifier_count),
        stringsAsFactors = FALSE
      )
      
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
  })
  
  # Conformance by type chart
  output$conformance_by_type_chart <- renderPlot({
    id_summary <- identifier_type_summary()
    
    conformance_data <- data.frame(
      Type = c("NPI", "CLIA", "NAIC"),
      Valid = c(id_summary$npi_valid, id_summary$clia_valid, id_summary$naic_valid),
      Invalid = c(id_summary$npi_invalid, id_summary$clia_invalid, id_summary$naic_invalid)
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
  })
  
  # Enhanced identifier type detail table with validation results and no identifier tracking
  output$identifier_type_table <- reactable::renderReactable({
    id_summary <- identifier_type_summary()
    
    type_data <- data.frame(
      Identifier_Type = c("NPI", "CLIA", "NAIC", "Other", "No Identifier Data"),
      Total_Count = c(id_summary$npi_count, id_summary$clia_count, 
                      id_summary$naic_count, id_summary$other_count,
                      id_summary$no_identifier_count),
      Valid_Count = c(id_summary$npi_valid, id_summary$clia_valid,
                      id_summary$naic_valid, 0, 0),  # Other and No Identifier are always invalid
      Invalid_Count = c(id_summary$npi_invalid, id_summary$clia_invalid,
                        id_summary$naic_invalid, id_summary$other_invalid, 0),
      Conformance_Rate = c(
        if(id_summary$npi_count > 0) paste0(round(id_summary$npi_valid / id_summary$npi_count * 100, 1), "%") else "N/A",
        if(id_summary$clia_count > 0) paste0(round(id_summary$clia_valid / id_summary$clia_count * 100, 1), "%") else "N/A",
        if(id_summary$naic_count > 0) paste0(round(id_summary$naic_valid / id_summary$naic_count * 100, 1), "%") else "N/A",
        "0%",
        "N/A"
      ),
      Percentage_of_Orgs = c(
        paste0(round(id_summary$npi_count / id_summary$total_organizations * 100, 1), "%"),
        paste0(round(id_summary$clia_count / id_summary$total_organizations * 100, 1), "%"),
        paste0(round(id_summary$naic_count / id_summary$total_organizations * 100, 1), "%"),
        paste0(round(id_summary$other_count / id_summary$total_organizations * 100, 1), "%"),
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
  
  # Enhanced issues detail table
  output$issues_detail_table <- reactable::renderReactable({
    summary <- quality_summary()
    id_summary <- identifier_type_summary()
    
    issues_data <- data.frame(
      Issue_Category = c("Identifier Type Validation", "Organization Names", "Address Completeness"),
      Total_Count = rep(summary$total_orgs, 3),
      Valid_Count = c(summary$valid_identifier_count, summary$valid_name_count, summary$valid_address_count),
      Invalid_Count = c(summary$total_orgs - summary$valid_identifier_count,
                       summary$total_orgs - summary$valid_name_count,
                       summary$total_orgs - summary$valid_address_count),
      Success_Rate = c(paste0(summary$identifier_percentage, "%"),
                      paste0(summary$name_percentage, "%"),
                      paste0(summary$address_percentage, "%")),
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
      )
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
  
  # Enhanced recommendations with specific US-Core guidance and no identifier alerts
  output$recommendations <- renderUI({
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
    if (summary$identifier_conformance_percentage < 80) {
      recommendations <- append(recommendations, 
        tags$div(class = "alert alert-warning", style = "margin-bottom: 10px;",
          tags$strong("US-Core Identifier Conformance Issues: "),
          paste0("Only ", summary$identifier_conformance_percentage, "% of organizations have conformant identifiers. "),
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
  
  # Download handler
  output$download_feedback_report <- downloadHandler(
    filename = function() {
      paste0("organization_data_quality_report_", Sys.Date(), ".csv")
    },
    content = function(file) {
      data <- filtered_org_data()
      
      report_data <- data %>%
        select(organization_name, valid_identifier, valid_name, valid_address, overall_quality, 
               identifier_conformant_count, identifier_total_count, identifier_conformance_rate,
               identifier_conformance_category, identifier_errors, identifier_counts, identifier_status,
               identifier_types, identifier_values) %>%
        mutate(
          identifier_issues = ifelse(!valid_identifier, "Missing or incomplete identifier data", "Valid"),
          name_issues = ifelse(!valid_name, "Placeholder name or too short", "Valid"),
          address_issues = ifelse(!valid_address, "Incomplete address information", "Valid"),
          quality_score = paste0(overall_quality, "/3"),
          conformance_summary = paste0(identifier_conformant_count, "/", identifier_total_count, " (", identifier_conformance_rate, "%)"),
          npi_count = map_dbl(identifier_counts, ~ .$NPI %||% 0),
          clia_count = map_dbl(identifier_counts, ~ .$CLIA %||% 0),
          naic_count = map_dbl(identifier_counts, ~ .$NAIC %||% 0),
          other_count = map_dbl(identifier_counts, ~ .$Other %||% 0),
          no_identifier = map_dbl(identifier_counts, ~ .$NoIdentifier %||% 0),
          npi_valid = map_dbl(identifier_counts, ~ .$NPI_valid %||% 0),
          clia_valid = map_dbl(identifier_counts, ~ .$CLIA_valid %||% 0),
          naic_valid = map_dbl(identifier_counts, ~ .$NAIC_valid %||% 0),
          us_core_compliant = ifelse(identifier_conformance_rate == 100, "Fully Compliant", 
                                    ifelse(identifier_conformance_rate > 0, "Partially Compliant", "Non-Compliant")),
          validation_errors = map_chr(identifier_errors, ~ if(is.null(.)) "" else paste(., collapse = "; ")),
          clean_identifier_types = str_replace_all(identifier_types, "<br/>", "; "),
          clean_identifier_values = str_replace_all(identifier_values, "<br/>", "; "),
          identifier_status_description = case_when(
            identifier_status == "no_identifiers" ~ "No identifier data provided",
            identifier_status == "invalid_only" ~ "Has identifiers but all are invalid",
            identifier_status == "all_valid" ~ "All identifiers are valid",
            identifier_status == "mixed_valid_invalid" ~ "Mix of valid and invalid identifiers",
            TRUE ~ "Unknown status"
          )
        ) %>%
        select(-identifier_counts, -identifier_errors, -identifier_types, -identifier_values)
      
      write.csv(report_data, file, row.names = FALSE)
    }
  )
}