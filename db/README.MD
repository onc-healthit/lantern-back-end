# Database Setup Scripts
The `.sql` files located in db/sql are mounted to the postgres service `docker-entrypoint-initdb.d` directory in the `docker-compose.yml` and `docker-compose-override.yml` files. From the postgres Docker image documentation https://hub.docker.com/_/postgres "If you would like to do additional initialization in an image derived from this one, add one or more *.sql, *.sql.gz, or *.sh scripts under /docker-entrypoint-initdb.d (creating the directory if necessary). After the entrypoint calls initdb to create the default postgres user and database, it will run any *.sql files, run any executable *.sh scripts, and source any non-executable *.sh scripts found in that directory to do further initialization before starting the service."

Changes to the database (new tables, new columns on exiisting tables etc..) should exist in `db/dbsetup.sql` as this file will setup the database from scratch.

# Database Migration
If changes are made to the database setup file `db/dbsetup.sql` then there should also be a migration script defined that when run will migrate an existing database to be compatible with the changes to `db/dbsetup.sql`.

Migration scripts should be placed in `db/migration/migrations`. For more information about how to write migration scripts see the tutorial here https://github.com/golang-migrate/migrate/blob/master/database/postgres/TUTORIAL.md

## Migrate Down
If you find yourself in the unlikely scenario of needing to undo migrations, by way of the `down.sql` scripts in the `db/migration/migrations` directory, you can change the line that reads `m.Steps(1)` in `db/migration/cmd/main.go` to `m.Steps(-1)` and run `make migrate_database` from the base directory.

## Migrate Validations
To migrate how the validations are stored in the db, you must run the validation migrations, which will remove the validation field from fhir_endpoints_info and fhir_endpoints_info_history tables, create the validations and validation_result tables, and add a field to reference the validation_result table in the fhir_endpoints_info and fhir_endpoints_info_history tables. To do this, follow the steps outlined below:

 * Follow the usual migration steps
 * Once the database has been migrated, start up Lantern with `make run`
 * Kill the endpoint manager: `docker stop lantern-back-end_endpoint_manager_1`
 * To migrate the validation information into the validation table: run `make migrate_validations direction=up`
    * You can run `make migrate_validations direction=down` to do a down migration (putting the validation information into the validation field)


# Database Schema

## fhir_endpoints table
The fhir_ endpoints table is the master list of FHIR service-base URLs that will be queried.
| Field        | Type           | Description  |
| ------------- |:-------------:| -----:|
| id     | INTEGER | Database ID of endpoint |
| url     | VARCHAR(500)      |   Service base URL of endpoint |
| list_source | VARCHAR(500)      |   Name/URL of list source that the endpoint came from |
| versions_response | JSONB      |    The response returned from querying the endpoint's versions operation |
| created_at | TIMESTAMPTZ      |    Timestamp of creation |
| updated_at | TIMESTAMPTZ      |    Timestamp of last update |

## fhir_endpoint_organizations table
The fhir_endpoint_organizations table contains all the organizations that each FHIR endpoint services.
| Field        | Type           | Description  |
| ------------- |:-------------:| -----:|
| id     | INTEGER | Database ID of organization |
| organization_name | VARCHAR(500)    | Name of the organization |
| organization_zipcode | VARCHAR(500) | Zip Code of the organization |
| organization_npi_id | VARCHAR(500)  | NPI ID of the organization |
| updated_at | TIMESTAMPTZ      |    Timestamp of last update |

## fhir_endpoint_organizations table
The fhir_endpoint_organizations table contains all the organizations that each FHIR endpoint services.
| Field        | Type           | Description  |
| ------------- |:-------------:| -----:|
| id     | INTEGER | fhir endpoint organization mapping database ID that is used to connect a single endpoint to one or many organizations |
| org_database_id  | INTEGER    | fhir endpoint organization ID referencing a fhir endpoint's organization in the fhir_endpoint_organizations table. |

