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

* **LANTERN_CAPQUERY_QRYINTVL**: The length of time between performing batch queries of endpoints for their capability statements. This is in minutes.

  Default value: 1440 (24 hours)

* **LANTERN_BROADCAST_EXCHANGE**: The name of the fanout exchange which broadcast the START/STOP message to subscribed queues.

  Default value: broadcast_exchange

* **LANTERN_BROADCAST_QUEUE**: The name of the queue that is ssubscribed to the exchange publishing the start/stop message. For multiple instances of capabilityQuerier, each instance must have a unique name.

  Default value: broadcast_queue

### Test Configuration

When testing, the capability querier uses the following environment variables:

* **LANTERN_TEST_QUSER** instead of LANTERN_QUSER: The user that the application will use to read and write from the queue.

  Default value: capabilityquerier

* **LANTERN_TEST_QPASSWORD** instead of LANTERN_QPASSWORD: The password for accessing the database as user LANTERN_QUSER.

  Default value: capabilityquerier

* **LANTERN_TEST_QNAME** instead of LANTERN_CAPQUERY_QNAME: The name of the queue being accessed.

  Default value: test-queue

* **LANTERN_TEST_ENDPTINFO_CAPQUERY_QNAME** instead of LANTERN_ENDPTINFO_CAPQUERY_QNAME: The name of the queue being accessed.

  Default value: test-endpoints-to-capability

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

To scale out the capability querier service edit the docker-compose.yml file 
to include another capabilityQuerier service. Under the environment define another name for the LANTERN_BROADCAST_QUEUE variable
```
capability_querier_2:
    environment:
        - LANTERN_BROADCAST_QUEUE=broadcast_queue_2
``` 

The value for LANTERN_BROADCAST_QUEUE can either be defined in directly in the docker-compose of separately in your .env file. Each capabilityQuerier must have a unique value for the LANTERN_BROADCAST_QUEUE variable.