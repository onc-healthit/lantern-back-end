# FHIR Endpoint Manager

## Configuration
The FHIR Endpoint Manager reads the following environment variables:

* **LANTERN_ENDPTMGR_DBHOST**: The hostname where the database is hosted.

  Default value: localhost

* **LANTERN_ENDPTMGR_DBPORT**: The port where the database is hosted.

  Default value: 5432

* **LANTERN_ENDPTMGR_DBUSER**: The database user that the application will use to read and write from the database.

  Default value: postgres

* **LANTERN_ENDPTMGR_DBPASS**: The password for accessing the database as user LANTERN_ENDPTMGR_DBUSER.

  Default value: postgrespassword

* **LANTERN_ENDPTMGR_DBNAME**: The name of the database being accessed.

  Default value: postgres

* **LANTERN_ENDPTMGR_DBSSLMODE**: The level of SSL certificate verification that is performed. For a production system, this should be set to 'verify-full'.

  Default value: disable

* **LANTERN_ENDPTMGR_LOGFILE**: The location of the logfile for log messages

  Default value: endpointmanagerLog.json

## Building and Running

The Endpoint Manager has not yet been dockerized. To run, perform the following commands:

```bash
go get ./... # You may have to set environment variable GO111MODULE=on
go mod download
go run endpointmanager/main.go
```