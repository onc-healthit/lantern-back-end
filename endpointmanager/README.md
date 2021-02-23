# FHIR Endpoint Manager

The FHIR Endpoint Manager is a service that coordinates the data capture and retrieval for FHIR endpoints.

## Configuration
The FHIR Endpoint Manager reads the following environment variables:

**These variables must be set on your system**

* **LANTERN_CHPLAPIKEY**: The key necessary for accessing CHPL

  Default value: \<none>

  You can obtain a CHPL API key [here](https://chpl.healthit.gov/#/resources/chpl-api).

* **LANTERN_DBUSER_READONLY**: The database user that the application will use to read from the database.

  Default value: none. This value is not used within the code. Suggested value: 'lantern_ro'

* **LANTERN_DBPASSWORD_READONLY**: The password for accessing the database as user LANTERN_DBUSER_READONLY.

  Default value: none. This value is not used within the code. Suggested value: 'postgrespassword_ro'

* **LANTERN_DBUSER_READWRITE**: The database user that the application will use to read or write to the database.

  Default value: none. This value is not used within the code. Suggested value: 'lantern_rw'

* **LANTERN_DBPASSWORD_READWRITE**: The password for accessing the database as user LANTERN_DBUSER_READWRITE.

  Default value: none. This value is not used within the code. Suggested value: 'postgrespassword_rw'

**These variables can use the default values *in development*. These should be set on the production system.**

* **LANTERN_DBHOST**: The hostname where the database is hosted.

  Default value: localhost

* **LANTERN_DBPORT**: The port where the database is hosted.

  Default value: 5432

* **LANTERN_DBUSER**: The database user that the application will use to read and write from the database.

  Default value: lantern

* **LANTERN_DBPASSWORD**: The password for accessing the database as user LANTERN_DBUSER.

  Default value: postgrespassword

* **LANTERN_DBNAME**: The name of the database being accessed.

  Default value: lantern

* **LANTERN_DBSSLMODE**: The level of SSL certificate verification that is performed. For a production system, this should be set to 'verify-full'.

  Default value: disable

* **LANTERN_ENDPTINFO_CAPQUERY_QNAME**: The name of the queue used by the endpointmanager and the capabilityquerier.

  Default value: endpoints-to-capability

* **LANTERN_EXPORT_NUMWORKERS**: The number of workers to use to parallelize creating the JSON export file and the JSON archive file.

  Default value: 25

* **LANTERN_EXPORT_DURATION**: The amount of time given to a worker (in seconds) to get a URL's history data from the database for the JSON export file and the JSON archive file.

  Default value: 240

### Test Configuration

When testing, the FHIR Endpoint Manager uses the following environment variables:

* **LANTERN_TEST_DBUSER** instead of LANTERN_DBUSER: The database user that the application will use to read and write from the database.

  Default value: lantern

* **LANTERN_TEST_DBPASSWORD** instead of LANTERN_DBPASSWORD: The password for accessing the database as user LANTERN_TEST_DBUSER.

  Default value: postgrespassword

* **LANTERN_TEST_DBNAME** instead of LANTERN_DBNAME: The name of the database being accessed.

  Default value: lantern_test

## Packages

The Endpoint Manager includes many packages with distinct purposes.

### Capability Handler

Takes messages off of the queue that include the capability statements of endpoints as well as additional data about the http interaction with the endpoint. Processes the endpoints (including linking them) and adds the data to the database.

### Capability Parser

Creates a model for capability statements and makes specific attributes of a capability statement queryable within the code. Can parse DSTU2, STU3, and R4 capability statements.

### CHPL Mapper

Maps endpoints to CHPL vendors and stores the mapping in the database. Eventually will map endpoints to CHPL products as well as additional information becomes available.

### CHPL Querier

Queries the CHPL service for CHPL product information and stores in the database.

### Config

Manages the configuration variables for all of the Endpoint Manager services.

### Endpoint Manager

Handles the object models and database storage for the endpoint information that Lantern is gathering.

### FHIR Endpoint Querier

Adds a list of endpoints to the database.

### NPPES Querier

Reads in a CSV file of NPPES data. You can find the latest monthly export of NPPES data here: http://download.cms.gov/nppes/NPI_Files.html

## Building and Running

The first time you run something, you may need to do the following in the directory where the main.go file is located:

```bash
go get ./... # You may have to set environment variable GO111MODULE=on
go mod download
```

### Base main.go

The Endpoint Manager main function is currently a stub function. You will see that the endpointmanager is running if you see "Started the endpoint manager." in as the output. 

Endpoint Manager functionality long term will rely on the lantern message queue (RabbitMQ) and the PostgreSQL database being available.

To run, perform the following commands:

```bash
cd endpointmanager/cmd
go run main.go
```

### Capability Receiver

Takes messages off of the queue that include the capability statements of endpoints as well as additional data about the http interaction with the endpoint. Processes the endpoints (including linking them) and adds the data to the database.

To run, perform the following commands:

```bash
cd endpointmanager/cmd/capabilityreceiver
go run main.go
```

The commands below assume that you are starting in the root directory of the Lantern backend project.

### CHPL Querier

Queries the CHPL service for CHPL product information and stores in the database.

Primarily uses the `chplquerier` package.

To run, perform the following commands:

```bash
cd endpointmanager/cmd/chplquerier
go run main.go
```

### Endpoint Populator

Parses a JSON file of endpoints and adds them to the database.

Primarily uses the `fhirendpointquerier` package.

To run, perform the following commands:

```bash
cd endpointmanager/cmd/endpointpopulator
go run main.go <path to endpoint json file>
```

### NPPES Populator

Reads in a CSV file of NPPES data. You can find the latest monthly export of NPPES data here: http://download.cms.gov/nppes/NPI_Files.html

Primarily uses the `nppesquerier` package.

To run, perform the following commands:


```bash
cd endpointmanager/cmd/nppesorgpopulator
go run main.go <path to nppes csv file>
```

### Expected Endpoint Source Formatting

The Endpoint Manager expects the format of an endpoint source list to be in one of the formats below:

Epic Endpoint Sources (JSON):

```
{
  "Entries": [
    {
      "OrganizationName": <name of the organization>,
      "FHIRPatientFacingURI": <location of the FHIR endpoint>
    },
    ...
  ]
}
```

Cerner Endpoint Sources (JSON):

```
{
  "endpoints": [
    {
      "name": <name of the organization>,
      "baseUrl": <location of the FHIR endpoint>,
      "type": <endpoint type>
    },
    ...
  ]
}
```

Lantern Endpoint Sources (JSON):

```
{
  "Endpoints": [
    {
      "URL": <location of the FHIR endpoint>,
      "OrganizationName": <name of the organization>,
      "NPIID": <organization npi id>
    },
    ...
  ]
}
```

NPPES Endpoint pfile (CSV):

```
"NPI","Endpoint Type","Endpoint Type Description","Endpoint","Affiliation","Endpoint Description","Affiliation Legal Business Name","Use Code","Use Description","Other Use Description","Content Type","Content Description","Other Content Description","Affiliation Address Line One","Affiliation Address Line Two","Affiliation Address City","Affiliation Address State","Affiliation Address Country","Affiliation Address Postal Code"
...
```

### Adding a New Endpoint List

To add a new endpoint list, add an entry to the EndpointResourcesList.json file located in the resources/prod_resources directory with the endpoint name, the name the endpoint source file will be saved as, and the endpoint URL. If the format does not match any of those listed above in the expected endpoint formats, add a new parser. See lantern-back-end/endpointmanager/pkg/fetcher/cernerlist.go, lantern-back-end/endpointmanager/pkg/fetcher/epiclist.go, or lantern-back-end/endpointmanager/pkg/fetcher/lanternlist.go for examples of the interface which endpoint list parsers need to adhere to.

## Endpoint Linker Algorithm Manual Corrections

To manually add a link between an endpoint and npi organization after the linker algorithm has been run, add the endpoint url and the npi id of the organization you want to link to the linkerMatchesWhitelist.json file. To manually remove a link between an endpoint and npi organization, add the linked endpoint url and the npi id of the organization you want to remove from the database to the linkerMatchesBlacklist.json file. Both files are found in the resources/prod_resources directory, and expect the following format:

```
[
  {
    "endpointURL": <url of endpoint>,
    "organizationID":<npi id of organization>
  },
  ...
]
```
