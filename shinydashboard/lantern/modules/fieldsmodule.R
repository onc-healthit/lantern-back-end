# Fields Module
library(reactable)

fieldsmodule_UI <- function(id) {
  ns <- NS(id)
  tagList(
    h1("FHIR CapabilityStatement / Conformance Fields"),
    p("This is the list of fields included in the FHIR CapabilityStatements / Conformance Resources from the endpoints."),
    tags$style(HTML("
      .field-list {
        display: grid;
        grid-template-columns: repeat(4, minmax(350px, auto));
        overflow-x: scroll;
        padding-bottom: 15px;
        resize: none;
      }
    ")),
    htmlOutput(ns("capstat_fields_text")),
    fluidRow(
      column(width = 5,
             h4("Required Fields"),
             reactable::reactableOutput(ns("capstat_fields_table_required")),
             h4("Optional Fields"),
             reactable::reactableOutput(ns("capstat_fields_table_optional"))),
      column(width = 7,
             h4("Supported CapabilityStatement / Conformance Fields"),
             uiOutput(ns("fields_plot"))
      )
    ),
    h1("FHIR CapabilityStatement / Conformance Extensions"),
    p("This is the list of extensions included in the FHIR CapabilityStatements / Conformance Resources from the endpoints."),
    tags$style(HTML("
      .extension-list {
        display: grid;
        grid-template-columns: repeat(3, minmax(480px, auto));
        overflow-x: scroll;
        padding-bottom: 15px;
        resize: none;
      }
    ")),
    htmlOutput(ns("capstat_extension_text")),
    fluidRow(
      column(width = 5,
             h4("Supported Extensions:"),
             reactable::reactableOutput(ns("capstat_extensions_table"))),
      column(width = 7,
             h4("Supported CapabilityStatement / Conformance Extensions"),
             uiOutput(ns("extensions_plot"))
      )
    )
  )
}

fieldsmodule <- function(
  input,
  output,
  session,
  sel_fhir_version,
  sel_vendor
) {

  ns <- session$ns

  dstu2 <- c("0.4.0", "0.5.0", "1.0.0", "1.0.1", "1.0.2")
  stu3 <- c("1.1.0", "1.2.0", "1.4.0", "1.6.0", "1.8.0", "3.0.0", "3.0.1", "3.0.2")
  r4 <- c("3.2.0", "3.3.0", "3.5.0", "3.5a.0", "4.0.0", "4.0.1")


  get_capstat_values <- function(extension) {
    res <- isolate(app_data$capstat_fields())
    version_filter <- FALSE

    req(sel_fhir_version())
    if (sel_fhir_version() != ui_special_values$ALL_FHIR_VERSIONS) {
      res <- res %>% filter(fhir_version == sel_fhir_version())
      version_filter <- TRUE
    }

    if (extension) {
      res <- res %>%
      filter(extension == "true")
    } else {
      res <- res %>%
      filter(extension == "false")
    }
    res <- res %>%
    group_by(field) %>%
    arrange(fhir_version, .by_group = TRUE)

    # Unknown FHIR versions default to DSTU2
    res <- res %>% mutate(fhir_version_name = case_when(
    fhir_version %in% dstu2 ~ "DSTU2",
    fhir_version %in% stu3 ~ "STU3",
    fhir_version %in% r4 ~ "R4",
    TRUE ~ "DSTU2"
    ))

    res <- res %>% summarise(fhir_version_names = paste(unique(fhir_version_name), collapse = ", "))

    if (version_filter) {
      res$field_version <- res$field
    } else {
      res$field_version <- paste(res$field, " (", res$fhir_version_names, ")", sep = "")
    }

    res
}

output$capstat_fields_text <- renderUI({
    col <- get_capstat_values(FALSE) %>% pull(3)
    liElem <- paste("<li>", col, "</li>", collapse = " ")
    divElem <- paste("<div class='field-list'>", liElem, "</div>")
    fullHtml <- paste("Lantern checks for the following fields: ", divElem)
    HTML(fullHtml)
  })

output$capstat_extension_text <- renderUI({
    col <- get_capstat_values(TRUE) %>% pull(3)
    liElem <- paste("<li>", col, "</li>", collapse = " ")
    divElem <- paste("<div class='extension-list'>", liElem, "</div>")
    fullHtml <- paste("Lantern checks for the following extensions: ", divElem)
    HTML(fullHtml)
})

  selected_fhir_endpoints <- reactive({
    res <- isolate(app_data$capstat_fields())
    req(sel_fhir_version(), sel_vendor())
    # If the selected dropdown value for the fhir verison is not the default "All FHIR Versions", filter
    # the capability statement fields by which fhir verison they're associated with
    if (sel_fhir_version() != ui_special_values$ALL_FHIR_VERSIONS) {
      res <- res %>% filter(fhir_version == sel_fhir_version())
    }
    # Same as above but with the vendor dropdown
    if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
      res <- res %>% filter(vendor_name == sel_vendor())
    }
    res
  })

  capstat_field_count <- reactive({
    get_capstat_fields_count(selected_fhir_endpoints(), "false")
  })

  capstat_extension_count <- reactive({
    get_capstat_fields_count(selected_fhir_endpoints(), "true")
  })

  # Required Capability Statement fields that we are tracking
  required_fields <- c("status", "kind", "fhirVersion", "format", "date")


   output$capstat_fields_table_required <- reactable::renderReactable({
     reactable(
              capstat_field_count() %>% filter(Fields %in% required_fields) %>% rename("FHIR Version" = fhir_version),
              columns = list(
                Endpoints = colDef(
                  aggregate = "sum",
                  format = list(aggregated = colFormat(prefix = "Total: "))
                ),
                Fields = colDef(
                  minWidth = 150
                ),
                "FHIR Version" = colDef(
                  align = "center"
                )
              ),
              groupBy = "Fields",
              sortable = TRUE,
              searchable = TRUE,
              striped = TRUE,
              showSortIcon = TRUE,
              defaultPageSize = 5

     )
  })

   output$capstat_fields_table_optional <- reactable::renderReactable({
     reactable(
              capstat_field_count() %>% filter(!(Fields %in% required_fields)) %>% rename("FHIR Version" = fhir_version),
              columns = list(
                Endpoints = colDef(
                  aggregate = "sum",
                  format = list(aggregated = colFormat(prefix = "Total: "))
                ),
                Fields = colDef(
                  minWidth = 150
                ),
                "FHIR Version" = colDef(
                  align = "center"
                )
              ),
              groupBy = "Fields",
              sortable = TRUE,
              searchable = TRUE,
              striped = TRUE,
              showSortIcon = TRUE,
              defaultPageSize = 50

     )
  })

  # Table of the extension counts
   output$capstat_extensions_table <- reactable::renderReactable({
     reactable(
              capstat_extension_count() %>% rename("FHIR Version" = fhir_version),
              columns = list(
                Endpoints = colDef(
                  aggregate = "sum",
                  format = list(aggregated = colFormat(prefix = "Total: "))
                )
              ),
              groupBy = "Fields",
              sortable = TRUE,
              searchable = TRUE,
              striped = TRUE,
              showSortIcon = TRUE,
              defaultPageSize = 10

     )
  })



  vendor <- reactive({
    sel_vendor()
  })

  plot_height <- reactive({
    max(nrow(capstat_field_count()) * 25, 400)
  })

  plot_height_extensions <- reactive({
    max(nrow(capstat_extension_count()) * 25, 400)
  })

  output$fields_plot <- renderUI({
    if (nrow(capstat_field_count()) != 0) {
      tagList(
        plotOutput(ns("fields_bar_plot"), height = plot_height())
      )
    }
    else {
      tagList(
        plotOutput(ns("fields_bar_empty_plot"), height = plot_height())
      )
    }
  })
  output$fields_bar_plot <- renderCachedPlot({
    ggplot(capstat_field_count(), aes(x = fct_rev(as.factor(Fields)), y = Endpoints, fill = fhir_version)) +
      geom_col(width = 0.8) +
      geom_text(aes(label = stat(y)), position = position_stack(vjust = 0.5)) +
      theme(legend.position = "top") +
      theme(text = element_text(size = 14)) +
      labs(x = "", y = "Number of Endpoints", fill = "FHIR Version", title = vendor()) +
      scale_y_continuous(sec.axis = sec_axis(~., name = "Number of Endpoints")) +
      coord_flip()
  },
    sizePolicy = sizeGrowthRatio(width = 400,
                                  height = 400,
                                  growthRate = 1.2),
    res = 72,
    cache = "app",
    cacheKeyExpr = {
      list(sel_fhir_version(), sel_vendor(), app_data$last_updated())
    }
  )
  output$fields_bar_empty_plot <- renderPlot({
    ggplot(capstat_field_count()) +
      geom_col(width = 0.8) +
      geom_text(aes(label = stat(y)), position = position_stack(vjust = 0.5)) +
      theme(legend.position = "top") +
      theme(text = element_text(size = 14)) +
      labs(x = "", y = "Number of Endpoints", fill = "FHIR Version", title = vendor()) +
      theme(axis.text.x = element_blank(),
      axis.text.y = element_blank(), axis.ticks = element_blank()) +
      scale_y_continuous(sec.axis = sec_axis(~., name = "Number of Endpoints")) +
      coord_flip() +
      annotate("text", label = "There are no FHIR CapabilityStatement / Conformance fields supported by the endpoints\nthat pass the selected filtering criteia", x = 1, y = 2, size = 4.5, colour = "red", hjust = 0.5)
  })

  output$extensions_plot <- renderUI({
    if (nrow(capstat_extension_count()) != 0) {
      tagList(
        plotOutput(ns("extensions_bar_plot"), height = plot_height_extensions())
      )
    }
    else {
      tagList(
        plotOutput(ns("extensions_bar_empty_plot"), height = plot_height_extensions())
      )
    }
  })

  output$extensions_bar_plot <- renderCachedPlot({
    ggplot(capstat_extension_count(), aes(x = fct_rev(as.factor(Fields)), y = Endpoints, fill = fhir_version)) +
      geom_col(width = 0.8) +
      geom_text(aes(label = stat(y)), position = position_stack(vjust = 0.5)) +
      theme(legend.position = "top") +
      theme(text = element_text(size = 14)) +
      labs(x = "", y = "Number of Endpoints", fill = "FHIR Version", title = vendor()) +
      scale_y_continuous(sec.axis = sec_axis(~., name = "Number of Endpoints")) +
      coord_flip()
  },
    sizePolicy = sizeGrowthRatio(width = 400,
                                  height = 400,
                                  growthRate = 1.2),
    res = 72,
    cache = "app",
    cacheKeyExpr = {
      list(sel_fhir_version(), sel_vendor(), app_data$last_updated())
    }
  )
  output$extensions_bar_empty_plot <- renderPlot({
    ggplot(capstat_extension_count()) +
      geom_col(width = 0.8) +
      geom_text(aes(label = stat(y)), position = position_stack(vjust = 0.5)) +
      theme(legend.position = "top") +
      theme(text = element_text(size = 14)) +
      labs(x = "", y = "Number of Endpoints", fill = "FHIR Version", title = vendor()) +
      theme(axis.text.x = element_blank(),
      axis.text.y = element_blank(), axis.ticks = element_blank()) +
      scale_y_continuous(sec.axis = sec_axis(~., name = "Number of Endpoints")) +
      coord_flip() +
      annotate("text", label = "There are no FHIR Capability Extensions supported by the endpoints\nthat pass the selected filtering criteia", x = 1, y = 2, size = 4.5, colour = "red", hjust = 0.5)
  })
}
