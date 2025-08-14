library(shiny)
library(shinydashboard)
library(DBI)
library(RPostgres)
library(jsonlite)
library(httr)

payerregistrationmodule_UI <- function(id) {
  ns <- NS(id)

  tagList(
    # === Styles for errors (dark/bold red just on the input) ===
    tags$style(HTML("
      input.input-error {
        border: 2px solid #b71c1c !important;
        box-shadow: 0 0 4px rgba(183,28,28,0.8);
      }
      .error-message {
        color: #b71c1c;
        font-size: 12px;
        margin-top: 3px;
      }
    ")),

    # JS handler: view switcher
    tags$script(HTML(sprintf("
      Shiny.addCustomMessageHandler('showRegistrationView', function(x){
        var form = document.getElementById(x.formId);
        var success = document.getElementById(x.successId);
        if (!form || !success) return;

        if (x.show === 'success') {
          form.style.display = 'none';
          success.style.display = '';
        } else {
          form.style.display = '';
          success.style.display = 'none';

          // Reset any existing v2 checkbox widget(s) and clear Shiny token
          try { if (window.grecaptcha && grecaptcha.reset) { grecaptcha.reset(); } } catch(e) {}
          Shiny.setInputValue('%s', null, {priority: 'event', nonce: Math.random()});
        }
      });
    ", ns("recaptcha_token")))),

    # === Tiny JS helper: toggle 'input-error' on an input by id ===
    tags$script(HTML("
      Shiny.addCustomMessageHandler('toggleInputError', function(x) {
        var el = document.getElementById(x.id);
        if (!el) return;
        if (x.add) { el.classList.add('input-error'); }
        else { el.classList.remove('input-error'); }
      });
    ")),

    # Info button
    fluidRow(
      column(width = 12,
        actionButton(
          ns("pr_show_info"),
          "Info",
          icon = tags$i(class = "fa fa-question-circle", "aria-hidden" = "true",
                        role = "presentation", "aria-label" = "question-circle icon")
        )
      )
    ),

    # Top note banner
    fluidRow(
      column(width = 8,
        div(
          style = "margin-bottom: 15px; padding: 7px; background-color: #e3f2fd; border-left: 4px solid #2196f3;",
          p(
            strong("Note:"),
            "Alternatively, you can register FHIR endpoint(s) in FHIR bundle or CSV format by sending an email to",
            tags$a("Lantern Support.",
              href = "mailto:lantern.support@example.com",
              class = "lantern-url",
              style = "color: #1976d2;"
            ),
            style = "margin: 0; color: #1565c0;"
          )
        )
      )
    ),

    div(id = ns("form_container"),
      # Main column layout
      fluidRow(
        column(width = 8,

          # --- FHIR Endpoint Section ---
          div(
            style = "background-color: #fff; padding: 20px; margin-bottom: 20px; border-radius: 5px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);",

            # FHIR Endpoint URL (static input; error below)
            div(
              style = "margin-bottom: 12px;",
              textInput(
                ns("fhir_endpoint"),
                label = "FHIR Endpoint",
                placeholder = "Enter FHIR endpoint URL",
                width = "100%"
              ),
              uiOutput(ns("fhir_endpoint_error_msg"))
            ),

            # User-facing website
            div(
              style = "margin-bottom: 12px;",
              textInput(
                ns("user_website"),
                label = "User-facing website for the endpoint",
                placeholder = "Enter website URL",
                width = "100%"
              )
            ),

            # Type of FHIR Endpoint
            div(
              style = "margin-bottom: 0;",
              selectInput(
                ns("fhir_type"),
                label = "Type of FHIR Endpoint",
                choices = list(
                  "Select" = "",
                  "Payer to payer API" = "payer_to_payer",
                  "Provider Access API" = "provider_access",
                  "Patient Access API" = "patient_access",
                  "Prior Authorization API" = "prior_authorization"
                ),
                width = "100%"
              )
            )
          ),

          # --- Organization Section ---
          div(
            style = "background-color: #fff; padding: 20px; margin-bottom: 20px; border-radius: 5px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);",
            h4("Organization", style = "margin-bottom: 15px; color: #333; margin-top: 0;"),

            fluidRow(
              column(width = 6,
                div(
                  style = "margin-bottom: 12px;",
                  textInput(ns("org_name"), label = "Name", width = "100%")
                )
              ),
              column(width = 6,
                div(
                  style = "margin-bottom: 12px;",
                  textInput(ns("payer_id"), label = "Payer ID/ EDI ID", width = "100%")
                )
              )
            ),

            fluidRow(
              column(width = 6,
                div(
                  style = "margin-bottom: 12px;",
                  textInput(ns("address1"), label = "Address Line 1", width = "100%")
                )
              ),
              column(width = 6,
                div(
                  style = "margin-bottom: 12px;",
                  textInput(ns("address2"), label = "Address Line 2", width = "100%")
                )
              )
            ),

            fluidRow(
              column(width = 4,
                div(
                  style = "margin-bottom: 12px;",
                  textInput(ns("city"), label = "City", width = "100%")
                )
              ),
              column(width = 4,
                div(
                  style = "margin-bottom: 12px;",
                  textInput(ns("state"), label = "State", width = "100%")
                )
              ),
              column(width = 4,
                div(
                  style = "margin-bottom: 12px;",
                  textInput(ns("zipcode"), label = "Zipcode", width = "100%")
                )
              )
            )
          ),

          # Additional Organizations (Dynamic)
          uiOutput(ns("additional_orgs_ui")),

          # Add Additional Organization Button
          div(
            style = "text-align: center; margin: 20px 0;",
            actionButton(
              ns("add_organization"),
              "+ ADD ADDITIONAL ORGANIZATION",
              style = "background-color: #4CAF50; color: white; border: none; padding: 8px 16px; border-radius: 4px; font-weight: bold; font-size: 14px;"
            )
          )
        )
      ),

      # Contact Information and Submit Section
      fluidRow(
        column(width = 8,

          # Contact Information Section
          div(
            style = "background-color: #fff; padding: 20px; margin-bottom: 20px; border-radius: 5px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);",
            h4("Contact Information", style = "margin-bottom: 15px; color: #333; margin-top: 0;"),

            fluidRow(
              column(width = 6,
                div(
                  style = "margin-bottom: 12px;",
                  textInput(ns("contact_name"), label = "Contact name", width = "100%")
                )
              ),
              column(width = 6,
                div(
                  style = "margin-bottom: 12px;",
                  textInput(ns("contact_email"), label = "Contact email", width = "100%"),
                  uiOutput(ns("contact_email_error_msg"))
                )
              )
            )
          ),

          # --- reCAPTCHA  ---
          tags$div(
            class = "g-recaptcha",
            `data-sitekey` = Sys.getenv("LANTERN_RECAPTCHA_SITE_KEY"),
            `data-callback` = "onRecaptcha",
            `data-expired-callback` = "onRecaptchaExpired",
            `data-error-callback` = "onRecaptchaError"
          ),
          tags$script(src = "https://www.google.com/recaptcha/api.js"),

          # Success/expired/error callbacks
          tags$script(HTML(sprintf("
            function onRecaptcha(token) {
              Shiny.setInputValue('%s', token, {priority: 'event'});
            }
            function onRecaptchaExpired() {
              Shiny.setInputValue('%s', null, {priority: 'event', nonce: Math.random()});
            }
            function onRecaptchaError() {
              Shiny.setInputValue('%s', null, {priority: 'event', nonce: Math.random()});
            }
          ", ns("recaptcha_token"), ns("recaptcha_token"), ns("recaptcha_token")))),

          # Just-in-time sync on form submit
          tags$script(HTML(sprintf("
          $(document).on('click', '#%s', function() {
            try {
              var t = (window.grecaptcha && grecaptcha.getResponse) ? grecaptcha.getResponse() : null;
              Shiny.setInputValue('%s', t || null, {priority: 'event', nonce: Math.random()});
            } catch(e) { /* ignore */ }
          });
        ", ns("submit_registration"), ns("recaptcha_token")))),

          # 6) Server-triggered reset: reset THIS widget + clear Shiny input
          tags$script(HTML(sprintf("
            Shiny.addCustomMessageHandler('resetRecaptcha', function(){
              try { if (window.grecaptcha) grecaptcha.reset(); } catch(e) {}
              Shiny.setInputValue('%s', null, {priority: 'event', nonce: Math.random()});
            });
          ", ns("recaptcha_token")))),

          # Submit button
          div(style = "margin-top: 16px;",
            actionButton(
              ns("submit_registration"),
              "SUBMIT",
              style = "background-color: #9c27b0; color: white; border: none; padding: 12px 40px; border-radius: 4px; font-weight: bold; font-size: 16px;"
            )
          ),

          # Success/Error Messages
          div(id = ns("message_area"), style = "margin-top: 20px;")
        )
      )
    ),

    div(
      id = ns("success_container"),
      style = "display:none; padding: 20px;",
      h2("Payer endpoint Self Registration"),
      div(style="margin-top: 10px; color: #2e7d32; font-weight: 600;",
          "Thank you for registering your FHIR endpoint. We will contact you if additional information is required."
      ),
      div(style="margin-top: 24px;",
          actionButton(
            ns("new_registration"),
            "+ REGISTER NEW FHIR ENDPOINT",
            style = "background-color:#2962ff; color:#fff; border:none; padding:10px 16px; border-radius:4px; font-weight:600;"
          )
      )
    )
  )
}

payerregistrationmodule <- function(input, output, session) {
  ns <- session$ns

  # ---- Logging helper ----
  log_submission <- function(event, details = "") {
    stamp <- format(Sys.time(), "%Y-%m-%d %H:%M:%S %Z")
    message(sprintf("[PAYER_REG][%s][session=%s][event=%s] %s",
                    stamp, session$token, event, details))
  }

  # Helper operator for null coalescing
  `%||%` <- function(a, b) if (is.null(a)) b else a

  # ===================== DATABASE FUNCTIONS =====================
  
  # Function to save payer registration data to the database
  save_payer_registration <- function(form_data) {
    tryCatch({
      # Begin transaction
      dbBegin(db_connection)
      
      # 1. Insert into payers table
      payer_query <- "
        INSERT INTO payers (contact_name, contact_email, submission_time)
        VALUES ($1, $2, $3)
        RETURNING id
      "
      
      payer_result <- dbGetQuery(
        db_connection,
        payer_query,
        params = list(
          form_data$contact_name %||% "",
          form_data$contact_email,
          form_data$submission_time
        )
      )
      
      payer_id <- payer_result$id[1]
      
      # 2. Insert main organization into payer_endpoints
      main_org_address <- list(
        address1 = form_data$address1 %||% "",
        address2 = form_data$address2 %||% "",
        city = form_data$city %||% "",
        state = form_data$state %||% "",
        zipcode = form_data$zipcode %||% ""
      )
      
      endpoint_query <- "
        INSERT INTO payer_endpoints (
          payer_id, url, name, edi_id, address, user_facing_url
        )
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id
      "
      
      # Convert payer_id to integer, handle empty strings
      edi_id_value <- if(is.null(form_data$payer_id) || form_data$payer_id == "") {
        NA_integer_
      } else {
        tryCatch(as.integer(form_data$payer_id), error = function(e) NA_integer_)
      }
      
      main_endpoint_result <- dbGetQuery(
        db_connection,
        endpoint_query,
        params = list(
          payer_id,
          form_data$fhir_endpoint,
          form_data$org_name %||% "",
          edi_id_value,
          toJSON(main_org_address, auto_unbox = TRUE),
          form_data$user_website %||% ""
        )
      )
      
      main_endpoint_id <- main_endpoint_result$id[1]
      
      # 3. Insert additional organizations
      if (length(form_data$additional_organizations) > 0) {
        for (org in form_data$additional_organizations) {
          if (!is.null(org$name) && org$name != "") {
            add_org_address <- list(
              address1 = org$address1 %||% "",
              address2 = org$address2 %||% "",
              city = org$city %||% "",
              state = org$state %||% "",
              zipcode = org$zipcode %||% ""
            )
            
            # Convert additional org payer_id to integer
            add_edi_id_value <- if(is.null(org$payer_id) || org$payer_id == "") {
              NA_integer_
            } else {
              tryCatch(as.integer(org$payer_id), error = function(e) NA_integer_)
            }
            
            dbGetQuery(
              db_connection,
              endpoint_query,
              params = list(
                payer_id,
                form_data$fhir_endpoint, # Same endpoint URL
                org$name,
                add_edi_id_value,
                toJSON(add_org_address, auto_unbox = TRUE),
                form_data$user_website %||% ""
              )
            )
          }
        }
      }
      
      # Note: payer_info table will be populated later in the population process
      
      # Commit transaction
      dbCommit(db_connection)
      
      return(list(
        success = TRUE,
        payer_id = payer_id,
        main_endpoint_id = main_endpoint_id,
        message = "Payer registration saved successfully"
      ))
      
    }, error = function(e) {
      # Rollback transaction on error
      dbRollback(db_connection)
      
      return(list(
        success = FALSE,
        error = as.character(e),
        message = paste("Error saving payer registration:", e$message)
      ))
    })
  }

  # Function to check if email already exists
  check_email_exists <- function(email) {
    tryCatch({
      query <- "SELECT COUNT(*) as count FROM payers WHERE LOWER(contact_email) = LOWER($1)"
      result <- dbGetQuery(db_connection, query, params = list(email))
      return(result$count[1] > 0)
    }, error = function(e) {
      log_submission("db_error", paste("Error checking email:", e$message))
      return(FALSE) # Assume email doesn't exist if there's an error
    })
  }

  # ===================== VALIDATION =====================
  
  validate_form <- function() {
    errors <- list()

    # Mandatory fields
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

    # Check for duplicate email in database
    if (!is.null(input$contact_email) && input$contact_email != "" && 
        length(errors) == 0) { # Only check if no other email errors
      if (check_email_exists(input$contact_email)) {
        errors$contact_email <- "This email address has already been registered"
      }
    }

    errors
  }

  # Error messages (inline, under inputs)
  output$fhir_endpoint_error_msg <- renderUI({
    errs <- validate_form()
    if ("fhir_endpoint" %in% names(errs)) {
      div(class = "error-message", errs$fhir_endpoint)
    }
  })
  
  output$contact_email_error_msg <- renderUI({
    errs <- validate_form()
    if ("contact_email" %in% names(errs)) {
      div(class = "error-message", errs$contact_email)
    }
  })

  # Toggle red border class on actual input elements (no re-render)
  observe({
    errs <- validate_form()

    session$sendCustomMessage('toggleInputError', list(
      id  = ns("fhir_endpoint"),
      add = "fhir_endpoint" %in% names(errs)
    ))
    session$sendCustomMessage('toggleInputError', list(
      id  = ns("contact_email"),
      add = "contact_email" %in% names(errs)
    ))
  })

  # ===================== ADDITIONAL ORGANIZATIONS =====================
  
  additional_orgs_count <- reactiveVal(0)

  output$additional_orgs_ui <- renderUI({
    count <- additional_orgs_count()
    if (count == 0) return(NULL)

    org_forms <- lapply(1:count, function(i) {
      div(
        style = "background-color: #fff; padding: 12px; margin-bottom: 12px; border-radius: 4px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); border-left: 3px solid #4CAF50;",
        div(
          style = "display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px;",
          h4(paste("Additional Organization", i), style = "margin: 0; color: #333; font-size: 16px;"),
          actionButton(ns(paste0("remove_org_", i)), "Remove",
                       style = "background-color: #f44336; color: white; border: none; padding: 3px 8px; border-radius: 3px; font-size: 11px;")
        ),

        # Row 1
        fluidRow(
          column(width = 6,
            div(style = "margin-bottom: 10px;",
                textInput(ns(paste0("additional_org_name_", i)), label = "Name", width = "100%"))
          ),
          column(width = 6,
            div(style = "margin-bottom: 10px;",
                textInput(ns(paste0("additional_payer_id_", i)), label = "Payer ID/ EDI ID", width = "100%"))
          )
        ),

        # Row 2
        fluidRow(
          column(width = 6,
            div(style = "margin-bottom: 10px;",
                textInput(ns(paste0("additional_address1_", i)), label = "Address Line 1", width = "100%"))
          ),
          column(width = 6,
            div(style = "margin-bottom: 10px;",
                textInput(ns(paste0("additional_address2_", i)), label = "Address Line 2", width = "100%"))
          )
        ),

        # Row 3
        fluidRow(
          column(width = 6,
            div(style = "margin-bottom: 10px;",
                textInput(ns(paste0("additional_city_", i)), label = "City", width = "100%"))
          ),
          column(width = 6,
            div(style = "margin-bottom: 10px;",
                textInput(ns(paste0("additional_state_", i)), label = "State", width = "100%"))
          )
        ),

        # Row 4
        fluidRow(
          column(width = 6,
            div(style = "margin-bottom: 0;",
                textInput(ns(paste0("additional_zipcode_", i)), label = "Zipcode", width = "100%"))
          )
        )
      )
    })

    do.call(tagList, org_forms)
  })

  observeEvent(input$add_organization, {
    additional_orgs_count(additional_orgs_count() + 1)
    showNotification("Additional organization form added!", type = "message", duration = 3)
  })

  observe({
    count <- additional_orgs_count()
    if (count > 0) {
      lapply(1:count, function(i) {
        observeEvent(input[[paste0("remove_org_", i)]], {
          if (additional_orgs_count() > 0) {
            additional_orgs_count(additional_orgs_count() - 1)
            showNotification("Organization removed!", type = "message", duration = 3)
          }
        }, ignoreInit = TRUE)
      })
    }
  })

  # Collect dynamic orgs
  get_additional_orgs_data <- function() {
    count <- additional_orgs_count()
    if (count == 0) return(list())

    additional_orgs <- list()
    for (i in 1:count) {
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
    additional_orgs
  }

  # ===================== FORM SUBMISSION =====================
  
  observeEvent(input$submit_registration, {
    errors <- c()

    # Validate form
    errors <- validate_form()

    if (length(errors) > 0) {
      log_submission("submit_invalid", paste(errors, collapse = "; "))
      showNotification("Please fix the errors in the form before submitting.",
                       type = "error", duration = 5)
      return()
    }
    
    # Obtain reCAPTCHA token
    token <- input$recaptcha_token # set by JS callback
    if (is.null(token) || identical(token, "")) {
      log_submission("captcha_missing", "No reCAPTCHA token present")
      showNotification("Please complete the CAPTCHA.", type = "error", duration = 5)
      return()
    }

    res <- httr::POST(
      url = "https://www.google.com/recaptcha/api/siteverify",
      body = list(
        secret = Sys.getenv("LANTERN_RECAPTCHA_SECRET"),
        response = token
      ),
      encode = "form"
    )

    # Verify reCAPTCHA
    verify <- httr::content(res)
    if (!isTRUE(verify$success)) {
      err <- tryCatch(paste(verify[["error-codes"]], collapse = ","), error = function(e) "unknown")
      log_submission("captcha_fail",
                      paste("Error:", err,
                      "| FHIR Endpoint:", input$fhir_endpoint,
                      "| Contact Email:", input$contact_email))
      showNotification("Please verify that you are not a robot.", type = "error", duration = 5)
      session$sendCustomMessage('resetRecaptcha', list())
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
    
    # Save to database
    tryCatch({
      # Save the payer registration data
      save_result <- save_payer_registration(form_data)
      
      if (save_result$success) {
        # --- SUCCESS LOG ---
        log_submission(
          "submit_ok",
          paste(
            "FHIR Endpoint:", input$fhir_endpoint,
            "| Contact Email:", input$contact_email,
            "| Endpoint Type:", input$fhir_type %||% "",
            "| Org Name:", input$org_name %||% "",
            "| Extra Orgs:", length(get_additional_orgs_data()),
            "| Payer ID:", save_result$payer_id,
            "| Endpoint ID:", save_result$main_endpoint_id
          )
        )

        # Show the success screen
        session$sendCustomMessage(
          'showRegistrationView',
          list(formId = ns("form_container"),
              successId = ns("success_container"),
              show = "success")
        )

        # Reset Captcha
        session$sendCustomMessage('resetRecaptcha', list())
        
      } else {
        # Database save failed
        log_submission("db_save_fail", paste("Database error:", save_result$error))
        showNotification(paste("Error saving registration:", save_result$message), 
                        type = "error", duration = 10)
        session$sendCustomMessage('resetRecaptcha', list())
      }
      
    }, error = function(e) {
      # Show error message
      log_submission("submit_error", paste("Unexpected error:", e$message))
      showNotification(paste("Error submitting registration:", e$message), 
                      type = "error", duration = 10)
      session$sendCustomMessage('resetRecaptcha', list())
    })
  })

  # ===================== FORM RESET =====================
  
  observeEvent(input$new_registration, {
    # Clear form fields
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
    # Reset Captcha
    session$sendCustomMessage('resetRecaptcha', list())
    # Reset dynamic orgs
    additional_orgs_count(0)

    # Hide success, show form
    session$sendCustomMessage(
      'showRegistrationView',
      list(formId = ns("form_container"),
          successId = ns("success_container"),
          show = "form")
    )
  })

  # ===================== INFO MODAL =====================
  
  observeEvent(input$pr_show_info, {
    showModal(modalDialog(
      title = "Payer Registration – Information",
      easyClose = TRUE,
      size = "l",
      p(HTML("
        <b>What is this?</b><br>
        This section will guide payers through the self-registration process.<br><br>
        <b>How to use:</b><br>
        1) Enter your FHIR endpoint URL and contact email (required).<br>
        2) Optionally add organization details and additional organizations.<br>
        3) Complete the CAPTCHA and submit.<br><br>
        <i>Your registration will be reviewed and validated before being added to our system.</i>
      "))
    ))
  })
}