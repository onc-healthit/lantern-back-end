#!/bin/sh

# This script assumes that you are using the postgres service defined in the docker-compose.yml for the database 
# If you use your own database solution, this script will need to query the database you chose to use instead of querying the postgres docker container

SHELL=/bin/sh
PATH=/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin

EMAIL=
DB_NAME=
DB_USER=
QUERY_INTERVAL=
log_file="/etc/lantern/logs/data_collection_check_logs.txt"
current_datetime=$(date +"%Y-%m-%d %H:%M:%S")

QUERY_INTERVAL_SECONDS=$((${QUERY_INTERVAL}*60))
DATE=$(date +%s)
PASTDATE=$((${DATE}-${QUERY_INTERVAL_SECONDS}))
QUERY=$(echo "SELECT count(*) FROM fhir_endpoints_info_history WHERE floor(extract(epoch from fhir_endpoints_info_history.updated_at)) BETWEEN ${PASTDATE} AND ${DATE}")
COUNT=$(docker exec -t lantern-back-end_postgres_1 psql -t -U${DB_USER} -d ${DB_NAME} -c "${QUERY}") || {
    echo "$current_datetime - Error: Database is down" >> $log_file
    echo "Error: Database is down" | /usr/bin/mail -s "Cron Job Error" ${EMAIL}
    exit 0
}
NUMBER=$(echo ${COUNT} | tr -cd '[[:digit:]]')
   
if [ "${NUMBER}" -eq "0" ]; then  
    echo "$current_datetime - Error: Lantern data collection has stopped" >> $log_file
    echo "Error: Lantern data collection has stopped" | /usr/bin/mail -s "Cron Job Error" ${EMAIL}
    exit 0
fi

echo "$current_datetime - Lantern data collection successful" >> $log_file