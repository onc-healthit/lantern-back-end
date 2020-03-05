# FHIR Endpoint Manager

The FHIR Endpoint Manager is a service that coordinates the data capture and retrieval for FHIR endpoints.

## Configuration
The FHIR Endpoint Manager reads the following environment variables:

**These variables must be set on your system**

* **LANTERN_CHPLAPIKEY**: The key necessary for accessing CHPL

  Default value: \<none>

  You can obtain a CHPL API key [here](https://chpl.healthit.gov/#/resources/chpl-api).

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

### Test Configuration

When testing, the FHIR Endpoint Manager uses the following environment variables:

* **LANTERN_TEST_DBUSER** instead of LANTERN_DBUSER: The database user that the application will use to read and write from the database.

  Default value: lantern

* **LANTERN_TEST_DBPASSWORD** instead of LANTERN_DBPASSWORD: The password for accessing the database as user LANTERN_TEST_DBUSER.

  Default value: postgrespassword

* **LANTERN_TEST_DBNAME** instead of LANTERN_DBNAME: The name of the database being accessed.

  Default value: lantern_test

## Building and Running

The Endpoint Manager main function is currently a stub function. You will see that the endpointmanager is running if you see "Started the endpoint manager." in as the output. 

Endpoint Manager functionality long term will rely on the lantern message queue (RabbitMQ) and the PostgreSQL database being available.

### Using Docker-Compose

The Endpoint Querier has been added to the application docker-compose file. See the [top-level README](../README.md) for how to run docker-compose.

### Using the Individual Docker Container

The instructions below assume that you are in `endpointmanager/`.

To build Docker container run the following command.

```bash
docker build -t endpointmanager .
```

To start the Docker container that you just built run:

```bash
docker run -it endpointmanager
```

### Running alone

The instructions below assume that you are in `endpointmanager/`.

The Endpoint Manager has not yet been dockerized. To run, perform the following commands:

```bash
go get ./... # You may have to set environment variable GO111MODULE=on
go mod download
go run cmd/main.go
```

