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
  fhir_endpoints_metadata     = tbl(db_connection, "fhir_endpoints_metadata"),
  fhir_endpoints_info_history = tbl(db_connection, "fhir_endpoints_info_history"),
  end_org                     = tbl(db_connection, "endpoint_organization"),
  hit_prod                    = tbl(db_connection, "healthit_products"),
  endpoint_export             = tbl(db_connection, "endpoint_export"),
  organization_location       = tbl(db_connection, "organization_location"),
  vendors                     = tbl(db_connection, "vendors"),
  endpoint_export_mv          = tbl(db_connection, "endpoint_export_mv"),
  mv_endpoint_totals          = tbl(db_connection, "mv_endpoint_totals"),
  mv_vendor_fhir_counts       = tbl(db_connection, "mv_vendor_fhir_counts"),
  mv_response_tally           = tbl(db_connection, "mv_response_tally"),
  mv_contacts_info            = tbl(db_connection, "mv_contacts_info")
)

valid_fhir_versions <- c("No Cap Stat", "0.4.0", "0.5.0", "1.0.0", "1.0.1", "1.0.2", "1.1.0", "1.2.0", "1.4.0", "1.6.0", "1.8.0", "3.0.0", "3.0.1", "3.0.2", "3.2.0", "3.3.0", "3.5.0", "3.5a.0", "4.0.0", "4.0.1")
