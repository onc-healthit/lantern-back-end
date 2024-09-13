#!/bin/sh

EMAIL=

SHELL=/bin/sh
PATH=/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin
log_file="/etc/lantern/logs/automatic_endpoint_update_logs.txt"
current_datetime=$(date +"%Y-%m-%d %H:%M:%S")

chmod +rx query-endpoint-resources.sh; ./query-endpoint-resources.sh

cd ../resources
cp prod_resources resources
docker cp resources lantern-back-end_endpoint_manager_1:/etc/lantern

docker exec lantern-back-end_endpoint_manager_1 /etc/lantern/populateEndpoints.sh || {
    echo "$current_datetime - Lantern failed to save endpoint information in database after endpoint resource list updates." >> $log_file
    echo "Lantern failed to save endpoint information in database after endpoint resource list updates." | /usr/bin/mail -s "Automatic endpoint update and database population error." ${EMAIL}
    exit 0
}

rm -r resources

echo "$current_datetime - Automatic Endpoint Update complete" >> $log_file