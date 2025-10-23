### Capability Receiver

Takes messages off of the queue that include either the Capability Statement of an endpoint or the response from a $versions operation, as well as additional data about the http interaction with the endpoint. Runs validations, pulls out all defined resources in the Capability Statement, as well as all fields and extensions in the Capability Statement with data. Matches the endpoint to CHPL vendor and product information in the database. Saves the data in the database.

## Configuration
The Capability Receiver reads the following environment variables:

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

### Test Configuration

When testing, the Capability Receiver uses the following environment variables:

* **LANTERN_TEST_DBUSER** instead of LANTERN_DBUSER: The database user that the application will use to read and write from the database.

  Default value: lantern

* **LANTERN_TEST_DBPASSWORD** instead of LANTERN_DBPASSWORD: The password for accessing the database as user LANTERN_TEST_DBUSER.

  Default value: postgrespassword

* **LANTERN_TEST_DBNAME** instead of LANTERN_DBNAME: The name of the database being accessed.

  Default value: lantern_test

## Packages

The Capability Receiver includes many packages with distinct purposes.

### Capability Handler

Takes messages off of the queue that include either the Capability Statement of an endpoint or the response from a $versions operation, as well as additional data about the http interaction with the endpoint. Runs validations, pulls out all defined resources in the Capability Statement, as well as all fields and extensions in the Capability Statement with data. Saves the data in the database.

### CHPL Mapper

Maps endpoints to CHPL vendors and stores the mapping in the database. Eventually will map endpoints to CHPL products as well as additional information becomes available.

## Building and Running

The Capability Receiver currently connects to the lantern message queue (RabbbitMQ). All log messages are written to stdout.

### Using Docker-Compose

The Capability Receiver has been added to the application docker-compose file. See the [top-level README](../README.md) for how to run docker-compose.

### Using the Individual Docker Container

At this time, it's not recommended to start this as an individual container because of the dependence on the message queue and the database.

### Running alone

To run, perform the following commands:

```bash
cd cmd/capabilityreceiver
go run main.go
```

## Tracking New FHIR Capability Statement Fields

To start tracking a new FHIR capability statement field, the field must be added in accordance with the functionality in the capabilityreceiver/pkg/capabilityhandler/includedfields.go file, which is responsible for tracking if certain FHIR capability statement fields exist. To begin, add a list entry of fields representing the path to the new field to the fieldsList at the beginning of the RunIncludedFieldsChecks function in the capabilityreceiver/pkg/capabilityhandler/includedfields.go file. The path should be a list of all the capability statement fields that must be accessed to reach where the new field is stored in the capability statement, with the last element in the list being the name of the newly added field. If any of the included fields in the path to the new field are arrays of interfaces rather than a single interface, check to make sure the field name is included in the arrayFields list at the top of the capabilityreceiver/pkg/capabilityhandler/includedfields.go file, and if it is not, add the name of the field to that list. A field will be recorded as a supported field with 'Exists' in the includedFields structure set to true if there is at least one instance of that field being used in any of the possible locations specified for it. 

For example, if a new version of FHIR is published that has foo as a field, with foo being nested within a field called bar, you would add an entry to the fieldsList structure that looked like this:

```
{"bar", "foo"}
```

## Tracking New FHIR Capability Statement Extensions

To start tracking a new FHIR capability statement extension, the extension must be added in accordance with the functionality in the capabilityreceiver/pkg/capabilityhandler/includedfields.go file, which is responsible for tracking FHIR extensions. To begin, add a list entry of fields representing the path to the new FHIR extension to the extensionList at the beginning of the RunIncludedExtensionsChecks function in the capabilityreceiver/pkg/capabilityhandler/includedfields.go file. The path should be a list of all the capability statement fields that must be accessed to reach where the new extension is stored in the capability statement, with the second to last element in the list being the extension url and the last element in the list being the name of the FHIR extension. If an extension can be stored at multiple different locations in the capability statement, add an entry for each path to the extensionList with each having the same extension url and name. If any of the included fields in the path to the extension are arrays of interfaces rather than a single interface, check to make sure the field name is included in the arrayFields list at the top of the capabilityreceiver/pkg/capabilityhandler/includedfields.go file, and if it is not, add the name of the field to that list. An extension will be recorded as a supported extension with 'Exists' in the includedFields structure set to true if there is at least one instance of that extension being used in any of the possible locations specified for it.

For example, if a new version of FHIR is published that has capabilitystatement-foo as an extension with url http://example.org/fhir/capabilitystatement-foo, with this extension being nested within an extension field array which is nested inside field called bar, you would add an entry to the extensionList structure that looked like this:

```
{"bar", "extension", "http://example.org/fhir/capabilitystatement-foo", "capabilitystatement-foo"}
```

If the bar field was an array of interfaces, you would add "bar" to the end of the arrayFields list.

## Adding New Manual CHPL Product Matches
Start by viewing which FHIR endpoints do not yet have a mapped HealthIT Product and also have a populated software field in their capability statement by executing the following query against the Lantern database.
`SELECT DISTINCT healthit_product_id, capability_statement->'software'->>'name', capability_statement->'software'->>'version' FROM fhir_endpoints_info WHERE capability_statement->>'software' IS NOT NULL;`

Next, search through the HealthIT Products for a product that has a name similar to one of the names which was advertised by the software name field in the capability statemnt, returned by the query above.
Given that software names as advertised by capability statements won't always align exactly with what is in CHPL (the healthit_products table) you may have to try different variations of the advertised name before a product is found using the query below.
`SELECT * FROM healthit_products WHERE name ILIKE '%foobar%'`
If one of the resulting matches has a matching name, and a version that represents the version advertised by the capability statement, we will associate the CHPLID of the matching HealthIT Product to the name and version as advertised by the capability statement. For example, if the capability statement advertised software with name 'foobar' version '2.0.1' and there is a HealthIT Product with name 'foobar' and version '2.0', the CHPLID for the given HealthIT product can be associated with the software name and version advertised by the capability statement. This represented in the `lantern-back-end/resources/prod_resources/CHPLProductMapping.json` file as an entry that adheres to the following format.
```
{
     "name": "product name as it appears in the capability statement",
     "version": "product version at it appears in the capability statement",
     "CHPLID": "Given CHPL ID of the product"
}