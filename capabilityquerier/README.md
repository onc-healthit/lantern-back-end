# Capability Querier

The capability querier is a service that queries endpoints for their capability statements, and sends those capability statements along with any additional relevant information to a queue.

## Configuration
The capability querier reads the following environment variables:

**These variables can use the default values *in development*. These should be set on the production system.**

* **LANTERN_QHOST**: The hostname where the queue is hosted.

  Default value: localhost

* **LANTERN_QPORT**: The port where the queue is hosted.

  Default value: 5672

* **LANTERN_QUSER**: The user that the application will use to read and write from the queue.

  Default value: capabilityquerier

* **LANTERN_QPASSWORD**: The password for accessing the database as user LANTERN_QUSER.

  Default value: capabilityquerier

* **LANTERN_CAPQUERY_QNAME**: The name of the queue being accessed.

  Default value: capability-statements

* **LANTERN_CAPQUERY_NUMWORKERS**: The number of workers to use to parallelize processing of the capability statements.

  Default value: 10

* **LANTERN_ENDPTINFO_CAPQUERY_QNAME**: The name of the queue used by the endpointmanager and the capabilityquerier.

  Default value: endpoints-to-capability

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

* **LANTERN_EXPORTFILE_WAIT**: The length of time between getting the final endpoints off of the queue and creating the JSON export file of data. This is in seconds.

  Default value: 300 (5 minutes)

* **LANTERN_PRUNING_THRESHOLD**: The length of time in which fhir endpoint info history entries stored in the database which are older than this threshold will be pruned. This is in minutes.

  Default value: 43800 (1 month)

### Test Configuration

When testing, the capability querier uses the following environment variables:

* **LANTERN_TEST_QUSER** instead of LANTERN_QUSER: The user that the application will use to read and write from the queue.

  Default value: capabilityquerier

* **LANTERN_TEST_QPASSWORD** instead of LANTERN_QPASSWORD: The password for accessing the database as user LANTERN_QUSER.

  Default value: capabilityquerier

* **LANTERN_TEST_QNAME** instead of LANTERN_CAPQUERY_QNAME: The name of the queue being accessed.

  Default value: test-queue

* **LANTERN_TEST_ENDPTINFO_CAPQUERY_QNAME** instead of LANTERN_ENDPTINFO_CAPQUERY_QNAME: The name of the queue used by the endpointmanager and the capabilityquerier.

  Default value: test-endpoints-to-capability

* **LANTERN_TEST_DBUSER** instead of LANTERN_DBUSER: The database user that the application will use to read and write from the database.

  Default value: lantern

* **LANTERN_TEST_DBPASSWORD** instead of LANTERN_DBPASSWORD: The password for accessing the database as user LANTERN_TEST_DBUSER.

  Default value: postgrespassword

* **LANTERN_TEST_DBNAME** instead of LANTERN_DBNAME: The name of the database being accessed.

  Default value: lantern_test

## Building and Running

The capability querier currently connects to the lantern message queue (RabbbitMQ). All log messages are written to stdout.

### Using Docker-Compose

The Endpoint Querier has been added to the application docker-compose file. See the [top-level README](../README.md) for how to run docker-compose.

### Using the Individual Docker Container

At this time, it's not recommended to start this as an individual container because of the dependence on the endpointlist file which is in another go project. This is challenging to manage for starting a single instance and not worth pursuing given that starting this container with all the other containers or running alone should be sufficient.

### Running alone

The instructions below assume that you are in `capabilityquerier/`.

The capability querier has not yet been dockerized. To run, perform the following commands:

```bash
go get ./... # You may have to set environment variable GO111MODULE=on
go mod download
go run cmd/main.go
```

## Scaling

To scale out the capability querier service edit the docker-compose.yml and docker-compose.override.yml file to include additional capability querier services. 
```
capability_querier_2:
  ...
``` 

Each instance of the capability querier will query endpoints from the same endpoints-to-capability queue in a round robin style.