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
library(shinybusy)
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
qry_interval_seconds <- (strtoi(Sys.getenv("LANTERN_CAPQUERY_QRYINTVL")) * 60)
database_fetch <- reactiveVal(0)

validation_group_list <- fromJSON(here(root, "validation_groups.json"))
validation_group_names <- names(validation_group_list)

# Define magic numbers for user interface
ui_special_values <- list(
  "ALL_FHIR_VERSIONS" = "All FHIR Versions",
  "ALL_DEVELOPERS" = "All Developers",
  "ALL_RESOURCES" = "All Resources"
)

# The list of fhir versions and vendors are unlikely to change during a user's session
# we'll update them on timer, but not refresh the UI
app <<- list(
  fhir_version_list      = reactiveVal(get_fhir_version_list(endpoint_export_tbl)),
  vendor_list            = get_vendor_list(endpoint_export_tbl),
  http_response_code_tbl =
    read_csv(here(root, "http_codes.csv"), col_types = cols(code = "i")) %>%
    mutate(code_chr = as.character(code)),
  zip_to_zcta =
    read_csv(here(root, "zipcode_zcta.csv"), col_types = cols(zipcode = "c", zcta = "c"))
)

# define global app_data which is computed at application startup, and
# refreshed at interval specified by refresh_timeout_minutes in configuration.yml
app_data <<- list(
  fhir_endpoint_totals = reactiveVal(NULL),        # count of endpoints, indexed and nonindexed
  response_tally = reactiveVal(NULL),              # counts of http responses
  http_pct = reactiveVal(NULL),                    # percentage of http responses for each endpoint
  vendor_count_tbl = reactiveVal(NULL),            # endpoint counts by vendor
  endpoint_resource_types = reactiveVal(NULL),     # Resource types from capability statement by endpoint
  capstat_fields = reactiveVal(NULL),              # fields from the capability statement
  capstat_fields_list = reactiveVal(NULL),         # the list of fields we keep track of in a capability statement
  capstat_values = reactiveVal(NULL),              # values of specific fields from the capability statement
  last_updated = reactiveVal(NULL),                # time app_data was last updated
  security_endpoints = reactiveVal(NULL),          # security auth types supported by each endpoint
  security_endpoints_tbl = reactiveVal(NULL),      # list of endpoints filterable by auth type
  auth_type_counts = reactiveVal(NULL),            # count and pct of endpoints by auth type and fhir_version
  endpoint_security_counts = reactiveVal(NULL),    # summary table of endpoint counts with security resource in cap statement
  security_code_list = reactiveVal(NULL),          # list of supported auth types for UI dropdown
  smart_response_capabilities = reactiveVal(NULL), # smart core capabilities by endpoint, vendor, fhir_version
  well_known_endpoints_tbl = reactiveVal(NULL),    # endpoints returning smart core capabilities JSON doc
  well_known_endpoints_no_doc = reactiveVal(NULL), # well known endpoints reached, but no JSON doc returned
  endpoint_locations = reactiveVal(NULL),          # endpoints with location information mappings
  implementation_guide = reactiveVal(NULL),        # implementation_guide table
  capstat_sizes_tbl = reactiveVal(NULL),           # capability statement size by vendor, fhir_version
  validation_tbl = reactiveVal(NULL)               # validation rules and results
)

# Define observer based on a refresh_timeout to refetch data from the database
updater <- observe({

  invalidateLater(config_yaml$refresh_timeout_minutes * 60 * 1000) # convert minutes to milliseconds
  # Database fetch is a reactive val that is set to 1 when the global app_data tables must be re-populated and is set to 0 when it is completed
  database_fetch(1)

})

onStop(function() {
  updater$suspend()
})
