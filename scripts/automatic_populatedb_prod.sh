#!/bin/sh

EMAIL=

# Commenting out SHELL and PATH variables as they are causing Go version error during the execution of query-endpoint-resources.sh
#SHELL=/bin/sh
#PATH=/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin
log_file="/etc/lantern/logs/automatic_populatedb_prod_logs.txt"
current_datetime=$(date +"%Y-%m-%d %H:%M:%S")
population_job_start_datetime="$current_datetime"
LOGFILE=populatedb_logs_$(date +%Y%m%d%H%M%S).txt

chmod +rx query-endpoint-resources.sh; ./query-endpoint-resources.sh

# Load environment variables from .env file
cd ..
export $(cat .env)
cd scripts

echo "$(date +"%Y-%m-%d %H:%M:%S") - Downloading CHPL Service Base URL List..." >> $log_file

# Download CHPL Service Base URL List CSV for tracking developers sharing list sources
CHPL_CSV_URL="https://chpl.healthit.gov/rest/service-base-url-list/download?api_key=${LANTERN_CHPLAPIKEY}"
CHPL_CSV_PATH="../resources/prod_resources/chpl_service_base_url_list.csv"
CHPL_CSV_PATH_CONTAINER="/etc/lantern/resources/chpl_service_base_url_list.csv"

curl -s -o "$CHPL_CSV_PATH" "$CHPL_CSV_URL" || {
  echo "$(date +"%Y-%m-%d %H:%M:%S") - Failed to download CHPL Service Base URL List CSV." >> $log_file
}

if [ -f "$CHPL_CSV_PATH" ] && [ -s "$CHPL_CSV_PATH" ]; then
  echo "$(date +"%Y-%m-%d %H:%M:%S") - Parsing CHPL Service Base URL List..." >> $log_file
  docker exec lantern-back-end-endpoint_manager-1 /bin/sh -c "cd /go/src/app/cmd/chplsharedlistsources && go run main.go $CHPL_CSV_PATH_CONTAINER" || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Failed to parse CHPL Service Base URL List." >> $log_file
  }
  echo "$(date +"%Y-%m-%d %H:%M:%S") - done" >> $log_file
fi

echo "$(date +"%Y-%m-%d %H:%M:%S") - Populating db with endpoint information..." >> $log_file

if docker exec lantern-back-end-endpoint_manager-1 /etc/lantern/populatedb.sh; then
  echo "$(date +"%Y-%m-%d %H:%M:%S") - Inserting old CHPL query errors into its history table..." >> $log_file

  docker exec lantern-back-end-postgres-1 psql -U lantern -d lantern -c "INSERT INTO endpoint_query_errors_history(list_source, error_message, queried_at, created_at) SELECT list_source, error_message, queried_at, created_at FROM endpoint_query_errors WHERE created_at < '$population_job_start_datetime';" >> $log_file 2>&1 || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Failed to insert old CHPL query errors into its history table." >> $log_file
  }
  
  echo "$(date +"%Y-%m-%d %H:%M:%S") - Deleting old CHPL query errors..." >> $log_file

  docker exec lantern-back-end-postgres-1 psql -U lantern -d lantern -c "DELETE FROM endpoint_query_errors WHERE created_at < '$population_job_start_datetime';" >> $log_file 2>&1 || {
    echo "$(date +"%Y-%m-%d %H:%M:%S") - Failed to delete old CHPL query errors." >> $log_file
  }
else
  echo "$(date +"%Y-%m-%d %H:%M:%S") - Lantern failed to save endpoint information in database." >> $log_file
  echo "Lantern failed to save endpoint information in database." | /usr/bin/mail -s "Automatic prod database population error." ${EMAIL}
fi

echo "$(date +"%Y-%m-%d %H:%M:%S") - done" >> $log_file

docker cp lantern-back-end-endpoint_manager-1:/etc/lantern/populatedb_logs.txt /etc/lantern/logs/populatedb_logs/${LOGFILE}