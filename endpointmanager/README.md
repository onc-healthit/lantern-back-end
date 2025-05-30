# FHIR Endpoint Manager

The FHIR Endpoint Manager is a service that coordinates the data capture and retrieval for FHIR endpoints.

## Configuration
The FHIR Endpoint Manager reads the following environment variables:

**These variables must be set on your system**

* **LANTERN_CHPLAPIKEY**: The key necessary for accessing CHPL

  Default value: \<none>

  You can obtain a CHPL API key [here](https://chpl.healthit.gov/#/resources/api).

* **LANTERN_1UP_CLIENT_SECRET**: The client secret necessary for accessing the 1Up API
  
  Default value: none. 
  
  You can obtain a 1UP client secret by following the registration instructions [here](https://1up.health/docs/start/quick-start-guide/register).

* **LANTERN_1UP_CLIENT_ID**: The client id necessary for accessing the 1Up API
  
  Default value: none. 
  
  You can obtain a 1UP client ID by following the registration instructions [here](https://1up.health/docs/start/quick-start-guide/register).

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

* **LANTERN_QHOST**: The hostname where the queue is hosted.

  Default value: localhost

* **LANTERN_QPORT**: The port where the queue is hosted.

  Default value: 5672

* **LANTERN_QUSER**: The user that the application will use to read and write from the queue.

  Default value: capabilityquerier

* **LANTERN_QPASSWORD**: The password for accessing the database as user LANTERN_QUSER.

  Default value: capabilityquerier

* **LANTERN_QUERY_NUMWORKERS**: The number of workers to use to parallelize processing of the capability statements and version responses.

  Default value: 10

* **LANTERN_CAPQUERY_QRYINTVL**: The length of time between performing batch queries of endpoints for their capability statements. This is in minutes.

  Default value: 1380 (23 hours)

* **LANTERN_EXPORT_NUMWORKERS**: The number of workers to use to parallelize creating the JSON export file and the JSON archive file.

  Default value: 25

* **LANTERN_EXPORT_DURATION**: The amount of time given to a worker (in seconds) to get a URL's history data from the database for the JSON export file and the JSON archive file.

  Default value: 240

* **LANTERN_PRUNING_THRESHOLD**: The length of time (in minutes) determining how old a fhir_endpoints_info_history entry has to be in order to be considered for pruning. Only entries equal to or older than this threshold will undergo pruning.

  Default value: 43800 (~ 30 days)
  
### Test Configuration

When testing, the FHIR Endpoint Manager uses the following environment variables:

* **LANTERN_TEST_DBUSER** instead of LANTERN_DBUSER: The database user that the application will use to read and write from the database.

  Default value: lantern

* **LANTERN_TEST_DBPASSWORD** instead of LANTERN_DBPASSWORD: The password for accessing the database as user LANTERN_TEST_DBUSER.

  Default value: postgrespassword

* **LANTERN_TEST_DBNAME** instead of LANTERN_DBNAME: The name of the database being accessed.

  Default value: lantern_test

* **LANTERN_TEST_QUSER** instead of LANTERN_QUSER: The user that the application will use to read and write from the queue.

  Default value: capabilityquerier

* **LANTERN_TEST_QPASSWORD** instead of LANTERN_QPASSWORD: The password for accessing the database as user LANTERN_QUSER.

  Default value: capabilityquerier

## Packages

The Endpoint Manager includes many packages with distinct purposes.

### Archive File

Creates an archive of the data from the fhir_endpoints, fhir_endpoints_info and vendors tables in a JSON format.

### Capability Parser

Creates a model for capability statements and makes specific attributes of a capability statement queryable within the code. Can parse DSTU2, STU3, and R4 capability statements.

### CHPL Querier

Queries the CHPL service for CHPL product information and stores in the database.

### Config

Manages the configuration variables for all of the Endpoint Manager services.

### Endpoint Linker

Links endpoints to organizations, either by the NPI ID (preferred), or by the organization name.

### Endpoint Manager

Handles the object models and database storage for the endpoint information that Lantern is gathering.

### Fetcher
Contains parsers for the different endpoint list formats that Lantern takes in. Selects the correct endpoint list parser and adds the endpoint list information to the database.  

### FHIR Endpoint Querier

Adds a list of endpoints to the database.

### Helpers

Contains helpful functions that are used commonly throughout the project, such as a string array contains function and a fail on error function.

### History Pruning

Prunes the fhir_endpoints_info_history table to remove consecutive duplicate endpoint entries older than the 2x the LANTERN_PRUNING_THRESHOLD environment variable and deletes any associated validation table entries.


### NPPES Querier

Reads in a CSV file of NPPES data. You can find the latest monthly export of NPPES data here: http://download.cms.gov/nppes/NPI_Files.html

### Send Endpoints

Gets current list of endpoints and sends each one to the capabilityquerier queue. It continues to repeat this action every time the query interval period has passed.

### Smart Parser

Creates a model for smart responses so they can be parsed and analyzed. 

### Test Helper

Contains helpful functions used in testing throughout the project.

### Versions Operator Parser

Creates a model for the versions operator response and makes the version and default version information easily accessible within the code. 

### Workers

Contains the code needed for creating, starting, and stopping workers used to parallelize processing.

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

The commands below assume that you are starting in the root directory of the Lantern backend project.

### Archive File
Creates an archive of the data from the fhir_endpoints, fhir_endpoints_info and vendors tables between the given dates in a JSON format and saves it to the given 'file' name.

Primarily uses the `archivefile` package.

```bash
cd endpointmanager/cmd/archivefile
go run main.go <start date> <end date> <file name>
```

### CHPL Populator 

Queries the CHPL service for the CHPL list of endpoint lists and stores in a JSON file.

To run, perform the following commands:

```bash
cd endpointmanager/cmd/CHPLpopulator
go run main.go <CHPL Endpoint List URL> <JSON file to save CHPL Endpoint List>
```

### CHPL Querier

Queries the CHPL service for CHPL product information and stores in the database.

Primarily uses the `chplquerier` package.

To run, perform the following commands:

```bash
cd endpointmanager/cmd/chplquerier
go run main.go
```

### CHPL Update Check
Queries the CHPL service for the CHPL list of endpoint lists and checks to see if the list has been updated. If it has, the CHPL endpoint list JSON file in the `resources/prod_resources` is updated with these changes, and file is created with all the updated CHPL endpoint list URLs listed to be used by the automated cron job CHPL update check to send an email with these URLs.

To run, perform the following commands:

```bash
cd endpointmanager/cmd/CHPLupdatecheck
go run main.go <CHPL Endpoint List URL> <JSON file to save CHPL Endpoint List>
```

### Data Validation
Checks if the number of endpoints in the fhir_endpoints table is greater than what could be queried in the query interval and displays a warning if it is.

To run, perform the following commands:

```bash
cd endpointmanager/cmd/datavalidation 
go run main.go
```

### Endpoint Exporter
Copies the entire contents of endpoint_export view into a csv which will be written to /tmp.

To run, perform the following commands:

```bash
cd endpointmanager/cmd/endpointexporter 
go run main.go
```

### Endpoint Linker

Links endpoints to organizations, either by the NPI ID (preferred), or by the organization name.

Primarily uses the `endpointlinker` package.

To run, perform the following commands:

```bash
cd endpointmanager/cmd/endpointlinker 
go run main.go
```

Or to run with printed linker results and information:
```bash
cd endpointmanager/cmd/endpointlinker
go run main.go --verbose
```

### Endpoint Populator

Parses a JSON file of endpoints and adds them to the database.

Primarily uses the `fhirendpointquerier` package.

To run, perform the following commands:

```bash
cd endpointmanager/cmd/endpointpopulator
go run main.go <path to endpoint json file>
```

### Endpoint Webscraper

Queries an endpoint list URL whose endpoint list is contained within an HTML table and uses web sraping to pull the endpoints out of the table and save them into a JSON file in the Lantern endpoint list format.

To run, perform the following commands:

```bash
cd endpointmanager/cmd/endpointwebscraper
go run main.go <Endpoint list name> <Endpoint list URL> <JSON file name to save endpoint list to>
```

### History Pruning
Prunes the fhir_endpoints_info_history table to remove consecutive duplicate endpoint entries older than the pruning threshold environment variable.

Primarily uses the `historypruning` package.

```bash
cd endpointmanager/cmd/historypruning 
go run main.go
```

### Send Endpoints
Gets current list of endpoints sends each one to the capabilityquerier queue. It continues to repeat this action every time the query interval period has passed.

Primarily uses the `sendendpoints` package.

To run, perform the following commands:

```bash
cd endpointmanager/cmd/sendendpoints 
go run main.go
```

### Expected Endpoint Source Formatting

The Endpoint Manager expects the format of an endpoint source list to be in one of the formats below:

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

FHIR Bundle of FHIR Endpoint Resources (JSON):

```
{
   "resourceType": "Bundle",
   "entry": [
     {
       "resource": {
         "name": <name of the organization>,
         "managingOrganiation": { "display": <text for resource>, "reference": <organization name> },
         "address": <location of the FHIR endpoint>
       }
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

To manually add a link between an endpoint and npi organization after the linker algorithm has been run, add the endpoint url and the npi id of the organization you want to link to the linkerMatchesAllowlist.json file. To manually remove a link between an endpoint and npi organization, add the linked endpoint url and the npi id of the organization you want to remove from the database to the linkerMatchesBlocklist.json file. Both files are found in the resources/prod_resources directory, and expect the following format:

```
[
  {
    "endpointURL": <url of endpoint>,
    "organizationID":<npi id of organization string>
  },
  ...
]
```

**Note:** The npi organizationID should be a string


## Endpoint Info History Pruning

The history pruning algorithm is run every 23 hours using a cron job (more details on how to set up in main README). The pruning algorithm will iterate over all of the fhir_endpoint_info_history entries for each distinct FHIR endpoint URL that have entered_at dates that are older than the time determined by subtracting the LANTERN_PRUNING_THRESHOLD from the current time, and also have entered_at dates that are newer than the current time minus the LANTERN_PRUNING_THRESHOLD times 2. Having a lower limit of the LANTERN_PRUNING_THRESHOLD times 2 ensures that the algorithm does not repeat all of the pruning checks on the same entries after every 23 hours, but that it also does not miss any entries that have not yet been pruned. The LANTERN_PRUNING_THRESHOLD, which set to one month by default, ensures that there is always data newer than the LANTERN_PRUNING_THRESHOLD that is not pruned, since an entry has to be older than the threshold in order to be considered for pruning.

The pruning algorithm will remove any consecutive duplicate entries in the fhir_endpoint_info_history table. A fhir_endpoint_info_history entry is considered a duplicate if there is an older consecutive entry that that has the same stored information for the endpoint's TLS version, MIME types, and SMART response, and if the newer entry's stored capability statement only differs by fields included in a list of ignored fields, such as the CapabilityStatement.date field. If a fhir_endpoint_info_history entry is found to be a duplicate of an older consecutive entry, it is deleted from the table, and this continues until only the oldest of the consecutive duplicated entries remains. This pruning strategy is advantageous in that there will always be a duration of at least LANTERN_PRUNING_THRESHOLD minutes worth of queries in the history table for each endpoint, therefore Lantern can inspect LANTERN_PRUNING_THRESHOLD minutes worth of data to see how every endpoint responded within each query interval while still saving storage space by removing duplicate data or data which only differs in the values reported for fields in the ignored fields set. Keeping all entries containing any unique data allows Lantern to keep track of how each endpoint has changed over long periods of time.
