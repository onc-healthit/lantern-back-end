#!/bin/sh

EMAIL=

SHELL=/bin/sh
PATH=/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin

chmod +rx query-endpoint-resources.sh; ./query-endpoint-resources.sh
docker exec lantern-back-end_endpoint_manager_1 /etc/lantern/populateEndpoints.sh || echo "Lantern failed to save endpoint information in database after endpoint resource list updates." | /usr/bin/mail -s "Automatic endpoint update and database population error." ${EMAIL}