## list_source_info table
The list_source_info table stores information about whether each list source came from CHPL or not.
| Field        | Type           | Description  |
| ------------- |:-------------:| -----:|
| list_source     | VARCHAR(500)  | Name/URL of list sources that Lantern collects  |
| is_chpl     | BOOLEAN     |  Whether the list source comes from chpl (true) or not (false). |
| updated_at  | TIMESTAMPTZ | Timestamp when this list source was last seen in CHPL response during population run. Used to identify stale sources. |


## fhir_endpoints_info table
The fhir_endpoints_info table contains the data collected from the last query of the FHIR endpoint at `url` while requesting the version given by `requested_version` and represents the most up to date information
| Field        | Type           | Description  |
| ------------- |:-------------:| -----:|
| id     | INTEGER | Database ID of endpoint |
| healthit_mapping_id     | INTEGER | healthit product mapping id that links an endpoint to the healthit_products_map table in order to map an endpoint to one or many healthit products |
| vendor_id     | INTEGER | Vendor ID of endpoint |
| url     | VARCHAR(500)      |   Service base URL of endpoint |
| tls_version     | VARCHAR(500)      |   Transport Layer Security (TLS) version of endpoint |
| mime_types | VARCHAR(500)[]      |    MIME types this endpoint supports |
| capability_statement     | JSONB      |   Capability statement receieved from endpoint |
| validation_result_id     | INTEGER      |   Validation id referencing the validation_results table |
| included_fields | JSONB      |    Structure that shows which capability statement fields and extensions are supported/unsupported by endpoint |
| operation_resource | JSONB     |    Stores the resources and their supported operations from the FHIR endpoint. |
| supported_profiles | JSONB | Stores the supported FHIR profiles from the FHIR endpoint. |
| created_at | TIMESTAMPTZ      |    Timestamp of creation |
| updated_at | TIMESTAMPTZ      |    Timestamp of last update |
| smart_response     | JSONB      |   SMART response receieved from endpoint|
| metadata_id  | INTEGER | Metadata ID referencing the fhir_endpoints_metadata table |
| requested_fhir_version  | VARCHAR(500)  | The FHIR version requested when querying the endpoint. Defaults to 'None' for endpoint entries where no specific FHIR version was requested. |
| capability_fhir_version  | VARCHAR(500)  | The FHIR version pulled out of the capability statement. |

## fhir_endpoints_info_history table
The fhir_endpoints_info_history table contains the history of the fhir_endpoints_info table. The operation field of the fhir_endpoints_info_history table represents if the entry was inserted for the first time (I) ie: The first query ever performed at the given `url` with the given `requested_version`, if the information retrieved from querying the `url` with the `requested_version` for an existing info entry was updated in any way (U) or if the info entry was removed (D). Deletion occurs in the case where a URL was once in a vendor list and was being queried by Lantern, but no longer exists in a vendor list and therefore will no longer exist in the `fhir_endpoints` table and will no longer be queried.
| Field        | Type           | Description  |
| ------------- |:-------------:| -----:|
| operation     | CHAR(1) | Entry operation (I for insert, U for Update, D for Delete)  |
| entered_at | TIMESTAMPTZ      |    Timestamp of creation |
| user_id     | VARCHAR(500) | database user ID |
| id     | INTEGER | database ID of endpoint |
| healthit_mapping_id     | INTEGER | healthit product mapping id that links an endpoint to the healthit_products_map table in order to map an endpoint to one or many healthit products |
| vendor_id     | INTEGER | vendor ID of endpoint |
| url     | VARCHAR(500)      |   Service base URL of endpoint |
| tls_version     | VARCHAR(500)      |   Transport Layer Security (TLS) version of endpoint |
| mime_types | VARCHAR(500)[]      |    MIME types this endpoint supports |
| capability_statement     | JSONB      |   Capability statement receieved from endpoint |
| validation_result_id     | INTEGER      |   Validation id referencing the validation_results table |
| included_fields | JSONB      |    Structure that shows which capability statement fields and extensions are supported/unsupported by endpoint |
| operation_resource | JSONB     |    Stores the resources and their supported operations from the FHIR endpoint. |
| supported_profiles | JSONB | Stores the supported FHIR profiles from the FHIR endpoint. |
| created_at | TIMESTAMPTZ      |    Timestamp of creation |
| updated_at | TIMESTAMPTZ      |    Timestamp of last update |
| smart_response     | JSONB      |   SMART response receieved from endpoint|
| metadata_id  | INTEGER  | Metadata ID referencing the fhir_endpoints_metadata table |
| requested_fhir_version  | VARCHAR(500)  | The FHIR version requested when querying the endpoint. Defaults to 'None' for endpoint entries where no specific FHIR version was requested. |
| capability_fhir_version  | VARCHAR(500)  | The FHIR version pulled out of the capability statement. |

