# Lantern

Lantern is an open source tool developed by the Office of the National Coordinator for Health Information Technology (ONC) and Mettle Solutions, LLC. that monitors and provides analytics about the availability and adoption of FHIR API service base URLs (endpoints) across healthcare organizations in the United States. It also gathers information about FHIR Capability Statements returned by these endpoints and provides visualizations to show FHIR adoption and patient data availability. For more information, check out the “About Lantern” page on our website, [https://www.lantern.healthit.gov](https://lantern.healthit.gov/?tab=dashboard_tab).

# Index
* [Running Lantern - Basic Flow](#running-lantern---basic-flow)
* [Using Docker Compose](#using-docker-compose)
* [Testing Lantern - Basic Flow](#testing-lantern---basic-flow)
* [Make Commands](#make-commands)
* [Configure Data Collection Failure System](#configure-data-collection-failure-system)
* [Configure Backup System](#configure-backup-system)
* [Running Lantern Services Individually](#running-lantern-services-individually)
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

## Clean your environment

**This is optional!**

If you'd like to start with a clean slate, run:

```bash
make clean
```

This removes all docker images, networks, and local volumes.

## Start Lantern

### Production Environment

1. To start up Lantern with a production environment, in your terminal, run:

    ```bash
    make run_prod
    ```

    This starts all of the following services:
    * **PostgreSQL** - application database (port 5432)
    * **LanternMQ (RabbitMQ)** - the message queue (localhost:15672)
    * **Capability Querier** - queries the endpoints for their Capability Statements every 23 hours. Starting the service the first time will also query the endpoints.
    * **Capability Receiver** - receives the Capability Statements from the queue, performs validations and saves the results to the database table `fhir_endpoints_info`
    * **Endpoint Manager** - sends endpoints to the capability querying queues
    *	**Shinydashboard** – the website (localhost:8090)

    **Note**: While running with the production environment, you will not be able to externally connect to the database or the message queue. Only the services can access those ports.

2. **If you have a clean database or want to update the data in your database**
    1. Run the following command to update the endpoint lists found in the production resources directory, `lantern-back-end/resources/prod_resources`.
     ```bash
      make update_source_data_prod
      ```
      This command will automatically query all the endpoint sources listed in `EndpointResourceList.json` for their endpoints. It will also query CHPL for its list of endpoint list sources, which is saved to `CHPLEndpointResourcesList.json`, and then query the endpoint lists (that have parsers written for them) for their endpoints. It then saves the first 10 endpoints from each list into their own files in the development resources directory, lantern-back-end/resources/dev_resources.
      
    **Note**: Google chrome must be installed in order to run this command and run the webscrapers needed for certain list sources.
    
    2. Run the following command to add data to the database, using the data found in `lantern-back-end/resources/prod_resources`.

    ```bash
    make populatedb_prod
    ```
    This runs the following tasks inside the endpoint manager container:
    * query the **National Plan & Provider Enumeration System (NPPES)** for their monthly export endpoint and NPI data files. **Note**: The command removes any entries in the NPI data file that are not organization entries.
    * the **endpoint populator**, which iterates over the list of endpoint sources and adds them to the database.
    * the **CHPL querier**, which requests Health IT product, vendor, and certification information from CHPL and adds these to the database
    * the **NPPES endpoint populator**, which adds endpoint data from the monthly NPPES export to the database. 
    * the **NPPES org populator**, which adds provider data from the monthly NPPES export to the database.
    * the **data validator**, which ensures that the amount of data in the database can successfully be quried in the 23 hour query interval.

3. **If you want to re-query and re-receive capability statements outside the refresh interval**, run the following:

    ```bash
    docker restart lantern-back-end_endpoint_manager_1
    ```

### Development Environment

1. To start up Lantern with a development environment, in your terminal, run:

    ```bash
    make run
    ```

    This starts all of the following services:
    * **PostgreSQL** - application database (localhost:5432)
    * **LanternMQ (RabbitMQ)** - the message queue (localhost:15672)
    * **Capability Querier** - queries the endpoints for their FHIR Capability Statements every 23 hours. Starting the service the first time will also query the endpoints.
    * **Capability Receiver** - receives the Capability Statements from the queue, performs validations and saves the results to the database table `fhir_endpoints_info`
    * **Endpoint Manager** - sends endpoints to the capability querying queues
    *	**Shinydashboard** – the website (localhost:3838)

    **Note**: If you are interested in viewing the queue data and interacting with the queue directly, you can go to localhost:15672 while the system is running (only works while running as development). Use lanternadmin as both the username and password on the login page.

2. **If you have a clean database or want to update the data in your database**
   1. run the following command to update the endpoint lists found in the production resources directory, `lantern-back-end/resources/prod_resources`.
        ```bash
         make update_source_data
         ```
         This command will automatically query all the endpoint sources listed in `EndpointResourceList.json` for their endpoints. It will also query CHPL for its list of endpoint list sources, which is saved to `CHPLEndpointResourcesList.json`, and then query the endpoint lists (that have parsers written for them) for their endpoints. It then saves the first 10 endpoints from each list into their own files in the development resources directory, lantern-back-end/resources/dev_resources. It then queries NPPES for its endpoint and NPI data files, cut out all the entries in the NPI data file that are not organization entries, and then creates a copy of each file and reduce them to 1000 lines for development purposes.

        **Note**: Resources can be moved from lantern-back-end/resources/prod_resources to lantern-back-end/resources/dev_resources to be used in the development environment.

        **Note**: Google Chrome must be installed in order to run this command and run the webscrapers needed for certain list sources.

        **Note**: If you only want to query the endpoint sources without also querying NPPES, you can run `make update_source_data_prod` (see **Production Environment** section).

    2. Run the following command to add data to the database, using the data found in `lantern-back-end/resources/dev_resources`.
   
         ```bash
         make populatedb
         ```
 
        **Note**: The `dev_resources` directory contains fewer endpoints, to reduce unnecessary load on the servers hosting the endpoints we are querying while developing locally.
        
       This runs the following tasks inside the endpoint manager container:
         * the **endpoint populator**, which iterates over the list of endpoint sources and adds them to the database.
        * the **CHPL querier**, which requests Health IT product, vendor, and certification information from CHPL and adds these to the database
        * the **NPPES endpoint populator**, which adds endpoint data from the monthly NPPES export to the database. 
        * the **NPPES org populator**, which adds provider data from the monthly NPPES export to the database.
        * the **data validator**, which ensures that the amount of data in the database can successfully be quried in the 23 hour query interval.

       The `populate_db.sh` script expects the `dev_resources` directory to contain the following files:
         * **endpoint_pfile.csv** - endpoint_pfile from the data dissemination package downloaded from https://download.cms.gov/nppes/NPI_Files.html
         * **npidata_pfile.csv** - npidata_pfile from the data dissemination package downloaded from https://download.cms.gov/nppes/NPI_Files.html 
         **Note**: This file can take a very long time to load so for development purposes, the load time can be reduced by only using the first 100000 entries. The first 100000 entries can be obtained by running `head -n 100000 npidata_pfile_20050523-20191110.csv >> npidata_pfile.csv`. Alternatively, running `make update_source_data` adds truncated npi files to the `dev_resources` directory as well.


3. **If you want to re-query and re-receive Capability Statements outside the refresh interval** run the following:

    ```bash
    docker restart lantern-back-end_endpoint_manager_1
    ```

## Stop Lantern

### Production Environment

To stop Lantern when running with a production environment, run:
```bash
make stop_prod
```

### Development Environment
To stop Lantern when running with a development environment, run:
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
  * `db/migration/certs`

# Using Docker Compose

Lantern is a multi-container application that runs using Docker, and therefore Docker must be downloaded in order to set up the project. You can download Docker Desktop [here](https://docs.docker.com/get-docker/), which includes Docker Compose. You can find more information about Docker compose [here](https://docs.docker.com/compose/). Using Docker compose is a three-step process:
  1. Define app environment with a Dockerfile so it can be reproduced anywhere
  2. Define services that make up app in docker-compose.yml so they can be run together in an isolated environment
  3. Run docker compose up and the Docker compose command starts and runs the entire app. Lantern's Makefile command `make run` and `make run prod` can be used to run the docker compose command to start the application

# Windows Instructions
The Lantern system was developed on Mac computers and the production and staging servers run Linux. The code in GitHub will run as-is while running on either of these operating systems. If running the system on a Windows machine, however, a few changes will have to be made to the code directly. 
  * In all script files, change end of line sequence from CRLF to LF   
    * If using Visual Studio Code for development, it can easily be changed by clicking an option in the bottom right-hand corner to change the end of line sequence 
  * In the wait-for-it.sh file in the scripts directory, change #!/usr/bin/env bash to #!/bin/bash 
  * The environmental variable for the CHPL API key may not work. If so, the key can be entered manually into endpointmanager/pkg/chplquerier/chplquerier.go file, where originally viper.GetString("chplapikey") 
  * The make populatedb command in the Makefile may not work due to the machine generating the name of the container as lantern-back-end-endpoint_manager-1 rather than lantern-back-end_endpoint_manager_1. The fix is to change the make command to: docker exec -it lantern-back-end-endpoint_manager-1 /etc/lantern/populatedb.sh  
    * This may also have to be done for any other make commands that specifically reference Docker containers 

# Testing Lantern - Basic Flow

There are three types of tests for Lantern and three corresponding commands:

| test type | command |
| --- | --- |
| unit | `make test` |
| integration | `make test_int` |
| end-to-end |  `make test_e2e` |

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
|`make populatedb` | Should be used with development environment by running `make run` first. Populates the database with the endpoint resource list information and NPPES information found in the `resources/dev_resources` directory.| 
|`make populatedb_prod` | Should be used with production environment by running `make run_prod` first. Populates the database with the endpoint resource list information found in the `resources/prod_resources` directory, and queries NPPES for its latest information and automatically stores it in the database before deleting the files.| 
|`make backup_database` | saves a database backup .sql file in the lantern base directory with name lantern_backup_`<timestamp>`.sql|
|`make restore_database file=<backup file name>` | restores the backup database that the 'file' parameter is set to|
|`make migrate_database direction=<up/down> force_version=<migration version number to force db to>` | Starts the postgres service and runs the next `*.up.sql` migration in the `db/migration/migrations` directory that has not yet been run. Must run this command the number times equal to the number of migrations you want to run. The optional direction parameter can be included to specify whether you want to run an up or down migration, and the force_version parameter can be included to force the database to a specific migration version before running the next migration. If the direction parameter is omitted, it will run an up migration by default, and if the force_version parameter is omitted, it runs the next migration that has not yet been run. The order of these two paramaters is important, and so if you want to include the force_version parameter, you must also include the direction parameter before it. |
|`make update_source_data` | Automatically queries the endpoint lists listed in the `EndpointResourcesList.json` file found in the `resources/prod_resources` directory, queries CHPL for its list of endpoint lists and stores the data in a file in the same directory, queries the NPPES NPI and endpoint data, and stores truncated versions of the files in the `resources/dev_resources` directory. |
|`make update_source_data_prod` | Automatically queries the endpoint lists listed in the `EndpointResourcesList.json` file found in the `resources/prod_resources` directory and queries CHPL for its list of endpoint lists and stores the data in a file in the same directory. |
|  `make lint` | Runs the R and golang linters |
|  `make lint_go` | Runs the golang lintr |
|  `make lint_R` | Runs the R lintr |
| `make history_pruning` | Prunes the fhir_endpoint_info_history table to remove duplicate entries |
| `make create_archive start=<start date> end=<end date> file=<archive file name>` | Creates an archive of the data in the database between the given dates in a JSON format and saves it to the given 'file' name. The dates format is '2021-01-31' (year, month, date). Example: `make create_archive start=2020-06-01 end=2021-06-01 file=archive_file.json`. Note: If the archive period includes any time between the current date and the LANTERN_PRUNING_THRESHOLD, then the given number of updates might be higher than expected because the history pruning algorithm is only run on data older than the threshold. |
|  `make migrate_validations direction=<up/down>` | Runs validation migrations when direction is set to up. If direction is set to down, undos validation migrations |
|  `make migrate_resources direction=<up/down>` | Runs resources migrations when direction is set to up. If direction is set to down, undos resources migrations |

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

# Configure Backup System

You can configure a backup system using cron and the backup.sh script located in the scripts directory to send an email notification whenever the current backup becomes available.

To set up the script for this backup system, you must insert the correct information into the following variables located at the beginning of the backup script. The DB_NAME and DB_USER variables used in the script should match their corresponding environmental variable (shown in parentheses below) defined in the .env file:
  * Set the EMAIL variable to the email you want the backup system to send alerts to
  * Set the DB_NAME variable to name of your database (LANTERN_DBNAME)
  * Set the DB_USER variable to the name of the database user (LANTERN_DBUSER)
  * Set the BACKUP_DIR variable to wherever you want to save the backup file

To configure this script to run using cron, do:
 * Use `crontab -e` to open up and edit the current user’s cron jobs in the crontab file
 * Add `Minute(0-59) Hour(0-24) Day_of_month(1-31) Month(1-12) Day_of_week(0-6) <Full Path to backup.sh>` to the crontab file
  * A `*` can be added to any field in the crontab expression to mean always
  * A `*/` can be added before a number in any field to execute the script to run every certain amount of time
  * Example: Add `0 */23 * * * <Full Path to backup.sh>` to run the script at minute 0 of every 23rd hour
 * To display all scheduled cron jobs for the current user, you can use `crontab -l`
 * You can halt the cron job by opening up the crontab file and commenting out the job with `#` or delete the crontab expression from the crontab file

# Configure CHPL Endpoint List Updater

You can configure a CHPL Endpoint List update checker using cron and the chpl_endpoint_list_check.sh script located in the scripts directory to send an email notification whenever the CHPL Endpoint List has been updated with new URLs. This script queries CHPL's list of endpoints and compares it to the CHPL endpoint list currently stored in the `resources/prod_resources` directory. If the CHPL list has been updated, the system will send an alert email with the updated URLs, and it will automatically update the list stored in `resources/prod_resources` with the new endpoint lists. 

To set up the script for this CHPL endpoint list updater, you must insert the correct information into the following variables located at the beginning of the update checker script.
  * Set the EMAIL variable to the email you want the CHPL endpoint list updater to send alerts to

To configure this script to run using cron, do:
 * Use `crontab -e` to open up and edit the current user’s cron jobs in the crontab file
 * Add `Minute(0-59) Hour(0-24) Day_of_month(1-31) Month(1-12) Day_of_week(0-6) cd <Full path to scripts directory> && ./chpl_endpoint_list_check.sh` to the crontab file
  * A `*` can be added to any field in the crontab expression to mean always
  * A `*/` can be added before a number in any field to execute the script to run every certain amount of time
  * Example: Add `0 */23 * * * cd <Full path to scripts directory> && ./chpl_endpoint_list_check.sh` to run the script at minute 0 of every 23rd hour
 * To display all scheduled cron jobs for the current user, you can use `crontab -l`
 * You can halt the cron job by opening up the crontab file and commenting out the job with `#` or delete the crontab expression from the crontab file

# Configure Automatic Endpoint Resource Updater

You can configure an automatic endpoint resource update checker using cron and the `automatic_endpoint_update.sh` script located in the scripts directory to automatically update the resource lists found in the `resources/prod_resources` directory and then save all this information in the database. The task will send an email notification if the automatic update fails to be saved in the database. 

To set up the script for this automatic endpoint resource updater, you must insert the correct information into the following variables located at the beginning of the update checker script.
  * Set the EMAIL variable to the email you want the automatic endpoint resource updater to send failure alerts to in the `automatic_endpoint_update.sh` script

To configure this script to run using cron, do:
 * Use `crontab -e` to open up and edit the current user’s cron jobs in the crontab file
 * Add `Minute(0-59) Hour(0-24) Day_of_month(1-31) Month(1-12) Day_of_week(0-6) cd <Full path to scripts directory> && ./automatic_endpoint_update.sh` to the crontab file
  * A `*` can be added to any field in the crontab expression to mean always
  * A `*/` can be added before a number in any field to execute the script to run every certain amount of time
  * Example: Add `1 * 1 */1 * cd <Full path to scripts directory> && ./automatic_endpoint_update.sh` to run the script at minute 1 on day 1 in every month
 * To display all scheduled cron jobs for the current user, you can use `crontab -l`
 * You can halt the cron job by opening up the crontab file and commenting out the job with `#` or delete the crontab expression from the crontab file

 # Configure Automatic Production Database Population

You can configure an automatic production database population system using cron and the `automatic_populatedb_prod.sh` script located in the scripts directory to save all the endpoint information from the endpoint resource lists found in the `resources/prod_resources` directory into the database. The script also downloads the most recent NPPES file, stores all the information from that file into the database, and then deletes the NPPES file in order to save storage space. The task will send an email notification if the endpoint list information or NPPES information fails to be saved in the database. 

To set up the script for this automatic production database population system, you must insert the correct information into the following variables located at the beginning of the automatic production database populator script.
  * Set the EMAIL variable to the email you want the automatic production database populator to send failure alerts to in the `automatic_populatedb_prod.sh` script

To configure this script to run using cron, do:
 * Use `crontab -e` to open up and edit the current user’s cron jobs in the crontab file
 * Add `Minute(0-59) Hour(0-24) Day_of_month(1-31) Month(1-12) Day_of_week(0-6) cd <Full path to scripts directory> && ./automatic_populatedb_prod.sh` to the crontab file
  * A `*` can be added to any field in the crontab expression to mean always
  * A `*/` can be added before a number in any field to execute the script to run every certain amount of time
  * Example: Add `1 * 1 */1 * cd <Full path to scripts directory> && ./automatic_populatedb_prod.sh` to run the script at minute 1 on day 1 in every month
 * To display all scheduled cron jobs for the current user, you can use `crontab -l`
 * You can halt the cron job by opening up the crontab file and commenting out the job with `#` or delete the crontab expression from the crontab file


# Configure History Pruning System

You can configure a system to run the history pruning using cron and the history_prune.sh script located in the scripts directory to prune the fhir_endpoints_info_history table.
    * NOTE: The history pruning process already runs automatically by the endpoint manager every query interval after it finishes sending all the endpoints to the capability querier.
To configure this script to run using cron, do:
 * Use `crontab -e` to open up and edit the current user’s cron jobs in the crontab file
 * Add `Minute(0-59) Hour(0-24) Day_of_month(1-31) Month(1-12) Day_of_week(0-6) cd <Full Path to script directory> && ./history_prune.sh` to the crontab file
  * A `*` can be added to any field in the crontab expression to mean always
  * A `*/` can be added before a number in any field to execute the script to run every certain amount of time
  * Example: Add `0 */23 * * * cd <Full Path to script directory> && ./history_prune.sh` to run the script at minute 0 of every 23rd hour
 * To display all scheduled cron jobs for the current user, you can use `crontab -l`
 * You can halt the cron job by opening up the crontab file and commenting out the job with `#` or delete the crontab expression from the crontab file

# Perform History Cleanup

You can perform a manual history cleanup operation which will prune all repetitive entries, determined using comparisons from the history pruning algorithm, present in the fhir_endpoints_info_history table. It will also prune the corresponding entries from the validations and validation_results table. It is a two-step process.

Step 1: Collect the identifiers of repetitive entries

Change directory to the /scripts inside lantern-back-end and run:

 ```bash
    ./duplicate_info_history_check.sh
  ```

This will start capturing the identifiers of repetitive entries in the fhir_endpoints_info_history table and store it in duplicateInfoHistoryIds.csv file inside the /home directory of the lantern-back-end_endpoint_manager_1 container.

To retrieve the csv file, change directory to /lantern-back-end and run:

 ```bash
    docker cp lantern-back-end_endpoint_manager_1:/home/duplicateInfoHistoryIds.csv .
  ```

Step 2: Perform the history cleanup

Change directory to the /scripts inside lantern-back-end and run:

 ```bash
    ./history_cleanup.sh
  ```

This will start the deletion of data from fhir_endpoints_info_history table using the captured identifiers of repetitive entries.

 ```bash
    ./validations_cleanup.sh
  ```

This will start the deletion of data from validations and validation_results tables using the captured identifiers of repetitive entries.

Note: Ensure that the duplicateInfoHistoryIds.csv file is present in /lantern-back-end before executing the above scripts.

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

### Make Commands
To make it easier, there are make commands that run the lintrs. The Go and R lintrs can be run on the Lantern code at any time using
`make lint`
There are also separate make commands to run the Go lintr and R lintr, which are
`make lint_go`
and
`make lint_R`.

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

Copyright 2024 Mettle Solutions, LLC.

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

```
http://www.apache.org/licenses/LICENSE-2.0
```

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.


