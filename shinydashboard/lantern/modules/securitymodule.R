# Security Module - Performance Optimized while maintaining exact data accuracy

log_duration <- function(label, expr) {
  start_time <- Sys.time()
  result <- expr
  duration <- Sys.time() - start_time
  result
}

auth_color_map <- data.frame(
  code = c("OAuth", "SMART-on-FHIR", "Basic", "Certificates", "UDAP", "None"),
  bg_color = c("#e6f0ff", "#e9f8ee", "#fde8e8", "#f3e8fd", "#e5e7eb", "#e5e7eb"),
  text_color = c("#1a4ed8", "#2b9348", "#b02a2a", "#6b21a8", "#374151", "#374151"),
  stringsAsFactors = FALSE
)

securitymodule_UI <- function(id) {
  ns <- NS(id)

  tagList(
    # CUSTOM CSS to match dashboard visual system
    tags$style(HTML("
          .plot-card {
            background: #ffffff;
            padding: 16px;
            border-radius: 10px;
            box-shadow: 0 2px 6px rgba(0,0,0,0.08);
            transition: transform 0.18s ease, box-shadow 0.18s ease;
            margin-bottom: 25px;
          }

          .section-title {
            font-size: 1.4rem;
            font-weight: 600;
            margin: 25px 0 15px;
          }

          .modern-security-table {
            margin-top: 12px;
            border-radius: 6px;
            overflow: hidden;
            box-shadow: 0 1px 3px rgba(0,0,0,0.05);
          }

          #security_search_query {
            border: 2px solid #e1e4e8;
            border-radius: 8px;
            padding: 10px 16px;
            font-size: 14px;
            transition: border-color 0.2s ease, box-shadow 0.2s ease;
            width: 100%;
          }

          #security_search_query:focus {
            border-color: #667eea;
            box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
            outline: none;
          }

          .pagination-wrapper {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-top: 20px;
          }

          .page-selector-wrapper {
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 10px;
          }

          .modern-nav-button {
            background: linear-gradient(90deg, #667eea, #764ba2);
            color: white;
            font-weight: 600;
            padding: 10px 20px;
            border-radius: 8px;
            border: none;
            transition: background 0.2s ease;
          }

          .modern-nav-button:hover {
            background: linear-gradient(90deg, #5a67d8, #6b46c1);
            cursor: pointer;
          }

          .version-badge {
            display: inline-block;
            background: #e6f0ff;
            color: #1a4ed8;
            padding: 4px 10px;
            border-radius: 12px;
            font-size: 0.8rem;
            font-weight: 600;
          }

          .auth-badge {
            padding: 4px 10px;
            border-radius: 12px;
            font-size: 0.8rem;
            font-weight: 600;
            display: inline-block;
            white-space: nowrap;
          }

          .auth-badge-oauth {
            background: #e6f0ff;
            color: #1a4ed8;
          }

          .auth-badge-smart {
            background: #e9f8ee;
            color: #2b9348;
          }

          .auth-badge-basic {
            background: #fde8e8;
            color: #b02a2a;
          }

          .auth-badge-cert {
            background: #f3e8fd;
            color: #6b21a8;
          }

          .auth-badge-default {
            background: #e5e7eb;
            color: #374151;
          }

          .auth-card {
            background: #ffffff;
            padding: 16px;
            border-radius: 10px;
            box-shadow: 0 2px 6px rgba(0,0,0,0.08);
            transition: transform 0.18s ease, box-shadow 0.18s ease;
            margin-bottom: 20px;
            text-align: left;
          }

          .auth-icon {
            font-size: 24px;
            margin-bottom: 6px;
          }

          .auth-percent {
            font-size: 1.5rem;
            font-weight: 700;
            margin-bottom: 4px;
          }

          .auth-method {
            font-size: 1.1rem;
            font-weight: 600;
          }

          .auth-count {
            font-size: 0.95rem;
            color: #555;
            margin-bottom: 4px;
          }
        ")),

    uiOutput(ns("auth_summary_cards_ui")),

    # Summary Tables
    fluidRow(
      column(6,
        div(class = "plot-card security-table-section",
            h4("Endpoint Summary"),
            tableOutput(ns("endpoint_summary_table"))
        )
      ),
      column(6,
        div(class = "plot-card security-table-section",
            h4("Authorization Type Counts"),
            reactable::reactableOutput(ns("auth_type_count_table"))
        )
      )
    ),

    div(class = "section-title", "Endpoints by Authorization Type"),

    # Filter + Search block
    div(class = "plot-card",
      uiOutput("show_security_filter"),
      fluidRow(
        column(6,
          textInput(ns("security_search_query"), NULL,
                    placeholder = "🔍 Search by URL, Organization, Developer, TLS, etc.")
        )
      )),

      div(class = "modern-security-table",
        reactable::reactableOutput(ns("security_endpoints"))
      ),

      # Pagination block
      div(class = "pagination-wrapper",
        column(3,
          actionButton(ns("security_prev_page"),
                       label = tagList(tags$i(class = "fa fa-arrow-left", style = "margin-right: 8px;"), "Previous"),
                       class = "modern-nav-button")
        ),
        column(6,
          div(class = "page-selector-wrapper",
            numericInput(ns("security_page_selector"),
                         label = NULL,
                         value = 1, min = 1, max = 1,
                         step = 1, width = "80px"),
            textOutput(ns("current_security_page_info"), inline = TRUE)
          )
        ),
        column(3,
          actionButton(ns("security_next_page"),
                       label = tagList("Next", tags$i(class = "fa fa-arrow-right", style = "margin-left: 8px;")),
                       class = "modern-nav-button")
        )
      )
    )
}

# TODO:
# Consider conditional styling of status badges based on auth method strength

securitymodule <- function(
  input,
  output,
  session,
  sel_fhir_version,
  sel_vendor,
  sel_auth_type_code
) {
  ns <- session$ns
  security_page_size <- 10
  security_page_state <- reactiveVal(1)
  current_request_id <- reactiveVal(0)

  observe({
    new_page <- security_page_state()
    current_selector <- input$security_page_selector
    if (is.null(current_selector) || is.na(current_selector) || !is.numeric(current_selector) || current_selector != new_page) {
      isolate({ updateNumericInput(session, "security_page_selector", max = security_total_pages(), value = new_page) })
    }
  })

  observeEvent(input$security_page_selector, {
    current_input <- input$security_page_selector
    if (!is.null(current_input) && !is.na(current_input) && is.numeric(current_input) && current_input > 0) {
      new_page <- max(1, min(current_input, security_total_pages()))
      if (new_page != security_page_state()) security_page_state(new_page)
      if (new_page != current_input) updateNumericInput(session, "security_page_selector", value = new_page)
    } else {
      invalidateLater(100)
      updateNumericInput(session, "security_page_selector", value = security_page_state())
    }
  }, ignoreInit = TRUE)

  observeEvent(input$security_next_page, {
    if (security_page_state() < security_total_pages()) security_page_state(security_page_state() + 1)
  })

  observeEvent(input$security_prev_page, {
    if (security_page_state() > 1) security_page_state(security_page_state() - 1)
  })

  output$security_prev_button_ui <- renderUI({ if (security_page_state() > 1) actionButton(ns("security_prev_page"), "Previous", icon = icon("arrow-left")) })
  output$security_next_button_ui <- renderUI({ if (security_page_state() < security_total_pages()) actionButton(ns("security_next_page"), "Next", icon = icon("arrow-right")) })
  output$current_security_page_info <- renderText({ paste("of", security_total_pages()) })

  observeEvent(list(sel_fhir_version(), sel_vendor(), sel_auth_type_code(), input$security_search_query), {
    security_page_state(1)
  })

  output$auth_type_count_table <- reactable::renderReactable({
    df <- isolate(get_auth_type_count(db_connection)) %>% select(-Percent)

    # --- Compute unified percentages over the total ---
    total_endpoints <- sum(df$Endpoints, na.rm = TRUE)
    df <- df %>%
      dplyr::mutate(
        Percent = round((Endpoints / total_endpoints) * 100, 1)
      )

    # --- Add styled badges for Code and FHIR Version ---
    df[["Code"]] <- vapply(df[["Code"]], function(code) {
      class <- dplyr::case_when(
        code == "OAuth" ~ "auth-badge-oauth",
        code == "SMART-on-FHIR" ~ "auth-badge-smart",
        code == "Basic" ~ "auth-badge-basic",
        code == "Certificates" ~ "auth-badge-cert",
        code == "UDAP" ~ "auth-badge-udap",
        TRUE ~ "auth-badge-default"
      )
      as.character(htmltools::span(class = paste("auth-badge", class), code))
    }, character(1))

    df[["FHIR Version"]] <- vapply(df[["FHIR Version"]], function(version) {
      as.character(htmltools::span(class = "version-badge", version))
    }, character(1))

    reactable::reactable(
      df,
      columns = list(
        `Code` = colDef(name = "Authorization Type", html = TRUE),
        `FHIR Version` = colDef(name = "FHIR Version", html = TRUE),
        `Endpoints` = colDef(name = "Endpoint Count", align = "right"),
        `Percent` = colDef(name = "Percentage", align = "right", 
                          cell = function(value) paste0(value, "%"))
      ),
      highlight = TRUE,
      striped = TRUE,
      bordered = FALSE,
      showSortIcon = TRUE,
      defaultPageSize = 10
    )
  })

  output$endpoint_summary_table <- renderTable(
    isolate(get_endpoint_security_counts(db_connection))
  )

  security_base_sql <- reactive({
    log_duration("security_base_sql", {
      req(sel_fhir_version(), sel_vendor(), sel_auth_type_code())
      versions <- paste0("'", sel_fhir_version(), "'", collapse = ", ")
      vendor_filter <- if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) paste0("AND vendor_name = '", sel_vendor(), "'") else ""
      search_filter <- if (!is.null(input$security_search_query) && input$security_search_query != "") {
        q <- gsub("'", "''", input$security_search_query)
        paste0("AND (url ILIKE '%", q, "%' OR condensed_organization_names ILIKE '%", q, "%' OR vendor_name ILIKE '%", q, "%' OR capability_fhir_version ILIKE '%", q, "%' OR tls_version ILIKE '%", q, "%')")
      } else ""
      paste0("FROM security_endpoints_distinct_mv WHERE capability_fhir_version IN (", versions, ") AND code = '", sel_auth_type_code(), "' ", vendor_filter, " ", search_filter)
    })
  })

  security_total_pages <- reactive({
    log_duration("security_total_pages", {
      count_query <- paste0("SELECT COUNT(*) as count ", security_base_sql())
      count <- tbl(db_connection, sql(count_query)) %>% collect() %>% pull(count)
      max(1, ceiling(count / security_page_size))
    })
  })

  selected_endpoints <- reactive({
    req(sel_fhir_version(), sel_vendor(), sel_auth_type_code())
    request_id <- isolate(current_request_id()) + 1
    current_request_id(request_id)
    log_duration("selected_endpoints", {
      limit <- security_page_size
      offset <- (security_page_state() - 1) * security_page_size
      query <- paste0("SELECT * ", security_base_sql(), " ORDER BY url LIMIT ", limit, " OFFSET ", offset)
      result <- tbl(db_connection, sql(query)) %>% collect()
      if (request_id == isolate(current_request_id())) result else data.frame()
    })
  })

  auth_summary_cards_df <- reactive({
    raw <- get_auth_type_count(db_connection)

    # Aggregate endpoints by Code
    summarized_auth_card_data <- raw %>%
      dplyr::group_by(Code) %>%
      dplyr::summarise(Endpoints = sum(Endpoints, na.rm = TRUE), .groups = "drop")

    # Compute percentages after ungrouping
    total_endpoints <- sum(summarized_auth_card_data$Endpoints, na.rm = TRUE)

    summarized_auth_card_data <- summarized_auth_card_data %>%
      dplyr::mutate(
        Percent = round((Endpoints / total_endpoints) * 100, 1)
      )

    # Add missing "None" row with 0 values if needed
    all_types <- dplyr::tibble(code = auth_color_map$code)
    merged <- all_types %>%
      dplyr::left_join(summarized_auth_card_data, by = c("code" = "Code")) %>%
      dplyr::mutate(
        Endpoints = ifelse(is.na(Endpoints), 0, Endpoints),
        Percent = ifelse(is.na(Percent), 0, Percent),
        method = code
      )

    # Merge in color map
    final <- merged %>%
      dplyr::left_join(auth_color_map, by = "code") %>%
      dplyr::select(method, count = Endpoints, percent = Percent, bg_color, text_color)

    final
  })

  output$auth_summary_cards_ui <- renderUI({
    df <- auth_summary_cards_df()
    req(nrow(df) > 0)

    fluidRow(
      lapply(1:nrow(df), function(i) {
        card <- df[i, ]
        column(4, div(class = "auth-card",
          style = paste0("background:", card$bg_color, ";"),
          div(class = "auth-method", style = paste0("color:", card$text_color, ";"), card$method),
          div(class = "auth-count", style = paste0("color:", card$text_color, ";"), paste0(format(card$count, big.mark = ","), " endpoints")),
          div(class = "auth-percent", style = paste0("color:", card$text_color, ";"), paste0(card$percent, "%"))
        ))
      })
    )
  })

  output$security_endpoints <- reactable::renderReactable({
    log_duration("renderReactable_security_endpoints", {
      reactable::reactable(
        selected_endpoints(),
        columns = list(
          url = colDef(name = "URL", html = TRUE),
          condensed_organization_names = colDef(name = "Organization", html = TRUE),
          vendor_name = colDef(name = "Developer"),
          capability_fhir_version = colDef(
            name = "FHIR Version",
            sortable = TRUE,
            cell = function(value) {
              if (!is.na(value) && value != "") {
                htmltools::span(class = "version-badge", value)
              } else {
                value
              }
            }
          ),
          tls_version = colDef(name = "TLS Version"),
          code = colDef(
            name = "Authorization",
            sortable = TRUE,
            cell = function(value) {
              if (is.na(value) || value == "") return(value)
              class <- dplyr::case_when(
                value == "OAuth" ~ "auth-badge-oauth",
                value == "SMART-on-FHIR" ~ "auth-badge-smart",
                value == "Basic" ~ "auth-badge-basic",
                value == "Certificates" ~ "auth-badge-cert",
                TRUE ~ "auth-badge-default"
              )
              htmltools::span(class = paste("auth-badge", class), value)
            }
          )
        ),
        sortable = TRUE,
        showSortIcon = TRUE
      )
    })
  })

}