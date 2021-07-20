library(DT)
library(purrr)

validationsmodule_UI <- function(id) {

  ns <- NS(id)

  tagList(
    fluidRow(
      column(width = 4,
            "sidebar",
             h3("Validation Details"),
             DT::dataTableOutput(ns("validation_details_table"))
            ),
      column(width = 8,
          "main",
          fluidRow(
            column(width = 12,
                h3("Validation Results Count"),
                uiOutput(ns("validation_results_plot"))
            )
          ), fluidRow(
            column(width = 12,
              h3("Validation Failure Details"),
              DT::dataTableOutput(ns("validation_failure_table"))
            )
          )
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

  output$endpoint_count <- renderText({
    paste("Matching Endpoints:", nrow(selected_fhir_endpoints()))
  })

  validation_rules <- reactive ({
    res <- isolate(app_data$validation_tbl())
    res <- res %>%
           filter(reference != "") %>%
           mutate(rule_name = paste("Name:", rule_name)) %>%
           mutate(comment = paste("Comment:", comment)) %>%
           mutate(num = row_number()) %>%
           distinct(num, rule_name, comment)
    x <- data.frame("validation_details" = c(rbind(res$num, res$rule_name, res$comment)))
    x
  })

  selected_validations <- reactive({
    res <- isolate(app_data$validation_tbl())
    req(sel_fhir_version(), sel_vendor(), sel_validation_group())
    if (sel_fhir_version() != ui_special_values$ALL_FHIR_VERSIONS) {
      res <- res %>% filter(fhir_version == sel_fhir_version())
    }
    if (sel_validation_group() != "All Groups") {
      if (sel_validation_group() == "Other") {
        res <- res %>% filter(reference == "")
      } else if (sel_validation_group() == "HTTP") {
        res <- res %>% filter(reference == "http://hl7.org/fhir/http.html")
      } else if (sel_validation_group() == "Capability Statements") {
        res <- res %>% filter(reference == "http://hl7.org/fhir/capabilitystatement.html")
      } else if (sel_validation_group() == "SMART") {
        res <- res %>% filter(reference == "http://www.hl7.org/fhir/smart-app-launch/conformance/index.html")
      } else if (sel_validation_group() == "US-CORE") {
        res <- res %>% filter(reference == "https://www.hl7.org/fhir/us/core/CapabilityStatement-us-core-server.html" | reference == "https://www.hl7.org/fhir/us/core/security.html")
      } else if (sel_validation_group() == "Certification Criteria") {
        res <- res %>% filter(reference == "https://www.healthit.gov/cures/sites/default/files/cures/2020-03/APICertificationCriterion.pdf")
      }
    }
    if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
      res <- res %>% filter(vendor_name == sel_vendor())
    }
    res
  })

  select_validation_results <- reactive ({
    res <- selected_validations()
    res <- res %>% 
            group_by(rule_name, valid) %>%
            count() %>%
            rename(count = n) %>%
            select(rule_name, valid, count)
    res

  })

  failed_validation_results <- reactive ({
    res <- selected_validations()
    res <- res %>% 
          filter(valid == FALSE)
    res 
  })

  output$validation_details_table <- DT::renderDataTable({
    datatable(validation_rules() %>% select(validation_details),
              colnames = c("Rules"),
              rownames = FALSE,
              options = list(scrollX = TRUE, scrollY = 700, paging = FALSE, dom = 't', ordering = FALSE)
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
      scale_fill_manual(values = c("red", "seagreen3")) +
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
              options = list(scrollX = TRUE)
            )
  })

}