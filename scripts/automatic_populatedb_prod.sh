#!/bin/sh

EMAIL=

# Commenting out SHELL and PATH variables as they are causing Go version error during the execution of query-endpoint-resources.sh
#SHELL=/bin/sh
#PATH=/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin
log_file="/etc/lantern/logs/automatic_populatedb_prod_logs.txt"
current_datetime=$(date +"%Y-%m-%d %H:%M:%S")
LOGFILE=populatedb_logs_$(date +%Y%m%d%H%M%S).txt

chmod +rx query-endpoint-resources.sh; ./query-endpoint-resources.sh

# Load environment variables from .env file
cd ..
export $(cat .env)
cd scripts

echo "$current_datetime - Downloading CHPL Service Base URL List..." >> $log_file

# Download CHPL Service Base URL List CSV for tracking developers sharing list sources
CHPL_CSV_URL="https://chpl.healthit.gov/rest/service-base-url-list/download?api_key=${LANTERN_CHPLAPIKEY}"
CHPL_CSV_PATH="../resources/prod_resources/chpl_service_base_url_list.csv"
CHPL_CSV_PATH_CONTAINER="/etc/lantern/resources/chpl_service_base_url_list.csv"

curl -s -o "$CHPL_CSV_PATH" "$CHPL_CSV_URL" || {
  echo "$current_datetime - Failed to download CHPL Service Base URL List CSV." >> $log_file
}

if [ -f "$CHPL_CSV_PATH" ] && [ -s "$CHPL_CSV_PATH" ]; then
  echo "$current_datetime - Parsing CHPL Service Base URL List..." >> $log_file
  docker exec lantern-back-end-endpoint_manager-1 /bin/sh -c "cd /go/src/app/cmd/chplsharedlistsources && go run main.go $CHPL_CSV_PATH_CONTAINER" || {
    echo "$current_datetime - Failed to parse CHPL Service Base URL List." >> $log_file
  }
  echo "$current_datetime - done" >> $log_file
fi

echo "$current_datetime - Populating db with endpoint information..." >> $log_file

docker exec lantern-back-end-endpoint_manager-1 /etc/lantern/populatedb.sh || {
  echo "$current_datetime - Lantern failed to save endpoint information in database." >> $log_file
  echo "Lantern failed to save endpoint information in database." | /usr/bin/mail -s "Automatic prod database population error." ${EMAIL}
}

echo "$current_datetime - done" >> $log_file

docker cp lantern-back-end-endpoint_manager-1:/etc/lantern/populatedb_logs.txt /etc/lantern/logs/populatedb_logs/${LOGFILE}