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

root <- ifelse(Sys.getenv("HOME")=='/home/shiny',".","lantern")
config_yaml <- yaml::read_yaml(here(root,"configuration.yml"))
purrr::walk(config_yaml$libraries, library, character.only = T)
purrr::walk(config_yaml$function_files, source)
purrr::walk(config_yaml$module_files, source)

# Load table of http response codes and descriptions
http_response_code_tbl <- read_csv(here(root,"http_codes.csv")) %>% mutate(code_chr=as.character(code))

# Get the list of distinct fhir versions for use in filtering
fhir_version_list <- as.list(endpoint_export_tbl %>%
                               arrange(fhir_version) %>%
                               distinct("FHIR Version"=fhir_version))

# Get the list of distinct vendors for use in filtering
vendor_list <- as.list(endpoint_export_tbl %>% distinct(vendor_name) %>% arrange(vendor_name) %>% pull(vendor_name))

