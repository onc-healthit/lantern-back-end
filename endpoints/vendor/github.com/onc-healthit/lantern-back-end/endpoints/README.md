
# FHIR Endpoint Querier
The FHIR Endpoint Querier is a service that retrieves the response code and response time for a FHIR API endpoint.

## Configuration
The FHIR Endpoint Querier reads the following environment variables:

* **LANTERN_ENDPTQRY_PORT**: The port where the metrics gathered from the FHIR endpoints will be hosted.

  Default value: 3333

* **LANTERN_ENDPTQRY_LOGFILE**: The location of the logfile for log messages

  Default value: endpointQuerierLog.json

* **LANTERN_ENDPTQRY_QUERY_INTERVAL**: Number of minutes to wait between queries (Note: actual time between queries will be greater since the time it takes to run the queries is non-zero)

  Default value: 10m

## Building And Running

After the Endpoint Querier starts, all output is directed to the configured log file. To check that the endpoint querier is running as expected, navigate to `http://localhost:<configured port>/metrics` to see the metrics being collected by the querier.

The instructions below assume that you are in the `endpoints/` directory.

### Using Docker-Compose

The Endpoint Querier has been added to the application docker-compose file. See the [top-level README](../README.md) for how to run docker-compose.

### Using the Individual Docker Container

To build Docker container run the following command.

```bash
docker build -t endpoint_querier .
```

To start the Docker container that you just built run:

```bash
docker run -p 3333:3333 -it endpoint_querier --name <container name>
```

### Running without Docker

The Endpoint Querier takes one arguement, a JSON file containing the endpoints which the service should query. The list of endpoints provided in `<project_root>/endpoints/resources/EndpointSources.json` was taken from https://fhirendpoints.github.io/data.json.

```bash
go get ./... # You may have to set environment variable GO111MODULE=on
go mod download
go run *.go resources/EndpointSources.json
```