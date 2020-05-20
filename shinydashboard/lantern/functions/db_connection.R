# Database connection functions

# Read database connection information from .Renviron file
# local development: readRenviron("../.env"); db_config$host = "localhost"
# db_config <- config::get("lantern") 

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
metrics_values      <- tbl(con, "metrics_values")
metrics_labels      <- tbl(con, "metrics_labels")
end_org             <- tbl(con, "endpoint_organization")
hit_prod            <- tbl(con, "healthit_products")
endpoint_export     <- tbl(con, "endpoint_export")
vendors             <- tbl(con, "vendors")