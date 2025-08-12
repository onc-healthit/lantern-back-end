library(shiny)
library(shinydashboard)

payerregistrationmodule_UI <- function(id) {
  
  ns <- NS(id)
  
  tagList(
    fluidRow(
      column(width = 8,
        div(
          style = "margin-bottom: 15px; padding: 7px; background-color: #e3f2fd; border-left: 4px solid #2196f3;",
          p(strong("Note:"), "Alternatively, you can register FHIR endpoint(s) in FHIR bundle or CSV format by sending an email to", 
            tags$a("Lantern Support.", 
                   href = "mailto:lantern.support@example.com", 
                   class = "lantern-url",
                   style = "color: #1976d2;"),
            style = "margin: 0; color: #1565c0;")
        )
      )
    ),
    
    # Single column layout
    fluidRow(
      column(width = 8, offset = 2,
        # FHIR Endpoint Section
        div(
          style = "background-color: #fff; padding: 20px; margin-bottom: 20px; border-radius: 5px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);",

          # FHIR Endpoint URL
          div(
            style = "margin-bottom: 12px;",
            textInput(ns("fhir_endpoint"), 
                     label = "FHIR Endpoint", 
                     placeholder = "Enter FHIR endpoint URL",
                     width = "100%"),
            uiOutput(ns("fhir_endpoint_error_msg"))
          ),
          
          # User-facing website
          div(
            style = "margin-bottom: 12px;",
            textInput(ns("user_website"), 
                     label = "User-facing website for the endpoint", 
                     placeholder = "Enter website URL",
                     width = "100%")
          ),
          
          # Type of FHIR Endpoint
          div(
            style = "margin-bottom: 0;",
            selectInput(ns("fhir_type"), 
                       label = "Type of FHIR Endpoint",
                       choices = list(
                         "Select" = "",
                         "Payer to payer API" = "payer_to_payer",
                         "Provider Access API" = "provider_access",
                         "Patient Access API" = "patient_access",
                         "Prior Authorization API" = "prior_authorization"
                       ),
                       width = "100%")
          )
        ),
        
        # Organization Section
        div(
          style = "background-color: #fff; padding: 20px; margin-bottom: 20px; border-radius: 5px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);",
          h4("Organization", style = "margin-bottom: 15px; color: #333; margin-top: 0;"),
          
          fluidRow(
            column(width = 6,
              div(
                style = "margin-bottom: 12px;",
                textInput(ns("org_name"), 
                         label = "Name", 
                         width = "100%")
              )
            ),
            column(width = 6,
              div(
                style = "margin-bottom: 12px;",
                textInput(ns("payer_id"), 
                         label = "Payer ID/ EDI ID", 
                         width = "100%")
              )
            )
          ),
          
          fluidRow(
            column(width = 6,
              div(
                style = "margin-bottom: 12px;",
                textInput(ns("address1"), 
                         label = "Address Line 1", 
                         width = "100%")
              )
            ),
            column(width = 6,
              div(
                style = "margin-bottom: 12px;",
                textInput(ns("address2"), 
                         label = "Address Line 2", 
                         width = "100%")
              )
            )
          ),
          
          fluidRow(
            column(width = 4,
              div(
                style = "margin-bottom: 12px;",
                textInput(ns("city"), 
                         label = "City", 
                         width = "100%")
              )
            ),
            column(width = 4,
              div(
                style = "margin-bottom: 12px;",
                textInput(ns("state"), 
                         label = "State", 
                         width = "100%")
              )
            ),
            column(width = 4,
              div(
                style = "margin-bottom: 12px;",
                textInput(ns("zipcode"), 
                         label = "Zipcode", 
                         width = "100%")
              )
            )
          )
        ),
        
        # Additional Organizations (Dynamic)
        uiOutput(ns("additional_orgs_ui")),
        
        # Add Additional Organization Button
        div(
          style = "text-align: center; margin: 20px 0;",
          actionButton(ns("add_organization"), 
                      "+ ADD ADDITIONAL ORGANIZATION",
                      style = "background-color: #4CAF50; color: white; border: none; padding: 8px 16px; border-radius: 4px; font-weight: bold; font-size: 14px;")
        )
      )
    ),
    
    # Contact Information and Submit Section
    fluidRow(
      column(width = 8, offset = 2,
        # Contact Information Section
        div(
          style = "background-color: #fff; padding: 20px; margin-bottom: 20px; border-radius: 5px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);",
          h4("Contact Information", style = "margin-bottom: 15px; color: #333; margin-top: 0;"),
          
          fluidRow(
            column(width = 6,
              div(
                style = "margin-bottom: 12px;",
                textInput(ns("contact_name"), 
                         label = "Contact name", 
                         width = "100%")
              )
            ),
            column(width = 6,
              div(
                style = "margin-bottom: 12px;",
                textInput(ns("contact_email"), 
                         label = "Contact email", 
                         width = "100%"),
                uiOutput(ns("contact_email_error_msg"))
              )
            )
          )
        ),
        
        # reCAPTCHA and Submit Section
        div(
          style = "background-color: #fff; padding: 20px; margin-bottom: 20px; border-radius: 5px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);",
          
          # reCAPTCHA placeholder
          div(
            style = "margin-bottom: 20px; padding: 20px; border: 1px solid #ccc; background-color: #f9f9f9; width: fit-content;",
            checkboxInput(ns("recaptcha_verified"), "I'm not a robot", value = FALSE),
            div(style = "font-size: 12px; color: #666; margin-top: 5px;", 
                "reCAPTCHA", br(), "Privacy - Terms")
          ),
          
          # Submit Button
          div(
            style = "text-align: left;",
            actionButton(ns("submit_registration"), 
                        "SUBMIT",
                        style = "background-color: #9c27b0; color: white; border: none; padding: 12px 40px; border-radius: 4px; font-weight: bold; font-size: 16px;")
          )
        ),
        
        # Success/Error Messages
        div(id = ns("message_area"), style = "margin-top: 20px;")
      )
    )
  )
}

