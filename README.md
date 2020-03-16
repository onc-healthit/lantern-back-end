# Lantern
* [Running Lantern - Basic Flow](#running-lantern---basic-flow)
* [Testing Lantern - Basic Flow](#testing-lantern---basic-flow)
* [Make Commands](#make-commands)
* [Running Lantern Services Individually](#running-lantern-services-individually)
* [Using Docker Compose](#using-docker-compose)
* [Testing - Details](#testing---details)
* [Contributing](#contributing)
* [License](#license)

# Running Lantern - Basic Flow

**Note:** Before running the commands below, make sure that you either:
* have created a hidden file named `.env` in the root directory of the project containing the environment variables specified in the `env.sample` file.
* have the variables specified in the `env.sample` file defined as system environment variables.

## Setup your environment

To run Lantern, several environment variables need to be set. These are defined within each project's README. Each README defines the variables that *must* be set on the system vs those whose default values are sufficient.

  * [Endpoint Manager](endpointmanager/README.md)
  * [Network Statistics Querier](networkstatsquerier/README.md)
  * [Capability Querier](capabilityquerier/README.md)
  * [Lantern Message Queue](lanternmq/README.md)

## Clean your environment

**This is optional!**

If you'd like to start with a clean slate, run:

```bash
make clean
```

This removes all docker images, networks, and local volumes.

## Start Lantern

1. In your terminal, run:

    ```bash
    make run
    ```

    This starts all of the services except for the endpointmanager:
    * **Lantern Front End** - The front end for the Lantern application (localhost:8090)
    * **Grafana** - the data visualization service (localhost:80)
    * **PostgreSQL** - application database
    * **LanternMQ (RabbitMQ)** - the message queue (localhost:15672)
    * **Prometheus / Prometheus remote storage adapter for PostgreSQL** - continuously queries the endpoints to determine response status and response time
    * **Capability Querier** - queries the endpoints for their capability statements once a day. Kicks off the initial query immediately.

2. **If you have a clean database** 
    1. Run the following command to begin populating the database:

        ```bash
        make populatedb
        ```

        This runs:
        * the **endpoint populator**, which iterates over the list of endpoint sources and adds them to the database.
        * the **CHPL querier**, which requests health IT product information from CHPL and adds these to the database
        * the **NPPES populator**, which adds provider data from the monthly NPPES export to the database. 
          * this is an optional item to add to the database, and can take around and hour to load.

    1. Open a new tab and run the following:

        ```bash
        cd endpointmanager/cmd/capabilityreceiver
        go run main.go
        cd ../../..
        ```

        This receives messages off of the queue. This also does some endpoint linking and processing. This action runs forever.

1. **If you want to requery and rereceive capability statements**, open two new tabs and run the following:

    In the first tab (this runs forever), run:

    ```bash
    cd capabilityquerier/cmd
    go run main.go
    cd ../..
    ```

    If you do not already have the capability receiver running, in the second tab (this runs forever), run:

    ```bash
    cd endpointmanager/cmd/capabilityreceiver
    go run main.go
    cd ../../..
    ```

## Stop Lantern

Run

```bash
make stop
```

## Starting Services Behind SSL-Inspecting Proxy

If you are operating behind a proxy that does SSL-Inspection you will have to copy the certificates that are used by the proxy into a `certs` directory at the root directory of the project. Docker-Compose will copy these certs into the containers that need to use the certificates.

# Testing Lantern - Basic Flow

There are three types of tests for Lantern and three corresponding commands:

| test type | command |
| --- | --- |
| unit | `make test` |
| integration | `make test_int` |
| end-to-end |  `make test_e2e` |
| all tests | `make test_all` |

# Make Commands

| make command | action |
| --- | --- |
|`make run` | runs docker-compose for a development environment |
|`make run_prod` | runs docker-compose for a production environment |
| `make stop` | runs docker-compose `down` for a development environment |
| `make stop_prod` | runs docker-compose `down` for a development environment |
| `make clean` | runs docker-compose `down` with the `--rmi local -v` tags to remove local images and volumes. Runs this for all docker-compose setups. Before running, it confirms with the user that they actually want to clean.
| `make clean_remote` | runs docker-compose `down` with the `--rmi all -v` tags to remove all images and volumes. Runs this for all docker-compose setups. Before running, it confirms with the user that they actually want to clean. |
| `make test` | runs unit tests | 
| `make test_int` | runs integration tests |
|  `make test_e2e` | runs end-to-end tests |
|`make test_all` | runs all tests and ends if any of the tests fail| 


# Running Lantern Services Individually

## Internal Services

See each internal service's README to see how to run that service as a standalone service.

  * [Endpoint Manager](endpointmanager/README.md)
  * [Network Statistics Querier](networkstatsquerier/README.md)
  * [Capability Querier](capabilityquerier/README.md)
  * [Lantern Message Queue](lanternmq/README.md)

## External Services

* Prometheus, the Prometheus Remote Storage Adapter for PostgreSQL, and PostgreSQL
* RabbitMQ
* Grafana


### Prometheus, the Prometheus Remote Storage Adapter for PostgreSQL, and PostgreSQL

Prometheus is used to capture time-series-based data from the FHIR API endpoints. It stores these in the PostgreSQL database using Timescale's PostgreSQL extension, [`pg_prometheus`](https://github.com/timescale/pg_prometheus) as well as the the [Prometheus remote storage adapter for PostgreSQL](https://github.com/timescale/prometheus-postgresql-adapter).

The PostgreSQL database is used to store all information related to the FHIR API endpoints. This includes the timeseries data captured by Prometheus as well as information from the capability statement, information gathered from Inferno, information about the EHR vendors from [CHPL](https://chpl.healthit.gov/#/search), and information about the provider organization using the endpoint.

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
    - url: "http://prometheus_postgresql_adapter:9201/write"
remote_read:
    - url: "http://prometheus_postgresql_adapter:9201/read"
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
    - targets: ['endpoint_querier:3333']
```

#### Initializing the Database by hand

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

  Copy/paste the contents of dbsetup.sql or other sql commands into the command prompt.

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

Grafana creates many of the visualizations used by Lantern through querying the PostgreSQL database.

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

# Using Docker Compose

All relevant docker-compose instructions are included in the Makefile under the appropriate `make` commands.

## No Containers

**If you have no containers** in your environment from a previous run of docker-compose, you will need to run `docker-compose up`.

### Development Environment

For a *development* environment, run:

```bash
docker-compose up
```

This will create the containers, start up all the services, as well as publish ports.

### Production Environment

For a *production* environment, run:

```bash
docker-compose -f docker-compose.yml up
```

This will create the containers, start all of the services, and will only expose Grafana on port 80.

To start the services in the background, add `-d` to your `docker-compose up` command.

## Existing Containers

**If you already have containers** in your environment from a previous run of docker-compose, you should run `docker-compose start`.

### Development Environment

For a *development* environment, run:

```bash
docker-compose start
```

This will start up all the services as well as publish ports.

### Production Environment 

For a *production* environment, run:

```bash
docker-compose -f docker-compose.yml start
```

This will start all of the services and will only expose Grafana on port 80.


## Stopping the Services

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

# Testing - Details

The test instructions in the Makefile include several additional flags to ensure that tests are run atomically and to check any resource usage conflicts due to parallelization. These are not listed below to reduce duplication. See the Makefile for the details.

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

## Running Integration Tests

To run integration tests (which take a long time to run), add the `integration` tag:

```bash
go test -tags=integration ./...
```

## Running End to End Tests
Running this command will build the project containers and then run e2e/integration_tests/*.go
```bash
docker-compose -f docker-compose.yml -f docker-compose.override.yml -f docker-compose.test.yml up --build --abort-on-container-exit
```
To bring down all containers and remove all volumes used in the end-to-end tests you can run:
```bash
docker-compose -f docker-compose.yml -f docker-compose.override.yml -f docker-compose.test.yml down -v
```
A successful test will show the following output before the docker containers are stopped.

```
lantern-e2e                      | ok  	github.com/onc-healthit/lantern-back-end/e2e/integration_tests	30.031s
lantern-e2e exited with code 0
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


