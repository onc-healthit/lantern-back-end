### Capability Receiver

Takes messages off of the queue that include the capability statements of endpoints as well as additional data about the http interaction with the endpoint. Processes the endpoints (including linking them) and adds the data to the database.

To run, perform the following commands:

```bash
cd cmd/capabilityreceiver
go run main.go
```


## Configuration
The FHIR Endpoint Manager reads the following environment variables:

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

### Test Configuration

When testing, the FHIR Endpoint Manager uses the following environment variables:

* **LANTERN_TEST_DBUSER** instead of LANTERN_DBUSER: The database user that the application will use to read and write from the database.

  Default value: lantern

* **LANTERN_TEST_DBPASSWORD** instead of LANTERN_DBPASSWORD: The password for accessing the database as user LANTERN_TEST_DBUSER.

  Default value: postgrespassword

* **LANTERN_TEST_DBNAME** instead of LANTERN_DBNAME: The name of the database being accessed.

  Default value: lantern_test
