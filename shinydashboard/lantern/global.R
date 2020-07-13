# Any code in this file is guaranteed to be called before either
# ui.R or server.R

#
# Lantern metrics dashboard
# This Shiny application will display metrics on FHIR endpoints as
# monitored by the Lantern application.
#
library(here)
library(yaml)
library(config)
library(shiny)
library(shinydashboard)
library(tidyverse)
shinyOptions(cache = memoryCache(max_size = 20e6, max_age = 3600))

root <- ifelse(Sys.getenv("HOME") == "/home/shiny", ".", "lantern")
config_yaml <- yaml::read_yaml(here(root, "configuration.yml"))
purrr::walk(config_yaml$libraries, library, character.only = T)
purrr::walk(config_yaml$function_files, source)
purrr::walk(config_yaml$module_files, source)

version_string <- read_file("VERSION")
version_number <- strsplit(version_string, "=")[[1]][2]
version_title <- paste("Version ", version_number)

# Define magic numbers for user interface
ui_special_values <- list(
  "ALL_FHIR_VERSIONS" = "All FHIR Versions",
  "ALL_VENDORS" = "All Vendors"
  )

# The list of fhir versions and vendors are unlikely to change during a user's session
# we'll update them on timer, but not refresh the UI
app <<- list(
  fhir_version_list      = get_fhir_version_list(endpoint_export_tbl),
  vendor_list            = get_vendor_list(endpoint_export_tbl),
  http_response_code_tbl =
    read_csv(here(root, "http_codes.csv"), col_types = cols(code = "i")) %>%
    mutate(code_chr = as.character(code))
)

app_data <<- list(
  fhir_endpoint_totals    = get_endpoint_totals_list(db_tables),
  response_tally          = get_response_tally_list(db_tables),
  http_pct                = get_http_response_summary_tbl(db_tables),
  endpoint_resource_types = get_fhir_resource_types(db_connection),
  last_updated            = now()
)

# we need a table with the code as a factor for use in ggplot
app_data$http_pctf <- app_data$http_pct %>%
    filter(http_response > 0, http_response != 200) %>%
    mutate(name = as.factor(as.character(id)), code_f = as.factor(code))

app_data$http_summary <- app_data$http_pct %>%
    left_join(app$http_response_code_tbl, by = c("code" = "code_chr")) %>%
    select(id, code, label) %>%
    group_by(code, label) %>%
    summarise(count = n())
  
app_data$vendor_count_tbl <- get_fhir_version_vendor_count(endpoint_export_tbl)
  
app_data$vc_totals <- app_data$vendor_count_tbl %>%
    filter(!(vendor_name == "Unknown")) %>%
    group_by(vendor_name) %>%
    summarise(total = sum(n))

# Define observer based on a refresh_timeout to refetch data from the database
updater <- observe({

  invalidateLater(config_yaml$refresh_timeout_minutes * 60 * 1000) # convert minutes to milliseconds

  app$fhir_version_list <<- get_fhir_version_list(endpoint_export_tbl)

  app_data$fhir_endpoint_totals <<- get_endpoint_totals_list(db_tables)

  app_data$response_tally <<- get_response_tally_list(db_tables)

  app_data$http_pct <<- get_http_response_summary_tbl(db_tables)

  app_data$http_pctf <<- app_data$http_pct %>%
    filter(http_response > 0, http_response != 200) %>%
    mutate(name = as.factor(as.character(id)), code_f = as.factor(code))

  app_data$http_pctf <<- app_data$http_pct %>%
    filter(http_response > 0, http_response != 200) %>%
    mutate(name = as.factor(as.character(id)), code_f = as.factor(code))

  app_data$http_summary <<- app_data$http_pct %>%
    left_join(app$http_response_code_tbl, by = c("code" = "code_chr")) %>%
    select(id, code, label) %>%
    group_by(code, label) %>%
    summarise(count = n())

  app_data$vendor_count_tbl <<- get_fhir_version_vendor_count(endpoint_export_tbl)

  app_data$endpoint_resource_types <<- get_fhir_resource_types(db_connection)

  app_data$last_updated <<- now()

})

onStop(function() {
  updater$suspend()
})
