# Fields Module
library(reactable)

fieldsmodule_UI <- function(id) {
  ns <- NS(id)
  tagList(
    h1("FHIR CapabilityStatement / Conformance Fields"),
    p("This is the list of fields included in the FHIR CapabilityStatements / Conformance Resources from the endpoints."),
    tags$style(HTML("
      .field-list ul {
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
             h2("Required Fields"),
             reactable::reactableOutput(ns("capstat_fields_table_required")),
             h2("Optional Fields"),
             reactable::reactableOutput(ns("capstat_fields_table_optional"))),
      column(width = 7,
             h2("Supported CapabilityStatement / Conformance Fields"),
             uiOutput(ns("fields_plot"))
      )
    ),
    h1("FHIR CapabilityStatement / Conformance Extensions"),
    p("This is the list of extensions included in the FHIR CapabilityStatements / Conformance Resources from the endpoints."),
    tags$style(HTML("
      .extension-list ul{
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
             h2("Supported Extensions:"),
             reactable::reactableOutput(ns("capstat_extensions_table"))),
      column(width = 7,
             h2("Supported CapabilityStatement / Conformance Extensions"),
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

  get_capstat_values_mv <- function(extension) {
    req(sel_fhir_version())

    if (extension) {
      mv <- "mv_capstat_values_extension"
    } else {
      mv <- "mv_capstat_values_fields"
    }

    # Build filtering conditions for the SQL query
    fhir_versions <- paste0("'", paste(sel_fhir_version(), collapse = "','"), "'")

    # Direct query to the required materialized view
    query <- paste0("
        SELECT DISTINCT field_version 
        FROM ", mv, "
        WHERE fhir_version IN (", fhir_versions, ")
        ORDER BY field_version
      ")

    # Execute the query
    res <- dbGetQuery(db_connection, query) %>% collect()

    # Remove fhir_version info if only one fhir version selected
    if (length(sel_fhir_version()) == 1) {
      res <- res %>% mutate(field_version = sub(" \\(.*\\)", "", field_version), field_version)
    }

    return(res)
}

get_capstat_fields_mv <- function(db_connection, fhir_version = NULL, vendor = NULL) {
  # Start with base query
  query <- tbl(db_connection, "mv_capstat_fields")

  # Apply filters in SQL before collecting data
  if (!is.null(fhir_version) && length(fhir_version) > 0) {
    query <- query %>% filter(fhir_version %in% !!fhir_version)
  }

  if (!is.null(vendor) && vendor != ui_special_values$ALL_DEVELOPERS) {
    query <- query %>% filter(vendor_name == !!vendor)
  }

  # Collect the data after applying filters in SQL
  result <- query %>% collect()

  return(result)
}

get_capstat_fields_count <- function(sel_fhir_version, sel_vendor, extensionBool) {
  # Build filtering conditions for the SQL query
  fhir_versions <- paste0("'", paste(sel_fhir_version, collapse = "','"), "'")
  vendor_filter <- if(!is.null(sel_vendor) && sel_vendor != ui_special_values$ALL_DEVELOPERS) {
    paste0("AND vendor_name = '", sel_vendor, "'")
  } else {
    ""
  }
  
  # Direct query to the materialized view
  query <- paste0("
      SELECT field as \"Fields\", fhir_version, COUNT(*) as \"Endpoints\"
      FROM mv_capstat_fields
      WHERE fhir_version IN (", fhir_versions, ")
      ", vendor_filter, "
      ", "AND exist = 'true'", 
      " AND extension = '", extensionBool, "'",
      "
      GROUP BY field, fhir_version
      ORDER BY field, fhir_version
    ")
  
  # Execute the query
  res <- dbGetQuery(db_connection, query) %>% collect()

  res <- res %>% mutate(Endpoints = as.integer(Endpoints)) %>% as_tibble()
  return(res)
}

output$capstat_fields_text <- renderUI({
    col <- get_capstat_values_mv(FALSE)
    liElem <- tagList()
    if (length(col) > 0) {
      liElem <- apply(col, 1, function(x) tags$li(x["field_version"]))
    }
    ulElem <- tags$ul(liElem, tabindex = "0")
    divElem <- div(ulElem, class = "field-list")
    tagList(HTML("Lantern checks for the following fields: "), divElem)
  })

output$capstat_extension_text <- renderUI({
    col <- get_capstat_values_mv(TRUE)
    liElem <- tagList()
    if (length(col) > 0) {
      liElem <- apply(col, 1, function(x) tags$li(x["field_version"]))
    }
    ulElem <- tags$ul(liElem, tabindex = "0")
    divElem <- div(ulElem, class = "extension-list")
    tagList(HTML("Lantern checks for the following extensions: "), divElem)
})

  selected_fhir_endpoints <- reactive({
    # Get current filter values
    current_fhir <- sel_fhir_version()
    current_vendor <- sel_vendor()

    req(current_fhir, current_vendor)

    # Get filtered data from the materialized view function
    res <- get_capstat_fields_mv(
      db_connection,
      fhir_version = current_fhir,
      vendor = current_vendor
    )

    res
  })

  capstat_field_count <- reactive({
    get_capstat_fields_count(sel_fhir_version(), sel_vendor(), "false")
  })

  capstat_extension_count <- reactive({
    get_capstat_fields_count(sel_fhir_version(), sel_vendor(), "true")
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
    } else {
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
      list(sel_fhir_version(), sel_vendor(), now("UTC"))
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
    } else {
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
      list(sel_fhir_version(), sel_vendor(), now("UTC"))
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
