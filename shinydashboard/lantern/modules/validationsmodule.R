library(DT)
library(purrr)
library(reactable)

validationsmodule_UI <- function(id) {

  ns <- NS(id)

  tagList(
    fluidRow(
      column(width = 12,
        p("For information about the validation rules that Lantern evaluates, including their descriptions and references, please see the",
                a("documentation available here.", href = "Lantern Validation Rules and Descriptions.pdf", target = "_blank", class = "lantern-url")
        )
      )
    ),
    # Row for validation results chart
    fluidRow(
      column(width = 12,
        h2("Validation Results Count"),
        htmlOutput(ns("anchorlink")),
        uiOutput(ns("validation_results_plot"))
      )
    ),
    fluidRow(
      column(width = 12,
        p("The ONC Final Rule requires endpoints to support FHIR version 4.0.1, but we have included all endpoints for reference"),
        p("*Note: The messagingEndptRule is not broken, there is an issue with the Capability Statement invariant ", a("(cpb-3).", href = "http://hl7.org/fhir/capabilitystatement.html#invs", target = "_blank", class = "lantern-url"),
        "The invariant states that the Messaging endpoint has to be present when the kind is 'instance', and Messaging endpoint cannot be present when kind is NOT 'instance', but the FHIRPath expression is \"messaging.endpoint.empty() or kind = 'instance'\", which
         is not consistent with the expectation for the invariant and will not properly evaluate the conditions required.")
      )
    ),
    # Row for validation rules table and validation failure chart
    fluidRow(
      column(width = 3,
        h3("Validation Details"),
        p("Click on a rule below to filter the validation failure details table."),
        reactable::reactableOutput(ns("validation_details_table"))
      ),
      column(width = 9,
        h3("Validation Failure Details"),
        htmlOutput(ns("anchorpoint")),
        htmlOutput(ns("failure_table_subtitle")),
        tags$p("The URL for each endpoint in the table below can be clicked on to see additional information for that individual endpoint.", role = "comment"),
        reactable::reactableOutput(ns("validation_failure_table")),
        fluidRow(
          column(3, 
            div(style = "display: flex; justify-content: flex-start;", 
                uiOutput(ns("validation_prev_page_ui"))
            )
          ),
          column(6,
            div(style = "display: flex; justify-content: center; align-items: center; gap: 10px; margin-top: 8px;",
                numericInput(ns("validation_page_selector"), label = NULL, value = 1, min = 1, step = 1, width = "80px"),
                textOutput(ns("validations_current_page_info"), inline = TRUE)
            )
          ),
          column(3, 
            div(style = "display: flex; justify-content: flex-end;",
                uiOutput(ns("validation_next_page_ui"))
            )
          )
        ),
        p("A green check icon indicates that an endpoint has successfully returned a Conformance Resource/Capability Statement. A red X icon indicates the endpoint did not return a Conformance Resource/Capability Statement.")
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

  # Get total using COUNT
  validation_total_pages <- reactive({
    req(sel_fhir_version(), sel_vendor(), sel_validation_group())

    selected_rule <- if (!is.null(getReactableState("validation_details_table")$selected)) {
      deframe(validation_rules()[getReactableState("validation_details_table")$selected, "rule_name"])
    } else {
      "NO_RULES"
    }

    fhir_versions <- paste0("'", paste(sel_fhir_version(), collapse = "','"), "'")
    vendor_filter <- if(sel_vendor() != ui_special_values$ALL_DEVELOPERS) paste0("AND vendor_name = '", sel_vendor(), "'") else ""
    validation_group_filter <- if(sel_validation_group() != "All Groups") {
      references <- paste0("'", paste(validation_group_list[[sel_validation_group()]], collapse = "','"), "'")
      paste0("AND reference IN (", references, ")")
    } else {
      ""

    }

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

  observe({
    updateNumericInput(session, "validation_page_selector", 
                      max = validation_total_pages(),
                      value = validation_page_state())
  })

  # Handle page selector input
  observeEvent(input$validation_page_selector, {
    if (!is.null(input$validation_page_selector) && !is.na(input$validation_page_selector)) {
      new_page <- max(1, min(input$validation_page_selector, validation_total_pages()))
      validation_page_state(new_page)

      if (new_page != input$validation_page_selector) {
        updateNumericInput(session, "validation_page_selector", value = new_page)
      }
    }
  })

  output$validation_prev_page_ui <- renderUI({
    if (validation_page_state() > 1) {
      actionButton(ns("validation_prev_page"), "Previous", icon = icon("arrow-left"))
    } else {
      NULL
    }
  })

  output$validation_next_page_ui <- renderUI({
    if (validation_page_state() < validation_total_pages()) {
      actionButton(ns("validation_next_page"), "Next", icon = icon("arrow-right"))
    } else {
      NULL
    }
  })

  observeEvent(input$validation_next_page, {
    message("NEXT PAGE BUTTON CLICKED")
    if (validation_page_state() < validation_total_pages()) {
      new_page <- validation_page_state() + 1
      validation_page_state(new_page)
    }
  })

  observeEvent(input$validation_prev_page, {
    message("PREV PAGE BUTTON CLICKED")
    if (validation_page_state() > 1) {
      new_page <- validation_page_state() - 1
      validation_page_state(new_page)
    }
  })
  
  output$validations_current_page_info <- renderText({
    paste("of", validation_total_pages())
  })

  output$anchorpoint <- renderUI({
    HTML("<span id='anchorid'></span>")
  })

  output$anchorlink <- renderUI({
    HTML("<p>See additional validation details and failure information <a class=\"lantern-url\" href='#anchorid'>below</a></p>")
  })


  # Reset page to 1 whenever filters or selected rule changes
  observeEvent(list(sel_fhir_version(), sel_vendor(), sel_validation_group(), getReactableState("validation_details_table")$selected), {
    validation_page_state(1)
    updateNumericInput(session, "validation_page_selector", value = 1)
  })

  # Function to directly query validation results plot data from materialized view
  get_validation_plot_data <- function() {
    # Direct query to the materialized view
    tbl(db_connection, sql("SELECT * FROM mv_validation_results_plot")) %>%
      collect()
  }

  # Create table with all the distinct validation rule names
  validation_rules <- reactive({
    req(sel_fhir_version(), sel_vendor(), sel_validation_group())
    
    # Build filtering conditions for the SQL query
    fhir_versions <- paste0("'", paste(sel_fhir_version(), collapse = "','"), "'")
    vendor_filter <- if(sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
      paste0("AND vendor_name = '", sel_vendor(), "'")
    } else {
      ""
    }
    
    validation_group_filter <- if(sel_validation_group() != "All Groups") {
      references <- paste0("'", paste(validation_group_list[[sel_validation_group()]], collapse = "','"), "'")
      paste0("AND reference IN (", references, ")")
    } else {
      ""
    }
    
    # Query to get rule names based on filters
    query <- paste0("
      SELECT DISTINCT rule_name
      FROM mv_validation_results_plot
      WHERE fhir_version IN (", fhir_versions, ")
      ", vendor_filter, "
      ", validation_group_filter, "
      ORDER BY rule_name
    ")
    
    # Execute the query
    res <- dbGetQuery(db_connection, query)
    
    return(res)
  })

  # Create table for validation rule details table
  validation_details <- reactive({
    res <- validation_rules()
    
    fhir_version_filter <- FALSE
    req(sel_fhir_version())
    if (length(sel_fhir_version()) != 1 || sel_fhir_version() == "Unknown") {
      # Get version information directly from the materialized view
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

  # Create table containing all the validations that match current selected filtering criteria
  selected_validations <- reactive({
    # Get validation data directly from the validation_tbl function
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

  # Creates table containing the filtered validation's rule name, if its valid, and it's count
  select_validation_results <- reactive({
    req(sel_fhir_version(), sel_vendor(), sel_validation_group())
    
    # Build query with filters
    fhir_versions <- paste0("'", paste(sel_fhir_version(), collapse = "','"), "'")
    vendor_filter <- if(sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
      paste0("AND vendor_name = '", sel_vendor(), "'")
    } else {
      ""
    }
    
    validation_group_filter <- if(sel_validation_group() != "All Groups") {
      references <- paste0("'", paste(validation_group_list[[sel_validation_group()]], collapse = "','"), "'")
      paste0("AND reference IN (", references, ")")
    } else {
      ""
    }
    
    # Execute the filtered query
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

  # Creates a table of all the failed filtered validations, further filtering by the selected rule from the validation details table
  # Paginated using SQL LIMIT OFFSET
  paged_failed_validation_results <- reactive({
    req(sel_fhir_version(), sel_vendor(), sel_validation_group())
    
    # Get the selected rule if available
    selected_rule <- if (!is.null(getReactableState("validation_details_table")) && !is.null(getReactableState("validation_details_table")$selected)) {
      deframe(validation_rules()[getReactableState("validation_details_table")$selected, "rule_name"])
    } else {
      "NO_RULES"  # Default when no rule is selected
    }
    
    # Build filtering conditions
    fhir_versions <- paste0("'", paste(sel_fhir_version(), collapse = "','"), "'")
    vendor_filter <- if(sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
      paste0("AND vendor_name = '", sel_vendor(), "'")
    } else {
      ""
    }
    
    validation_group_filter <- if(sel_validation_group() != "All Groups") {
      references <- paste0("'", paste(validation_group_list[[sel_validation_group()]], collapse = "','"), "'")
      paste0("AND reference IN (", references, ")")
    } else {
      ""
    }

    limit <- validations_page_size
    offset <- (validation_page_state() - 1) * validations_page_size
    
    # Query to get failed validations for the selected rule
    query <- paste0("
      SELECT fhir_version, url, expected, actual, vendor_name
      FROM mv_validation_failures
      WHERE rule_name = '", selected_rule, "'
      AND fhir_version IN (", fhir_versions, ") ",
      vendor_filter, " ", 
      validation_group_filter, " ",
      "ORDER BY url LIMIT ", limit, " OFFSET ", offset
    )
    
    # Execute query
    res <- dbGetQuery(db_connection, query)
    
    # Add clickable URL links
    res <- res %>%
      mutate(url = paste0("<a class=\"lantern-url\" tabindex=\"0\" aria-label=\"Press enter to open pop up modal containing additional information for this endpoint.\" onkeydown = \"javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)\" onclick=\"Shiny.setInputValue(\'endpoint_popup\',&quot;", url, "&&", "None", "&quot,{priority: \'event\'});\">", url, "</a>"))
    
    return(res)
  })

  output$validation_details_table <-  reactable::renderReactable({
    reactable(validation_details() %>% select(entry),
              columns = list(
                entry = colDef(name = "Validation Rules", html = TRUE)
              ),
              selection = "single",
              onClick = "select",
              defaultSelected = c(1),
              pagination = FALSE,
              height = 500
    )
  })

  # Reactive to calculate the plot height for the validation tables based on how many rows are in the resulting selected validation results
  validation_plot_height <- reactive({
    max(nrow(select_validation_results()) * 25, 400)
  })

  # Calls function to render the validation results count chart if there is data or an empty plot if no data
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

  # Renders the validation result count chart which displays the number of endpoints that failed or passed each validation test
  output$validation_bar_plot <- renderCachedPlot({
    ggplot(select_validation_results(), aes(x = fct_rev(as.factor(rule_name)), y = count, fill = valid)) +
      geom_col(width = 0.8) +
      geom_text(aes(label = stat(y)), position = position_stack(vjust = 0.5)) +
      ggtitle("Validation Results") +
      theme(plot.title = element_text(hjust = 0.5)) +
      theme(legend.position = "bottom") +
      theme(legend.title = element_blank()) +
      theme(text = element_text(size = 14)) +
      labs(x = "", y = "", fill = "Valid") +
      scale_y_continuous(sec.axis = sec_axis(~.)) +
      scale_fill_manual(values = c("Failure" = "red", "Success" = "seagreen3"), limits = c("Failure", "Success")) +
      guides(fill = guide_legend(reverse = TRUE)) +
      coord_flip()
  },
    sizePolicy = sizeGrowthRatio(width = 400,
                                 height = 400,
                                 growthRate = 1.2),
    res = 72,
    cache = "app",
    cacheKeyExpr = {
      list(sel_fhir_version(), sel_vendor(), sel_validation_group(), now("UTC"))
    })

  # Renders an empty validation result count chart when no data available
  output$validation_bar_empty_plot <- renderPlot({
    ggplot(select_validation_results()) +
      geom_col(width = 0.8) +
      labs(x = "", y = "") +
      theme(axis.text.x = element_blank(),
            axis.text.y = element_blank(), axis.ticks = element_blank()) +
      annotate("text", label = "There are no validation results for the endpoints\nthat pass the selected filtering criteia", x = 1, y = 2, size = 4.5, colour = "red", hjust = 0.5)
  })

  cap_stat_icon <- function(fhir_version) {
    icon <- tagAppendAttributes(shiny::icon("check-circle-o"), style = "color: green", "aria-hidden" = "true")
    if (fhir_version == "No Cap Stat") {
      icon <- tagAppendAttributes(shiny::icon("times-circle-o"), style = "color: red", "aria-hidden" = "true")
    }
    icon
  }

  output$failure_table_subtitle <- renderUI({
    p(paste("Rule: ", deframe(validation_rules()[getReactableState("validation_details_table")$selected, "rule_name"])))
  })


  # Renders the validation failure data table which contains the endpoints that failed validation tests and what the expected and actual values were
  output$validation_failure_table <- reactable::renderReactable({
    paged_data <- paged_failed_validation_results()
    reactable(paged_data,
              defaultColDef = colDef(
                style = function(value, index) {
                  if (paged_data$fhir_version[index] == "No Cap Stat") {
                    list(background = "rgba(0, 0, 0, 0.03)")
                  }
                }
              ),
              columns = list(
                fhir_version = colDef(name = "FHIR Version",
                                      cell = function(value, index) {
                                        image <- cap_stat_icon(paged_data$fhir_version[index])
                                        tagList(
                                          div(style = list(display = "inline-block", width = "45px"), image),
                                          value
                                        )
                                      }),
                url = colDef(name = "URL", html = TRUE, minWidth = 300),
                expected = colDef(name = "Expected Value"),
                actual = colDef(name = "Actual Value"),
                vendor_name = colDef(name = "Certified API Developer Name")
              )
    )
  })
}