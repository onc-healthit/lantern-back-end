library(DT)
library(purrr)

validationsmodule_UI <- function(id) {

  ns <- NS(id)

  tagList(
    fluidRow(
      column(width = 12,
        p("The ONC Final Rule requires endpoints to support FHIR version 4.0.1, but we have included all endpoints for reference")
      )
    ),
    fluidRow(
      column(width = 12,
        h3("Validation Results Count"),
        uiOutput(ns("validation_results_plot"))
      )
    ),
    fluidRow(
      column(width = 3,
        h3("Validation Details"),
        p("Click on a rule below to filter the validation failure details table."),
        DT::dataTableOutput(ns("validation_details_table"))
      ),
      column(width = 9,
        h3("Validation Failure Details"),
        DT::dataTableOutput(ns("validation_failure_table"))
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

  validation_rules <- reactive({
    res <- selected_validations()
    res <- res %>%
           distinct(rule_name) %>%
           arrange(rule_name)
    res
  })

  validation_details <- reactive({
    res <- validation_rules()
    res <- res %>%
      mutate(comment_line = paste("Comment:", validation_rules_descriptions[rule_name])) %>%
      mutate(rule_name_line = paste("Name:", rule_name)) %>%
      mutate(num = paste(row_number(), ".")) %>%
      distinct(num, rule_name_line, comment_line) %>%
      mutate(entry = paste(num,  rule_name_line, comment_line, sep = "<br>")) %>%
      select(entry)
    res
  })

  selected_validations <- reactive({
    res <- isolate(app_data$validation_tbl())
    req(sel_fhir_version(), sel_vendor(), sel_validation_group())
    if (sel_fhir_version() != ui_special_values$ALL_FHIR_VERSIONS) {
      res <- res %>% filter(fhir_version == sel_fhir_version())
    }
    if (sel_validation_group() != "All Groups") {
      res <- res %>% filter(reference %in% validation_group_list[[sel_validation_group()]])
    }
    if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
      res <- res %>% filter(vendor_name == sel_vendor())
    }
    res
  })

  select_validation_results <- reactive({
    res <- selected_validations()
    res <- res %>%
      group_by(rule_name, valid) %>%
      count() %>%
      rename(count = n) %>%
      select(rule_name, valid, count)
    res
  })

  failed_validation_results <- reactive({
    res <- selected_validations()
    if (length(input$validation_details_table_rows_selected) > 0) {
      selected_rule <- deframe(validation_rules()[input$validation_details_table_rows_selected, "rule_name"])
      res <- res %>%
        filter(rule_name == selected_rule)
    } else {
      res <- res %>%
        filter(rule_name == "NO_RULES")
    }
    res <- res %>%
        filter(valid == FALSE)
    res
  })

  output$validation_details_table <- DT::renderDataTable({
    datatable(validation_details() %>% select(entry),
      colnames = "",
      rownames = FALSE,
      escape = FALSE,
      selection = list(mode = "single", selected = c(1), target = "row"),
      options = list(scrollX = TRUE, scrollY = 500, scrollCollapse = TRUE, paging = FALSE, dom = "t", ordering = FALSE)
    )
  })

  validation_plot_height <- reactive({
    max(nrow(select_validation_results()) * 25, 400)
  })

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
  output$validation_bar_plot <- renderCachedPlot({
    ggplot(select_validation_results(), aes(x = fct_rev(as.factor(rule_name)), y = count, fill = valid)) +
      geom_col(width = 0.8) +
      geom_text(aes(label = stat(y)), position = position_stack(vjust = 0.5)) +
      ggtitle("Validation Results") +
      theme(plot.title = element_text(hjust = 0.5)) +
      theme(legend.position = "bottom") +
      theme(text = element_text(size = 14)) +
      labs(x = "", y = "", fill = "Valid") +
      scale_y_continuous(sec.axis = sec_axis(~.)) +
      scale_fill_manual(values = c("FALSE" = "red", "TRUE" = "seagreen3"), limits = c("FALSE", "TRUE")) +
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

  output$validation_bar_empty_plot <- renderPlot({
    ggplot(select_validation_results()) +
    geom_col(width = 0.8) +
    labs(x = "", y = "") +
    theme(axis.text.x = element_blank(),
    axis.text.y = element_blank(), axis.ticks = element_blank()) +
    annotate("text", label = "There are no validation results for the endpoints\nthat pass the selected filtering criteia", x = 1, y = 2, size = 4.5, colour = "red", hjust = 0.5)
  })

    output$validation_failure_table <- DT::renderDataTable({
    datatable(failed_validation_results() %>% select(url, expected, actual, vendor_name, fhir_version),
              colnames = c("URL", "Expected Value", "Actual Value", "Certified API Developer Name", "FHIR Version"),
              rownames = FALSE,
              selection = "none",
              caption = paste("Rule: ", deframe(validation_rules()[input$validation_details_table_rows_selected, "rule_name"])),
              options = list(scrollX = TRUE)
            )
  })
}
