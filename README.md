# FHIR Target Querier
A service to send http requests to get capability statements from FHIR endpoints

## Building And Running

The Endpoint Querier takes one arguement, a JSON file containing the endpoints which the service should query. The list of endpoints provided in `<project_root>/resources/EndpointSources.json` was taken from https://fhirendpoints.github.io/data.json.

```bash
go install ./...
go run src/endpoints/*.go ./resources/EndpointSources.json
```

## Starting Prometheus via Docker Container
You'll still need a prometheus.yml configuration file for this, see https://github.com/prometheus/prometheus/blob/master/documentation/examples/prometheus.yml make sure that the configuration has [the FHIR Querier as a Target](#adding-the-fhir-querier-service-as-a-target)
```bash
docker run -p 9090:9090 -v <absolute path to config>/prometheus.yml:/etc/prometheus/prometheus.yml prom/prometheus
```
***IMPORTANT:*** If running docker on a Mac you'll have to configure the target location to be `host.docker.internal:8443` instead of localhost. For Linux systems, starting the Prometheus docker container with `--network=host` instead of `-p 9090:9090` will work and you can specify the target location to be `localhost:8443`.

## Starting Prometheus via Local Clone
Alternatively you can checkout Prometheus source code and run it locally
```bash
git clone https://github.com/prometheus/prometheus.git
cd prometheus
make build
```
Configure prometheus.yml to scrape from the fhir_target.
For our current purposes the example prometheus configuration at `prometheus/documentation/examples/prometheus.yml` will suffice
#### Adding the FHIR Querier service as a target
Make sure the config file contains the following:
```yaml
scrape_configs:
  # The job name is added as a label `job=<job_name>` to any timeseries scraped from this config.

  - job_name: 'FHIRQUERY'

    # metrics_path defaults to '/metrics'
    # scheme defaults to 'http'.

    static_configs:
    - targets: ['localhost:8080']
```
Finally, you can start an instance of prometheus by running:
```bash
./prometheus --config.file=prometheus.yml
```

## Starting Grafana
Make sure that you have Docker installed and running on your machine
```bash
docker run -d -p 3000:3000 grafana/grafana
```

## Viewing Colllected Data In Grafana
1. Navigate to `http://localhost:3000/` in a web browser
2. Login using Username: admin, Password admin
3. Add a Prometheus datasource, running on `http://localhost:9090` by default. Select access Browser and then Save
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
