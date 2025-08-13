library(shiny)
library(shinydashboard)

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
    tags$script(HTML("
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
        }
      });
    ")),

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

          # reCAPTCHA and Submit Section
          div(
            style = "background-color: #fff; padding: 20px; margin-bottom: 20px; border-radius: 5px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);",

            # reCAPTCHA placeholder
            div(
              style = "margin-bottom: 20px; padding: 20px; border: 1px solid #ccc; background-color: #f9f9f9; width: fit-content;",
              checkboxInput(ns("recaptcha_verified"), "I'm not a robot", value = FALSE),
              div(style = "font-size: 12px; color: #666; margin-top: 5px;", "reCAPTCHA", br(), "Privacy - Terms")
            ),

            # Submit Button
            div(
              style = "text-align: left;",
              actionButton(
                ns("submit_registration"),
                "SUBMIT",
                style = "background-color: #9c27b0; color: white; border: none; padding: 12px 40px; border-radius: 4px; font-weight: bold; font-size: 16px;"
              )
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

  # ----------------- Validation -----------------
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

  # ----------------- Additional Orgs -----------------
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

  # ----------------- Submit -----------------
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
    
    # Check reCAPTCHA
    if (!isTRUE(input$recaptcha_verified)) {
      log_submission("captcha_fail",
                     paste("FHIR Endpoint:", input$fhir_endpoint,
                           "Contact Email:", input$contact_email))
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
      # --- SUCCESS LOG (before clearing form) ---
      log_submission(
        "submit_ok",
        paste(
          "FHIR Endpoint:", input$fhir_endpoint,
          "| Contact Email:", input$contact_email,
          "| Endpoint Type:", input$fhir_type %||% "",
          "| Org Name:", input$org_name %||% "",
          "| Extra Orgs:", length(get_additional_orgs_data())
        )
      )

      # Show the success screen
      session$sendCustomMessage(
        'showRegistrationView',
        list(formId = ns("form_container"),
            successId = ns("success_container"),
            show = "success")
      )

      # Simulate form submission
      # save_payer_registration(form_data)
      
    }, error = function(e) {
      # Show error message
      showNotification(paste("Error submitting registration:", e$message), 
                      type = "error", duration = 10)
    })
  })

  # ----------------- New Endpoint Button -----------------
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
    updateCheckboxInput(session, "recaptcha_verified", value = FALSE)
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

  # ----------------- Info Modal -----------------
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
      <i>(Replace this placeholder with final copy.)</i>
    "))
  ))
})
}