
# Network Statistics Querier
The Network Statistics Querier is a service that retrieves the response code and response time for a FHIR API endpoint.

## Configuration
The Network Statistics Querier reads the following environment variables:

**These variables must be set on your system**

\<none>

**These variables can use the default values *in development*. These should be set on the production system.**

* **LANTERN_ENDPTQRY_PORT**: The port where the metrics gathered from the FHIR endpoints will be hosted.

  Default value: 3333

* **LANTERN_ENDPTQRY_QUERY_INTERVAL**: Number of minutes to wait between queries (Note: actual time between queries will be greater since the time it takes to run the queries is non-zero)

  Default value: 10

## Building And Running

After the Network Statistics Querier starts, all output is directed to the configured log file. To check that the endpoint querier is running as expected, navigate to `http://localhost:<configured port>/metrics` to see the metrics being collected by the querier.

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

The Network Statistics Querier takes two arguments, a JSON file containing the endpoints which the service should query, and the source of that list. The list of endpoints provided in `<project_root>/endpoints/resources/` are:
* CernerEndpointSources.json from Cerner's endpoint list. The expected source for querying this list is `Cerner` (e.g. `go run *.go resources/CernerEndpointSources.json Cerner`).
* EpicEndpointSources.json from Epic's endpoint list. The expected source for querying this list is `Epic`.
* For any other endpoint source JSON files, the expected source can be any string you want saved in the database as the list source.

```bash
go get ./... # You may have to set environment variable GO111MODULE=on
go mod download
go run *.go resources/<endpoint_list>.json <source>
```

### Expected Endpoint Source Formatting

The Network Statistics Querier expects the format of an endpoint source list to be in the below format, unless one of the exceptions noted below.

```
{
  "Entries": [
    {
      "OrganizationName": <name of the organization>,
      "FHIRPatientFacingURI": <location of the FHIR endpoint>
    },
    ...
  ]
}
```

Exceptions:
* Cerner