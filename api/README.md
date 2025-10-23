# API
The API service provides a REST API which can be used to download FHIR endpoints and organizations data. 

## Configuration
The API reads the following environment variables:

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

## Building and Running

The API connects to the postgres database. All log messages are written to stdout.

### Using Docker-Compose

The API has been added to the application docker-compose file. See the [top-level README](../README.md) for how to run docker-compose.

### Using the Individual Docker Container

At this time, it's not recommended to start this as an individual container because of the dependence on the postgres database. Restarting the API container would restart the postgres database as well and many other containers depend on the postgres container.