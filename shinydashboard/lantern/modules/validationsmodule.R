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

  output$anchorpoint <- renderUI({
    HTML("<span id='anchorid'></span>")
  })

  output$anchorlink <- renderUI({
    HTML("<p>See additional validation details and failure information <a class=\"lantern-url\" href='#anchorid'>below</a></p>")
  })

  # Create table with all the distinct validation rule names
  validation_rules <- reactive({
    res <- selected_validations() %>% distinct(url, fhir_version, vendor_name, rule_name, valid, expected, actual, comment, reference) %>% select(url, fhir_version, vendor_name, rule_name, valid, expected, actual, comment, reference)
    res <- res %>%
           distinct(rule_name) %>%
           arrange(rule_name)
    res
  })

  # Create table for validation rule details table
  validation_details <- reactive({
    res <- validation_rules()

    fhir_version_filter <- FALSE
    req(sel_fhir_version())
    if (length(sel_fhir_version()) != 1 || sel_fhir_version() == "Unknown") {
      versions <- get_validation_versions()
      res <- res %>%
      left_join(versions %>% select(validation_name, fhir_version_names),
        by = c("rule_name" = "validation_name")) %>%
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
        mutate(entry = paste(num,  rule_name_line, versions_line, comment_line, sep = "<br>")) %>%
        select(entry)
      } else {
        res <- res %>%
        distinct(num, rule_name_line, comment_line) %>%
        mutate(entry = paste(num,  rule_name_line, comment_line, sep = "<br>")) %>%
        select(entry)
      }

    res
  })

  # Create table containing all the validations that pass current selected filtering criteria
  selected_validations <- reactive({
    res <- isolate(app_data$validation_tbl())
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

  get_validation_versions <- reactive({
    res <- isolate(app_data$validation_tbl())
    res <- res %>%
    filter(fhir_version != "Unknown", fhir_version != "No Cap Stat") %>%
    group_by(rule_name) %>%
    rename(validation_name = rule_name) %>%
    arrange(fhir_version, .by_group = TRUE) %>%
    mutate(fhir_version_name = case_when(
      fhir_version %in% dstu2 ~ "DSTU2",
      fhir_version %in% stu3 ~ "STU3",
      fhir_version %in% r4 ~ "R4",
      TRUE ~ "DSTU2"
    )) %>%
    summarise(fhir_version_names = paste(unique(fhir_version_name), collapse = ", "))
    res
  })

  # Creates table containing the filtered validation's rule name, if its valid, and it'c count
  select_validation_results <- reactive({
    res <- selected_validations() %>% distinct(url, fhir_version, vendor_name, rule_name, valid, expected, actual, comment, reference) %>% select(url, fhir_version, vendor_name, rule_name, valid, expected, actual, comment, reference)
    res <- res %>%
      group_by(rule_name, valid) %>%
      count() %>%
      rename(count = n) %>%
      select(rule_name, valid, count) %>%
      mutate(valid = if_else(valid == TRUE, "Success", "Failure"))
    res
  })

  # Creates a table of all the failed filtered validations, further filtering by the selected rule from the validation details table
  failed_validation_results <- reactive({
    res <- selected_validations() %>%
    mutate(url = linkURL) %>%
    distinct(url, fhir_version, vendor_name, rule_name, valid, expected, actual, comment, reference) %>%
    select(url, fhir_version, vendor_name, rule_name, valid, expected, actual, comment, reference)
    if (!is.null(getReactableState("validation_details_table")) && !is.null(getReactableState("validation_details_table")$selected)) {
      selected_rule <- deframe(validation_rules()[getReactableState("validation_details_table")$selected, "rule_name"])
      res <- res %>%
        filter(rule_name == selected_rule)
    } else {
      res <- res %>%
        filter(rule_name == "NO_RULES")
    }
    res <- res %>%
        filter(valid == FALSE)
    res %>% select(fhir_version, url, expected, actual, vendor_name)
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
      list(sel_fhir_version(), sel_vendor(), sel_validation_group(), app_data$last_updated())
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
    reactable(failed_validation_results(),
              defaultColDef = colDef(
                style = function(value, index) {
                  if (failed_validation_results()$fhir_version[index] == "No Cap Stat") {
                    list(background = "rgba(0, 0, 0, 0.03)")
                  }
                }
              ),
              columns = list(
                fhir_version = colDef(name = "FHIR Version",
                    cell = function(value, index) {
                        image <- cap_stat_icon(failed_validation_results()$fhir_version[index])
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
