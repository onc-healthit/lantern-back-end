# Lantern
* Lantern Services
  * [Endpoint Manager](endpointmanager/README.md)
  * [Endpoint Querier](endpoints/README.md)
  * [Lantern Message Queue](lanternmq/README.md)
* [Additional Services](#additional-services)
  * [Using docker-compose](#using-docker-compose)
    * [Starting the Services](#starting-the-services)
    * [Stopping the Services](#stopping-the-services)
    * [Starting Services Behind SSL-Inspecting Proxy](#starting-services-behind-ssl-inspecting-proxy)
  * [Running the Services Individually](#running-the-services-individually)
    * [Endpoint Manager](#endpoint-manager)
    * [Endpoint Querier](#endpoint-querier)
    * [Prometheus with Remote Storage (PostgreSQL)](#prometheus-with-remote-storage-postgresql)
      * [Prometheus Configuration File](#prometheus-configuration-file)
    * [RabbitMQ](#rabbitmq)
    * [Grafana](#grafana)
      * [Viewing Collected Data in Grafana](#viewing-collected-data-in-grafana)
* [Testing](#testing)
  * [Running All Unit Tests](#running-all-unit-tests)
  * [Running Tests With Coverage](#running-tests-with-coverage)
  * [Running End to End Tests](#running-end-to-end-tests)
* [Contributing](#contributing)
  * [Lintr](#lintr)
  * [Govendor](#govendor)
* [License](#license)

# Additional Services

The Lantern infrastructure relies on several additional services. These include:
* Prometheus
* Prometheus remote storage adapter for PostgreSQL
* PostgreSQL database
* Grafana

Prometheus is used to capture time-series-based data from the FHIR API endpoints. It stores these in the PostgreSQL database using Timescale's PostgreSQL extension, [`pg_prometheus`](https://github.com/timescale/pg_prometheus) as well as the the [Prometheus remote storage adapter for PostgreSQL](https://github.com/timescale/prometheus-postgresql-adapter).

The PostgreSQL database is used to store all information related to the FHIR API endpoints. This includes the timeseries data captured by Prometheus as well as information from the capability statement, information gathered from Inferno, information about the EHR vendors from [CHPL](https://chpl.healthit.gov/#/search), and information about the provider organization using the endpoint.

Grafana creates many of the visualizations used by Lantern through querying the PostgreSQL database.

## Using docker-compose

All of the required services to run the Lantern back-end are contained in the docker-compose file.

### Starting the Services

**Notice:** Before running `docker-compose up` make sure that you have created a hidden file named `.env` containing the environment variables specified in the `env.sample` file located alongside `docker-compose.yml`

**If you have no containers** in your environment from a previous run of docker-compose, you will need to run `docker-compose up`.

For a *development* environment, run:

```bash
docker-compose up
```

This will create the containers, start up all the services, as well as publish ports.

For a *production* environment, run:

```bash
docker-compose -f docker-compose.yml up
```

This will create the containers, start all of the services, and will only expose Grafana on port 80.

To start the services in the background, add `-d` to your `docker-compose up` command.

**If you already have containers** in your environment from a previous run of docker-compose, you should run `docker-compose start`.

For a *development* environment, run:

```bash
docker-compose start
```

This will start up all the services as well as publish ports.

For a *production* environment, run:

```bash
docker-compose -f docker-compose.yml start
```

This will start all of the services and will only expose Grafana on port 80.


### Stopping the Services

To stop the services and retain the containers and network, run:

```bash
docker-compose stop
```

To stop the services and remove the containers and networks, run:

```bash
docker-compose down
```

To stop the services, remove the containers and networks, images, and volumes, run:

```bash
docker-compose down --rmi all -v
```

### Starting Services Behind SSL-Inspecting Proxy
If you are operating behind a proxy that does SSL-Inspection you will have to copy the certificates that are used by the proxy into the docker containers that will be communicating through the proxy. Currently the endpoint_querier is the only contianer that has such a requirement, the volumes entry `- ./certs/:/etc/ssl/certs` in the endpoint_querier service section  of `docker-compose.override.yml` will mount a directory named `certs` located in the root of this project into the location of the docker container where the container's OS will look for certificates. If you are operating behind an SSL-Inspecting proxy **you will have to copy your certificates into this directory.** The changes in the `docker-compose.override.yml` file will be applied if you run `docker-compose up`.

## Running The Services Individually

### Endpoint Manager

See the [Endpoint Manager documentation](endpointmanager/README.md).

### Endpoint Querier

See the [Endpoint Querier documentation](endpoints/README.md)

### Prometheus with Remote Storage (PostgreSQL)

In order for Prometheus to use a remote PostgreSQL database for storage there are 3 components that need to be in place. A helpful tutorial that details each of the steps can be found [here](https://docs.timescale.com/latest/tutorials/prometheus-adapter).

If you haven't created docker volumes for Prometheus pr PostgreSQL, you'll have to run the following commands.

```bash
docker volume ls # To list already created volumes
docker volume create prometheusData # Volume to persist Prometheus data
docker volume create pgdata # Volume to persist data written to PostgreSQL database
```

The three components are:

* [PostgreSQL Database with the pg_prometheus extension](https://github.com/timescale/pg_prometheus)
* [Prometheus remote storage adapter for PostgreSQL](https://github.com/timescale/prometheus-postgresql-adapter)
* [Prometheus](https://github.com/prometheus/prometheus)

These need to be started in the following order with the given commands:

1. PostgreSQL Database

    ```bash
    docker run --name pg_prometheus -e POSTGRES_PASSWORD=<postgrespassword> -e POSTGRES_USER=lantern -e POSTGRES_DB=lantern -p 5432:5432 --volume pgdata:/var/lib/postgresql/data timescale/pg_prometheus:latest-pg11 postgres -csynchronous_commit=off
    ```

    Note that this will create a database called `lantern` with the admin user name `lantern`.

2. Prometheus remote storage adapter for PostgreSQL

    It is important that the `pg_prometheus` container started in step 1 is up and running before starting the `prometheus_postgresql_adapter` container, as the `prometheus_postgresql_adapter` container will need to run database setup tasks the first time that it is run.

    ```bash
    docker run --name prometheus_postgresql_adapter --link pg_prometheus -p 9201:9201 timescale/prometheus-postgresql-adapter:latest -pg-host=pg_prometheus -pg-password=<postgrespassword> -pg-database=lantern -pg-user=lantern -pg-prometheus-log-samples
    ```

3. Prometheus

    Both the `pg_prometheus` and `prometheus_postgresql_adapter` containers need to be running before running this container.

    ```bash
    docker run -p 9090:9090 --link prometheus_postgresql_adapter -v <AbsoluePathToConfig>/prometheus.yml:/etc/prometheus/prometheus.yml --volume prometheusData:/prometheus prom/prometheus
    ```

#### [Prometheus Configuration File](prometheus.yml)

**In order to instruct Prometheus to write the collected data to a Postgres database**, you'll have to configure the location of the Postgres remote storage adapter. For example:

```yml
remote_write:
    - url: "http://prometheus_postgres_adapter:9201/write"
remote_read:
    - url: "http://prometheus_postgres_adapter:9201/read"
```

**To add the Endpoint Querier as a Prometheus target**, the information below should be contained in the configuration file.

```yaml
scrape_configs:
  # The job name is added as a label `job=<job_name>` to any timeseries scraped from this config.

  - job_name: 'FHIRQUERY'

    # metrics_path defaults to '/metrics'
    # scheme defaults to 'http'.

    static_configs:
    - targets: ['<endpoint_querier_container_name>:<endpoint_port>']
```

For example:

```yaml
scrape_configs:
  # The job name is added as a label `job=<job_name>` to any timeseries scraped from this config.

  - job_name: 'FHIRQUERY'

    # metrics_path defaults to '/metrics'
    # scheme defaults to 'http'.

    static_configs:
    - targets: ['endpoint_querier_1:3333']
```

#### Initializing the Database

The database initialization script is planned to be included as part of the docker-compose file. Until then, initialize the database using the following commands:

* If you have postgres installed locally (which you can do with `brew install postgresql`), you can do:

  ```
  psql -h <container name> -p <port> -U <username> -d <database> -a -f <setup script>
  ```

  For example:

  ```
  psql -h pg_prometheus -p 5432 -U postgres -d postgres -a -f endpointmanager/dbsetup.sql
  ```

* If you don't have postgres installed locally, you can open the database in docker and then copy past the commands from the dbsetup.sql file in. Open the database in docker:

  ```
  docker exec -it <container name> psql -U <username>
  ```

  Connect to your database using `\c <database name>`

  Copy/paste the contents of dbsetup.sql into the command prompt.

### RabbitMQ

The instructions below are modified from https://hub.docker.com/_/rabbitmq.

To start the RabbitMQ docker container, run

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

### Grafana
To start Grafana, run the following:

```bash
docker run -d -p 3000:3000 grafana/grafana
```

You can check that Grafana is up by nagivating to `http://localhost:3000` and using username and password `admin:admin`.

#### Viewing Collected Data in Grafana

1. Navigate to `http://localhost:3000/` in a web browser
2. Login using Username: admin, Password admin
3. Add PostgreSQL as data source
    - If you are running the postgres database on a local docker container and are publishing port 5432, location is `localhost:5432` or `host.docker.internal:5432` (if on a MAC).
    - If you started the postgres database using the docker-compose file in this repository (#starting-all-services-using-docker-compose) then the postgres database will be located at `pg_prometheus:5432`
    - Enter `lantern` in the Database and User fields and enterthe PostgreSQL password you started the PostgreSQL docker container with in the Password field. Toggle "TimescaleDB" to on. Finally select `disable` for SSL Mode.
4. From the main page create a Dashboard, adding visualizations for the metrics you would like to explore.

# Testing
## Running All Unit Tests

To run all tests in a project run:

```bash
go test -count=1 ./...
```

This will search the current directory and all sub directories for files matching the pattern `*_test.go`. Adding `-count=1` ensures that your test results will not be cached. You also have the option to specify a package location or file if you do not want to run all tests at once.

## Running Tests With Coverage

If you are interested in having coverage information displayed along with the pass/fail status' of the tests you can run the tests with the `--covermode=count` option.

```bash
go test -covermode=count ./...
```

If you are interested in more in-depth coverage analysis you'll have to generate a coverage report, in the following command the coverage report is named coverage.out

```bash
go test -coverprofile=coverage.out ./...
```

Using the generated coverage file you can use `go tool cover` to view the coverage report in your browser and see which lines are or are not being covered.

```bash
go tool cover -html=coverage.out
```

### Running End to End Tests
Running this command will build the project containers and then run e2e/integration_tests/*.go
```bash
docker-compose -f docker-compose.yml -f docker-compose.override.yml -f docker-compose.test.yml up --build --abort-on-container-exit
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
