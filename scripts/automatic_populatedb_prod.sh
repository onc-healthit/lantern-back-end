#!/bin/sh

EMAIL=

# Commenting out SHELL and PATH variables as they are causing Go version error during the execution of query-endpoint-resources.sh
#SHELL=/bin/sh
#PATH=/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin
log_file="/etc/lantern/logs/automatic_populatedb_prod_logs.txt"
current_datetime=$(date +"%Y-%m-%d %H:%M:%S")
LOGFILE=populatedb_logs_$(date +%Y%m%d%H%M%S).txt

chmod +rx query-endpoint-resources.sh; ./query-endpoint-resources.sh

echo "$current_datetime - Populating db with endpoint information..." >> $log_file

docker exec lantern-back-end_endpoint_manager_1 /etc/lantern/populatedb.sh || {
  echo "$current_datetime - Lantern failed to save endpoint information in database." >> $log_file
  echo "Lantern failed to save endpoint information in database." | /usr/bin/mail -s "Automatic prod database population error." ${EMAIL}
}

echo "$current_datetime - done" >> $log_file

docker cp lantern-back-end_endpoint_manager_1:/etc/lantern/populatedb_logs.txt /etc/lantern/logs/populatedb_logs/${LOGFILE}