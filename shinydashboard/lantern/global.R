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
purrr::walk(config_yaml$libraries, library, character.only = TRUE)
purrr::walk(config_yaml$function_files, source)
purrr::walk(config_yaml$module_files, source)

version_string <- read_file("VERSION")
version_number <- strsplit(version_string, "=")[[1]][2]
version_title <- paste("Version ", version_number)
devbanner <- Sys.getenv("LANTERN_BANNER_TEXT")
qry_interval_seconds <- (strtoi(Sys.getenv("LANTERN_CAPQUERY_QRYINTVL")) * 60)

validation_group_list <- fromJSON(here(root, "validation_groups.json"))
validation_rules_descriptions <- fromJSON(here(root, "rule_descriptions.json"))
validation_group_names <- names(validation_group_list)

valid_fhir_versions <- c("No Cap Stat", "0.4.0", "0.5.0", "1.0.0", "1.0.1", "1.0.2", "1.1.0", "1.2.0", "1.4.0", "1.6.0", "1.8.0", "3.0.0", "3.0.1", "3.0.2", "3.2.0", "3.3.0", "3.5.0", "3.5a.0", "4.0.0", "4.0.1", "4.1.0", "4.3.0", "4.2.0", "4.4.0", "4.5.0", "4.6.0", "5.0.0")

# Normalize FHIR version strings to 3-part format (e.g., "4.4" -> "4.4.0", "4" -> "4.0.0")
# Vectorized to work inside dplyr::mutate
normalize_fhir_version <- function(version) {
  dplyr::case_when(
    is.na(version) | version == "" | version == "No Cap Stat" | version == "Unknown" ~ version,
    !grepl("\\.", version) ~ paste0(version, ".0.0"),
    !grepl("\\..+\\.", version) ~ paste0(version, ".0"),
    TRUE ~ version
  )
}

# Expand normalized 3-part FHIR versions back to include short forms for DB filtering
# e.g., c("4.4.0") -> c("4.4.0", "4.4"), c("4.0.0") -> c("4.0.0", "4.0", "4")
expand_fhir_versions_for_db <- function(versions) {
  expanded <- versions
  for (v in versions) {
    if (is.na(v) || v %in% c("", "No Cap Stat", "Unknown")) next
    parts <- strsplit(v, "\\.")[[1]]
    if (length(parts) == 3) {
      if (parts[3] == "0") {
        expanded <- c(expanded, paste(parts[1], parts[2], sep = "."))
      }
      if (parts[2] == "0" && parts[3] == "0") {
        expanded <- c(expanded, parts[1])
      }
    }
  }
  unique(expanded)
}

dstu2 <- c("0.4.0", "0.4", "0.5.0", "0.5", "1.0.0", "1.0", "1", "1.0.1", "1.0.2")
stu3 <- c("1.1.0", "1.1", "1.2.0", "1.2", "1.4.0", "1.4", "1.6.0", "1.6", "1.8.0", "1.8", "3.0.0", "3.0", "3", "3.0.1", "3.0.2")
r4 <- c("3.2.0", "3.2", "3.3.0", "3.3", "3.5.0", "3.5", "3.5a.0", "4.0.0", "4.0", "4", "4.0.1")
r4b <- c("4.1.0", "4.1", "4.3.0", "4.3")
r5 <- c("4.2.0", "4.2", "4.4.0", "4.4", "4.5.0", "4.5", "4.6.0", "4.6", "5.0.0", "5.0", "5")

# Define magic numbers for user interface
ui_special_values <- list(
  "ALL_DEVELOPERS" = "All Developers",
  "ALL_RESOURCES" = "All Resources",
  "ALL_PROFILES" = "All Profiles"
)

# HTTP response codes and their labels. Static, so it is read once at startup rather than by the
# data refresh.
http_response_codes <- read_csv(here(root, "http_codes.csv"), col_types = cols(code = "i")) %>%
  mutate(code_chr = as.character(code))
