# Shared pagination scaffolding used by every paginated table module (endpoints, organizations,
# security, validations, values, profile, smartresponse, resource, contacts). Each of those
# modules used to independently reimplement this exact block (a page-state reactiveVal, a
# numericInput <-> page-state reconciler that breaks its own feedback loop via isolate() and
# invalidateLater(100), next/prev button click handlers, and next/prev button renderUI blocks) --
# ~150-250 lines per module, duplicated 9 times. Only this scaffolding is shared; each module's
# own SQL/filter-building and total_pages logic (which vary too much to safely unify) stay in
# the module itself.
#
# create_pager() creates the page-state reactiveVal, wires up the reconciler and next/prev
# handlers, renders the next/prev buttons, and returns the page-state reactiveVal for the calling
# module to use exactly as it used to use its own local xxx_page_state reactiveVal (e.g. for its
# total_pages/data-fetch OFFSET calculation, and for its own filter-change reset observer, which
# stays in the module since its watched reactives differ per module).
#
# id parameters must match the module's own UI element ids (as passed to input$/ns()) -- create_pager
# does not rename or introduce any new input/output ids, it just centralizes the server-side wiring.
create_pager <- function(input, output, session,
                          page_selector_id, prev_button_id, next_button_id,
                          prev_output_id, next_output_id,
                          total_pages) {
  ns <- session$ns
  page_state <- reactiveVal(1)

  # Break the feedback loop with isolate(): whenever page_state changes (via the next/prev
  # buttons or the module's own filter-change reset), push the new value into the numericInput.
  observe({
    new_page <- page_state()
    current_selector <- input[[page_selector_id]]

    if (is.null(current_selector) ||
        is.na(current_selector) ||
        !is.numeric(current_selector) ||
        current_selector != new_page) {

      isolate({
        updateNumericInput(session, page_selector_id, max = total_pages(), value = new_page)
      })
    }
  })

  # Handle the user typing directly into the page-selector numericInput.
  observeEvent(input[[page_selector_id]], {
    current_input <- input[[page_selector_id]]

    if (!is.null(current_input) &&
        !is.na(current_input) &&
        is.numeric(current_input) &&
        current_input > 0) {

      new_page <- max(1, min(current_input, total_pages()))

      if (new_page != page_state()) {
        page_state(new_page)
      }

      if (new_page != current_input) {
        updateNumericInput(session, page_selector_id, value = new_page)
      }
    } else {
      # If input is invalid (empty, NA, or <= 0), reset to current page.
      # Use a small delay to prevent immediate feedback loop.
      invalidateLater(100)
      updateNumericInput(session, page_selector_id, value = page_state())
    }
  }, ignoreInit = TRUE)

  observeEvent(input[[next_button_id]], {
    if (page_state() < total_pages()) {
      page_state(page_state() + 1)
    }
  })

  observeEvent(input[[prev_button_id]], {
    if (page_state() > 1) {
      page_state(page_state() - 1)
    }
  })

  output[[prev_output_id]] <- renderUI({
    if (page_state() > 1) {
      actionButton(ns(prev_button_id), "Previous", icon = icon("arrow-left"))
    } else {
      NULL
    }
  })

  output[[next_output_id]] <- renderUI({
    if (page_state() < total_pages()) {
      actionButton(ns(next_button_id), "Next", icon = icon("arrow-right"))
    } else {
      NULL
    }
  })

  page_state
}
