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
devbanner <- Sys.getenv("LANTERN_BANNER_TEXT")

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

# define global app_data which is computed at application startup, and 
# refreshed at interval specified by refresh_timeout_minutes in configuration.yml
app_data <<- list(
  "fhir_version_list",           # list of fhir_versions reported by endpoints
  "fhir_endpoint_totals",        # count of endpoints, indexed and nonindexed
  "response_tally",              # counts of http responses
  "http_pct",                    # percentage of http responses for each endpoint
  "http_pctf",                   # http percentages with status as factors for graphing
  "http_summary",                # counts of all http_responses ever
  "vendor_count_tbl",            # endpoint counts by vendor
  "endpoint_resource_types",     # Resource types from capability statement by endpoint
  "capstat_fields",              # fields from the capability statement
  "last_updated",                # time app_data was last updated
  "avg_response_time",           # mean response time for endpoints by refresh period
  "vc_totals",                   # counts of endpoints by vendor
  "security_endpoints",          # security auth types supported by each endpoint
  "security_endpoints_tbl",      # list of endpoints filterable by auth type
  "auth_type_counts",            # count and pct of endpoints by auth type and fhir_version
  "endpoint_security_counts",    # summary table of endpoint counts with security resource in cap statement
  "security_code_list",          # list of supported auth types for UI dropdown
  "smart_response_capabilities", # smart core capabilities by endpoint, vendor, fhir_version
  "well_known_endpoints_tbl",    # endpoints returning smart core capabilities JSON doc
  "well_known_endpoints_no_doc", # well known endpoints reached, but no JSON doc returned
  "well_known_endpoint_counts"   # summary table of well known URI endpoints
)
app_data$fhir_endpoint_totals = get_endpoint_totals_list(db_tables)
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

  app_data$capstat_fields <<- get_capstat_fields(db_connection)

  app_data$last_updated <<- now()

  app_data$avg_response_time <<- get_avg_response_time(db_connection)

  app_data$vc_totals <<- app_data$vendor_count_tbl %>%
    filter(!(vendor_name == "Unknown")) %>%
    group_by(vendor_name) %>%
    summarise(total = sum(n))

  app_data$security_endpoints <<- get_security_endpoints(db_connection)

  app_data$security_endpoints_tbl <<- get_security_endpoints_tbl(db_connection)

  app_data$auth_type_counts <<- get_auth_type_count(app_data$security_endpoints)

  app_data$endpoint_security_counts <<- get_endpoint_security_counts(db_connection)

  app_data$security_code_list <<- app_data$security_endpoints %>%
    distinct(code) %>%
    pull(code)
  
  app_data$smart_response_capabilities <<- get_smart_response_capabilities(db_connection)

  app_data$well_known_endpoints_tbl    <<- get_well_known_endpoints_tbl(db_connection)
  app_data$well_known_endpoints_no_doc <<- get_well_known_endpoints_no_doc(db_connection)
  app_data$well_known_endpoint_counts  <<- get_well_known_endpoint_counts(db_connection)
})

onStop(function() {
  updater$suspend()
})
