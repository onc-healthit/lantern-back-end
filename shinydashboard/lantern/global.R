# Any code in this file is guaranteed to be called before either
# ui.R or server.R

#
# Lantern metrics dashboard
# This Shiny application will display metrics on FHIR endpoints as 
# monitored by the Lantern application. 
# 

config_yaml <- yaml::read_yaml("configuration.yml")
purrr::walk(config_yaml$libraries, library, character.only = T)
purrr::walk(config_yaml$function_files, source)
purrr::walk(config_yaml$module_files, source)

# Load table of http response codes and descriptions
root <- ifelse(Sys.getenv("HOME")=='/home/shiny',".","lantern")
http_response_code_tbl <- read_csv(here(root,"http_codes.csv")) %>% mutate(code_chr=as.character(code))





# we want the current set of http response codes from the endpoint monitoring
# first get the entries from the metrics_labels table for http_request_responses
http_response_ids <- metrics_labels %>%
  filter(metric_name == "AllEndpoints_http_request_responses") %>%
  select(id)

# next, right_join against the value for each endpoint
http_response_values <- metrics_values %>%
  right_join(http_response_ids, by = c("labels_id" = "id"))

# Compute the percentage of each response code for all responses received
http_pct <- as_tibble(http_response_values %>% 
                        mutate(code=as.character(value)) %>%
                        group_by(labels_id,code,value) %>% 
                        summarise(Percentage=n()) %>% 
                        group_by(labels_id) %>% 
                        mutate(Percentage=Percentage/sum(Percentage,na.rm = TRUE)*100)
)
# we want to graph all non-200 results by response code, but they need to be factors
# so they can be shown as separate categories on the graph, rather than as a scalar value
http_pctf <- http_pct %>% 
  filter(value != 200) %>% 
  mutate(name=as.factor(labels_id), code=as.factor(code)) 

# create a summary table to show the response codes received along with 
# the description for each code
http_summary <- http_pct %>%
  left_join(http_response_code_tbl, by=c("code" = "code_chr")) %>%
  select(code,label) %>%
  group_by("HTTP Response" = code,"Status"=label) %>%
  summarise(Count=n()) 

# Get the FHIR version for each endpoint
fhir_version_tbl <- as_tibble(tbl(con,sql("select id,url,vendor,capability_statement->>'fhirVersion' as FHIR from fhir_endpoints_info where capability_statement->>'fhirVersion' IS NOT NULL")))

# Get the count of endpoints by vendor, and use "Unknown" for any entries
# where the vendor field is empty
fhir_version_vendor_count <- fhir_version_tbl %>%
  mutate(vendor = na_if(vendor,"")) %>%
  tidyr::replace_na(list(vendor="Unknown")) %>%
  group_by(vendor,fhir) %>%
  tally() %>%
  select(Vendor=vendor,"FHIR Version"=fhir,"Count"=n)

# Get the list of distinct fhir versions for use in filtering
fhir_version_list <- as.list(fhir_version_tbl %>% distinct("FHIR Version"=fhir))

# Get the list of distinct vendors for use in filtering
vendor_list <- as.list(as_tibble(fhir_endpoints_info %>% distinct(vendor)) %>% mutate(vendor = na_if(vendor,"")) %>% tidyr::replace_na(list(vendor="Unknown")) %>% pull(vendor))

