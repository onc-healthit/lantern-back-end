
# FHIR Endpoint Querier
A service to send http requests to get capability statements from FHIR endpoints

## Configuration
The FHIR Endpoint Querier reads the following environment variables:

* **LANTERN_ENDPTQRY_PORT**: The port where the metrics gathered from the FHIR endpoints will be hosted.

  Default value: 3333

* **LANTERN_ENDPTQRY_LOGFILE**: The location of the logfile for log messages

  Default value: endpointQuerierLog.json

## Building And Running

The Endpoint Querier takes one arguement, a JSON file containing the endpoints which the service should query. The list of endpoints provided in `<project_root>/endpoints/resources/EndpointSources.json` was taken from https://fhirendpoints.github.io/data.json.

```bash
go get ./... # You may have to set environment variable GO111MODULE=on
go mod download
go run endpoints/*.go ./endpoints/resources/EndpointSources.json
```

## Building And Running via Docker Container
To build Docker container run the following command.
```bash
cd endpoints
docker build -t endpoint_querier .
```
To start the Docker container that you just bult run:
```bash
docker run -p 3333:3333 -it endpoint_querier
```