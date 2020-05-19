# Functions to compute metrics on endpoints

get_endpoint_totals <- function(fhir_endpoints_con,fhir_endpoints_info_con) {
  # Get the table of fhir endpoints. There may be endpoints we have not reached
  # so get counts of indexed and non-indexed endpoints
  fhir_endpoints_tbl <- as_tibble(fhir_endpoints_con)
  fhir_endpoints_info_tbl <- as_tibble(fhir_endpoints_info_con)
  fhir_endpoint_totals <- list(
    "all_endpoints"     = nrow(fhir_endpoints_tbl),
    "indexed_endpoints" = nrow(fhir_endpoints_info_tbl %>% filter(http_response != 0)),
    "nonindexed_endpoints" = nrow(fhir_endpoints_tbl) - nrow(fhir_endpoints_info_tbl %>% filter(http_response != 0))
  )
}


get_response_tally <- function(fhir_endpoints_info_con) {
  # get the endpoint tally by http_response received 
  curr_http_response_tally <- as_tibble(fhir_endpoints_info_con) %>%
    select(http_response) %>%
    group_by(http_response) %>%
    tally()

  # Get the list of most recent HTTP responses when requesting the capability statement from the 
  # fhir_endpoints 
  response_tally <- list(
    "http_200" = max((curr_http_response_tally %>% filter(http_response==200))$n,0),
    "http_404" = max((curr_http_response_tally %>% filter(http_response==404))$n,0),
    "http_503" = max((curr_http_response_tally %>% filter(http_response==503))$n,0)
  )
}