## fhir_endpoints_metadata table
The fhir_endpoints_metadata table contains the metadata information collected from the last query of the FHIR endpoint at `url` and represents the most up to date information
| Field        | Type           | Description  |
| ------------- |:-------------:| -----:|
| id     | INTEGER | Database ID of endpoint |
| url     | VARCHAR(500)      |   Service base URL of endpoint |
| http_response     | INTEGER    |   HTTP response receieved from endpoint metadata url |
| availability     | DECIMAL(5,4)    |   All-time availability percentage. The number of total HTTP 200 responses that have ever been received from this endpoint divided by the total number of HTTP request attempt |
| errors     | VARCHAR(500)   |   Errors receieved from querying endpoint |
| response_time_seconds     | DECIMAL(7,4)    |   HTTP response time of endpoint |
| smart_http_response     | INTEGER    |  HTTP response receieved from endpoint SMART url |
| requested_fhir_version  | VARCHAR(500)  | The FHIR version requested when querying the endpoint. Defaults to 'None' for endpoint entries where no specific FHIR version was requested. |
| created_at | TIMESTAMPTZ      |    Timestamp of creation |
| updated_at | TIMESTAMPTZ      |    Timestamp of last update |

## validation_results table
| Field        | Type           | Description  |
| ------------- |:-------------:| -----:|
| id     | INTEGER | Database ID of the validation result ID entry |

## validations table
| Field        | Type           | Description  |
| ------------- |:-------------:| -----:|
| rule_name     | VARCHAR(500) | Name of the validation rule |
| valid     | BOOLEAN | Whether the endpoint passed (true) or failed (false) the validation check|
| expected     | VARCHAR(500) | The expected result of the validation check |
| actual     | VARCHAR(500) | The actual returned result of the validation check on an endpoint |
| comment     | VARCHAR(500) | Any comments that give more details about the validation |
| reference     | VARCHAR(500) | Reference URL for validation rule |
| implementation_guide     | VARCHAR(500) | Implementation guide that the validation rule is associated with if one exists |
| validation_result_id     | INTEGER | ID referencing the validation result table which groups validations for a single endpoint together |

## endpoint_organization table
The endpoint_organization table stores the matches made by the endpoint linker algorithm between endpoints and NPI organizations.
| Field        | Type           | Description  |
| ------------- |:-------------:| -----:|
| url     | VARCHAR(500) | Service base URL of endpoint |
| confidence     | VARCHAR(500) | Confidence percentage that match created by linking algorithm is accurate |
| created_at | TIMESTAMPTZ      |    Timestamp of creation |
| updated_at | TIMESTAMPTZ      |    Timestamp of last update |

## vendors table
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
| practice_type | VARCHAR(500) | The practice type of the product, either Ambulatory or Inpatient |
| created_at | TIMESTAMPTZ      |    Timestamp of creation |
| updated_at | TIMESTAMPTZ      |    Timestamp of last update |

## healthit_products_map table
The healthit_products_map table stores IDs that link FHIR endpoints to one or many health IT products
| Field        | Type           | Description  |
| ------------- |:-------------:| -----:|
| id     | INTEGER | healthit product mapping database ID that is used to connect a single endpoint to one or many healthit products |
| healthit_product_id     | INTEGER    | healthIT product ID referencing a healthIT product in the healthit_products table. |

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
| requested_fhir_version  | VARCHAR(500)  | The FHIR version requested when querying the endpoint. Defaults to 'None' for endpoint entries where no specific FHIR version was requested. |
| created_at | TIMESTAMPTZ | Timestamp of creation |
| updated_at | TIMESTAMPTZ | Timestamp of last update |