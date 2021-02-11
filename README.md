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
  * [Capability Querier](capabilityquerier/README.md)
  * [Capability Receiver](capabilityreceiver/README.md)
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

    This starts all of the following services:
    * **PostgreSQL** - application database
    * **LanternMQ (RabbitMQ)** - the message queue (localhost:15672)
    * **Capability Querier** - queries the endpoints for their capability statements once a day. Kicks off the initial query immediately.
    * **Capability Receiver** - receives the capability statements from the queue, peforms validations and saves the results to fhir_endpoints_info
    * **Endpoint Manager** - sends endpoints to the capability querying queues


2. **If you have a clean database or want to update the data in your database** 
    1. Run the following command to update your endpoint resource files found in `lantern-back-end/resources/prod_resources`. This command will automatically query all the endpoint sources listed in EndpointResourceList.json, which can be found in `lantern-back-end/resources/prod_resources`, and NPPES for their endpoint and npi data files. 

     ```bash
      make update_source_data
      ```

    2. Run the following command to begin populating the database usinig the data found in `lantern-back-end/resources/<dev_resources|prod_resources>`
      -Note: If you are doing development use the `dev_resources` directory as it contains less endpoints which reduces unnecessary load on the servers hosting the endpoints we are querying.

    The populated db scrpt expects the resources directory to contain the following files:
      * **CernerEndpointSources.json** - JSON file containing endpoint information from Cerner
      * **EpicEndpointSources.json** - JSON file containing endpoint information from Epic
      * **LanternEndpointSources.json** - JSON file containing endpoint information reported directly to Lantern
      * **endpoint_pfile.csv** - enpoint_pfile from the data dissemination package downloaded from https://download.cms.gov/nppes/NPI_Files.html
      * **npidata_pfile.csv** - npidata_pfile from the data dissemination package downloaded from https://download.cms.gov/nppes/NPI_Files.html 
        * NOTE: This file can take a very long time to load so for development purposes, the load time can be reduced by only using the first 100000 entries. The first 100000 entries can be obtained by running `head -n 100000 npidata_pfile_20050523-20191110.csv >> npidata_pfile.csv`
      * **linkerMatchesWhitelist and linkerMatchesBlacklist** - whitelist and blacklist files used in manually correcting the endpoint to npi organization linker. To manually add/remove endpoint to npi organization links in the database, see endpointmanager README on format for adding links to whitelist and blacklist files

      ```bash
      make populatedb
      ```

      This runs the following tasks inside the endpoint manager container:
      * the **endpoint populator**, which iterates over the list of endpoint sources and adds them to the database.
      * the **CHPL querier**, which requests health IT product information from CHPL and adds these to the database
      * the **NPPES endpoint populator**, which adds endpoint data from the monthly NPPES export to the database. 
      * the **NPPES org populator**, which adds provider data from the monthly NPPES export to the database. 
        * this is item will take an hour to load if you use the full npidata_pfile

3. **If you want to requery and rereceive capability statements outside the refresh interval** run the following:

    ```bash
    docker restart lantern-back-end_endpoint_manager_1
    ```

## Stop Lantern

Run

```bash
make stop
```

## Starting Services Behind SSL-Inspecting Proxy

If you are operating behind a proxy that does SSL-Inspection you will have to copy the certificates that are used by the proxy into the following directories:
  * `capabalitiyquerier/certs/`
  * `capabilitiyreceiver/certs/`
  * `endpointmanager/certs/`
  * `lanternmq/certs`
  * `shinydashboard/certs/`
  * `e2e/certs`

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
|`make backup_database` | saves a database backup .sql file in the lantern base directory with name lantern_backup_`<timestamp>`.sql|
|`make restore_database file=<backup file name>` | restores the backup database that the 'file' parameter is set to|
|`make migrate_database` | Starts the postgres service and runs the next `*.up.sql` migration in the `db/migration/migrations` directory that has not yet been run. Must run this command the number times equal to the number of migrations you want to run. |
|`make update_source_data` |Automatically queries the Epic and Cerner endpoint source websites and the NPPES npi and endpoint data and stores these resource files in the resources/prod_resources directory |
|  `make lint` | Runs the R and golang linters |
|  `make lint_go` | Runs the golang lintr |
|  `make lint_R` | Runs the R lintr |
| `make json_export file=<export file name>` | Exports the history of the endpoint data to a JSON file specified by the 'file' parameter |
| `make history_pruning` | Prunes the fhir_endpoint_info_history table to remove duplicate entries |
| `make create_archive start=<start date> end=<end date> file=<archive file name>` | Creates an archive of the data in the database between the given dates in a JSON format and saves it to the given 'file' name. The dates format is '2021-01-31' (year, month, date). Example: `make create_archive start=2020-06-01 end=2021-06-01 file=archive_file.json` |

# Configure Data Collection Failure System

You can configure a data collection failure system using cron and the data_collection_check.sh script located in the scripts directory to send an email notification if the lantern data collection goes down for any reason. 

The data_collection_check.sh script runs outside of Lantern and periodically checks to see if data has been written to the fhir_endpoints_info within the last N many minutes, where N is the Lantern query interval. If data has not been written within said interval, or the database is down, then the script sends an alert to the set email address.

To set up the script for this data collection failure system, you must insert the correct information into the following variables located at the beginning of the data_collection_check script. The DB_NAME, DB_USER, and QUERY_INTERVAL variables used in the script should match their corresponding environmental variable (shown in parentheses below) defined in the .env file:
  * Set the EMAIL variable to the email you want the failure system to send alerts to
  * Set the DB_NAME variable to name of your database (LANTERN_DBNAME)
  * Set the DB_USER variable to the name of the database user (LANTERN_DBUSER)
  * Set the QUERY_INTERVAL variable to the capability querier query interval in minutes (LANTERN_CAPQUERY_QRYINTVL)

