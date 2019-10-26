# FHIR Target Querier
A service to send http requests to get capability statements from FHIR endpoints

## Building And Running

The Endpoint Querier takes one arguement, a JSON file containing the endpoints which the service should query. The list of endpoints provided in `<project_root>/endpoints/resources/EndpointSources.json` was taken from https://fhirendpoints.github.io/data.json.

```bash
go get ./...
go install ./...
go run endpoints/*.go ./endpoints/resources/EndpointSources.json
```

## Building And Running via Docker Container
To build Docker container run the following command. NOTE: If you are behind a corperate proxy, the dependencies might not be able to be pulled down.
```bash
cd endpoints
docker build -t endpoint_querier .
```
To start the Docker container that you just bult run:
```bash
docker run -p 8443:8443 -it endpoint_querier
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
docker run --name pg_prometheus -d -e POSTGRES_PASSWORD=<postgrespassword> -p 5432:5432 --volume pgdata:/var/lib/postgresql/data timescale/pg_prometheus:latest postgres -csynchronous_commit=off
```
2. [PostgreSQL remote storage adapter to facilitate communication between Prometheus and the Database](https://github.com/timescale/prometheus-postgresql-adapter)
It is important that the pg_prometheus container started in step 1 is up and running before starting the prometheus-postgresql-adapter, as the prometheus-postgresql-adapter will need to run database setup tasks the first time that it is run.
```bash
docker run --name prometheus_postgresql_adapter --link pg_prometheus -d -p 9201:9201 timescale/prometheus-postgresql-adapter:latest -pg-host=pg_prometheus -pg-password=<postgrespassword> -pg-prometheus-log-samples
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
  - If using PostgreSQL remote storage, add a PostgreSQL data source, running on `localhost:5432` or `host.docker.internal:5432` (if on a MAC). Enter `postgres` in the Database and User fields and enterthe PostgreSQL password you started the PostgreSQL docker container with in the Password field. Finally select `disable` for SSL Mode.
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
