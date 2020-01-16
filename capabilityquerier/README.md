# Capability Querier

The capability querier is a service that queries endpoints for their capability statements, and sends those capability statements along with any additional relevant information to a queue.

## Configuration
The capability querier reads the following environment variables:

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

* **LANTERN_CAPQUERY_QRYINTVL**: The length of time between performing batch queries of endpoints for their capability statements. This is in minutes.

  Default value: 1440 (24 hours)


### Test Configuration

When testing, the capability querier uses the following environment variables:

* **LANTERN_TEST_QUSER** instead of LANTERN_QUSER: The user that the application will use to read and write from the queue.

  Default value: capabilityquerier

* **LANTERN_TEST_QPASSWORD** instead of LANTERN_QPASSWORD: The password for accessing the database as user LANTERN_QUSER.

  Default value: capabilityquerier

* **LANTERN_TEST_CAPQUERY_QNAME** instead of LANTERN_CAPQUERY_QNAME: The name of the queue being accessed.

  Default value: capability-statements-test

## Building and Running

The capability querier currently connects to the lantern message queue (RabbbitMQ). All log messages are written to stdout.

The instructions below assume that you are in `capabilityquerier/`.

The capability querier has not yet been dockerized. To run, perform the following commands:

```bash
go get ./... # You may have to set environment variable GO111MODULE=on
go mod download
go run cmd/main.go
```