To configure this script to run using cron, do:
 * Use `crontab -e` to open up and edit the current user’s cron jobs in the crontab file
 * Add `Minute(0-59) Hour(0-24) Day_of_month(1-31) Month(1-12) Day_of_week(0-6) <Full Path to data_collection_checks.sh>` to the crontab file
  * A `*` can be added to any field in the crontab expression to mean always
  * A `*/` can be added before a number in any field to execute the script to run every certain amount of time
  * Example: Add `0 */23 * * * <Full Path to data_collection_checks.sh>` to run the script at minute 0 of every 23rd hour
 * To display all scheduled cron jobs for the current user, you can use `crontab -l`
 * You can halt the cron job by opening up the crontab file and commenting out the job with `#` or delete the crontab expression from the crontab file

# Running Lantern Services Individually

## Internal Services

See each internal service's README to see how to run that service as a standalone service.

  * [Endpoint Manager](endpointmanager/README.md)
  * [Capability Querier](capabilityquerier/README.md)
  * [Capability Receiver](capabilityreceiver/README.md)
  * [Lantern Message Queue](lanternmq/README.md)

## External Services

* PostgreSQL
* RabbitMQ

#### Initializing the Database by hand

* If you have postgres installed locally (which you can do with `brew install postgresql`), you can do:

  ```
  psql -h <container name> -p <port> -U <username> -d <database> -a -f <setup script>
  ```

  For example:

  ```
  psql -h postgres -p 5432 -U postgres -d postgres -a -f endpointmanager/dbsetup.sql
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

You can also check that you have access to the admin page by navigating to `http://localhost:15672` and using username and password `lanternadmin:lanternadmin`.

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

# Hosting

To start Lantern automatically on linux system reboot create a systemd service file named Lantern-app.service in /etc/systemd/system directory

`sudo vi /etc/systemd/system/lantern-app.service`

Contents of the lantern-app.service file:
```
[Unit]
Description=Lantern Application Service
Requires=docker.service
After=docker.service

[Service]
Restart=always
Type=forking
TimeoutStartSec=0
WorkingDirectory=/home/centos/lantern-back-end
ExecStart=/bin/bash -c 'make run_prod'
ExecStop=/bin/bash -c 'make stop_prod'

[Install]
WantedBy=multi-user.target
```  

Enable the Lantern-App service on start up
`sudo systemctl enable lantern-app`

Start the Lantern service
`sudo systemctl start lantern-app`

To stop the service 
`sudo systemctl stop lantern-app`

To view the status
`systemctl status lantern-app`

Changes made to the lantern-app.service file will need to be reloaded and restarted 
```
sudo systemctl daemon-reload
sudo systemctl restart lantern-app
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
Code going through PR should pass the golang lintr invoked by running:
```bash
golangci-lint run -E gofmt
```
You may have to install golangci-lint first. To do this on a Mac you can run:
```bash
brew install golangci/tap/golangci-lint
```
More information about golangci-lint can be found [here](https://github.com/golangci/golangci-lint)

Code should also run the R lintr without receiving any lintr warning messages, invoked by running:
```bash
Rscript -e lintr::lint_dir(linters = lintr::with_defaults(object_usage_linter=NULL, closed_curly_linter = NULL, open_curly_linter = NULL, line_length_linter = NULL, object_name_linter = NULL))
```

You may have to install R and the R lintr package first. To do this on a Mac, you can install R from the internet, and then you can run:
```bash
echo 'install.packages("lintr", dependencies = TRUE, repos="http://cran.rstudio.com/")' | R --save
```

Or you may run the lintr.sh script in the ./scripts directory which will install the lintr package if not already installed, then print out and throw an error if the lintr recommends any changes to the R code

More information about the R lintr can be found [here](https://github.com/jimhester/lintr)


## Running Shiny Load Test
To run a load test, install shiny load test and shiny cannon [here](https://rstudio.github.io/shinyloadtest/)

Record a user session in the R console
```
shinyloadtest::record_session('http://localhost:3838/', output_file = '/path/to/output/recording.log')
```
This will open a browser with the shiny dashboard loaded. Interact with the app and close the tab/browser to end session recording.

Run a load test in the terminal using the recorded session
```
shinycannon <path/to/recording.log> http://localhost:3838/ --workers 5 --loaded-duration-minutes 2 --output-dir /Path/for/output/run
```
The workers are number of concurrent users to simulate. The loaded duration minutes is how long to run the test for once it warms up (reaches the specified number of workers). A worker will repeat the session as many times as possible within the loaded duration. 

Analyze the results
```
df <- shinyloadtest::load_runs("5 workers" = "/path/to/test/run")
shinyloadtest::shinyloadtest_report(df, "/path/for/output/report.html")
```
This will load the results of the test run into a dataframe then generate a report. 

## GoMod
If you make changes in one package and would like to use those changes in another package that depends on the first package that you change, commit your code and run `make update_mods branch=<your_working_branch>` This is especially relevant when running your new code in docker images built for the e2e, capabilityquerier, endpointmanager, capabilityreciever packages as the go.mod files are what will be used to determine which versions of the packages should be checked out when the docker images are built. Your final commit in a PR should be the go.mod and go.sum updates that occur as a result of running `make update_mods branch=<your_working_branch>`


# License

Copyright 2021 The MITRE Corporation

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

```
http://www.apache.org/licenses/LICENSE-2.0
```

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.


