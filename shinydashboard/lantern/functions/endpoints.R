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

get_http_response_summary <- function(fhir_endpoints_info_history) {
  # Compute the percentage of each response code for all responses received
  as_tibble(fhir_endpoints_info_history) %>%
            select(id,http_response) %>%
            mutate(code=as.character(http_response)) %>%
            group_by(id,code,http_response) %>% 
            summarise(Percentage=n()) %>%
            group_by(id) %>% 
            mutate(Percentage=Percentage/sum(Percentage,na.rm = TRUE)*100) %>%
            ungroup()
}

get_fhir_version_vendor_count <- function(endpoint_tbl) {
  # Get the count of endpoints by vendor
  endpoint_tbl %>%
    group_by(vendor_name,fhir_version) %>%
    tally() %>%
    select(Vendor=vendor_name,"FHIR Version"=fhir_version,"Count"=n)
}

# Get the list of distinct fhir versions for use in filtering
get_fhir_version_list <- function(endpoint_tbl) {
    as.list(endpoint_tbl %>%
            arrange(fhir_version) %>%
            distinct("FHIR Version"=fhir_version))
}

get_vendor_list <- function(endpoint_tbl) {
# Get the list of distinct vendors for use in filtering
  vendor_list <- as.list(endpoint_tbl %>%
                         distinct(vendor_name) %>%
                         arrange(vendor_name) %>%
                         pull(vendor_name))
}