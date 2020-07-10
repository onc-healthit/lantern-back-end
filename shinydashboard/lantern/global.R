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

root <- ifelse(Sys.getenv("HOME") == "/home/shiny", ".", "lantern")
config_yaml <- yaml::read_yaml(here(root, "configuration.yml"))
purrr::walk(config_yaml$libraries, library, character.only = T)
purrr::walk(config_yaml$function_files, source)
purrr::walk(config_yaml$module_files, source)

version_string <- read_file("../../version.txt")
version_number <- strsplit(version_string, "=")[[1]][2]

# Load table of http response codes and descriptions
http_response_code_tbl <-
  read_csv(here(root, "http_codes.csv"), col_types = cols(code = "i")) %>%
  mutate(code_chr = as.character(code))

# Define magic numbers for user interface
ui_special_values <- list(
  "ALL_FHIR_VERSIONS" = "99",
  "ALL_VENDORS" = "99"
  )
