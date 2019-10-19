package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// TODO: configuration file or commandline arguments
const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "H0p0np0p"
	dbname   = "postgres"
	sslmode  = "disable"
)

const dbSetup = `CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE fhir_endpoints (
    id                      SERIAL PRIMARY KEY,
    url                     VARCHAR(500) UNIQUE,
    fhir_version            VARCHAR(500),
    authorization_standard  VARCHAR(500),
    location                JSONB, -- location of IP address from ipstack.com.
    metadata                JSONB,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE provider_organizations ( -- https://data.medicare.gov/Hospital-Compare/Hospital-General-Information/xubh-q36u. Group practices:  https://data.medicare.gov/Physician-Compare/Physician-Compare-National-Downloadable-File/mj5m-pzi6 could get each group practice and address from this if can’t find a better data source
    id                      SERIAL PRIMARY KEY,
    name                    VARCHAR(500),
    url                     VARCHAR(500),
    location                JSONB,
    organization_type       VARCHAR(500), -- hospital or group practice
    hospital_type           VARCHAR(500), -- hospital type
    ownership               VARCHAR(500), -- hospital ownership
    beds                    INTEGER, -- hospital. can help show relative size. This is in https://hifld-geoplatform.opendata.arcgis.com/datasets/hospitals/
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE healthit_products (
    id                      SERIAL PRIMARY KEY,
    name                    VARCHAR(500),
    version                 VARCHAR(500),
    developer               VARCHAR(500),
    location                JSONB,
    authorization_standard  VARCHAR(500),
    api_syntax              VARCHAR(500),
    api_url                 VARCHAR(500),
    certification_criteria  JSONB,
    certification_status    VARCHAR(500),
    certification_date      DATE,
    certification_edition   VARCHAR(500),
    last_modified_in_chpl   DATE,
    chpl_id                 VARCHAR(500),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT healthit_product_info UNIQUE(name, version)
);

CREATE TRIGGER set_timestamp_fhir_endpoints
BEFORE UPDATE ON fhir_endpoints
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

CREATE TRIGGER set_timestamp_provider_organizations
BEFORE UPDATE ON provider_organizations
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

CREATE TRIGGER set_timestamp_healthit_products
BEFORE UPDATE ON healthit_products
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();
`

var db *sql.DB

func main() {
	//var endpoint models.FHIREndpoint
	var err error
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// calling db.Ping to create a connection to the database.
	// db.Open only validates the arguments, it does not create the connection.
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")
}
