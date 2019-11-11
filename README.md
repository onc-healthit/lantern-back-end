# Lantern

- [FHIR Endpoint Manager](#fhir-endpoint-manager)
  * [Configuration](#configuration)
  * [Building And Running](#building-and-running)
- [FHIR Endpoint Querier](#fhir-endpoint-querier)
  * [Configuration](#configuration-1)
  * [Building And Running](#building-and-running-1)
  * [Building And Running via Docker Container](#building-and-running-via-docker-container)
- [Additional Services](#additional-services)
  * [Starting All Services Using docker-compose](#starting-all-services-using-docker-compose)
  * [Starting Prometheus via Docker Container](#starting-prometheus-via-docker-container)
  * [Starting Prometheus via Local Clone](#starting-prometheus-via-local-clone)
  * [Prometheus With Remote Storage (PostgreSQL)](#prometheus-with-remote-storage--postgresql-)
      - [Adding the FHIR Querier service as a target](#adding-the-fhir-querier-service-as-a-target)
  * [Starting RabbitMQ](#starting-rabbitmq)
  * [Starting Grafana](#starting-grafana)
  * [Viewing Colllected Data In Grafana](#viewing-colllected-data-in-grafana)
- [Testing](#testing)
    + [Running All Unit Tests](#running-all-unit-tests)
    + [Running Tests With Coverage](#running-tests-with-coverage)
- [Contributing](#contributing)
  * [Lintr](#lintr)
  * [Govendor](#govendor)
- [License](#license)

# FHIR Endpoint Manager

## Configuration
The FHIR Endpoint Manager reads the following environment variables:

* **LANTERN_ENDPTMGR_DBHOST**: The hostname where the database is hosted.

  Default value: localhost

* **LANTERN_ENDPTMGR_DBPORT**: The port where the database is hosted.

  Default value: 5432

* **LANTERN_ENDPTMGR_DBUSER**: The database user that the application will use to read and write from the database.

  Default value: lantern

* **LANTERN_ENDPTMGR_DBPASS**: The password for accessing the database as user LANTERN_ENDPTMGR_DBUSER.

  Default value: postgrespassword

* **LANTERN_ENDPTMGR_DBNAME**: The name of the database being accessed.

  Default value: lantern

* **LANTERN_ENDPTMGR_DBSSLMODE**: The level of SSL certificate verification that is performed. For a production system, this should be set to 'verify-full'.

  Default value: disable

* **LANTERN_ENDPTMGR_LOGFILE**: The location of the logfile for log messages

  Default value: endpointmanagerLog.json

## Building and Running

```bash
go get ./... # You may have to set environment variable GO111MODULE=on
go mod download
go run endpointmanager/main.go
```

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
# Additional Services
The data collected by the endpoint querier can then be collected by Prometheus, which can be written to a Postgres database using the Prometheus Postgres storage adapter. This data can ultimately be viewed in Grafana. Below is information about how to start these additional services.

## Starting All Services Using docker-compose
All of the required services to run the Lantern back end are contained in the docker-compose file.

**Notice:** Before running `docker-compose up` make sure that you have created a hidden file named `.env` containing the environment variables specified in the `env.sample` file located alongside `docker-compose.yml`

To run all of the dockerized services for a development environment, run

```bash
docker-compose up
```

This will start endpoint querier, Prometheus, Postgres, Prometheus Postgres storage adapter, Grafana, and Rabbitmq; will setup the networking between the related services; and will publish ports.

To run all of the dockerized services for a production environment, run

```bash
docker-compose -f docker-compose.yml up
```

This will start endpoint querier, Prometheus, Postgres, Prometheus Postgres storage adapter, Grafana, and Rabbitmq; will setup the networking between the related services; and will only publish port 3000 to port 80 for Grafana.

To start all of the services in the background, add `-d` to your `docker-compose up` command.

To stop everything and remove the containers and network, run:
```bash
docker-compose down
```

To stop everything and keep the containers/volumes run:
```bash
docker-compose stop
```

If you stopped the containers and wish to restart them you can run:
```bash
docker-compose start
```

## Starting Prometheus via Docker Container
You'll still need a prometheus.yml configuration file for this, see https://github.com/prometheus/prometheus/blob/master/documentation/examples/prometheus.yml make sure that the configuration has [the FHIR Querier as a Target](#adding-the-fhir-querier-service-as-a-target)
```bash
docker run -p 9090:9090 -v <absolute path to config>/prometheus.yml:/etc/prometheus/prometheus.yml prom/prometheus
```
***IMPORTANT:*** If running docker on a Mac you'll have to configure the [FHIR Querier Target](#adding-the-fhir-querier-service-as-a-target) to be `host.docker.internal:<endpoint_port>` instead of localhost. For Linux systems, starting the Prometheus docker container with `--network=host` instead of `-p 9090:9090` will work and you can specify the target location to be `localhost:<endpoint_port>`.

## Starting Prometheus via Local Clone
Alternatively you can checkout Prometheus source code and run it locally
```bash
git clone https://github.com/prometheus/prometheus.git
cd prometheus
make build
```
Configure prometheus.yml to scrape from the fhir_target.
For our current purposes the example prometheus configuration at `prometheus/documentation/examples/prometheus.yml` will suffice
Finally, you can start an instance of prometheus by running:
```bash
./prometheus --config.file=prometheus.yml
```

## Prometheus With Remote Storage (PostgreSQL)
In order for Prometheus to use a remote PostgreSQL database for storage there are 3 components that need to be in place. A helpful tutorial that details each of the steps can be found [here](https://docs.timescale.com/latest/tutorials/prometheus-adapter)
If you haven't created docker volumes for Prometheus pr PostgreSQL, you'll have to run the following commands.
```bash
docker volume ls # To list already created volumes
docker volume create prometheusData # Volume to persist Prometheus data
docker volume create pgdata # Volume to persist data written to PostgreSQL database
```
1. [PostgreSQL Database with the pg_prometheus extension](https://github.com/timescale/pg_prometheus)

    ```bash
    docker run --name pg_prometheus -d -e POSTGRES_PASSWORD=<postgrespassword> -e POSTGRES_USER=lantern -e POSTGRES_DB=lantern -p 5432:5432 --volume pgdata:/var/lib/postgresql/data timescale/pg_prometheus:latest-pg11 postgres -csynchronous_commit=off
    ```
   
    Note that this will create a database called `lantern` with the admin user name `lantern`.

2. [PostgreSQL remote storage adapter to facilitate communication between Prometheus and the Database](https://github.com/timescale/prometheus-postgresql-adapter)

    It is important that the pg_prometheus container started in step 1 is up and running before starting the prometheus-postgresql-adapter, as the prometheus-postgresql-adapter will need to run database setup tasks the first time that it is run.

    ```bash
    docker run --name prometheus_postgresql_adapter --link pg_prometheus -d -p 9201:9201 timescale/prometheus-postgresql-adapter:latest -pg-host=pg_prometheus -pg-password=<postgrespassword> -pg-database=lantern -pg-user=lantern -pg-prometheus-log-samples
    ```
3. [Prometheus instance with remote storage adapter configuration](https://github.com/timescale/prometheus-postgresql-adapter)

    ```bash
    docker run -p 8080:9090 --link prometheus_postgresql_adapter -v <AbsoluePathToConfig>/prometheus.yml:/etc/prometheus/prometheus.yml --volume prometheusData:/prometheus prom/prometheus
    ```

#### Adding the FHIR Querier service as a target
Make sure the config file contains the following:
```yaml
scrape_configs:
  # The job name is added as a label `job=<job_name>` to any timeseries scraped from this config.

  - job_name: 'FHIRQUERY'

    # metrics_path defaults to '/metrics'
    # scheme defaults to 'http'.

    static_configs:
    - targets: ['localhost:<endpoint_port>']
```
In order to instruct Prometheus to write the collected data to a Postgres database, you'll have to configure the location of the Postgres remote storage adapter.
**Important:** If you are on a MAC you may have to use host.docker.internal instead of localhost here.
```yml
remote_write:
    - url: "http://localhost:9201/write"
remote_read:
    - url: "http://localhost:9201/read"
```

## Starting RabbitMQ

Modified from https://hub.docker.com/_/rabbitmq:

```bash
docker run -d --hostname lantern-mq --name lantern-mq -p 15672:15672 -p 5672:5672 rabbitmq:3-management
```

This will start a RabbitMQ container listening on the default port of 5672. If you give that a minute, then do `docker logs lantern-mq`, you'll see in the output a block similar to:

```
 node           : rabbit@lantern-mq
 home dir       : /var/lib/rabbitmq
 config file(s) : /etc/rabbitmq/rabbitmq.conf
 cookie hash    : 2VgNGhlcNws2enUk77Sv9w==
 log(s)         : <stdout>
 database dir   : /var/lib/rabbitmq/mnesia/rabbit@lantern-mq
```

You can also check that you have access to the admin page by navigating to `http://localhost:15672` and using username and password `guest:guest`.

## Starting Grafana
Make sure that you have Docker installed and running on your machine
```bash
docker run -d -p 3000:3000 grafana/grafana
```

## Viewing Colllected Data In Grafana
1. Navigate to `http://localhost:3000/` in a web browser
2. Login using Username: admin, Password admin
3. Add a datasource
  - If using Prometheus without remote storage, add a Prometheus datasource, running on `http://localhost:9090` by default. Select access Browser and then Save
  - If using PostgreSQL remote storage, add a PostgreSQL data source.
    - If you are running the postgres database on a local docker container and are publishing port 5432, location is `localhost:5432` or `host.docker.internal:5432` (if on a MAC).
    - If you started the postgres database using the docker-compose file in this repository (#starting-all-services-using-docker-compose) then the postgres database will be located at `pg_prometheus:5432`
    - Enter `lantern` in the Database and User fields and enterthe PostgreSQL password you started the PostgreSQL docker container with in the Password field. Finally select `disable` for SSL Mode.
4. From the main page create a Dashboard, adding visualizations for the metrics you would like to explore

# Testing
### Running All Unit Tests
To run all tests in a project run:
```bash
go test ./...
```
This will search the current directory and all sub directories for files matching the pattern `*_test.go`. You also have the option to specify a package location or file if you do not want to run all tests at once.
### Running Tests With Coverage
If you are interested in having coverage information displayed along with the pass/fail status' of the tests you can run the tests with the `--covermode=count` option.
```bash
go test -covermode=count ./...
ok      command-line-arguments  0.009s  coverage: 31.8% of statements
```
If you are interested in more in-depth coverage analysis you'll have to generate a coverage report, in the following command the coverage report is named coverage.out
```bash
go test -coverprofile=coverage.out ./...
```
Using the generated coverage file you can use `go tool cover` to view the coverage report in your browser and see which lines are or are not being covered.
```bash
go tool cover -html=coverage.out
```

# Contributing
## Lintr
Code going through PR should pass the lintr invoked by running:
```bash
golangci-lint run -E gofmt
```
You may have to install golangci-lint first. To do this on a Mac you can run:
```bash
brew install golangci/tap/golangci-lint
```
More information about golangci-lint can be found [here](https://github.com/golangci/golangci-lint)

## Govendor
Dependencies required for each package are cached in a `vendor/` directory within each package. Go will search for dependencies within the `vendor/` directory at build-time.
To cache dependencies for a package using the govendor tool:
```bash
go get -u github.com/kardianos/govendor # Download govendor
cd <your package>
govendor init # You may need to add your go/bin directory to your PATH if govendor is not found. This will create a vendor directory
govendor add +external # Copy external package dependencies into vendor directory, dependencies will appear the same as they do in src/
govendor add +local # Copy package dependencies that share the same project root into the vendor directory
```
If you add dependencies to your package, or there are updates to dependencies (either local or external) you will have to run the following commands in order to make sure that the vendor directory reflects the updates.
```bash
govendor update +external # Update external dependencies
govendor update +local # Update dependencies that share the same project root
```

# License

Copyright 2019 The MITRE Corporation

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

```
http://www.apache.org/licenses/LICENSE-2.0
```

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
