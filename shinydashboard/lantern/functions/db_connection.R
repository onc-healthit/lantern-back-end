# Database connection functions

# Read database connection information from .Renviron file
# If doing local development: you can readRenviron("../.env")
# and set the db_config$host = "localhost"

db_config <- list("dbname" = Sys.getenv("LANTERN_DBNAME"),
                  "host" = Sys.getenv("LANTERN_DBHOST"),
                  "port" = Sys.getenv("LANTERN_DBPORT"),
                  "user" = Sys.getenv("LANTERN_DBUSER"),
                  "password" = Sys.getenv("LANTERN_DBPASSWORD")
)

# Connect to the Lantern database
con <- dbConnect(RPostgres::Postgres(),
                 dbname = db_config$dbname, 
                 host = db_config$host, # i.e. 'ec2-54-83-201-96.compute-1.amazonaws.com'
                 port = db_config$port, 
                 user = db_config$user,
                 password = db_config$password
)

# Make connections to the various lantern tables
fhir_endpoints      <- tbl(con, "fhir_endpoints")
fhir_endpoints_info <- tbl(con, "fhir_endpoints_info")
fhir_endpoints_info_history <- tbl(con,"fhir_endpoints_info_history")
metrics_values      <- tbl(con, "metrics_values")
metrics_labels      <- tbl(con, "metrics_labels")
end_org             <- tbl(con, "endpoint_organization")
hit_prod            <- tbl(con, "healthit_products")
endpoint_export     <- tbl(con, "endpoint_export")
vendors             <- tbl(con, "vendors")

# Get the Endpoint export table and clean up for UI
endpoint_export_tbl <- as_tibble(endpoint_export) %>%
  mutate(vendor_name = na_if(vendor_name,"")) %>%
  tidyr::replace_na(list(vendor_name="Unknown")) %>%
  tidyr::replace_na(list(fhir_version="Unknown"))