#!/bin/sh

SHELL=/bin/sh
PATH=/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin

chmod +rx query-endpoint-resources.sh; ./query-endpoint-resources.sh
exec docker exec -it lantern-back-end_endpoint_manager_1 /etc/lantern/populateEndpoints.sh