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
purrr::walk(config_yaml$libraries, library, character.only = TRUE)
purrr::walk(config_yaml$function_files, source)
purrr::walk(config_yaml$module_files, source)

version_string <- read_file("VERSION")
version_number <- strsplit(version_string, "=")[[1]][2]
version_title <- paste("Version ", version_number)
devbanner <- Sys.getenv("LANTERN_BANNER_TEXT")
qry_interval_seconds <- (strtoi(Sys.getenv("LANTERN_CAPQUERY_QRYINTVL")) * 60)
database_fetch <- reactiveVal(0)

validation_group_list <- fromJSON(here(root, "validation_groups.json"))
validation_rules_descriptions <- fromJSON(here(root, "rule_descriptions.json"))
validation_group_names <- names(validation_group_list)

valid_fhir_versions <- c("No Cap Stat", "0.4.0", "0.5.0", "1.0.0", "1.0.1", "1.0.2", "1.1.0", "1.2.0", "1.4.0", "1.6.0", "1.8.0", "3.0.0", "3.0.1", "3.0.2", "3.2.0", "3.3.0", "3.5.0", "3.5a.0", "4.0.0", "4.0.1")

dstu2 <- c("0.4.0", "0.5.0", "1.0.0", "1.0.1", "1.0.2")
stu3 <- c("1.1.0", "1.2.0", "1.4.0", "1.6.0", "1.8.0", "3.0.0", "3.0.1", "3.0.2")
r4 <- c("3.2.0", "3.3.0", "3.5.0", "3.5a.0", "4.0.0", "4.0.1")

# Define magic numbers for user interface
ui_special_values <- list(
  "ALL_DEVELOPERS" = "All Developers",
  "ALL_RESOURCES" = "All Resources",
  "ALL_PROFILES" = "All Profiles"
)

# The list of fhir versions and vendors are unlikely to change during a user's session
# we'll update them on timer, but not refresh the UI
app <<- list(
  fhir_version_list_no_capstat      = reactiveVal(NULL),
  fhir_version_list      = reactiveVal(NULL),
  distinct_fhir_version_list_no_capstat      = reactiveVal(NULL),
  distinct_fhir_version_list      = reactiveVal(NULL),
  vendor_list            = reactiveVal(NULL),
  http_response_code_tbl = reactiveVal(NULL),
  zip_to_zcta = reactiveVal(NULL),
  endpoint_export_tbl = reactiveVal(NULL)
)


time_until_next_run <- function() {
  current_time <- Sys.time()
  message("current_time ", current_time)
  current_hour <- as.numeric(format(current_time, "%H"))
  current_minute <- as.numeric(format(current_time, "%M"))

  hours_until_2am <- ifelse(current_hour >= 6, 24 - current_hour + 6, 6 - current_hour)
  time_until_next_run <- (hours_until_2am * 60 * 60) - (current_minute * 60)
  message("time_until_next_run: ", time_until_next_run)
  return(time_until_next_run)
}

updater <- observe({
  time_until_next_run_value <- time_until_next_run()
  invalidateLater(time_until_next_run_value * 1000)
  database_fetch(1)
})

onStop(function() {
  updater$suspend()
})
