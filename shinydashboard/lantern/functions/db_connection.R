# Database connection functions
library(RPostgres)
# Read database connection information from .Renviron file
# If doing local development: you can readRenviron("../.env")
# and set the db_config$host = "localhost"

db_config <- list("dbname" = Sys.getenv("LANTERN_DBNAME"),
                  "host" = Sys.getenv("LANTERN_DBHOST"),
                  "port" = Sys.getenv("LANTERN_DBPORT"),
                  "user" = Sys.getenv("LANTERN_DBUSER"),
                  "password" = Sys.getenv("LANTERN_DBPASSWORD")
)

db_config$host <- ifelse(Sys.getenv("HOME") == "/home/shiny", db_config$host, "localhost")

# Connect to the Lantern database
db_connection <-
  dbConnect(
    RPostgres::Postgres(),
    dbname = db_config$dbname,
    host = db_config$host, # i.e. 'ec2-54-83-201-96.compute-1.amazonaws.com'
    port = db_config$port,
    user = db_config$user,
    password = db_config$password
)

# Make connections to the various lantern tables
db_tables <- list(
  fhir_endpoints              = tbl(db_connection, "fhir_endpoints"),
  fhir_endpoints_info         = tbl(db_connection, "fhir_endpoints_info"),
  fhir_endpoints_info_history = tbl(db_connection, "fhir_endpoints_info_history"),
  end_org                     = tbl(db_connection, "endpoint_organization"),
  hit_prod                    = tbl(db_connection, "healthit_products"),
  endpoint_export             = tbl(db_connection, "endpoint_export"),
  vendors                     = tbl(db_connection, "vendors")
)

# Get the Endpoint export table and clean up for UI
endpoint_export_tbl <- db_tables$endpoint_export %>%
  collect() %>%
  mutate(vendor_name = na_if(vendor_name, "")) %>%
  tidyr::replace_na(list(vendor_name = "Unknown")) %>%
  tidyr::replace_na(list(fhir_version = "Unknown"))
