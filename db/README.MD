# Database Setup Scripts
The `.sql` files located in db/sql are mounted to the postgres service `docker-entrypoint-initdb.d` directory in the `docker-compose.yml` and `docker-compose-override.yml` files. From the postgres Docker image documentation https://hub.docker.com/_/postgres "If you would like to do additional initialization in an image derived from this one, add one or more *.sql, *.sql.gz, or *.sh scripts under /docker-entrypoint-initdb.d (creating the directory if necessary). After the entrypoint calls initdb to create the default postgres user and database, it will run any *.sql files, run any executable *.sh scripts, and source any non-executable *.sh scripts found in that directory to do further initialization before starting the service."

Changes to the database (new tables, new columns on exiisting tables etc..) should exist in `db/dbsetup.sql` as this file will setup the database from scratch.

# Database Migration
If changes are made to the database setup file `db/dbsetup.sql` then there should also be a migration script defined that when run will migrate an existing database to be compatible with the changes to `db/dbsetup.sql`.

Migration scripts should be placed in `db/migration/migratrions`. For more information about how to write migration scripts see the tutorial here https://github.com/golang-migrate/migrate/blob/master/database/postgres/TUTORIAL.md

## Migrate Down
If you find yourself in the unlikely scenario of needing to undo migrations, by way of the `down.sql` scripts in the `db/migration/migrations` directory, you can change the line that reads `m.Steps(1)` in `db/migration/cmd/main.go` to `m.Steps(-1)` and run `make migrate_database` from the base directory.


# Database Schema

## fhir_endpoints table
The fhir_ endpoints table is the master list of FHIR service-base URLs that will be queried.
| Field        | Type           | Description  |
| ------------- |:-------------:| -----:|
| id     | INTEGER | Database ID of endpoint |
| url     | VARCHAR(500)      |   Service base URL of endpoint |
| organization_names | VARCHAR(500)[]      |    Organization names this endpoint is reported as |
| npi_ids | VARCHAR(500)[]      |    NPI IDs of matching organizations |
| list_source | VARCHAR(500)      |   Name/URL of list source that the ednpoint came from |
| created_at | TIMESTAMPTZ      |    Timestamp of creation |
| updated_at | TIMESTAMPTZ      |    Timestamp of last update |

## npi_contacts table
The entries in this table are from the endpoints file in the NPPES monthly data dissemination package.
| Field        | Type           | Description  |
| ------------- |:-------------:| -----:|
| id     | INTEGER | Database ID of endpoint |
| npi_id     | VARCHAR(500) | NPI ID of the NPPES contact entry |
| endpoint_type     | VARCHAR(500) | Endpoint type in NPPES, should always equal "FHIR" |
| endpoint_type_description     | VARCHAR(500) | Endpoint type description in NPPES, should always equal "FHIR_URL"  |
| endpoint     | VARCHAR(500) | Service base URL of FHIR endpoint |
| valid_url     | BOOLEAN | Whether or not the endpoint field is formatted as a URL  |
| affiliation     | VARCHAR(500) | Whether or not the endpoint is affiliated to organizations other than what is listed in this entry. Values are 'Y' or 'N' |
| endpoint_description     | VARCHAR(500) | Brief description of endpoint by registerer |
| affiliation_legal_business_name     | VARCHAR(500) | Legal busiiness name of the affiliated organization |
| use_code     | VARCHAR(500) | Use code of the FHIR endpoint, values include "HIE", "DIRECT", "OTHER" or empty |
| use_description     | VARCHAR(500) | Description of the use code, values include "Health Information Exchange (HIE)", "Direct", "Other or empty |
| other_use_description     | VARCHAR(500) | Free text field describing other use of endpoint |
| content_type     | VARCHAR(500) | Description of content at endpoint, values are "CSV", "Other" or empty |
| content_description     | VARCHAR(500) | Free text field describing content hosted at endpoint |
| other_content_description     | VARCHAR(500) | database ID of endpoint |
| location     | JSONB | JSON of the location of the endpoint affiliated organization |
| created_at | TIMESTAMPTZ      |    Timestamp of creation |
| updated_at | TIMESTAMPTZ      |    Timestamp of last update |