payerregistrationmodule <- function(
  input,
  output,
  session
) {
  ns <- session$ns
  
  # Helper operator for null coalescing
  `%||%` <- function(a, b) if (is.null(a)) b else a
  
  # Reactive values to store additional organizations
  additional_orgs_count <- reactiveVal(0)
  
  # Render additional organization forms dynamically
  output$additional_orgs_ui <- renderUI({
    count <- additional_orgs_count()
    if (count == 0) return(NULL)
    
    org_forms <- lapply(1:count, function(i) {
      div(
        style = "background-color: #fff; padding: 12px; margin-bottom: 12px; border-radius: 4px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); border-left: 3px solid #4CAF50;",
        div(
          style = "display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px;",
          h4(paste("Additional Organization", i), style = "margin: 0; color: #333; font-size: 16px;"),
          actionButton(ns(paste0("remove_org_", i)), 
                      "Remove", 
                      style = "background-color: #f44336; color: white; border: none; padding: 3px 8px; border-radius: 3px; font-size: 11px;")
        ),
        
        # Row 1: Name and Payer ID/EDI ID
        fluidRow(
          column(width = 6,
            div(
              style = "margin-bottom: 10px;",
              textInput(ns(paste0("additional_org_name_", i)), 
                       label = "Name",
                       placeholder = "",
                       width = "100%")
            )
          ),
          column(width = 6,
            div(
              style = "margin-bottom: 10px;",
              textInput(ns(paste0("additional_payer_id_", i)), 
                       label = "Payer ID/ EDI ID",
                       placeholder = "",
                       width = "100%")
            )
          )
        ),
        
        # Row 2: Address Line 1 and Address Line 2
        fluidRow(
          column(width = 6,
            div(
              style = "margin-bottom: 10px;",
              textInput(ns(paste0("additional_address1_", i)), 
                       label = "Address Line 1",
                       placeholder = "",
                       width = "100%")
            )
          ),
          column(width = 6,
            div(
              style = "margin-bottom: 10px;",
              textInput(ns(paste0("additional_address2_", i)), 
                       label = "Address Line 2",
                       placeholder = "",
                       width = "100%")
            )
          )
        ),
        
        # Row 3: City and State
        fluidRow(
          column(width = 6,
            div(
              style = "margin-bottom: 10px;",
              textInput(ns(paste0("additional_city_", i)), 
                       label = "City",
                       placeholder = "",
                       width = "100%")
            )
          ),
          column(width = 6,
            div(
              style = "margin-bottom: 10px;",
              textInput(ns(paste0("additional_state_", i)), 
                       label = "State",
                       placeholder = "",
                       width = "100%")
            )
          )
        ),
        
        # Row 4: Zipcode (left side only)
        fluidRow(
          column(width = 6,
            div(
              style = "margin-bottom: 0;",
              textInput(ns(paste0("additional_zipcode_", i)), 
                       label = "Zipcode",
                       placeholder = "",
                       width = "100%")
            )
          )
        )
      )
    })
    
    do.call(tagList, org_forms)
  })
  
  # Validation function
  validate_form <- function() {
    errors <- list()
    
    # Check mandatory fields
    if (is.null(input$fhir_endpoint) || input$fhir_endpoint == "") {
      errors$fhir_endpoint <- "FHIR Endpoint is mandatory"
    }
    
    if (is.null(input$contact_email) || input$contact_email == "") {
      errors$contact_email <- "Contact Email is mandatory"
    }
    
    # Basic email validation
    if (!is.null(input$contact_email) && input$contact_email != "" && 
        !grepl("^[^@]+@[^@]+\\.[^@]+$", input$contact_email)) {
      errors$contact_email <- "Please enter a valid email address"
    }
    
    # URL validation for FHIR endpoint
    if (!is.null(input$fhir_endpoint) && input$fhir_endpoint != "" && 
        !grepl("^https?://", input$fhir_endpoint)) {
      errors$fhir_endpoint <- "Please enter a valid URL (starting with http:// or https://)"
    }
    
    return(errors)
  }
  
  # Show/hide validation errors using renderUI
  output$fhir_endpoint_error_msg <- renderUI({
    errors <- validate_form()
    if ("fhir_endpoint" %in% names(errors)) {
      div(style = "color: #d32f2f; font-size: 12px; margin-top: 5px;", errors$fhir_endpoint)
    }
  })
  
  output$contact_email_error_msg <- renderUI({
    errors <- validate_form()
    if ("contact_email" %in% names(errors)) {
      div(style = "color: #d32f2f; font-size: 12px; margin-top: 5px;", errors$contact_email)
    }
  })
  
  # Handle add additional organization
  observeEvent(input$add_organization, {
    current_count <- additional_orgs_count()
    additional_orgs_count(current_count + 1)
    
    showNotification("Additional organization form added!", 
                    type = "message", duration = 3)
  })
  
  # Handle remove organization buttons dynamically
  observe({
    count <- additional_orgs_count()
    if (count > 0) {
      lapply(1:count, function(i) {
        observeEvent(input[[paste0("remove_org_", i)]], {
          # Reduce count and trigger UI refresh
          current_count <- additional_orgs_count()
          if (current_count > 0) {
            additional_orgs_count(current_count - 1)
            showNotification("Organization removed!", 
                            type = "message", duration = 3)
          }
        }, ignoreInit = TRUE)
      })
    }
  })
  
  # Helper function to get additional organizations data
  get_additional_orgs_data <- function() {
    count <- additional_orgs_count()
    if (count == 0) return(list())
    
    additional_orgs <- list()
    for (i in 1:count) {
      # Only add if the organization name is not empty
      org_name <- input[[paste0("additional_org_name_", i)]]
      if (!is.null(org_name) && org_name != "") {
        additional_orgs[[length(additional_orgs) + 1]] <- list(
          id = i,
          name = org_name %||% "",
          payer_id = input[[paste0("additional_payer_id_", i)]] %||% "",
          address1 = input[[paste0("additional_address1_", i)]] %||% "",
          address2 = input[[paste0("additional_address2_", i)]] %||% "",
          city = input[[paste0("additional_city_", i)]] %||% "",
          state = input[[paste0("additional_state_", i)]] %||% "",
          zipcode = input[[paste0("additional_zipcode_", i)]] %||% ""
        )
      }
    }
    return(additional_orgs)
  }
  
  # Handle form submission
  observeEvent(input$submit_registration, {
    
    # Validate form
    errors <- validate_form()
    
    if (length(errors) > 0) {
      showNotification("Please fix the errors in the form before submitting.", 
                      type = "error", duration = 5)
      return()
    }
    
    # Check reCAPTCHA
    if (!input$recaptcha_verified) {
      showNotification("Please verify that you are not a robot.", 
                      type = "error", duration = 5)
      return()
    }
    
    # Collect form data
    form_data <- list(
      fhir_endpoint = input$fhir_endpoint,
      user_website = input$user_website,
      fhir_type = input$fhir_type,
      org_name = input$org_name,
      payer_id = input$payer_id,
      address1 = input$address1,
      address2 = input$address2,
      city = input$city,
      state = input$state,
      zipcode = input$zipcode,
      contact_name = input$contact_name,
      contact_email = input$contact_email,
      additional_organizations = get_additional_orgs_data(),
      submission_time = Sys.time()
    )
    
    # TODO: Here we would typically save the data to a database
    # For now, we'll just show a success message
    
    tryCatch({
      # Simulate form submission
      # save_payer_registration(form_data)
      
      # Clear form
      updateTextInput(session, "fhir_endpoint", value = "")
      updateTextInput(session, "user_website", value = "")
      updateSelectInput(session, "fhir_type", selected = "")
      updateTextInput(session, "org_name", value = "")
      updateTextInput(session, "payer_id", value = "")
      updateTextInput(session, "address1", value = "")
      updateTextInput(session, "address2", value = "")
      updateTextInput(session, "city", value = "")
      updateTextInput(session, "state", value = "")
      updateTextInput(session, "zipcode", value = "")
      updateTextInput(session, "contact_name", value = "")
      updateTextInput(session, "contact_email", value = "")
      updateCheckboxInput(session, "recaptcha_verified", value = FALSE)
      
      # Clear additional organizations
      additional_orgs_count(0)
      
      # Show success message
      output$message_area <- renderUI({
        div(
          style = "padding: 15px; background-color: #d4edda; border: 1px solid #c3e6cb; border-radius: 4px; color: #155724;",
          tags$strong("Success!"), " Your payer endpoint registration has been submitted successfully. 
          You will receive a confirmation email shortly."
        )
      })
      
    }, error = function(e) {
      # Show error message
      showNotification(paste("Error submitting registration:", e$message), 
                      type = "error", duration = 10)
    })
  })
}