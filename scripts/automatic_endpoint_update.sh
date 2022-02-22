#!/bin/sh

SHELL=/bin/sh
PATH=/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin

chmod +rx query-endpoint-resources.sh; ./query-endpoint-resources.sh
chmod +rx populatedb_prod.sh; ./populatedb_prod.sh