## fhir_endpoints_info table
The fhir_endpoints_info table contains the data collected from the last query of the FHIR endpoint at `url` and represents the most up to date information
| Field        | Type           | Description  |
| ------------- |:-------------:| -----:|
| id     | integer | Database ID of endpoint |
| healthit_product_id     | INTEGER | Healthit product ID of endpoint |
| vendor_id     | INTEGER | Vendor ID of endpoint |
| url     | VARCHAR(500)      |   Service base URL of endpoint |
| tls_version     | VARCHAR(500)      |   Transport Layer Security (TLS) version of endpoint |
| mime_types | VARCHAR(500)[]      |    MIME types this endpoint supports |
| http_response     | INTEGER | HTTP response receieved from endpoint metadata url |
| errors     | VARCHAR(500)      |   Errors receieved from querying endpoint  |
| capability_statement     | JSONB      |   Capability statement receieved from endpoint |
| validation     | JSONB      |   Validation information for ONC conformance criteria |
| included_fields | JSONB      |    Structure that shows which capability statement fields and extensions are supported/unsupported by endpoint |
| supported_resources | VARCHAR(500)[]      |    Stores all the FHIR resources the endpoint supports |
| response_time_seconds | DECIMAL(64,4)     |    HTTP response time of endpoint |
| created_at | TIMESTAMPTZ      |    Timestamp of creation |
| updated_at | TIMESTAMPTZ      |    Timestamp of last update |
| smart_http_response     | INTEGER | HTTP response receieved from endpoint SMART url |
| smart_response     | JSONB      |   SMART response receieved from endpoint|
| availability     | DECIMAL(64,4)      |   All-time availability percentage. The number of total HTTP 200 responses that have ever been received from this endpoint divided by the total number of HTTP request attempts|

## fhir_endpoints_info_history table
The fhir_endpoints_info_history table contains the history of the fhir_endpoints_info table. The operation field of the fhir_endpoints_info_history table represents if the entry was inserted for the first time (I) ie: The first query ever performed at the given `url`, if the information retrieved from querying the `url` for an existing info entry was updated in any way (U) or if the info entry was removed (D). Deletion occurs in the case where a URL was once in a vendor list and was being queried by Lantern, but no longer exists in a vendor list and therefore will no longer exist in the `fhir_endpoints` table and will no longer be queried.
| Field        | Type           | Description  |
| ------------- |:-------------:| -----:|
| operation     | CHAR(1) | Entry operation (I for insert, U for Update, D for Delete)  |
| entered_at | TIMESTAMPTZ      |    Timestamp of creation |
| user_id     | VARCHAR(500) | database user ID |
| id     | INTEGER | database ID of endpoint |
| healthit_product_id     | INTEGER | healthit product ID of endpoint |
| vendor_id     | INTEGER | vendor ID of endpoint |
| url     | VARCHAR(500)      |   Service base URL of endpoint |
| tls_version     | VARCHAR(500)      |   Transport Layer Security (TLS) version of endpoint |
| mime_types | VARCHAR(500)[]      |    MIME types this endpoint supports |
| http_response     | INTEGER | HTTP response receieved from endpoint metadata url |
| errors     | VARCHAR(500)      |   Errors receieved from querying endpoint  |
| capability_statement     | JSONB      |   Capability statement receieved from endpoint |
| validation     | JSONB      |   Validation information for ONC conformance criteria |
| included_fields | JSONB      |    Structure that shows which capability statement fields and extensions are supported/unsupported by endpoint |
| supported_resources | VARCHAR(500)[]      |    Stores all the FHIR resources the endpoint supports |
| response_time_seconds | DECIMAL(64,4)     |    HTTP response time of endpoint |
| created_at | TIMESTAMPTZ      |    Timestamp of creation |
| updated_at | TIMESTAMPTZ      |    Timestamp of last update |
| smart_http_response     | INTEGER | HTTP response receieved from endpoint SMART url |
| smart_response     | JSONB      |   SMART response receieved from endpoint|
| availability     | DECIMAL(64,4)      |   All-time availability percentage. The number of total HTTP 200 responses that have ever been received from this endpoint divided by the total number of HTTP request attempts|


## vendor table
The vendor table stores health IT product vendor information gathered from CHPL.
| Field        | Type           | Description  |
| ------------- |:-------------:| -----:|
| id     | INTEGER | database ID of Vendor |
| name     | VARCHAR(500)      |   Name of vendor |
| developer_code     | VARCHAR(500)      |   Vendor developer code |
| url     | VARCHAR(500)      |   URL of vendor website |
| location     | JSONB      |   Vendor site locations |
| status     | VARCHAR(500)       |   Vendor status active or not |
| last_modified_in_chpl | TIMESTAMPTZ      |    Timestamp of last time vendor modified in CHPL |
| chpl_id | INTEGER      |    Vendor CHPL ID |
| created_at | TIMESTAMPTZ      |    Timestamp of creation |
| updated_at | TIMESTAMPTZ      |    Timestamp of last update |

## healthit_products table
The healthit_products table stores health IT product information gathered from CHPL.
| Field        | Type           | Description  |
| ------------- |:-------------:| -----:|
| id     | INTEGER | Database ID of healthit product |
| name     | VARCHAR(500)      |   Name of healthit product |
| version     | VARCHAR(500)      |   Version of healthit product |
| vendor_id     | INTEGER |  ID of healthit product vendor |
| location     | JSONB      |   Healthit product locations |
| authorization_standard     | VARCHAR(500)      |   Healthit product authorization standard |
| api_syntax    | VARCHAR(500)      |   Syntax of healthit product API |
| api_url    | VARCHAR(500)      |   URL of healthit product API |
| certification_criteria     | JSONB      |   Healthit product certification criteria |
| certification_status     | VARCHAR(500)      |   Status of healthit product certification |
| certification_date    | DATE      |   Date of healthit product certification |
| certification_edition     | VARCHAR(500)      |   healthit product certification edition |
| last_modified_in_chpl | DATE     |    Timestamp of last time healthit product modified in CHPL |
| chpl_id | VARCHAR(500)      |    Healthit product CHPL ID |
| created_at | TIMESTAMPTZ      |    Timestamp of creation |
| updated_at | TIMESTAMPTZ      |    Timestamp of last update |

## certification_criteria table
The certification_criteria table stores certification criteria information gathered from CHPL.
| Field        | Type           | Description  |
| ------------- |:-------------:| -----:|
| id | INTEGER | Database ID of certification criteria |
| certification_id | INTEGER | ID of the certification criteria saved in CHPL |
|	cerification_number | VARCHAR(500) | The name of the certification criteria |
|	title | VARCHAR(500) | The title of the certification criteria |
| certification_edition_id  | INTEGER | The ID of the certification edition the criteria was defined in |
|	certification_edition | VARCHAR(500) | The certification edition of the criteria |
|	description | VARCHAR(500) | Description of the certification criteria |
|	removed | BOOLEAN | Whether the criteria has been removed |
| created_at | TIMESTAMPTZ | Timestamp of creation |
| updated_at | TIMESTAMPTZ | Timestamp of last update |

## product_criteria table
The product criteria table maps the Health IT product Database ID to entries in the certification_criteria based on the product's `certification_criteria` field.
| Field        | Type           | Description  |
| ------------- |:-------------:| -----:|
| healthit_product_id | INTEGER | Database ID of healthit product |
| certification_id | INTEGER | ID of the certification criteria saved in CHPL |
| certification_number | VARCHAR(500) | The name of the certification criteria |
| created_at | TIMESTAMPTZ | Timestamp of creation |
| updated_at | TIMESTAMPTZ | Timestamp of last update |

# fhir_endpoints_availability table
The FHIR endpoints availability table keeps track of how many HTTP 200 responses have been received from a URL and how many total responses have been received to cache the availability to reduce the burden of having to calculate the availability of a given endpoint.
| Field        | Type           | Description  |
| ------------- |:-------------:| -----:|
| url     | VARCHAR(500)      |   Service base URL of endpoint |
| http_200_count | BIGINT | Count of HTTP 200 responses ever received from endpoint |
| http_all_count | BIGINT | Total count of all HTTP requests sent to the endpoint |
| created_at | TIMESTAMPTZ | Timestamp of creation |
| updated_at | TIMESTAMPTZ | Timestamp of last update |